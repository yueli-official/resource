package controller

import (
	"context"

	v1 "platform/products/resource/api/api/v1"
	"platform/products/resource/api/internal/assetclient"
	"platform/products/resource/api/internal/catalog"
)

// Assets handles the operator (JWT) file-management endpoints for a resource.
type Assets struct{ svc *catalog.Service }

func NewAssets(svc *catalog.Service) *Assets { return &Assets{svc: svc} }

func (c *Assets) AddAsset(ctx context.Context, req *v1.AddAssetReq) (*v1.AddAssetRes, error) {
	owner, err := subject(ctx)
	if err != nil {
		return nil, err
	}
	out, err := c.svc.AddAsset(ctx, owner, bearerOf(ctx), req.ID, req.Filename, req.Size, req.Multipart)
	if err != nil {
		return nil, err
	}
	return &v1.AddAssetRes{
		UploadURL:     out.UploadURL,
		UploadToken:   out.UploadToken,
		Method:        firstNonEmpty(out.Method, "PUT"),
		UploadHeaders: out.UploadHeaders,
		UploadID:      out.UploadID,
		PartSize:      out.PartSize,
		PartCount:     out.PartCount,
	}, nil
}

func (c *Assets) MultipartPartURL(ctx context.Context, req *v1.MultipartPartURLReq) (*v1.MultipartPartURLRes, error) {
	owner, err := subject(ctx)
	if err != nil {
		return nil, err
	}
	out, err := c.svc.MultipartPartURL(ctx, owner, bearerOf(ctx), req.ID, assetclient.MultipartPartURLInput{
		UploadToken: req.UploadToken,
		PartNumber:  req.PartNumber,
	})
	if err != nil {
		return nil, err
	}
	return &v1.MultipartPartURLRes{UploadURL: out.UploadURL, UploadHeaders: out.UploadHeaders}, nil
}

func (c *Assets) CompleteMultipart(ctx context.Context, req *v1.MultipartCompleteReq) (*v1.MultipartCompleteRes, error) {
	owner, err := subject(ctx)
	if err != nil {
		return nil, err
	}
	parts := make([]assetclient.MultipartPartInput, 0, len(req.Parts))
	for _, part := range req.Parts {
		parts = append(parts, assetclient.MultipartPartInput{PartNumber: part.PartNumber, ETag: part.ETag})
	}
	if err := c.svc.CompleteMultipart(ctx, owner, bearerOf(ctx), req.ID, assetclient.MultipartCompleteInput{
		UploadToken: req.UploadToken,
		Parts:       parts,
	}); err != nil {
		return nil, err
	}
	return &v1.MultipartCompleteRes{Completed: true}, nil
}

func (c *Assets) AbortMultipart(ctx context.Context, req *v1.MultipartAbortReq) (*v1.MultipartAbortRes, error) {
	owner, err := subject(ctx)
	if err != nil {
		return nil, err
	}
	if err := c.svc.AbortMultipart(ctx, owner, bearerOf(ctx), req.ID, assetclient.MultipartAbortInput{
		UploadToken: req.UploadToken,
	}); err != nil {
		return nil, err
	}
	return &v1.MultipartAbortRes{Aborted: true}, nil
}

func (c *Assets) FinalizeAsset(ctx context.Context, req *v1.FinalizeAssetReq) (*v1.FinalizeAssetRes, error) {
	owner, err := subject(ctx)
	if err != nil {
		return nil, err
	}
	ra, err := c.svc.FinalizeAsset(ctx, owner, bearerOf(ctx), req.ID, req.UploadToken, req.Label)
	if err != nil {
		return nil, err
	}
	return &v1.FinalizeAssetRes{Asset: assetView(ra)}, nil
}

func (c *Assets) RemoveAsset(ctx context.Context, req *v1.RemoveAssetReq) (*v1.RemoveAssetRes, error) {
	owner, err := subject(ctx)
	if err != nil {
		return nil, err
	}
	if err := c.svc.RemoveAsset(ctx, owner, bearerOf(ctx), req.ID, req.AssetID); err != nil {
		return nil, err
	}
	return &v1.RemoveAssetRes{Removed: true}, nil
}

func (c *Assets) AddCover(ctx context.Context, req *v1.AddCoverReq) (*v1.AddCoverRes, error) {
	owner, err := subject(ctx)
	if err != nil {
		return nil, err
	}
	out, err := c.svc.AddCover(ctx, owner, bearerOf(ctx), req.ID, req.Filename, req.Mime, req.Size)
	if err != nil {
		return nil, err
	}
	return &v1.AddCoverRes{UploadURL: out.UploadURL, UploadToken: out.UploadToken, Method: "PUT", UploadHeaders: out.UploadHeaders}, nil
}

func (c *Assets) FinalizeCover(ctx context.Context, req *v1.FinalizeCoverReq) (*v1.FinalizeCoverRes, error) {
	owner, err := subject(ctx)
	if err != nil {
		return nil, err
	}
	assetID, coverURL, err := c.svc.FinalizeCover(ctx, owner, bearerOf(ctx), req.ID, req.UploadToken)
	if err != nil {
		return nil, err
	}
	return &v1.FinalizeCoverRes{CoverAssetID: assetID, CoverURL: coverURL}, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
