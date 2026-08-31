# api-students

REST API sederhana untuk data mahasiswa. Dibuat sebagai Tugas Mandiri Pertemuan 2 —
REST API & HTTP Deep Dive (Praktikum Pemrograman Backend Lanjut).

## Menjalankan

```bash
go mod tidy
go run .
```

Server berjalan di `http://localhost:3000`.

## Kontrak API

| Metode | Endpoint | Parameter | Contoh Body | Status Mungkin | Contoh Respons |
|--------|----------|-----------|-------------|-----------------|----------------|
| GET | `/api/v1/students` | `page`, `limit`, `search`, `sort` (`id`\|`name`\|`grade`\|`created_at`), `order` (`asc`\|`desc`), `is_active`, `min_grade`, `max_grade` | - | 200 | lihat contoh #5 di bawah |
| GET | `/api/v1/students/:id` | - | - | 200, 400, 404 | lihat contoh #7 |
| POST | `/api/v1/students` | - | `{"nim":"434241074","name":"Diaul","grade":90}` | 201, 409, 415, 422 | lihat contoh #1 |
| PUT | `/api/v1/students/:id` | - | `{"nim":"434241074","name":"Diaul Revisi","grade":95,"is_active":false}` | 200, 400, 404, 422 | lihat contoh #9 |
| PATCH | `/api/v1/students/:id` | - | `{"grade":98}` | 200, 400, 404, 422 | lihat contoh #11 |
| DELETE | `/api/v1/students/:id` | - | - | 204, 400, 404 | lihat contoh #12 |

## Contoh Respons per Skenario Pengujian

### 1) 201 Created — buat mahasiswa pertama
```
HTTP/1.1 201 Created
Content-Type: application/json
Location: /api/v1/students/1
X-Request-Id: <uuid>

{
  "success": true,
  "message": "mahasiswa berhasil dibuat",
  "data": {
    "id": 1,
    "nim": "434241074",
    "name": "Diaul",
    "grade": 90,
    "is_active": true,
    "created_at": "2026-08-23T10:00:00+07:00"
  }
}
```

### 2) 409 Conflict — NIM duplikat
```
HTTP/1.1 409 Conflict

{
  "success": false,
  "message": "nim sudah terdaftar"
}
```

### 3) 422 Unprocessable Entity — validasi gagal
```
HTTP/1.1 422 Unprocessable Entity

{
  "success": false,
  "message": "validasi gagal",
  "errors": {
    "nim": "wajib diisi",
    "name": "wajib diisi",
    "grade": "harus di antara 0 dan 100"
  }
}
```

### 4) 415 Unsupported Media Type — tanpa Content-Type
```
HTTP/1.1 415 Unsupported Media Type

{
  "success": false,
  "message": "Content-Type harus application/json"
}
```

*(Setelah ini, 3 mahasiswa tambahan dibuat: Bagas (id 2, grade 75), Citra (id 3, grade 88), Dian (id 4, grade 60, kemudian di-PATCH `is_active=false`))*

### 5) 200 OK — daftar dengan paginasi & sort (`page=1&limit=2&sort=name&order=desc`)
```
HTTP/1.1 200 OK

{
  "success": true,
  "message": "daftar mahasiswa berhasil diambil",
  "data": [
    { "id": 1, "nim": "434241074", "name": "Diaul", "grade": 90, "is_active": true, "created_at": "..." },
    { "id": 4, "nim": "434241077", "name": "Dian", "grade": 60, "is_active": false, "created_at": "..." }
  ],
  "meta": { "page": 1, "limit": 2, "total": 4, "total_pages": 2 }
}
```

### 6) 200 OK — search & filter (`search=dia&is_active=true&min_grade=80`)
```
HTTP/1.1 200 OK

{
  "success": true,
  "message": "daftar mahasiswa berhasil diambil",
  "data": [
    { "id": 1, "nim": "434241074", "name": "Diaul", "grade": 90, "is_active": true, "created_at": "..." }
  ],
  "meta": { "page": 1, "limit": 10, "total": 1, "total_pages": 1 }
}
```
*(Dian tidak ikut kena filter `search=dia` karena `is_active`-nya sudah `false`.)*

### 7) 400 Bad Request — id bukan angka
```
HTTP/1.1 400 Bad Request

{
  "success": false,
  "message": "id harus berupa angka positif"
}
```

### 8) 404 Not Found — id tidak ada
```
HTTP/1.1 404 Not Found

{
  "success": false,
  "message": "mahasiswa tidak ditemukan"
}
```

### 9) 200 OK — PUT (ganti seluruh isi)
```
HTTP/1.1 200 OK

{
  "success": true,
  "message": "mahasiswa berhasil diganti seluruhnya",
  "data": {
    "id": 1,
    "nim": "434241074",
    "name": "Diaul Revisi",
    "grade": 95,
    "is_active": false,
    "created_at": "2026-08-23T10:00:00+07:00"
  }
}
```

### 10) 422 — PUT tanpa field wajib (`name` kosong)
```
HTTP/1.1 422 Unprocessable Entity

{
  "success": false,
  "message": "validasi gagal",
  "errors": {
    "name": "wajib diisi pada PUT"
  }
}
```

### 11) 200 OK — PATCH (ubah sebagian, hanya `grade`)
```
HTTP/1.1 200 OK

{
  "success": true,
  "message": "mahasiswa berhasil diperbarui sebagian",
  "data": {
    "id": 1,
    "nim": "434241074",
    "name": "Diaul Revisi",
    "grade": 98,
    "is_active": false,
    "created_at": "2026-08-23T10:00:00+07:00"
  }
}
```
*(Bandingkan dengan #9: hanya `grade` yang berubah dari 95 ke 98; `name` dan `is_active` tetap seperti hasil PUT sebelumnya — inilah bukti nyata bedanya PUT dan PATCH.)*

### 12) 204 No Content — DELETE
```
HTTP/1.1 204 No Content
```
*(tanpa body)*

### 13) 404 — DELETE id yang sudah dihapus
```
HTTP/1.1 404 Not Found

{
  "success": false,
  "message": "mahasiswa tidak ditemukan"
}
```

---

## ✅ Status Pengujian

Semua endpoint telah diuji menggunakan Insomnia dengan collection yang terorganisir.
Test 6 (data tambahan) dibuat via VSCode, bukan via Insomnia.

### Insomnia Collection: `Tugas 2 API Students`

```
📁 Tugas 2 API Students
  ├─ 📄 Test 1 - Health
  ├─ 📄 Test 2 - POST sukses
  ├─ 📄 Test 3 - NIM duplikat
  ├─ 📄 Test 4 - No Content-Type
  ├─ 📄 Test 5 - Validasi gagal
  ├─ 📄 Test 6 - Buat data tambahan   → (via VSCode)
  ├─ 📄 Test 7 - List + pagination
  ├─ 📄 Test 8 - Search + filter
  ├─ 📄 Test 9 - Get by id
  ├─ 📄 Test 10 - 404 & 400
  ├─ 📄 Test 11 - PUT
  ├─ 📄 Test 12 - PUT 422
  ├─ 📄 Test 13 - PATCH
  ├─ 📄 Test 14 - DELETE
  └─ 📄 Test 15 - PUT vs PATCH proof
```

---

## Sumber bantuan

- <Saya menggunakan claude AI untuk membantu menulis README ini. dan mengoreksi beberapa barisan code yang kurang efektif, lalu juga membantu saya dalam sesi brainstorming karena modul 2 ini cukup sulit bagi saya.>