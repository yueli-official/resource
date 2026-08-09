package controller

import (
	"testing"

	"github.com/yueli-official/foundation/go/identifier"

	v1 "github.com/yueli-official/resource/api/api/v1"
)

func TestNormalizeDeliveryPayloadIDsUsesIdentifierModule(t *testing.T) {
	existing := identifier.MustNew().String()
	original := &v1.DeliveryPayloadView{Items: []*v1.DeliveryItemView{
		{ID: existing, Kind: "asset_file"},
		{ID: "draft-delivery-1", Kind: "netdisk", Netdisk: &v1.NetdiskDeliveryView{URL: "https://example.com/file"}},
	}}
	got := normalizeDeliveryPayloadIDs(original)
	if got == original || got.Items[1] == original.Items[1] || got.Items[1].Netdisk == original.Items[1].Netdisk {
		t.Fatal("normalization must not mutate the request payload")
	}
	if got.Items[0].ID != existing {
		t.Fatalf("existing ID = %q, want %q", got.Items[0].ID, existing)
	}
	if got.Items[1].ID == original.Items[1].ID {
		t.Fatalf("draft ID was not replaced: %q", got.Items[1].ID)
	}
	value, err := identifier.Parse(got.Items[1].ID)
	if err != nil || value.Version() != 7 {
		t.Fatalf("assigned ID = %q, error = %v, want UUIDv7", got.Items[1].ID, err)
	}
}
