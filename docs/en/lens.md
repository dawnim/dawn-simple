# Developer tools

PrismGo Lens is a standalone developer tool for application information, documentation, routes, commands, and logs.

```sh
git clone https://github.com/prismgo/lens.git /tmp/admin-lens
cd /tmp/admin-lens
go build -o /tmp/admin-lens-cli ./cmd/prismgo-lens
cd /path/to/your/project
/tmp/admin-lens-cli --project . execute-tool application-info e30=
```

`e30=` represents an empty JSON object. Configure the stdio MCP command as:

```sh
/tmp/admin-lens-cli --project /absolute/project/path mcp
```

Keep the CLI at a stable path for regular use. Project conventions live in each repository's `AGENTS.md`. Connect database tools to development or dedicated test databases, and redact log output.
