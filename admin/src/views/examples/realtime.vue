<template>
  <ElCard>
    <template #header>{{ t('实时示例') }}</template>
    <p class="mb-4">{{ t('通过共享WebSocket读取业务缓存，可在缓存管理中查看或清理') }}</p>
    <ElButton :disabled="realtime.status.value !== 'ready'" :loading="loading" @click="read">{{ t('读取问候') }}</ElButton>
    <p class="mt-4">{{ greeting }}</p>
  </ElCard>
</template>
<script setup lang="ts">
  import { ref } from 'vue'
  import { useI18n } from 'vue-i18n'
  import { realtime } from '@dawnim/admin'
  const { t } = useI18n()
  const loading = ref(false)
  const greeting = ref('')
  async function read() {
    loading.value = true
    try { greeting.value = (await realtime.request<{ greeting: string }>('app.greeting')).greeting }
    catch (error) { greeting.value = error instanceof Error ? error.message : String(error) }
    finally { loading.value = false }
  }
</script>
