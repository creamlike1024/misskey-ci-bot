package main

import (
    "fmt"
    "github.com/go-co-op/gocron"
    log "github.com/sirupsen/logrus"
    "time"
)

var CronJobSlice []CronJob

type CronJob struct {
    action string
    cron   string
}

func CronJobRunner() {
    s := gocron.NewScheduler(time.Local)
    s.TagsUnique()
    for _, job := range CronJobSlice {
        switch job.action {
        case "reindex":
            s.Tag(job.action).Cron(job.cron).Do(func() {
                err := DbReindex()
                CronJobRepoter(err, job.action)
            })
        case "backup":
            s.Tag(job.action).Cron(job.cron).Do(func() {
                err := DbBackup(false)
                CronJobRepoter(err, job.action)
            })
        case "analyse":
            s.Tag(job.action).Cron(job.cron).Do(func() {
                err := DbAnalyze()
                CronJobRepoter(err, job.action)
            })
        }
    }
    s.StartBlocking()
}

func CronJobRepoter(err error, action string) {
    // 根据任务执行结果发送通知
    if err != nil {
        log.Infof("CronJob: %s excuted", action)
        IndependentNotification <- fmt.Sprintf("定时任务：%s 执行失败", action)
        return
    }
    log.Infof("CronJob: %s excute success", action)
    IndependentNotification <- fmt.Sprintf("定时任务：%s 执行成功", action)
}

func GetCronJobsStatus() []CronJob {
    var cronjobStatus []CronJob
    // 将 cronjob 和对应的 nextRunTime 放到一个 map 里
    for _, job := range CronJobSlice {
        cronjobStatus = append(cronjobStatus, CronJob{
            action: job.action,
            cron:   job.cron,
        })
    }
    return cronjobStatus
}
