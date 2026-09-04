// Package reserr declares Resource's immutable public Problem contract.
package reserr

import (
	"fmt"
	"net/http"

	"github.com/yueli-official/foundation/go/problem"
)

var (
	DescriptorRateLimited = descriptor("common.rate_limited", http.StatusTooManyRequests)
	DescriptorValidation  = descriptor("common.validation_failed", http.StatusBadRequest)
	DescriptorInternal    = descriptor("common.internal", http.StatusInternalServerError)
)

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

func InvalidState(_ string) error {
	return mapped(CodeInvalidState, nil)
}

func InvalidInput(_ string) error {
	return mapped(CodeInvalidInput, nil)
}

func AssetNotFound(assetID string) error {
	return mapped(CodeAssetNotFound, map[string]any{"assetId": assetID})
}

func UpstreamFailed(_ string) error {
	return mapped(CodeUpstreamFailed, nil)
}

func RevisionConflict() error {
	return mapped(CodeRevisionConflict, nil)
}

func PreconditionRequired() error {
	return mapped(CodePreconditionRequired, nil)
}
