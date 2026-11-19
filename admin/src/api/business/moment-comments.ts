import { createResourceApi, type ListQuery } from '@dawnim/admin'

export interface MomentComment {
  id: number
  created_at: number
  updated_at: number
  moment_id: number
  user_id: number
  reply_id: number
  content: string
  region: string
}
export interface CreateMomentCommentInput {
  moment_id: number
  user_id: number
  reply_id?: number
  content: string
  region?: string
}
export type UpdateMomentCommentInput = Partial<CreateMomentCommentInput>
export interface MomentCommentQuery extends ListQuery {
  moment_id?: number
  user_id?: number
  content?: string
  region?: string
}
export const defaults = { user_id: 0, reply_id: 0, content: '', region: '' }

export default createResourceApi<
  MomentComment,
  CreateMomentCommentInput,
  UpdateMomentCommentInput,
  MomentCommentQuery
>('/moment-comments')
