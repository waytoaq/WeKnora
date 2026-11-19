<template>
  <div class="graph-schema-settings">
    <div class="section-header">
      <h2>{{ t('graphSchema.title') }}</h2>
      <p class="section-description">{{ t('graphSchema.description') }}</p>
    </div>

    <t-tabs v-model="activeTab" class="schema-tabs">
      <t-tab-panel value="attributes" :label="t('graphSchema.tabs.attributes')">
        <div class="toolbar">
          <t-button theme="primary" size="small" @click="openAttrDialog()">
            {{ t('graphSchema.actions.create') }}
          </t-button>
          <t-button variant="outline" size="small" :loading="loadingAttrs" @click="loadAttributes">
            {{ t('common.refresh') }}
          </t-button>
        </div>
        <t-table
          row-key="id"
          :data="attributes"
          :columns="attrColumns"
          :loading="loadingAttrs"
          size="small"
          hover
        >
          <template #op="{ row }">
            <t-space>
              <t-link theme="primary" @click="openAttrDialog(row)">{{ t('common.edit') }}</t-link>
              <t-popconfirm :content="t('graphSchema.confirmDelete')" @confirm="removeAttribute(row)">
                <t-link theme="danger">{{ t('common.delete') }}</t-link>
              </t-popconfirm>
            </t-space>
          </template>
        </t-table>
      </t-tab-panel>

      <t-tab-panel value="entities" :label="t('graphSchema.tabs.entities')">
        <div class="toolbar">
          <t-button theme="primary" size="small" @click="openEntityDialog()">
            {{ t('graphSchema.actions.create') }}
          </t-button>
          <t-button variant="outline" size="small" :loading="loadingEntities" @click="loadEntities">
            {{ t('common.refresh') }}
          </t-button>
        </div>
        <t-table
          row-key="id"
          :data="entities"
          :columns="entityColumns"
          :loading="loadingEntities"
          size="small"
          hover
        >
          <template #category="{ row }">
            {{ categoryLabel(row.category) }}
          </template>
          <template #attrs="{ row }">
            {{ formatAttrNames(row.attribute_ids) }}
          </template>
          <template #op="{ row }">
            <t-space>
              <t-link theme="primary" @click="openEntityDialog(row)">{{ t('common.edit') }}</t-link>
              <t-popconfirm :content="t('graphSchema.confirmDelete')" @confirm="removeEntity(row)">
                <t-link theme="danger">{{ t('common.delete') }}</t-link>
              </t-popconfirm>
            </t-space>
          </template>
        </t-table>
      </t-tab-panel>

      <t-tab-panel value="relations" :label="t('graphSchema.tabs.relations')">
        <div class="toolbar">
          <t-button theme="primary" size="small" @click="openRelationDialog()">
            {{ t('graphSchema.actions.create') }}
          </t-button>
          <t-button variant="outline" size="small" :loading="loadingRelations" @click="loadRelations">
            {{ t('common.refresh') }}
          </t-button>
        </div>
        <t-table
          row-key="id"
          :data="relations"
          :columns="relationColumns"
          :loading="loadingRelations"
          size="small"
          hover
        >
          <template #source="{ row }">{{ entityName(row.source_entity_id) }}</template>
          <template #target="{ row }">{{ entityName(row.target_entity_id) }}</template>
          <template #op="{ row }">
            <t-space>
              <t-link theme="primary" @click="openRelationDialog(row)">{{ t('common.edit') }}</t-link>
              <t-popconfirm :content="t('graphSchema.confirmDelete')" @confirm="removeRelation(row)">
                <t-link theme="danger">{{ t('common.delete') }}</t-link>
              </t-popconfirm>
            </t-space>
          </template>
        </t-table>
      </t-tab-panel>

      <t-tab-panel value="tasks" :label="t('graphSchema.tabs.tasks')">
        <div class="toolbar">
          <t-button theme="primary" size="small" @click="openTaskDialog">
            {{ t('graphSchema.tasks.create') }}
          </t-button>
          <t-select
            v-model="taskStatusFilter"
            :options="taskStatusFilterOptions"
            size="small"
            style="width: 140px"
            clearable
            :placeholder="t('graphSchema.tasks.filterAll')"
          />
          <t-button variant="outline" size="small" :loading="loadingTasks" @click="loadTasks">
            {{ t('common.refresh') }}
          </t-button>
        </div>
        <t-table
          row-key="id"
          :data="filteredTasks"
          :columns="taskColumns"
          :loading="loadingTasks"
          size="small"
          hover
          :empty="t('graphSchema.tasks.empty')"
        >
          <template #task_kind="{ row }">
            {{ kindLabel(row.task_kind) }}
          </template>
          <template #status="{ row }">
            <t-tag :theme="statusTheme(row.status)" variant="light">{{ statusLabel(row.status) }}</t-tag>
          </template>
          <template #progress="{ row }">
            <div class="progress-cell">
              <t-progress
                :percentage="Number(row.progress ?? 0)"
                :status="progressStatus(row.status)"
                size="small"
              />
            </div>
          </template>
          <template #created_at="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
          <template #error="{ row }">
            <t-tooltip v-if="row.error_message" :content="row.error_message">
              <span class="err err-ellipsis">{{ row.error_message }}</span>
            </t-tooltip>
            <span v-else>—</span>
          </template>
          <template #op="{ row }">
            <div class="op-links">
              <t-link theme="primary" @click="openTaskDetail(row)">
                {{ t('graphSchema.tasks.detail') }}
              </t-link>
              <t-link
                v-if="row.status === 'pending' || row.status === 'running'"
                theme="warning"
                @click="cancelTask(row)"
              >
                {{ t('graphSchema.tasks.cancel') }}
              </t-link>
              <t-link
                v-if="row.status === 'failed' || row.status === 'cancelled'"
                theme="success"
                @click="retryTask(row)"
              >
                {{ t('graphSchema.tasks.retry') }}
              </t-link>
            </div>
          </template>
        </t-table>
      </t-tab-panel>
    </t-tabs>

    <!-- Attribute dialog -->
    <t-dialog
      v-model:visible="attrDialogVisible"
      :header="attrForm.id ? t('graphSchema.attr.edit') : t('graphSchema.attr.create')"
      :confirm-btn="{ content: t('common.save'), loading: saving }"
      @confirm="saveAttribute"
    >
      <t-form :data="attrForm" label-width="96px">
        <t-form-item :label="t('graphSchema.fields.name')">
          <t-input v-model="attrForm.name" />
        </t-form-item>
        <t-form-item :label="t('graphSchema.fields.code')">
          <t-input v-model="attrForm.code" :disabled="!!attrForm.id" />
        </t-form-item>
        <t-form-item :label="t('graphSchema.fields.category')">
          <t-input v-model="attrForm.category" />
        </t-form-item>
        <t-form-item :label="t('graphSchema.attr.valueType')">
          <t-select v-model="attrForm.value_type" :options="valueTypeOptions" />
        </t-form-item>
        <t-form-item :label="t('graphSchema.attr.dataType')">
          <t-input v-model="attrForm.data_type" />
        </t-form-item>
        <t-form-item :label="t('graphSchema.attr.isLabel')">
          <t-switch v-model="attrForm.is_label" />
        </t-form-item>
        <t-form-item :label="t('graphSchema.attr.isStatic')">
          <t-switch v-model="attrForm.is_static" />
        </t-form-item>
        <t-form-item v-if="attrForm.value_type === 'enum'" :label="t('graphSchema.attr.enumValues')">
          <t-textarea v-model="enumValuesText" :placeholder="t('graphSchema.attr.enumPlaceholder')" />
        </t-form-item>
        <t-form-item :label="t('graphSchema.fields.description')">
          <t-textarea v-model="attrForm.description" />
        </t-form-item>
      </t-form>
    </t-dialog>

    <!-- Entity dialog -->
    <t-dialog
      v-model:visible="entityDialogVisible"
      :header="entityForm.id ? t('graphSchema.entity.edit') : t('graphSchema.entity.create')"
      :confirm-btn="{ content: t('common.save'), loading: saving }"
      @confirm="saveEntity"
    >
      <t-form :data="entityForm" label-width="96px">
        <t-form-item :label="t('graphSchema.fields.name')">
          <t-input v-model="entityForm.name" />
        </t-form-item>
        <t-form-item :label="t('graphSchema.fields.code')">
          <t-input v-model="entityForm.code" :disabled="!!entityForm.id" />
        </t-form-item>
        <t-form-item :label="t('graphSchema.fields.category')">
          <t-select v-model="entityForm.category" :options="entityCategoryOptions" />
        </t-form-item>
        <t-form-item :label="t('graphSchema.entity.identityKey')">
          <t-input v-model="entityForm.identity_key" />
        </t-form-item>
        <t-form-item :label="t('graphSchema.fields.attributes')">
          <t-select
            v-model="entityForm.attribute_ids"
            multiple
            filterable
            :options="attributeOptions"
            :placeholder="t('graphSchema.fields.selectAttributes')"
          />
        </t-form-item>
        <t-form-item :label="t('graphSchema.fields.description')">
          <t-textarea v-model="entityForm.description" />
        </t-form-item>
      </t-form>
    </t-dialog>

    <!-- Relation dialog -->
    <t-dialog
      v-model:visible="relationDialogVisible"
      :header="relationForm.id ? t('graphSchema.relation.edit') : t('graphSchema.relation.create')"
      :confirm-btn="{ content: t('common.save'), loading: saving }"
      @confirm="saveRelation"
    >
      <t-form :data="relationForm" label-width="96px">
        <t-form-item :label="t('graphSchema.fields.name')">
          <t-input v-model="relationForm.name" />
        </t-form-item>
        <t-form-item :label="t('graphSchema.fields.code')">
          <t-input v-model="relationForm.code" :disabled="!!relationForm.id" />
        </t-form-item>
        <t-form-item :label="t('graphSchema.fields.category')">
          <t-input v-model="relationForm.category" />
        </t-form-item>
        <t-form-item :label="t('graphSchema.relation.source')">
          <t-select v-model="relationForm.source_entity_id" :options="entityOptions" filterable />
        </t-form-item>
        <t-form-item :label="t('graphSchema.relation.target')">
          <t-select v-model="relationForm.target_entity_id" :options="entityOptions" filterable />
        </t-form-item>
        <t-form-item :label="t('graphSchema.fields.attributes')">
          <t-select
            v-model="relationForm.attribute_ids"
            multiple
            filterable
            :options="attributeOptions"
          />
        </t-form-item>
        <t-form-item :label="t('graphSchema.fields.description')">
          <t-textarea v-model="relationForm.description" />
        </t-form-item>
      </t-form>
    </t-dialog>

    <!-- Task create dialog -->
    <t-dialog
      v-model:visible="taskDialogVisible"
      :header="t('graphSchema.tasks.create')"
      :confirm-btn="{ content: t('graphSchema.tasks.submit'), loading: saving }"
      width="640px"
      @confirm="submitTask"
    >
      <t-form label-width="110px">
        <t-form-item :label="t('graphSchema.tasks.kind')">
          <t-select v-model="taskForm.task_kind" :options="taskKindOptions" />
        </t-form-item>
        <t-form-item v-if="taskForm.task_kind === 'entity'" :label="t('graphSchema.tasks.entityDef')">
          <t-select v-model="taskForm.entity_def_id" :options="entityOptions" filterable />
        </t-form-item>
        <t-form-item v-if="taskForm.task_kind === 'relation'" :label="t('graphSchema.tasks.relationDef')">
          <t-select v-model="taskForm.relation_def_id" :options="relationOptions" filterable />
        </t-form-item>
        <t-form-item v-if="taskForm.task_kind === 'attribute'" :label="t('graphSchema.tasks.attributeDefOptional')">
          <t-select
            v-model="taskForm.attribute_def_id"
            :options="attributeOptions"
            filterable
            clearable
          />
        </t-form-item>
        <t-form-item :label="t('graphSchema.tasks.nameCol')">
          <t-input v-model="taskForm.name_col" :placeholder="taskForm.task_kind === 'relation' ? 'source' : 'name'" />
        </t-form-item>
        <t-form-item v-if="taskForm.task_kind === 'relation'" :label="t('graphSchema.tasks.targetCol')">
          <t-input v-model="taskForm.target_col" placeholder="target" />
        </t-form-item>
        <t-form-item v-if="taskForm.task_kind === 'attribute'" :label="t('graphSchema.tasks.attrCol')">
          <t-input v-model="taskForm.attr_col" placeholder="attribute" />
        </t-form-item>
        <t-form-item :label="t('graphSchema.tasks.rows')">
          <t-textarea
            v-model="taskRowsText"
            :autosize="{ minRows: 6, maxRows: 12 }"
            :placeholder="t('graphSchema.tasks.rowsPlaceholder')"
          />
        </t-form-item>
      </t-form>
    </t-dialog>

    <!-- Task detail drawer -->
    <t-drawer
      v-model:visible="taskDetailVisible"
      :header="t('graphSchema.tasks.detail')"
      size="480px"
      :footer="false"
    >
      <template v-if="taskDetail">
        <t-descriptions :column="1" size="small" bordered>
          <t-descriptions-item :label="t('graphSchema.tasks.kind')">
            {{ kindLabel(taskDetail.task_kind) }}
          </t-descriptions-item>
          <t-descriptions-item :label="t('graphSchema.tasks.status')">
            <t-tag :theme="statusTheme(taskDetail.status)" variant="light">
              {{ statusLabel(taskDetail.status) }}
            </t-tag>
          </t-descriptions-item>
          <t-descriptions-item :label="t('graphSchema.tasks.progress')">
            <t-progress
              :percentage="Number(taskDetail.progress ?? 0)"
              :status="progressStatus(taskDetail.status)"
              size="small"
            />
          </t-descriptions-item>
          <t-descriptions-item :label="t('graphSchema.tasks.createdAt')">
            {{ formatTime(taskDetail.created_at) }}
          </t-descriptions-item>
          <t-descriptions-item :label="t('graphSchema.tasks.startedAt')">
            {{ formatTime(taskDetail.started_at) }}
          </t-descriptions-item>
          <t-descriptions-item :label="t('graphSchema.tasks.finishedAt')">
            {{ formatTime(taskDetail.finished_at) }}
          </t-descriptions-item>
          <t-descriptions-item :label="t('graphSchema.tasks.rowCount')">
            {{ taskRowCount(taskDetail) }}
          </t-descriptions-item>
          <t-descriptions-item
            v-if="taskSourceRef(taskDetail)?.entity_def_id"
            :label="t('graphSchema.tasks.entityDef')"
          >
            {{ entityName(String(taskSourceRef(taskDetail)?.entity_def_id)) }}
          </t-descriptions-item>
          <t-descriptions-item
            v-if="taskSourceRef(taskDetail)?.relation_def_id"
            :label="t('graphSchema.tasks.relationDef')"
          >
            {{ relationName(String(taskSourceRef(taskDetail)?.relation_def_id)) }}
          </t-descriptions-item>
          <t-descriptions-item
            v-if="taskSourceRef(taskDetail)?.attribute_def_id"
            :label="t('graphSchema.tasks.attributeDef')"
          >
            {{ attributeName(String(taskSourceRef(taskDetail)?.attribute_def_id)) }}
          </t-descriptions-item>
          <t-descriptions-item :label="t('graphSchema.tasks.error')">
            <span v-if="taskDetail.error_message" class="err">{{ taskDetail.error_message }}</span>
            <span v-else>{{ t('graphSchema.tasks.noError') }}</span>
          </t-descriptions-item>
        </t-descriptions>
        <div class="detail-actions">
          <t-button
            v-if="taskDetail.status === 'pending' || taskDetail.status === 'running'"
            theme="warning"
            variant="outline"
            @click="cancelTask(taskDetail)"
          >
            {{ t('graphSchema.tasks.cancel') }}
          </t-button>
          <t-button
            v-if="taskDetail.status === 'failed' || taskDetail.status === 'cancelled'"
            theme="primary"
            @click="retryTask(taskDetail)"
          >
            {{ t('graphSchema.tasks.retry') }}
          </t-button>
        </div>
      </template>
    </t-drawer>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { MessagePlugin } from 'tdesign-vue-next'
import {
  cancelGraphIngestTask,
  createGraphAttribute,
  createGraphEntity,
  createGraphIngestTask,
  createGraphRelation,
  deleteGraphAttribute,
  deleteGraphEntity,
  deleteGraphRelation,
  getGraphIngestTask,
  listGraphAttributes,
  listGraphEntities,
  listGraphIngestTasks,
  listGraphRelations,
  retryGraphIngestTask,
  updateGraphAttribute,
  updateGraphEntity,
  updateGraphRelation,
  type GraphIngestSourceRef,
  type KBGraphAttribute,
  type KBGraphEntity,
  type KBGraphIngestTask,
  type KBGraphRelation,
} from '@/api/knowledge-base/graph-schema'

const props = defineProps<{ kbId: string }>()
const { t } = useI18n()

const activeTab = ref('attributes')
const saving = ref(false)
const loadingAttrs = ref(false)
const loadingEntities = ref(false)
const loadingRelations = ref(false)
const loadingTasks = ref(false)

const attributes = ref<KBGraphAttribute[]>([])
const entities = ref<KBGraphEntity[]>([])
const relations = ref<KBGraphRelation[]>([])
const tasks = ref<KBGraphIngestTask[]>([])
const taskStatusFilter = ref<string>('')
const taskDetailVisible = ref(false)
const taskDetail = ref<KBGraphIngestTask | null>(null)

const attrDialogVisible = ref(false)
const entityDialogVisible = ref(false)
const relationDialogVisible = ref(false)
const taskDialogVisible = ref(false)

const enumValuesText = ref('')
const taskRowsText = ref('')

const emptyAttr = (): KBGraphAttribute => ({
  category: '',
  name: '',
  code: '',
  value_type: 'free',
  data_type: 'string',
  is_label: false,
  is_static: true,
  description: '',
  enum_values: [],
})
const emptyEntity = (): KBGraphEntity => ({
  category: 'thing',
  name: '',
  code: '',
  identity_key: '',
  description: '',
  attribute_ids: [],
})
const emptyRelation = (): KBGraphRelation => ({
  category: '',
  name: '',
  code: '',
  source_entity_id: '',
  target_entity_id: '',
  description: '',
  attribute_ids: [],
})

const attrForm = reactive<KBGraphAttribute>(emptyAttr())
const entityForm = reactive<KBGraphEntity>(emptyEntity())
const relationForm = reactive<KBGraphRelation>(emptyRelation())
const taskForm = reactive({
  task_kind: 'entity' as 'entity' | 'relation' | 'attribute',
  entity_def_id: '',
  relation_def_id: '',
  attribute_def_id: '',
  name_col: 'name',
  target_col: 'target',
  attr_col: 'attribute',
})

const valueTypeOptions = computed(() => [
  { label: t('graphSchema.attr.valueTypes.free'), value: 'free' },
  { label: t('graphSchema.attr.valueTypes.single'), value: 'single' },
  { label: t('graphSchema.attr.valueTypes.enum'), value: 'enum' },
])
const entityCategoryOptions = computed(() => [
  { label: t('graphSchema.entity.categories.person'), value: 'person' },
  { label: t('graphSchema.entity.categories.event'), value: 'event' },
  { label: t('graphSchema.entity.categories.place'), value: 'place' },
  { label: t('graphSchema.entity.categories.thing'), value: 'thing' },
  { label: t('graphSchema.entity.categories.organization'), value: 'organization' },
])
const taskKindOptions = computed(() => [
  { label: t('graphSchema.tasks.kinds.entity'), value: 'entity' },
  { label: t('graphSchema.tasks.kinds.relation'), value: 'relation' },
  { label: t('graphSchema.tasks.kinds.attribute'), value: 'attribute' },
])
const taskStatusFilterOptions = computed(() => [
  { label: t('graphSchema.tasks.filterAll'), value: '' },
  { label: t('graphSchema.tasks.statuses.pending'), value: 'pending' },
  { label: t('graphSchema.tasks.statuses.running'), value: 'running' },
  { label: t('graphSchema.tasks.statuses.succeeded'), value: 'succeeded' },
  { label: t('graphSchema.tasks.statuses.failed'), value: 'failed' },
  { label: t('graphSchema.tasks.statuses.cancelled'), value: 'cancelled' },
])

const attributeOptions = computed(() =>
  attributes.value.map((a) => ({ label: `${a.name} (${a.code})`, value: a.id! })),
)
const entityOptions = computed(() =>
  entities.value.map((e) => ({ label: `${e.name} (${e.code})`, value: e.id! })),
)
const relationOptions = computed(() =>
  relations.value.map((r) => ({ label: `${r.name} (${r.code})`, value: r.id! })),
)
const filteredTasks = computed(() => {
  if (!taskStatusFilter.value) return tasks.value
  return tasks.value.filter((task) => task.status === taskStatusFilter.value)
})

const attrColumns = [
  { colKey: 'name', title: () => t('graphSchema.fields.name') },
  { colKey: 'code', title: () => t('graphSchema.fields.code') },
  { colKey: 'value_type', title: () => t('graphSchema.attr.valueType') },
  { colKey: 'category', title: () => t('graphSchema.fields.category') },
  { colKey: 'op', title: () => t('common.operation'), width: 120 },
]
const entityColumns = [
  { colKey: 'name', title: () => t('graphSchema.fields.name') },
  { colKey: 'code', title: () => t('graphSchema.fields.code') },
  { colKey: 'category', title: () => t('graphSchema.fields.category') },
  { colKey: 'attrs', title: () => t('graphSchema.fields.attributes') },
  { colKey: 'op', title: () => t('common.operation'), width: 120 },
]
const relationColumns = [
  { colKey: 'name', title: () => t('graphSchema.fields.name') },
  { colKey: 'code', title: () => t('graphSchema.fields.code') },
  { colKey: 'source', title: () => t('graphSchema.relation.source') },
  { colKey: 'target', title: () => t('graphSchema.relation.target') },
  { colKey: 'op', title: () => t('common.operation'), width: 120 },
]
const taskColumns = [
  { colKey: 'task_kind', title: () => t('graphSchema.tasks.kind'), width: 110 },
  { colKey: 'status', title: () => t('graphSchema.tasks.status'), width: 100 },
  { colKey: 'progress', title: () => t('graphSchema.tasks.progress'), width: 140 },
  { colKey: 'created_at', title: () => t('graphSchema.tasks.createdAt'), width: 160 },
  { colKey: 'error', title: () => t('graphSchema.tasks.error'), ellipsis: true },
  { colKey: 'op', title: () => t('common.operation'), width: 180 },
]

function normalizeIDs(raw: unknown): string[] {
  if (Array.isArray(raw)) return raw.map(String)
  if (typeof raw === 'string') {
    try {
      const parsed = JSON.parse(raw)
      return Array.isArray(parsed) ? parsed.map(String) : []
    } catch {
      return []
    }
  }
  return []
}

function categoryLabel(cat: string) {
  const key = `graphSchema.entity.categories.${cat}`
  const label = t(key)
  return label === key ? cat : label
}

function entityName(id: string) {
  return entities.value.find((e) => e.id === id)?.name || id
}

function relationName(id: string) {
  return relations.value.find((r) => r.id === id)?.name || id
}

function attributeName(id: string) {
  return attributes.value.find((a) => a.id === id)?.name || id
}

function formatAttrNames(ids: unknown) {
  const list = normalizeIDs(ids)
  if (!list.length) return '—'
  return list
    .map((id) => attributes.value.find((a) => a.id === id)?.name || id)
    .join(', ')
}

function kindLabel(kind?: string) {
  const key = `graphSchema.tasks.kinds.${kind || 'entity'}`
  const label = t(key)
  return label === key ? kind : label
}

function statusLabel(s?: string) {
  const key = `graphSchema.tasks.statuses.${s || 'pending'}`
  const label = t(key)
  return label === key ? s : label
}

function statusTheme(s?: string): 'default' | 'primary' | 'success' | 'warning' | 'danger' {
  switch (s) {
    case 'succeeded':
      return 'success'
    case 'failed':
      return 'danger'
    case 'running':
      return 'primary'
    case 'cancelled':
      return 'warning'
    default:
      return 'default'
  }
}

function progressStatus(s?: string): 'success' | 'error' | 'warning' | 'active' | undefined {
  switch (s) {
    case 'succeeded':
      return 'success'
    case 'failed':
      return 'error'
    case 'cancelled':
      return 'warning'
    case 'running':
      return 'active'
    default:
      return undefined
  }
}

function formatTime(value?: string) {
  if (!value) return '—'
  const d = new Date(value)
  if (Number.isNaN(d.getTime())) return value
  return d.toLocaleString()
}

function taskSourceRef(task?: KBGraphIngestTask | null): GraphIngestSourceRef | null {
  if (!task?.source_ref) return null
  if (typeof task.source_ref === 'string') {
    try {
      return JSON.parse(task.source_ref) as GraphIngestSourceRef
    } catch {
      return null
    }
  }
  return task.source_ref as GraphIngestSourceRef
}

function taskRowCount(task?: KBGraphIngestTask | null) {
  const rows = taskSourceRef(task)?.rows
  return Array.isArray(rows) ? rows.length : 0
}

async function loadAttributes() {
  if (!props.kbId) return
  loadingAttrs.value = true
  try {
    const data = await listGraphAttributes(props.kbId)
    attributes.value = (data || []).map((a) => ({
      ...a,
      enum_values: normalizeIDs(a.enum_values),
    }))
  } catch (e: any) {
    MessagePlugin.error(e?.message || t('common.loadFailed'))
  } finally {
    loadingAttrs.value = false
  }
}

async function loadEntities() {
  if (!props.kbId) return
  loadingEntities.value = true
  try {
    const data = await listGraphEntities(props.kbId)
    entities.value = (data || []).map((e) => ({
      ...e,
      attribute_ids: normalizeIDs(e.attribute_ids),
    }))
  } catch (e: any) {
    MessagePlugin.error(e?.message || t('common.loadFailed'))
  } finally {
    loadingEntities.value = false
  }
}

async function loadRelations() {
  if (!props.kbId) return
  loadingRelations.value = true
  try {
    const data = await listGraphRelations(props.kbId)
    relations.value = (data || []).map((r) => ({
      ...r,
      attribute_ids: normalizeIDs(r.attribute_ids),
    }))
  } catch (e: any) {
    MessagePlugin.error(e?.message || t('common.loadFailed'))
  } finally {
    loadingRelations.value = false
  }
}

async function loadTasks() {
  if (!props.kbId) return
  loadingTasks.value = true
  try {
    tasks.value = (await listGraphIngestTasks(props.kbId)) || []
  } catch (e: any) {
    MessagePlugin.error(e?.message || t('common.loadFailed'))
  } finally {
    loadingTasks.value = false
  }
}

function openAttrDialog(row?: KBGraphAttribute) {
  Object.assign(attrForm, emptyAttr(), row || {})
  enumValuesText.value = (attrForm.enum_values || []).join('\n')
  attrDialogVisible.value = true
}

function openEntityDialog(row?: KBGraphEntity) {
  Object.assign(entityForm, emptyEntity(), row || {}, {
    attribute_ids: normalizeIDs(row?.attribute_ids),
  })
  entityDialogVisible.value = true
}

function openRelationDialog(row?: KBGraphRelation) {
  Object.assign(relationForm, emptyRelation(), row || {}, {
    attribute_ids: normalizeIDs(row?.attribute_ids),
  })
  relationDialogVisible.value = true
}

async function saveAttribute() {
  saving.value = true
  try {
    const enum_values =
      attrForm.value_type === 'enum'
        ? enumValuesText.value
            .split(/[\n,]/)
            .map((s) => s.trim())
            .filter(Boolean)
        : []
    const body = { ...attrForm, enum_values }
    if (attrForm.id) {
      await updateGraphAttribute(props.kbId, attrForm.id, body)
    } else {
      await createGraphAttribute(props.kbId, body)
    }
    attrDialogVisible.value = false
    MessagePlugin.success(t('common.saveSuccess'))
    await loadAttributes()
  } catch (e: any) {
    MessagePlugin.error(e?.message || t('common.saveFailed'))
  } finally {
    saving.value = false
  }
}

async function saveEntity() {
  saving.value = true
  try {
    const body = { ...entityForm, attribute_ids: entityForm.attribute_ids || [] }
    if (entityForm.id) {
      await updateGraphEntity(props.kbId, entityForm.id, body)
    } else {
      await createGraphEntity(props.kbId, body)
    }
    entityDialogVisible.value = false
    MessagePlugin.success(t('common.saveSuccess'))
    await loadEntities()
  } catch (e: any) {
    MessagePlugin.error(e?.message || t('common.saveFailed'))
  } finally {
    saving.value = false
  }
}

async function saveRelation() {
  saving.value = true
  try {
    const body = { ...relationForm, attribute_ids: relationForm.attribute_ids || [] }
    if (relationForm.id) {
      await updateGraphRelation(props.kbId, relationForm.id, body)
    } else {
      await createGraphRelation(props.kbId, body)
    }
    relationDialogVisible.value = false
    MessagePlugin.success(t('common.saveSuccess'))
    await loadRelations()
  } catch (e: any) {
    MessagePlugin.error(e?.message || t('common.saveFailed'))
  } finally {
    saving.value = false
  }
}

async function removeAttribute(row: KBGraphAttribute) {
  try {
    await deleteGraphAttribute(props.kbId, row.id!)
    MessagePlugin.success(t('common.deleteSuccess'))
    await loadAttributes()
  } catch (e: any) {
    MessagePlugin.error(e?.message || t('common.deleteFailed'))
  }
}

async function removeEntity(row: KBGraphEntity) {
  try {
    await deleteGraphEntity(props.kbId, row.id!)
    MessagePlugin.success(t('common.deleteSuccess'))
    await loadEntities()
  } catch (e: any) {
    MessagePlugin.error(e?.message || t('common.deleteFailed'))
  }
}

async function removeRelation(row: KBGraphRelation) {
  try {
    await deleteGraphRelation(props.kbId, row.id!)
    MessagePlugin.success(t('common.deleteSuccess'))
    await loadRelations()
  } catch (e: any) {
    MessagePlugin.error(e?.message || t('common.deleteFailed'))
  }
}

function parseRows(text: string): Record<string, string>[] {
  const lines = text
    .split('\n')
    .map((l) => l.trim())
    .filter(Boolean)
  if (lines.length < 2) return []
  const headers = lines[0].split(/[,\t]/).map((h) => h.trim())
  return lines.slice(1).map((line) => {
    const cols = line.split(/[,\t]/)
    const row: Record<string, string> = {}
    headers.forEach((h, i) => {
      row[h] = (cols[i] || '').trim()
    })
    return row
  })
}

function openTaskDialog() {
  taskForm.task_kind = 'entity'
  taskForm.entity_def_id = ''
  taskForm.relation_def_id = ''
  taskForm.attribute_def_id = ''
  taskForm.name_col = 'name'
  taskForm.target_col = 'target'
  taskForm.attr_col = 'attribute'
  taskRowsText.value = ''
  if (!entities.value.length) loadEntities()
  if (!relations.value.length) loadRelations()
  if (!attributes.value.length) loadAttributes()
  taskDialogVisible.value = true
}

async function openTaskDetail(row: KBGraphIngestTask) {
  taskDetail.value = row
  taskDetailVisible.value = true
  if (!row.id) return
  try {
    const fresh = await getGraphIngestTask(props.kbId, row.id)
    if (fresh) taskDetail.value = fresh
  } catch {
    // keep list snapshot if detail fetch fails
  }
}

async function submitTask() {
  const rows = parseRows(taskRowsText.value)
  if (!rows.length) {
    MessagePlugin.warning(t('graphSchema.tasks.rowsRequired'))
    return
  }
  if (taskForm.task_kind === 'entity' && !taskForm.entity_def_id) {
    MessagePlugin.warning(t('graphSchema.tasks.entityDefRequired'))
    return
  }
  if (taskForm.task_kind === 'relation' && !taskForm.relation_def_id) {
    MessagePlugin.warning(t('graphSchema.tasks.relationDefRequired'))
    return
  }
  saving.value = true
  try {
    const column_mapping: Record<string, string> = {}
    if (taskForm.task_kind === 'relation') {
      column_mapping.source = taskForm.name_col || 'source'
      column_mapping.target = taskForm.target_col || 'target'
    } else if (taskForm.task_kind === 'attribute') {
      column_mapping.name = taskForm.name_col || 'name'
      column_mapping.attribute = taskForm.attr_col || 'attribute'
    } else {
      column_mapping.name = taskForm.name_col || 'name'
    }
    await createGraphIngestTask(props.kbId, {
      task_kind: taskForm.task_kind,
      source_type: 'table_mapping',
      source_ref: {
        entity_def_id: taskForm.entity_def_id || undefined,
        relation_def_id: taskForm.relation_def_id || undefined,
        attribute_def_id: taskForm.attribute_def_id || undefined,
        column_mapping,
        rows,
      },
    })
    taskDialogVisible.value = false
    taskRowsText.value = ''
    MessagePlugin.success(t('graphSchema.tasks.created'))
    activeTab.value = 'tasks'
    await loadTasks()
  } catch (e: any) {
    MessagePlugin.error(e?.message || t('common.saveFailed'))
  } finally {
    saving.value = false
  }
}

async function cancelTask(row: KBGraphIngestTask) {
  try {
    const updated = await cancelGraphIngestTask(props.kbId, row.id!)
    MessagePlugin.success(t('graphSchema.tasks.cancelled'))
    await loadTasks()
    if (taskDetail.value?.id === row.id && updated) {
      taskDetail.value = updated
    }
  } catch (e: any) {
    MessagePlugin.error(e?.message || t('common.operationFailed'))
  }
}

async function retryTask(row: KBGraphIngestTask) {
  try {
    const updated = await retryGraphIngestTask(props.kbId, row.id!)
    MessagePlugin.success(t('graphSchema.tasks.retried'))
    await loadTasks()
    if (taskDetail.value?.id === row.id && updated) {
      taskDetail.value = updated
    }
  } catch (e: any) {
    MessagePlugin.error(e?.message || t('common.operationFailed'))
  }
}

let pollTimer: ReturnType<typeof setInterval> | null = null

watch(activeTab, (tab) => {
  if (tab === 'attributes') loadAttributes()
  if (tab === 'entities') {
    loadAttributes()
    loadEntities()
  }
  if (tab === 'relations') {
    loadEntities()
    loadRelations()
  }
  if (tab === 'tasks') {
    loadAttributes()
    loadEntities()
    loadRelations()
    loadTasks()
  }
})

watch(
  () => taskForm.task_kind,
  (kind) => {
    if (kind === 'relation' && taskForm.name_col === 'name') {
      taskForm.name_col = 'source'
    }
    if (kind !== 'relation' && taskForm.name_col === 'source') {
      taskForm.name_col = 'name'
    }
  },
)

watch(
  () => props.kbId,
  async (id) => {
    if (!id) return
    await Promise.all([loadAttributes(), loadEntities(), loadRelations(), loadTasks()])
  },
)

onMounted(async () => {
  await Promise.all([loadAttributes(), loadEntities(), loadRelations(), loadTasks()])
  pollTimer = setInterval(async () => {
    if (activeTab.value !== 'tasks') return
    await loadTasks()
    if (taskDetailVisible.value && taskDetail.value?.id) {
      try {
        const fresh = await getGraphIngestTask(props.kbId, taskDetail.value.id)
        if (fresh) taskDetail.value = fresh
      } catch {
        // ignore polling detail errors
      }
    }
  }, 5000)
})

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
})
</script>

<style scoped lang="less">
.graph-schema-settings {
  .section-header {
    margin-bottom: 16px;
    h2 {
      margin: 0 0 6px;
      font-size: 18px;
      color: var(--text-primary, #0f1b3d);
    }
    .section-description {
      margin: 0;
      color: var(--text-secondary, #5a6a85);
      font-size: 13px;
    }
  }
  .toolbar {
    display: flex;
    align-items: center;
    gap: 8px;
    margin: 12px 0;
    flex-wrap: wrap;
  }
  .progress-cell {
    min-width: 110px;
  }
  .op-links {
    display: flex;
    gap: 10px;
    flex-wrap: wrap;
  }
  .err {
    color: var(--error-color, #d54941);
    font-size: 12px;
  }
  .err-ellipsis {
    display: inline-block;
    max-width: 220px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    vertical-align: bottom;
  }
  .detail-actions {
    display: flex;
    gap: 8px;
    margin-top: 16px;
  }
}
</style>
