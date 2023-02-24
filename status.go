package main

import (
    "encoding/json"
    "io"
    "os"
    "time"

    log "github.com/sirupsen/logrus"
)

type Status struct {
    LatestVersion string
    PublishedDate string
    IsBuilt       bool
    IsDeployed    bool
}

var IsFirstRun bool
var NewRelease chan bool
var BuiltInProgress chan bool
var BuiltSuccess chan bool
var StartDeploy chan bool

func ReadStatus() Status {
    statusFile, err := os.Open("status.json")
    if err != nil {
        log.Panic(err)
    }
    defer func(statusFile *os.File) {
        err := statusFile.Close()
        if err != nil {
            log.Panic(err)
        }
    }(statusFile)
    var status Status
    statusBytes, _ := io.ReadAll(statusFile)
    err = json.Unmarshal(statusBytes, &status)
    if err != nil {
        log.Panic(err)
    }
    return status
}

func WriteStatus(status Status) {
    statusBytes, _ := json.Marshal(status)
    statusFile, err := os.Create("status.json")
    if err != nil {
        log.Panic(err)
    }
    defer func(statusFile *os.File) {
        err := statusFile.Close()
        if err != nil {
            log.Panic(err)
        }
    }(statusFile)
    _, err = statusFile.Write(statusBytes)
    if err != nil {
        log.Panic(err)
    }
}

func InitStatus() {
    // 如果 status.json 不存在，则创建一个
    if _, err := os.Stat("status.json"); os.IsNotExist(err) {
        log.Info("首次运行，生成 status.json")
        RefreshStatus()
        IsFirstRun = true
    }
}

func RefreshStatus() {
    releaseInfo := GetReleasesInfo()
    for _, release := range releaseInfo {
        if !release.Prerelease {
            status := Status{
                LatestVersion: release.TagName,
                PublishedDate: release.PublishedAt.Format(time.RFC3339),
                IsBuilt:       false,
                IsDeployed:    false,
            }
            WriteStatus(status)
            break
        }
    }
}
