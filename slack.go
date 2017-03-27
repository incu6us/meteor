package main

import (
	"encoding/json"
	"bytes"
	"github.com/incu6us/meteor/internal/utils/httputils"
)

func sendSlack(slackUrl, taskName, result string) (string, error) {
	var slackWebHookUrl string

	if slackUrl != "" {
		return slackMessage(slackUrl, result)
	}

	slackWebHookUrl = new(TaskConfig).taskConfig(taskName).Slack.WebHookUrl

	if slackWebHookUrl != "" {
		return slackMessage(slackWebHookUrl, result)
	}
	return "", nil
}

func slackMessage(slackUrl, result string) (string, error) {
	payload := make(map[string]string)
	payload["text"] = result
	jsonPaylod, _ := json.Marshal(payload)

	header := make(map[string]string)
	header["Content-type"] = "application/json"

	body := bytes.NewBuffer(jsonPaylod)
	resp, err := httputils.NewHTTPUtil().PostData(slackUrl, header, body, nil)
	if err != nil {
		return "", err
	}
	return string(resp), nil
}
