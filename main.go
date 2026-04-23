package main

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/W-Floyd/corrugation/backend"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/danielgtaylor/huma/v2/autopatch"
	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/foolin/goview"
)

type Options struct {
	Address string `help:"Address to listen on" default:"0.0.0.0"`
	Port    int    `help:"Port to listen on" default:"8001"`
	Dist    string `help:"Dist path" default:"./dist"`
	Data    string `help:"Data path" default:"./data"`
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {

	cli := humacli.New(func(hooks humacli.Hooks, options *Options) {

		log.Println("init backend")
		if _, err := os.Stat(options.Data); os.IsNotExist(err) {
			err := os.Mkdir(options.Data, 0755)
			if err != nil {
				log.Fatalln(err)
			}
		}
		err := backend.ConnectDB(filepath.Join(options.Data, "db.sqlite"))
		if err != nil {
			log.Fatalln(err)
		}
		err = backend.InitAndMigrateDB()
		if err != nil {
			log.Fatalln(err)
		}

		assets := []string{}

		err = filepath.WalkDir(options.Dist, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() {
				assets = append(assets, path)
			}
			return nil
		})

		if err != nil {
			log.Fatalln(err)
		}

		// Create a new router & API
		router := http.NewServeMux()
		api := humago.New(router, huma.DefaultConfig("My API", "1.0.0"))

		autopatch.AutoPatch(api)

		c := goview.DefaultConfig
		c.DisableCache = true

		for _, asset := range assets {
			contents, err := os.ReadFile(asset)
			if err != nil {
				log.Fatalln(err)
			}
			var contentType string
			switch filepath.Ext(asset) {
			case ".js":
				contentType = "text/javascript"
			case ".css":
				contentType = "text/css"
			default:
				contentType = http.DetectContentType(contents)
			}

			path := strings.TrimPrefix(filepath.Clean(asset), filepath.Clean(options.Dist))

			if path == "/index.html" {
				path = "/{$}"
			}

			huma.Register(api,
				huma.Operation{
					Path:   path,
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

		backend.RegisterHandlers(api)

		// Tell the CLI how to start your router.
		hooks.OnStart(func() {
			err := http.ListenAndServe(fmt.Sprintf("%s:%d", options.Address, options.Port), router)
			if err != nil {
				log.Fatalln(err)
			}
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
