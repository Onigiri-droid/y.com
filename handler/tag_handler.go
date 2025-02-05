package handler

import (
	"api-service/model"
	"net/http"

	"github.com/labstack/echo/v4"
)

// Получение всех тегов
func (h *PostHandler) GetAllTags(c echo.Context) error {
	tags := make([]model.Tag, 0)
	err := h.DB.NewSelect().Model(&tags).Scan(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении тегов"})
	}

	return c.JSON(http.StatusOK, tags)
}