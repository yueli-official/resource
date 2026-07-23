// Package reserr declares the resource-site error codes (namespace resource.*)
// and their HTTP status, registered with the shared gokit/errs catalog.
package reserr

import (
	"net/http"

	"platform/gokit/errs"
)

var (
	CodeNotFound             = errs.Register("resource.not_found", http.StatusNotFound)
	CodeForbidden            = errs.Register("resource.forbidden", http.StatusForbidden)
	CodeSlugTaken            = errs.Register("resource.slug_taken", http.StatusConflict)
	CodeInvalidType          = errs.Register("resource.invalid_type", http.StatusBadRequest)
	CodeInvalidState         = errs.Register("resource.invalid_state", http.StatusBadRequest)
	CodeInvalidInput         = errs.Register("resource.invalid_input", http.StatusBadRequest)
	CodeAssetNotFound        = errs.Register("resource.asset_not_found", http.StatusNotFound)
	CodeUpstreamFailed       = errs.Register("resource.upstream_failed", http.StatusBadGateway)
	CodeRevisionConflict     = errs.Register("resource.site_profile_revision_conflict", 412)
	CodePreconditionRequired = errs.Register("resource.site_profile_precondition_required", 428)
)

// NotFound is returned when a resource id/slug does not exist or is not visible.
func NotFound(id string) *errs.Coded {
	return errs.New(CodeNotFound, "resource not found", map[string]any{"id": id})
}

// Forbidden is returned when the caller is not the resource owner.
func Forbidden() *errs.Coded { return errs.New(CodeForbidden, "not the resource owner", nil) }

// SlugTaken is returned when a generated/explicit slug collides (after retries).
func SlugTaken(slug string) *errs.Coded {
	return errs.New(CodeSlugTaken, "slug already taken", map[string]any{"slug": slug})
}

// InvalidType is returned when type is not in the configured set.
func InvalidType(t string) *errs.Coded {
	return errs.New(CodeInvalidType, "unknown resource type", map[string]any{"type": t})
}

// InvalidState is returned when a publish constraint is unmet (no asset).
func InvalidState(detail string) *errs.Coded {
	return errs.New(CodeInvalidState, "invalid state: "+detail, nil)
}

// InvalidInput is returned for malformed request input not caught by binding.
func InvalidInput(detail string) *errs.Coded {
	return errs.New(CodeInvalidInput, "invalid input: "+detail, nil)
}

// AssetNotFound is returned when an assetId is not part of the resource.
func AssetNotFound(assetID string) *errs.Coded {
	return errs.New(CodeAssetNotFound, "asset not part of this resource", map[string]any{"assetId": assetID})
}

// UpstreamFailed wraps a downstream (asset) failure with a trimmed summary —
// never the raw downstream body.
func UpstreamFailed(summary string) *errs.Coded {
	return errs.New(CodeUpstreamFailed, "upstream service failed: "+summary, nil)
}

func RevisionConflict() *errs.Coded {
	return errs.New(CodeRevisionConflict, "site profile revision does not match", nil)
}

func PreconditionRequired() *errs.Coded {
	return errs.New(CodePreconditionRequired, "If-Match is required for site profile updates", nil)
}
