# Dawn 开发文档

在线文档：[https://dawnim.github.io/dawn-simple/](https://dawnim.github.io/dawn-simple/)

从 [dawn-simple](https://github.com/dawnim/dawn-simple) 开始，按业务需要扩展模型、接口、菜单和页面。[文档首页](index.md) · [文档站构建](testing.md#文档站)

## 入门

| 文档 | 内容 |
|---|---|
| [快速开始](quickstart.md) | 环境准备、安装初始化与首次运行 |
| [创建空白项目](create-project.md) | 独立接入服务端与前端库，搭建业务骨架 |
| [架构与目录](architecture.md) | 框架关系、目录职责与应用生命周期 |

## 业务开发

| 文档 | 内容 |
|---|---|
| [后端开发](backend.md) | 模型、迁移、CRUD、菜单、缓存与队列 |
| [前端开发](frontend.md) | 页面注册、公共组件、工作台、主题与多语言 |
| [权限与数据范围](authorization.md) | 身份认证、角色授权、管理级别与记录归属 |
| [配置](configuration.md) | 数据库、Redis、文件存储、系统参数与日志 |
| [聊天与实时通信](chat.md) | 管理员聊天、消息投递与 WebSocket 扩展 |

## 运行与交付

| 文档 | 内容 |
|---|---|
| [服务运行](runtime.md) | 本地开发、终端接管、平滑重载与生产部署 |
| [测试与发布](testing.md) | 检查命令、数据库矩阵、框架联调与版本发布 |
| [开发工具](lens.md) | PrismGo Lens 接入 |

教程中的命令默认在 `dawn-simple` 根目录执行。框架维护命令单独标注所在仓库。路由、容器、事件和队列的基础用法见 [PrismGo 文档](prismgo/README.md)，源文件通过 Git 子模块引用官方仓库。
