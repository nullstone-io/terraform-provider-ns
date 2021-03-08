package ns

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
)

var (
	ApiKeyEnvVar   = "NULLSTONE_API_KEY"
	AddressEnvVar  = "NULLSTONE_ADDR"
	DefaultAddress = "https://api.nullstone.io"
	TraceEnvVar = "NULLSTONE_TRACE"
)

func NewConfig() Config {
	cfg := Config{
		BaseAddress: DefaultAddress,
		ApiKey:      os.Getenv(ApiKeyEnvVar),
	}
	if val := os.Getenv(AddressEnvVar); val != "" {
		cfg.BaseAddress = val
	}
	if val := os.Getenv(TraceEnvVar); val != "" {
		cfg.IsTraceEnabled = true
	}
	return cfg
}

type Config struct {
	BaseAddress    string
	ApiKey         string
	IsTraceEnabled bool
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
	bt := baseTransport
	if c.IsTraceEnabled {
		bt = &tracingTransport{BaseTransport: bt}
	}
	return &apiKeyTransport{BaseTransport: bt, ApiKey: c.ApiKey}
}

var _ http.RoundTripper = &tracingTransport{}

type tracingTransport struct {
	BaseTransport http.RoundTripper
}

func (t *tracingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	raw, _ := httputil.DumpRequestOut(r, true)
	log.Printf("[DEBUG] %s", string(raw))
	return t.BaseTransport.RoundTrip(r)
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
