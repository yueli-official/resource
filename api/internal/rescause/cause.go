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

func NotFoundError(id string) error               { return New(KindNotFound, id, "") }
func ForbiddenError() error                       { return New(KindForbidden, "", "") }
func AuthorizationUnavailableError() error        { return New(KindAuthorizationUnavailable, "", "") }
func SlugTakenError(slug string) error            { return New(KindSlugTaken, slug, "") }
func InvalidTypeError(value string) error         { return New(KindInvalidType, value, "") }
func InvalidStateError(diagnostic string) error   { return New(KindInvalidState, "", diagnostic) }
func InvalidInputError(diagnostic string) error   { return New(KindInvalidInput, "", diagnostic) }
func AssetNotFoundError(id string) error          { return New(KindAssetNotFound, id, "") }
func UpstreamFailedError(diagnostic string) error { return New(KindUpstreamFailed, "", diagnostic) }
func RevisionConflictError() error                { return New(KindRevisionConflict, "", "") }
func PreconditionRequiredError() error            { return New(KindPreconditionRequired, "", "") }

func NotFound(id string) error               { return NotFoundError(id) }
func Forbidden() error                       { return ForbiddenError() }
func AuthorizationUnavailable() error        { return AuthorizationUnavailableError() }
func SlugTaken(slug string) error            { return SlugTakenError(slug) }
func InvalidType(value string) error         { return InvalidTypeError(value) }
func InvalidState(diagnostic string) error   { return InvalidStateError(diagnostic) }
func InvalidInput(diagnostic string) error   { return InvalidInputError(diagnostic) }
func AssetNotFound(id string) error          { return AssetNotFoundError(id) }
func UpstreamFailed(diagnostic string) error { return UpstreamFailedError(diagnostic) }
func RevisionConflict() error                { return RevisionConflictError() }
func PreconditionRequired() error            { return PreconditionRequiredError() }
