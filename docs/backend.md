# 后端开发

后台业务集中在 `app/admin`，共享实体放在 `app/models`。接口复用应用的数据库、Redis、授权与响应格式。

## 生成业务代码

在“开发工具 → 代码生成”选择数据表，配置字段、搜索项和业务目录：

| 模式 | 输出 |
|---|---|
| 引用模式 | 模型、通用 CRUD 注册、前端 API、字段语言包及 `CrudPage` 入口 |
| 源码模式 | 模型、独立控制器、前端 API、字段语言包、列表、搜索栏和编辑弹窗源码 |

两种模式使用相同的权限、数据范围和列表交互。前端目录默认 `business`，可自定义；中文字段名读取表注释，英文按字段名生成。

下载或生成到项目时可逐文件编辑，并勾选“生成表迁移”输出结构快照。直接写入项目仅对本地环境的超级管理员开放，已有文件保持原样。生成后按页面中的接入说明注册路由、菜单和前端页面，再检查字段类型、校验与业务权限。

## 手写 CRUD

以业务记录为例，创建 `app/models/record.go`：

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

`Base` 提供主键、毫秒时间和软删除。为业务表创建迁移后，在 `app/admin/routes.go` 注册资源：

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

`Fields` 声明可写字段，`Scope` 约束查询与写入范围，`Prepare` 设置服务端归属。`Validate` 可补充业务校验，`WriteScope` 可进一步限制写入，`AfterWrite` 在成功提交后执行。

在现有 `Configure` 中设置 `o.Routes = RegisterRoutes`，并追加菜单与权限：

```go
o.Menus = append(o.Menus, sdk.Menu{
	Key: "app.records", ParentKey: "dawn.development", Title: "业务记录",
	Path: "/records", Component: "/records", Permission: "app.records.list",
})
for action, title := range map[string]string{
	"list": "查看", "create": "新增", "update": "修改", "delete": "删除", "export": "导出",
} {
	o.Permissions = append(o.Permissions, sdk.Permission{
		Key: "app.records." + action, Title: title, ParentKey: "app.records",
	})
}
```

给角色授权后，按[前端开发](frontend.md#业务列表)注册 `/records` 页面。项目改名后，将示例中的模块导入替换为自己的 `go.mod` 模块名。

### 菜单与多语言

功能菜单和操作权限在 `Configure` 中定义，随代码发布，无需向线上数据库补录菜单。稳定 key 用于授权与关联，`Title` 使用中文原文；角色授权和自定义菜单保存在数据库。

在 `app/providers/app_service_provider.go` 的 `Boot` 中注册英文：

```go
import "github.com/prismgo/framework/translation"

translation.AddLines(map[string]any{
	"messages.业务记录": "Records",
}, "en", "admin")
```

服务端按请求的 `Accept-Language` 翻译菜单与权限名称，中文直接返回原文。前端切换语言时更新导航、标签页和菜单管理列表，无需刷新页面。页面自身文案通过 `createAdmin.messages` 注册，见[主题与语言](frontend.md#主题与语言)。

## 接口约定

资源直接位于 `/admin` 下，路径使用复数和连字符，层级表达真实从属关系。

| 操作 | 请求 |
|---|---|
| 列表、详情 | `GET /admin/records`、`GET /admin/records/{id}` |
| 新增 | `POST /admin/records` |
| 局部修改 | `PATCH /admin/records/{id}` |
| 完整 | `PUT /admin/records/{id}` |
| 删除、批量删除 | `DELETE /admin/records/{id}`、`DELETE /admin/records?ids=1,2` |
| 导入、模板 | `POST /admin/records/imports`，文件字段为 `file`；`GET /admin/records/import-template` |
| 导出 | `GET /admin/records`，请求头 `Accept: text/csv` |

列表参数为 `page`、`page_size`、`sort_by`、`sort_order`，默认页码 1、每页 20 条。排序使用字段白名单，每页条数按请求值处理。导出沿用筛选与数据范围，忽略分页；批量删除要求明确 ID，并在同一事务中完成。

| 响应 | 格式 |
|---|---|
| 分页列表 | `{data:[],meta:{page,page_size,total}}` |
| 树列表 | `{data:[]}` |
| 详情、新增、修改 | `{data:...}` |
| 错误 | `{message,errors?}`，使用对应 HTTP 状态 |

新增返回 `201` 和 `Location`，删除返回 `204`，参数校验失败返回 `422`。JSON 字段使用 `snake_case`，时间使用 Unix 毫秒。前端通过公共 `request` 解包响应。

局部修改（PATCH）只更新提交字段，未提交字段保持原值；完整修改（PUT）提交全部可写内容，必填字段须提交，可选字段省略时恢复默认值或零值。`0`、`false` 和空字符串均正常保存，`null` 仅用于可空字段。

通用 CRUD 的命名路由为 `<权限前缀>.index/store/show/update/destroy/destroy-many`。可用 `route:list` 查询，通过 `route.URL` 生成地址。复杂业务使用独立控制器，并在入口校验权限与[数据范围](authorization.md)。

## 数据库迁移

迁移放在 `database/migrations`，使用唯一时间戳文件名，并在启动包中空导入。结构快照固定保存字段、索引和注释；后续变更新增迁移。

注册时在当前文件中定义函数，框架据此识别迁移文件名。内置表迁移示例：

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

dawn-server 的安装迁移包含系统表、聊天表与索引，dawn-simple 通过统一的 `sdk.Up`、`sdk.Down` 接入。业务基础数据在 `database/seeders/database_seeder.go` 中追加；结构与数据分别管理。

```sh
go run . migrate
go run . db:seed
go run . migrate:status
go run . migrate:rollback
```

回滚撤销最近一批迁移，可能删除对应表。生产执行迁移、填充或安装时加 `--force`，由一个实例执行，并保留同版的迁移与 Seeder 源码。

## 业务缓存

在 `app/constants/key.go` 使用框架的定义类型：

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

在 `Configure` 中追加 `o.CacheKeys = append(o.CacheKeys, appconst.TempOrderSummary)`；`appconst` 导入业务常量包。通过 `a.Key(appconst.TempOrderSummary, orderID)` 生成带应用与环境前缀的 Key，再使用 `a.Redis` 读写。

锁、永久缓存、临时缓存分别以 `Lock`、`Forever`、`Temp` 命名。业务只定义自己的内容，框架定义直接引用。管理策略见[缓存配置](configuration.md#缓存管理)。

## 队列与调度

默认队列直接引用框架的 `QueueDefault`、`QueueParallel`。需要业务队列时，在 `app/constants/queue.go` 定义名称，并加入 `config/horizon.go` 对应组的 `queue` 列表：

| 组 | 用途 | worker |
|---|---|---|
| `supervisor` | `default` 与需要串行处理的任务 | 固定 1 个 |
| `multiples` | `parallel` 与可并发任务 | 自动伸缩，默认最多 4 个 |

业务任务放在 `app/jobs`，调度放在 `app/schedule/register.go`，使用队列与调度 API。任务注册需在生产者与 worker 中一致；任务设计应允许重试，业务事件在事务提交后派发。

同一部署保留一个串行组与一个 cron 实例。扩容 Horizon 节点设置 `HORIZON_SERIAL=false`，通过 `HORIZON_CONCURRENCY` 调整并发上限。
