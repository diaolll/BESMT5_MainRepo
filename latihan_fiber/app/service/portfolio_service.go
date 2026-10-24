package service

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"

	"latihan_fiber/app/model"
	"latihan_fiber/app/repository"
	"latihan_fiber/helper"
)

// PortfolioService adalah satu-satunya layer yang mengenal PostgreSQL dan
// MongoDB sekaligus. Data mahasiswa tetap diverifikasi ke PostgreSQL,
// sedangkan portofolio disimpan di MongoDB.
type PortfolioService struct {
	portfolios repository.PortfolioRepository
	students   repository.StudentRepository
	perms      *helper.PermissionSet
}

func NewPortfolioService(
	portfolios repository.PortfolioRepository,
	students repository.StudentRepository,
	perms *helper.PermissionSet,
) *PortfolioService {
	return &PortfolioService{portfolios: portfolios, students: students, perms: perms}
}

func (s *PortfolioService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}
	studentID, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id mahasiswa harus berupa angka positif")
	}

	student, err := s.students.FindByID(ctx, studentID)
	if err != nil {
		return translateStudentError(err)
	}
	if !CanAccessStudent(current, studentOwnerID(student), s.perms, "student:read:any") {
		return helper.Forbidden("tidak berhak mengakses portofolio mahasiswa lain")
	}

	portfolio, err := s.portfolios.GetByStudent(ctx, studentID)
	if err != nil {
		return translatePortfolioError(err)
	}
	return helper.Success(c, fiber.StatusOK, "portofolio berhasil diambil", portfolio)
}

func (s *PortfolioService) Replace(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}
	studentID, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id mahasiswa harus berupa angka positif")
	}

	student, err := s.students.FindByID(ctx, studentID)
	if err != nil {
		return translateStudentError(err)
	}
	if !CanAccessStudent(current, studentOwnerID(student), s.perms, "student:update:any") {
		return helper.Forbidden("tidak berhak mengubah portofolio mahasiswa lain")
	}

	var req model.ReplacePortfolioRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	normalizePortfolioRequest(&req)
	if errs := validatePortfolioRequest(req); errs != nil {
		return helper.Validation(errs)
	}

	portfolio, err := s.portfolios.Replace(ctx, model.Portfolio{
		StudentID: studentID, Headline: req.Headline, Bio: req.Bio,
		Projects:     projectsFromRequest(req.Projects),
		Certificates: certificatesFromRequest(req.Certificates),
	})
	if err != nil {
		return translatePortfolioError(err)
	}
	return helper.Success(c, fiber.StatusOK, "portofolio berhasil disimpan", portfolio)
}

func (s *PortfolioService) AddProject(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}
	studentID, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id mahasiswa harus berupa angka positif")
	}

	student, err := s.students.FindByID(ctx, studentID)
	if err != nil {
		return translateStudentError(err)
	}
	if !CanAccessStudent(current, studentOwnerID(student), s.perms, "student:update:any") {
		return helper.Forbidden("tidak berhak mengubah portofolio mahasiswa lain")
	}

	var req model.AddProjectRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	req.Title = strings.TrimSpace(req.Title)
	req.Link = strings.TrimSpace(req.Link)
	req.Technologies = normalizeTechnologies(req.Technologies)
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	portfolio, err := s.portfolios.AddProject(ctx, studentID, model.Project{
		Title: req.Title, Technologies: req.Technologies,
		Link: req.Link, Year: req.Year,
	})
	if err != nil {
		return translatePortfolioError(err)
	}
	location := fmt.Sprintf("/api/v1/students/%d/portfolio", studentID)
	return helper.Created(c, "proyek berhasil ditambahkan", portfolio, location)
}

func (s *PortfolioService) DeleteProject(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}
	studentID, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id mahasiswa harus berupa angka positif")
	}

	student, err := s.students.FindByID(ctx, studentID)
	if err != nil {
		return translateStudentError(err)
	}
	if !CanAccessStudent(current, studentOwnerID(student), s.perms, "student:update:any") {
		return helper.Forbidden("tidak berhak mengubah portofolio mahasiswa lain")
	}

	portfolio, err := s.portfolios.DeleteProject(ctx, studentID, c.Params("projectId"))
	if err != nil {
		return translatePortfolioError(err)
	}
	return helper.Success(c, fiber.StatusOK, "proyek berhasil dihapus", portfolio)
}

func (s *PortfolioService) Technologies(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	rows, err := s.portfolios.TopTechnologies(ctx, 10)
	if err != nil {
		return helper.Internal(err)
	}
	return helper.Success(c, fiber.StatusOK, "statistik teknologi berhasil diambil", rows)
}

func studentOwnerID(student model.Student) int {
	if student.OwnerID == nil {
		return 0
	}
	return *student.OwnerID
}

func normalizePortfolioRequest(req *model.ReplacePortfolioRequest) {
	req.Headline = strings.TrimSpace(req.Headline)
	req.Bio = strings.TrimSpace(req.Bio)
	for i := range req.Projects {
		req.Projects[i].Title = strings.TrimSpace(req.Projects[i].Title)
		req.Projects[i].Link = strings.TrimSpace(req.Projects[i].Link)
		req.Projects[i].Technologies = normalizeTechnologies(req.Projects[i].Technologies)
	}
	for i := range req.Certificates {
		req.Certificates[i].Name = strings.TrimSpace(req.Certificates[i].Name)
		req.Certificates[i].Issuer = strings.TrimSpace(req.Certificates[i].Issuer)
		req.Certificates[i].Date = strings.TrimSpace(req.Certificates[i].Date)
	}
}

func validatePortfolioRequest(req model.ReplacePortfolioRequest) map[string]string {
	errs := helper.ValidateStruct(req)
	if errs == nil {
		errs = map[string]string{}
	}
	for i, project := range req.Projects {
		for field, message := range helper.ValidateStruct(project) {
			errs[fmt.Sprintf("projects[%d].%s", i, field)] = message
		}
	}
	for i, certificate := range req.Certificates {
		for field, message := range helper.ValidateStruct(certificate) {
			errs[fmt.Sprintf("certificates[%d].%s", i, field)] = message
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errs
}

func projectsFromRequest(items []model.ProjectRequest) []model.Project {
	result := make([]model.Project, 0, len(items))
	for _, item := range items {
		result = append(result, model.Project{
			Title: item.Title, Technologies: item.Technologies,
			Link: item.Link, Year: item.Year,
		})
	}
	return result
}

func certificatesFromRequest(items []model.CertificateRequest) []model.Certificate {
	result := make([]model.Certificate, 0, len(items))
	for _, item := range items {
		result = append(result, model.Certificate{
			Name: item.Name, Issuer: item.Issuer, Date: item.Date,
		})
	}
	return result
}

// normalizeTechnologies merapikan nama teknologi tanpa mengubah kapitalisasi
// pertama yang dipilih pengguna. Duplikat dibandingkan tanpa membedakan huruf
// besar-kecil agar "Go" dan "go" tidak dihitung dua kali pada statistik.
func normalizeTechnologies(items []string) []string {
	result := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		key := strings.ToLower(item)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, item)
	}
	return result
}

func translatePortfolioError(err error) error {
	switch {
	case errors.Is(err, repository.ErrInvalidID):
		return helper.BadRequest("id proyek tidak valid")
	case errors.Is(err, repository.ErrNotFound):
		return helper.NotFound("portofolio belum dibuat")
	case errors.Is(err, repository.ErrProjectNotFound):
		return helper.NotFound("proyek tidak ditemukan")
	case errors.Is(err, repository.ErrProjectLimit):
		return helper.NewError(fiber.StatusConflict, "PROJECT_LIMIT",
			"jumlah proyek sudah mencapai batas 20")
	default:
		return helper.Internal(err)
	}
}
