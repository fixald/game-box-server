# AI 开发入口

本目录由 backend-agent 负责。开始前必须阅读仓库根目录 `AGENTS.md`、`DEVELOPMENT_PLAN.md`、本目录 `AI_PLAN.md`、`API_RULES.md` 和 `../API_ROUTE_RULES.md`。

负责 Go API、业务逻辑、模型、数据库、缓存、权限、审计、定时任务和服务端测试。修改路由或 handler 前，必须同步 router、权限、错误码、Swagger/接口文档及受影响的前端计划；不得修改 `client/` 或 `admin-ui/` 的业务实现。

优先验证：`go test ./...`，并根据 `Makefile` 使用项目已有检查命令。
