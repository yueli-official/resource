package runtime

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/gogf/gf/v2/net/ghttp"
	foundationopenapi "github.com/yueli-official/foundation/go/goframe/openapi"
)

const openAPIOutputEnv = "RESOURCE_OPENAPI_OUTPUT"

func OpenAPIRequested() bool {
	return os.Getenv(openAPIOutputEnv) != ""
}

func ExportOpenAPIIfRequested(server *ghttp.Server) (handled bool, err error) {
	output := os.Getenv(openAPIOutputEnv)
	if output == "" {
		return false, nil
	}
	if err := foundationopenapi.Export(foundationopenapi.ExportConfig{
		Server: server, Output: output, Overwrite: true,
	}); err != nil {
		return true, err
	}
	return true, canonicalizeJSONFile(output)
}

func canonicalizeJSONFile(path string) error {
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var document any
	if err := json.Unmarshal(body, &document); err != nil {
		return fmt.Errorf("parse exported OpenAPI: %w", err)
	}
	canonical, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		return fmt.Errorf("canonicalize exported OpenAPI: %w", err)
	}
	canonical = append(canonical, '\n')
	if err := os.WriteFile(path, canonical, 0o644); err != nil {
		return fmt.Errorf("write canonical OpenAPI: %w", err)
	}
	return nil
}
