// Package deploy owns Resource's production binding checks.
package deploy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	jose "github.com/go-jose/go-jose/v4"
)

const maxResponseBytes = int64(1 << 20)

var requiredAssetPaths = []string{
	"/api/v1/asset-references",
	"/api/v1/assets/finalize",
	"/api/v1/assets/upload-init",
}

type BindingConfig struct {
	IdentityIssuer            string
	IdentityDiscoveryURL      string
	IdentityJWKSURL           string
	AssetBaseURL              string
	AllowInsecureIdentityHTTP bool
	AllowInsecureAssetHTTP    bool
}

func WaitForBindings(ctx context.Context, config BindingConfig, interval time.Duration) error {
	if interval <= 0 {
		interval = 2 * time.Second
	}
	var lastErr error
	for {
		if err := checkBindings(ctx, config); err == nil {
			return nil
		} else {
			lastErr = err
		}
		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return fmt.Errorf("bindings did not converge: %w", errors.Join(lastErr, ctx.Err()))
		case <-timer.C:
		}
	}
}

func checkBindings(ctx context.Context, config BindingConfig) error {
	config.IdentityIssuer = strings.TrimRight(strings.TrimSpace(config.IdentityIssuer), "/")
	config.IdentityDiscoveryURL = strings.TrimSpace(config.IdentityDiscoveryURL)
	config.IdentityJWKSURL = strings.TrimSpace(config.IdentityJWKSURL)
	config.AssetBaseURL = strings.TrimRight(strings.TrimSpace(config.AssetBaseURL), "/")
	if config.IdentityDiscoveryURL == "" && config.IdentityIssuer != "" {
		config.IdentityDiscoveryURL = config.IdentityIssuer + "/.well-known/openid-configuration"
	}
	if config.IdentityJWKSURL == "" && config.IdentityIssuer != "" {
		config.IdentityJWKSURL = config.IdentityIssuer + "/oauth2/jwks.json"
	}
	for name, candidate := range map[string]struct {
		value         string
		allowInsecure bool
	}{
		"identity issuer":        {config.IdentityIssuer, config.AllowInsecureIdentityHTTP},
		"identity discovery URL": {config.IdentityDiscoveryURL, config.AllowInsecureIdentityHTTP},
		"identity JWKS URL":      {config.IdentityJWKSURL, config.AllowInsecureIdentityHTTP},
		"asset base URL":         {config.AssetBaseURL, config.AllowInsecureAssetHTTP},
	} {
		if candidate.value == "" {
			return fmt.Errorf("%s is required", name)
		}
		if err := validateURL(candidate.value, candidate.allowInsecure); err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
	}
	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	var discovery struct {
		Issuer                string `json:"issuer"`
		AuthorizationEndpoint string `json:"authorization_endpoint"`
		TokenEndpoint         string `json:"token_endpoint"`
		JWKSURI               string `json:"jwks_uri"`
	}
	if err := getJSON(ctx, client, config.IdentityDiscoveryURL, &discovery); err != nil {
		return fmt.Errorf("identity discovery: %w", err)
	}
	if strings.TrimRight(discovery.Issuer, "/") != config.IdentityIssuer {
		return fmt.Errorf("identity issuer %q does not match %q", discovery.Issuer, config.IdentityIssuer)
	}
	if discovery.AuthorizationEndpoint == "" || discovery.TokenEndpoint == "" || discovery.JWKSURI == "" {
		return errors.New("identity discovery is missing OIDC endpoints")
	}
	var set jose.JSONWebKeySet
	if err := getJSON(ctx, client, config.IdentityJWKSURL, &set); err != nil {
		return fmt.Errorf("identity JWKS: %w", err)
	}
	usableKey := false
	for index := range set.Keys {
		key := &set.Keys[index]
		if key.Valid() && key.IsPublic() && strings.TrimSpace(key.KeyID) != "" &&
			(key.Use == "" || key.Use == "sig") {
			usableKey = true
			break
		}
	}
	if !usableKey {
		return errors.New("identity JWKS has no usable public signing key")
	}
	if err := getJSON(ctx, client, config.AssetBaseURL+"/readyz", &map[string]any{}); err != nil {
		return fmt.Errorf("asset readiness: %w", err)
	}
	var document struct {
		Paths map[string]json.RawMessage `json:"paths"`
	}
	if err := getJSON(ctx, client, config.AssetBaseURL+"/api.json", &document); err != nil {
		return fmt.Errorf("asset OpenAPI: %w", err)
	}
	for _, path := range requiredAssetPaths {
		if _, ok := document.Paths[path]; !ok {
			return fmt.Errorf("asset contract is missing required path %s", path)
		}
	}
	return nil
}

func validateURL(value string, allowInsecureHTTP bool) error {
	parsed, err := url.Parse(value)
	if err != nil || !parsed.IsAbs() || parsed.Host == "" {
		return errors.New("must be an absolute URL")
	}
	if parsed.User != nil || parsed.Fragment != "" {
		return errors.New("must not contain credentials or a fragment")
	}
	switch strings.ToLower(parsed.Scheme) {
	case "https":
		return nil
	case "http":
		if allowInsecureHTTP || isLoopbackHost(parsed.Hostname()) {
			return nil
		}
		return errors.New("plain HTTP requires an explicit insecure-HTTP opt-in")
	default:
		return errors.New("scheme must be HTTP or HTTPS")
	}
}

func isLoopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	address := net.ParseIP(host)
	return address != nil && address.IsLoopback()
}

func getJSON(ctx context.Context, client *http.Client, endpoint string, target any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	request.Header.Set("Accept", "application/json")
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, maxResponseBytes+1))
	if err != nil {
		return err
	}
	if int64(len(body)) > maxResponseBytes {
		return fmt.Errorf("response exceeds %d bytes", maxResponseBytes)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("GET %s returned %s", endpoint, response.Status)
	}
	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("decode GET %s: %w", endpoint, err)
	}
	return nil
}
