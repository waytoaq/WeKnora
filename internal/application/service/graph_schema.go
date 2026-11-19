package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	apperrors "github.com/Tencent/WeKnora/internal/errors"
	"github.com/Tencent/WeKnora/internal/logger"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/Tencent/WeKnora/internal/types/interfaces"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

type graphSchemaService struct {
	repo        interfaces.GraphSchemaRepository
	graphEngine interfaces.RetrieveGraphRepository
	taskClient  interfaces.TaskEnqueuer
}

// NewGraphSchemaService creates a GraphSchemaService.
func NewGraphSchemaService(
	repo interfaces.GraphSchemaRepository,
	graphEngine interfaces.RetrieveGraphRepository,
	taskClient interfaces.TaskEnqueuer,
) interfaces.GraphSchemaService {
	return &graphSchemaService{repo: repo, graphEngine: graphEngine, taskClient: taskClient}
}

func (s *graphSchemaService) tenantID(ctx context.Context) (uint64, error) {
	tid, ok := types.TenantIDFromContext(ctx)
	if !ok || tid == 0 {
		return 0, apperrors.NewUnauthorizedError("tenant required")
	}
	return tid, nil
}

func normalizeCode(code string) string {
	return strings.TrimSpace(code)
}

func parseStringIDs(raw types.JSON) ([]string, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var ids []string
	if err := json.Unmarshal(raw, &ids); err != nil {
		return nil, fmt.Errorf("invalid attribute_ids: %w", err)
	}
	return ids, nil
}

func encodeStringIDs(ids []string) types.JSON {
	if ids == nil {
		ids = []string{}
	}
	b, _ := json.Marshal(ids)
	return types.JSON(b)
}

func (s *graphSchemaService) validateAttributeIDs(ctx context.Context, tenantID uint64, kbID string, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	found, err := s.repo.GetAttributesByIDs(ctx, tenantID, kbID, ids)
	if err != nil {
		return err
	}
	if len(found) != len(ids) {
		return apperrors.NewBadRequestError("attribute_ids 必须属于当前知识库")
	}
	return nil
}

// --- Attributes ---

func (s *graphSchemaService) ListAttributes(ctx context.Context, kbID string) ([]*types.KBGraphAttribute, error) {
	tid, err := s.tenantID(ctx)
	if err != nil {
		return nil, err
	}
	return s.repo.ListAttributes(ctx, tid, kbID)
}

func (s *graphSchemaService) CreateAttribute(ctx context.Context, kbID string, attr *types.KBGraphAttribute) (*types.KBGraphAttribute, error) {
	tid, err := s.tenantID(ctx)
	if err != nil {
		return nil, err
	}
	attr.ID = uuid.New().String()
	attr.TenantID = tid
	attr.KnowledgeBaseID = kbID
	attr.Code = normalizeCode(attr.Code)
	attr.Name = strings.TrimSpace(attr.Name)
	if attr.Code == "" || attr.Name == "" {
		return nil, apperrors.NewBadRequestError("名称与代码不能为空")
	}
	if attr.ValueType == "" {
		attr.ValueType = types.GraphAttrValueFree
	}
	if attr.DataType == "" {
		attr.DataType = "string"
	}
	if len(attr.EnumValues) == 0 {
		attr.EnumValues = types.JSON([]byte("[]"))
	}
	now := time.Now()
	attr.CreatedAt = now
	attr.UpdatedAt = now
	if err := s.repo.CreateAttribute(ctx, attr); err != nil {
		if isUniqueViolation(err) {
			return nil, apperrors.NewConflictError("属性代码已存在")
		}
		return nil, err
	}
	return attr, nil
}

func (s *graphSchemaService) UpdateAttribute(ctx context.Context, kbID, id string, patch *types.KBGraphAttribute) (*types.KBGraphAttribute, error) {
	tid, err := s.tenantID(ctx)
	if err != nil {
		return nil, err
	}
	existing, err := s.repo.GetAttribute(ctx, tid, kbID, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.NewNotFoundError("属性不存在")
		}
		return nil, err
	}
	if patch.Name != "" {
		existing.Name = strings.TrimSpace(patch.Name)
	}
	if patch.Code != "" {
		existing.Code = normalizeCode(patch.Code)
	}
	if patch.Category != "" {
		existing.Category = patch.Category
	}
	if patch.ValueType != "" {
		existing.ValueType = patch.ValueType
	}
	if patch.DataType != "" {
		existing.DataType = patch.DataType
	}
	existing.IsLabel = patch.IsLabel
	existing.IsStatic = patch.IsStatic
	if patch.Description != "" || patch.Description == "" && patch.Name != "" {
		existing.Description = patch.Description
	}
	if patch.EnumValues != nil {
		existing.EnumValues = patch.EnumValues
	}
	existing.UpdatedAt = time.Now()
	if err := s.repo.UpdateAttribute(ctx, existing); err != nil {
		if isUniqueViolation(err) {
			return nil, apperrors.NewConflictError("属性代码已存在")
		}
		return nil, err
	}
	return existing, nil
}

func (s *graphSchemaService) DeleteAttribute(ctx context.Context, kbID, id string) error {
	tid, err := s.tenantID(ctx)
	if err != nil {
		return err
	}
	return s.repo.DeleteAttribute(ctx, tid, kbID, id)
}

// --- Entities ---

func (s *graphSchemaService) ListEntities(ctx context.Context, kbID string) ([]*types.KBGraphEntity, error) {
	tid, err := s.tenantID(ctx)
	if err != nil {
		return nil, err
	}
	return s.repo.ListEntities(ctx, tid, kbID)
}

func (s *graphSchemaService) CreateEntity(ctx context.Context, kbID string, ent *types.KBGraphEntity) (*types.KBGraphEntity, error) {
	tid, err := s.tenantID(ctx)
	if err != nil {
		return nil, err
	}
	ids, err := parseStringIDs(ent.AttributeIDs)
	if err != nil {
		return nil, apperrors.NewBadRequestError(err.Error())
	}
	if err := s.validateAttributeIDs(ctx, tid, kbID, ids); err != nil {
		return nil, err
	}
	ent.ID = uuid.New().String()
	ent.TenantID = tid
	ent.KnowledgeBaseID = kbID
	ent.Code = normalizeCode(ent.Code)
	ent.Name = strings.TrimSpace(ent.Name)
	if ent.Code == "" || ent.Name == "" {
		return nil, apperrors.NewBadRequestError("名称与代码不能为空")
	}
	ent.AttributeIDs = encodeStringIDs(ids)
	now := time.Now()
	ent.CreatedAt = now
	ent.UpdatedAt = now
	if err := s.repo.CreateEntity(ctx, ent); err != nil {
		if isUniqueViolation(err) {
			return nil, apperrors.NewConflictError("实体代码已存在")
		}
		return nil, err
	}
	return ent, nil
}

func (s *graphSchemaService) UpdateEntity(ctx context.Context, kbID, id string, patch *types.KBGraphEntity) (*types.KBGraphEntity, error) {
	tid, err := s.tenantID(ctx)
	if err != nil {
		return nil, err
	}
	existing, err := s.repo.GetEntity(ctx, tid, kbID, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.NewNotFoundError("实体不存在")
		}
		return nil, err
	}
	if patch.Name != "" {
		existing.Name = strings.TrimSpace(patch.Name)
	}
	if patch.Code != "" {
		existing.Code = normalizeCode(patch.Code)
	}
	if patch.Category != "" {
		existing.Category = patch.Category
	}
	if patch.IdentityKey != "" || patch.Name != "" {
		existing.IdentityKey = patch.IdentityKey
	}
	existing.Description = patch.Description
	if patch.AttributeIDs != nil {
		ids, err := parseStringIDs(patch.AttributeIDs)
		if err != nil {
			return nil, apperrors.NewBadRequestError(err.Error())
		}
		if err := s.validateAttributeIDs(ctx, tid, kbID, ids); err != nil {
			return nil, err
		}
		existing.AttributeIDs = encodeStringIDs(ids)
	}
	existing.UpdatedAt = time.Now()
	if err := s.repo.UpdateEntity(ctx, existing); err != nil {
		if isUniqueViolation(err) {
			return nil, apperrors.NewConflictError("实体代码已存在")
		}
		return nil, err
	}
	return existing, nil
}

func (s *graphSchemaService) DeleteEntity(ctx context.Context, kbID, id string) error {
	tid, err := s.tenantID(ctx)
	if err != nil {
		return err
	}
	return s.repo.DeleteEntity(ctx, tid, kbID, id)
}

// --- Relations ---

func (s *graphSchemaService) ListRelations(ctx context.Context, kbID string) ([]*types.KBGraphRelation, error) {
	tid, err := s.tenantID(ctx)
	if err != nil {
		return nil, err
	}
	return s.repo.ListRelations(ctx, tid, kbID)
}

func (s *graphSchemaService) CreateRelation(ctx context.Context, kbID string, rel *types.KBGraphRelation) (*types.KBGraphRelation, error) {
	tid, err := s.tenantID(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.repo.GetEntity(ctx, tid, kbID, rel.SourceEntityID); err != nil {
		return nil, apperrors.NewBadRequestError("本端实体不存在")
	}
	if _, err := s.repo.GetEntity(ctx, tid, kbID, rel.TargetEntityID); err != nil {
		return nil, apperrors.NewBadRequestError("对端实体不存在")
	}
	ids, err := parseStringIDs(rel.AttributeIDs)
	if err != nil {
		return nil, apperrors.NewBadRequestError(err.Error())
	}
	if err := s.validateAttributeIDs(ctx, tid, kbID, ids); err != nil {
		return nil, err
	}
	rel.ID = uuid.New().String()
	rel.TenantID = tid
	rel.KnowledgeBaseID = kbID
	rel.Code = normalizeCode(rel.Code)
	rel.Name = strings.TrimSpace(rel.Name)
	if rel.Code == "" || rel.Name == "" {
		return nil, apperrors.NewBadRequestError("名称与代码不能为空")
	}
	rel.AttributeIDs = encodeStringIDs(ids)
	now := time.Now()
	rel.CreatedAt = now
	rel.UpdatedAt = now
	if err := s.repo.CreateRelation(ctx, rel); err != nil {
		if isUniqueViolation(err) {
			return nil, apperrors.NewConflictError("关系代码已存在")
		}
		return nil, err
	}
	return rel, nil
}

func (s *graphSchemaService) UpdateRelation(ctx context.Context, kbID, id string, patch *types.KBGraphRelation) (*types.KBGraphRelation, error) {
	tid, err := s.tenantID(ctx)
	if err != nil {
		return nil, err
	}
	existing, err := s.repo.GetRelation(ctx, tid, kbID, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.NewNotFoundError("关系不存在")
		}
		return nil, err
	}
	if patch.Name != "" {
		existing.Name = strings.TrimSpace(patch.Name)
	}
	if patch.Code != "" {
		existing.Code = normalizeCode(patch.Code)
	}
	if patch.Category != "" {
		existing.Category = patch.Category
	}
	if patch.SourceEntityID != "" {
		if _, err := s.repo.GetEntity(ctx, tid, kbID, patch.SourceEntityID); err != nil {
			return nil, apperrors.NewBadRequestError("本端实体不存在")
		}
		existing.SourceEntityID = patch.SourceEntityID
	}
	if patch.TargetEntityID != "" {
		if _, err := s.repo.GetEntity(ctx, tid, kbID, patch.TargetEntityID); err != nil {
			return nil, apperrors.NewBadRequestError("对端实体不存在")
		}
		existing.TargetEntityID = patch.TargetEntityID
	}
	existing.Description = patch.Description
	if patch.AttributeIDs != nil {
		ids, err := parseStringIDs(patch.AttributeIDs)
		if err != nil {
			return nil, apperrors.NewBadRequestError(err.Error())
		}
		if err := s.validateAttributeIDs(ctx, tid, kbID, ids); err != nil {
			return nil, err
		}
		existing.AttributeIDs = encodeStringIDs(ids)
	}
	existing.UpdatedAt = time.Now()
	if err := s.repo.UpdateRelation(ctx, existing); err != nil {
		if isUniqueViolation(err) {
			return nil, apperrors.NewConflictError("关系代码已存在")
		}
		return nil, err
	}
	return existing, nil
}

func (s *graphSchemaService) DeleteRelation(ctx context.Context, kbID, id string) error {
	tid, err := s.tenantID(ctx)
	if err != nil {
		return err
	}
	return s.repo.DeleteRelation(ctx, tid, kbID, id)
}

func (s *graphSchemaService) LoadBundle(ctx context.Context, kbID string) (*types.GraphSchemaBundle, error) {
	tid, err := s.tenantID(ctx)
	if err != nil {
		return nil, err
	}
	return s.repo.LoadBundle(ctx, tid, kbID)
}

// CompileExtractHints builds extract prompt constraints from KB schema.
func (s *graphSchemaService) CompileExtractHints(ctx context.Context, kbID string) (*types.GraphExtractHints, error) {
	tid, err := s.tenantID(ctx)
	if err != nil {
		return nil, err
	}
	bundle, err := s.repo.LoadBundle(ctx, tid, kbID)
	if err != nil {
		return nil, err
	}
	return CompileExtractHintsFromBundle(bundle), nil
}

// CompileExtractHintsFromBundle is a pure helper used by extract and tests.
func CompileExtractHintsFromBundle(bundle *types.GraphSchemaBundle) *types.GraphExtractHints {
	hints := &types.GraphExtractHints{}
	if bundle == nil {
		return hints
	}
	var b strings.Builder
	b.WriteString("Follow the knowledge-base schema below. Only extract entities, relations, and attributes that match these definitions.\n")
	if len(bundle.Entities) > 0 {
		b.WriteString("Allowed entity types:\n")
		for _, e := range bundle.Entities {
			hints.EntityCodes = append(hints.EntityCodes, e.Code)
			line := fmt.Sprintf("- %s (code=%s, category=%s)", e.Name, e.Code, e.Category)
			if e.Description != "" {
				line += ": " + e.Description
			}
			b.WriteString(line + "\n")
		}
	}
	if len(bundle.Relations) > 0 {
		b.WriteString("Allowed relation types (use these as relation tags):\n")
		for _, r := range bundle.Relations {
			tag := r.Name
			if tag == "" {
				tag = r.Code
			}
			hints.Tags = append(hints.Tags, tag)
			hints.RelationCodes = append(hints.RelationCodes, r.Code)
			line := fmt.Sprintf("- %s (code=%s)", r.Name, r.Code)
			if r.Description != "" {
				line += ": " + r.Description
			}
			b.WriteString(line + "\n")
		}
	}
	if len(bundle.Attributes) > 0 {
		b.WriteString("Allowed attributes:\n")
		for _, a := range bundle.Attributes {
			hints.AttributeCodes = append(hints.AttributeCodes, a.Code)
			line := fmt.Sprintf("- %s (code=%s, value_type=%s)", a.Name, a.Code, a.ValueType)
			if a.Description != "" {
				line += ": " + a.Description
			}
			b.WriteString(line + "\n")
		}
	}
	if len(bundle.Entities) > 0 || len(bundle.Relations) > 0 || len(bundle.Attributes) > 0 {
		hints.CustomInstructions = strings.TrimSpace(b.String())
	}
	return hints
}

// --- Ingest tasks ---

func (s *graphSchemaService) ListIngestTasks(ctx context.Context, kbID string) ([]*types.KBGraphIngestTask, error) {
	tid, err := s.tenantID(ctx)
	if err != nil {
		return nil, err
	}
	return s.repo.ListIngestTasks(ctx, tid, kbID)
}

func (s *graphSchemaService) GetIngestTask(ctx context.Context, kbID, id string) (*types.KBGraphIngestTask, error) {
	tid, err := s.tenantID(ctx)
	if err != nil {
		return nil, err
	}
	task, err := s.repo.GetIngestTask(ctx, tid, kbID, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.NewNotFoundError("任务不存在")
		}
		return nil, err
	}
	return task, nil
}

func (s *graphSchemaService) CreateIngestTask(ctx context.Context, kbID string, task *types.KBGraphIngestTask) (*types.KBGraphIngestTask, error) {
	tid, err := s.tenantID(ctx)
	if err != nil {
		return nil, err
	}
	switch task.TaskKind {
	case types.GraphIngestKindEntity, types.GraphIngestKindRelation, types.GraphIngestKindAttribute:
	default:
		return nil, apperrors.NewBadRequestError("无效的任务类型")
	}
	if task.SourceType == "" {
		task.SourceType = types.GraphIngestSourceTableMapping
	}
	if len(task.SourceRef) == 0 {
		task.SourceRef = types.JSON([]byte("{}"))
	}
	var ref types.GraphIngestSourceRef
	if err := json.Unmarshal(task.SourceRef, &ref); err != nil {
		return nil, apperrors.NewBadRequestError("source_ref 格式无效").WithDetails(err.Error())
	}
	if len(ref.Rows) == 0 {
		return nil, apperrors.NewBadRequestError("请提供至少一行表映射数据")
	}
	switch task.TaskKind {
	case types.GraphIngestKindEntity:
		if ref.EntityDefID == "" {
			return nil, apperrors.NewBadRequestError("实体入库需指定实体定义")
		}
		if _, err := s.repo.GetEntity(ctx, tid, kbID, ref.EntityDefID); err != nil {
			return nil, apperrors.NewBadRequestError("实体定义不存在")
		}
	case types.GraphIngestKindRelation:
		if ref.RelationDefID == "" {
			return nil, apperrors.NewBadRequestError("关系入库需指定关系定义")
		}
		if _, err := s.repo.GetRelation(ctx, tid, kbID, ref.RelationDefID); err != nil {
			return nil, apperrors.NewBadRequestError("关系定义不存在")
		}
	case types.GraphIngestKindAttribute:
		if ref.AttributeDefID != "" {
			if _, err := s.repo.GetAttribute(ctx, tid, kbID, ref.AttributeDefID); err != nil {
				return nil, apperrors.NewBadRequestError("属性定义不存在")
			}
		}
	}
	task.ID = uuid.New().String()
	task.TenantID = tid
	task.KnowledgeBaseID = kbID
	task.Status = types.GraphIngestStatusPending
	task.Progress = 0
	task.ErrorMessage = ""
	now := time.Now()
	task.CreatedAt = now
	task.UpdatedAt = now
	if err := s.repo.CreateIngestTask(ctx, task); err != nil {
		return nil, err
	}
	if err := enqueueGraphSchemaIngest(ctx, s.taskClient, tid, kbID, task.ID); err != nil {
		logger.Warnf(ctx, "failed to enqueue graph schema ingest task %s: %v", task.ID, err)
		task.Status = types.GraphIngestStatusFailed
		task.ErrorMessage = "enqueue failed: " + err.Error()
		fin := time.Now()
		task.FinishedAt = &fin
		task.UpdatedAt = fin
		_ = s.repo.UpdateIngestTask(ctx, task)
		return task, nil
	}
	return task, nil
}

func (s *graphSchemaService) CancelIngestTask(ctx context.Context, kbID, id string) (*types.KBGraphIngestTask, error) {
	tid, err := s.tenantID(ctx)
	if err != nil {
		return nil, err
	}
	task, err := s.repo.GetIngestTask(ctx, tid, kbID, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.NewNotFoundError("任务不存在")
		}
		return nil, err
	}
	if task.Status == types.GraphIngestStatusCancelled {
		return task, nil
	}
	if task.Status == types.GraphIngestStatusSucceeded || task.Status == types.GraphIngestStatusFailed {
		return nil, apperrors.NewBadRequestError("任务已结束，无法取消")
	}
	now := time.Now()
	task.Status = types.GraphIngestStatusCancelled
	task.FinishedAt = &now
	task.UpdatedAt = now
	if err := s.repo.UpdateIngestTask(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *graphSchemaService) RetryIngestTask(ctx context.Context, kbID, id string) (*types.KBGraphIngestTask, error) {
	tid, err := s.tenantID(ctx)
	if err != nil {
		return nil, err
	}
	task, err := s.repo.GetIngestTask(ctx, tid, kbID, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.NewNotFoundError("任务不存在")
		}
		return nil, err
	}
	if task.Status != types.GraphIngestStatusFailed && task.Status != types.GraphIngestStatusCancelled {
		return nil, apperrors.NewBadRequestError("仅失败或已取消的任务可重试")
	}
	now := time.Now()
	task.Status = types.GraphIngestStatusPending
	task.Progress = 0
	task.ErrorMessage = ""
	task.StartedAt = nil
	task.FinishedAt = nil
	task.UpdatedAt = now
	if err := s.repo.UpdateIngestTask(ctx, task); err != nil {
		return nil, err
	}
	if err := enqueueGraphSchemaIngest(ctx, s.taskClient, tid, kbID, task.ID); err != nil {
		logger.Warnf(ctx, "failed to re-enqueue graph schema ingest task %s: %v", task.ID, err)
		task.Status = types.GraphIngestStatusFailed
		task.ErrorMessage = "enqueue failed: " + err.Error()
		fin := time.Now()
		task.FinishedAt = &fin
		task.UpdatedAt = fin
		_ = s.repo.UpdateIngestTask(ctx, task)
		return nil, apperrors.NewInternalServerError("重新入队失败: " + err.Error())
	}
	return task, nil
}

// GraphSchemaIngestPayload is the asynq payload for TypeGraphSchemaIngest.
type GraphSchemaIngestPayload struct {
	TenantID        uint64 `json:"tenant_id"`
	KnowledgeBaseID string `json:"knowledge_base_id"`
	TaskID          string `json:"task_id"`
}

func enqueueGraphSchemaIngest(ctx context.Context, client interfaces.TaskEnqueuer, tenantID uint64, kbID, taskID string) error {
	if client == nil {
		return fmt.Errorf("task client unavailable")
	}
	payload, err := json.Marshal(GraphSchemaIngestPayload{
		TenantID: tenantID, KnowledgeBaseID: kbID, TaskID: taskID,
	})
	if err != nil {
		return err
	}
	task := asynq.NewTask(types.TypeGraphSchemaIngest, payload)
	// MaxRetry(0): failures are persisted as failed; users retry explicitly via
	// RetryIngestTask to avoid status flicker from queue auto-retries.
	_, err = client.Enqueue(task, asynq.Queue(types.QueueGraph), asynq.MaxRetry(0))
	return err
}

func (s *graphSchemaService) ProcessIngestTask(ctx context.Context, t *asynq.Task) error {
	var p GraphSchemaIngestPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("invalid payload: %w", err)
	}
	task, err := s.repo.GetIngestTask(ctx, p.TenantID, p.KnowledgeBaseID, p.TaskID)
	if err != nil {
		return err
	}
	switch task.Status {
	case types.GraphIngestStatusCancelled, types.GraphIngestStatusSucceeded:
		return nil
	case types.GraphIngestStatusFailed:
		// Stale queue message after a previous failure; ignore unless retried.
		return nil
	}
	now := time.Now()
	task.Status = types.GraphIngestStatusRunning
	task.StartedAt = &now
	task.Progress = 10
	task.ErrorMessage = ""
	task.UpdatedAt = now
	_ = s.repo.UpdateIngestTask(ctx, task)

	if err := s.runIngest(ctx, task); err != nil {
		fin := time.Now()
		task.Status = types.GraphIngestStatusFailed
		task.ErrorMessage = err.Error()
		task.FinishedAt = &fin
		task.UpdatedAt = fin
		_ = s.repo.UpdateIngestTask(ctx, task)
		// Persist failure without returning err so asynq won't auto-retry.
		logger.Warnf(ctx, "graph schema ingest task %s failed: %v", task.ID, err)
		return nil
	}

	// Re-check cancel
	fresh, _ := s.repo.GetIngestTask(ctx, p.TenantID, p.KnowledgeBaseID, p.TaskID)
	if fresh != nil && fresh.Status == types.GraphIngestStatusCancelled {
		return nil
	}
	fin := time.Now()
	task.Status = types.GraphIngestStatusSucceeded
	task.Progress = 100
	task.FinishedAt = &fin
	task.UpdatedAt = fin
	task.ErrorMessage = ""
	return s.repo.UpdateIngestTask(ctx, task)
}

func (s *graphSchemaService) runIngest(ctx context.Context, task *types.KBGraphIngestTask) error {
	var ref types.GraphIngestSourceRef
	if len(task.SourceRef) > 0 {
		if err := json.Unmarshal(task.SourceRef, &ref); err != nil {
			return fmt.Errorf("invalid source_ref: %w", err)
		}
	}
	if len(ref.Rows) == 0 {
		return fmt.Errorf("source_ref.rows 为空，请提供表映射行数据")
	}
	graph := &types.GraphData{}
	switch task.TaskKind {
	case types.GraphIngestKindEntity:
		ent, err := s.repo.GetEntity(ctx, task.TenantID, task.KnowledgeBaseID, ref.EntityDefID)
		if err != nil {
			return fmt.Errorf("实体定义不存在: %w", err)
		}
		nameCol := ref.ColumnMapping["name"]
		if nameCol == "" {
			nameCol = "name"
		}
		for _, row := range ref.Rows {
			name := strings.TrimSpace(row[nameCol])
			if name == "" {
				continue
			}
			attrs := []string{}
			for col, val := range row {
				if col == nameCol || strings.TrimSpace(val) == "" {
					continue
				}
				attrs = append(attrs, col+":"+val)
			}
			graph.Node = append(graph.Node, &types.GraphNode{
				Name:           name,
				Attributes:     attrs,
				EntityCode:     ent.Code,
				EntityCategory: ent.Category,
			})
		}
	case types.GraphIngestKindRelation:
		rel, err := s.repo.GetRelation(ctx, task.TenantID, task.KnowledgeBaseID, ref.RelationDefID)
		if err != nil {
			return fmt.Errorf("关系定义不存在: %w", err)
		}
		srcCol := ref.ColumnMapping["source"]
		tgtCol := ref.ColumnMapping["target"]
		if srcCol == "" {
			srcCol = "source"
		}
		if tgtCol == "" {
			tgtCol = "target"
		}
		relType := rel.Name
		if relType == "" {
			relType = rel.Code
		}
		for _, row := range ref.Rows {
			src := strings.TrimSpace(row[srcCol])
			tgt := strings.TrimSpace(row[tgtCol])
			if src == "" || tgt == "" {
				continue
			}
			graph.Node = append(graph.Node,
				&types.GraphNode{Name: src},
				&types.GraphNode{Name: tgt},
			)
			graph.Relation = append(graph.Relation, &types.GraphRelation{
				Node1:        src,
				Node2:        tgt,
				Type:         relType,
				RelationCode: rel.Code,
			})
		}
	case types.GraphIngestKindAttribute:
		// Attribute ingest attaches attribute values onto named entities.
		nameCol := ref.ColumnMapping["name"]
		if nameCol == "" {
			nameCol = "name"
		}
		attrCol := ref.ColumnMapping["attribute"]
		if attrCol == "" {
			attrCol = "attribute"
		}
		attrPrefix := ""
		if ref.AttributeDefID != "" {
			if attr, err := s.repo.GetAttribute(ctx, task.TenantID, task.KnowledgeBaseID, ref.AttributeDefID); err == nil && attr != nil {
				if attr.Code != "" {
					attrPrefix = attr.Code + ":"
				} else if attr.Name != "" {
					attrPrefix = attr.Name + ":"
				}
			}
		}
		for _, row := range ref.Rows {
			name := strings.TrimSpace(row[nameCol])
			attrVal := strings.TrimSpace(row[attrCol])
			if name == "" || attrVal == "" {
				continue
			}
			if attrPrefix != "" && !strings.Contains(attrVal, ":") {
				attrVal = attrPrefix + attrVal
			}
			graph.Node = append(graph.Node, &types.GraphNode{
				Name:       name,
				Attributes: []string{attrVal},
			})
		}
	default:
		return fmt.Errorf("unsupported task_kind: %s", task.TaskKind)
	}
	if len(graph.Node) == 0 && len(graph.Relation) == 0 {
		return fmt.Errorf("映射结果为空，请检查列映射与行数据")
	}
	if s.graphEngine == nil {
		return fmt.Errorf("graph engine unavailable")
	}
	ns := types.NameSpace{KnowledgeBase: task.KnowledgeBaseID, Knowledge: ref.KnowledgeID}
	if ns.Knowledge == "" {
		ns.Knowledge = "schema-ingest:" + task.ID
	}
	task.Progress = 60
	task.UpdatedAt = time.Now()
	_ = s.repo.UpdateIngestTask(ctx, task)
	if err := s.graphEngine.AddGraph(ctx, ns, []*types.GraphData{graph}); err != nil {
		return err
	}
	task.Progress = 90
	task.UpdatedAt = time.Now()
	_ = s.repo.UpdateIngestTask(ctx, task)
	return nil
}
