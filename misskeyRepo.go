package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    md "github.com/JohannesKaufmann/html-to-markdown"
    log "github.com/sirupsen/logrus"
    "github.com/yuin/goldmark"
    "io"
    "net/http"
    "strings"
    "time"
)

type ReleaseInfo []struct {
    HtmlUrl         string    `json:"html_url"`
    TagName         string    `json:"tag_name"`
    TargetCommitish string    `json:"target_commitish"`
    Prerelease      bool      `json:"prerelease"`
    PublishedAt     time.Time `json:"published_at"`
}

func GetChangeLog() string {
    // 获取 changelog
    // 发送 GET 请求
    resp, err := http.Get("https://raw.githubusercontent.com/misskey-dev/misskey/master/CHANGELOG.md")
    if err != nil {
        log.Error(err)
        return "获取 changelog 失败"
    }
    defer func(Body io.ReadCloser) {
        err := Body.Close()
        if err != nil {
            log.Error(err)
        }
    }(resp.Body)
    // 解析响应
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        log.Error(err)
        return "获取 changelog 失败"
    }
    // 去除注释
    bodyString := string(body)
    bodyString = strings.Split(bodyString, "-->")[1]
    // 转换为 html
    var buf bytes.Buffer
    err = goldmark.Convert([]byte(bodyString), &buf)
    if err != nil {
        log.Error(err)
        return "获取 changelog 失败"
    }
    // 提取第一段 h2
    changeLogHtml := strings.ReplaceAll(strings.Split(buf.String(), "<h2>")[1], "</h2>", "")
    // 转换回 markdown
    converter := md.NewConverter("", true, nil)
    changeLog, err := converter.ConvertString(changeLogHtml)
    if err != nil {
        log.Error(err)
        return "获取 changelog 失败"
    }
    return changeLog
}

func GetReleasesInfo() ReleaseInfo {
    // 获取 release
    // GET 请求
    req, err := http.NewRequest("GET", "https://api.github.com/repos/misskey-dev/misskey/releases", nil)
    if err != nil {
        log.Error(err)
    }
    req.Header.Set("Accept", "application/vnd.github.v3+json")
    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ConfigInstance.GithubToken))
    req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
    // 发送 GET 请求
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        log.Error(err)
    }
    defer func(Body io.ReadCloser) {
        err := Body.Close()
        if err != nil {
            log.Error(err)
        }
    }(resp.Body)
    // 解析响应
    var releaseInfo ReleaseInfo
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        log.Error(err)
    }
    _ = json.Unmarshal(body, &releaseInfo)
    return releaseInfo
}
