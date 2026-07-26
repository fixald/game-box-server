# server API 开发约束

所有接口必须先在 router 分组中注册，再实现 handler。路由、权限中间件、Swagger 注释和前端 API 模块必须一起提交。

## 路径命名空间

- `/api/v1/admin/*`：后台管理和运营业务。
- `/api/v1/client/*`：客户端用户和业务接口。
- `/api/v1/public/*`：明确的公开资源。
- `/api/v1/system/*`：健康检查、指标和运维接口。
- 当前 `/api/v1/login`、`/api/v1/menu`、`/api/v1/sys-user` 等 go-admin 系统接口为兼容区；不得新增无命名空间业务接口。

## 路由规则

- 资源使用小写 kebab-case 复数，ID 使用 `/:id`。
- 列表和详情使用 GET，创建 POST，更新 PUT，删除 DELETE；改变状态使用 POST 子动作。
- 查询参数使用 `page`、`pageSize`、`keyword`；不要新增 `pageIndex`。
- 认证和权限必须挂在 router group middleware，不能只在 handler 中判断。
- 每个新接口添加 Swagger `@Router`、权限声明和测试；确认 API 记录写入权限表。
- 客户端埋点统一使用 `POST /api/v1/client/events`，禁止为单个页面新增专用埋点路径。

跨工程完整约定见 `../API_ROUTE_RULES.md`。AI 修改路由前必须用 `rg` 同时核对 router、Swagger 和三个前端工程引用。
