package oldbackend

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func handleLogin(c echo.Context) error {
	var req loginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, "invalid request")
	}

	if req.Username != viper.GetString("username") || req.Password != viper.GetString("password") {
		return c.JSON(http.StatusUnauthorized, "invalid credentials")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": req.Username,
		"exp": time.Now().Add(72 * time.Hour).Unix(),
	})

	signed, err := token.SignedString([]byte(viper.GetString("jwt-secret")))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "failed to sign token")
	}

	return c.JSON(http.StatusOK, map[string]string{"token": signed})
}

func jwtMiddleware() echo.MiddlewareFunc {
	return echojwt.WithConfig(echojwt.Config{
		SigningKey: []byte(viper.GetString("jwt-secret")),
	})
}
