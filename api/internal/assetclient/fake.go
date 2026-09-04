package assetclient

import (
	"context"
	"strings"
	"sync"

	"github.com/yueli-official/foundation/go/identifier"
	reserr "github.com/yueli-official/resource/api/internal/rescause"
)

// Fake is an in-memory AssetClient for tests. It mirrors visibility
// from init→finalize so public assets get a media key.
// FailNext makes the next call return an upstream error (resilience tests).
type Fake struct {
	mu         sync.Mutex
	pending    map[string]string // uploadToken -> visibility
	deleted    []string
	refs       []ReferenceInput
	unrefs     []ReferenceInput
	parts      []int
	completes  []MultipartCompleteInput
	aborts     []string
	failNext   bool
	PublicBase string
}

func NewFake() *Fake {
	return &Fake{pending: map[string]string{}, PublicBase: "http://asset.test/public"}
}

// FailNext arms a one-shot upstream failure on the next call.
func (f *Fake) FailNext() { f.mu.Lock(); f.failNext = true; f.mu.Unlock() }

func (f *Fake) tripped() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.failNext {
		f.failNext = false
		return true
	}
	return false
}

func (f *Fake) UploadInit(_ context.Context, _ string, in InitInput) (InitOutput, error) {
	if f.tripped() {
		return InitOutput{}, reserr.UpstreamFailed("fake upstream 503")
	}
	tok := "faketok-" + identifier.MustNew().String()
	f.mu.Lock()
	f.pending[tok] = in.Visibility
	f.mu.Unlock()
	if in.Multipart {
		return InitOutput{
			UploadToken: tok,
			Method:      "MULTIPART",
			UploadID:    "fake-upload-" + tok,
			PartSize:    64 << 20,
			PartCount:   1,
		}, nil
	}
	return InitOutput{UploadURL: "http://asset.test/api/v1/assets/blob/" + tok, UploadToken: tok}, nil
}

func (f *Fake) MultipartPartURL(_ context.Context, _ string, in MultipartPartURLInput) (MultipartPartURLOutput, error) {
	if f.tripped() {
		return MultipartPartURLOutput{}, reserr.UpstreamFailed("fake upstream 503")
	}
	f.mu.Lock()
	f.parts = append(f.parts, in.PartNumber)
	f.mu.Unlock()
	return MultipartPartURLOutput{UploadURL: "http://asset.test/multipart/" + in.UploadToken}, nil
}

func (f *Fake) CompleteMultipart(_ context.Context, _ string, in MultipartCompleteInput) error {
	if f.tripped() {
		return reserr.UpstreamFailed("fake upstream 503")
	}
	f.mu.Lock()
	f.completes = append(f.completes, in)
	f.mu.Unlock()
	return nil
}

func (f *Fake) AbortMultipart(_ context.Context, _ string, in MultipartAbortInput) error {
	if f.tripped() {
		return reserr.UpstreamFailed("fake upstream 503")
	}
	f.mu.Lock()
	f.aborts = append(f.aborts, in.UploadToken)
	f.mu.Unlock()
	return nil
}

func (f *Fake) Finalize(_ context.Context, _, uploadToken string) (View, error) {
	if f.tripped() {
		return View{}, reserr.UpstreamFailed("fake upstream 503")
	}
	f.mu.Lock()
	vis := f.pending[uploadToken]
	f.mu.Unlock()
	if vis == "" {
		vis = "public"
	}
	id := identifier.MustNew().String()
	v := View{ID: id, MediaKey: strings.ReplaceAll(id, "-", ""), Size: 1234, Mime: "application/zip", Filename: "file.zip", Visibility: vis}
	if vis == "public" {
		v.MediaKey = strings.ReplaceAll(id, "-", "")
	}
	return v, nil
}

func (f *Fake) Delete(_ context.Context, _, assetID string) error {
	if f.tripped() {
		return reserr.UpstreamFailed("fake upstream 503")
	}
	f.mu.Lock()
	f.deleted = append(f.deleted, assetID)
	f.mu.Unlock()
	return nil
}

func (f *Fake) RegisterReference(_ context.Context, _ string, in ReferenceInput) error {
	if f.tripped() {
		return reserr.UpstreamFailed("fake upstream 503")
	}
	f.mu.Lock()
	f.refs = append(f.refs, in)
	f.mu.Unlock()
	return nil
}

func (f *Fake) UnregisterReference(_ context.Context, _ string, in ReferenceInput) error {
	if f.tripped() {
		return reserr.UpstreamFailed("fake upstream 503")
	}
	f.mu.Lock()
	f.unrefs = append(f.unrefs, in)
	f.mu.Unlock()
	return nil
}

// Deleted returns the asset ids Delete was called with (cleanup assertions).
func (f *Fake) Deleted() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.deleted...)
}
