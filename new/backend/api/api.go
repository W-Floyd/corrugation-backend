package api

import (
	"net/http"

	"github.com/W-Floyd/corrugation/backend/variables"
	"github.com/labstack/echo/v4"
)

func Version(c echo.Context) error {
	return c.JSON(http.StatusOK, variables.Version)
}
