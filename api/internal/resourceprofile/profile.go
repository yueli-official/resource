package resourceprofile

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/yueli-official/foundation/go/siteprofile"
)

type Manager struct {
	module     *siteprofile.Service
	store      *siteprofile.PostgresStore
	definition siteprofile.CompiledDefinition
	clock      siteprofile.Clock
}

func Definition() siteprofile.CompiledDefinition {
	definition := siteprofile.DefaultDefinition()
	definition.RequireTagline = true
	definition.RequireLogo = true
	definition.RequireFooterTagline = true
	definition.RequireCopyright = true
	return siteprofile.MustCompileDefinition(definition)
}

func InitialProfile(name, tagline string) siteprofile.Profile {
	if strings.TrimSpace(name) == "" {
		name = "月离资源"
	}
	if strings.TrimSpace(tagline) == "" {
		tagline = "软件、设计素材与脚本"
	}
	return siteprofile.Profile{
		Identity: siteprofile.Identity{Name: name, Tagline: tagline},
		Branding: siteprofile.Branding{Logo: &siteprofile.Visual{
			Kind: siteprofile.VisualIcon, Ref: "i-tabler-package", Alt: name,
		}},
		Footer: siteprofile.Footer{
			Tagline: tagline, Copyright: "© 2026 Yueli",
			LinkGroups: []siteprofile.LinkGroup{},
			Social:     []siteprofile.SocialLink{},
			Legal:      []siteprofile.Link{},
			Compliance: siteprofile.Compliance{Records: []siteprofile.ComplianceRecord{}},
		},
		Support: siteprofile.Support{Contacts: []siteprofile.Contact{}},
	}
}

func NewMemory() *Manager {
	definition := Definition()
	clock := siteprofile.SystemClock{}
	return &Manager{
		module: siteprofile.NewMemory(definition, clock), definition: definition, clock: clock,
	}
}

func NewPostgres(db *sql.DB) (*Manager, error) {
	if db == nil {
		return nil, errors.New("resourceprofile: database is required")
	}
	store, err := siteprofile.NewPostgresStore(db, siteprofile.DefaultPostgresPrefix)
	if err != nil {
		return nil, err
	}
	definition := Definition()
	clock := siteprofile.SystemClock{}
	module, err := siteprofile.New(store, definition, clock)
	if err != nil {
		return nil, err
	}
	return &Manager{module: module, store: store, definition: definition, clock: clock}, nil
}

func (m *Manager) Get(ctx context.Context) (siteprofile.Snapshot, error) {
	return m.module.Get(ctx)
}

func (m *Manager) Schema() siteprofile.FormSchema {
	return m.module.Schema()
}

func (m *Manager) PublicAt(ctx context.Context) (siteprofile.PublicProjection, error) {
	return m.module.PublicAt(ctx, m.clock.Now())
}

func (m *Manager) Replace(ctx context.Context, command siteprofile.ReplaceCommand) (siteprofile.ReplaceResult, error) {
	return m.module.Replace(ctx, command)
}

func (m *Manager) ReplaceTx(
	ctx context.Context,
	tx *sql.Tx,
	command siteprofile.ReplaceCommand,
) (siteprofile.ReplaceResult, error) {
	if m.store == nil {
		return m.module.Replace(ctx, command)
	}
	bound, err := m.store.Bind(tx)
	if err != nil {
		return siteprofile.ReplaceResult{}, err
	}
	module, err := siteprofile.New(bound, m.definition, m.clock)
	if err != nil {
		return siteprofile.ReplaceResult{}, err
	}
	return module.Replace(ctx, command)
}
