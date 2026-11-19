# 数据库结构维护

Dawn提供两种可混用的方式：PrismGo迁移记录按顺序执行的变更；`dawn:schema`将实际数据库对齐到当前完整结构。迁移通常已完成升级，完整同步可继续处理遗漏的列、类型、长度、默认值或已删除的结构。

所有命令在应用根目录执行。能力由dawn-server提供，dawn-simple无需复制命令实现。

## 预览与同步

```sh
# 执行尚未运行的迁移
go run . migrate

# 可选：只读检查并预览全部差异
go run . dawn:schema check

# 检查并展示差异，执行同步，成功后自动复查
go run . dawn:schema sync

# 可选：单独再次验证最终结构
go run . dawn:schema check
```

预览列表显示计划中的变更，不是SQL报错或已完成的操作；“注意”表示变更风险，不代表执行失败。

`check`每项操作单独一段，空行分隔，列出对象、操作、风险等级、当前结构和目标结构，空结构显示`-`。删表、删列、改列、约束和唯一性变更标记“注意”；支持ANSI时仅该项首行标红，“当前结构”标题为绿色，“目标结构”标题为黄色，结构内容保持终端默认颜色。检查发现差异仍表示检查成功，默认返回0并显示差异数量；`--strict`用于CI发现差异时返回非零状态。检查不会执行DDL。

`sync`自身完成“检查并展示差异→执行同步→成功后复查”，不要求先运行`check`，也不读取上次检查的结果。先单独运行`check`便于人工审阅；直接运行`sync`也会检查并立即执行，不等待交互确认。无需逐项确认或额外的风险许可参数；生产环境沿用`--force`：

```sh
go run . migrate --force
go run . dawn:schema sync --force
```

| 选项 | 用途 |
|---|---|
| `--database=mysql` | 使用`database.connections.mysql`连接，默认使用应用默认连接 |
| `--tables=orders,users` | 本次只比较或同步这些表，不改变完整定义 |
| `--json` | 输出`action`、`changes`、`in_sync`，每项包含`risk`布尔值 |
| `--strict` | `check`发现差异时返回非零状态，默认检查完成返回0 |
| `--force` | 允许生产环境执行同步，不会关闭数据库约束或清空业务数据 |
| `--ansi` / `--no-ansi` | 强制启用或关闭终端颜色 |

## 一份完整定义

内置表统一定义在dawn-server的`database/migrations/install.go`，首次迁移和完整同步使用同一份结构。dawn-simple只保留一个安装迁移入口。

开发者通过`dawn.ServiceProvider.Schema`配置最终目标。使用普通GORM结构描述字段、索引和约束，不需要另写一份字段配置。推荐将业务定义放在`database/schema/schema.go`：

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

// 嵌入内置快照，新增字段，也可用同名字段覆盖原定义
// 修改默认字段时，同时修改应用实际使用的模型与接口
type User struct {
    migrations.User
    Nickname string `gorm:"size:200;not null;default:''"`
    Level    int    `gorm:"not null;default:1"`
}

func Configure(d *sdk.Definition) {
    d.Tables["orders"] = sdk.Table{Model: &Order{}, Comment: "订单表"}
    d.Tables["users"] = sdk.Table{Model: &User{}, Comment: "系统:普通用户表"}
}
```

在应用的`bootstrap/extensions.go`中，将回调交给现有扩展实例：

```go
import appschema "your/module/database/schema"

// 保留原来的Configure和其他扩展
dawn.ServiceProvider{Configure: admin.Configure, Schema: appschema.Configure}
```

`Tables`的键就是实际表名，可以追加、替换或`delete(d.Tables, "old_table")`。允许修改内置表，没有“只能修改Dawn表”的范围限制。需要同步的所有业务表都放进这份最终定义；**数据库中存在、最终定义中不存在的表会被删除，表中多出的字段、索引和约束也会被移除**。首次加入完整同步时，先补全现有业务结构，再查看`check`结果。

`Keep`默认保留PrismGo的`migrations`、SQLite内部`sqlite_*`及知识库运行时生成的`knowledge_vectors_*`。这些结构分别由迁移器、数据库和知识索引管理，不属于静态安装结构。开发者可按实际需要调整`d.Keep`中的表名匹配规则；有明确目标定义的表仍会比较，不因匹配`Keep`而跳过。

只修改Go代码、未更新这份目标定义的数据库操作，无法被结构同步推断。默认字段的结构更改也不会自动改写框架运行时模型，调用方需要同步适配。

## 与迁移配合

| 需求 | 选择 |
|---|---|
| 按顺序改名、搬数据、分拆表、回填后收紧约束 | 新增PrismGo迁移 |
| 修改长度、类型、默认值、注释，增删表列或索引 | 更新完整定义，再执行`check`、`sync` |
| 已执行迁移里的结构快照后来有调整 | `migrate`不会重放旧文件；通过完整定义同步已有库 |
| 回退已发布变更 | 原生迁移回滚；完整同步只有向当前目标收敛，没有历史回滚 |

两种方式可以自由组合。原生迁移不要求使用完整同步；业务迁移也能复用定义的`Create(db)`、`Drop(db)`，避免重复声明结构。`sync`不登记或删除迁移执行记录，不填充或覆盖账号、配置、用户编号密钥；初始化数据仍由Seeder维护。

改名不能仅在完整定义中更换名称，否则会按删除旧对象、创建新对象处理。要保留原数据，先在原生迁移中明确重命名，再更新最终定义。数据类型转换、非空和唯一约束由数据库实际验证，不会静默改写冲突数据。

`AutoMigrate`属于GORM第三方库，不是Dawn自己维护的同步引擎；它不会删除多余结构，不能替代完整同步。PrismGo的`schema.Bind(db).SyncModels(...)`只补建模型表和缺失列，同样不替代这套命令，也没有删除第三方能力。

## 数据库差异与运行

| 数据库 | 结构变更行为 |
|---|---|
| MySQL / MariaDB | 使用GORM迁移器和数据库DDL更新表列、索引、主键、唯一、外键、检查约束与注释；DDL可能隐式提交，前面已成功的操作不能靠事务全部撤销 |
| PostgreSQL | 使用对应迁移器更新结构，支持字段和表注释；复杂类型转换仍受数据库转换规则限制 |
| SQLite | 不支持原生改列，改列或约束时按最终定义重建表，在事务内复制保留字段、恢复索引与自增序列；复制或约束检查失败则回滚，不存储字段长度限制和注释 |

MySQL/MariaDB为外键自动建立的必需支撑索引属于约束结构，不会因模型未声明`index`就删除。替换外键支撑索引或修改其依赖列时，先移除受影响外键，完成字段和索引变更后恢复约束。

新建与修改表遵循目标中的外键依赖。同步失败会指出失败对象；MySQL等数据库可能已完成部分DDL，处理数据冲突后重新执行`check`，再同步剩余差异。存储过程、视图和触发器不由GORM模型声明，应由原生迁移维护。

启动检查只读检查目标中必需的表和列。缺失时显示更新提示并等待，不退出反复重启，也不会自动执行DDL；修改已执行迁移后可显式运行上述命令。执行维护时由一个实例负责即可。
