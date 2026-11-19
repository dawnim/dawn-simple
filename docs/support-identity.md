# 网站与客服身份双向接入

simple内置接入，网站和客服共用`users`。外部网站保留自己的用户体系，只需一个站点后端接口和一个JS身份回调。客服身份初始化不会创建咨询、排队或占用客服容量。

应用的网站入口中，允许的网站来源填写`*`或留空表示任意网站；也可以填写完整来源，例如`https://www.example.com`。跨域接入使用Bearer，不向Dawn携带网站Cookie；服务全局跨域配置见[部署配置](configuration.md)。网站来源配置不代替后端身份验证。

## simple接入

官方演示应用开启“允许本站用户身份同步”（`allow_local_users`），官网自动加载其公开标识。进入官网时恢复本站会话；已有客服游客时认领同一用户，否则创建本站游客，再将其身份交给客服。

- 游客右上角显示“完善信息”，可以注册或切换到已有账号登录
- 注册在原`users`记录设置账号、密码、昵称，保留内部ID、`unique_id`与咨询记录
- 登录已有账号时，将当前应用内的游客咨询归入该账号，保留会话ID、人工接待和评价
- 退出时注销当前网站和客服凭证，恢复独立游客；其他浏览器不会因此被注销
- 同一浏览器的标签页通过网站会话存储事件同步身份

官网的`POST /api/support-identities`只允许官方`dawn`应用，客户端不能指定别的应用。它调用扩展公开的`LocalIdentity`，返回本站公开资料、需要新建时的本站登录凭证，以及客服票据。源码位于[dawn-simple/app/web](https://github.com/dawnim/dawn-simple/tree/main/app/web)。

其他共用Dawn用户表的网站可在自己的接口中直接调用：

```go
session, ticket, err := manager.LocalIdentity(ctx, appID, accountToken, visitorTicket)
if err != nil {
    return err
}
// accountToken来自当前网站请求；appID由网站后端配置，不能由访客任意指定
// 将新生成的session.Token交给网站已有的登录存储，空值表示沿用现有凭证
result := map[string]any{
    "user": session.User.Profile(), "token": session.Token,
    "support": map[string]any{"ticket": ticket},
}
```

`LocalIdentity`只认领游客。注册账号需要正常登录，不能用游客票据恢复注册账号。公开的`Register(ctx, current, username, password, nickname)`在`current`为本站游客会话时原地升级；传空`customer.Session{}`则创建新账号。游客升级后，旧游客凭证全部失效。

## 第三方：一个后端接口与一段JS

### 1. 配置应用

在后台创建客服应用，配置网站入口和“身份签名密钥”。将公开标识交给前端，将签名密钥保存在第三方服务器。外部接入不需要开启`allow_local_users`，也不需要连接Dawn数据库。

第三方用户ID按`客服应用＋iss＋sub`隔离，例如站点用户`608`与Dawn用户编号各自保持原值。

### 2. 网站后端提供身份接口

例如`POST /api/support-identities`，接收可选的`visitor_ticket`。接口必须根据网站自己的Cookie或Bearer读取当前用户，不使用浏览器提交的`user_id`认领账号。

1. 网站已有游客或注册用户：直接使用当前用户
2. 网站没有用户且有游客票据：向Dawn后端交换，按返回的稳定`visitor_id`创建或恢复本站游客
3. 两者都没有：按网站原有规则创建游客
4. 用应用密钥签发该用户的身份声明，返回公开资料与`support.identity`

```json
{
  "data": {
    "user": { "id": "608", "nickname": "游客608", "type": "guest" },
    "support": { "identity": "后端签发的JWT" }
  }
}
```

签名使用标准JWT库。以下Go示例使用`github.com/golang-jwt/jwt/v5`：

```go
claims := jwt.MapClaims{
    "iss": "my-site", "sub": userID, "aud": appPublicKey,
    "type": userType, // guest或member，由网站后端决定
    "exp": time.Now().Add(time.Minute).Unix(), "jti": rand.Text(),
    "user": map[string]any{"nickname": nickname, "avatar": avatar},
}
proof, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
    SignedString([]byte(identitySecret))
```

`sub`必须为稳定的非空字符串，`iss`是固定的站点标识。`jti`每次生成，声明只能交换一次；`type`省略时表示`member`。`user`支持`nickname/avatar/email/mobile`，从站点可信资料读取。密钥不进入浏览器，用户登录凭证也不传给Dawn。

### 3. 接收客服先创建的游客

只有请求带有`visitor_ticket`且网站当前没有用户时，才需要这一过程。网站服务器发送：

```http
POST https://support.example.com/api/support/应用公开标识/visitor-ticket-exchanges
Authorization: Bearer 后端交换用途JWT
Content-Type: application/json

{"ticket":"visitor_ticket的值"}
```

交换用途JWT仍用同一个应用密钥签发，包含：

```go
claims := jwt.MapClaims{
    "iss": "my-site", "sub": "server", "aud": appPublicKey,
    "scope": "visitor_exchange",
    "exp": time.Now().Add(time.Minute).Unix(), "jti": rand.Text(),
}
```

成功返回201：

```json
{"data":{"visitor_id":"7539553662","type":"guest","external_user_id":"608"}}
```

`external_user_id`仅在该游客已经绑定相同`iss`时返回。优先恢复该用户，否则用`应用公开标识＋visitor_id`作为站点内唯一绑定键创建游客，防止网络重试重复建人。站点完成自己的会话后，按上节返回该游客的签名身份，SDK会把它关联到原客服访客。

票据有效期1分钟，只能由一个接收方认领；有效期内同一接收方可以重试，但每次请求的JWT使用新的`jti`。原游客凭证已注销或用户已注册时票据失效。不能只凭公开的`visitor_id`认领用户。

### 4. 加载客服JS

先加载脚本，不设置`data-app`，再调用`init`，避免自动初始化先于回调注册：

```html
<script src="https://support.example.com/support-assets/widget.js"></script>
<script>
DawnSupport.init({
  server: 'https://support.example.com',
  app: '应用公开标识',
  resolveIdentity: async ({ visitor_ticket }) => {
    // api沿用网站自己的请求工具，向本站发送Cookie或Bearer
    const result = await api.post('/api/support-identities', { visitor_ticket })
    // 在这里将result.user同步到网站自己的用户状态
    return result.support
  },
  onError: error => console.error(error.message)
})
</script>
```

上例`api.post`已经在统一请求层解包`data`；使用原生`fetch`时自行读取一次`result.data`。SDK负责准备游客票据、验证交换结果与保存客服凭证，不需要自行操作iframe或`postMessage`。

```js
// 网站注册或登录成功后
await DawnSupport.syncIdentity()

// 网站退出登录完成、清除自身登录状态后
await DawnSupport.reset()
```

`syncIdentity()`按需调用回调，不变的身份沿用客服凭证。`reset()`注销当前客服凭证、清空当前身份窗口，再通过回调恢复独立游客。仅收起窗口使用`close()`，保留聊天及输入内容。

没有身份对接的网站只加载普通`data-app`脚本即可使用匿名客服。也可以直接调用`identify({identity})`或`identify({ticket})`，不使用回调。可选的`onSessionChange({user, verified})`只通知公开资料，不公开登录凭证。

## 资料、权限与归并

身份可信程度与账户类型分别显示：站点后端确认的游客可以是“已验证＋游客”，仍不具备注册账号能力。业务工具的`require_identity`要求已验证身份，`require_account`另外要求注册用户。Go工具从`ToolActor.UserType`读取触发请求时的账户类型，不接受模型修改。

网站负责账号资料，客服负责咨询、接待、评价和应用内备注。前端`attributes`仅作为上下文。注册或登录后，当前应用内应归属的游客会话整体转移，保留会话ID和接待记录，取消旧AI执行与待确认操作；其他应用和其他账号不会被合并。旧游客声明不能将注册账号降回游客，退出后也不能用旧游客票据恢复账号历史。

同站点游客注册时优先保持原站点用户ID。如果站点注册时生成了新ID，登录后的签名身份与当前客服游客凭证共同确认归属；Dawn会保留咨询并切换到新身份。接入网站自身的游客订单、购物车等数据由网站按自己的业务规则处理。

## 公共接口

| 入口 | 用途 |
|---|---|
| `POST /api/support/{key}/identities` | 建立或恢复客服身份，不创建咨询 |
| `POST /api/support/{key}/visitor-tickets` | 当前客服游客获取反向票据，需要客服Bearer |
| `POST /api/support/{key}/visitor-ticket-exchanges` | 外部服务器用交换用途JWT验证反向票据 |
| `Manager.Identify` | Go侧建立或恢复身份 |
| `Manager.LocalIdentity` | 共用users表的网站接入 |
| `Manager.ExchangeVisitor` | Go侧验证外部服务器的反向交换 |
| `Manager.ClaimVisitor` | 可信后端认领游客，接收方标识由后端固定 |

现有消息、人工接待、独立链接和主题配置见[聊天与网站客服](chat.md)。
