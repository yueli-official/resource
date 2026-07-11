package assetclient

import (
	"context"
	"net/url"
	"strings"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"

	"platform/products/resource/api/internal/reserr"
)

// httpClient is the real AssetClient, talking to the asset service over HTTP.
// Upload/finalize/delete forward the caller's bearer token (operator == asset
// owner). The resource site is fully free, so every file is a public asset
// delivered from its stable CDN URL — no signed/gated delivery.
type httpClient struct {
	base     string
	siteSlug string
	spaceKey string
}

// NewHTTP builds an HTTP-backed asset client rooted at the asset service base URL.
func NewHTTP(baseURL, siteSlug, spaceKey string) Client {
	return &httpClient{base: strings.TrimRight(baseURL, "/"), siteSlug: siteSlug, spaceKey: spaceKey}
}

func (c *httpClient) post(ctx context.Context, bearer, path string, body g.Map) (*gjson.Json, error) {
	cli := g.Client()
	cli.SetHeader("Authorization", "Bearer "+bearer)
	cli.ContentJson()
	resp, err := cli.Post(ctx, c.base+path, body)
	if err != nil {
		return nil, reserr.UpstreamFailed("asset service unreachable")
	}
	defer resp.Close()
	j := gjson.New(resp.ReadAllString())
	if code := j.Get("code").String(); code != "ok" {
		return nil, reserr.UpstreamFailed(code)
	}
	return j, nil
}

func (c *httpClient) UploadInit(ctx context.Context, bearer string, in InitInput) (InitOutput, error) {
	j, err := c.post(ctx, bearer, "/api/v1/assets/upload-init", g.Map{
		"filename": in.Filename, "mime": in.Mime, "size": in.Size,
		"category": in.Category, "siteKey": c.siteSlug, "spaceKey": c.spaceKey, "profileKey": in.Category,
		"visibility": in.Visibility, "multipart": in.Multipart,
	})
	if err != nil {
		return InitOutput{}, err
	}
	return InitOutput{
		UploadURL:     j.Get("data.uploadUrl").String(),
		UploadToken:   j.Get("data.uploadToken").String(),
		Method:        j.Get("data.method").String(),
		UploadHeaders: stringMap(j.Get("data.uploadHeaders").Map()),
		UploadID:      j.Get("data.uploadId").String(),
		PartSize:      j.Get("data.partSize").Int64(),
		PartCount:     j.Get("data.partCount").Int(),
	}, nil
}

func (c *httpClient) MultipartPartURL(ctx context.Context, bearer string, in MultipartPartURLInput) (MultipartPartURLOutput, error) {
	j, err := c.post(ctx, bearer, "/api/v1/assets/multipart/part-url", g.Map{
		"uploadToken": in.UploadToken,
		"partNumber":  in.PartNumber,
	})
	if err != nil {
		return MultipartPartURLOutput{}, err
	}
	return MultipartPartURLOutput{
		UploadURL:     j.Get("data.uploadUrl").String(),
		UploadHeaders: stringMap(j.Get("data.uploadHeaders").Map()),
	}, nil
}

func (c *httpClient) CompleteMultipart(ctx context.Context, bearer string, in MultipartCompleteInput) error {
	parts := make([]g.Map, 0, len(in.Parts))
	for _, part := range in.Parts {
		parts = append(parts, g.Map{"partNumber": part.PartNumber, "etag": part.ETag})
	}
	_, err := c.post(ctx, bearer, "/api/v1/assets/multipart/complete", g.Map{
		"uploadToken": in.UploadToken,
		"parts":       parts,
	})
	return err
}

func (c *httpClient) AbortMultipart(ctx context.Context, bearer string, in MultipartAbortInput) error {
	_, err := c.post(ctx, bearer, "/api/v1/assets/multipart/abort", g.Map{
		"uploadToken": in.UploadToken,
	})
	return err
}

func stringMap(raw map[string]any) map[string]string {
	if len(raw) == 0 {
		return nil
	}
	out := make(map[string]string, len(raw))
	for k, v := range raw {
		out[k] = g.NewVar(v).String()
	}
	return out
}

func (c *httpClient) Finalize(ctx context.Context, bearer, uploadToken string) (View, error) {
	j, err := c.post(ctx, bearer, "/api/v1/assets/finalize", g.Map{"uploadToken": uploadToken})
	if err != nil {
		return View{}, err
	}
	return View{
		ID: j.Get("data.asset.id").String(), CdnURL: j.Get("data.asset.cdnUrl").String(),
		Size: j.Get("data.asset.size").Int64(), Mime: j.Get("data.asset.mime").String(),
		Filename: j.Get("data.asset.filename").String(), Visibility: j.Get("data.asset.visibility").String(),
	}, nil
}

func (c *httpClient) RegisterReference(ctx context.Context, bearer string, in ReferenceInput) error {
	_, err := c.post(ctx, bearer, "/api/v1/asset-references", g.Map{
		"assetId": in.AssetID, "siteKey": c.siteSlug, "refType": in.RefType, "refId": in.RefID,
		"refLabel": in.RefLabel, "refUrl": in.RefURL,
	})
	return err
}

func (c *httpClient) UnregisterReference(ctx context.Context, bearer string, in ReferenceInput) error {
	cli := g.Client()
	cli.SetHeader("Authorization", "Bearer "+bearer)
	q := url.Values{}
	q.Set("assetId", in.AssetID)
	q.Set("siteKey", c.siteSlug)
	q.Set("refType", in.RefType)
	q.Set("refId", in.RefID)
	resp, err := cli.Delete(ctx, c.base+"/api/v1/asset-references?"+q.Encode())
	if err != nil {
		return reserr.UpstreamFailed("asset service unreachable")
	}
	defer resp.Close()
	if code := gjson.New(resp.ReadAllString()).Get("code").String(); code != "ok" {
		return reserr.UpstreamFailed(code)
	}
	return nil
}

func (c *httpClient) Delete(ctx context.Context, bearer, assetID string) error {
	cli := g.Client()
	cli.SetHeader("Authorization", "Bearer "+bearer)
	resp, err := cli.Delete(ctx, c.base+"/api/v1/assets/"+assetID)
	if err != nil {
		return reserr.UpstreamFailed("asset service unreachable")
	}
	defer resp.Close()
	if code := gjson.New(resp.ReadAllString()).Get("code").String(); code != "ok" {
		return reserr.UpstreamFailed(code)
	}
	return nil
}
