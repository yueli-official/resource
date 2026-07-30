package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"
)

func EnableEnvironmentConfig() error {
	if err := applyDatabaseURL(); err != nil {
		return err
	}
	config := g.Cfg()
	if _, enabled := config.GetAdapter().(*environmentAdapter); enabled {
		return nil
	}
	config.SetAdapter(&environmentAdapter{base: config.GetAdapter()})
	return nil
}

type environmentAdapter struct {
	base gcfg.Adapter
}

func (adapter *environmentAdapter) Available(ctx context.Context, resource ...string) bool {
	return adapter.base.Available(ctx, resource...)
}

func (adapter *environmentAdapter) Get(ctx context.Context, pattern string) (any, error) {
	if raw, found := os.LookupEnv(environmentKey(pattern)); found {
		return decodeEnvironmentValue(raw), nil
	}
	value, err := adapter.base.Get(ctx, pattern)
	if err != nil {
		return nil, err
	}
	return overlayEnvironment(pattern, value), nil
}

func (adapter *environmentAdapter) Data(ctx context.Context) (map[string]any, error) {
	data, err := adapter.base.Data(ctx)
	if err != nil {
		return nil, err
	}
	overlaid, _ := overlayEnvironment("", data).(map[string]any)
	return overlaid, nil
}

func (adapter *environmentAdapter) AddWatcher(name string, fn gcfg.WatcherFunc) {
	if watcher, ok := adapter.base.(gcfg.WatcherAdapter); ok {
		watcher.AddWatcher(name, fn)
	}
}

func (adapter *environmentAdapter) RemoveWatcher(name string) {
	if watcher, ok := adapter.base.(gcfg.WatcherAdapter); ok {
		watcher.RemoveWatcher(name)
	}
}

func (adapter *environmentAdapter) GetWatcherNames() []string {
	if watcher, ok := adapter.base.(gcfg.WatcherAdapter); ok {
		return watcher.GetWatcherNames()
	}
	return nil
}

func (adapter *environmentAdapter) IsWatching(name string) bool {
	if watcher, ok := adapter.base.(gcfg.WatcherAdapter); ok {
		return watcher.IsWatching(name)
	}
	return false
}

func overlayEnvironment(pattern string, value any) any {
	if pattern != "" {
		if raw, found := os.LookupEnv(environmentKey(pattern)); found {
			return decodeEnvironmentValue(raw)
		}
	}
	switch current := value.(type) {
	case map[string]any:
		copied := make(map[string]any, len(current))
		for key, child := range current {
			childPattern := key
			if pattern != "" {
				childPattern = pattern + "." + key
			}
			copied[key] = overlayEnvironment(childPattern, child)
		}
		return copied
	case []any:
		copied := make([]any, len(current))
		for index, child := range current {
			copied[index] = overlayEnvironment(pattern, child)
		}
		return copied
	default:
		return value
	}
}

func environmentKey(pattern string) string {
	return "GF_" + strings.ToUpper(strings.ReplaceAll(pattern, ".", "_"))
}

func decodeEnvironmentValue(raw string) any {
	var decoded any
	if json.Valid([]byte(raw)) && json.Unmarshal([]byte(raw), &decoded) == nil {
		return decoded
	}
	return raw
}

func applyDatabaseURL() error {
	raw := strings.TrimSpace(os.Getenv("RESOURCE_DATABASE_URL"))
	if raw == "" {
		return nil
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Hostname() == "" || parsed.User == nil ||
		(parsed.Scheme != "postgres" && parsed.Scheme != "postgresql") {
		return fmt.Errorf("RESOURCE_DATABASE_URL must be an absolute PostgreSQL URL")
	}
	name := strings.TrimPrefix(parsed.EscapedPath(), "/")
	name, err = url.PathUnescape(name)
	if err != nil || strings.TrimSpace(name) == "" || strings.Contains(name, "/") {
		return fmt.Errorf("RESOURCE_DATABASE_URL must contain one database name")
	}
	port := parsed.Port()
	if port == "" {
		port = "5432"
	}
	password, _ := parsed.User.Password()
	values := map[string]string{
		"GF_DATABASE_DEFAULT_TYPE":  "pgsql",
		"GF_DATABASE_DEFAULT_HOST":  parsed.Hostname(),
		"GF_DATABASE_DEFAULT_PORT":  port,
		"GF_DATABASE_DEFAULT_USER":  parsed.User.Username(),
		"GF_DATABASE_DEFAULT_PASS":  password,
		"GF_DATABASE_DEFAULT_NAME":  name,
		"GF_DATABASE_DEFAULT_EXTRA": parsed.RawQuery,
	}
	for key, value := range values {
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("set %s from RESOURCE_DATABASE_URL: %w", key, err)
		}
	}
	return nil
}
