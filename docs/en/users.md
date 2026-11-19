# Users and public IDs

Dawn provides a built-in `users` system. Embed `models.User` in a business model to add fields. Administrators use `models.AdminUser` and `admin_users`; their credentials and permissions are separate. `AdminUser.UserID` can link a regular user, but the link itself grants no admin permissions.

## User types

`type` supports `guest`, `member`, and `ai`. Each support application owns one AI user. Application-specific visitor attributes live in `user_traits`. Backend authorization uses admin identity, independent of regular-user type.

## Authentication and visitor recovery

Regular users and visitors authenticate through PrismGo Session and Redis. SQL retains users, attributes, and chat records. Website and support credentials are validated separately and cannot be reused across support applications. External identity requires a signed assertion or a local single-use ticket.

Browsers retain random visitor credentials, valid for 180 days by default. Reopening chat with valid credentials restores the same identity and history. Clearing browser storage or losing Redis authentication data prevents anonymous recovery through nickname, IP, or UA. Verified accounts can recover history after authenticating again. Registering a local guest upgrades the existing user and preserves its ID and history. Signing into an existing account merges guest consultations in the current application; old credentials are revoked after commit. See [Website and support identity](support-identity.md).

For login, logout, and framework configuration, see [Chat](chat.md) and [Authentication and sessions](authorization.md#authentication-and-sessions).

## Visitor management

The admin visitor directory is global. Permissions `dawn.members.list/create/update/delete` independently control viewing, creation, updates, and deletion. `GET /admin/members` supports `query`, `type`, `status`, and standard pagination, defaulting to 10 rows ordered by descending ID. `GET /admin/members/{id}` returns one user.

`POST /admin/members` creates a regular user with required `username` (3–80 characters) and `password` (at least 6 characters, at most 72 bytes). Writable profile fields are `nickname`, `avatar`, `email`, `mobile`, `locale`, and `status`, limited to 100, 500, 180, 40, and 16 characters, and status 1 (enabled) or 2 (disabled). User type, internal ID, public ID, and audit timestamps are assigned by the server.

`PATCH /admin/members/{id}` changes submitted profile fields; `PUT` resets all profile fields, using defaults for omissions. `DELETE /admin/members/{id}` deletes one user; `DELETE /admin/members?ids=1,2` deletes an explicit batch in one transaction. If any row is missing, no rows are deleted. Disabling or deleting a user revokes their login credentials.

## Public IDs

`users.id` is the internal primary key. `users.unique_id` is a unique string derived from it. Creating individual `models.User` records through the ORM fills the public ID in the same transaction through a hook. Direct SQL inserts must call the same codec and write it within their transaction.

Initial installation generates a random 256-bit key in the `__user_identity` record of `admin_settings`. It is hidden from settings forms and has no editing API. Reinstallation never overwrites existing configuration. IDs use radix-9 AES-FF1 format-preserving encryption with alphabet `012356789` and no leading zero. Starting at 7 digits, each contiguous range of original IDs has one fixed width. Width increases one digit after its range is exhausted. Rules are fixed in framework code, without a width setting.

`identity.Load(ctx, db)` reuses the global `ForeverSettings` group Hash: group `__user_identity`, field `key`. There is no separate identity cache. It uses the application's existing Redis connection and application/environment namespace, without expiry, and retains the `__revision` version. Internal groups starting with two underscores are excluded from admin viewing, editing, and clearing. Cache hits do not query SQL; a miss reloads the same database record without generating a new key. After committing, `db:seed` clears this group cache to prevent reinstallation from retaining old configuration.

| Absolute integer value | ID width |
| --- | --- |
| 1–1,889,568 | 7 |
| 1,889,569–18,895,680 | 8 |
| 18,895,681–171,950,688 | 9 |
| 171,950,689–1,549,445,760 | 10 |
| 1,549,445,761–13,946,901,408 | 11 |

Each tier adds `32 × 9^(width-2)` values per sign. Ranges are contiguous and do not overlap or skip original IDs. Each tier adds nine times the capacity of the preceding tier. Widths 7–10 together cover about 1.549 billion positive IDs, with a separate equal capacity for negatives. The maximum width is 21 digits, covering the entire nonzero `int64` range.

Only the range containing the original ID's absolute value determines its width. Width selection is not random and does not depend on current user counts or the largest stored ID. With a fixed key, the same integer always produces the same public ID, and different integers cannot collide. Positive and negative values with the same magnitude use the same width.

The middle digit at zero-based index `len(value)/2` is odd for positive numbers and even for negative numbers. Even widths use the right of the two middle positions. User primary keys remain positive; negative numbers are available for other business uses.

Encoding subtracts the range's first value to obtain a zero-based offset, maps that offset into the valid character domain, and repeats FF1 permutations until the leading-digit and sign constraints hold. Decoding finds the range from the public ID's width, reverses the permutation, adds back the range's first value, and restores the sign. Every nonzero `int64` input can be encoded and recovered; strings outside the character rules or unassigned output strings may be invalid. Recovering the original value does not query a database.

**Back up this record with the database. Never change or delete it.** A cache miss with a missing installation record, or invalid configuration, returns an error rather than silently generating another key. `dawn:install` is for full initialization. Restoring a database with existing users requires its matching identity configuration and clearing any identity cache from the previous environment. Use `migrate:fresh --seed` when rebuilding a database so its schema and cache are updated together.

```go
import "github.com/dawnim/dawn-server/support/identity"

codec, err := identity.Load(ctx, db)
if err != nil {
    return err
}
uniqueID, err := codec.ID2Big(int64(user.ID))
if err != nil {
    return err
}
internalID, err := codec.Big2ID(uniqueID)
if err != nil {
    return err
}
_ = internalID

// The general-purpose codec also supports negatives and returns int64.
encoded, err := codec.ID2Big(-123)
if err != nil {
    return err
}
original, err := codec.Big2ID(encoded) // original == -123
if err != nil {
    return err
}
_ = original
```

Call `Load` once per batch and reuse `codec`. `ID2Big` and `Big2ID` perform only in-memory computation without accessing SQL or Redis; an acquired codec is safe for concurrent use. Standalone programs can construct `identity.Codec{Key: key}` directly, where `key` is the original installation key as 64 hexadecimal characters, without initializing a database or cache.

Public profiles and support responses omit internal user primary keys. Visitor realtime identities and message subjects use `unique_id`; admin management still operates by internal primary key.

Always treat public IDs as strings in JSON and frontend code to avoid JavaScript integer precision loss. `ID2Big` accepts nonzero `int64` values, and `Big2ID` returns `int64`, including its minimum negative value. Zero, invalid characters, leading zeros, and out-of-tier values return errors translatable for the request locale.

The permutation hides sequential changes within a tier. Length reveals the approximate absolute-value range, and the middle digit reveals the sign. IDs are not authentication credentials and do not provide tamper detection; reads and writes still require credentials and scopes. Preserve the key and rules for stored IDs. Changing rules requires rebuilding IDs and their references, without legacy encoding compatibility.

## Built-in table identification

Built-in tables retain responsibility-based names such as `admin_*`, `chat_*`, and `users`, without an extra product prefix. Supported databases use table comments beginning with `系统:`, for example `系统:普通用户表`. Business tables use their own comments. SQLite has no native table comments.
