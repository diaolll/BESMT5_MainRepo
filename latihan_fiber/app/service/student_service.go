package service

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"latihan_fiber/app/model"
	"latihan_fiber/app/repository"
	"latihan_fiber/helper"
)

type StudentService struct {
	repo  repository.StudentRepository
	perms *helper.PermissionSet
}

func NewStudentService(repo repository.StudentRepository, perms *helper.PermissionSet) *StudentService {
	return &StudentService{repo: repo, perms: perms}
}

func (s *StudentService) List(c *fiber.Ctx) error {
	format, err := helper.Negotiate(c, helper.FormatJSON, helper.FormatCSV)
	if err != nil {
		return err
	}
	q, err := helper.ParseCursorQuery(c)
	if err != nil {
		return err
	}
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	rows, err := s.repo.FindAfterCursor(ctx, q)
	if err != nil {
		return helper.Internal(err)
	}
	hasMore := len(rows) > q.Limit
	if hasMore {
		rows = rows[:q.Limit]
	}
	if format == helper.FormatCSV {
		return helper.WriteStudentsCSV(c, rows)
	}
	meta := &model.CursorMeta{Limit: q.Limit, HasMore: hasMore}
	if hasMore && len(rows) > 0 {
		last := rows[len(rows)-1]
		meta.NextCursor = helper.EncodeCursor(last.CreatedAt, last.ID)
	}
	return helper.SuccessCursor(c, "daftar student berhasil diambil", rows, meta)
}

func (s *StudentService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}
	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}
	student, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateStudentError(err)
	}
	ownerID := 0
	if student.OwnerID != nil {
		ownerID = *student.OwnerID
	}
	if !CanAccessStudent(current, ownerID, s.perms, "student:read:any") {
		return helper.Forbidden("tidak berhak mengakses data student lain")
	}
	return helper.Success(c, fiber.StatusOK, "student ditemukan", student)
}

func (s *StudentService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}
	var req model.CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	req.NIM = strings.TrimSpace(req.NIM)
	req.Name = strings.TrimSpace(req.Name)

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	ownerID := current.UserID
	newStudent, err := s.repo.Create(ctx, model.Student{
		NIM: req.NIM, Name: req.Name, Grade: req.Grade, IsActive: true,
		OwnerID: &ownerID,
	})
	if err != nil {
		return translateStudentError(err)
	}
	return helper.Created(c, "student berhasil dibuat", newStudent,
		"/api/v1/students/"+strconv.Itoa(newStudent.ID))
}

func (s *StudentService) Replace(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}
	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateStudentError(err)
	}
	ownerID := 0
	if existing.OwnerID != nil {
		ownerID = *existing.OwnerID
	}
	if !CanAccessStudent(current, ownerID, s.perms, "student:update:any") {
		return helper.Forbidden("tidak berhak mengubah data student lain")
	}
	var req model.ReplaceStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	result, err := s.repo.Update(ctx, model.Student{
		ID: id, NIM: strings.TrimSpace(req.NIM), Name: strings.TrimSpace(req.Name),
		Grade: req.Grade, IsActive: req.IsActive,
		OwnerID: existing.OwnerID,
	})
	if err != nil {
		return translateStudentError(err)
	}
	return helper.Success(c, fiber.StatusOK, "student berhasil diganti seluruhnya", result)
}

func (s *StudentService) Patch(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}
	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}
	var req model.PatchStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	if IsEmptyPatchStudent(req) {
		return helper.BadRequest("tidak ada field yang diubah")
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	currentData, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateStudentError(err)
	}
	ownerID := 0
	if currentData.OwnerID != nil {
		ownerID = *currentData.OwnerID
	}
	if !CanAccessStudent(current, ownerID, s.perms, "student:update:any") {
		return helper.Forbidden("tidak berhak mengubah data student lain")
	}
	updated, _ := ApplyPatchStudent(currentData, req)
	updated.OwnerID = currentData.OwnerID
	result, err := s.repo.Update(ctx, updated)
	if err != nil {
		return translateStudentError(err)
	}
	return helper.Success(c, fiber.StatusOK, "student berhasil diperbarui sebagian", result)
}

func (s *StudentService) Delete(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}
	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}
	// Cek ownership/permission sebelum hapus
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateStudentError(err)
	}
	ownerID := 0
	if existing.OwnerID != nil {
		ownerID = *existing.OwnerID
	}
	if !CanAccessStudent(current, ownerID, s.perms, "student:delete") {
		// fallback cek permission langsung jika owner check fails but user has delete:any? Per CanAccessStudent already handles
		return helper.Forbidden("tidak berhak menghapus data student lain")
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return translateStudentError(err)
	}
	return helper.NoContent(c)
}

func translateStudentError(err error) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.NotFound("student tidak ditemukan")
	case errors.Is(err, repository.ErrDuplicate):
		return helper.Conflict("nim sudah dipakai")
	default:
		return helper.Internal(err)
	}
}
