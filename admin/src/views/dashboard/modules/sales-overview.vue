<template>
  <div class="art-card h-105 p-5 mb-5 max-sm:mb-4">
    <div class="art-card-header">
      <div class="title">
        <h4>{{ t('近期登录统计') }}</h4>
      </div>
    </div>
    <ArtLineChart
      height="calc(100% - 40px)"
      :data="yData"
      :xAxisData="xData"
      :showAreaColor="true"
      :showAxisLine="false"
    />
  </div>
</template>

<script setup lang="ts">
  import { useI18n } from 'vue-i18n'
  const { t } = useI18n()
  import { ref, onMounted } from 'vue'
  import { ArtLineChart } from '@dawnim/admin'
  import { fetchLoginChart } from '../../../api/dashboard'

  /**
   * 登录数据
   */
  const yData = ref<number[]>([])

  /**
   * 时间数据
   */
  const xData = ref<string[]>([])

  onMounted(async () => {
    fetchLoginChart().then((data) => {
      yData.value = data.login_count
      xData.value = data.login_date
    })
  })
</script>
