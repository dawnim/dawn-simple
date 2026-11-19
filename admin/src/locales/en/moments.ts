export default {
  自由交流: 'Discussion',
  问题求助: 'Help',
  作品分享: 'Showcase',
  AI趣闻: 'AI stories',
  动态: 'Moment',
  帖子: 'Post',
  fields: {
    moments: {
      title: 'Title',
      board: 'Board',
      content: 'Content',
      images: 'Images',
      location: 'Location',
      region: 'IP region',
      created_at: 'Created at',
      updated_at: 'Updated at',
      user_id: 'User ID',
      format: 'Format',
      is_anonymous: 'Anonymous'
    },
    moment_comments: {
      moment_id: 'Moment ID',
      user_id: 'User ID',
      reply_id: 'Reply ID',
      content: 'Content',
      region: 'IP region',
      created_at: 'Created at',
      updated_at: 'Updated at'
    }
  }
}
