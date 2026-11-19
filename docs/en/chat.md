# Chat and website support

The header chat entry supports both admin private messages and customer support. Regular administrators use a small window; support agents use a larger window with a conversation list, messages, and customer details. One capsule at the top right cycles through private chat → support → fullscreen support → private chat, alongside theme and close icons. The top left shows the current administrator's avatar, today's replies, and today's customers. Server permissions remain enforced; support access does not expose other administrators' private chats.

Opening the window shows all permitted conversations, prioritizing unread ones. Reading history does not require reception mode and does not claim a conversation. Opening a conversation marks it read and clears its badge. Enabling Start receiving assigns waiting customers according to app permissions and capacity; agents can reply only to assigned conversations. Disabling the switch stops new assignments and disables the composer. Heartbeats never enable reception automatically. Daily statistics use UTC calendar days, as does the dashboard. Replies count successfully sent public agent text; customers are deduplicated across claims and transfers that day. Private chats, internal notes, AI replies, and failed sends are excluded. AI, waiting, human, and closed states use pale blue, orange, green, and gray labels.

See [AI support](ai.md) for configuration and knowledge, [Users](users.md) for models and IDs, and [GeoIP](geoip.md) for IP details.

Visitor and admin chat support Unicode emoji. Visitor chat uses the admin private-chat avatar, bubble, and color style, with light, dark, and system themes.

## Standalone chat link

```text
https://support.example.com/chat/public-app-key?user=alice&user_id=42&mobile=13800000000&nick=Alice
```

Extra query parameters become application-specific customer attributes visible to agents. `entry` selects an entry, `theme` selects `light`, `dark`, or `system`, and `locale` selects `zh` or `en`. Unknown business fields retain their names and are never credentials. Set allowed origins to `*` or leave empty to allow any website. See [Identity integration](support-identity.md) for simple and external sites.

Prefer JS attributes or `postMessage` for sensitive information; query strings may enter host website access logs. Never put authentication tickets in ordinary query parameters.

## Floating widget

```html
<script src="https://support.example.com/support-assets/widget.js"
        data-app="public-app-key" data-theme="system" data-locale="en" defer></script>
```

Drag the launcher vertically; it snaps to either screen edge and remembers its position on that website. When collapsed, only the button receives clicks, with no fullscreen transparent overlay. Chat runs in an iframe; Shadow DOM isolates launcher styles. Mobile chat opens fullscreen. On desktop, clicking outside the chat window temporarily hides it; reopening preserves the conversation, draft, and scroll position.

Pass dynamic attributes through initialization:

Optionally configure `menu` to add custom launcher items. `label` is text or a function receiving the locale, `icon` is an image URL, and `onClick` receives the support API. Set `menuLayout` to `vertical` (default) or `horizontal`. Without a menu the launcher opens support directly.

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
  app: 'public-app-key',
  theme: 'system',
  locale: 'en',
  linkRules: [{ from: 'https://docs.example.com/product/', to: '/docs/' }],
  closeOnMobileNavigate: true,
  attributes: { user: 'alice', user_id: '42', mobile: '13800000000', nick: 'Alice', plan: 'standard' }
})
DawnSupport.open()
DawnSupport.close()
DawnSupport.setTheme('dark')
DawnSupport.setLocale('zh')
DawnSupport.identify({ attributes: { plan: 'pro' } }) // Unverified context only
DawnSupport.identify({ identity: signedIdentity, attributes: { plan: 'standard' } })
// After the website signs out or switches accounts, revoke its current support identity.
DawnSupport.reset()
```

`data-app` enables automatic initialization. Do not combine it with manual JS initialization. `destroy()` removes the launcher, iframe, state, and context timer. Host and iframe messages validate origin and window source. Allowed origins manage website integration; they do not replace authentication.

`locale` is optional and defaults to the browser language (`zh/en`). `setLocale()` updates visitor interface labels and API `Accept-Language` without reloading the iframe. Messages, composer draft, scroll, and connection remain intact; existing message contents and authored branding are not translated. The generic widget does not read application-specific language storage.

`linkRules` is optional. Rules match complete URL prefixes in order. `from` is an absolute URL ending in `/`; `to` may be absolute or host-root-relative. Rules apply to both message links and source citations and retain suffix paths, query strings, and anchors. Original knowledge sources remain unchanged. Unmatched links retain their address. Links open in new tabs by default; provide `onNavigate(url)` to use your router. The example website opens docs inside its persistent shell. Ctrl/Command clicks retain normal browser new-tab behavior.

`closeOnMobileNavigate` defaults to `false`. When enabled, clicking a mapped link on screens up to 600px closes chat; reopening preserves the conversation, scroll, and draft. Desktop chat stays open. These options do not alter unconfigured sites or standalone chat. The persistent website shell belongs to simple, not the shared support script.

Business chat can reuse the current identity through a single-use WS ticket without reading private widget storage:

```js
const { ticket, ws_url } = await DawnSupport.createRealtimeTicket(resume)
const socket = new WebSocket(new URL(ws_url, location.href))
socket.onopen = () => socket.send(JSON.stringify({ type: 'auth', ticket }))
```

Omit `resume` initially and use `ready.data.resume` when reconnecting. Deduplicate events by `channel/id` and send `ack`; refetch business state on `resync_required`. Ticket creation waits for identity synchronization and each ticket works once. `onSessionChange({user, verified})` notifies the host after synchronization: discard the previous account's state and reconnect. Close business sockets on logout. The host owns its menus, chat panel, and APIs, with server-side member authorization. See the [community example](moments.md) for the complete flow.

## Visitor recovery and attributes

Visitors use high-entropy random Bearer credentials; the server stores only digests. Standalone chat stores credentials on the support domain. Widgets also retain app-specific credentials on the host website to work with third-party storage restrictions. Valid credentials restore the same visitor and history. Clearing storage or changing browser/domain prevents reliable anonymous matching; authenticated local or external identities can recover their accounts.

Requests supply IP and UA. Browsers add language, timezone, screen and viewport sizes, page URL/title, referrer, and UTM parameters. These browser values are mutable context; verification status is separate. Dawn does not fingerprint browsers or merge accounts by IP or phone number.

When input preview is enabled, throttled WS events send the current draft to the assigned agent; the latest snapshot is stored in member extension JSON. Unsent text never enters chat messages, knowledge, or AI prompts.

## Local user authentication

Website registration and login use `/api/account/registrations` and `/api/account/sessions`. Website credentials are separate from admin credentials. Signed-in users obtain single-use support tickets for apps permitting local identity.

```http
POST /api/account/support-tickets
Authorization: Bearer website-token
Content-Type: application/json

{"app_id": 1}
```

Pass `data.ticket` to `DawnSupport.identify({ticket})`. Standalone chat also accepts `/chat/public-key#ticket=single-use-ticket` and removes the fragment after reading it. Tickets can be used once. simple includes bidirectional guest synchronization and in-place registration; see [simple integration](support-identity.md#simple-integration).

## External user authentication

External products verify their users on their own server, then sign an HS256 JWT with the support application's identity secret. Include `iss`, `sub`, `aud`, `exp`, and a unique `jti`.

```go
claims := jwt.MapClaims{
    "iss": "my-product", "sub": "customer-42", "aud": supportPublicKey,
    "exp": time.Now().Add(time.Minute).Unix(), "jti": rand.Text(),
}
proof, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(identitySecret))
```

The browser receives only the signed assertion, never the secret. Exchange it through `DawnSupport.identify({identity: proof})`. Assertions are single-use. External IDs are isolated by application and issuer; submitting another `user_id` cannot access that person's data. Call `reset()` when signing out or switching website accounts. JWT expiry applies to the exchange; the resulting support session has its own lifetime.

Verified user IDs appear green in admin; unverified attributes appear red. Verification confirms the source of a user ID, independently of guest/member type. It does not verify UA, IP, or other browser attributes.

## HTTP and realtime protocol

Public APIs live under `/api/support/{public_key}`; admin APIs live under `/admin`. HTTP and WS use Bearer credentials or single-use tickets, without third-party Cookies. Empty `WS_PUBLIC_URL` uses the current hostname, WS port, and path for direct HTTP visitor-page access. Behind a reverse proxy, set `WS_PUBLIC_URL=/ws` and proxy it to WS, or configure a full public WSS URL.

| API | Purpose |
|---|---|
| `GET /api/support/{key}/configuration` | Public branding and entry settings |
| `POST /api/support/{key}/identities` | Create or restore identity without opening a consultation |
| `POST /api/support/{key}/sessions` | Create or restore visitor chat and exchange trusted identity |
| `DELETE /api/support/{key}/session` | Revoke current support credentials |
| `GET /api/support/{key}/conversations` | Current user's conversations in the app |
| `GET/POST /api/support/{key}/conversations/{id}/messages` | Message history and sending |
| `PUT /api/support/{key}/conversations/{id}/read-position` | Monotonic read-position update |
| `POST /api/support/{key}/conversations/{id}/human-requests` | Request an agent |
| `GET /api/support/{key}/conversations/{id}` | Current conversation and latest human session |
| `POST /api/support/{key}/conversations/{id}/endings` | Visitor immediately ends waiting or human service; `session_id`, `version` |
| `PUT /api/support/{key}/conversations/{id}/closing-request` | Accept or decline closure; `session_id`, `version`, `approved` |
| `GET /api/support/{key}/conversations/{id}/support-sessions/{session_id}` | View a service round and rating state |
| `POST /api/support/{key}/conversations/{id}/ratings` | Rate and accept closure, or rate a completed round; `session_id`, `version`, `rating` (1 poor/2 neutral/3 good), optional `comment` |
| `POST /api/support/{key}/conversations/{id}/messages/{message_id}/retries` | Retry a failed AI answer to the current question |
| `POST /api/support/{key}/realtime-tickets` | Single-use WS login ticket |
| `PUT /api/support/{key}/actions/{id}` | Confirm or decline a displayed tool action |

Send `auth` first after connecting WS. On success, receive `ready` and recovery credentials. Deduplicate and acknowledge events by `channel` and `id`. Streams additionally carry execution attempt, version, and increasing fragment sequence. Each support app has its own identity realm; credentials from one app cannot subscribe to another app's messages.

Developers can register business messages and channels on the existing `realtime` instance without another connection. `Hub.Handle` registers messages, `Hub.Validate` validates identity, `Hub.Authorize` controls subscriptions, and `Hub.AuthorizeEvent(ctx, principal, message)` rechecks actual deliveries. Backend support checks conversation and app scopes; preserve these checks when changing hooks.

```ts
createAdmin({
  support: { customerInfo: CustomerBusinessCard },
  pages: { '/support/apps': () => import('./MySupportApps.vue') }
})
```

The business card receives `customer` and `conversation`. Use public `request` for business APIs, which must still enforce admin permissions and scopes.

Conversation events include the latest human round in `conversation.support_session`. Clients accept snapshots only when `conversation.version` is current or newer and refetch state after reconnecting. Historical system messages are display-only. Closure cards use `support_session.id/version`, preventing stale cards from closing a later round. The theme button cycles system, light, and dark and retains the choice. Emoji render as whole graphemes aligned with inline text, without splitting skin tones, composites, or flags.
