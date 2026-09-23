package config

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"latihan_fiber/app/model"
	"latihan_fiber/helper"
	"latihan_fiber/middleware"
	"latihan_fiber/route"
)

func NewApp(
	logger *slog.Logger,
	deps route.Dependencies,
) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      GetEnv("APP_NAME", "Praktikum Backend Lanjut"),
		ErrorHandler: newErrorHandler(logger),
		BodyLimit:    1 * 1024 * 1024,
	})

	middleware.Register(app, logger, GetEnv("ALLOWED_ORIGINS", ""))
	route.Register(app, deps)

	app.Use(func(c *fiber.Ctx) error {
		return helper.NotFound("endpoint tidak ditemukan")
	})

	return app
}

func newErrorHandler(logger *slog.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		requestID := helper.RequestID(c)
		var appErr *helper.AppError
		switch {
		case errors.As(err, &appErr):
			// sudah benar, biarkan
		case errors.Is(err, fiber.ErrRequestEntityTooLarge):
			appErr = &helper.AppError{
				Status: fiber.StatusRequestEntityTooLarge,
				Code:   "PAYLOAD_TOO_LARGE",
				Message: "ukuran body melebihi batas yang diizinkan",
			}
		default:
			var fiberErr *fiber.Error
			if errors.As(err, &fiberErr) {
				appErr = &helper.AppError{
					Status: fiberErr.Code, Code: "HTTP_ERROR",
					Message: fiberErr.Message,
				}
			} else {
				appErr = helper.Internal(err)
			}
		}
		// FIX: sebelumnya level tertukar — 4xx harus WARN, 5xx ERROR
		if appErr.Status >= fiber.StatusInternalServerError {
			// 5xx = ERROR, sertakan cause jika ada
			causeMsg := ""
			if appErr.Cause() != nil {
				causeMsg = appErr.Cause().Error()
			} else {
				causeMsg = appErr.Error()
			}
			logger.Error("request_failed",
				slog.String("request_id", requestID),
				slog.String("path", c.Path()),
				slog.String("code", appErr.Code),
				slog.Int("status", appErr.Status),
				slog.String("error", causeMsg))
		} else {
			logger.Warn("request_rejected",
				slog.String("request_id", requestID),
				slog.String("path", c.Path()),
				slog.String("code", appErr.Code),
				slog.Int("status", appErr.Status))
		}
		return c.Status(appErr.Status).JSON(model.ErrorResponse{
			Success:   false,
			Code:      appErr.Code,
			Message:   appErr.Message,
			Fields:    appErr.Fields,
			RequestID: requestID,
		})
	}
}
