package main

import (
	"fmt"
	"github.com/gookit/color"
	"time"

	"github.com/imroc/req/v3"
)

func (r *Tagger) GetRecaptcha() (string, error) {
	type task struct {
		ErrorId          int    `json:"errorId"`
		ErrorCode        string `json:"errorCode"`
		ErrorDescription string `json:"errorDescription"`
		TaskId           string `json:"taskId"`
	}
	type result struct {
		ErrorId          int    `json:"errorId"`
		ErrorCode        string `json:"errorCode"`
		ErrorDescription string `json:"errorDescription"`
		Solution         struct {
			GRecaptchaResponse string `json:"gRecaptchaResponse"`
		} `json:"solution"`
		Status string `json:"status"`
	}

	color.Yellowln("开始解决验证码...")

	client := req.C()
	client.ImpersonateChrome()
	client.SetTimeout(2 * time.Minute)
	client.SetCommonRetryCount(2)
	client.SetBaseURL("https://api.yescaptcha.com")

	var taskResp task
	_, err := client.R().SetBodyJsonMarshal(map[string]any{
		"clientKey": r.yescaptchaToken,
		"task": map[string]any{
			"websiteURL": "https://login.wordpress.org/",
			"websiteKey": "6LckXrgUAAAAANrzcMN7iy_WxvmMcseaaRW-YFts",
			"pageAction": "login",
			"type":       "RecaptchaV3TaskProxylessM1S7",
		},
	}).SetSuccessResult(&taskResp).SetErrorResult(&taskResp).Post("/createTask")
	if err != nil {
		return "", err
	}
	if taskResp.ErrorId != 0 {
		return "", fmt.Errorf("创建验证码任务失败: %s", taskResp.ErrorDescription)
	}

	color.Greenln("验证码任务创建成功，任务 ID: " + taskResp.TaskId)

	var resp result
	timeout := time.After(120 * time.Second)
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			color.Yellowln("获取验证码结果...")

			_, err = client.R().SetBodyJsonMarshal(map[string]any{
				"clientKey": r.yescaptchaToken,
				"taskId":    taskResp.TaskId,
			}).SetSuccessResult(&resp).SetErrorResult(&resp).Post("/getTaskResult")
			if err != nil {
				return "", err
			}
			if resp.ErrorId != 0 && resp.Status != "processing" {
				return "", fmt.Errorf("获取验证码结果失败: %s", resp.ErrorDescription)
			}
			if resp.Status == "ready" {
				color.Greenln("验证码解决成功！")
				return resp.Solution.GRecaptchaResponse, nil
			}
		case <-timeout:
			return "", fmt.Errorf("获取验证码结果超时")
		}
	}
}
