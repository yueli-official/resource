// Package reserr declares Resource's immutable public Problem contract.
package reserr

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/yueli-official/foundation/go/problem"
)

// Cause preserves the product-owned semantic cause for errors.Is/As while its
// wrapped Problem is the only value projected by the HTTP writer.
type Cause struct {
	Code       string
	Diagnostic string
	problem    error
}

func (cause *Cause) Error() string { return cause.Code }
func (cause *Cause) Unwrap() error { return cause.problem }
func (cause *Cause) Is(target error) bool {
	other, ok := target.(*Cause)
	return ok && cause.Code == other.Code
}

var (
	DescriptorRateLimited = descriptor("common.rate_limited", http.StatusTooManyRequests)
	DescriptorValidation  = descriptor("common.validation_failed", http.StatusBadRequest)
	DescriptorInternal    = descriptor("common.internal", http.StatusInternalServerError)
)

func mapped(code, diagnostic string, params problem.Parameters) error {
	value, ok := DescriptorForCode(code)
	if !ok {
		return fmt.Errorf("resource public error code is not declared: %s", code)
	}
	result, err := problem.NewError(value, params)
	if err != nil {
		return fmt.Errorf("resource public error %s: %w", code, err)
	}
	return &Cause{Code: code, Diagnostic: diagnostic, problem: result}
}

func IsCode(err error, code string) bool {
	return errors.Is(err, &Cause{Code: code})
}

func NotFound(id string) error {
	return mapped(CodeNotFound, "", map[string]any{"id": id})
}

func Forbidden() error {
	return mapped(CodeForbidden, "", nil)
}

func AuthorizationUnavailable() error {
	return mapped(CodeAuthorizationUnavailable, "", nil)
}

func SlugTaken(slug string) error {
	return mapped(CodeSlugTaken, "", map[string]any{"slug": slug})
}

func InvalidType(resourceType string) error {
	return mapped(CodeInvalidType, "", map[string]any{"type": resourceType})
}

func InvalidState(diagnostic string) error {
	return mapped(CodeInvalidState, diagnostic, nil)
}

func InvalidInput(diagnostic string) error {
	return mapped(CodeInvalidInput, diagnostic, nil)
}

func AssetNotFound(assetID string) error {
	return mapped(CodeAssetNotFound, "", map[string]any{"assetId": assetID})
}

func UpstreamFailed(diagnostic string) error {
	return mapped(CodeUpstreamFailed, diagnostic, nil)
}

func RevisionConflict() error {
	return mapped(CodeRevisionConflict, "", nil)
}

func PreconditionRequired() error {
	return mapped(CodePreconditionRequired, "", nil)
}
