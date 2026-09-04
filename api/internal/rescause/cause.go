// Package rescause owns Resource domain and provider failure semantics without
// coupling them to HTTP or the public Problem representation.
package rescause

import "errors"

type Kind string

const (
	KindNotFound                 Kind = "not_found"
	KindForbidden                Kind = "forbidden"
	KindAuthorizationUnavailable Kind = "authorization_unavailable"
	KindSlugTaken                Kind = "slug_taken"
	KindInvalidType              Kind = "invalid_type"
	KindInvalidState             Kind = "invalid_state"
	KindInvalidInput             Kind = "invalid_input"
	KindAssetNotFound            Kind = "asset_not_found"
	KindUpstreamFailed           Kind = "upstream_failed"
	KindRevisionConflict         Kind = "revision_conflict"
	KindPreconditionRequired     Kind = "precondition_required"
)

type Cause struct {
	Kind       Kind
	Value      string
	Diagnostic string
}

func (cause *Cause) Error() string { return string(cause.Kind) }
func (cause *Cause) Is(target error) bool {
	other, ok := target.(*Cause)
	return ok && cause.Kind == other.Kind
}

func Is(err error, kind Kind) bool { return errors.Is(err, &Cause{Kind: kind}) }

func New(kind Kind, value, diagnostic string) error {
	return &Cause{Kind: kind, Value: value, Diagnostic: diagnostic}
}

func NotFound(id string) error               { return New(KindNotFound, id, "") }
func Forbidden() error                       { return New(KindForbidden, "", "") }
func AuthorizationUnavailable() error        { return New(KindAuthorizationUnavailable, "", "") }
func SlugTaken(slug string) error            { return New(KindSlugTaken, slug, "") }
func InvalidType(value string) error         { return New(KindInvalidType, value, "") }
func InvalidState(diagnostic string) error   { return New(KindInvalidState, "", diagnostic) }
func InvalidInput(diagnostic string) error   { return New(KindInvalidInput, "", diagnostic) }
func AssetNotFound(id string) error          { return New(KindAssetNotFound, id, "") }
func UpstreamFailed(diagnostic string) error { return New(KindUpstreamFailed, "", diagnostic) }
func RevisionConflict() error                { return New(KindRevisionConflict, "", "") }
func PreconditionRequired() error            { return New(KindPreconditionRequired, "", "") }
