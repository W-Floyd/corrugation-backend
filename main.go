package main

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/W-Floyd/corrugation-backend/backend"
	"github.com/W-Floyd/corrugation-backend/frontend"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/danielgtaylor/huma/v2/autopatch"
	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/foolin/goview"
)

type Options struct {
	Port int `help:"Port to listen on" short:"p" default:"8001"`
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

//go:embed assets
var assets embed.FS

func main() {

	embedListing, err := assets.ReadDir("assets")
	if err != nil {
		log.Fatalln(err)
	}

	cli := humacli.New(func(hooks humacli.Hooks, options *Options) {

		// Create a new router & API
		router := http.NewServeMux()
		api := humago.New(router, huma.DefaultConfig("My API", "1.0.0"))

		autopatch.AutoPatch(api)

		c := goview.DefaultConfig
		c.DisableCache = true

		c.Funcs = template.FuncMap{
			"unescapeHTML": func(s string) any {
				return template.HTML(s)
			},
			"copy": func() string {
				return time.Now().Format("2006")
			},
			"componentButtonRound": frontend.ComponentButtonRound,
		}

		e := goview.New(c)

		for _, asset := range embedListing {
			contents, err := assets.ReadFile("assets/" + asset.Name())
			if err != nil {
				log.Fatalln(err)
			}
			var contentType string
			switch filepath.Ext(asset.Name()) {
			case ".js":
				contentType = "text/javascript"
			case ".css":
				contentType = "text/css"
			default:
				contentType = http.DetectContentType(contents)
			}

			huma.Register(api,
				huma.Operation{
					Path:   "/" + asset.Name(),
					Method: http.MethodGet,
					Hidden: true,
				}, func(ctx context.Context, i *struct{}) (output *backend.BytesOutput, err error) {
					output = &backend.BytesOutput{
						Body:        contents,
						ContentType: contentType,
					}
					return
				})
		}

		huma.Register(api, huma.Operation{
			Path:   "/{$}",
			Method: http.MethodGet,
			Hidden: true,
		}, func(ctx context.Context, input *struct {
		}) (output *backend.BytesOutput, err error) {

			v := NewCustomResponseWriter()

			err = e.Render(v, http.StatusOK, "index", goview.M{
				"title": "Corrugation",
			})

			output = &backend.BytesOutput{
				Body:        v.body,
				ContentType: "text/html",
			}

			return
		})

		backend.RegisterHandlers(api)

		// Tell the CLI how to start your router.
		hooks.OnStart(func() {
			http.ListenAndServe(fmt.Sprintf(":%d", options.Port), router)
		})

	})

	// Run the CLI. When passed no commands, it starts the server.
	cli.Run()

}

type CustomResponseWriter struct {
	body       []byte
	statusCode int
	header     http.Header
}

func NewCustomResponseWriter() *CustomResponseWriter {
	return &CustomResponseWriter{
		header: http.Header{},
	}
}

func (w *CustomResponseWriter) Header() http.Header {
	return w.header
}

func (w *CustomResponseWriter) Write(b []byte) (int, error) {
	w.body = append(w.body, b...)
	return 0, nil
}

func (w *CustomResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}
