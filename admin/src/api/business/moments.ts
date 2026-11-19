import { createResourceApi, type Column, type ListQuery } from '@dawnim/admin'

export interface Moments {
  id: number
  created_at: number
  updated_at: number
  user_id: number
  format: string
  title: string
  board: string
  content: string
  images: string
  location: string
  region: string
  is_anonymous: boolean
}
export interface CreateMomentsInput {
  user_id: number
  format: string
  title: string
  board: string
  content: string
  images: string
  location?: string
  region?: string
  is_anonymous?: boolean
}
export type UpdateMomentsInput = Partial<CreateMomentsInput>
export interface MomentsQuery extends ListQuery {
  title?: string
  board?: string
  content?: string
  region?: string
  format?: string
}
export const columns = (t: (key: string) => string): Column[] =>
  ([
    { prop: 'title', label: 'fields.moments.title', required: true, maxlength: 100 },
    { prop: 'format', label: 'fields.moments.format', required: true, options: [{ value: 'moment', label: t('动态') }, { value: 'post', label: t('帖子') }] },
    {
      prop: 'board', label: 'fields.moments.board', required: true,
      options: [
        { value: 'general', label: t('自由交流') },
        { value: 'help', label: t('问题求助') },
        { value: 'showcase', label: t('作品分享') },
        { value: 'ai', label: t('AI趣闻') }
      ]
    },
    { prop: 'user_id', label: 'fields.moments.user_id', type: 'number', required: true },
    { prop: 'content', label: 'fields.moments.content', type: 'textarea', required: true, formSlot: 'content' },
    { prop: 'images', label: 'fields.moments.images', type: 'textarea', hidden: true, formSlot: 'images' },
    { prop: 'region', label: 'fields.moments.region' },
    { prop: 'is_anonymous', label: 'fields.moments.is_anonymous', type: 'boolean' },
    { prop: 'location', label: 'fields.moments.location' },
    { prop: 'created_at', label: 'fields.moments.created_at', type: 'date', readonly: true },
    { prop: 'updated_at', label: 'fields.moments.updated_at', type: 'date', readonly: true }
  ] as Column[]).map((column) => ({ ...column, label: t(column.label) }))
export const search: string[] = ['title', 'format', 'board', 'content', 'region']
export const defaults = { title: '', board: 'general', format: 'moment', content: '', images: '[]', location: '', region: '', user_id: 0, is_anonymous: false }

export default createResourceApi<Moments, CreateMomentsInput, UpdateMomentsInput, MomentsQuery>('/moments')
