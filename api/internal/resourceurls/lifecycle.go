// Package resourceurls adapts Resource's stable ID detail routes and taxonomy
// slug routes to the Foundation URL lifecycle module.
package resourceurls

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/yueli-official/foundation/go/identifier"
	"github.com/yueli-official/foundation/go/urllifecycle"
)

const (
	ResourceKind urllifecycle.ResourceKind = "resource.item"
	CategoryKind urllifecycle.ResourceKind = "resource.category"
	TagKind      urllifecycle.ResourceKind = "resource.tag"
)

type Lifecycle struct {
	module   urllifecycle.Module
	postgres *urllifecycle.PostgresAdapter
}

type TaxonomyState struct {
	ID   string
	Kind urllifecycle.ResourceKind
	Slug string
}

func Definition(origin string) urllifecycle.Definition {
	return urllifecycle.Definition{
		Version:       urllifecycle.DefinitionVersion,
		TrustedOrigin: strings.TrimRight(origin, "/"),
		ResourceKinds: []urllifecycle.ResourceKindDefinition{
			{Key: ResourceKind}, {Key: CategoryKind}, {Key: TagKind},
		},
		Namespaces: []urllifecycle.NamespaceDefinition{
			{Key: "resource.items", PathPrefix: "/resources"},
			{Key: "resource.categories", PathPrefix: "/category"},
			{Key: "resource.tags", PathPrefix: "/tags"},
		},
		Limits: urllifecycle.Limits{MaxPageSize: 1000},
	}
}

func NewPostgres(ctx context.Context, db *sql.DB, instanceKey, origin string) (*Lifecycle, error) {
	catalog, err := urllifecycle.Compile(Definition(origin))
	if err != nil {
		return nil, err
	}
	adapter, err := urllifecycle.NewPostgres(ctx, catalog, urllifecycle.PostgresOptions{
		DB: db, InstanceKey: instanceKey,
	})
	if err != nil {
		return nil, err
	}
	return &Lifecycle{module: adapter, postgres: adapter}, nil
}

func NewMemory(origin string) (*Lifecycle, error) {
	catalog, err := urllifecycle.Compile(Definition(origin))
	if err != nil {
		return nil, err
	}
	module, err := urllifecycle.NewMemory(catalog, urllifecycle.MemoryOptions{})
	if err != nil {
		return nil, err
	}
	return &Lifecycle{module: module}, nil
}

func (l *Lifecycle) Resolver() urllifecycle.Resolver {
	if l == nil {
		return nil
	}
	return l.module
}

func (l *Lifecycle) ReconcileAll(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	rows, err := tx.QueryContext(ctx, `SELECT id::text FROM resources WHERE published_at IS NOT NULL ORDER BY id`)
	if err != nil {
		return err
	}
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			_ = rows.Close()
			return err
		}
		ids = append(ids, id)
	}
	if err := rows.Close(); err != nil {
		return err
	}
	if err := l.ReconcileResources(ctx, tx, ids); err != nil {
		return err
	}
	if err := l.ReconcileTaxonomies(ctx, tx, "resource taxonomy startup reconciliation"); err != nil {
		return err
	}
	return tx.Commit()
}

func (l *Lifecycle) ReconcileResources(ctx context.Context, tx *sql.Tx, ids []string) error {
	module, err := l.bound(tx)
	if err != nil {
		return err
	}
	for _, id := range ids {
		var published bool
		err := tx.QueryRowContext(ctx,
			`SELECT published_at IS NOT NULL FROM resources WHERE id = $1::uuid`, id,
		).Scan(&published)
		if errors.Is(err, sql.ErrNoRows) {
			continue
		}
		if err != nil {
			return err
		}
		if published {
			if err := ensure(ctx, module, resourceRoute(id), urllifecycle.LocalRef{Path: "/resources/" + id}, "resource published"); err != nil {
				return err
			}
		}
	}
	return nil
}

func (l *Lifecycle) DeleteResource(ctx context.Context, tx *sql.Tx, id string) error {
	module, err := l.bound(tx)
	if err != nil {
		return err
	}
	return gone(ctx, module, resourceRoute(id), "resource deleted")
}

func (l *Lifecycle) ReconcileTaxonomies(ctx context.Context, tx *sql.Tx, reason string) error {
	module, err := l.bound(tx)
	if err != nil {
		return err
	}
	rows, err := tx.QueryContext(ctx, `
SELECT taxonomy.id::text, taxonomy.taxonomy, term.slug
FROM taxonomies taxonomy
JOIN terms term ON term.id = taxonomy.term_id
WHERE taxonomy.status = 'active'
ORDER BY taxonomy.id`)
	if err != nil {
		return err
	}
	desired := map[string]TaxonomyState{}
	for rows.Next() {
		var state TaxonomyState
		var kind string
		if err := rows.Scan(&state.ID, &kind, &state.Slug); err != nil {
			_ = rows.Close()
			return err
		}
		state.Kind = CategoryKind
		if kind == "tag" {
			state.Kind = TagKind
		}
		desired[routeID(taxonomyRoute(state))] = state
	}
	if err := rows.Close(); err != nil {
		return err
	}
	active, err := listTaxonomyActive(ctx, module)
	if err != nil {
		return err
	}
	changes := make([]urllifecycle.ResourceChange, 0)
	for key, state := range desired {
		inspection, found := active[key]
		if !found {
			next := urllifecycle.ActiveRoute{Canonical: taxonomyRef(state)}
			changes = append(changes, urllifecycle.ResourceChange{
				Route: taxonomyRoute(state), Desired: &next, Departures: releaseDeparture(),
			})
			continue
		}
		delete(active, key)
		if inspection.Active.Canonical.Path != taxonomyRef(state).Path {
			next := *inspection.Active
			next.Canonical = taxonomyRef(state)
			changes = append(changes, urllifecycle.ResourceChange{
				Route: taxonomyRoute(state), ExpectedRevision: inspection.Revision, Desired: &next,
				Departures: redirectDeparture(),
			})
		}
	}
	for _, inspection := range active {
		if inspection.Route == nil {
			continue
		}
		changes = append(changes, goneChange(*inspection.Route, inspection.Revision))
	}
	if len(changes) == 0 {
		return nil
	}
	_, err = module.Apply(ctx, changeSet(reason, changes), urllifecycle.ApplyOptions{})
	return err
}

func (l *Lifecycle) MergeTaxonomy(ctx context.Context, tx *sql.Tx, source, target TaxonomyState) error {
	module, err := l.bound(tx)
	if err != nil {
		return err
	}
	if err := ensure(ctx, module, taxonomyRoute(target), taxonomyRef(target), "resource taxonomy merge target repair"); err != nil {
		return err
	}
	sourceInspection, found, err := inspect(ctx, module, taxonomyRoute(source))
	if err != nil {
		return err
	}
	if !found {
		return nil
	}
	change := urllifecycle.Merge(
		meta("resource taxonomy merged"),
		taxonomyRoute(source),
		sourceInspection.Revision,
		taxonomyRoute(target),
		urllifecycle.DefaultPermanentRedirect(),
	)
	_, err = module.Apply(ctx, change, urllifecycle.ApplyOptions{})
	return err
}

func (l *Lifecycle) bound(tx *sql.Tx) (urllifecycle.Module, error) {
	if l == nil || l.module == nil {
		return nil, fmt.Errorf("resource URL lifecycle is not configured")
	}
	if l.postgres == nil {
		return l.module, nil
	}
	if tx == nil {
		return nil, fmt.Errorf("resource URL lifecycle requires the product transaction")
	}
	return l.postgres.Bind(tx)
}

func ensure(ctx context.Context, module urllifecycle.Module, key urllifecycle.RouteKey, ref urllifecycle.LocalRef, reason string) error {
	inspection, found, err := inspect(ctx, module, key)
	if err != nil {
		return err
	}
	if found {
		if inspection.Active.Canonical.Path == ref.Path {
			return nil
		}
		change := urllifecycle.Rename(meta(reason), key, inspection.Revision, *inspection.Active, ref, urllifecycle.DefaultPermanentRedirect())
		_, err = module.Apply(ctx, change, urllifecycle.ApplyOptions{})
		return err
	}
	change := urllifecycle.Claim(meta(reason), urllifecycle.ClaimSpec{
		Route: key, Active: urllifecycle.ActiveRoute{Canonical: ref},
	})
	_, err = module.Apply(ctx, change, urllifecycle.ApplyOptions{})
	return err
}

func gone(ctx context.Context, module urllifecycle.Module, key urllifecycle.RouteKey, reason string) error {
	inspection, found, err := inspect(ctx, module, key)
	if err != nil || !found {
		return err
	}
	change := urllifecycle.Retire(meta(reason), urllifecycle.RetireGone(key, inspection.Revision))
	_, err = module.Apply(ctx, change, urllifecycle.ApplyOptions{})
	return err
}

func inspect(ctx context.Context, module urllifecycle.Reader, key urllifecycle.RouteKey) (urllifecycle.Inspection, bool, error) {
	value, err := module.Inspect(ctx, urllifecycle.InspectQuery{Route: &key})
	if err == nil {
		return value, true, nil
	}
	var lifecycleErr *urllifecycle.Error
	if errors.As(err, &lifecycleErr) && lifecycleErr.Kind == urllifecycle.ErrorNotFound {
		return urllifecycle.Inspection{}, false, nil
	}
	return urllifecycle.Inspection{}, false, err
}

func listTaxonomyActive(ctx context.Context, module urllifecycle.Module) (map[string]urllifecycle.Inspection, error) {
	result := map[string]urllifecycle.Inspection{}
	after := ""
	for {
		page, err := module.List(ctx, urllifecycle.ListQuery{After: after, Limit: 1000})
		if err != nil {
			return nil, err
		}
		for _, item := range page.Items {
			if item.Route == nil || item.Active == nil {
				continue
			}
			if item.Route.Resource.Kind == CategoryKind || item.Route.Resource.Kind == TagKind {
				result[routeID(*item.Route)] = item
			}
		}
		if page.Next == "" {
			return result, nil
		}
		after = page.Next
	}
}

func resourceRoute(id string) urllifecycle.RouteKey {
	return urllifecycle.RouteKey{Resource: urllifecycle.ResourceKey{Kind: ResourceKind, ID: id}}
}

func taxonomyRoute(state TaxonomyState) urllifecycle.RouteKey {
	return urllifecycle.RouteKey{Resource: urllifecycle.ResourceKey{Kind: state.Kind, ID: state.ID}}
}

func taxonomyRef(state TaxonomyState) urllifecycle.LocalRef {
	prefix := "/category/"
	if state.Kind == TagKind {
		prefix = "/tags/"
	}
	return urllifecycle.LocalRef{Path: prefix + state.Slug}
}

func routeID(route urllifecycle.RouteKey) string {
	return string(route.Resource.Kind) + "\x00" + route.Resource.ID
}

func meta(reason string) urllifecycle.MutationMeta {
	return urllifecycle.MutationMeta{
		CommandID: urllifecycle.CommandID(identifier.MustNew().String()),
		Actor:     urllifecycle.ActorRef{Kind: "system", ID: "resource"},
		Reason:    reason,
	}
}

func changeSet(reason string, changes []urllifecycle.ResourceChange) urllifecycle.ChangeSet {
	return urllifecycle.ChangeSet{
		CommandID: urllifecycle.CommandID(identifier.MustNew().String()),
		Actor:     urllifecycle.ActorRef{Kind: "system", ID: "resource"},
		Reason:    reason, ResourceChanges: changes,
	}
}

func goneChange(key urllifecycle.RouteKey, revision urllifecycle.RouteRevision) urllifecycle.ResourceChange {
	return urllifecycle.ResourceChange{
		Route: key, ExpectedRevision: revision,
		Departures: urllifecycle.DeparturePolicy{
			Canonical: urllifecycle.FormerOutcome{Kind: urllifecycle.FormerGone},
			Aliases:   urllifecycle.FormerOutcome{Kind: urllifecycle.FormerGone},
		},
	}
}

func redirectDeparture() urllifecycle.DeparturePolicy {
	return urllifecycle.DeparturePolicy{
		Canonical: urllifecycle.FormerOutcome{
			Kind: urllifecycle.FormerRedirectToCurrent, Redirect: urllifecycle.DefaultPermanentRedirect(),
		},
		Aliases: urllifecycle.FormerOutcome{Kind: urllifecycle.FormerRelease},
	}
}

func releaseDeparture() urllifecycle.DeparturePolicy {
	return urllifecycle.DeparturePolicy{
		Canonical: urllifecycle.FormerOutcome{Kind: urllifecycle.FormerRelease},
		Aliases:   urllifecycle.FormerOutcome{Kind: urllifecycle.FormerRelease},
	}
}
