package cmd

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/skip2/go-qrcode"
)

func qrGenerate(c echo.Context) error {

	id := c.Param("id")

	idInt, err := strconv.Atoi(id)

	if err != nil {
		return c.String(http.StatusBadRequest, "id "+id+" could not be parsed as an integer")
	}

	recLevel := 0

	if hasForm("level", c) {
		recLevelS := c.Param("level")

		recLevel, err = strconv.Atoi(recLevelS)

		if err != nil {
			return c.String(http.StatusBadRequest, "level "+id+" could not be parsed as an integer")
		}

	}

	code, err := qrcode.Encode(strconv.Itoa(idInt), qrcode.RecoveryLevel(recLevel), 1024)

	if err != nil {
		return c.String(http.StatusInternalServerError, "Error generating QR code")
	}

	return c.Blob(http.StatusOK, "image/png", code)

}
