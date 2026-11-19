import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { getAuthConfig } from '@/api/auth'

/** Built-in wordmark when no system_name override is configured. */
export const DEFAULT_SYSTEM_NAME = '知识'

export const useBrandingStore = defineStore('branding', () => {
  const systemName = ref('')
  const logo = ref('')
  const loaded = ref(false)
  let loadPromise: Promise<void> | null = null

  const displayName = computed(() => {
    const name = systemName.value.trim()
    return name || DEFAULT_SYSTEM_NAME
  })

  const logoUrl = computed(() => logo.value.trim())

  function apply(next: { system_name?: string | null; logo?: string | null }) {
    if (typeof next.system_name === 'string') {
      systemName.value = next.system_name
    }
    if (typeof next.logo === 'string') {
      logo.value = next.logo
    }
    if (typeof document !== 'undefined') {
      document.title = displayName.value
    }
  }

  async function load(force = false): Promise<void> {
    if (loaded.value && !force) return
    if (loadPromise && !force) return loadPromise
    loadPromise = (async () => {
      try {
        const response = await getAuthConfig()
        if (response.success) {
          apply({
            system_name: response.system_name ?? '',
            logo: response.logo ?? '',
          })
        }
      } finally {
        loaded.value = true
        loadPromise = null
      }
    })()
    return loadPromise
  }

  return {
    systemName,
    logo,
    loaded,
    displayName,
    logoUrl,
    apply,
    load,
  }
})
