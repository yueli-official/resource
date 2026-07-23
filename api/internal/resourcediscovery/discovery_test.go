package resourcediscovery

import (
	"strings"
	"testing"
	"time"

	"github.com/gogf/gf/v2/os/gtime"
	"github.com/yueli-official/foundation/go/discovery"
	"github.com/yueli-official/foundation/go/siteprofile"

	"platform/products/resource/api/internal/model"
)

func TestProjectResourceProducesProductAndSitemapProjection(t *testing.T) {
	module := discovery.MustCompile(discovery.Definition{
		ContractVersion: discovery.ContractVersion,
		Site: discovery.SiteProfile{
			Origin: "https://resource.example", Name: "Resources", DefaultLocale: "zh-CN",
		},
	})
	updated := gtime.NewFromTime(time.Date(2026, 7, 2, 0, 0, 0, 0, time.UTC))
	projection, err := ProjectResource(module, &model.Resource{
		ID: "one", Title: "Tool", Summary: "Useful",
		Status: model.StatusPublished, UpdatedAt: updated,
	}, nil, "Resources")
	if err != nil {
		t.Fatal(err)
	}
	if projection.Sitemap == nil ||
		projection.Sitemap.Location != "https://resource.example/resources/one" {
		t.Fatalf("resource sitemap is missing: %#v", projection.Sitemap)
	}
	if !strings.Contains(string(projection.Head.StructuredData[0].JSON), `"Product"`) {
		t.Fatalf("product schema is missing: %s", projection.Head.StructuredData[0].JSON)
	}
}

func TestManagerTracksSiteProfileRevision(t *testing.T) {
	config := Config{
		Origin: "https://resource.example", Locale: "zh-CN",
		TTL: time.Minute, Clock: time.Now,
	}
	manager, err := NewManager(nil, config, siteprofile.Snapshot{
		Revision: 1,
		Profile: siteprofile.Profile{
			Identity: siteprofile.Identity{Name: "First", Description: "First description"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := manager.Refresh(siteprofile.Snapshot{
		Revision: 2,
		Profile: siteprofile.Profile{
			Identity: siteprofile.Identity{Name: "Second", Description: "Second description"},
		},
	}); err != nil {
		t.Fatal(err)
	}
	projection, err := manager.ProjectResource(&model.Resource{
		ID: "one", Title: "Tool", Summary: "Useful", Status: model.StatusPublished,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if manager.ProfileRevision() != 2 || manager.SiteName() != "Second" {
		t.Fatalf("revision=%d name=%q", manager.ProfileRevision(), manager.SiteName())
	}
	if !strings.Contains(string(projection.Head.StructuredData[0].JSON), `"Second"`) {
		t.Fatalf("updated Site Profile did not reach Discovery: %s", projection.Head.StructuredData[0].JSON)
	}
}
