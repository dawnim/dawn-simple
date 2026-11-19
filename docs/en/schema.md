# Database schema maintenance

Dawn provides two complementary workflows: PrismGo migrations record ordered changes; `dawn:schema` aligns the database with the complete current definition. Migrations usually finish an upgrade; full synchronization can also reconcile missing columns, types, lengths, defaults, and removed objects.

Run commands from the application root. dawn-server provides the implementation; dawn-simple does not duplicate it.

## Preview and synchronize

```sh
# Run pending migrations
go run . migrate

# Optional: inspect changes without writing
go run . dawn:schema check

# Inspect and display changes, apply them, then recheck on success
go run . dawn:schema sync

# Optional: run a separate final check
go run . dawn:schema check
```

The preview lists planned changes, not SQL errors or completed operations. Attention markers indicate risk, not execution failures.

`check` is read-only. Each operation appears in its own paragraph, separated by a blank line, with the object, operation, risk level, current definition, and target definition. Empty definitions display `-`. Table/column removal, column alterations, constraints, and uniqueness changes are marked for attention. On ANSI terminals, only that operation's first line is red; the current/target headings are green/yellow, while definitions keep the terminal's default color. Finding differences still means the check completed successfully: the default exit status is 0 and the count is shown. Use `--strict` when CI should return a nonzero status for drift.

`sync` inspects and displays current differences, applies them, and rechecks on success. A separate `check` is optional and helps with manual review; `sync` does not reuse a previous check result. Running `sync` directly also inspects the database and immediately applies changes without waiting for interactive confirmation. There are no per-item confirmations or extra risk permissions. Production uses the usual `--force`:

```sh
go run . migrate --force
go run . dawn:schema sync --force
```

| Option | Purpose |
|---|---|
| `--database=mysql` | Use `database.connections.mysql`; defaults to the application's default connection |
| `--tables=orders,users` | Compare or synchronize only these tables for this invocation |
| `--json` | Return `action`, `changes`, and `in_sync`; each change includes a boolean `risk` |
| `--strict` | Return a nonzero status when `check` finds differences; the default is 0 after a completed check |
| `--force` | Permit production synchronization; does not disable data constraints or clear business data |
| `--ansi` / `--no-ansi` | Force or disable terminal colors |

## One complete definition

Built-in tables are defined once in dawn-server's `database/migrations/install.go`. Initial migration and synchronization share this definition. dawn-simple retains a single installation migration entry.

Customize the final target through `dawn.ServiceProvider.Schema`. Ordinary GORM structs define columns, indexes, and constraints. Put business definitions in `database/schema/schema.go`:

```go
package schema

import (
    "github.com/dawnim/dawn-server/database/migrations"
    sdk "github.com/dawnim/dawn-server/database/schema"
)

type Order struct {
    ID    uint   `gorm:"primaryKey"`
    Code  string `gorm:"size:80;not null;uniqueIndex"`
    Total int64  `gorm:"not null;default:0;check:orders_total,total >= 0"`
}

type User struct {
    migrations.User
    Nickname string `gorm:"size:200;not null;default:''"`
    Level    int    `gorm:"not null;default:1"`
}

func Configure(d *sdk.Definition) {
    d.Tables["orders"] = sdk.Table{Model: &Order{}, Comment: "Orders"}
    d.Tables["users"] = sdk.Table{Model: &User{}, Comment: "System: users"}
}
```

In the existing provider instance in `bootstrap/extensions.go`:

```go
import appschema "your/module/database/schema"

dawn.ServiceProvider{Configure: admin.Configure, Schema: appschema.Configure}
```

Map keys are physical table names. Add, replace, or `delete(d.Tables, "old_table")`. Built-in tables can be changed; there is no built-in-only management boundary. **Tables absent from the complete target are dropped, and extra columns, indexes, and constraints are removed.** Describe all existing business tables before adopting full synchronization and review `check` first.

`Keep` initially preserves PrismGo's `migrations`, SQLite's internal `sqlite_*`, and runtime-generated `knowledge_vectors_*` tables. Their lifecycles belong to the migration engine, database, and knowledge index builder. Applications may customize these table patterns. Explicit target definitions are still compared even when they match a `Keep` pattern.

Changes made elsewhere without updating the target cannot be inferred. Embedding a snapshot can add fields or override a field with the same Go name. Schema overrides do not automatically change runtime models or APIs; adapt their callers too.

## Combine with migrations

| Requirement | Workflow |
|---|---|
| Rename, move data, split tables, backfill before tightening constraints | Add an ordered PrismGo migration |
| Change types, lengths, defaults, comments, tables, columns, or indexes | Update the target, then run `check` and `sync` |
| Edit a snapshot inside an already executed migration | `migrate` does not replay it; synchronize the complete target |
| Roll back a released change | Native migration rollback; synchronization has no historical rollback |

Use either workflow or both. Business migrations may reuse a definition's `Create(db)` and `Drop(db)`. Synchronization leaves migration history, accounts, configuration values, and user identity keys unchanged. Seeders own initialization data.

Changing only a name in the target means drop-and-create. To retain data, rename explicitly in a native migration first, then update the target. Type conversions, non-null values, and uniqueness are validated by the database; conflicting data is not silently rewritten.

`AutoMigrate` is a third-party GORM function. It does not delete extra structure and cannot replace full synchronization. PrismGo's `schema.Bind(db).SyncModels(...)` only creates missing model tables and columns. Neither third-party API has been removed.

## Database behavior and startup

| Database | Behavior |
|---|---|
| MySQL / MariaDB | Uses GORM migrators and database DDL for columns, indexes, primary/unique/foreign/check constraints, and comments. DDL can commit implicitly, so earlier operations may survive a later failure |
| PostgreSQL | Uses its migrator and supports table/column comments. Complex type conversions remain subject to database casting rules |
| SQLite | Rebuilds changed tables inside a transaction, copies retained columns, restores indexes and AUTOINCREMENT high-water marks, and checks foreign keys. Copy/constraint failures roll back. SQLite does not enforce string lengths or store comments |

MySQL/MariaDB indexes automatically created to support foreign keys are part of the constraint structure; omitting an `index` tag does not remove a required supporting index. When replacing supporting indexes or changing dependency columns, synchronization removes affected foreign keys, updates columns and indexes, and restores the constraints.

Target foreign-key dependencies determine table creation order. A failure identifies the operation and object. After resolving data conflicts, run `check` again and synchronize the remaining changes. Stored procedures, views, and triggers are maintained through native migrations rather than GORM models.

Startup performs an inexpensive, read-only check for required target tables and columns. Missing structure produces an update message and waits without repeatedly exiting or running automatic DDL. Run maintenance from one instance.
