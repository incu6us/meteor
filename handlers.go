package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/incu6us/meteor/internal/utils/passwd"
)

func httpSecret(user, realm string) string {
	if user == conf.General.Username {
		return passwd.GeneratePassword().GenApr1Password(conf.General.Password)
	}
	return ""
}

func SlackHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var data url.Values
		var err error

		byteData, _ := ioutil.ReadAll(r.Body)
		log.Printf("Debug from Slack: %s", byteData)

		if data, err = url.ParseQuery(string(byteData)); err != nil {
			log.Printf("Error to parse string from Slack: %v", err)
		}

		token := data.Get("token")
		dirtyTask := data.Get("text")
		command := data.Get("command")
		responseUrl := data.Get("response_url")

		taskName := strings.Split(dirtyTask, " ")[0]
		params := make(map[string]string, 5)
		if len(strings.Split(dirtyTask, " ")) > 1 {
			for _, paramPairs := range strings.Split(strings.Split(dirtyTask, " ")[1], ";") {
				kv := strings.Split(paramPairs, "=")
				params[strings.TrimSpace(kv[0])] = kv[1]
			}
		}

		log.Printf("Slack params: %#v", params)

		if token == conf.General.SlackToken {
			switch command {
			case COMMAND_TASKLIST:
				h.ServeHTTP(w, r)
			case COMMAND_TASKRUN:
				go executeHttpTask(w, taskName, params, responseUrl)
				sendSlack(responseUrl, "", "Task was succefully queued!")
			}
		} else {
			io.WriteString(w, fmt.Sprintf("Wrong slack-token accepted: %s", token))
		}
	})
}

func SlackListFunc(w http.ResponseWriter, r *http.Request) {

	var listOfTasks bytes.Buffer
	var files []os.FileInfo
	var err error

	listOfTasks.WriteString("Tasks list:\n")

	if files, err = ioutil.ReadDir(TASKS_DIR); err != nil {
		log.Println(err)
		listOfTasks.WriteString("`empty`")
		w.Write(listOfTasks.Bytes())
		return
	}

	for _, file := range files {
		listOfTasks.WriteString("\t`" + file.Name() + "`\n")
	}

	w.Write(listOfTasks.Bytes())
}

func RunFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	vars := mux.Vars(r)

	//var msg = make(chan string)

	taskName := vars["taskName"]
	log.Printf("> PARAMS: %#v", r.URL.Query())
	params := make(map[string]string, 5)
	for k, values := range r.URL.Query() {
		if len(values) > 0 {
			params[k] = values[0]
		}
	}

	executeHttpTask(w, taskName, params, "")
}
