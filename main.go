package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/abbot/go-http-auth"
	"github.com/gorilla/mux"

	"github.com/incu6us/meteor/internal/utils/config"
	"github.com/incu6us/meteor/internal/utils/passwd"
)

const (
	Workspace        = "workspace"
	TasksDir         = "tasks"
	COMMAND_TASKLIST = "/tasklist"
	COMMAND_TASKRUN  = "/taskrun"
)

var (
	conf        = config.GetConfig()
	APP_PATH, _ = os.Getwd()
	WORKSPACE   = APP_PATH + string(filepath.Separator) + Workspace
	TASKS_DIR   = APP_PATH + string(filepath.Separator) + TasksDir
)

func main() {

	log.Printf("ROOT PATH: %s", APP_PATH)
	if conf.General.Username != "" {
		log.Printf("ROOT HEADER PASSWORD: %s", passwd.GeneratePassword().GetPasswdForHeader(
			conf.General.Username, conf.General.Password),
		)
	}

	router := mux.NewRouter().StrictSlash(true)

	// curl example: -H 'Authorization: Basic dXNlcjpQQHNzdzByZA=='
	authenticator := auth.NewBasicAuthenticator("meteor", httpSecret)

	router.HandleFunc("/api/task/run/{taskName}", auth.JustCheck(authenticator, RunFunc))

	router.Handle("/api/integration/slack/list", SlackHandler(http.HandlerFunc(SlackListFunc)))
	router.Handle("/api/integration/slack/run", SlackHandler(http.HandlerFunc(nil)))

	log.Printf("Start listening on %s", conf.General.Listen)
	if err := http.ListenAndServe(conf.General.Listen, router); err != nil {
		log.Panicln(err)
	}
}

func exists(path string) (bool) {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}

	if os.IsNotExist(err) {
		return false
	}

	return false
}
