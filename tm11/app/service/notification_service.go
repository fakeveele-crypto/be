package service

import (
	"github.com/gofiber/fiber/v2"
	"latihan-fiber/tm11/app/repository"
	"latihan-fiber/tm11/helper"
)

type NotificationService struct {
	notifications repository.NotificationRepository
}

func NewNotificationService(notifications repository.NotificationRepository) *NotificationService {
	return &NotificationService{notifications: notifications}
}

func (s *NotificationService) Mine(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	notifications, err := s.notifications.FindByUser(ctx, current.UserID, 50)
	if err != nil {
		return helper.Internal(err)
	}
	return helper.Success(c, fiber.StatusOK, "notifikasi berhasil diambil", notifications)
}
