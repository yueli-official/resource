package reserr

import (
	"errors"
	"testing"
)

func TestCausePreservesSemanticIdentityWithoutExposingDiagnostic(t *testing.T) {
	err := InvalidInput("database constraint detail")
	var cause *Cause
	if !errors.As(err, &cause) {
		t.Fatal("expected typed Resource cause")
	}
	if cause.Diagnostic != "database constraint detail" {
		t.Fatalf("diagnostic = %q", cause.Diagnostic)
	}
	if err.Error() != CodeInvalidInput {
		t.Fatalf("public error text = %q", err.Error())
	}
	if !IsCode(err, CodeInvalidInput) {
		t.Fatal("errors.Is identity was not preserved")
	}
}
