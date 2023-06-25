package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
)

type Config struct {
	GithubUsername   string
	GithubToken      string
	ActionRepo       string
	TelegramChatId   int64
	TelegramBotToken string
	DisplayTimezone  string
}

var ConfigInstance Config
var SSHConfigInstance SSHClientConfig
var DeployConfigInstance DeployConfig

func init() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.SetConfigType("yaml")
	err := viper.ReadInConfig()
	if err != nil {
		log.Panic(err)
	}
	log.Info("Using config file: ", viper.ConfigFileUsed())
	// 读取配置
	ConfigInstance = Config{
		GithubUsername:   viper.GetString("github.username"),
		GithubToken:      viper.GetString("github.token"),
		ActionRepo:       viper.GetString("github.action-repo"),
		TelegramChatId:   viper.GetInt64("telegram.chat-id"),
		TelegramBotToken: viper.GetString("telegram.bot-token"),
		DisplayTimezone:  viper.GetString("telegram.display-timezone"),
	}
	// 读取 Deploy 配置
	DeployConfigInstance = DeployConfig{
		Path:              viper.GetString("deploy.path"),
		DockerComposeFile: viper.GetString("deploy.docker-compose-file"),
		DbName:            viper.GetString("deploy.misskey-db-name"),
		DbUser:            viper.GetString("deploy.misskey-db-user"),
		WebContainerName:  viper.GetString("deploy.misskey-container-name"),
		DbContainerName:   viper.GetString("deploy.db-container-name"),
		IsAutoDeploy:      viper.GetBool("deploy.auto-deploy"),
		BackupScript:      viper.GetString("deploy.backup-script"),
	}
	if viper.GetString("deploy.method") == "ssh" {
		DeployConfigInstance.IsLocal = false
	} else if viper.GetString("deploy.method") == "local" {
		DeployConfigInstance.IsLocal = true
	} else {
		log.Panic("Invalid deploy method: ", viper.GetString("deploy.method"))
	}
	// 初始化 SSH 配置
	if !DeployConfigInstance.IsLocal {
		SSHConfigInstance = SSHClientConfig{
			User:                 viper.GetString("ssh.user"),
			Host:                 viper.GetString("ssh.host"),
			Port:                 22,
			Password:             viper.GetString("ssh.password"),
			PrivateKey:           "",
			PrivateKeyPassPhrase: viper.GetString("ssh.key-passphrase"),
		}
		// 如果有端口，就读取端口
		if viper.GetInt("ssh.port") != 0 {
			SSHConfigInstance.Port = viper.GetInt("ssh.port")
		}
		// 如果有私钥，就读取私钥
		if viper.GetString("ssh.key-file") != "" {
			log.Info("Using private key: ", viper.GetString("ssh.key-file"))
			sshKey, err := os.ReadFile(viper.GetString("ssh.key-file"))
			if err != nil {
				log.Panic(err)
			}
			SSHConfigInstance.PrivateKey = string(sshKey)
		}
	}
	// 配置命令
	CommandsInstance = Commands{
		Cd:                  fmt.Sprintf("cd %s", DeployConfigInstance.Path),
		Pull:                fmt.Sprintf("docker compose -f %s pull -q", DeployConfigInstance.DockerComposeFile),
		Down:                fmt.Sprintf("docker compose -f %s down", DeployConfigInstance.DockerComposeFile),
		Up:                  fmt.Sprintf("docker compose -f %s up -d", DeployConfigInstance.DockerComposeFile),
		Prune:               "docker system prune -f",
		StartWeb:            fmt.Sprintf("docker compose -f %s start %s", DeployConfigInstance.DockerComposeFile, DeployConfigInstance.WebContainerName),
		StopWeb:             fmt.Sprintf("docker compose -f %s stop %s", DeployConfigInstance.DockerComposeFile, DeployConfigInstance.WebContainerName),
		RestartWeb:          fmt.Sprintf("docker compose -f %s restart %s", DeployConfigInstance.DockerComposeFile, DeployConfigInstance.WebContainerName),
		RestartAll:          fmt.Sprintf("docker compose -f %s restart", DeployConfigInstance.DockerComposeFile),
		DbReindex:           fmt.Sprintf("docker compose -f %s exec db reindexdb -U %s -d %s", DeployConfigInstance.DockerComposeFile, DeployConfigInstance.DbUser, DeployConfigInstance.DbName),
		DbAnalyze:           fmt.Sprintf("docker compose -f %s exec db psql -U %s -c 'ANALYZE;'", DeployConfigInstance.DockerComposeFile, DeployConfigInstance.DbUser),
		DbVacuumFull:        fmt.Sprintf("docker compose -f %s exec db psql -U %s -c 'VACUUM FULL;'", DeployConfigInstance.DockerComposeFile, DeployConfigInstance.DbUser),
		DbVacuumFullAnalyze: fmt.Sprintf("docker compose -f %s exec db psql -U %s -c 'VACUUM(FULL, ANALYZE);'", DeployConfigInstance.DockerComposeFile, DeployConfigInstance.DbUser),
		DbVacuum:            fmt.Sprintf("docker compose -f %s exec db psql -U %s -c 'VACUUM;'", DeployConfigInstance.DockerComposeFile, DeployConfigInstance.DbUser),
		DbVacuumAnalyze:     fmt.Sprintf("docker compose -f %s exec db psql -U %s -c 'VACUUM(ANALYZE);'", DeployConfigInstance.DockerComposeFile, DeployConfigInstance.DbUser),
	}
	cronjobs := viper.Get("cron")
	// 从 cronjons 里读取所有的 cronjob 到 CronJobSlice
	for _, cronjob := range cronjobs.([]interface{}) {
		cronjobMap := cronjob.(map[string]interface{})
		cronjobAction := cronjobMap["action"].(string)
		cronjobExpression := cronjobMap["cron"].(string)
		cronjobInstance := CronJob{
			action: cronjobAction,
			cron:   cronjobExpression,
		}
		CronJobSlice = append(CronJobSlice, cronjobInstance)
	}
	// 初始化 channel
	IsFirstRun = false
	NewRelease = make(chan bool)
	BuiltInProgress = make(chan bool)
	BuiltSuccess = make(chan bool)
	IndependentNotification = make(chan string)
	StartDeploy = make(chan bool)
}

func main() {
	// 初始化状态
	InitStatus()
	// 启动定时任务
	go CronJobRunner()
	// 启动所有 goroutine
	go func() {
		for {
			CheckNewRelease()
		}
	}()
	go CheckBuiltStatus()
	for {
		// 如果 Telegram Bot 崩溃了，就重启
		RunTelegramBot()
	}
}
