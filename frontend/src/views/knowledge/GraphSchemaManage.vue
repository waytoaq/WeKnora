<template>
  <main class="graph-schema-manage">
    <header class="page-header" style="--wails-draggable: drag">
      <div class="title-row" style="--wails-draggable: drag">
        <h2 style="--wails-draggable: drag">
          <t-icon name="root-list" size="22px" class="title-icon" />
          {{ t('graphSchemaManage.title') }}
        </h2>
        <div class="tools" style="--wails-draggable: no-drag">
          <t-select
            v-model="selectedKbId"
            class="kb-select"
            filterable
            clearable
            :loading="loadingKbs"
            :options="kbOptions"
            :placeholder="t('graphSchemaManage.selectKb')"
            @change="onKbChange"
          />
          <t-button variant="outline" size="small" :loading="loadingKbs" @click="loadKnowledgeBases">
            {{ t('common.refresh') }}
          </t-button>
        </div>
      </div>
      <p class="subtitle" style="--wails-draggable: drag">{{ t('graphSchemaManage.subtitle') }}</p>
    </header>

    <section class="page-body">
      <div v-if="loadingKbs && !kbOptions.length" class="state-box">
        <t-loading :text="t('common.loading')" />
      </div>
      <div v-else-if="!kbOptions.length" class="state-box">
        <t-empty :description="t('graphSchemaManage.noKb')" />
      </div>
      <div v-else-if="!selectedKbId" class="state-box">
        <t-empty :description="t('graphSchemaManage.pickKbHint')" />
      </div>
      <div v-else class="schema-panel">
        <div class="kb-meta">
          <span class="kb-name">{{ selectedKbName }}</span>
          <t-link theme="primary" hover="color" @click="openKb">
            {{ t('graphSchemaManage.openKb') }}
          </t-link>
        </div>
        <GraphSchemaSettings :key="selectedKbId" :kb-id="selectedKbId" />
      </div>
    </section>
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { MessagePlugin } from 'tdesign-vue-next'
import { listKnowledgeBases } from '@/api/knowledge-base'
import GraphSchemaSettings from './settings/GraphSchemaSettings.vue'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()

type KbOption = { label: string; value: string; type?: string; isDefault?: boolean }

const loadingKbs = ref(false)
const kbOptions = ref<KbOption[]>([])
const selectedKbId = ref('')

const selectedKbName = computed(() => {
  return kbOptions.value.find((k) => k.value === selectedKbId.value)?.label || selectedKbId.value
})

function pickInitialKbId(options: KbOption[], queryKb: string, current: string): string {
  if (queryKb && options.some((k) => k.value === queryKb)) return queryKb
  if (current && options.some((k) => k.value === current)) return current
  const defaultKb = options.find((k) => k.isDefault)
  if (defaultKb) return defaultKb.value
  if (options.length === 1) return options[0].value
  return ''
}

async function loadKnowledgeBases() {
  loadingKbs.value = true
  try {
    const res: any = await listKnowledgeBases()
    const list = Array.isArray(res?.data) ? res.data : Array.isArray(res) ? res : []
    kbOptions.value = list
      .filter((kb: any) => kb?.type !== 'faq')
      .map((kb: any) => ({
        label: kb.is_default ? `${kb.name || kb.id}（${t('graphSchemaManage.defaultBadge')}）` : (kb.name || kb.id),
        value: kb.id,
        type: kb.type,
        isDefault: !!kb.is_default,
      }))
    const queryKb = typeof route.query.kbId === 'string' ? route.query.kbId : ''
    const next = pickInitialKbId(kbOptions.value, queryKb, selectedKbId.value)
    selectedKbId.value = next
    if (next && queryKb !== next && !queryKb) {
      router.replace({ path: '/platform/graph-schema', query: { kbId: next } })
    }
  } catch (e: any) {
    MessagePlugin.error(e?.message || t('common.loadFailed'))
  } finally {
    loadingKbs.value = false
  }
}

function onKbChange(value: string) {
  const next = value || ''
  selectedKbId.value = next
  router.replace({
    path: '/platform/graph-schema',
    query: next ? { kbId: next } : {},
  })
}

function openKb() {
  if (!selectedKbId.value) return
  router.push(`/platform/knowledge-bases/${selectedKbId.value}`)
}

watch(
  () => route.query.kbId,
  (kbId) => {
    if (typeof kbId === 'string' && kbId && kbId !== selectedKbId.value) {
      if (kbOptions.value.some((k) => k.value === kbId)) {
        selectedKbId.value = kbId
      }
    }
  },
)

onMounted(loadKnowledgeBases)
</script>

<style scoped lang="less">
.graph-schema-manage {
  height: 100%;
  display: flex;
  flex-direction: column;
  min-height: 0;
  padding: 20px 24px 24px;
  background:
    radial-gradient(1200px 480px at 12% -10%, rgba(42, 100, 246, 0.08), transparent 60%),
    linear-gradient(180deg, #f5f8ff 0%, #eef3fb 100%);
}

.page-header {
  margin-bottom: 16px;
}

.title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;

  h2 {
    margin: 0;
    display: inline-flex;
    align-items: center;
    gap: 8px;
    font-size: 22px;
    font-weight: 650;
    color: #0f1b3d;
  }
}

.title-icon {
  color: #2a64f6;
}

.subtitle {
  margin: 8px 0 0;
  color: #5a6a85;
  font-size: 13px;
  line-height: 1.5;
}

.tools {
  display: flex;
  align-items: center;
  gap: 8px;
}

.kb-select {
  min-width: 260px;
}

.page-body {
  flex: 1;
  min-height: 0;
  overflow: auto;
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.72);
  border: 1px solid rgba(42, 100, 246, 0.08);
  box-shadow: 0 10px 30px rgba(15, 27, 61, 0.06);
  backdrop-filter: blur(10px);
  padding: 16px 18px 20px;
}

.state-box {
  min-height: 280px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.kb-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 8px;
  padding-bottom: 10px;
  border-bottom: 1px solid rgba(15, 27, 61, 0.06);
}

.kb-name {
  font-size: 14px;
  font-weight: 600;
  color: #0f1b3d;
}
</style>
