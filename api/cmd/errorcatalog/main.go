package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/yueli-official/resource/api/internal/reserr"
)

type document struct {
	SchemaVersion string                `json:"schemaVersion"`
	Errors        []reserr.CatalogEntry `json:"errors"`
}

func main() {
	output := flag.String("output", filepath.FromSlash("contracts/errors/catalog.json"), "错误目录输出路径")
	check := flag.Bool("check", false, "只检查已提交目录是否最新")
	flag.Parse()
	data, err := json.MarshalIndent(document{
		SchemaVersion: "resource.yueli.dev/error-catalog/v1",
		Errors:        reserr.Catalog(),
	}, "", "  ")
	if err != nil {
		exit(err)
	}
	data = append(data, '\n')
	if *check {
		current, err := os.ReadFile(*output)
		if err != nil {
			exit(err)
		}
		if !bytes.Equal(bytes.ReplaceAll(current, []byte("\r\n"), []byte("\n")), data) {
			exit(fmt.Errorf("错误目录已漂移，请执行 go run ./cmd/errorcatalog"))
		}
		return
	}
	if err := os.MkdirAll(filepath.Dir(*output), 0o755); err != nil {
		exit(err)
	}
	if err := os.WriteFile(*output, data, 0o644); err != nil {
		exit(err)
	}
}

func exit(err error) {
	fmt.Fprintln(os.Stderr, "errorcatalog:", err)
	os.Exit(1)
}
