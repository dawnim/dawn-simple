<template>
  <ElDialog v-model="visible" :title="t('评论列表')" width="960px" align-center>
    <div class="comment-post">#{{ moment.id }} {{ moment.title }}</div>
    <ElForm v-if="editing" :model="form" label-width="90px" class="comment-form">
      <ElFormItem :label="field('fields.moment_comments.content')" required>
        <ElInput v-model="form.content" type="textarea" :rows="3" maxlength="300" show-word-limit />
      </ElFormItem>
      <ElFormItem :label="field('fields.moment_comments.user_id')" required>
        <ElInputNumber v-model="form.user_id" :min="0" />
      </ElFormItem>
      <ElFormItem :label="field('fields.moment_comments.reply_id')">
        <ElInputNumber v-model="form.reply_id" :min="0" />
      </ElFormItem>
      <ElFormItem :label="field('fields.moment_comments.region')">
        <ElInput v-model="form.region" maxlength="100" />
      </ElFormItem>
      <ElButton type="primary" :loading="saving" @click="submit">{{ t('保存') }}</ElButton>
      <ElButton @click="editing = false">{{ t('取消') }}</ElButton>
    </ElForm>
    <div class="comment-toolbar">
      <ElButton v-if="can('create')" type="primary" @click="startCreate">{{
        t('新增')
      }}</ElButton>
      <ElButton
        v-if="can('delete')"
        :disabled="selectedRows.length === 0"
        @click="deleteSelectedRows(api.removeMany, changed)"
        >{{ t('删除') }}</ElButton
      >
      <ElInput
        v-model="keyword"
        clearable
        class="comment-search"
        :placeholder="t('搜索评论内容')"
        @keyup.enter="doSearch"
        @clear="doSearch"
      />
      <ElButton @click="doSearch">{{ t('搜索') }}</ElButton>
      <ElButton @click="refreshData">{{ t('刷新') }}</ElButton>
    </div>
    <ArtTable
      row-key="id"
      :data="data"
      :loading="loading"
      :columns="tableColumns"
      :pagination="pagination"
      @selection-change="handleSelectionChange"
      @sort-change="handleSortChange"
      @pagination:size-change="handleSizeChange"
      @pagination:current-change="handleCurrentChange"
    >
      <template #operation="{ row }">
        <div class="flex gap-2">
          <DawnButton v-if="can('update')" type="secondary" @click="startEdit(row)" />
          <DawnButton
            v-if="can('delete')"
            type="error"
            @click="deleteRow(row, api.remove, changed)"
          />
        </div>
      </template>
    </ArtTable>
  </ElDialog>
</template>
<script setup lang="ts">
  import { computed, reactive, ref, watch } from 'vue'
  import { useI18n } from 'vue-i18n'
  import {
    ElDialog,
    ElForm,
    ElFormItem,
    ElInput,
    ElInputNumber,
    ElButton,
    ElMessage
  } from 'element-plus'
  import {
    ArtTable,
    DawnButton,
    dateLabel,
    useTable,
    useAdmin,
    useAuth
  } from '@dawnim/admin'
  import api, { defaults } from '../../../../api/business/moment-comments'
  import zh from '../../../../locales/zh/moments'
  import en from '../../../../locales/en/moments'

  const props = defineProps<{ modelValue: boolean; moment: Record<string, any> }>()
  const emit = defineEmits<{ 'update:modelValue': [boolean] }>()
  const { t } = useI18n({ useScope: 'global' })
  const { t: field } = useI18n({
    useScope: 'local',
    messages: { zh, en },
    fallbackFormat: true,
    missingWarn: false,
    fallbackWarn: false
  })
  const { hasAuth } = useAuth()
  const can = (action: string) => hasAuth('app.moment_comments.' + action)
  const visible = computed({
    get: () => props.modelValue,
    set: (value) => emit('update:modelValue', value)
  })

  const { columns: tableColumns, data, loading, pagination, searchParams, getData, refreshData, handleSortChange, handleSizeChange, handleCurrentChange } = useTable({
    core: {
      apiFn: (params: any) => api.list({ ...params, moment_id: props.moment.id }),
      columnsFactory: () => [
        { type: 'selection' },
        { prop: 'id', label: 'ID', width: 90, align: 'center', sortable: true },
        { prop: 'user_id', label: field('fields.moment_comments.user_id'), minWidth: 110, sortable: true },
        { prop: 'reply_id', label: field('fields.moment_comments.reply_id'), minWidth: 110, sortable: true },
        { prop: 'content', label: field('fields.moment_comments.content'), minWidth: 240, showOverflowTooltip: true },
        { prop: 'region', label: field('fields.moment_comments.region'), minWidth: 120 },
        {
          prop: 'created_at',
          label: field('fields.moment_comments.created_at'),
          minWidth: 160,
          formatter: (row: Record<string, any>) => dateLabel(row.created_at)
        },
        { prop: 'operation', label: t('操作'), width: 100, fixed: 'right', useSlot: true }
      ]
    }
  })
  const { deleteRow, deleteSelectedRows, handleSelectionChange, selectedRows } = useAdmin()

  const keyword = ref('')
  const editing = ref(false)
  const editingId = ref<number | null>(null)
  const saving = ref(false)
  const form = reactive<Record<string, any>>({ ...defaults })

  function changed() {
    refreshData()
  }
  function doSearch() {
    Object.assign(searchParams, { content: keyword.value || undefined })
    getData()
  }
  function startCreate() {
    editingId.value = null
    Object.assign(form, defaults, { user_id: props.moment.user_id || 0 })
    editing.value = true
  }
  function startEdit(row: Record<string, any>) {
    editingId.value = row.id
    Object.assign(form, {
      user_id: row.user_id,
      reply_id: row.reply_id,
      content: row.content,
      region: row.region
    })
    editing.value = true
  }
  async function submit() {
    if (!String(form.content).trim()) {
      ElMessage.warning(t('评论内容不能为空'))
      return
    }
    const input = {
      moment_id: props.moment.id,
      user_id: Number(form.user_id),
      reply_id: Number(form.reply_id) || 0,
      content: String(form.content).trim(),
      region: String(form.region || '').trim()
    }
    saving.value = true
    try {
      if (editingId.value) await api.update(editingId.value, input)
      else await api.create(input)
      ElMessage.success(t(editingId.value ? '修改成功' : '新增成功'))
      editing.value = false
      refreshData()
    } finally {
      saving.value = false
    }
  }
  watch(
    () => props.moment.id,
    () => {
      editing.value = false
      keyword.value = ''
      Object.assign(searchParams, { content: undefined })
      refreshData()
    }
  )
</script>
<style scoped>
  .comment-post {
    margin-bottom: 12px;
    color: var(--el-text-color-secondary);
  }
  .comment-form {
    padding: 12px;
    margin-bottom: 12px;
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 6px;
  }
  .comment-toolbar {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-bottom: 12px;
  }
  .comment-search {
    width: 220px;
  }
</style>
