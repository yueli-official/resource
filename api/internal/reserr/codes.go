// Package reserr declares Resource's immutable public Problem contract.
package reserr

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/yueli-official/foundation/go/problem"
	"github.com/yueli-official/resource/api/internal/rescause"
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

func InvalidState(diagnostic string) error {
	return mapped(CodeInvalidState, nil)
}

func InvalidInput(diagnostic string) error {
	return mapped(CodeInvalidInput, nil)
}

func AssetNotFound(assetID string) error {
	return mapped(CodeAssetNotFound, map[string]any{"assetId": assetID})
}

func UpstreamFailed(diagnostic string) error {
	return mapped(CodeUpstreamFailed, nil)
}

func RevisionConflict() error {
	return mapped(CodeRevisionConflict, nil)
}

func PreconditionRequired() error {
	return mapped(CodePreconditionRequired, nil)
}

// MapCause is the single application boundary from domain/provider causes to
// the public error catalog. Unknown failures stay unknown for safe 500 handling.
func MapCause(err error) error {
	var cause *rescause.Cause
	if !errors.As(err, &cause) {
		return err
	}
	switch cause.Kind {
	case rescause.KindNotFound:
		return NotFound(cause.Value)
	case rescause.KindForbidden:
		return Forbidden()
	case rescause.KindAuthorizationUnavailable:
		return AuthorizationUnavailable()
	case rescause.KindSlugTaken:
		return SlugTaken(cause.Value)
	case rescause.KindInvalidType:
		return InvalidType(cause.Value)
	case rescause.KindInvalidState:
		return InvalidState(cause.Diagnostic)
	case rescause.KindInvalidInput:
		return InvalidInput(cause.Diagnostic)
	case rescause.KindAssetNotFound:
		return AssetNotFound(cause.Value)
	case rescause.KindUpstreamFailed:
		return UpstreamFailed(cause.Diagnostic)
	case rescause.KindRevisionConflict:
		return RevisionConflict()
	case rescause.KindPreconditionRequired:
		return PreconditionRequired()
	default:
		return err
	}
}
