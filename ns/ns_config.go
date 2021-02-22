package ns

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
)

var (
	ApiKeyEnvVar   = "NULLSTONE_API_KEY"
	AddressEnvVar  = "NULLSTONE_ADDR"
	DefaultAddress = "https://api.nullstone.io"
)

func NewConfig() Config {
	cfg := Config{
		BaseAddress: DefaultAddress,
		ApiKey:      os.Getenv(ApiKeyEnvVar),
	}
	if val := os.Getenv(AddressEnvVar); val != "" {
		cfg.BaseAddress = val
	}
	return cfg
}

type Config struct {
	BaseAddress string
	ApiKey      string
}

func (c *Config) ConstructUrl(reqPath string) (*url.URL, error) {
	u, err := url.Parse(c.BaseAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid nullstone API base address (%s): %w", c.BaseAddress, err)
	}
	u.Path = path.Join(u.Path, reqPath)
	return u, nil
}

func (c *Config) CreateTransport(baseTransport http.RoundTripper) http.RoundTripper {
	return &apiKeyTransport{BaseTransport: baseTransport, ApiKey: c.ApiKey}
}

var _ http.RoundTripper = &apiKeyTransport{}

type apiKeyTransport struct {
	BaseTransport http.RoundTripper
	ApiKey        string
}

func (t *apiKeyTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Set("Authorization", "Bearer "+t.ApiKey)
	return t.BaseTransport.RoundTrip(r)
}
