package helper

import (
	"encoding/base64"
	"errors"
	"strconv"
	"strings"
	"time"

	"latihan_fiber/app/model"

	"github.com/gofiber/fiber/v2"
)

var ErrInvalidCursor = errors.New("cursor tidak valid")

func EncodeCursor(createdAt time.Time, id int) string {
	raw := strconv.FormatInt(createdAt.UTC().UnixNano(), 10) + "|" + strconv.Itoa(id)
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func DecodeCursor(encoded string) (model.Cursor, error) {
	if strings.TrimSpace(encoded) == "" {
		return model.Cursor{}, ErrInvalidCursor
	}
	decoded, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return model.Cursor{}, ErrInvalidCursor
	}
	parts := strings.Split(string(decoded), "|")
	if len(parts) != 2 {
		return model.Cursor{}, ErrInvalidCursor
	}
	nanos, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return model.Cursor{}, ErrInvalidCursor
	}
	id, err := strconv.Atoi(parts[1])
	if err != nil || id < 1 {
		return model.Cursor{}, ErrInvalidCursor
	}
	return model.Cursor{CreatedAt: time.Unix(0, nanos).UTC(), ID: id}, nil
}

func ParseCursorQuery(c *fiber.Ctx) (model.CursorQuery, error) {
	q := model.CursorQuery{
		Limit:  c.QueryInt("limit", 10),
		Search: strings.TrimSpace(c.Query("search")),
	}
	if q.Limit < 1 {
		q.Limit = 10
	}
	if q.Limit > 100 {
		q.Limit = 100
	}
	if raw := c.Query("is_active"); raw != "" {
		if v, err := strconv.ParseBool(raw); err == nil {
			q.IsActive = &v
		} else {
			return model.CursorQuery{}, BadRequest("is_active harus berupa boolean")
		}
	}
	// grade filters for students reuse same helper
	if raw := c.Query("min_grade"); raw != "" {
		if v, err := strconv.ParseFloat(raw, 64); err == nil {
			q.MinGrade = &v
		} else {
			return model.CursorQuery{}, BadRequest("min_grade harus berupa angka")
		}
	}
	if raw := c.Query("max_grade"); raw != "" {
		if v, err := strconv.ParseFloat(raw, 64); err == nil {
			q.MaxGrade = &v
		} else {
			return model.CursorQuery{}, BadRequest("max_grade harus berupa angka")
		}
	}
	if raw := strings.TrimSpace(c.Query("cursor")); raw != "" {
		cur, err := DecodeCursor(raw)
		if err != nil {
			return model.CursorQuery{}, BadRequest("cursor tidak valid")
		}
		q.After = &cur
	}
	return q, nil
}
