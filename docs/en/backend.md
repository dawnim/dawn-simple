# Backend

Put admin business code in `app/admin` and shared entities in `app/models`. APIs reuse the application's SQL, Redis, authorization, and response conventions.

## Generate business code

Start ordinary CRUD development with the code generator, then add business logic to its output. Create the business table through a migration or schema sync first. Under Developer tools → Code generator, select a table, mode, resource name, and business directory:

| Mode | Output |
|---|---|
| Reference | Model, generic CRUD registration, frontend API, field translations, and a `CrudPage` entry |
| Source | Model, standalone controller, frontend API, field translations, and source for the list, filters, and edit dialog |

Both modes use the same permissions, data scopes, and list interactions. The frontend directory defaults to `business` and is configurable. Chinese field labels come from column comments; English labels derive from field names.

Edit individual files before downloading or generating them. Select Generate table migration to include a schema snapshot. Direct project writes are available only to local-environment super administrators and preserve existing files. Follow the generated integration instructions to register routes, menus, and pages, then check field types, validation, and business permissions.

### CLI and AI integration

`dawn-server` provides the command. Run it from an application that has registered the extension; it shares the admin UI generator and templates:

```sh
# List business tables as JSON
go run . dawn:generate --json

# Preview full source; files maps relative paths to source code
go run . dawn:generate orders --mode=full --preview --json

# Generate into the current project; reference mode is the default
go run . dawn:generate orders

# Choose output, resource, frontend subdirectory and migration snapshot
go run . dawn:generate orders --output=/tmp/orders-module --resource=orders --directory=sales --migration

# Extension development with a different Go module and frontend root
go run . dawn:generate orders --output=../orders-extension --module=example.com/orders-extension --backend=app --frontend=src --mode=full
```

| Argument/option | Default and meaning |
|---|---|
| `table` | Omit to list business tables; no naming-prefix restriction; built-in and migration metadata tables are excluded |
| `--mode` | `simple` for shared CRUD, `full` for editable source |
| `--output` | Application root; accepts an absolute path or a path relative to the working directory |
| `--backend`, `--frontend` | `app` and `admin/src` within the output root; model imports follow the backend directory |
| `--directory`, `--resource` | `business` frontend subdirectory; resource defaults to the hyphenated table name |
| `--module` | `dawn.module`; specify the target module when generating into another Go module |
| `--database` | Optional configured database connection name |
| `--migration` | Include an independent snapshot migration of the current table |
| `--preview`, `--json` | Preview without writing; machine-readable JSON contains `output`, `preview`, and `files` |

Generation only reads database metadata; it does not create tables or write business data. Existing source files are preserved, and a failed write removes files created by that call. The target can be separate from the running application; generation does not create its `go.mod` or `package.json`. Merge routes, menus with icons, permissions, and frontend `pages` into the existing registration points as shown in the integration examples.

Both modes generate field-presence requirements, required-string and length validation, nullable types, defaults, and unique indexes. The server assigns `created_by` and `department_id` and applies the corresponding data scope. SQL types cannot identify business enums, relations, or uploads; adjust these after generation. For standard status values, add `dict: 'data_status'` to the frontend column and the appropriate enum validation to the model.

## Customize generated CRUD

The following example explains the key integration points in generated code, using `app/models/record.go`:

```go
package models

import base "github.com/dawnim/dawn-server/app/models"

type Record struct {
	base.Base
	Name         string `json:"name" gorm:"not null;size:100;comment:名称" binding:"required"`
	CreatedBy    uint   `json:"created_by" gorm:"not null;index;comment:创建管理员ID"`
	DepartmentID uint   `json:"department_id" gorm:"not null;default:0;index;comment:归属部门ID"`
}
```

`Base` supplies the primary key, millisecond timestamps, and soft deletion. Create a migration for the business table, then register its resource in `app/admin/routes.go`:

```go
package admin

import (
	sdk "github.com/dawnim/dawn-server/app/admin"
	crud "github.com/dawnim/dawn-server/app/admin/controllers"
	"github.com/dawnim/dawn-simple/app/models"
	"github.com/prismgo/framework/route"
	"gorm.io/gorm"
)

func RegisterRoutes(a *sdk.App, g *route.Registrar) {
	crud.RegisterResource(a, g, crud.Resource[models.Record]{
		Path: "/records", Permission: "app.records",
		Fields:   []string{"name", "created_by", "department_id"},
		Required: []string{"name"}, Search: []string{"name"}, Importable: true,
		Scope: func(db *gorm.DB, actor *sdk.Actor) *gorm.DB {
			return db.Scopes(actor.Scope("created_by", "department_id"))
		},
		Prepare: func(row *models.Record, actor *sdk.Actor, editing bool) {
			if !editing {
				row.CreatedBy = actor.User.ID
				row.DepartmentID = actor.User.DepartmentID
			}
		},
	})
}
```

`Fields` declares writable fields. `Scope` limits reads and writes. `Prepare` assigns ownership on the server. Use `Validate` for business checks, `WriteScope` for additional write restrictions, and `AfterWrite` for work after a successful commit.

For custom business handling, register standard routes through the same entry point. Handlers use the `func(*gin.Context)` signature:

```go
crud.RegisterResourceRoutes(a, g, "/records", "app.records", crud.ResourceHandlers{
	Index: controller.Index, Store: controller.Store, Show: controller.Show,
	Update: controller.Update, Destroy: controller.Destroy,
})
```

`Index/Show` require `.list`, `Store` requires `.create`, `Update` accepts PUT/PATCH and requires `.update`, and `Destroy` registers both member and collection deletion with `.delete`. The delete handler reads explicit IDs through `response.IDs(c)` and performs the batch in one transaction. Omitted handlers create no routes and need no placeholder methods.

`RegisterResource` and source-mode generated controllers share this registration. Handlers own field validation, data scopes, transactions, and responses. When resources share a permission prefix, set `Resource.Name` or `ResourceHandlers.Name` to give each resource distinct route names. Register operations such as login, reception, and AI assistance separately with `g.Get/Post/Put/...`. Reuse `RegisterResourceFiles` for CSV representations, imports, and templates when needed; they are not automatically enabled for every resource.

CSV uploads are previewed before confirmation. `POST collection/imports?preview=true` checks CSV syntax and headers without writing, returning `{data:{columns:[{key,label}],rows:a_two_dimensional_array}}`. Confirmation resubmits the same file to `POST collection/imports`. Database unique indexes identify existing records. Only CSV columns are overwritten; primary keys, creation timestamps, and ownership stay unchanged. GORM automatically refreshes `updated_at` on updates. Later rows override earlier duplicates within the file. Updates require `.update` permission and the same write scope. Validation, authorization, or write errors roll back the entire import. Success returns `{data:{count:inserted_rows,updated:updated_rows}}`, counting repeated updates per CSV row. Templates contain only a header; add data rows before uploading. Uploading an unchanged template returns zero for both counts.

In the existing `Configure`, assign `o.Routes = RegisterRoutes` and append menus and permissions:

```go
o.Menus = append(o.Menus, sdk.Menu{
	Key: "app.records", ParentKey: "dawn.development", Title: "业务记录",
	Path: "/records", Component: "/records", Permission: "app.records.list", Icon: "ri:table-line",
})
for action, title := range map[string]string{
	"list": "查看", "create": "新增", "update": "修改", "delete": "删除", "export": "导出",
} {
	o.Permissions = append(o.Permissions, sdk.Permission{
		Key: "app.records." + action, Title: title, ParentKey: "app.records",
	})
}
```

Grant the role permissions, then register `/records` as described in [Business lists](frontend.md#business-lists). Replace the example module import when renaming your project.

### Menus and languages

Define feature menus and operation permissions in `Configure` and ship them with the code. Stable keys identify grants and relationships; `Title` uses the original Chinese text. SQL stores role grants and custom menus.

Register English translations in `Boot` in `app/providers/app_service_provider.go`:

```go
import "github.com/prismgo/framework/translation"

translation.AddLines(map[string]any{
	"messages.业务记录": "Records",
}, "en", "admin")
```

The server translates menu and permission labels using the request's `Accept-Language`; Chinese falls back to the original text. Frontend language changes update navigation, tabs, and menu management without reloading. Register page text through `createAdmin.messages`; see [Themes and languages](frontend.md#themes-and-languages).

## API conventions

Resources live directly under `/admin`. Use plural paths and hyphens; nesting expresses actual ownership.

| Operation | Request |
|---|---|
| List, detail | `GET /admin/records`, `GET /admin/records/{id}` |
| Create | `POST /admin/records` |
| Partial update | `PATCH /admin/records/{id}` |
| Full replacement | `PUT /admin/records/{id}` |
| Delete, batch delete | `DELETE /admin/records/{id}`, `DELETE /admin/records?ids=1,2` |
| Import, template | `POST /admin/records/imports` with a `file` field; `GET /admin/records/import-template` |
| Export | `GET /admin/records` with `Accept: text/csv` |

List parameters are `page`, `page_size`, `sort_by`, and `sort_order`, defaulting to page 1 and 10 rows. Sorting uses an allowlist; page size follows the request. Exports retain filters and scopes but ignore pagination. Batch deletion requires explicit IDs and runs in one transaction.

| Response | Format |
|---|---|
| Paginated list | `{data:[],meta:{page,page_size,total}}` |
| Tree | `{data:[]}` |
| Detail, create, update | `{data:...}` |
| Error | `{message,errors?}` with the appropriate HTTP status |

Creation returns `201` and `Location`; deletion returns `204`; validation errors return `422`. JSON fields use `snake_case`, and time uses Unix milliseconds. The shared frontend `request` layer unwraps responses.

PATCH changes only submitted fields. PUT requires all mandatory writable fields and resets omitted optional fields to defaults or zero values. `0`, `false`, and empty strings are saved normally; `null` is allowed only for nullable fields.

Standard resource route names default to `<permission-prefix>.index/store/show/update/destroy/destroy-many`. When imports are enabled, `.imports` and `.import-template` are named as well. Inspect them with `route:list` and generate URLs with `route.URL`. Use standalone controllers for complex operations and check permissions and [data scopes](authorization.md) at their entry points.

## Database migrations

Place ordered migrations in `database/migrations` with unique timestamp filenames and blank-import the package during startup. PrismGo records execution by migration name; editing an executed file does not replay it.

Developers may also maintain one complete target and run `dawn:schema check` / `dawn:schema sync`. Both workflows can be combined: run pending migrations, synchronize remaining differences, then check the result. Business tables and overrides of built-in fields participate in the complete target. See [Database schema maintenance](schema.md) for registration, red risk previews, deletion behavior, and database differences.

Define registration functions in the migration file itself so the framework can identify its filename. Built-in schema example:

```go
package migrations

import (
	sdk "github.com/dawnim/dawn-server/database/migrations"
	"github.com/prismgo/framework/database"
	"gorm.io/gorm"
)

func init() {
	database.RegisterMigration(
		func(db *gorm.DB) error { return sdk.Up(db) },
		func(db *gorm.DB) error { return sdk.Down(db) },
	)
}
```

The dawn-server installation migration includes system tables, chat tables, and indexes. dawn-simple integrates them through `sdk.Up` and `sdk.Down`. Add business seed data in `database/seeders/database_seeder.go`; manage schema and data separately.

```sh
go run . migrate
go run . db:seed
go run . migrate:status
go run . migrate:rollback
```

Rollback reverses the latest batch and may drop tables. In production, use `--force` for migration, seeding, or installation, run it from one instance, and retain matching migration and seeder sources.

## Business cache

Use the framework definition type in `app/constants/key.go`:

```go
package constants

import (
	"time"
	dawn "github.com/dawnim/dawn-server/constants"
)

var TempOrderSummary = dawn.Definition{
	Name: "TempOrderSummary", Pattern: "temp:order:summary:%d", Type: "string",
	Description: "订单摘要缓存", TTL: 5 * time.Minute, Readable: true, Clearable: true,
}
```

In `Configure`, append `o.CacheKeys = append(o.CacheKeys, appconst.TempOrderSummary)` with `appconst` importing your business constants. Generate application- and environment-prefixed keys with `a.Key(appconst.TempOrderSummary, orderID)`, then read and write through `a.Redis`.

Use `Lock`, `Forever`, and `Temp` prefixes for locks, permanent cache, and temporary cache. Business code defines its own keys and directly reuses framework definitions. See [Cache management](configuration.md#cache-management).

## Queues and schedules

Use framework `QueueDefault` and `QueueParallel` for default queues. Define additional business queue names in `app/constants/queue.go` and append them to the appropriate `queue` list in `config/horizon.go`:

| Group | Purpose | Workers |
|---|---|---|
| `supervisor` | `default` and serial tasks | Exactly 1 |
| `multiples` | `parallel` and concurrent tasks | Autoscaled, maximum 4 by default |

Put jobs in `app/jobs` and schedules in `app/schedule/register.go`. Register the same job types in producers and workers. Jobs must tolerate retries; dispatch business events after transaction commits.

Run only one serial group and one cron instance per deployment. On additional Horizon nodes, set `HORIZON_SERIAL=false`; adjust the concurrency limit with `HORIZON_CONCURRENCY`.

## Reusing attachment uploads

After authenticating the user and checking allowed file types, a business endpoint can reuse the file center and its deduplication. Here `file` and `header` come from a multipart upload and `a` is the admin application:

```go
uploader := &controllers.Controller{App: a}
attachment, err := uploader.SaveUpload(c, file, controllers.UploadInput{
    Name: header.Filename, Size: header.Size, Scope: fmt.Sprintf("user:%d", user.ID), OwnerID: user.ID,
}, limit)
```

`limit` is in bytes. `SaveUpload` computes and verifies digests and follows the admin upload settings. Identical content shares storage while attachments retain independent ownership in the unified file center. Set `Scope` and `OwnerID` from the authenticated user, never client input. The caller controls authentication, request limits, allowed types, and read authorization.

For instant uploads, let the browser calculate MD5, SHA-256, and byte size, then call this from a business preflight endpoint:

```go
input.Scope, input.OwnerID = fmt.Sprintf("user:%d", user.ID), user.ID
attachment, exists, err := uploader.CheckUpload(c, input)
```

On a hit, return the attachment ID without transferring content. Otherwise call `SaveUpload` with `UploadInput.MD5/SHA256` for server verification. Preflight enforces upload policies and checks the file in its recorded storage location, including after a driver change. Authorize reads before calling `uploader.ServeAttachment(c, attachment)` with `File` preloaded. See the [complete community upload example](https://github.com/dawnim/dawn-simple/blob/dev/app/web/moments.go).
