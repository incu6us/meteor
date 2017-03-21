package main

import (
	"os"
	"bufio"
	"os/exec"
	"log"
	"net/http"
	"fmt"
	"bytes"
	"github.com/incu6us/meteor/internal/utils/config"
	"github.com/gorilla/mux"
	"path/filepath"
	"strings"
	"github.com/abbot/go-http-auth"
	"github.com/incu6us/meteor/internal/utils/passwd"
	"io/ioutil"
	"github.com/naoina/toml"
	"github.com/incu6us/meteor/internal/utils/httputils"
	"encoding/json"
	"errors"
	"time"
)

const (
	Workspace = "workspace"
	TaskDir   = "tasks"
)

var (
	conf        = config.GetConfig()
	APP_PATH, _ = os.Getwd()
	WORKSPACE   = APP_PATH + string(filepath.Separator) + Workspace
	TASK_DIR    = APP_PATH + string(filepath.Separator) + TaskDir
)


func httpSecret(user, realm string) string {
	if user == conf.General.Username {
		return passwd.GeneratePassword().GenApr1Password(conf.General.Password)
	}
	return ""
}

func main() {

	log.Printf("ROOT PATH: %s", APP_PATH)
	if conf.General.Username != "" {
		log.Printf("ROOT HEADER PASSWORD: %s", passwd.GeneratePassword().GetPasswdForHeader(
			conf.General.Username, conf.General.Password),
		)
	}

	routes := make(map[string]func(http.ResponseWriter, *http.Request))
	routes["/api/task/run/{taskName}"] = Run

	router := mux.NewRouter().StrictSlash(true)

	// curl example: -H 'Authorization: Basic dXNlcjpQQHNzdzByZA=='
	authenticator := auth.NewBasicAuthenticator("meteor", httpSecret)
	for k, v := range routes{
		router.HandleFunc(k, auth.JustCheck(authenticator, v))
	}

	log.Printf("Start listening on %s", conf.General.Listen)
	if err := http.ListenAndServe(conf.General.Listen, router); err != nil {
		log.Panicln(err)
	}
}

type SlackDataBody struct {
	Text string
}

func Run(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	var startExecutionCommandTime time.Time
	var endExecutionCommandTime time.Duration

	vars := mux.Vars(r)

	//var msg = make(chan string)

	defer r.Body.Close()

	taskName := vars["taskName"]

	mess, err := sendSlack(taskName, "Job `"+taskName+"` has been started!")
	if err != nil{
		log.Printf("Slack error: %v", err)
	}
	if mess != "" {
		log.Printf("Slack message: %s", mess)
	}
	startExecutionCommandTime = time.Now()
	result, err := executeTask(taskName)
	if err != nil {
		endExecutionCommandTime = time.Now().Sub(startExecutionCommandTime)
		sendSlack(taskName, ":skull: Job `"+taskName+"` - *failed*!\n" +
			"Result:\n```"+result+"\n"+err.Error()+"```\nExecution time: *"+endExecutionCommandTime.String()+"*")
		w.Write([]byte(err.Error()))
		return
	}

	//fmt.Println(new(TaskConfig).taskConfig(taskName).Vars)

	w.Write([]byte(result))

	endExecutionCommandTime = time.Now().Sub(startExecutionCommandTime)
	mess, err = sendSlack(taskName, ":+1: Job `"+taskName+"` has been finished *successfully*!\n" +
		"Result:\n```"+result+"```\nExecution time: *"+endExecutionCommandTime.String()+"*")
	if err != nil{
		log.Printf("Slack error: %v", err)
	}
	if mess != "" {
		log.Printf("Slack message: %s", mess)
	}
}

func sendSlack(taskName, result string) (string, error) {
	slackUrl := new(TaskConfig).taskConfig(taskName).Slack.Url
	if slackUrl != ""{
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
	return "", nil
}
type TaskConfig struct {
	Vars []struct{
		Name string
		Value string
	}
	Slack struct{
		Url string
	}
}

func (t *TaskConfig) taskConfig(taskName string) *TaskConfig {
	var err error
	var confFile []byte

	if confFile, err = ioutil.ReadFile(
		TASK_DIR + string(filepath.Separator) + taskName + string(filepath.Separator) + "config",
	); err != nil {
		log.Printf("Error to open script file: %v", err)
	}

	if err := toml.Unmarshal(confFile, t); err != nil {
		log.Printf("TOML error: %v", err)
	}

	return t
}

func executeTask(taskName string) (string, error) {
	taskWorkspace := WORKSPACE + string(filepath.Separator) + taskName

	var globalVars = make(map[string]string)
	globalVars["$WORKSPACE"] = taskWorkspace
	globalVars["$TASKSPACE"] = TASK_DIR+string(filepath.Separator)+taskName

	if exists(taskWorkspace) == true {
		log.Printf("Task is already running. Workspace: %s - is busy. Wait a while", taskWorkspace)
		return "", errors.New("Task is already running. Workspace: "+taskWorkspace+" - is busy. Wait a while...")
	}

	//var msg = make(chan string)
	var err error
	var scriptFile *os.File

	defer scriptFile.Close()

	if err = os.MkdirAll(taskWorkspace, 0777); err != nil {
		log.Println(err)
	}
	os.Chdir(taskWorkspace)

	executeCmd := func(cmdStr string) (string, error) {
		var cmdOut []byte

		// $WORKSPACE global var
		if strings.Contains(cmdStr, "$WORKSPACE") {
			cmdStr = strings.Replace(cmdStr, "$WORKSPACE", globalVars["$WORKSPACE"], -1)+string(filepath.Separator)
		}

		// $TASKSPACE global var
		if strings.Contains(cmdStr, "$TASKSPACE") {
			cmdStr = strings.Replace(cmdStr, "$TASKSPACE", globalVars["$TASKSPACE"], -1)+string(filepath.Separator)
		}

		cmd := exec.Command(conf.General.CmdInterpreter, conf.General.CmdFlag, cmdStr)

		if cmdOut, err = cmd.CombinedOutput(); err != nil {
			log.Printf("!!! Error to execute line: %v\n%s", err, cmdOut)
			//msg <- fmt.Sprintf("!!! Error to execute line: %v", err)
			return "", errors.New(fmt.Sprintf("!!! Error to execute line: %v\n%s", err, cmdOut))
		}

		log.Printf("--- Output: %s", cmdOut)
		//msg <- fmt.Sprintf("--- Output: %s", cmdOut)
		if cmdOut != nil {
			return fmt.Sprintf("--- Output: %s", cmdOut), nil
		}
		return "", nil
	}

	if scriptFile, err = os.Open(
		TASK_DIR + string(filepath.Separator) + taskName + string(filepath.Separator) + "pipeline",
	); err != nil {
		//msg <- fmt.Sprintf("Error to open script file: %v", err)
		return "", err
	}

	scanner := bufio.NewScanner(scriptFile)
	scanner.Split(bufio.ScanLines)

	var buf bytes.Buffer
	log.Printf("Running a script: %s", taskName)

	for scanner.Scan() {
		var output string
		str := scanner.Text()
		if str != "" && !strings.HasPrefix(str, "#") {
			log.Printf("--- Running: %s\n", str)
			//msg <- fmt.Sprintf("--- Running: %s\n", str)
			buf.WriteString(fmt.Sprintf("--- Running: %s\n", str))
			if output, err = executeCmd(str);err != nil{
				cleanTaskWorkspace(taskWorkspace)
				return buf.String(), err
			}
			if	output != "" {
				buf.WriteString(output + "\n\n")
			}
		}
	}

	cleanTaskWorkspace(taskWorkspace)

	log.Println()
	log.Println()

	return buf.String(), nil
}

func cleanTaskWorkspace(taskWorkspace string) {
	log.Printf("Cleaning: %s", taskWorkspace)
	if err := os.RemoveAll(taskWorkspace); err != nil {
		log.Printf("Error cleaning workdir: %v", err)
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
