// Package assetclient is the resource site's gateway to the asset service.
// It is an interface so the catalog logic can be unit/integration-tested against a
// Fake without standing up a real asset service. Upload/finalize/delete forward
// the caller's user token (operator == asset owner). The resource site is fully
// free, so every file is a public asset delivered from its stable CDN URL.
package assetclient

import "context"

// InitInput begins an upload on the asset service.
type InitInput struct {
	Filename   string
	Mime       string
	Category   string
	Visibility string // public | private
	Size       int64
	Multipart  bool
}

// InitOutput is the asset service's signed upload short link the client PUTs to.
type InitOutput struct {
	UploadURL     string
	UploadToken   string
	Method        string
	UploadHeaders map[string]string
	UploadID      string
	PartSize      int64
	PartCount     int
}

// View is the asset metadata the resource site keeps a snapshot of.
type View struct {
	ID         string
	CdnURL     string // non-empty only for public assets
	Size       int64
	Mime       string
	Filename   string
	Visibility string
}

type ReferenceInput struct {
	AssetID, RefType, RefID string
	RefLabel, RefURL        string
}

type MultipartPartInput struct {
	PartNumber int
	ETag       string
}

type MultipartPartURLInput struct {
	UploadToken string
	PartNumber  int
}

type MultipartPartURLOutput struct {
	UploadURL     string
	UploadHeaders map[string]string
}

type MultipartCompleteInput struct {
	UploadToken string
	Parts       []MultipartPartInput
}

type MultipartAbortInput struct {
	UploadToken string
}

// Client is the asset-service contract the resource site depends on.
type Client interface {
	// UploadInit validates + reserves an upload, returning the blob short link.
	UploadInit(ctx context.Context, bearer string, in InitInput) (InitOutput, error)
	MultipartPartURL(ctx context.Context, bearer string, in MultipartPartURLInput) (MultipartPartURLOutput, error)
	CompleteMultipart(ctx context.Context, bearer string, in MultipartCompleteInput) error
	AbortMultipart(ctx context.Context, bearer string, in MultipartAbortInput) error
	// Finalize turns an uploaded blob into a recorded asset and returns its view.
	Finalize(ctx context.Context, bearer, uploadToken string) (View, error)
	RegisterReference(ctx context.Context, bearer string, in ReferenceInput) error
	UnregisterReference(ctx context.Context, bearer string, in ReferenceInput) error
	// Delete removes an asset (best-effort cleanup).
	Delete(ctx context.Context, bearer, assetID string) error
}
