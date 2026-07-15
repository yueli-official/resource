package assetclient

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestHTTPClientUsesConfiguredSiteContextForAssetWrites(t *testing.T) {
	var bodies []map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			bodies = append(bodies, map[string]any{"siteKey": r.URL.Query().Get("siteKey")})
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"code":"ok","data":{}}`))
			return
		}
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		var body map[string]any
		if err := json.Unmarshal(raw, &body); err != nil {
			values, parseErr := url.ParseQuery(string(raw))
			if parseErr != nil {
				t.Errorf("decode request: json=%v form=%v", err, parseErr)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			body = make(map[string]any, len(values))
			for key := range values {
				body[key] = values.Get(key)
			}
		}
		bodies = append(bodies, body)
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/api/v1/assets/upload-init" {
			_, _ = w.Write([]byte(`{"code":"ok","data":{"uploadUrl":"http://upload.test","uploadToken":"token"}}`))
			return
		}
		_, _ = w.Write([]byte(`{"code":"ok","data":{}}`))
	}))
	defer server.Close()

	client := NewHTTP(server.URL, "resource-main", "yueli")
	if _, err := client.UploadInit(context.Background(), "token", InitInput{
		Filename: "tool.zip", Mime: "application/zip", Category: "resource", Visibility: "public", Size: 12,
	}); err != nil {
		t.Fatalf("UploadInit() error = %v", err)
	}
	if err := client.RegisterReference(context.Background(), "token", ReferenceInput{
		AssetID: "asset-1", RefType: "resource-file", RefID: "resource-1",
	}); err != nil {
		t.Fatalf("RegisterReference() error = %v", err)
	}
	if err := client.UnregisterReference(context.Background(), "token", ReferenceInput{
		AssetID: "asset-1", RefType: "resource-file", RefID: "resource-1",
	}); err != nil {
		t.Fatalf("UnregisterReference() error = %v", err)
	}

	if len(bodies) != 3 {
		t.Fatalf("request count = %d, want 3", len(bodies))
	}
	for i, body := range bodies {
		if body["siteKey"] != "resource-main" {
			t.Fatalf("request %d siteKey = %#v, want resource-main", i, body["siteKey"])
		}
	}
	if bodies[0]["spaceKey"] != "yueli" {
		t.Fatalf("upload spaceKey = %#v, want yueli", bodies[0]["spaceKey"])
	}
}
