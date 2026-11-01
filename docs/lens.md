# 开发工具

Lens用于查询应用信息、文档、路由、命令和日志，单独构建，不参与应用生产编译。

```sh
git clone https://github.com/prismgo/lens.git /tmp/admin-lens
cd /tmp/admin-lens
go build -o /tmp/admin-lens-cli ./cmd/prismgo-lens
cd /你的项目目录
/tmp/admin-lens-cli --project . execute-tool application-info e30=
```

e30=表示空JSON对象。Agent的stdio MCP命令可配置为 `/tmp/admin-lens-cli --project /绝对项目目录 mcp`。需要长期使用时，把CLI放到固定路径。

项目规则放在AGENTS.md。数据库查询使用开发或专用测试连接，日志输出保留脱敏。工具不会自动获得生产环境权限。
