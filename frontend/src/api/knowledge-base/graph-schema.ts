import { get, post, put, del } from '../../utils/request'

export type GraphAttrValueType = 'enum' | 'single' | 'free'
export type GraphEntityCategory = 'person' | 'event' | 'place' | 'thing' | 'organization'
export type GraphIngestKind = 'entity' | 'relation' | 'attribute'
export type GraphIngestStatus = 'pending' | 'running' | 'succeeded' | 'failed' | 'cancelled'

export interface KBGraphAttribute {
  id?: string
  knowledge_base_id?: string
  category: string
  name: string
  code: string
  value_type: GraphAttrValueType
  data_type: string
  is_label: boolean
  is_static: boolean
  description: string
  enum_values: string[]
  created_at?: string
  updated_at?: string
}

export interface KBGraphEntity {
  id?: string
  knowledge_base_id?: string
  category: GraphEntityCategory | string
  name: string
  code: string
  identity_key: string
  description: string
  attribute_ids: string[]
  created_at?: string
  updated_at?: string
}

export interface KBGraphRelation {
  id?: string
  knowledge_base_id?: string
  category: string
  name: string
  code: string
  source_entity_id: string
  target_entity_id: string
  description: string
  attribute_ids: string[]
  created_at?: string
  updated_at?: string
}

export interface GraphIngestSourceRef {
  knowledge_id?: string
  entity_def_id?: string
  relation_def_id?: string
  attribute_def_id?: string
  column_mapping?: Record<string, string>
  rows?: Record<string, string>[]
}

export interface KBGraphIngestTask {
  id?: string
  knowledge_base_id?: string
  task_kind: GraphIngestKind
  status?: GraphIngestStatus
  source_type?: string
  source_ref?: GraphIngestSourceRef | Record<string, unknown>
  progress?: number
  error_message?: string
  started_at?: string
  finished_at?: string
  created_at?: string
  updated_at?: string
}

function base(kbId: string) {
  return `/api/v1/knowledge-bases/${kbId}/graph-schema`
}

function unwrap<T>(res: unknown): T {
  const r = res as { data?: T }
  return (r?.data ?? res) as T
}

export async function listGraphAttributes(kbId: string) {
  return unwrap<KBGraphAttribute[]>(await get(`${base(kbId)}/attributes`))
}

export async function createGraphAttribute(kbId: string, body: Partial<KBGraphAttribute>) {
  return unwrap<KBGraphAttribute>(await post(`${base(kbId)}/attributes`, body))
}

export async function updateGraphAttribute(kbId: string, attrId: string, body: Partial<KBGraphAttribute>) {
  return unwrap<KBGraphAttribute>(await put(`${base(kbId)}/attributes/${attrId}`, body))
}

export async function deleteGraphAttribute(kbId: string, attrId: string) {
  return unwrap(await del(`${base(kbId)}/attributes/${attrId}`))
}

export async function listGraphEntities(kbId: string) {
  return unwrap<KBGraphEntity[]>(await get(`${base(kbId)}/entities`))
}

export async function createGraphEntity(kbId: string, body: Partial<KBGraphEntity>) {
  return unwrap<KBGraphEntity>(await post(`${base(kbId)}/entities`, body))
}

export async function updateGraphEntity(kbId: string, entityId: string, body: Partial<KBGraphEntity>) {
  return unwrap<KBGraphEntity>(await put(`${base(kbId)}/entities/${entityId}`, body))
}

export async function deleteGraphEntity(kbId: string, entityId: string) {
  return unwrap(await del(`${base(kbId)}/entities/${entityId}`))
}

export async function listGraphRelations(kbId: string) {
  return unwrap<KBGraphRelation[]>(await get(`${base(kbId)}/relations`))
}

export async function createGraphRelation(kbId: string, body: Partial<KBGraphRelation>) {
  return unwrap<KBGraphRelation>(await post(`${base(kbId)}/relations`, body))
}

export async function updateGraphRelation(kbId: string, relationId: string, body: Partial<KBGraphRelation>) {
  return unwrap<KBGraphRelation>(await put(`${base(kbId)}/relations/${relationId}`, body))
}

export async function deleteGraphRelation(kbId: string, relationId: string) {
  return unwrap(await del(`${base(kbId)}/relations/${relationId}`))
}

export async function listGraphIngestTasks(kbId: string) {
  return unwrap<KBGraphIngestTask[]>(await get(`${base(kbId)}/tasks`))
}

export async function getGraphIngestTask(kbId: string, taskId: string) {
  return unwrap<KBGraphIngestTask>(await get(`${base(kbId)}/tasks/${taskId}`))
}

export async function createGraphIngestTask(kbId: string, body: Partial<KBGraphIngestTask>) {
  // Backend expects source_ref as JSON object; stringify enum_values style not needed.
  const payload = {
    ...body,
    source_ref: body.source_ref ?? {},
  }
  return unwrap<KBGraphIngestTask>(await post(`${base(kbId)}/tasks`, payload))
}

export async function cancelGraphIngestTask(kbId: string, taskId: string) {
  return unwrap<KBGraphIngestTask>(await post(`${base(kbId)}/tasks/${taskId}/cancel`, {}))
}

export async function retryGraphIngestTask(kbId: string, taskId: string) {
  return unwrap<KBGraphIngestTask>(await post(`${base(kbId)}/tasks/${taskId}/retry`, {}))
}
