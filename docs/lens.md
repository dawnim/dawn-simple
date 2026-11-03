# 开发工具

PrismGo Lens 提供应用信息、文档、路由、命令与日志查询，作为独立开发工具运行。

```sh
git clone https://github.com/prismgo/lens.git /tmp/admin-lens
cd /tmp/admin-lens
go build -o /tmp/admin-lens-cli ./cmd/prismgo-lens
cd /你的项目目录
/tmp/admin-lens-cli --project . execute-tool application-info e30=
```

`e30=` 表示空 JSON 对象。stdio MCP 命令配置为：

```sh
/tmp/admin-lens-cli --project /绝对项目目录 mcp
```

长期使用时将 CLI 放在固定路径。项目规范保留在各仓库的 `AGENTS.md`，数据库工具连接开发或专用测试库，日志输出脱敏。
