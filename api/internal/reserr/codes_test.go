package reserr

import (
	"errors"
	"testing"

	"github.com/yueli-official/foundation/go/problem"
	"github.com/yueli-official/resource/api/internal/rescause"
)

func TestMapCausePreservesDomainIdentityAndProjectsCatalogError(t *testing.T) {
	err := rescause.InvalidInput("database constraint detail")
	var cause *rescause.Cause
	if !errors.As(err, &cause) || cause.Diagnostic != "database constraint detail" {
		t.Fatalf("typed cause = %#v", cause)
	}
	if !rescause.Is(err, rescause.KindInvalidInput) {
		t.Fatal("errors.Is identity was not preserved")
	}
	mapped := MapCause(err)
	public, ok, mapErr := problem.FromError(mapped, "test-trace")
	if mapErr != nil || !ok || public.Code != CodeInvalidInput {
		t.Fatalf("mapped public error = %#v", mapped)
	}
}
