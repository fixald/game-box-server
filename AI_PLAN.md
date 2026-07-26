# server AI Development Plan

## 工程职责

基于 go-admin 实现 Go 后端，目录建议：

```text
app/
  games/ servers/ gifts/ content/ downloads/ cauth/ stats/
  admin/ api/ models/ service/
```

运营账号使用 go-admin 的 `sys_user` + Casbin；玩家账号使用独立 `users`/`gb_users`，禁止混表。

## 执行规范

每次只实现一个功能点。实现前检查现有 go-admin 约定和迁移方式；优先复用 JWT、Casbin、菜单、操作日志、代码生成和定时任务。

## Phase 0：基础设施

- `S.CORE.01` 统一响应封装：`code/message/data/requestId`
- `S.CORE.02` 错误码、请求追踪、幂等键中间件
- `S.CORE.03` MySQL/Redis/OSS/CDN 环境配置
- `S.CORE.04` 结构化日志、健康检查、Prometheus 指标
- `A.OPS.01` 配置运营、审核员、管理员角色和菜单骨架

## Phase 1：MVP 任务

| ID | 任务 | 主要接口/表 | 验收 |
|---|---|---|---|
| `S.AUTH.01` | 手机验证码登录注册 | `/api/auth/sms/*`；`users`、`sms_codes` | 60 秒限流、过期/错误/锁定错误码正确 |
| `S.AUTH.02` | Refresh Token 与登出 | `/api/auth/refresh`、`/logout`；`refresh_tokens` | 轮换、重放失效、登出失效 |
| `S.GAMES.01` | 游戏列表筛选分页 | `GET /api/games`；`games` | 下架不可见、组合筛选正确 |
| `S.GAMES.03` | 游戏详情 | `GET /api/games/:id` | 字段完整、无权限返回 404/下架态 |
| `A.GAME.01` | 游戏后台 CRUD | `/api/admin/games` | 上下架、上传、启动配置修改均审计 |
| `S.SERVER.01` | 新服推荐列表和状态缓存 | `GET /api/v1/client/game-servers`；`servers` | 支持 `page,pageSize,recommended=true`，状态延迟 ≤60 秒 |
| `S.SERVER.04` | 状态变更和批量维护 | `server_status_logs` | 状态机合法、操作可追溯 |
| `A.SERVER.01` | 区服 CRUD/CSV 导入 | `/api/admin/servers` | 失败行回报、批量操作审计 |
| `S.DL.01` | 最新版本查询 | `/api/client/versions/latest` | 返回版本、最低版本、签名、SHA、短期 URL |
| `S.DL.02` | 文件发布流水线 | `game_versions`、`game_files` | 未验证包不可发布 |
| `S.DL.03` | 下载/启动事件 | `POST /api/downloads/events` | 幂等、request_id、错误码完整 |
| `S.CONTENT.04` | 最小举报流程 | `/api/reports`；`reports` | 可提交、后台处理、有记录 |
| `S.SUPPORT.01` | 反馈接口 | `POST /api/feedback`；`feedbacks` | 支持附件、分类和处理状态 |
| `S.AUTH.05` | 实名适配层 | `users` 实名字段 | 不存明文身份证，失败可追踪 |
| `A.OPS.02` | 业务审计 | `operation_audit_logs` | 发布、上下架、封禁、库存等 100% 可追溯 |

### `A.BANNER.01` Banner 管理

Banner 是运营后台资源，不能与游戏图标、版本文件或文章内容混用。Banner 只允许引用已上传并完成安全校验的附件，发布时必须记录操作审计。

数据表：`banners`

```text
id, title, imageUrl, linkType, linkValue, position, weight,
gameId, startAt, endAt, status, sort, createdBy, updatedBy,
createdAt, updatedAt, deletedAt
```

字段规则：

- `position` 为运营位标识，例如 `home_top`；同一运营位按 `weight DESC, sort ASC, id DESC` 排序。
- `linkType` 只允许 `none/game/article/url`；外链必须经过 HTTPS 和域名白名单校验。
- `startAt < endAt`，时间统一存 UTC；查询接口按服务端时间判断是否生效。
- `status` 为 `draft/published/offline`；只有已发布且处于时间窗内的 Banner 才能出现在 C 端。
- 同一运营位允许多个时间窗，但发布操作必须校验时间窗重叠并给出冲突记录。
- 删除使用软删除；发布、撤回、删除、时间窗和链接修改必须写入审计。

后台接口：

```text
GET    /api/admin/banners              banner:list
POST   /api/admin/banners              banner:create
GET    /api/admin/banners/:id          banner:list
PUT    /api/admin/banners/:id          banner:update
DELETE /api/admin/banners/:id          banner:delete
POST   /api/admin/banners/:id/publish  banner:publish
POST   /api/admin/banners/:id/recall   banner:publish
```

C 端接口：

```text
GET /api/v1/client/banners?position=home_top
```

C 端只返回生效数据，并返回 `id,title,imageUrl,linkType,linkValue,gameId,startAt,endAt`；后台列表可返回草稿、下线和冲突信息。

验收要求：正常 CRUD、重复发布幂等、时间窗冲突、非法外链、越权操作、软删除恢复保护和审计快照测试齐全。

### C 端直播房间

```text
GET /api/v1/client/live/rooms?page=1&pageSize=6
```

接口需要客户端登录，按 `viewers DESC, sort ASC, id DESC` 分页返回 `gb_live_rooms` 中状态为 `live` 的房间；`gameId`、`serverId` 对外返回字符串 ID，未设置时返回 `null`。

响应项包含 `id,title,streamerName,streamerAvatar,coverUrl,viewers,gameId,gameName,serverId,serverName,status,roomUrl,startedAt,endedAt,sort`；业务 ID 格式分别为 `live_{id}`、`game_{id}`、`server_{id}`。分页响应包含 `list,page,pageSize,total,hasMore`，成功消息为 `ok`。

房间列表支持 `keyword,categoryId,gameId,status,sort,page,pageSize,viewers,startedAt,recommendation`；`viewers` 为最低观看人数，`startedAt` 为 RFC3339 起始时间，`sort` 支持 `popular/viewers`、`latest/newest`、`recommended`。

详情接口：`GET /api/v1/client/live/rooms/{id}`，需要客户端登录；`id` 使用 `live_{id}`，返回公告、主播信息、播放流、清晰度、游戏、区服、观看人数和直播状态。未绑定主播/区服时对应对象返回 `null`。

### C 端主播

```text
GET /api/v1/client/live/streamers/{id}
GET /api/v1/client/live/streamers/{id}/rooms?page=1&pageSize=10
GET /api/v1/client/live/streamers?sort=popular&page=1&pageSize=10
```

主播详情返回 `id,name,avatarUrl,coverUrl,description,fans,following,isLive,currentRoomId`；主播列表支持 `popular` 按粉丝数倒序，主播房间列表优先当前直播并分页。

关注接口：`POST/DELETE /api/v1/client/live/streamers/{id}/follow`，以及 `GET /api/v1/client/users/me/live/following`；关注操作幂等并在事务内维护粉丝数。

举报接口：`POST /api/v1/client/reports`，客户端登录后可提交 `live_room` 类型举报，返回 `reportId` 和 `submitted` 状态。

直播埋点统一使用 `POST /api/v1/client/events`，支持 `live_exposure`、`live_click`、`live_enter`、`live_follow`，记录 `resourceType/resourceId/source/sessionId`。

### C 端直播分类

```text
GET /api/v1/client/live/categories
```

无需登录，仅返回 `gb_live_categories.enabled = true` 的分类，按 `sort ASC, id ASC` 排序，响应 `data.list` 项包含 `id,name,type,sort,enabled`。

### `A.USER.01` 玩家用户查询与封禁

玩家账号必须使用独立的 `users`/`gb_users` 表，禁止查询或修改 `sys_user` 代替玩家账号。手机号、身份证摘要和实名状态属于敏感字段，后台响应必须脱敏。

数据表：`gb_users`、`user_bans`

`gb_users` 至少包含：

```text
id, phoneCiphertext, phoneHash, nickname, avatarUrl, status,
realNameStatus, lastLoginAt, registeredAt, createdAt, updatedAt
```

`user_bans` 至少包含：

```text
id, userId, banType, reason, source, startsAt, expiresAt,
status, operatorId, liftedAt, liftedBy, createdAt, updatedAt
```

规则：

- `banType` 为 `mute/login/game/all`；`all` 优先级最高。
- `status` 为 `active/lifted/expired`；临时封禁由服务端判断过期，不能依赖前端倒计时。
- 永久封禁使用 `expiresAt = null`；解封必须保留原记录并写入 `liftedAt/liftedBy`。
- 新增封禁、重复封禁、解封必须幂等；相同用户和相同有效封禁类型不得产生并行有效记录。
- 登录、刷新 Token、获取启动票据时必须校验用户是否被封禁；封禁生效后 Refresh Token 立即失效。
- 手机号查询使用 `phoneHash` 精确匹配，列表展示只返回如 `138****0000` 的脱敏值；日志禁止记录明文手机号。
- 封禁、解封、批量封禁必须在事务内更新状态并写入 `operation_audit_logs`。

后台接口：

```text
GET  /api/admin/users                         user:list
GET  /api/admin/users/:id                     user:list
GET  /api/admin/users/:id/bans                user:list
POST /api/admin/users/:id/ban                 user:ban
PUT  /api/admin/user-bans/:id                 user:ban
POST /api/admin/user-bans/:id/unban           user:unban
```

查询参数：`page,pageSize,keyword,status,realNameStatus,banType,registeredFrom,registeredTo`。

用户详情不得返回 `phoneCiphertext`、身份证明原文、Token、Refresh Token；可返回脱敏手机号、账号状态、实名状态、最近登录时间和有效封禁摘要。

错误码：

```text
10006 账号已封禁
10007 封禁记录不存在
10008 封禁状态已变更
```

验收要求：手机号脱敏和精确查询、分页筛选、临时/永久封禁、重复封禁幂等、自动过期、解封、Token 失效、权限不足、事务回滚和审计快照测试齐全。

## Phase 2/3 任务

- `S.GIFT.*`：礼包、兑换码、原子库存、发奖补偿、对账
- `S.CONTENT.*` / `A.CONTENT.01`：文章、评论、敏感词、审核和封禁
- `A.STATS.01`：DAU/MAU、下载、启动、留存、礼包指标
- `S.MSG.01`：站内信、开服提醒、审核结果
- `A.CHANNEL.01`、`A.ADS.01`、`S.PAY.*`：商业化能力

## 后端接口最低要求

- 使用统一响应和业务错误码：`1xxxx` 认证、`2xxxx` 游戏、`3xxxx` 区服、`4xxxx` 下载、`5xxxx` 礼包、`6xxxx` 内容、`9xxxx` 系统。
- 领礼包、兑换码、支付回调、版本发布必须幂等。
- 时间服务端存 UTC，API 使用 ISO8601。
- 不在普通日志中记录 Token、身份证、完整手机号和启动票据。

## API 契约（联调基线）

### 通用约定

| 项 | 约定 |
|---|---|
| C 端基础路径 | `/api` |
| 后台基础路径 | `/api/admin` |
| 内容类型 | `application/json`；文件上传使用 `multipart/form-data` |
| 鉴权 | `Authorization: Bearer <token>` |
| 追踪 | 支持 `X-Request-ID`，响应始终返回 `requestId` |
| 幂等 | 写接口可传 `Idempotency-Key`，重复请求返回首次结果 |
| 时间 | 数据库存 UTC；接口使用 ISO8601 |
| 分页 | `page` 从 1 开始，`pageSize` 默认 20、最大 100 |
| 删除 | 默认软删除；审计、登录、下载、领取和支付记录禁止物理删除 |

统一成功响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {},
  "requestId": "req_01JXYZ"
}
```

统一失败响应：

```json
{
  "code": 40001,
  "message": "验证码已过期",
  "data": null,
  "requestId": "req_01JXYZ"
}
```

分页 `data`：

```json
{"list": [], "page": 1, "pageSize": 20, "total": 0}
```

### C 端 API

| ID | 方法与路径 | 鉴权 | 请求关键字段 | 返回关键字段 |
|---|---|---|---|---|
| `S.AUTH.01` | `POST /api/auth/sms/send` | 否 | `phone` | `maskedPhone, expireIn` |
| `S.AUTH.01` | `POST /api/auth/sms/login` | 否 | `phone, code, deviceId?` | `accessToken, refreshToken, user` |
| `S.AUTH.02` | `POST /api/auth/refresh` | Refresh | `refreshToken` | 新 Token 对 |
| `S.AUTH.02` | `POST /api/auth/logout` | 登录 | `refreshToken?` | `null` |
| `S.GAMES.01` | `GET /api/games` | 游客 | `tag,status,sort,page,pageSize` | 游戏分页摘要 |
| `S.GAMES.03` | `GET /api/games/:id` | 游客 | 路径 `id` | 游戏详情、版本、公告、区服摘要 |
| `S.SERVER.01` | `GET /api/v1/client/game-servers` | 游客 | `page,pageSize,recommended=true` | 按 `recommendationWeight`、状态、在线人数和开服时间排序的推荐新服分页列表 |
| `S.DL.01` | `GET /api/client/versions/latest` | 游客/登录 | `gameId,platform,channel,version?` | 版本和签名下载信息 |
| `S.DL.03` | `POST /api/downloads/events` | 游客/登录 | `eventType,gameId,version,result,errorCode` | `eventId` |
| `S.DL.03` | `POST /api/games/:gameId/launch-ticket` | 登录 | `serverId,clientVersion` | `launchTicket,expiresAt` |
| `S.CONTENT.04` | `POST /api/reports` | 登录 | `targetType,targetId,reason,detail` | `reportId,status` |
| `S.SUPPORT.01` | `POST /api/feedback` | 游客/登录 | `category,detail,attachmentIds[]` | `feedbackId,status` |

游戏列表项至少包含：`id,name,iconUrl,gameType,versionTags,publisher,rating,downloadCount,status`。
区服至少包含：`id,name,imageUrl,openTime,status,mergeTime,minClientVersion,isRecommended,recommendationWeight`。

### 后台 API

| ID | 方法与路径 | 权限码建议 | 说明 |
|---|---|---|---|
| `A.GAME.01` | `GET/POST/PUT/DELETE /api/admin/games` | `game:list/create/update/delete` | CRUD，删除默认下架 |
| `A.GAME.01` | `POST /api/admin/games/:id/publish` | `game:publish` | 上下架、维护，必须审计 |
| `A.GAME.02` | `GET/POST /api/admin/versions` | `version:list/create` | 版本元数据与文件 |
| `A.GAME.02` | `POST /api/admin/versions/:id/verify` | `version:verify` | SHA、签名、文件清单校验 |
| `A.GAME.02` | `POST /api/admin/versions/:id/publish` | `version:publish` | 发布/撤回，必须审计 |
| `A.BANNER.01` | `GET/POST/PUT/DELETE /api/admin/banners` | `banner:list/create/update/delete` | Banner CRUD、时间窗、运营位和权重 |
| `A.BANNER.01` | `POST /api/admin/banners/:id/publish`、`/recall` | `banner:publish` | 发布/撤回，必须审计 |
| `A.SERVER.01` | `GET/POST/PUT /api/admin/servers` | `server:list/create/update` | 区服 CRUD |
| `A.SERVER.01` | `POST /api/admin/servers/import` | `server:import` | CSV 导入并返回失败行 |
| `A.SERVER.01` | `POST /api/admin/servers/batch-maintain` | `server:maintain` | 批量维护 |
| `A.OPS.02` | `GET /api/admin/audits` | `audit:list/export` | 按人、资源、时间检索 |
| `A.CONTENT.01` | `GET/PUT /api/admin/reports` | `report:list/process` | 举报工单处理 |
| `A.USER.01` | `GET /api/admin/users`、`GET /api/admin/users/:id` | `user:list` | 玩家查询，手机号脱敏、实名和封禁状态 |
| `A.USER.01` | `GET /api/admin/users/:id/bans` | `user:list` | 查询玩家封禁记录 |
| `A.USER.01` | `POST /api/admin/users/:id/ban` | `user:ban` | 禁言、登录封禁、游戏封禁、永久封禁 |
| `A.USER.01` | `PUT /api/admin/user-bans/:id`、`POST /api/admin/user-bans/:id/unban` | `user:ban/unban` | 更新或解除封禁，必须审计 |

后台写接口失败时不得产生半成功状态；外部 OSS/CDN 操作使用任务重试，状态变更和审计写入必须在事务内完成。

### 错误码

```text
10001 手机号格式错误       10002 短信发送过于频繁
10003 验证码错误           10004 验证码已过期
10005 Token 无效            10006 账号已封禁
20001 游戏不存在            20002 游戏已下架
30001 区服不存在            30002 区服维护中
40001 版本不存在            40002 下载地址已过期
40003 SHA 校验失败          40004 签名校验失败
50001 礼包已领取            50002 礼包已售罄
60001 无权操作              60002 内容待审核
90001 参数错误              90002 服务暂不可用
```

错误码必须稳定；`message` 面向用户，详细堆栈只进入结构化日志。

### 文件版本响应

版本接口必须返回 `version,minimumVersion,packageType,downloadUrl,expiresAt,fileSize,sha256,signature,mandatory,changelog`。
下载地址必须是短期签名 URL；客户端不得拼接 OSS/CDN 地址，也不得缓存永久地址。

## Phase 2/3 API 契约

### 礼包与任务

| ID | 方法与路径 | 鉴权 | 说明 |
|---|---|---|---|
| `S.GIFT.01` | `GET /api/gifts` | 游客/登录 | 按 `gameId,type,status,page,pageSize` 查询 |
| `S.GIFT.02` | `POST /api/gifts/:id/claim` | 登录 | 幂等领取，返回领取记录和奖励 |
| `S.GIFT.03` | `POST /api/gifts/redeem` | 登录 | `{code,gameId,serverId?}` 兑换码 |
| `S.GIFT.04` | `GET /api/checkins` | 登录 | 查询签到日历和当前连续天数 |
| `S.GIFT.04` | `POST /api/checkins` | 登录 | 当日签到，幂等 |
| `S.GIFT.05` | `GET /api/tasks` | 登录 | 查询任务和完成状态 |
| `S.GIFT.05` | `POST /api/tasks/:id/claim` | 登录 | 任务奖励领取，走统一发奖中台 |
| `A.GIFT.01` | `GET/POST/PUT/DELETE /api/admin/gifts` | 礼包权限 | 礼包 CRUD、库存、时间窗 |
| `A.GIFT.01` | `GET /api/admin/gifts/:id/claims` | 礼包查看 | 领取流水和风控信息 |
| `A.GIFT.01` | `POST /api/admin/gifts/:id/compensate` | 礼包补发 | 补发必须审计 |
| `A.GIFT.02` | `POST /api/admin/gift-codes/import` | 兑换码管理 | 批量导入并返回失败行 |
| `A.GIFT.02` | `POST /api/admin/gift-codes/:id/revoke` | 兑换码管理 | 作废未使用兑换码 |

礼包领取必须满足：用户/礼包/区服业务唯一约束、库存原子扣减、重复请求返回首次结果、发奖失败进入补偿状态。

### 内容、评论与审核

| ID | 方法与路径 | 鉴权 | 说明 |
|---|---|---|---|
| `S.CONTENT.01` | `GET /api/articles` | 游客 | 游戏、类型、状态、分页查询 |
| `S.CONTENT.01` | `GET /api/articles/:id` | 游客 | 文章详情和作者摘要 |
| `S.CONTENT.01` | `POST /api/articles` | 登录 | 创建草稿/提交审核 |
| `S.CONTENT.02` | `GET /api/articles/:id/comments` | 游客 | 评论分页 |
| `S.CONTENT.02` | `POST /api/articles/:id/comments` | 登录 | 敏感词和风险分处理 |
| `S.CONTENT.02` | `DELETE /api/comments/:id` | 本人/审核员 | 删除本人或违规评论 |
| `A.CONTENT.01` | `GET /api/admin/articles/audit-queue` | 内容审核 | 待审队列 |
| `A.CONTENT.01` | `POST /api/admin/articles/:id/audit` | 内容审核 | `{action,reason}` 通过/驳回/下线 |
| `A.CONTENT.01` | `GET/POST/DELETE /api/admin/sensitive-words` | 内容审核 | 敏感词 CRUD |
| `A.CONTENT.01` | `POST /api/admin/users/:id/mute` | 内容审核 | 禁言/解禁 |

### 消息通知

| ID | 方法与路径 | 鉴权 | 说明 |
|---|---|---|---|
| `S.MSG.01` | `GET /api/messages` | 登录 | 站内信分页 |
| `S.MSG.01` | `GET /api/messages/unread-count` | 登录 | 未读数量 |
| `S.MSG.01` | `POST /api/messages/:id/read` | 登录 | 标记已读，幂等 |
| `S.MSG.01` | `POST /api/messages/read-all` | 登录 | 全部标记已读 |
| `A.MSG.01` | `POST /api/admin/messages` | 消息运营 | 创建定向/全量消息 |
| `A.MSG.01` | `GET /api/admin/messages` | 消息运营 | 查询发送状态和统计 |

### 账户聚合未读统计

`GET /api/v1/client/users/me` 的 `data.user` 增加实时统计字段：

| 字段 | 计算规则 |
|---|---|
| `taskUnreadCount` | 当前可领取且尚未领取的任务数量（与任务中心 `claimable` 状态一致） |
| `messageUnreadCount` | 当前用户 `gb_messages.read_at IS NULL` 的站内信数量 |

未读统计随任务领取、消息标记已读/全部已读后的下一次请求实时更新；未登录或账号无效仍按原认证错误返回。

### 统计与可观测性

| ID | 方法与路径 | 鉴权 | 说明 |
|---|---|---|---|
| `S.OBS.01` | `GET /health` | 内部 | 健康检查 |
| `S.OBS.01` | `GET /metrics` | 内部 | Prometheus 指标 |
| `C.OBS.01` | `POST /api/client/crashes` | 游客/登录 | 崩溃堆栈、版本、设备信息 |
| `A.STATS.01` | `GET /api/admin/stats/overview` | 数据看板 | DAU/MAU、注册、下载、启动 |
| `A.STATS.01` | `GET /api/admin/stats/retention` | 数据看板 | 次留、7 留、渠道维度 |
| `A.STATS.01` | `GET /api/admin/stats/games/:id` | 数据看板 | 单游戏漏斗和区服导流 |
| `A.STATS.01` | `GET /api/admin/stats/export` | 数据导出 | 异步导出，返回任务 ID |

### 商业化、渠道与厂商

| ID | 方法与路径 | 鉴权 | 说明 |
|---|---|---|---|
| `S.PAY.01` | `POST /api/payments/orders` | 实名登录 | 创建充值订单 |
| `S.PAY.01` | `GET /api/payments/orders/:id` | 本人/财务 | 查询订单状态 |
| `S.PAY.01` | `POST /api/payments/callback/:provider` | 支付网关 | 验签、幂等、更新订单状态 |
| `S.PAY.01` | `POST /api/admin/payments/:id/refund` | 财务管理员 | 退款并审计 |
| `A.CHANNEL.01` | `GET/POST/PUT /api/admin/channels` | 渠道管理 | 渠道 CRUD |
| `A.CHANNEL.01` | `GET /api/admin/channels/:id/stats` | 渠道管理 | 归因和 ROI |
| `A.ADS.01` | `GET/POST/PUT/DELETE /api/admin/ads` | 广告管理 | 广告位和投放计划 |
| `A.ADS.01` | `POST /api/ads/events` | 游客/登录 | 曝光、点击事件 |
| `A.PARTNER.01` | `GET/POST/PUT /api/admin/partners` | 管理员 | 厂商/代理管理 |
| `A.PARTNER.01` | `PUT /api/admin/partners/:id/games` | 管理员 | 授权游戏范围 |
| `A.PARTNER.01` | `GET /api/partners/me/stats` | 厂商/代理 | 只读授权游戏数据 |

## 后端完成定义

- 有数据库迁移和索引。
- 有接口请求/响应示例和错误码。
- 有 Casbin 策略及后台资源范围校验。
- 有正常、重复、超时、权限不足和回滚测试。
- 高风险写操作有 `before_snapshot`/`after_snapshot` 审计。
