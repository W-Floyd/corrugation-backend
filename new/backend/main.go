package main

import (
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/W-Floyd/corrugation/backend/api"
	"github.com/W-Floyd/corrugation/backend/variables"
	"github.com/foolin/goview"
	"github.com/foolin/goview/supports/echoview-v4"
	"github.com/kelseyhightower/envconfig"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/gorm"
)

var (
	config Config
	DB     *gorm.DB
)

func main() {

	err := envconfig.Process(variables.Appname, &config)
	if err != nil {
		log.Fatal(err)
	}

	err = InitDB(&config)
	if err != nil {
		log.Fatal(err)
	}

	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	c := goview.DefaultConfig
	c.DisableCache = true
	c.Root = "frontend/views"

	c.Funcs = template.FuncMap{
		"unescapeHTML": unescapeHTML,
		"copyright":    copyright,
	}

	e.Renderer = echoview.New(c)

	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "index", echo.Map{
			"title": variables.Appname,
		})
	})

	e.Use(middleware.Static("frontend/assets"))

	r := e.Group("/api")
	r.GET("/version", api.Version)

	e.Logger.Fatal(e.Start(":" + strconv.Itoa(config.Port)))
}
