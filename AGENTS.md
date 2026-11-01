# 示例应用开发规范

## 目录

main.go与bootstrap负责入口和装配；后台菜单、路由、控制器、中间件放app/admin。共享模型放app/models，Key与队列名称放app/constants。业务前端在admin/src，默认仪表盘通过pages映射覆盖。按实际需要添加事件、Job等业务目录，不添加无关功能演示。

## 编码

- 优先标准库和已有运行时能力。直接实现简单逻辑；确有多处复用再提取函数。业务层按需要增加，不为形式创建Service/Repository。
- Go文件小写加下划线，一个实体一个文件；控制器按功能命名。共用字段放base.go，固定选项贴近实体。
- Register只绑定懒加载factory，Boot注册驱动、命令和监听器。连接由容器统一创建和释放，业务入口不另建一套资源。
- 时间统一int64 Unix毫秒，软删除0表示有效；建表显式使用整数时间列。结构与调用方同步修改，不写历史兼容代码。
- 管理接口默认/admin/。服务端检查权限与数据Scope；列表、详情、写入、导出、下载使用一致的数据范围，归属字段由服务端决定。
- dawn.*、app.*、custom.*是稳定的系统、业务、外链授权key；保存后不改名、不复用。
- 缓存集中定义，锁以Lock开头，永久缓存以Forever开头，临时缓存以Temp开头；声明TTL及可读可清策略。
- Event在事务提交后派发；Job注册类型、传递context并保证重试可接受。共享状态存数据库/Redis，goroutine有明确退出路径。
- 配置通过Add/Env注册，保留组件的命名空间、字段和层级；SQL与Redis连接放database，其他组件引用连接名。HTTP监听、代理和超时统一使用app.server，不在后台路由中重复覆盖。环境变量可覆盖默认值。密码、token和密钥不写日志；保密业务参数加密保存。
- HTTP跨域默认支持*，Bearer请求不携带跨域Cookie；凭证模式使用明确来源列表。WS按同一来源配置校验，握手后仍需认证。
- 修改Go文件后运行gofmt/goimports；公开扩展契约有简短说明，测试验证实际行为和权限边界。

## 变更与验证

保留其他任务的修改，确认当前Git根目录后再操作。依赖问题改源码或版本配置，不修改module cache和node_modules。

外部数据库测试只使用明确指定的专用空库。没有连接变量时跳过外部矩阵测试。只清理本次创建并确认的进程，不重启既有服务。

运行与改动相符的测试，分别报告代码检查、服务联调和界面验收。提交、推送、发布和部署按用户指令执行。

## 队列

- 固定使用config/horizon.go中的supervisor和multiples两组。
- 串行任务加入supervisor.queue，固定一个worker；并发任务加入multiples.queue，通过HORIZON_CONCURRENCY设置上限。
- 新任务复用上述两组，不按任务新建supervisor；队列名称统一定义在app/constants/queue.go。
- 同一部署只运行一份串行组。扩容Horizon节点设置HORIZON_SERIAL=false；跨实例的资源互斥使用业务锁。
- cron同一部署只启动一个实例。组合模式已有Horizon和cron，不重复启动独立进程。

## 运行与检查

```sh
go run . dawn:start all
go test -race ./...
go vet ./...
```

组合运行和独立进程使用对应Supervisor配置。前端在admin目录执行pnpm build；修改前端库后更新本地依赖再构建。

测试使用临时SQLite与隔离Redis，不能读取开发者.env后写入固定测试数据。保持401/403、数据Scope、CORS、事件、队列重试、下载归属、限流与组合退出的验证。

README说明配置和运行，docs说明结构和业务用法；新增示例同步入口说明。文档直接写操作、规则及设计收益，来源留在相关项目README鸣谢。

## 服务管理

启动使用 `dawn:start all`，本地前后端使用 `dawn:start local`。服务列表用逗号分隔，操作为start、stop、restart、reload、status；stop保持停止，reload仅用于HTTP/WS。每个服务独立进程，管理器只恢复期望运行的服务，不重复接管Horizon内部worker。

复用框架服务生命周期与监听器继承，重载必须等待新进程就绪；失败时保留旧进程。禁止提前取消在途HTTP请求或把WS连接直接迁移到新进程。验证异常拉起、主动停止、交接失败、端口占用、父进程退出和子孙进程回收。

配置直接使用标准字段和类型，不重复验证静态配置。应用日志默认storage/logs/dawn.log；新增配置同步.env.example，不删基础配置键。

## 语言与文案

固定文案的语言key使用中文原文，其他语言包使用相同key。动态字段按技术标识读取显示名，保留按模块组织的zh/en字段映射；接口字段名不翻译。缺少译文时使用key并渲染占位符。固定界面、错误和通知文案必须使用语言包；修改中文措辞时同步修改调用处、key和所有译文，不保留旧key别名。中文尽量不加中间空格，句末不加标点；命令、路径、代码保留必要空格。

动态内容通过占位符传入，禁止拼接后翻译。前端使用命名占位符，例如`t('您于{v0}执行了{v1}', { v0: time, v1: action })`；后端保留格式占位符并按请求locale翻译。各语言占位符名称、数量和类型保持一致，翻译可调整命名参数顺序。业务语言包按模块组织，生成页面自动加载本模块语言包，通用文案使用公共语言包。

## 命名

Go导出标识使用PascalCase，缩写保留ID、IP、HTTP、URL等标准大写；Go文件使用snake_case。前端变量和函数使用camelCase，组件名使用PascalCase。数据库列、HTTP/WS请求与响应字段统一使用小写snake_case，例如id、user_id、created_at；界面中的ID、IP、PID仍使用标准大写。第三方结构在接口边界适配，不向调用方暴露Go字段名，也不保留大小写两套字段或兼容别名；框架原生配置与公开组件参数遵循各自契约。

字段语言包固定使用fields.<模块>.<接口字段名>，zh/en使用相同层级和相同snake_case字段，与接口逐字一致。禁止同时定义ID/id、Host/host、JobName/job_name等同义映射；新增或改名时同步接口、页面、生成模板及语言包。菜单和权限的稳定key不因风格调整改名。
