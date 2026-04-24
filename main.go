package main

import (
	"archive/tar"
	"compress/gzip"
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
	Address                string `help:"Address to listen on" default:"0.0.0.0"`
	Port                   int    `help:"Port to listen on" default:"8083"`
	Dist                   string `help:"Dist path" default:"./dist"`
	Data                   string `help:"Data path" default:"./data"`
	OIDCDiscoveryURL       string `help:"OIDC discovery URL (e.g. https://authentik.example.com/application/o/<slug>/.well-known/openid-configuration); omit to disable auth"`
	OIDCClientID           string `help:"OAuth2 client ID registered in Authentik"`
	OIDCInsecureSkipVerify bool   `help:"Skip TLS certificate verification for OIDC discovery and JWKS requests"`
	InfinityAddress        string `help:"Infinity embeddings server address" default:"http://localhost:8002"`
	InfinityTextModel      string `help:"Infinity text embeddings model ID" default:"wkcn/TinyCLIP-ViT-8M-16-Text-3M-YFCC15M"`
	InfinityImageModel     string `help:"Infinity image embeddings model ID" default:"openai/clip-vit-large-patch14"`
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {

	cli := humacli.New(func(hooks humacli.Hooks, options *Options) {

		log.Println("init backend")
		backend.SetInfinityConfig(options.InfinityAddress, options.InfinityTextModel, options.InfinityImageModel)

		if _, err := os.Stat(options.Data); os.IsNotExist(err) {
			err := os.Mkdir(options.Data, 0755)
			if err != nil {
				log.Fatalln(err)
			}
		}

		dbPath := filepath.Join(options.Data, "db.sqlite")
		dbExists := fileExists(dbPath)

		err := backend.ConnectDB(dbPath)
		if err != nil {
			log.Fatalln(err)
		}
		err = backend.InitAndMigrateDB()
		if err != nil {
			log.Fatalln(err)
		}

		if !dbExists {
			if err := runLegacyMigration(options.Data); err != nil {
				log.Printf("legacy migration failed: %v", err)
			}
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
			oidcCfg, err = backend.FetchOIDCConfig(options.OIDCDiscoveryURL, options.OIDCInsecureSkipVerify)
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
			api.UseMiddleware(backend.NewAuthMiddleware(api, oidcCfg.Issuer, oidcCfg.JWKSURI, options.OIDCInsecureSkipVerify))
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
			go backend.BackfillEmbeddings()
			err := http.ListenAndServe(fmt.Sprintf("%s:%d", options.Address, options.Port), router)
			if err != nil {
				log.Fatalln(err)
			}
		})

	})

	// Run the CLI. When passed no commands, it starts the server.
	cli.Run()

}

func runLegacyMigration(dataPath string) error {
	tarPath := filepath.Join(dataPath, "legacy.tar.gz")
	storeJSON := filepath.Join(dataPath, "store.json")

	switch {
	case fileExists(storeJSON):
		log.Println("legacy store.json found, running migration")
		if err := buildLegacyTarGz(dataPath, tarPath); err != nil {
			return fmt.Errorf("build tar.gz: %w", err)
		}
	case fileExists(tarPath):
		log.Println("legacy.tar.gz found, running migration")
	default:
		return nil
	}

	f, err := os.Open(tarPath)
	if err != nil {
		return fmt.Errorf("open tar.gz: %w", err)
	}
	defer f.Close()

	result, err := backend.ImportFromReader(context.Background(), f, false)
	if err != nil {
		return fmt.Errorf("import: %w", err)
	}

	log.Printf("legacy migration complete: %d entities, %d artifacts imported", result.EntitiesImported, result.ArtifactsImported)
	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func buildLegacyTarGz(dataPath, tarPath string) error {
	out, err := os.Create(tarPath)
	if err != nil {
		return err
	}
	defer out.Close()

	gw := gzip.NewWriter(out)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	addFile := func(name, arcName string) error {
		data, err := os.ReadFile(name)
		if err != nil {
			return err
		}
		hdr := &tar.Header{
			Name: arcName,
			Mode: 0644,
			Size: int64(len(data)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		_, err = tw.Write(data)
		return err
	}

	if err := addFile(filepath.Join(dataPath, "store.json"), "store.json"); err != nil {
		return fmt.Errorf("store.json: %w", err)
	}

	artifactsDir := filepath.Join(dataPath, "artifacts")
	entries, err := os.ReadDir(artifactsDir)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read artifacts dir: %w", err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".webp") {
			continue
		}
		src := filepath.Join(artifactsDir, e.Name())
		if err := addFile(src, "artifacts/"+e.Name()); err != nil {
			log.Printf("skipping artifact %s: %v", e.Name(), err)
		}
	}

	return nil
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
