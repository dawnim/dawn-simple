<template>
  <div class="art-full-height">
    <CrudPage
      :api="api"
      :columns="columns"
      :search="search"
      :defaults="defaults"
      :action-width="220"
      form-width="1100px"
      permission="app.moments"
    >
      <template #actions="{ row }">
        <ElButton v-if="canComment" size="small" @click="openComments(row)">{{
          t('评论')
        }}</ElButton>
      </template>
      <template #content="{ form }">
        <MomentEditor
          v-model="form.content"
          :format="form.format"
          :record-id="form.id"
          :hidden="hiddenTools"
        />
      </template>
      <template #images="{ form }">
        <MomentImages v-model="form.images" />
      </template>
    </CrudPage>
    <CommentDialog v-if="current" v-model="commentVisible" :moment="current" />
  </div>
</template>
<script setup lang="ts">
  import { computed, ref } from 'vue'
  import { useI18n } from 'vue-i18n'
  import { ElButton } from 'element-plus'
  import { CrudPage, useAuth } from '@dawnim/admin'
  import api, {
    columns as getColumns,
    search,
    defaults
  } from '../../../api/business/moments'
  import zh from '../../../locales/zh/moments'
  import en from '../../../locales/en/moments'
  import CommentDialog from './modules/comment-dialog.vue'
  import MomentEditor from './modules/moment-editor.vue'
  import MomentImages from './modules/moment-images.vue'

  const { t } = useI18n({ useScope: 'global' })
  const { t: fieldText } = useI18n({
    useScope: 'local',
    messages: { zh, en },
    fallbackFormat: true,
    missingWarn: false,
    fallbackWarn: false
  })
  const columns = computed(() => getColumns(fieldText))
  const canComment = useAuth().hasAuth('app.moment_comments.list')
  const commentVisible = ref(false)
  const current = ref<Record<string, any> | null>(null)
  const hiddenTools = ref<string[]>([])
  fetch('/api/moment-config')
    .then((response) => response.json())
    .then((payload) => (hiddenTools.value = payload?.data?.toolbar_hidden || []))
    .catch(() => undefined)

  function openComments(row: Record<string, any>) {
    current.value = row
    commentVisible.value = true
  }
</script>
