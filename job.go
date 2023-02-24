package main

import (
    "time"

    log "github.com/sirupsen/logrus"
)

func CheckNewRelease() {
    defer func() {
        if err := recover(); err != nil {
            log.Error(err)
            log.Error("CheckNewRelease exited with error, will restart")
        }
    }()
    // 每 5 分钟检查一次
    for {
        releases := GetReleasesInfo()
        // 检查是否有新版本
        for _, release := range releases {
            if !release.Prerelease {
                status := ReadStatus()
                // 比较发布时间检查是否有新版本
                currentVersionPublishedDate, _ := time.Parse(time.RFC3339, status.PublishedDate)
                if release.PublishedAt.After(currentVersionPublishedDate) {
                    // 有新版本，更新 status.json 并向 Channel 发送消息
                    status.LatestVersion = release.TagName
                    status.PublishedDate = release.PublishedAt.Format(time.RFC3339)
                    status.IsBuilt = false
                    status.IsDeployed = false
                    WriteStatus(status)
                    NewRelease <- true
                    log.Infof("New misskey release: %s detected", release.TagName)
                }
                break
            }
        }
        time.Sleep(5 * time.Minute)
    }
}

func CheckBuiltStatus() {
    // 每 1 分钟检查一次
    for {
        <-BuiltInProgress
        for {
            // 每 90 秒检查一次，由于 Github api 有延迟因此放在循环开头，避免第一次检查时获取到上次的结果
            time.Sleep(90 * time.Second)
            runs := GetGithubActionRunStatus()
            if runs[0].Status == "completed" && runs[0].Conclusion == "success" {
                // 构建成功，更新 status.json
                status := ReadStatus()
                status.IsBuilt = true
                WriteStatus(status)
                BuiltSuccess <- true
                log.Infof("Workflow run: %d success", runs[0].Id)
                break
            } else if runs[0].Status == "completed" && runs[0].Conclusion == "failure" {
                BuiltSuccess <- false
                log.Infof("Workflow run: %d failure", runs[0].Id)
                break
            }
        }
    }
}
