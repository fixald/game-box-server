-- GameBox 本地/测试环境种子数据（MySQL 8+）
-- 使用方式：mysql -u<user> -p <database> < config/seed_gamebox.sql
-- 仅用于开发测试，不得导入生产环境。

SET time_zone = '+00:00';

INSERT IGNORE INTO gb_checkin_rewards (level,name,reward,icon,status,created_at,updated_at) VALUES
  (1,'初次签到','10 积分','✦','active',UTC_TIMESTAMP(),UTC_TIMESTAMP()),
  (3,'三日礼','30 积分','✦','active',UTC_TIMESTAMP(),UTC_TIMESTAMP()),
  (7,'七日礼包','礼包','🎁','active',UTC_TIMESTAMP(),UTC_TIMESTAMP()),
  (14,'半月奖励','SVIP经验','♛','active',UTC_TIMESTAMP(),UTC_TIMESTAMP()),
  (30,'满月大奖','豪华礼包','🎁','active',UTC_TIMESTAMP(),UTC_TIMESTAMP());

INSERT IGNORE INTO gb_tasks
  (code,category,title,description,icon,target,points,action_label,action_route,status,sort,created_at,updated_at)
VALUES
  ('task-1','daily','每日登录盒子','登录客户端即可完成','◷',1,20,'','', 'active',1,UTC_TIMESTAMP(),UTC_TIMESTAMP()),
  ('task-2','daily','观看直播 10 分钟','在直播频道观看任意直播','▶',10,50,'去直播','#/live','active',2,UTC_TIMESTAMP(),UTC_TIMESTAMP()),
  ('task-3','game','启动一次游戏','启动任意已安装的传奇游戏','⚔',1,20,'','', 'active',3,UTC_TIMESTAMP(),UTC_TIMESTAMP()),
  ('task-4','newbie','完善个人资料','上传头像并设置昵称','◇',1,100,'去完善','#/settings','active',4,UTC_TIMESTAMP(),UTC_TIMESTAMP()),
  ('task-5','social','关注 3 位主播','关注你喜欢的传奇主播','♡',3,30,'去关注','#/live','active',5,UTC_TIMESTAMP(),UTC_TIMESTAMP()),
  ('task-6','newbie','完成首次下载','下载并校验一款游戏','↓',1,0,'','', 'active',6,UTC_TIMESTAMP(),UTC_TIMESTAMP()),
  ('task-7','daily','浏览新服推荐','查看今日推荐新区和开服信息','◈',1,10,'去新服','#/','active',7,UTC_TIMESTAMP(),UTC_TIMESTAMP()),
  ('task-8','daily','完成一次签到','在任务中心完成每日签到','✓',1,20,'','', 'active',8,UTC_TIMESTAMP(),UTC_TIMESTAMP()),
  ('task-9','game','查看游戏详情','浏览任意一款游戏的详情页面','◉',1,15,'去游戏','#/games','active',9,UTC_TIMESTAMP(),UTC_TIMESTAMP()),
  ('task-10','game','进入推荐区服','从新服推荐中选择一个区服','⚑',1,0,'','', 'active',10,UTC_TIMESTAMP(),UTC_TIMESTAMP()),
  ('task-11','social','查看一条资讯','阅读平台最新游戏资讯','▤',1,10,'去资讯','#/news','active',11,UTC_TIMESTAMP(),UTC_TIMESTAMP()),
  ('task-12','newbie','完成首次区服选择','选择喜欢的游戏区服','◇',1,50,'','', 'active',12,UTC_TIMESTAMP(),UTC_TIMESTAMP());

UPDATE gb_tasks SET rewards = JSON_ARRAY(JSON_OBJECT('type','points','name','积分','amount',points,'icon','✦'))
WHERE code IN ('task-1','task-2','task-3','task-4','task-5','task-7','task-8','task-9','task-12');
UPDATE gb_tasks SET rewards = JSON_ARRAY(JSON_OBJECT('type','coupon','name','下载加速券','amount',1,'icon','⚡')) WHERE code='task-6';
UPDATE gb_tasks SET rewards = JSON_ARRAY(JSON_OBJECT('type','gift','name','区服礼包','icon','🎁')) WHERE code='task-10';


INSERT IGNORE INTO games
  (id,name,slug,description,icon_url,category,game_type,publisher,rating,download_count,version_tags,status,created_at,updated_at)
VALUES
  (1001,'星海远征','star-expedition','科幻题材多人在线 RPG','https://example.test/assets/star-expedition.png','RPG','online','GameBox Studio',4.8,125000,'official,hot,1.0','published',UTC_TIMESTAMP(),UTC_TIMESTAMP()),
  (1002,'像素勇者','pixel-heroes','轻量像素冒险游戏','https://example.test/assets/pixel-heroes.png','冒险','standalone','GameBox Studio',4.3,87000,'official,1.2','published',UTC_TIMESTAMP(),UTC_TIMESTAMP()),
  (1003,'测试下架游戏','offline-demo','用于测试下架不可见','https://example.test/assets/offline-demo.png','测试','online','QA',0,0,'test','offline',UTC_TIMESTAMP(),UTC_TIMESTAMP());

INSERT IGNORE INTO servers
  (id,game_id,name,image_url,open_time,status,merge_time,min_client_version,is_recommended,recommendation_weight,created_at,updated_at)
VALUES
  (2001,1001,'星海一区','https://example.test/assets/servers/star-1.png',UTC_TIMESTAMP() - INTERVAL 30 DAY,'normal',NULL,'1.0.0',1,80,UTC_TIMESTAMP(),UTC_TIMESTAMP()),
  (2002,1001,'星海二区','https://example.test/assets/servers/star-2.png',UTC_TIMESTAMP() - INTERVAL 2 DAY,'normal',NULL,'1.0.0',0,0,UTC_TIMESTAMP(),UTC_TIMESTAMP()),
  (2003,1001,'星海三区','https://example.test/assets/servers/star-3.png',UTC_TIMESTAMP() + INTERVAL 2 DAY,'preview',NULL,'1.1.0',1,100,UTC_TIMESTAMP(),UTC_TIMESTAMP()),
  (2004,1002,'像素一区','https://example.test/assets/servers/pixel-1.png',UTC_TIMESTAMP() - INTERVAL 60 DAY,'maintenance',NULL,'1.2.0',1,50,UTC_TIMESTAMP(),UTC_TIMESTAMP()),
  (2005,1002,'像素二区','https://example.test/assets/servers/pixel-2.png',UTC_TIMESTAMP() + INTERVAL 3 HOUR,'opening_soon',NULL,'1.2.0',1,95,UTC_TIMESTAMP(),UTC_TIMESTAMP()),
  (2006,1002,'像素三区','https://example.test/assets/servers/pixel-3.png',UTC_TIMESTAMP() - INTERVAL 1 HOUR,'hot',NULL,'1.2.0',1,70,UTC_TIMESTAMP(),UTC_TIMESTAMP()),
  (2007,1001,'星海四区','https://example.test/assets/servers/star-4.png',UTC_TIMESTAMP() + INTERVAL 1 DAY,'preview',NULL,'1.1.0',1,40,UTC_TIMESTAMP(),UTC_TIMESTAMP()),
  (2008,1001,'星海五区','https://example.test/assets/servers/star-5.png',UTC_TIMESTAMP() - INTERVAL 6 HOUR,'normal',NULL,'1.1.0',1,20,UTC_TIMESTAMP(),UTC_TIMESTAMP()),
  (2009,1002,'像素四区','https://example.test/assets/servers/pixel-4.png',UTC_TIMESTAMP() + INTERVAL 6 HOUR,'opening_soon',NULL,'1.2.0',0,0,UTC_TIMESTAMP(),UTC_TIMESTAMP()),
  (2010,1001,'星海测试维护服','https://example.test/assets/servers/star-maintenance.png',UTC_TIMESTAMP() + INTERVAL 2 HOUR,'maintenance',NULL,'1.1.0',1,120,UTC_TIMESTAMP(),UTC_TIMESTAMP());

INSERT IGNORE INTO banners
  (id,title,image_url,link_type,link_value,position,weight,game_id,start_at,end_at,status,sort,created_by,updated_by,created_at,updated_at)
VALUES
  (3001,'星海远征新服开启','https://example.test/assets/banner-star.jpg','game','1001','home_top',100,1001,UTC_TIMESTAMP() - INTERVAL 1 DAY,UTC_TIMESTAMP() + INTERVAL 30 DAY,'published',1,1,1,UTC_TIMESTAMP(),UTC_TIMESTAMP()),
  (3002,'像素勇者版本更新','https://example.test/assets/banner-pixel.jpg','game','1002','home_top',80,1002,UTC_TIMESTAMP() - INTERVAL 1 DAY,UTC_TIMESTAMP() + INTERVAL 15 DAY,'published',2,1,1,UTC_TIMESTAMP(),UTC_TIMESTAMP()),
  (3003,'待发布 Banner','https://example.test/assets/banner-draft.jpg','none','','home_top',10,NULL,UTC_TIMESTAMP(),UTC_TIMESTAMP() + INTERVAL 60 DAY,'draft',3,1,1,UTC_TIMESTAMP(),UTC_TIMESTAMP());

INSERT IGNORE INTO gb_users
  (id,phone_hash,phone_ciphertext,nickname,avatar_url,status,token_version,real_name_status,created_at,updated_at)
VALUES
  (4001,'a6942f9771d67f34034d2f1926988ed3fad3bf1b4e7cedb9a31f31398dea43bc','138****8000','星海玩家','https://example.test/avatar/1.png','active',1,'unverified',UTC_TIMESTAMP(),UTC_TIMESTAMP()),
  (4002,'f1d8142cbb59c0a2f93f91fbe934f83f9afbdab0b8fafaabad0f842b32aab322','139****9000','测试封禁用户','', 'banned',2,'unverified',UTC_TIMESTAMP(),UTC_TIMESTAMP()),
  (4003,'159e1d4d48ef2b3538470fc327cc7b7031e8a19211213bc1c12784b66f69e9a0','186****6000','反馈用户','', 'active',1,'verified',UTC_TIMESTAMP(),UTC_TIMESTAMP());

INSERT IGNORE INTO gb_user_bans
  (id,user_id,ban_type,reason,source,starts_at,expires_at,status,operator_id,created_at,updated_at)
VALUES
  (5001,4002,'login','测试永久封禁','seed',UTC_TIMESTAMP() - INTERVAL 1 DAY,NULL,'active',1,UTC_TIMESTAMP(),UTC_TIMESTAMP());

INSERT IGNORE INTO reports
  (id,user_id,target_type,target_id,reason,detail,status,created_at,updated_at)
VALUES
  (6001,4001,'comment','demo-comment-1','违规内容','测试举报工单，请处理','pending',UTC_TIMESTAMP(),UTC_TIMESTAMP()),
  (6002,4003,'user','4002','恶意行为','测试已处理举报','resolved',UTC_TIMESTAMP() - INTERVAL 1 DAY,UTC_TIMESTAMP());

INSERT IGNORE INTO feedbacks
  (id,user_id,category,detail,attachment_ids,status,result,created_at,updated_at)
VALUES
  (7001,4001,'bug','登录后游戏列表加载较慢','[]','pending','',UTC_TIMESTAMP(),UTC_TIMESTAMP()),
  (7002,4003,'suggestion','希望增加开服提醒','[]','processing','已转交运营',UTC_TIMESTAMP(),UTC_TIMESTAMP());

-- 登录联调：账号 user@example.com / Password123!
INSERT IGNORE INTO gb_agreements
  (id,version,title,content_url,summary,status,published_at,created_at,updated_at)
VALUES
  (8001,'2026-07-01','用户协议与隐私政策','https://cdn.example.com/agreements/2026-07-01.html','登录即表示同意最新用户协议与隐私政策','published',UTC_TIMESTAMP(),UTC_TIMESTAMP(),UTC_TIMESTAMP());

INSERT IGNORE INTO gb_users
  (id,phone_hash,phone_ciphertext,account,account_hash,password_hash,nickname,avatar_url,status,vip_level,token_version,agreement_version,real_name_status,created_at,updated_at)
VALUES
  (4004,'359ea74a80a57accd42a7311ed96eca04f3e631d0ab34ea76808c543240d8a68','138****0000','user@example.com','b4c9a289323b21a01c3e940f150eb9b8c542587f1abfd8f0e1cc1ffc5e475514','$2a$10$ik5p.0s0R1KW3ShenTW9x.DSflnD7jnV33M7EcgXkypOBjGrrvjsq','玩家','https://cdn.example.com/avatar.png','active',0,1,'2026-07-01','unverified',UTC_TIMESTAMP(),UTC_TIMESTAMP());
