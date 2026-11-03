<template>
  <div class="art-card p-5 flex-b mb-5 max-sm:mb-4">
    <div>
      <h2 class="text-2xl font-medium">{{ t('关于项目') }}</h2>
      <p class="text-g-700 mt-1">
        {{ systemName }}
        {{ t('是一款兼具设计美学与高效开发的后台系统') }}
      </p>
      <p class="text-g-700 mt-1">
        {{ t('使用了Go + Vue3 + Element Plus高性能、高颜值技术栈') }}
      </p>

      <div class="flex flex-wrap gap-3.5 max-w-150 mt-9">
        <div
          class="w-60 flex-cb h-12.5 px-3.5 border border-g-300 c-p rounded-lg text-sm bg-g-100 duration-300 hover:-translate-y-1 max-sm:w-full"
          v-for="link in linkList"
          :key="link.label"
          @click="goPage(link.url)"
        >
          <span class="text-g-700">{{ link.label }}</span>
          <ArtSvgIcon icon="ri:arrow-right-s-line" class="text-lg text-g-600" />
        </div>
      </div>
    </div>
    <img class="w-75 max-md:!hidden" src="../../../assets/draw1.png" alt="draw1" />
  </div>
</template>

<script setup lang="ts">
  import { useI18n } from 'vue-i18n'
  const { t } = useI18n()
  import { computed } from 'vue'
  import { ArtSvgIcon } from '@dawnim/admin'

  const systemName = 'Dawn'

  const linkList = computed(() => [
    {
      label: t('服务端'),
      url: 'https://github.com/dawnim/dawn-server'
    },
    {
      label: t('文档'),
      url: 'https://github.com/dawnim/dawn-server/blob/main/README.md'
    },
    {
      label: t('前端'),
      url: 'https://github.com/dawnim/dawn-admin'
    },
    {
      label: t('演示项目'),
      url: 'https://github.com/dawnim/dawn-simple'
    }
  ])

  /**
   * 在新标签页中打开指定 URL
   * @param url 要打开的网页地址
   */
  const goPage = (url: string): void => {
    window.open(url, '_blank', 'noopener,noreferrer')
  }
</script>
