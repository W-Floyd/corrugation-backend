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
)

type Options struct {
	Address          string `help:"Address to listen on" default:"0.0.0.0"`
	Port             int    `help:"Port to listen on" default:"8083"`
	Dist             string `help:"Dist path" default:"./dist"`
	Data             string `help:"Data path" default:"./data"`
	OIDCDiscoveryURL string `help:"OIDC discovery URL (e.g. https://authentik.example.com/application/o/<slug>/.well-known/openid-configuration); omit to disable auth"`
	OIDCClientID     string `help:"OAuth2 client ID registered in Authentik"`
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
		config := huma.DefaultConfig("My API", "1.0.0")

		var oidcCfg *backend.OIDCConfig
		if options.OIDCDiscoveryURL != "" {
			var err error
			oidcCfg, err = backend.FetchOIDCConfig(options.OIDCDiscoveryURL)
			if err != nil {
				log.Fatalf("fetch OIDC config: %v", err)
			}
			config.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
				"authentik": {
					Type:         "http",
					Scheme:       "bearer",
					BearerFormat: "JWT",
				},
			}
			backend.SetAuthConfig(backend.AuthFrontendConfig{
				Enabled:               true,
				AuthorizationEndpoint: oidcCfg.AuthorizationEndpoint,
				TokenEndpoint:         oidcCfg.TokenEndpoint,
				ClientID:              options.OIDCClientID,
			})
			log.Printf("OIDC auth enabled, issuer: %s", oidcCfg.Issuer)
		}

		api := humago.New(router, config)
		backend.RegisterAuthHandlers(api)

		if oidcCfg != nil {
			api.UseMiddleware(backend.NewAuthMiddleware(api, oidcCfg.Issuer, oidcCfg.JWKSURI))
		}

		autopatch.AutoPatch(api)

		var indexHTML []byte

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
				indexHTML = contents
				continue
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

		// Catch-all: serve index.html for any path not matched by API or assets.
		if indexHTML != nil {
			router.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.Write(indexHTML)
			})
		}

		backend.RegisterHandlers(api)
		router.HandleFunc("GET /ws", backend.WsHandler)

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
