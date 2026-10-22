// Command loadtest membuktikan ada atau tidaknya race condition pada
// endpoint KRS. Folder tools/ bukan bagian aplikasi dan tidak ikut di-deploy.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"latihan_fiber/app/model"
	"latihan_fiber/config"
	"latihan_fiber/database"
	"latihan_fiber/helper"
)

type testUser struct {
	ID       int
	Username string
	Role     string
	Token    string
}

type courseInfo struct {
	ID   int
	Code string
}

func main() {
	mode := flag.String("mode", "crowd", "mode uji: crowd | credits | swap")
	baseURL := flag.String("base", "http://localhost:3000", "base URL server")
	reset := flag.Bool("reset", true, "kembalikan data uji sebelum menjalankan uji")
	flag.Parse()

	config.LoadEnv()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	pool, err := database.NewPool(ctx)
	if err != nil {
		fail("gagal terhubung ke database: %v", err)
	}
	defer pool.Close()

	jwtSecret := config.GetEnv("JWT_SECRET", "")
	if len(jwtSecret) < 32 {
		fail("JWT_SECRET harus diisi minimal 32 karakter")
	}
	jwtManager := helper.NewJWTManager(
		jwtSecret,
		config.GetEnv("JWT_ISSUER", "praktikum-backend"),
		time.Duration(config.GetEnvInt("JWT_ACCESS_TTL_MINUTES", 15))*time.Minute,
	)

	users, err := loadUsers(ctx, pool, jwtManager)
	if err != nil {
		fail("gagal memuat akun uji: %v", err)
	}
	if len(users) < 60 {
		fail("akun uji kurang: butuh 60, ditemukan %d (jalankan migrations/007_seed_loadtest.sql)", len(users))
	}

	courses, err := loadCourses(ctx, pool)
	if err != nil {
		fail("gagal memuat kelas: %v", err)
	}

	client := &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        200,
			MaxIdleConnsPerHost: 200,
		},
	}

	if *reset {
		if err := resetData(ctx, pool); err != nil {
			fail("gagal mereset data: %v", err)
		}
	}

	switch *mode {
	case "crowd":
		runCrowd(ctx, pool, client, *baseURL, users, courses)
	case "credits":
		runCredits(ctx, pool, client, *baseURL, users)
	case "swap":
		runSwap(ctx, pool, client, *baseURL, users, courses)
	default:
		fail("mode tidak dikenal: %s", *mode)
	}
}

func loadUsers(
	ctx context.Context, pool *pgxpool.Pool, jwtManager *helper.JWTManager,
) ([]testUser, error) {
	rows, err := pool.Query(ctx,
		`SELECT id, username, role
         FROM users
         WHERE username LIKE 'mhs%'
         ORDER BY username
         LIMIT 60`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []testUser{}
	for rows.Next() {
		var u testUser
		if err := rows.Scan(&u.ID, &u.Username, &u.Role); err != nil {
			return nil, err
		}
		token, err := jwtManager.GenerateAccess(model.User{
			ID: u.ID, Username: u.Username, Role: u.Role,
		})
		if err != nil {
			return nil, err
		}
		u.Token = token
		users = append(users, u)
	}
	return users, rows.Err()
}

func loadCourses(ctx context.Context, pool *pgxpool.Pool) (map[string]int, error) {
	rows, err := pool.Query(ctx, "SELECT id, code FROM courses")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	courses := map[string]int{}
	for rows.Next() {
		var c courseInfo
		if err := rows.Scan(&c.ID, &c.Code); err != nil {
			return nil, err
		}
		courses[c.Code] = c.ID
	}
	return courses, rows.Err()
}

func loadThreeCreditCourses(ctx context.Context, pool *pgxpool.Pool) ([]courseInfo, error) {
	rows, err := pool.Query(ctx,
		`SELECT id, code FROM courses
         WHERE credits = 3 AND code <> 'IF399'
         ORDER BY code
         LIMIT 12`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	courses := []courseInfo{}
	for rows.Next() {
		var c courseInfo
		if err := rows.Scan(&c.ID, &c.Code); err != nil {
			return nil, err
		}
		courses = append(courses, c)
	}
	return courses, rows.Err()
}

func resetData(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx,
		`DELETE FROM enrollments
         WHERE user_id IN (SELECT id FROM users WHERE username LIKE 'mhs%')`)
	if err != nil {
		return err
	}
	_, err = pool.Exec(ctx,
		`UPDATE courses c
         SET enrolled_count = (
             SELECT COUNT(*) FROM enrollments e WHERE e.course_id = c.id
         )`)
	return err
}

// fire menjalankan n pekerjaan secara serentak dan menghitung hasilnya.
//
// Setiap goroutine menunggu di depan "gerbang" start yang sama. Gerbang
// baru dibuka setelah SELURUH goroutine siap, sehingga request benar-benar
// berangkat hampir bersamaan -- bukan bergiliran sesuai urutan dibuatnya.
func fire(n int, job func(i int) string) map[string]int {
	var wg sync.WaitGroup
	start := make(chan struct{})
	results := make(chan string, n)

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int, out chan<- string) {
			defer wg.Done()
			<-start
			out <- job(i)
		}(i, results)
	}

	close(start)
	wg.Wait()
	close(results)

	tally := map[string]int{}
	for r := range results {
		tally[r]++
	}
	return tally
}

func request(client *http.Client, method, url, token string, body any) string {
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return "ENCODE_ERROR"
		}
		reader = bytes.NewReader(encoded)
	}

	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return "REQUEST_ERROR"
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "NETWORK_ERROR"
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	var payload struct {
		Code string `json:"code"`
	}
	_ = json.Unmarshal(raw, &payload)
	if payload.Code != "" {
		return fmt.Sprintf("%d %s", resp.StatusCode, payload.Code)
	}
	return fmt.Sprintf("%d OK", resp.StatusCode)
}

func runCrowd(
	ctx context.Context, pool *pgxpool.Pool, client *http.Client,
	baseURL string, users []testUser, courses map[string]int,
) {
	courseID, ok := courses["IF399"]
	if !ok {
		fail("kelas IF399 tidak ditemukan (jalankan migrations/006_courses.sql)")
	}

	fmt.Printf("%d mahasiswa berebut kelas IF399 secara serentak...\n", len(users))
	start := time.Now()
	tally := fire(len(users), func(i int) string {
		url := fmt.Sprintf("%s/api/v1/courses/%d/enroll", baseURL, courseID)
		return request(client, http.MethodPost, url, users[i].Token, nil)
	})

	fmt.Println("hasil request:")
	printTally(tally)
	fmt.Printf("waktu total : %s\n", time.Since(start).Round(time.Millisecond))
	printCourse(ctx, pool, "IF399")
}

func runCredits(
	ctx context.Context, pool *pgxpool.Pool, client *http.Client,
	baseURL string, users []testUser,
) {
	courses, err := loadThreeCreditCourses(ctx, pool)
	if err != nil {
		fail("gagal memuat kelas 3 SKS: %v", err)
	}
	if len(courses) < 12 {
		fail("kelas 3 SKS kurang dari 12")
	}

	type job struct {
		token    string
		courseID int
	}
	jobs := make([]job, 0, 10*len(courses))
	for i := 0; i < 10; i++ {
		for _, course := range courses {
			jobs = append(jobs, job{token: users[i].Token, courseID: course.ID})
		}
	}

	fmt.Printf("10 mahasiswa, masing-masing mendaftar ke %d kelas 3 SKS secara serentak (batas 24 SKS)...\n",
		len(courses))
	start := time.Now()
	tally := fire(len(jobs), func(i int) string {
		j := jobs[i]
		url := fmt.Sprintf("%s/api/v1/courses/%d/enroll", baseURL, j.courseID)
		return request(client, http.MethodPost, url, j.token, nil)
	})

	fmt.Println("hasil request:")
	printTally(tally)
	fmt.Printf("waktu total : %s\n", time.Since(start).Round(time.Millisecond))
	printCreditCheck(ctx, pool)
}

func runSwap(
	ctx context.Context, pool *pgxpool.Pool, client *http.Client,
	baseURL string, users []testUser, courses map[string]int,
) {
	fromID, ok := courses["IF101"]
	if !ok {
		fail("kelas IF101 tidak ditemukan (jalankan migrations/006_courses.sql)")
	}
	toID, ok := courses["IF102"]
	if !ok {
		fail("kelas IF102 tidak ditemukan (jalankan migrations/006_courses.sql)")
	}

	groupA := make([]int, 0, 20)
	groupB := make([]int, 0, 20)
	for i := 0; i < 20; i++ {
		groupA = append(groupA, users[i].ID)
		groupB = append(groupB, users[i+20].ID)
	}
	if err := seedSwapEnrollments(ctx, pool, fromID, toID, groupA, groupB); err != nil {
		fail("gagal menyiapkan data swap: %v", err)
	}

	before, err := deadlockCount(ctx, pool)
	if err != nil {
		fail("gagal membaca statistik deadlock: %v", err)
	}

	type job struct {
		token string
		from  int
		to    int
	}
	jobs := make([]job, 0, 40)
	for i := 0; i < 20; i++ {
		jobs = append(jobs, job{token: users[i].Token, from: fromID, to: toID})
	}
	for i := 20; i < 40; i++ {
		jobs = append(jobs, job{token: users[i].Token, from: toID, to: fromID})
	}

	fmt.Println("20 pasang mahasiswa pindah kelas berlawanan arah secara serentak...")
	start := time.Now()
	tally := fire(len(jobs), func(i int) string {
		j := jobs[i]
		return request(client, http.MethodPost, baseURL+"/api/v1/enrollments/swap",
			j.token, map[string]int{
				"from_course_id": j.from,
				"to_course_id":   j.to,
			})
	})

	after, err := deadlockCount(ctx, pool)
	if err != nil {
		fail("gagal membaca statistik deadlock: %v", err)
	}

	fmt.Println("hasil request:")
	printTally(tally)
	fmt.Printf("waktu total : %s\n", time.Since(start).Round(time.Millisecond))
	fmt.Printf("deadlock terdeteksi PostgreSQL selama uji: %d\n", after-before)
	printCourse(ctx, pool, "IF101")
	printCourse(ctx, pool, "IF102")
}

func seedSwapEnrollments(
	ctx context.Context, pool *pgxpool.Pool, fromID, toID int,
	groupA, groupB []int,
) error {
	if _, err := pool.Exec(ctx,
		`INSERT INTO enrollments (course_id, user_id)
         SELECT $1, unnest($2::int[])
         ON CONFLICT DO NOTHING`, fromID, groupA); err != nil {
		return err
	}
	if _, err := pool.Exec(ctx,
		`INSERT INTO enrollments (course_id, user_id)
         SELECT $1, unnest($2::int[])
         ON CONFLICT DO NOTHING`, toID, groupB); err != nil {
		return err
	}
	_, err := pool.Exec(ctx,
		`UPDATE courses c
         SET enrolled_count = (
             SELECT COUNT(*) FROM enrollments e WHERE e.course_id = c.id
         )
         WHERE c.id IN ($1, $2)`, fromID, toID)
	return err
}

func deadlockCount(ctx context.Context, pool *pgxpool.Pool) (int64, error) {
	var count int64
	err := pool.QueryRow(ctx,
		`SELECT deadlocks FROM pg_stat_database WHERE datname = current_database()`).
		Scan(&count)
	return count, err
}

func printCourse(ctx context.Context, pool *pgxpool.Pool, code string) {
	var quota, enrolled, rows int
	err := pool.QueryRow(ctx,
		`SELECT c.quota, c.enrolled_count,
                (SELECT COUNT(*) FROM enrollments e WHERE e.course_id = c.id)
         FROM courses c WHERE c.code = $1`, code).
		Scan(&quota, &enrolled, &rows)
	if err != nil {
		fmt.Printf("%s: gagal membaca data: %v\n", code, err)
		return
	}
	status := "KONSISTEN"
	if enrolled != rows {
		status = "TIDAK KONSISTEN"
	}
	fmt.Printf("%s kuota=%d enrolled_count=%d baris enrollments=%d -> %s\n",
		code, quota, enrolled, rows, status)
}

func printCreditCheck(ctx context.Context, pool *pgxpool.Pool) {
	rows, err := pool.Query(ctx,
		`SELECT u.username, COALESCE(SUM(c.credits), 0) AS total
         FROM users u
         LEFT JOIN enrollments e ON e.user_id = u.id
         LEFT JOIN courses c ON c.id = e.course_id
         WHERE u.username LIKE 'mhs%'
         GROUP BY u.username
         ORDER BY u.username
         LIMIT 10`)
	if err != nil {
		fmt.Printf("gagal memeriksa SKS: %v\n", err)
		return
	}
	defer rows.Close()

	violations := 0
	maxCredits := 0
	for rows.Next() {
		var username string
		var total int
		if err := rows.Scan(&username, &total); err != nil {
			fmt.Printf("gagal membaca SKS: %v\n", err)
			return
		}
		if total > maxCredits {
			maxCredits = total
		}
		if total > 24 {
			violations++
		}
	}
	if violations == 0 {
		fmt.Printf("seluruh 10 mahasiswa <= 24 SKS (maksimum %d) -> KONSISTEN\n", maxCredits)
	} else {
		fmt.Printf("%d mahasiswa melebihi 24 SKS (maksimum %d) -> TIDAK KONSISTEN\n",
			violations, maxCredits)
	}
}

func printTally(tally map[string]int) {
	type pair struct {
		label string
		count int
	}
	pairs := make([]pair, 0, len(tally))
	for label, count := range tally {
		pairs = append(pairs, pair{label: label, count: count})
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].count != pairs[j].count {
			return pairs[i].count > pairs[j].count
		}
		return pairs[i].label < pairs[j].label
	})
	for _, p := range pairs {
		fmt.Printf("  %-32s %d\n", p.label, p.count)
	}
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
