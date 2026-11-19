package interfaces

import (
	"context"

	"github.com/Tencent/WeKnora/internal/types"
	"github.com/hibiken/asynq"
)

// GraphSchemaRepository persists KB-scoped graph attribute/entity/relation defs.
type GraphSchemaRepository interface {
	CreateAttribute(ctx context.Context, attr *types.KBGraphAttribute) error
	UpdateAttribute(ctx context.Context, attr *types.KBGraphAttribute) error
	GetAttribute(ctx context.Context, tenantID uint64, kbID, id string) (*types.KBGraphAttribute, error)
	ListAttributes(ctx context.Context, tenantID uint64, kbID string) ([]*types.KBGraphAttribute, error)
	DeleteAttribute(ctx context.Context, tenantID uint64, kbID, id string) error
	GetAttributesByIDs(ctx context.Context, tenantID uint64, kbID string, ids []string) ([]*types.KBGraphAttribute, error)

	CreateEntity(ctx context.Context, ent *types.KBGraphEntity) error
	UpdateEntity(ctx context.Context, ent *types.KBGraphEntity) error
	GetEntity(ctx context.Context, tenantID uint64, kbID, id string) (*types.KBGraphEntity, error)
	ListEntities(ctx context.Context, tenantID uint64, kbID string) ([]*types.KBGraphEntity, error)
	DeleteEntity(ctx context.Context, tenantID uint64, kbID, id string) error

	CreateRelation(ctx context.Context, rel *types.KBGraphRelation) error
	UpdateRelation(ctx context.Context, rel *types.KBGraphRelation) error
	GetRelation(ctx context.Context, tenantID uint64, kbID, id string) (*types.KBGraphRelation, error)
	ListRelations(ctx context.Context, tenantID uint64, kbID string) ([]*types.KBGraphRelation, error)
	DeleteRelation(ctx context.Context, tenantID uint64, kbID, id string) error

	LoadBundle(ctx context.Context, tenantID uint64, kbID string) (*types.GraphSchemaBundle, error)

	CreateIngestTask(ctx context.Context, task *types.KBGraphIngestTask) error
	UpdateIngestTask(ctx context.Context, task *types.KBGraphIngestTask) error
	GetIngestTask(ctx context.Context, tenantID uint64, kbID, id string) (*types.KBGraphIngestTask, error)
	ListIngestTasks(ctx context.Context, tenantID uint64, kbID string) ([]*types.KBGraphIngestTask, error)
}

// GraphSchemaService exposes KB graph schema CRUD and ingest task management.
type GraphSchemaService interface {
	ListAttributes(ctx context.Context, kbID string) ([]*types.KBGraphAttribute, error)
	CreateAttribute(ctx context.Context, kbID string, attr *types.KBGraphAttribute) (*types.KBGraphAttribute, error)
	UpdateAttribute(ctx context.Context, kbID, id string, attr *types.KBGraphAttribute) (*types.KBGraphAttribute, error)
	DeleteAttribute(ctx context.Context, kbID, id string) error

	ListEntities(ctx context.Context, kbID string) ([]*types.KBGraphEntity, error)
	CreateEntity(ctx context.Context, kbID string, ent *types.KBGraphEntity) (*types.KBGraphEntity, error)
	UpdateEntity(ctx context.Context, kbID, id string, ent *types.KBGraphEntity) (*types.KBGraphEntity, error)
	DeleteEntity(ctx context.Context, kbID, id string) error

	ListRelations(ctx context.Context, kbID string) ([]*types.KBGraphRelation, error)
	CreateRelation(ctx context.Context, kbID string, rel *types.KBGraphRelation) (*types.KBGraphRelation, error)
	UpdateRelation(ctx context.Context, kbID, id string, rel *types.KBGraphRelation) (*types.KBGraphRelation, error)
	DeleteRelation(ctx context.Context, kbID, id string) error

	LoadBundle(ctx context.Context, kbID string) (*types.GraphSchemaBundle, error)
	CompileExtractHints(ctx context.Context, kbID string) (*types.GraphExtractHints, error)

	ListIngestTasks(ctx context.Context, kbID string) ([]*types.KBGraphIngestTask, error)
	GetIngestTask(ctx context.Context, kbID, id string) (*types.KBGraphIngestTask, error)
	CreateIngestTask(ctx context.Context, kbID string, task *types.KBGraphIngestTask) (*types.KBGraphIngestTask, error)
	CancelIngestTask(ctx context.Context, kbID, id string) (*types.KBGraphIngestTask, error)
	// RetryIngestTask resets a failed/cancelled task to pending and re-enqueues it.
	RetryIngestTask(ctx context.Context, kbID, id string) (*types.KBGraphIngestTask, error)

	// ProcessIngestTask is the asynq handler for TypeGraphSchemaIngest.
	ProcessIngestTask(ctx context.Context, t *asynq.Task) error
}
