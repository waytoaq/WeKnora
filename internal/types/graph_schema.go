package types

import "time"

// Graph attribute value types.
const (
	GraphAttrValueEnum   = "enum"
	GraphAttrValueSingle = "single"
	GraphAttrValueFree   = "free"
)

// Graph entity categories (人/事/地/物/组织).
const (
	GraphEntityCategoryPerson       = "person"
	GraphEntityCategoryEvent        = "event"
	GraphEntityCategoryPlace        = "place"
	GraphEntityCategoryThing        = "thing"
	GraphEntityCategoryOrganization = "organization"
)

// Graph ingest task kinds and statuses.
const (
	GraphIngestKindEntity    = "entity"
	GraphIngestKindRelation  = "relation"
	GraphIngestKindAttribute = "attribute"

	GraphIngestStatusPending   = "pending"
	GraphIngestStatusRunning   = "running"
	GraphIngestStatusSucceeded = "succeeded"
	GraphIngestStatusFailed    = "failed"
	GraphIngestStatusCancelled = "cancelled"

	GraphIngestSourceTableMapping = "table_mapping"
	GraphIngestSourceManual       = "manual"
)

// KBGraphAttribute is a knowledge-base scoped attribute definition.
type KBGraphAttribute struct {
	ID              string    `json:"id" gorm:"type:varchar(36);primaryKey"`
	TenantID        uint64    `json:"tenant_id"`
	KnowledgeBaseID string    `json:"knowledge_base_id" gorm:"type:varchar(36);index"`
	Category        string    `json:"category" gorm:"type:varchar(64)"`
	Name            string    `json:"name" gorm:"type:varchar(200)"`
	Code            string    `json:"code" gorm:"type:varchar(128)"`
	ValueType       string    `json:"value_type" gorm:"type:varchar(32)"` // enum|single|free
	DataType        string    `json:"data_type" gorm:"type:varchar(32)"`
	IsLabel         bool      `json:"is_label"`
	IsStatic        bool      `json:"is_static"`
	Description     string    `json:"description" gorm:"type:text"`
	EnumValues      JSON      `json:"enum_values" gorm:"type:jsonb;default:'[]'"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (KBGraphAttribute) TableName() string { return "kb_graph_attributes" }

// KBGraphEntity is a knowledge-base scoped entity type definition.
type KBGraphEntity struct {
	ID              string    `json:"id" gorm:"type:varchar(36);primaryKey"`
	TenantID        uint64    `json:"tenant_id"`
	KnowledgeBaseID string    `json:"knowledge_base_id" gorm:"type:varchar(36);index"`
	Category        string    `json:"category" gorm:"type:varchar(32)"`
	Name            string    `json:"name" gorm:"type:varchar(200)"`
	Code            string    `json:"code" gorm:"type:varchar(128)"`
	IdentityKey     string    `json:"identity_key" gorm:"type:varchar(128)"`
	Description     string    `json:"description" gorm:"type:text"`
	AttributeIDs    JSON      `json:"attribute_ids" gorm:"type:jsonb;default:'[]'"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (KBGraphEntity) TableName() string { return "kb_graph_entities" }

// KBGraphRelation is a knowledge-base scoped relation type definition.
type KBGraphRelation struct {
	ID              string    `json:"id" gorm:"type:varchar(36);primaryKey"`
	TenantID        uint64    `json:"tenant_id"`
	KnowledgeBaseID string    `json:"knowledge_base_id" gorm:"type:varchar(36);index"`
	Category        string    `json:"category" gorm:"type:varchar(64)"`
	Name            string    `json:"name" gorm:"type:varchar(200)"`
	Code            string    `json:"code" gorm:"type:varchar(128)"`
	SourceEntityID  string    `json:"source_entity_id" gorm:"type:varchar(36)"`
	TargetEntityID  string    `json:"target_entity_id" gorm:"type:varchar(36)"`
	Description     string    `json:"description" gorm:"type:text"`
	AttributeIDs    JSON      `json:"attribute_ids" gorm:"type:jsonb;default:'[]'"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (KBGraphRelation) TableName() string { return "kb_graph_relations" }

// KBGraphIngestTask tracks structured table→graph ingest jobs.
type KBGraphIngestTask struct {
	ID              string     `json:"id" gorm:"type:varchar(36);primaryKey"`
	TenantID        uint64     `json:"tenant_id"`
	KnowledgeBaseID string     `json:"knowledge_base_id" gorm:"type:varchar(36);index"`
	TaskKind        string     `json:"task_kind" gorm:"type:varchar(32)"` // entity|relation|attribute
	Status          string     `json:"status" gorm:"type:varchar(32)"`
	SourceType      string     `json:"source_type" gorm:"type:varchar(32)"`
	SourceRef       JSON       `json:"source_ref" gorm:"type:jsonb;default:'{}'"`
	Progress        int        `json:"progress"`
	ErrorMessage    string     `json:"error_message" gorm:"type:text"`
	StartedAt       *time.Time `json:"started_at"`
	FinishedAt      *time.Time `json:"finished_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func (KBGraphIngestTask) TableName() string { return "kb_graph_ingest_tasks" }

// GraphSchemaBundle is the full schema snapshot for a knowledge base.
type GraphSchemaBundle struct {
	Attributes []*KBGraphAttribute `json:"attributes"`
	Entities   []*KBGraphEntity    `json:"entities"`
	Relations  []*KBGraphRelation  `json:"relations"`
}

// GraphExtractHints are compiled constraints injected into chunk extract prompts.
type GraphExtractHints struct {
	Tags                []string
	CustomInstructions  string
	EntityCodes         []string
	RelationCodes       []string
	AttributeCodes      []string
}

// GraphIngestSourceRef is the JSON shape stored in KBGraphIngestTask.SourceRef.
type GraphIngestSourceRef struct {
	KnowledgeID    string              `json:"knowledge_id,omitempty"`
	EntityDefID    string              `json:"entity_def_id,omitempty"`
	RelationDefID  string              `json:"relation_def_id,omitempty"`
	AttributeDefID string              `json:"attribute_def_id,omitempty"`
	ColumnMapping  map[string]string   `json:"column_mapping,omitempty"`
	Rows           []map[string]string `json:"rows,omitempty"`
}
