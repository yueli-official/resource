package resourcesearch

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/yueli-official/foundation/go/search"
)

const analyzer search.AnalyzerKey = "content-v1"

type Index struct {
	module   search.Module
	postgres *search.Postgres
}

func Definition() search.Definition {
	return search.Definition{
		Consumer: "resource.public", Version: 1,
		Analyzers: []search.AnalyzerDefinition{{Key: analyzer, QueryMode: search.QueryWeb, Required: []search.Capability{search.CapabilityFullText}}},
		Filters: []search.FilterDefinition{
			{Name: "id", MaxValues: 5000}, {Name: "type"}, {Name: "tags", MaxValues: 100, Facetable: true},
		},
		Limits: search.Limits{MaxPageSize: 100},
	}
}

func NewPostgres(ctx context.Context, db *sql.DB, site string) (*Index, error) {
	catalog, err := search.Compile(Definition())
	if err != nil {
		return nil, err
	}
	module, err := search.NewPostgres(ctx, catalog, search.PostgresOptions{
		DB: db, InstanceKey: "resource." + site,
		AnalyzerBindings: map[search.AnalyzerKey]string{analyzer: "chinese_zh"},
	})
	if err != nil {
		return nil, err
	}
	return &Index{module: module, postgres: module}, nil
}

func NewMemory() *Index { return &Index{module: search.NewMemory(search.MustCompile(Definition()))} }

type resourceRow struct {
	ID, Title, Summary, Description, Type, Status string
	Revision                                      uint64
	Tags                                          []string
	SortAt                                        time.Time
}

func scanRow(scanner interface{ Scan(...any) error }) (resourceRow, error) {
	var row resourceRow
	var tags []byte
	err := scanner.Scan(&row.ID, &row.Title, &row.Summary, &row.Description, &row.Type,
		&row.Status, &row.Revision, &tags, &row.SortAt)
	if err == nil {
		err = json.Unmarshal(tags, &row.Tags)
	}
	return row, err
}

const selectResource = `SELECT id,title,summary,description,type,status,search_revision,tags,
	COALESCE(published_at,updated_at) FROM resources`

func (index *Index) Hook(id string) func(context.Context, *sql.Tx) error {
	return func(ctx context.Context, tx *sql.Tx) error {
		row, err := scanRow(tx.QueryRowContext(ctx, selectResource+" WHERE id=$1", id))
		if err != nil {
			return err
		}
		projector, err := index.postgres.Bind(tx)
		if err != nil {
			return err
		}
		return index.applyRow(ctx, projector, row)
	}
}

func (index *Index) DeleteHook(id string, revision uint64) func(context.Context, *sql.Tx) error {
	return func(ctx context.Context, tx *sql.Tx) error {
		projector, err := index.postgres.Bind(tx)
		if err != nil {
			return err
		}
		_, err = projector.Apply(ctx, search.Batch{
			ID: search.BatchID(fmt.Sprintf("resource.item.%s.%d", id, revision+1)),
			Changes: []search.Change{search.Remove(
				search.DocumentKey{Kind: "resource", ID: search.DocumentID(id)},
				search.ProjectionRevision(revision+1),
			)},
		})
		return err
	}
}

func (index *Index) applyRow(ctx context.Context, projector search.Projector, row resourceRow) error {
	key := search.DocumentKey{Kind: "resource", ID: search.DocumentID(row.ID)}
	var change search.Change
	if row.Status == "published" {
		change = search.Upsert(search.SourceDocument{
			Key: key, Revision: search.ProjectionRevision(row.Revision), Analyzer: analyzer,
			Title: row.Title, Summary: row.Summary, Body: row.Description, Keywords: row.Tags, SortAt: row.SortAt.UTC(),
			Filters: search.FieldValues{
				"id": search.Keyword(row.ID), "type": search.Keyword(row.Type), "tags": search.Keywords(row.Tags...),
			},
			Visibility: search.VisibilityReference{ResourceType: "resource.item", ResourceID: row.ID},
		})
	} else {
		change = search.Remove(key, search.ProjectionRevision(row.Revision))
	}
	_, err := projector.Apply(ctx, search.Batch{
		ID:      search.BatchID(fmt.Sprintf("resource.item.%s.%d", row.ID, row.Revision)),
		Changes: []search.Change{change},
	})
	return err
}

func (index *Index) Reconcile(ctx context.Context, db *sql.DB) error {
	rows, err := db.QueryContext(ctx, selectResource)
	if err != nil {
		return err
	}
	var values []resourceRow
	for rows.Next() {
		row, scanErr := scanRow(rows)
		if scanErr != nil {
			_ = rows.Close()
			return scanErr
		}
		values = append(values, row)
	}
	if err := rows.Close(); err != nil {
		return err
	}
	for _, row := range values {
		if err := index.applyRow(ctx, index.module, row); err != nil {
			return err
		}
	}
	return nil
}

func (index *Index) Search(ctx context.Context, text, resourceType string, tags, ids []string, limit, offset int) (search.Page, error) {
	filters := []search.Filter{}
	if resourceType != "" {
		filters = append(filters, search.Equal("type", resourceType))
	}
	if len(tags) > 0 {
		filters = append(filters, search.All("tags", tags...))
	}
	if ids != nil {
		if len(ids) == 0 {
			return search.Page{}, nil
		}
		filters = append(filters, search.Any("id", ids...))
	}
	needed := offset + limit
	var result search.Page
	var cursor search.Cursor
	for len(result.Hits) < needed {
		page, err := index.module.Search(ctx, search.Query{
			Text: text, Analyzer: analyzer, Filters: filters,
			Page: search.PageRequest{Size: min(100, needed-len(result.Hits)), Cursor: cursor},
		})
		if err != nil {
			return search.Page{}, err
		}
		result.Plan, result.Total = page.Plan, page.Total
		result.Hits = append(result.Hits, page.Hits...)
		if page.NextCursor == "" {
			break
		}
		cursor = page.NextCursor
	}
	if offset >= len(result.Hits) {
		result.Hits = nil
	} else {
		result.Hits = result.Hits[offset:min(offset+limit, len(result.Hits))]
	}
	return result, nil
}
