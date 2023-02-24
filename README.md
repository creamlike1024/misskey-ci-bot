# misskey-ci-bot
A bot that can help you easily update misskey and do some daily maintenance jobs

# 说明

这是一个帮助更新 Misskey 实例的 Telegram bot，目前仅支持单点 docker 部署的实例
支持特性：
- 使用 Telegram 进行管理
- 自动检测 Misskey Release，有更新时发送最近 Changlog 并自动构建镜像（使用 github action）
- 可选于 tg 确认后部署或全自动部署
- 更新前自动备份数据库
- 一键执行 数据库 analyze, 数据库 REINDEX 和数据库备份
- 定时任务（ANALYZE, REINDEX, BACKUP）
- 本地或 SSH 部署（bot 可以不在 misskey 所在机器上运行）

自用偷懒的工具，写得很烂，不过还是有点用（x

使用该 bot 的实例：
- [x] [m.isle.moe](https://m.isle.moe)

# 使用

1. Fork [misskey-docker-ci](https://github.com/creamlike1024/misskey-docker-ci) 仓库，按照 README 配置好 secrets
2. 转到 `Github Settings ->  Developer settings -> Personal access tokens -> Fine-grained tokens` 生成一个新 token，仓库选择刚才 Fork 的 misskey-docker-ci，给予 Actions 的 `Read and Write` 权限，保存
3. 下载 Release。只编译了 linux amd64 版本，需要其它架构可自行进行编译
4. 填充 `config.yml` 配置文件
5. 启动 bot，保持运行