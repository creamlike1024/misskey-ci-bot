package main

import (
    "bytes"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "net/http"

    log "github.com/sirupsen/logrus"
)

type WorkflowRuns []struct {
    Id         int64  `json:"id"`
    Status     string `json:"status"`
    Conclusion string `json:"conclusion"`
    HtmlUrl    string `json:"html_url"`
}

func GetGithubActionRunStatus() WorkflowRuns {
    // GET 请求
    url := fmt.Sprintf("https://api.github.com/repos/%s/%s/actions/runs", ConfigInstance.GithubUsername, ConfigInstance.ActionRepo)
    req, err := http.NewRequest("GET", url, nil)
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
    // 取出 workflow_runs
    var bodyMap map[string]interface{}
    _ = json.NewDecoder(resp.Body).Decode(&bodyMap)
    workflowRunsBody, _ := json.Marshal(bodyMap["workflow_runs"])
    // 解析响应
    var workflowRuns WorkflowRuns
    _ = json.Unmarshal(workflowRunsBody, &workflowRuns)
    return workflowRuns
}

func RunGithubAction() bool {
    // POST 请求
    url := fmt.Sprintf("https://api.github.com/repos/%s/%s/actions/workflows/%s/dispatches", ConfigInstance.GithubUsername, ConfigInstance.ActionRepo, "build.yaml")
    // 设置 POST 请求的 body
    data := map[string]string{
        "ref": "main",
    }
    jsonStr, _ := json.Marshal(data)
    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
    if err != nil {
        log.Error(err)
    }
    req.Header.Set("Accept", "application/vnd.github.v3+json")
    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ConfigInstance.GithubToken))
    req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
    // 发送 POST 请求
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
    if resp.StatusCode == 204 {
        return true
    } else {
        return false
    }
}

func DeleteGithubActionRun(Ids []int64) []error {
    deleteRun := func(id int64) error {
        // DELETE 请求
        url := fmt.Sprintf("https://api.github.com/repos/%s/%s/actions/runs/%d", ConfigInstance.GithubUsername, ConfigInstance.ActionRepo, id)
        req, err := http.NewRequest("DELETE", url, nil)
        if err != nil {
            return err
        }
        req.Header.Set("Accept", "application/vnd.github.v3+json")
        req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ConfigInstance.GithubToken))
        req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
        // 发送 DELETE 请求
        client := &http.Client{}
        resp, err := client.Do(req)
        if err != nil {
            return err
        }
        defer func(Body io.ReadCloser) {
            err := Body.Close()
            if err != nil {
                log.Error(err)
            }
        }(resp.Body)
        // 解析响应
        if resp.StatusCode == 204 {
            return nil
        } else {
            return errors.New("delete failed")
        }
    }
    var errs []error
    for _, id := range Ids {
        err := deleteRun(id)
        if err != nil {
            errs = append(errs, err)
        }
    }
    if len(errs) > 0 {
        return errs
    }
    return nil
}
