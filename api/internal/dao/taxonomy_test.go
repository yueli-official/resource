package dao

import "testing"

func TestTaxonomyListOrderAllowlist(t *testing.T) {
	tests := []struct {
		name, sort, direction, want string
	}{
		{"default rejects SQL", "name; DROP TABLE terms", "desc;--", "t.name ASC"},
		{"count descending", "count", "desc", "post_count DESC, t.name ASC"},
		{"slug ascending", "slug", "asc", "t.slug ASC, t.name ASC"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := taxonomyListOrder(tt.sort, tt.direction); got != tt.want {
				t.Fatalf("taxonomyListOrder() = %q, want %q", got, tt.want)
			}
		})
	}
}
