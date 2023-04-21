package main

import (
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	log "github.com/sirupsen/logrus"
)

var IndependentNotification chan string

func RunTelegramBot() {
	startTime := time.Now()
	bot, err := tgbotapi.NewBotAPI(ConfigInstance.TelegramBotToken)
	if err != nil {
		log.Panic(err)
	}
	log.Info("Authorized on account ", bot.Self.UserName)
	defer func() {
		// 捕获 panic
		if err := recover(); err != nil {
			log.Error(err)
			log.Error("Telegram bot crashed, restarting...")
		}
	}()
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, _ := bot.GetUpdatesChan(u)
	// 监听新版本发布和构建状态
	go independentNotifier(bot)
	go notifyFirstRun()
	go notifyNewRelease()
	go notifyBuildStatus()
	go autoUpdateMisskey()
	for update := range updates {
		if update.Message.Chat.ID != ConfigInstance.TelegramChatId || update.Message == nil || update.Message.Time().Before(startTime) {
			continue
		}
		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "help":
				go func() {
					_, _ = bot.Send(help(update.Message))
				}()
			case "showActionStatus":
				go func() {
					_, _ = bot.Send(showActionStatus(update.Message))
				}()
			case "deleteAllWorkflowRuns":
				go func() {
					_, _ = bot.Send(deleteAllWorkflowRuns(update.Message))
				}()
			case "manuallyRunWorkflow":
				go func() {
					_, _ = bot.Send(manuallyRunWorkflow(update.Message))
				}()
			case "getChangeLog":
				go func() {
					_, _ = bot.Send(getChangeLog(update.Message))
				}()
			case "update":
				go func() {
					_, _ = bot.Send(updateMisskey(update.Message))
				}()
			case "forceUpdate":
				go forceUpdateMisskey()
			case "status":
				go func() {
					_, _ = bot.Send(getStatus(update.Message))
				}()
			case "cronStatus":
				go func() {
					cronStatus := GetCronJobsStatus()
					var text string
					for _, cronJob := range cronStatus {
						text += fmt.Sprintf("任务：%s\nCron 表达式：%s\n\n", cronJob.action, cronJob.cron)
					}
					_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, text))
				}()
			case "dbBackup":
				go func() {
					_, _ = bot.Send(dbBackup(update.Message))
				}()
			case "dbReindex":
				go func() {
					_, _ = bot.Send(dbReindex(update.Message))
				}()
			case "dbAnalyze":
				go func() {
					_, _ = bot.Send(dbAnalyze(update.Message))
				}()
			case "down":
				go func() {
					_, _ = bot.Send(down(update.Message))
				}()
			case "up":
				go func() {
					_, _ = bot.Send(up(update.Message))
				}()
			case "restart":
				go func() {
					_, _ = bot.Send(restart(update.Message))
				}()
			case "startMisskey":
				go func() {
					_, _ = bot.Send(startMisskey(update.Message))
				}()
			case "stopMisskey":
				go func() {
					_, _ = bot.Send(stopMisskey(update.Message))
				}()
			default:
				go func() {
					_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Unknown command"))
				}()
			}
		}
	}
}

func independentNotifier(bot *tgbotapi.BotAPI) tgbotapi.MessageConfig {
	for {
		text := <-IndependentNotification
		_, _ = bot.Send(tgbotapi.NewMessage(ConfigInstance.TelegramChatId, text))
	}
}

func notifyNewRelease() {
	for {
		<-NewRelease
		releases := GetReleasesInfo()
		var text string
		text += "检测到新版本！\n"
		for _, release := range releases {
			if !release.Prerelease {
				text += "当前最新版本为：" + release.TagName + "\n"
				loc, _ := time.LoadLocation(ConfigInstance.DisplayTimezone)
				text += fmt.Sprintf("发布时间：%s\n%s\n\n", release.PublishedAt.In(loc).Format("2006-01-02 15:04:05 MST -0700"), release.HtmlUrl)
				text += "最近的 ChangeLog：\n" + GetChangeLog() + "\n\n"
				ok := RunGithubAction()
				if ok {
					text += "已开始构建镜像..."
					BuiltInProgress <- true
					log.Info("Start building image...")
				} else {
					text += "自动开始构建镜像失败，请手动执行"
				}
				break
			}
		}
		IndependentNotification <- text
	}
}

func notifyBuildStatus() {
	for {
		result := <-BuiltSuccess
		if result {
			if DeployConfigInstance.IsAutoDeploy {
				StartDeploy <- true
				IndependentNotification <- "镜像构建成功！开始部署更新..."
				continue
			}
			IndependentNotification <- "镜像构建成功！输入 /update 开始更新"
		} else {
			IndependentNotification <- "镜像构建失败！请手动查看 workflow 运行情况"
		}
	}
}

func notifyFirstRun() {
	// 如果是第一次运行，发送消息
	if IsFirstRun {
		var text string
		text += "欢迎使用 Misskey Ci Bot！\n\n这似乎是第一次运行，设置默认状态：\n"
		status := ReadStatus()
		text += fmt.Sprintf("当前版本为：%s\n", status.LatestVersion) +
			fmt.Sprintf("是否已构建镜像 %t\n", status.IsBuilt) +
			fmt.Sprintf("是否已部署更新 %t\n\n", status.IsDeployed) +
			"我将在后台定时检测新版本并自动发起构建镜像，同时进行通知。\n" +
			"如果你还未更新这个版本，可以输入 /forceUpdate 手动执行更新\n\n" +
			"输入 /help 获取命令列表"
		IndependentNotification <- text
	}
}

func help(msg *tgbotapi.Message) tgbotapi.MessageConfig {
	text := "Available commands:\n" +
		"/help - 显示帮助\n" +
		"/showActionStatus - 显示最后一个 Action Run 状态\n" +
		"/deleteAllWorkflowRuns - 删除所有 workflow run\n" +
		"/manuallyRunWorkflow - 手动发起执行 workflow\n" +
		"/getChangeLog - 获取最近的 ChangeLog\n" +
		"/forceUpdate - 强制更新\n" +
		"/update - update Misskey\n" +
		"/status - 显示当前 status\n" +
		"/cronStatus - 显示当前 cron 任务状态\n" +
		"/dbBackup - 备份数据库\n" +
		"/dbReindex - 重建数据库索引\n" +
		"/dbAnalyze - 更新数据库统计数据，通常导入备份后执行\n" +
		"/down - Down 所有容器\n" +
		"/up - Up 所有容器\n" +
		"/restart - 重启所有容器\n" +
		"/startMisskey - 启动 Misskey 容器\n" +
		"/stopMisskey - 停止 Misskey 容器\n"
	m := tgbotapi.NewMessage(msg.Chat.ID, text)
	m.ReplyToMessageID = msg.MessageID
	return m
}

func showActionStatus(msg *tgbotapi.Message) tgbotapi.MessageConfig {
	runs := GetGithubActionRunStatus()
	// 如果没有运行记录，返回空
	if len(runs) == 0 {
		return tgbotapi.NewMessage(msg.Chat.ID, "没有 workflow 运行记录")
	}
	var text string
	text += fmt.Sprintf("Latest workflow run %d\nStatus: %s, %s", runs[0].Id, runs[0].Status, runs[0].Conclusion) +
		fmt.Sprintf("\n详情：%s", runs[0].HtmlUrl)
	m := tgbotapi.NewMessage(msg.Chat.ID, text)
	m.ReplyToMessageID = msg.MessageID
	return m
}

func deleteAllWorkflowRuns(msg *tgbotapi.Message) tgbotapi.MessageConfig {
	runs := GetGithubActionRunStatus()
	// 如果没有运行记录，返回空
	if len(runs) == 0 {
		return tgbotapi.NewMessage(msg.Chat.ID, "目前没有 workflow 运行记录")
	}
	var runIds []int64
	for _, run := range runs {
		runIds = append(runIds, run.Id)
	}
	result := DeleteGithubActionRun(runIds)
	var text string
	if result == nil {
		text = "已删除所有 workflow 运行记录"
	} else {
		text = "未能完全删除所有 workflow 运行记录"
	}
	m := tgbotapi.NewMessage(msg.Chat.ID, text)
	m.ReplyToMessageID = msg.MessageID
	return m
}

func manuallyRunWorkflow(msg *tgbotapi.Message) tgbotapi.MessageConfig {
	ok := RunGithubAction()
	var text string
	if ok {
		text = "新的 workflow 已经开始运行"
	} else {
		text = "执行失败"
	}
	m := tgbotapi.NewMessage(msg.Chat.ID, text)
	m.ReplyToMessageID = msg.MessageID
	return m
}

func getStatus(msg *tgbotapi.Message) tgbotapi.MessageConfig {
	status := ReadStatus()
	var text string
	text += fmt.Sprintf("最新 Misskey 版本：%s\n", status.LatestVersion)
	publishedTime, _ := time.Parse(time.RFC3339, status.PublishedDate)
	loc, _ := time.LoadLocation(ConfigInstance.DisplayTimezone)
	text += fmt.Sprintf("发布时间：%s\n", publishedTime.In(loc).Format("2006-01-02 15:04:05 MST -0700")) +
		fmt.Sprintf("是否已构建镜像：%t\n", status.IsBuilt) +
		fmt.Sprintf("是否已部署：%t", status.IsDeployed)
	m := tgbotapi.NewMessage(msg.Chat.ID, text)
	m.ReplyToMessageID = msg.MessageID
	return m
}

func dbBackup(msg *tgbotapi.Message) tgbotapi.MessageConfig {
	var text string
	IndependentNotification <- "开始备份数据库，请耐心等待"
	if err := DbBackup(false); err != nil {
		text += "备份失败"
	} else {
		text += "备份成功"
	}
	m := tgbotapi.NewMessage(msg.Chat.ID, text)
	m.ReplyToMessageID = msg.MessageID
	return m
}

func dbReindex(msg *tgbotapi.Message) tgbotapi.MessageConfig {
	var text string
	IndependentNotification <- "开始重建数据库索引，将会停止 Misskey 容器并需要较长时间，请耐心等待"
	if err := DbReindex(); err != nil {
		text += "重建索引失败"
	} else {
		text += "重建索引成功"
	}
	err := StartMisskeyContainer()
	if err != nil {
		text += "，重新启动 Misskey 容器失败"
	} else {
		text += "，重新启动 Misskey 容器成功"
	}
	m := tgbotapi.NewMessage(msg.Chat.ID, text)
	m.ReplyToMessageID = msg.MessageID
	return m
}

func dbAnalyze(msg *tgbotapi.Message) tgbotapi.MessageConfig {
	var text string
	IndependentNotification <- "开始更新数据库统计数据，请耐心等待"
	if err := DbAnalyze(); err != nil {
		text += "更新统计数据失败"
	} else {
		text += "更新统计数据成功"
	}
	m := tgbotapi.NewMessage(msg.Chat.ID, text)
	m.ReplyToMessageID = msg.MessageID
	return m
}

func up(msg *tgbotapi.Message) tgbotapi.MessageConfig {
	var text string
	err := ContainersUp()
	if err != nil {
		text += "失败"
	} else {
		text += "启动所有容器成功"
	}
	m := tgbotapi.NewMessage(msg.Chat.ID, text)
	m.ReplyToMessageID = msg.MessageID
	return m
}

func down(msg *tgbotapi.Message) tgbotapi.MessageConfig {
	var text string
	err := ContainersDown()
	if err != nil {
		text += "失败"
	} else {
		text += "停止所有容器成功"
	}
	m := tgbotapi.NewMessage(msg.Chat.ID, text)
	m.ReplyToMessageID = msg.MessageID
	return m
}

func restart(msg *tgbotapi.Message) tgbotapi.MessageConfig {
	var text string
	err := ContainersRestart()
	if err != nil {
		text += "失败"
	} else {
		text += "重启所有容器成功"
	}
	m := tgbotapi.NewMessage(msg.Chat.ID, text)
	m.ReplyToMessageID = msg.MessageID
	return m
}

func stopMisskey(msg *tgbotapi.Message) tgbotapi.MessageConfig {
	var text string
	err := StopMisskeyContainer()
	if err != nil {
		text += "失败"
	} else {
		text += "停止 Misskey 容器成功"
	}
	m := tgbotapi.NewMessage(msg.Chat.ID, text)
	m.ReplyToMessageID = msg.MessageID
	return m
}

func startMisskey(msg *tgbotapi.Message) tgbotapi.MessageConfig {
	var text string
	err := StartMisskeyContainer()
	if err != nil {
		text += "失败"
	} else {
		text += "启动 Misskey 容器成功"
	}
	m := tgbotapi.NewMessage(msg.Chat.ID, text)
	m.ReplyToMessageID = msg.MessageID
	return m
}

func getChangeLog(msg *tgbotapi.Message) tgbotapi.MessageConfig {
	m := tgbotapi.NewMessage(msg.Chat.ID, GetChangeLog())
	m.ReplyToMessageID = msg.MessageID
	return m
}

func updateMisskey(msg *tgbotapi.Message) tgbotapi.MessageConfig {
	var text string
	status := ReadStatus()
	if !status.IsBuilt {
		text += fmt.Sprintf("当前版本：%s 未构建，如果状态异常，输入 /manuallyRunWorkflow 手动发起构建或 /forceUpdate 强制更新\n", status.LatestVersion)
		m := tgbotapi.NewMessage(msg.Chat.ID, text)
		m.ReplyToMessageID = msg.MessageID
		return m
	}
	if status.IsDeployed {
		text += fmt.Sprintf("当前版本：%s 已部署，如果状态异常，输入 /forceUpdate 强制更新\n", status.LatestVersion)
		m := tgbotapi.NewMessage(msg.Chat.ID, text)
		m.ReplyToMessageID = msg.MessageID
		return m
	}
	err := DbBackup(true)
	if err != nil {
		text += "备份数据库失败，终止操作"
		m := tgbotapi.NewMessage(msg.Chat.ID, text)
		m.ReplyToMessageID = msg.MessageID
		return m
	}
	IndependentNotification <- "数据库备份完毕，开始部署更新，请耐心等待"
	err = Deploy()
	if err != nil {
		text += "部署更新失败"
	} else {
		text += fmt.Sprintf("Misskey %s 更新成功\n", status.LatestVersion)
		status.IsDeployed = true
		WriteStatus(status)
	}
	m := tgbotapi.NewMessage(msg.Chat.ID, text)
	m.ReplyToMessageID = msg.MessageID
	return m
}

func forceUpdateMisskey() {
	IndependentNotification <- "刷新状态..."
	RefreshStatus()
	ok := RunGithubAction()
	if ok {
		IndependentNotification <- "已开始构建镜像..."
		BuiltInProgress <- true
		log.Info("Start building image...")
	} else {
		IndependentNotification <- "自动开始构建镜像失败，请手动执行"
	}
}

func autoUpdateMisskey() {
	for {
		<-StartDeploy
		status := ReadStatus()
		err := DbBackup(true)
		if err != nil {
			IndependentNotification <- "数据库备份失败,终止部署"
			return
		}
		err = Deploy()
		if err != nil {
			IndependentNotification <- "部署更新失败"
		} else {
			text := fmt.Sprintf("Misskey %s 更新成功\n", status.LatestVersion)
			IndependentNotification <- text
			status.IsDeployed = true
			WriteStatus(status)
		}
	}
}
