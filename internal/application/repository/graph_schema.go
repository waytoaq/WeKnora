package repository

import (
	"context"

	"github.com/Tencent/WeKnora/internal/types"
	"github.com/Tencent/WeKnora/internal/types/interfaces"
	"gorm.io/gorm"
)

type graphSchemaRepository struct {
	db *gorm.DB
}

// NewGraphSchemaRepository creates a GraphSchemaRepository.
func NewGraphSchemaRepository(db *gorm.DB) interfaces.GraphSchemaRepository {
	return &graphSchemaRepository{db: db}
}

func (r *graphSchemaRepository) CreateAttribute(ctx context.Context, attr *types.KBGraphAttribute) error {
	return r.db.WithContext(ctx).Create(attr).Error
}

func (r *graphSchemaRepository) UpdateAttribute(ctx context.Context, attr *types.KBGraphAttribute) error {
	return r.db.WithContext(ctx).Save(attr).Error
}

func (r *graphSchemaRepository) GetAttribute(ctx context.Context, tenantID uint64, kbID, id string) (*types.KBGraphAttribute, error) {
	var attr types.KBGraphAttribute
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND knowledge_base_id = ? AND id = ?", tenantID, kbID, id).
		First(&attr).Error
	if err != nil {
		return nil, err
	}
	return &attr, nil
}

func (r *graphSchemaRepository) ListAttributes(ctx context.Context, tenantID uint64, kbID string) ([]*types.KBGraphAttribute, error) {
	var list []*types.KBGraphAttribute
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND knowledge_base_id = ?", tenantID, kbID).
		Order("created_at ASC").
		Find(&list).Error
	return list, err
}

func (r *graphSchemaRepository) DeleteAttribute(ctx context.Context, tenantID uint64, kbID, id string) error {
	return r.db.WithContext(ctx).
		Where("tenant_id = ? AND knowledge_base_id = ? AND id = ?", tenantID, kbID, id).
		Delete(&types.KBGraphAttribute{}).Error
}

func (r *graphSchemaRepository) GetAttributesByIDs(ctx context.Context, tenantID uint64, kbID string, ids []string) ([]*types.KBGraphAttribute, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var list []*types.KBGraphAttribute
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND knowledge_base_id = ? AND id IN ?", tenantID, kbID, ids).
		Find(&list).Error
	return list, err
}

func (r *graphSchemaRepository) CreateEntity(ctx context.Context, ent *types.KBGraphEntity) error {
	return r.db.WithContext(ctx).Create(ent).Error
}

func (r *graphSchemaRepository) UpdateEntity(ctx context.Context, ent *types.KBGraphEntity) error {
	return r.db.WithContext(ctx).Save(ent).Error
}

func (r *graphSchemaRepository) GetEntity(ctx context.Context, tenantID uint64, kbID, id string) (*types.KBGraphEntity, error) {
	var ent types.KBGraphEntity
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND knowledge_base_id = ? AND id = ?", tenantID, kbID, id).
		First(&ent).Error
	if err != nil {
		return nil, err
	}
	return &ent, nil
}

func (r *graphSchemaRepository) ListEntities(ctx context.Context, tenantID uint64, kbID string) ([]*types.KBGraphEntity, error) {
	var list []*types.KBGraphEntity
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND knowledge_base_id = ?", tenantID, kbID).
		Order("created_at ASC").
		Find(&list).Error
	return list, err
}

func (r *graphSchemaRepository) DeleteEntity(ctx context.Context, tenantID uint64, kbID, id string) error {
	return r.db.WithContext(ctx).
		Where("tenant_id = ? AND knowledge_base_id = ? AND id = ?", tenantID, kbID, id).
		Delete(&types.KBGraphEntity{}).Error
}

func (r *graphSchemaRepository) CreateRelation(ctx context.Context, rel *types.KBGraphRelation) error {
	return r.db.WithContext(ctx).Create(rel).Error
}

func (r *graphSchemaRepository) UpdateRelation(ctx context.Context, rel *types.KBGraphRelation) error {
	return r.db.WithContext(ctx).Save(rel).Error
}

func (r *graphSchemaRepository) GetRelation(ctx context.Context, tenantID uint64, kbID, id string) (*types.KBGraphRelation, error) {
	var rel types.KBGraphRelation
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND knowledge_base_id = ? AND id = ?", tenantID, kbID, id).
		First(&rel).Error
	if err != nil {
		return nil, err
	}
	return &rel, nil
}

func (r *graphSchemaRepository) ListRelations(ctx context.Context, tenantID uint64, kbID string) ([]*types.KBGraphRelation, error) {
	var list []*types.KBGraphRelation
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND knowledge_base_id = ?", tenantID, kbID).
		Order("created_at ASC").
		Find(&list).Error
	return list, err
}

func (r *graphSchemaRepository) DeleteRelation(ctx context.Context, tenantID uint64, kbID, id string) error {
	return r.db.WithContext(ctx).
		Where("tenant_id = ? AND knowledge_base_id = ? AND id = ?", tenantID, kbID, id).
		Delete(&types.KBGraphRelation{}).Error
}

func (r *graphSchemaRepository) LoadBundle(ctx context.Context, tenantID uint64, kbID string) (*types.GraphSchemaBundle, error) {
	attrs, err := r.ListAttributes(ctx, tenantID, kbID)
	if err != nil {
		return nil, err
	}
	ents, err := r.ListEntities(ctx, tenantID, kbID)
	if err != nil {
		return nil, err
	}
	rels, err := r.ListRelations(ctx, tenantID, kbID)
	if err != nil {
		return nil, err
	}
	return &types.GraphSchemaBundle{Attributes: attrs, Entities: ents, Relations: rels}, nil
}

func (r *graphSchemaRepository) CreateIngestTask(ctx context.Context, task *types.KBGraphIngestTask) error {
	return r.db.WithContext(ctx).Create(task).Error
}

func (r *graphSchemaRepository) UpdateIngestTask(ctx context.Context, task *types.KBGraphIngestTask) error {
	return r.db.WithContext(ctx).Save(task).Error
}

func (r *graphSchemaRepository) GetIngestTask(ctx context.Context, tenantID uint64, kbID, id string) (*types.KBGraphIngestTask, error) {
	var task types.KBGraphIngestTask
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND knowledge_base_id = ? AND id = ?", tenantID, kbID, id).
		First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *graphSchemaRepository) ListIngestTasks(ctx context.Context, tenantID uint64, kbID string) ([]*types.KBGraphIngestTask, error) {
	var list []*types.KBGraphIngestTask
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND knowledge_base_id = ?", tenantID, kbID).
		Order("created_at DESC").
		Find(&list).Error
	return list, err
}
