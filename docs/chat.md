# 聊天与网站客服接入

后台顶部聊天入口同时支持管理员私聊与客服会话。普通管理员使用小窗口，有客服权限时使用包含会话列表、聊天内容和客户资料的大窗口。右上角单个胶囊按“私聊→客服→全屏客服→私聊”循环切换，旁边保留主题与关闭图标；左上角显示当前管理员头像、今日回复消息数和今日接待客户数。权限由服务端检查，客服权限不会开放其他管理员的私聊。

打开窗口默认查看全部会话，优先打开未读会话；查看历史不需要开启接待，也不会认领会话，打开会话后记录已读并消除对应角标。开启“开始接待”开关后，系统按应用权限和容量分配等待中的客户，只能回复分配给自己的会话；关闭开关后不再分配新客户，输入框禁用。客服心跳不会自行开启接待。今日统计与系统仪表盘一致，按UTC自然日计算，回复仅统计公开且已发送成功的客服文本，接待按当日认领或转入的客户去重，私聊、内部备注、AI回复和失败消息不计入。AI接待、等待人工、人工接待与已结束使用淡蓝、淡橙、淡绿与淡灰标签。

客服配置、知识与业务扩展见[AI客服](ai.md)，用户模型和编号见[普通用户](users.md)，IP信息见[GeoIP](geoip.md)。

访客与后台聊天均支持Unicode表情。访客聊天沿用后台私聊的仿微信配色、头像和气泡，支持亮色、暗色与跟随系统主题。

## 独立聊天链接

```text
https://support.example.com/chat/应用公开标识?user=alice&user_id=42&mobile=13800000000&nick=Alice
```

任意额外查询参数会保存为该应用内的客户资料，后台可查看。`entry`与`theme`分别选择入口和`light`、`dark`或`system`主题，`locale`选择`zh`或`en`语言。未知业务字段保持原字段名，不被当作登录凭证。网站来源填写`*`或留空允许任意网站。simple与第三方双向身份接入见[接入教程](support-identity.md)。

敏感资料优先使用JS属性或`postMessage`，URL参数可能进入接入网站的访问日志。认证票据不放在普通查询参数中。

## 悬浮入口

```html
<script src="https://support.example.com/support-assets/widget.js"
        data-app="应用公开标识" data-theme="system" data-locale="zh" defer></script>
```

悬浮按钮可以上下拖动并自动贴到左右边缘，位置保存在当前网站。收起时仅按钮接收点击，不保留全屏透明遮罩；聊天页面放在独立iframe中，按钮使用Shadow DOM隔离样式。手机打开时使用全屏聊天；电脑端点击聊天窗口外部会临时收起窗口，再次打开保留会话、输入内容和滚动位置。

动态资料使用初始化参数传入。

可选配置 `menu` 为悬浮入口提供自定义菜单：`label` 可为文字或接收语言的函数，`icon` 是图片地址，`onClick` 接收客服 API；`menuLayout` 可设为 `vertical`（默认）或 `horizontal`。未配置菜单时直接打开客服。

```js
DawnSupport.init({
  app: 'public-key',
  menuLayout: 'vertical',
  menu: [
    { label: locale => locale === 'zh' ? '客服' : 'Support', icon: '/support.svg', onClick: support => support.open() },
    { label: locale => locale === 'zh' ? '聊天' : 'Chat', icon: '/chat.svg', onClick: () => openSiteChat() },
  ],
})
```

```js
DawnSupport.init({
  server: 'https://support.example.com',
  app: '应用公开标识',
  theme: 'system',
  locale: 'zh',
  linkRules: [{ from: 'https://docs.example.com/product/', to: '/docs/' }],
  closeOnMobileNavigate: true,
  attributes: { user: 'alice', user_id: '42', mobile: '13800000000', nick: 'Alice', plan: 'standard' }
})
DawnSupport.open()
DawnSupport.close()
DawnSupport.setTheme('dark')
DawnSupport.setLocale('en')
DawnSupport.identify({ attributes: { plan: 'pro' } }) // 更新未认证资料
DawnSupport.identify({ identity: signedIdentity, attributes: { plan: 'standard' } })
// 接入网站账号退出或切换后调用，解除当前客服身份并清除该网站保存的凭证
DawnSupport.reset()
```

`data-app`存在时脚本自动初始化，使用JS初始化时不要同时自动初始化。`destroy()`移除入口、iframe、事件状态与页面上下文定时器。宿主页面与iframe通信检查来源及窗口对象；来源限制用于网站接入管理，不代替身份认证。

`locale`可选，默认根据浏览器选择`zh/en`；`setLocale()`更新访客界面和后续请求的`Accept-Language`，不重建iframe，保留消息、输入草稿、滚动位置与连接。已有消息及配置的品牌内容不自动翻译，公共脚本不读取业务站点自己的语言存储。

`linkRules`默认不配置，按顺序匹配完整URL前缀，`from`使用以`/`结尾的绝对地址，`to`可以是绝对地址或宿主网站的根相对路径。规则同时处理回复正文中的链接和底部来源，保留后续路径、查询参数及锚点，不修改知识库原始来源。未匹配的链接保持原地址。默认在新标签页打开；接入方可在初始化参数提供`onNavigate(url)`交给自己的页面路由，官网示例用它在常驻外壳中打开文档。按住Ctrl或Command点击仍沿用浏览器的新标签页行为。

`closeOnMobileNavigate`默认`false`；设为`true`后，手机宽度不超过600px时点击映射链接自动收起客服，再次打开保留原会话、消息滚动位置与输入草稿，桌面保持展开。以上选项均为可选，未配置的网站与独立聊天页面保持原有行为。网站常驻外壳仅属于simple官网，不属于客服公共脚本。

业务聊天可以复用当前身份获取一次性WS票据，无需读取widget的私有存储：

```js
const { ticket, ws_url } = await DawnSupport.createRealtimeTicket(resume)
const socket = new WebSocket(new URL(ws_url, location.href))
socket.onopen = () => socket.send(JSON.stringify({ type: 'auth', ticket }))
```

`resume`首次省略，重连时传入`ready.data.resume`。收到事件后按`channel/id`去重并发送`ack`；`resync_required`时重新获取业务列表。票据等待当前身份同步完成，只能使用一次。`onSessionChange({user, verified})`在身份同步后通知接入方，业务应清空旧账号状态并重新连接；退出时关闭业务WS。菜单、聊天面板与业务接口由接入项目实现，服务端仍检查成员权限。完整流程见[网站动态开发实例](moments.md)。

## 访客恢复与资料

访客使用高强度随机Bearer凭证，服务器只保存摘要。独立页面在客服域保存凭证，悬浮入口也在宿主网站保存对应应用的凭证，处理浏览器限制第三方存储的情况。相同凭证会恢复同一访客与咨询记录；清理浏览器存储、换浏览器或换域名时，匿名身份无法可靠关联，登录或外部身份认证后可恢复对应账号。

系统采集请求IP与UA，浏览器补充语言、时区、屏幕、窗口、当前页面、标题、来源页面和UTM参数。浏览器资料可被调用方修改，只用作咨询上下文；身份验证状态单独展示。不开启浏览器指纹，也不使用IP或手机号自动合并账号。

启用输入预览后，访客输入通过WS节流传给当前接待客服，最新快照存入成员扩展JSON。未发送内容不进入聊天消息、知识库或AI提示词。

## 本站用户认证

官网注册与登录使用`/api/account/registrations`、`/api/account/sessions`，本站登录凭证与后台管理员凭证分开。已登录用户为允许本站身份的客服应用获取一次性票据。

```text
POST /api/account/support-tickets
Authorization: Bearer 本站登录凭证
Content-Type: application/json

{"app_id": 1}
```

将响应的`data.ticket`交给`DawnSupport.identify({ticket})`。直接打开聊天页面时也可以使用`/chat/公开标识#ticket=一次性票据`，页面读取后移除fragment。票据只可使用一次。simple已内置双向游客同步和注册原地升级，见[本站接入](support-identity.md#simple接入)。

## 外部用户认证

外部产品在自己的服务器验证当前登录状态，再用对应客服应用的身份签名密钥生成HS256 JWT。必须包含`iss`、`sub`、`aud`、`exp`与唯一的`jti`。

```go
claims := jwt.MapClaims{
    "iss": "my-product", "sub": "customer-42", "aud": supportPublicKey,
    "exp": time.Now().Add(time.Minute).Unix(), "jti": rand.Text(),
}
proof, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(identitySecret))
```

浏览器只接收已签名声明，不能获得签名密钥。使用`DawnSupport.identify({identity: proof})`交换身份，同一声明不能重复使用。外部编号按应用和签发者隔离，不能通过传入另一个`user_id`查询其他人的数据。接入网站退出或切换账号时必须调用`reset()`；JWT的有效期用于一次性交换，交换后的客服会话按自己的有效期保存。

后台中已认证的用户编号显示为绿色，未认证资料显示为红色。验证表示用户编号来源已经确认；游客和注册用户分别判断，不表示浏览器、UA、IP及其他扩展字段也已验证。

## HTTP与实时协议

公开接口位于`/api/support/{public_key}`，后台接口位于`/admin`。HTTP与WS均使用Bearer或一次性票据，不依赖第三方Cookie。`WS_PUBLIC_URL`留空时使用当前访问域名、WS端口和路径，支持直接访问HTTP服务上的访客页面。反向代理部署时设置`WS_PUBLIC_URL=/ws`并把该路径转发到WS服务，也可填写完整公网WSS地址。

| 接口 | 用途 |
|---|---|
| `GET /api/support/{key}/configuration` | 获取可公开的品牌与入口设置 |
| `POST /api/support/{key}/identities` | 建立或恢复身份，不创建咨询 |
| `POST /api/support/{key}/sessions` | 建立或恢复访客，交换可信身份 |
| `DELETE /api/support/{key}/session` | 注销当前客服凭证 |
| `GET /api/support/{key}/conversations` | 当前用户的应用内会话 |
| `GET/POST /api/support/{key}/conversations/{id}/messages` | 历史消息与发送 |
| `PUT /api/support/{key}/conversations/{id}/read-position` | 单调更新阅读位置 |
| `POST /api/support/{key}/conversations/{id}/human-requests` | 申请人工 |
| `GET /api/support/{key}/conversations/{id}` | 当前会话与最近人工接待状态 |
| `POST /api/support/{key}/conversations/{id}/endings` | 客户直接结束当前排队或人工接待，传`session_id`、`version` |
| `PUT /api/support/{key}/conversations/{id}/closing-request` | 客户确认或拒绝结束，传`session_id`、`version`、`approved` |
| `GET /api/support/{key}/conversations/{id}/support-sessions/{session_id}` | 查看指定接待与评价状态 |
| `POST /api/support/{key}/conversations/{id}/ratings` | 评分并同意结束，或评价已结束接待，传`session_id`、`version`、`rating`（1差评/2一般/3好评）、`comment`（可选） |
| `POST /api/support/{key}/conversations/{id}/messages/{message_id}/retries` | 重试当前问题的失败AI回复 |
| `POST /api/support/{key}/realtime-tickets` | 获取一次性WS登录票据 |
| `PUT /api/support/{key}/actions/{id}` | 确认或拒绝已经展示的工具操作 |

WS连接后首先发送`auth`，成功后接收`ready`与恢复凭证。事件按`channel`和`id`去重并确认；流式回复另带执行次数、版本与递增片段序号。每个客服应用使用独立身份域，不能用一个应用的访客凭证订阅另一个应用的消息。

开发者可继续使用内置`realtime`实例注册业务消息和频道，不另建连接。`Hub.Handle`注册消息，`Hub.Validate`验证身份，`Hub.Authorize`控制频道订阅，`Hub.AuthorizeEvent(ctx, principal, message)`在实际投递时重新校验消息。后台客服使用会话与应用范围检查；更改钩子时应保留这些检查。

```ts
createAdmin({
  support: { customerInfo: CustomerBusinessCard },
  pages: { '/support/apps': () => import('./MySupportApps.vue') }
})
```

客户业务卡片接收`customer`与`conversation`，可通过公开的`request`调用业务接口。业务接口仍需执行后台权限和数据范围检查。

接待事件中的`conversation.support_session`包含最近人工接待。客户端只接受不低于当前`conversation.version`的状态快照，重连时重新获取当前状态；历史系统消息只负责展示，不覆盖当前状态。结束卡使用`support_session.id/version`校验，避免旧卡关闭新接待。访客主题按钮按跟随系统、亮色、暗色循环切换并记住选择。表情按完整字素显示，字号与行内文字一致，不拆分肤色、组合表情或旗帜。
