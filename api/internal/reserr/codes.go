// Package reserr declares Resource's immutable public Problem contract.
package reserr

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/yueli-official/foundation/go/problem"
)

const (
	CodeNotFound                 = "resource.not_found"
	CodeForbidden                = "resource.forbidden"
	CodeSlugTaken                = "resource.slug_taken"
	CodeInvalidType              = "resource.invalid_type"
	CodeInvalidState             = "resource.invalid_state"
	CodeInvalidInput             = "resource.invalid_input"
	CodeAssetNotFound            = "resource.asset_not_found"
	CodeUpstreamFailed           = "resource.upstream_failed"
	CodeRevisionConflict         = "resource.site_profile_revision_conflict"
	CodePreconditionRequired     = "resource.site_profile_precondition_required"
	CodeAuthorizationUnavailable = "resource.authorization_unavailable"
)

var (
	DescriptorRateLimited = descriptor("common.rate_limited", http.StatusTooManyRequests)
	DescriptorValidation  = descriptor("common.validation_failed", http.StatusBadRequest)
	DescriptorInternal    = descriptor("common.internal", http.StatusInternalServerError)

	descriptors = map[string]problem.Descriptor{
		CodeNotFound:                 descriptor(CodeNotFound, http.StatusNotFound),
		CodeForbidden:                descriptor(CodeForbidden, http.StatusForbidden),
		CodeSlugTaken:                descriptor(CodeSlugTaken, http.StatusConflict),
		CodeInvalidType:              descriptor(CodeInvalidType, http.StatusBadRequest),
		CodeInvalidState:             descriptor(CodeInvalidState, http.StatusBadRequest),
		CodeInvalidInput:             descriptor(CodeInvalidInput, http.StatusBadRequest),
		CodeAssetNotFound:            descriptor(CodeAssetNotFound, http.StatusNotFound),
		CodeUpstreamFailed:           descriptor(CodeUpstreamFailed, http.StatusBadGateway),
		CodeRevisionConflict:         descriptor(CodeRevisionConflict, http.StatusPreconditionFailed),
		CodePreconditionRequired:     descriptor(CodePreconditionRequired, http.StatusPreconditionRequired),
		CodeAuthorizationUnavailable: descriptor(CodeAuthorizationUnavailable, http.StatusServiceUnavailable),
	}
)

func descriptor(code string, status int) problem.Descriptor {
	return problem.MustDescriptor(
		problem.MustKind(code, status),
		"https://errors.yueli.dev/problems/"+code,
	)
}

func DescriptorForCode(code string) (problem.Descriptor, bool) {
	value, ok := descriptors[code]
	return value, ok
}

type CatalogEntry struct {
	Code   string `json:"code"`
	Status int    `json:"status"`
}

func Catalog() []CatalogEntry {
	result := make([]CatalogEntry, 0, len(descriptors))
	for code, value := range descriptors {
		result = append(result, CatalogEntry{Code: code, Status: value.Kind().Status()})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Code < result[j].Code })
	return result
}

func mapped(code string, params problem.Parameters) error {
	value, ok := DescriptorForCode(code)
	if !ok {
		return fmt.Errorf("resource public error code is not declared: %s", code)
	}
	result, err := problem.NewError(value, params)
	if err != nil {
		return fmt.Errorf("resource public error %s: %w", code, err)
	}
	return result
}

func NotFound(id string) error {
	return mapped(CodeNotFound, map[string]any{"id": id})
}

func Forbidden() error {
	return mapped(CodeForbidden, nil)
}

func AuthorizationUnavailable() error {
	return mapped(CodeAuthorizationUnavailable, nil)
}

func SlugTaken(slug string) error {
	return mapped(CodeSlugTaken, map[string]any{"slug": slug})
}

func InvalidType(resourceType string) error {
	return mapped(CodeInvalidType, map[string]any{"type": resourceType})
}

func InvalidState(detail string) error {
	return mapped(CodeInvalidState, map[string]any{"detail": detail})
}

func InvalidInput(detail string) error {
	return mapped(CodeInvalidInput, map[string]any{"detail": detail})
}

func AssetNotFound(assetID string) error {
	return mapped(CodeAssetNotFound, map[string]any{"assetId": assetID})
}

func UpstreamFailed(summary string) error {
	return mapped(CodeUpstreamFailed, map[string]any{"detail": summary})
}

func RevisionConflict() error {
	return mapped(CodeRevisionConflict, nil)
}

func PreconditionRequired() error {
	return mapped(CodePreconditionRequired, nil)
}
