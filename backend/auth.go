package backend

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

type OIDCConfig struct {
	Issuer                string `json:"issuer"`
	JWKSURI               string `json:"jwks_uri"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
}

func FetchOIDCConfig(discoveryURL string) (*OIDCConfig, error) {
	resp, err := http.Get(discoveryURL) //nolint:gosec
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var cfg OIDCConfig
	if err := json.NewDecoder(resp.Body).Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// AuthFrontendConfig is returned to the frontend so it can initiate the OIDC flow.
type AuthFrontendConfig struct {
	Enabled               bool   `json:"enabled"`
	AuthorizationEndpoint string `json:"authorizationEndpoint,omitempty"`
	TokenEndpoint         string `json:"tokenEndpoint,omitempty"`
	ClientID              string `json:"clientId,omitempty"`
}

var globalAuthConfig = AuthFrontendConfig{}

func SetAuthConfig(cfg AuthFrontendConfig) {
	globalAuthConfig = cfg
}

var GetAuthConfigOperation = huma.Operation{
	Method:      http.MethodGet,
	Path:        "/api/auth/config",
	OperationID: "get-auth-config",
	Summary:     "Get OIDC auth configuration for the frontend",
}

func GetAuthConfigHandler(_ context.Context, _ *struct{}) (*struct{ Body AuthFrontendConfig }, error) {
	return &struct{ Body AuthFrontendConfig }{Body: globalAuthConfig}, nil
}

func RegisterAuthHandlers(api huma.API) {
	huma.Register(api, GetAuthConfigOperation, GetAuthConfigHandler)
}

func newJWKSet(jwksURL string) jwk.Set {
	cache := jwk.NewCache(context.Background())
	if err := cache.Register(jwksURL, jwk.WithMinRefreshInterval(10*time.Minute)); err != nil {
		panic("failed to register jwk location: " + err.Error())
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if _, err := cache.Refresh(ctx, jwksURL); err != nil {
		panic("failed to fetch jwks on startup: " + err.Error())
	}
	return jwk.NewCachedSet(cache, jwksURL)
}

// NewAuthMiddleware guards all /api/ paths (except /api/auth/) with OIDC JWT validation.
func NewAuthMiddleware(api huma.API, issuer, jwksURL string) func(huma.Context, func(huma.Context)) {
	keySet := newJWKSet(jwksURL)
	return func(ctx huma.Context, next func(huma.Context)) {
		path := ctx.URL().Path
		if !strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/api/auth/") {
			next(ctx)
			return
		}
		token := strings.TrimPrefix(ctx.Header("Authorization"), "Bearer ")
		if token == "" {
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "Unauthorized")
			return
		}
		_, err := jwt.ParseString(token,
			jwt.WithKeySet(keySet),
			jwt.WithValidate(true),
			jwt.WithIssuer(issuer),
		)
		if err != nil {
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "Unauthorized")
			return
		}
		next(ctx)
	}
}
