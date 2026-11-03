<template>
  <time class="header-clock" :datetime="now.toISOString()" :title="formattedTime">{{
    formattedTime
  }}</time>
</template>
<script setup lang="ts">
  import { ref, computed, onMounted, onUnmounted } from 'vue'
  defineOptions({ inheritAttrs: false })
  const props = defineProps<{ timezone: string; serverTimeOffset: number }>()
  const tick = ref(Date.now())
  const now = computed(() => new Date(tick.value + props.serverTimeOffset))
  const formattedTime = computed(() => {
    const formatter = new Intl.DateTimeFormat('zh-CN', {
      timeZone: props.timezone,
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
      hourCycle: 'h23',
      timeZoneName: 'longOffset'
    })
    const parts = Object.fromEntries(
      formatter.formatToParts(now.value).map((part) => [part.type, part.value])
    )
    return `${parts.year}-${parts.month}-${parts.day} ${parts.hour}:${parts.minute}:${parts.second} ${props.timezone} (${parts.timeZoneName.replace('GMT', 'UTC')})`
  })
  let timer: ReturnType<typeof setInterval>
  onMounted(() => {
    timer = setInterval(() => (tick.value = Date.now()), 1000)
  })
  onUnmounted(() => clearInterval(timer))
</script>
<style scoped>
  .header-clock {
    color: var(--el-text-color-secondary);
    font-size: 12px;
    white-space: nowrap;
  }
  @media (max-width: 767px) {
    .header-clock {
      max-width: 190px;
      overflow: hidden;
      text-overflow: ellipsis;
    }
  }
</style>
