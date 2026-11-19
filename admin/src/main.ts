import { createAdmin } from '@dawnim/admin'
import '@dawnim/admin/style.css'
import HeaderTools from './components/HeaderTools.vue'

createAdmin({
  apiBase: import.meta.env.VITE_ADMIN_PREFIX || '/admin',
  header: { before: HeaderTools },
  messages: {
    en: {
      '网友动态': 'Community moments',
      '每日发布动态上限': 'Daily post limit',
      '每日AI创作总次数': 'Daily AI draft limit',
      '每个IP每天最多使用2次': 'Each IP can generate up to 2 drafts per day',
      '每位用户每天可发布的动态数量': 'Number of posts each user may publish per day',
      '每日AI自动动态上限': 'Daily automated AI post limit',
      '设为0关闭，依据近期文档阅读生成': 'Set to 0 to disable; based on recent documentation visits',
      'Brave搜索密钥': 'Brave Search key',
      '选填，为AI自动动态提供近期趣闻素材': 'Optional; provides recent stories for automated AI posts',
      '隐藏的编辑器工具': 'Hidden editor tools',
      '逗号分隔的工具名，如 media,alignment；留空显示全部': 'Comma-separated tool keys, e.g. media,alignment; leave empty to show all',
      '自定义表情短码': 'Custom emoji shortcodes',
      '每行一个，格式 名称=表情，正文中用 :名称: 引用': 'One per line as name=emoji; use :name: in posts',
      '添加图片': 'Add images',
      '图库图片': 'Gallery images',
      '移除图片': 'Remove image',
      '搜索图片': 'Search images',
      '上传图片': 'Upload images',
      '上传失败': 'Upload failed',
      '暂无图片': 'No images',
      '最多选择9张图片': 'Choose up to 9 images',
      '评论': 'Comments',
      '评论列表': 'Comments',
      '搜索评论内容': 'Search comments',
      '评论内容不能为空': 'Comment content is required',
      '保存': 'Save',
      '新增': 'Create',
      '删除': 'Delete',
      '编辑': 'Edit',
      '操作': 'Actions',
      '新增成功': 'Created',
      '修改成功': 'Updated'
    }
  },
  pages: {
    '/support/moments': () => import('./views/business/moments/index.vue'),
    '/dashboard/statistics': () => import('./views/dashboard/index.vue'),
    '/dashboard/work': () => import('./views/dashboard/work.vue')
  }
}).mount('#app')
