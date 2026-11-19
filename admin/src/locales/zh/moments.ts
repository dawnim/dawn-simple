export default {
  fields: {
    moments: {
      title: '标题',
      board: '板块',
      content: '内容',
      images: '图库图片',
      location: '标记地点',
      region: 'IP属地',
      created_at: '创建时间',
      updated_at: '更新时间',
      user_id: '用户ID',
      format: '内容形式',
      is_anonymous: '匿名发布'
    },
    moment_comments: {
      moment_id: '动态ID',
      user_id: '用户ID',
      reply_id: '回复ID',
      content: '评论内容',
      region: 'IP属地',
      created_at: '创建时间',
      updated_at: '更新时间'
    }
  }
}
