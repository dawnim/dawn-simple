# Bidirectional website and support identity

simple includes integration with a shared `users` table. External websites retain their user system and need only one backend endpoint and one JS identity callback. Initializing identity does not create a consultation, queue entry, or agent assignment.

Set allowed website origins to `*` or leave empty for any website; alternatively enter full origins such as `https://www.example.com`. Cross-origin integration uses Bearer and does not send website Cookies to Dawn. See [Configuration](configuration.md) for global CORS. Origin settings do not replace backend identity verification.

## simple integration

The official demo app enables `allow_local_users`; the website loads its public key automatically. Opening the website restores its session. An existing support guest is claimed as the same user; otherwise the site creates a guest and shares its identity with support.

- Guests see Complete profile and can register or sign into an existing account
- Registration sets username, password, and nickname on the existing `users` row, preserving internal ID, `unique_id`, and history
- Existing-account login transfers the current app's guest consultations while retaining conversation IDs, human rounds, and ratings
- Signing out revokes the current website and support credentials and creates an independent guest; other browsers remain signed in
- Storage events synchronize identity across tabs in the same browser

The website's `POST /api/support-identities` is restricted to the official `dawn` app; clients cannot choose another app. It calls the public `LocalIdentity` API and returns public website information, a website token only when newly created, and a support ticket. See [dawn-simple/app/web](https://github.com/dawnim/dawn-simple/tree/main/app/web).

Other websites sharing Dawn's user table can call it from their own endpoint:

```go
session, ticket, err := manager.LocalIdentity(ctx, appID, accountToken, visitorTicket)
if err != nil {
    return err
}
// accountToken comes from the current site request; appID is server-configured.
// Store a newly issued session.Token in the site's existing login store.
result := map[string]any{
    "user": session.User.Profile(), "token": session.Token,
    "support": map[string]any{"ticket": ticket},
}
```

`LocalIdentity` claims guests only. Registered accounts require normal authentication and cannot be recovered through guest tickets. `Register(ctx, current, username, password, nickname)` upgrades a local guest session in place; passing an empty `customer.Session{}` creates a new account. Upgrading invalidates all old guest credentials.

## External websites: one endpoint and one script

### 1. Configure the app

Create a support application with a website entry and identity signing secret. Give the public key to the frontend; keep the secret on the external server. External integration needs neither `allow_local_users` nor direct access to Dawn's SQL.

External IDs are isolated by support app + `iss` + `sub`. A website ID such as `608` and Dawn's public ID retain their own values.

### 2. Provide a website identity endpoint

For example, `POST /api/support-identities` accepts an optional `visitor_ticket`. Resolve the current user from the website's Cookie or Bearer. Never claim an account using a browser-submitted `user_id`.

1. If the website already has a guest or member, use that user
2. With no website user but a visitor ticket, exchange it server-to-server and create or recover a local guest using the stable `visitor_id`
3. Otherwise create a guest according to the website's existing rules
4. Sign an identity assertion and return public details and `support.identity`

```json
{
  "data": {
    "user": { "id": "608", "nickname": "Guest608", "type": "guest" },
    "support": { "identity": "server-signed-JWT" }
  }
}
```

Use a standard JWT library, such as Go's `github.com/golang-jwt/jwt/v5`:

```go
claims := jwt.MapClaims{
    "iss": "my-site", "sub": userID, "aud": appPublicKey,
    "type": userType, // guest or member, selected by the website backend
    "exp": time.Now().Add(time.Minute).Unix(), "jti": rand.Text(),
    "user": map[string]any{"nickname": nickname, "avatar": avatar},
}
proof, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
    SignedString([]byte(identitySecret))
```

`sub` must be a stable nonempty string; `iss` is a fixed website identifier. Generate a new `jti` for each single-use assertion. Omitted `type` means `member`. Read `nickname/avatar/email/mobile` in `user` from trusted site data. Neither the signing secret nor website login credentials are sent to the browser/Dawn respectively.

### 3. Accept a support-created guest

This is needed only when a request includes `visitor_ticket` and the website has no current user. The website server sends:

```http
POST https://support.example.com/api/support/public-app-key/visitor-ticket-exchanges
Authorization: Bearer exchange-purpose-JWT
Content-Type: application/json

{"ticket":"visitor_ticket-value"}
```

Sign the exchange JWT using the same app secret and include:

```go
claims := jwt.MapClaims{
    "iss": "my-site", "sub": "server", "aud": appPublicKey,
    "scope": "visitor_exchange",
    "exp": time.Now().Add(time.Minute).Unix(), "jti": rand.Text(),
}
```

Success returns 201:

```json
{"data":{"visitor_id":"7539553662","type":"guest","external_user_id":"608"}}
```

`external_user_id` is returned only when the visitor is already linked to the same `iss`. Restore that user first. Otherwise create a guest with a unique site-side key of app public key + `visitor_id`, preventing duplicates during retries. Establish the website session, then return its signed identity as above. The SDK links it to the original support visitor.

Tickets last 1 minute and can be claimed by one recipient. The same recipient may retry within that time with a new JWT `jti` for each request. Revoking guest credentials or registering the user invalidates the ticket. A public `visitor_id` alone cannot claim a user.

### 4. Load the support script

Load without `data-app`, then call `init`, so automatic initialization does not run before callback registration:

```html
<script src="https://support.example.com/support-assets/widget.js"></script>
<script>
DawnSupport.init({
  server: 'https://support.example.com',
  app: 'public-app-key',
  resolveIdentity: async ({ visitor_ticket }) => {
    // Use the site's existing request client and its Cookie or Bearer.
    const result = await api.post('/api/support-identities', { visitor_ticket })
    // Synchronize result.user into the site's user state here.
    return result.support
  },
  onError: error => console.error(error.message)
})
</script>
```

This example assumes `api.post` unwraps `data` in its shared request layer. With native `fetch`, read `result.data` once. The SDK prepares guest tickets, checks exchanges, and stores support credentials; no manual iframe or `postMessage` work is required.

```js
// After website registration or login succeeds
await DawnSupport.syncIdentity()

// After website sign-out completes and its local login state is cleared
await DawnSupport.reset()
```

`syncIdentity()` invokes the callback as needed and retains support credentials for unchanged identities. `reset()` revokes the current support credential, clears the identity window, and uses the callback to establish an independent guest. Use `close()` merely to hide chat while retaining its content and draft.

Sites without identity integration can load the ordinary `data-app` script for anonymous support. Direct `identify({identity})` or `identify({ticket})` calls are also supported without a callback. Optional `onSessionChange({user, verified})` exposes public information only, never login credentials.

## Attributes, permissions, and merging

Verification and account type are independent: a server-confirmed guest may be verified while still lacking member capabilities. Tool `require_identity` requires a verified identity; `require_account` additionally requires a registered user. Go tools read the account type from `ToolActor.UserType` captured for the request, never from model arguments.

The website owns account information; support owns consultations, service rounds, ratings, and app-specific notes. Frontend `attributes` are context only. Registration/login transfers eligible guest conversations within the current app, retaining conversation IDs and rounds while cancelling old AI runs and pending actions. Other apps and accounts are not merged. Old guest assertions cannot downgrade members; old guest tickets cannot recover account history after sign-out.

Prefer retaining the website's guest ID during registration. If registration creates a new ID, the new signed identity and existing support guest credential jointly establish ownership. Dawn keeps the consultations and switches identity. The website handles its own guest orders, carts, and other business data.

## Public APIs

| Entry | Purpose |
|---|---|
| `POST /api/support/{key}/identities` | Establish or restore identity without a consultation |
| `POST /api/support/{key}/visitor-tickets` | Current support guest requests a reverse ticket using support Bearer |
| `POST /api/support/{key}/visitor-ticket-exchanges` | External server verifies a reverse ticket using an exchange-purpose JWT |
| `Manager.Identify` | Establish or restore identity in Go |
| `Manager.LocalIdentity` | Website integration with shared `users` |
| `Manager.ExchangeVisitor` | Verify an external server's reverse exchange in Go |
| `Manager.ClaimVisitor` | Trusted backend claims a guest with a server-fixed recipient |

For messages, human service, standalone links, and themes, see [Chat and website support](chat.md).
