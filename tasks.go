package main

import (
	"bufio"
	"bytes"
	"log"
	"strings"
	"fmt"
	"os"
	"io/ioutil"
	"path/filepath"
	"github.com/naoina/toml"
	"os/exec"
	"net/http"
	"time"
	"errors"
)

type TaskConfig struct {
	Vars []struct {
		Name  string
		Value string
	}
	Slack struct {
		WebHookUrl string `toml:"webhook-url"`
	}
}

func (t *TaskConfig) taskConfig(taskName string) *TaskConfig {
	var err error
	var confFile []byte

	if confFile, err = ioutil.ReadFile(
		TASKS_DIR + string(filepath.Separator) + taskName + string(filepath.Separator) + "config",
	); err != nil {
		log.Printf("Error to open script file: %v", err)
	}

	toml.Unmarshal(confFile, t)
	//if err := toml.Unmarshal(confFile, t); err != nil {
	//	log.Printf("TOML error: %v", err)
	//}

	return t
}

func executeTask(taskName string, params map[string]string) (string, error) {
	taskWorkspace := WORKSPACE + string(filepath.Separator) + taskName

	var globalVars = make(map[string]string)
	globalVars["$WORKSPACE"] = taskWorkspace
	globalVars["$TASKSPACE"] = TASKS_DIR + string(filepath.Separator) + taskName
	for k, v := range params{
		globalVars[fmt.Sprintf("$%s", strings.Replace(k, " ", "", -1))] = v
	}

	if exists(taskWorkspace) == true {
		log.Printf("Task is already running. Workspace: %s - is busy. Wait a while", taskWorkspace)
		return "", errors.New("Task is already running. Workspace: " + taskWorkspace + " - is busy. Wait a while...")
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
			cmdStr = strings.Replace(cmdStr, "$WORKSPACE", globalVars["$WORKSPACE"], -1) + string(filepath.Separator)
		}

		// $TASKSPACE global var
		if strings.Contains(cmdStr, "$TASKSPACE") {
			cmdStr = strings.Replace(cmdStr, "$TASKSPACE", globalVars["$TASKSPACE"], -1) + string(filepath.Separator)
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
		TASKS_DIR + string(filepath.Separator) + taskName + string(filepath.Separator) + "pipeline",
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
			if output, err = executeCmd(str); err != nil {
				cleanTaskWorkspace(taskWorkspace)
				return buf.String(), err
			}
			if output != "" {
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

func executeHttpTask(w http.ResponseWriter, taskName string, params map[string]string, responseUrl string) {
	var startExecutionCommandTime time.Time
	var endExecutionCommandTime time.Duration

	mess, err := sendSlack(responseUrl, taskName, "Job `"+taskName+"` has been started!")
	if err != nil {
		log.Printf("Slack error: %v", err)
	}
	if mess != "" {
		log.Printf("Slack message: %s", mess)
	}
	startExecutionCommandTime = time.Now()
	result, err := executeTask(taskName, params)
	if err != nil {
		endExecutionCommandTime = time.Now().Sub(startExecutionCommandTime)
		sendSlack(responseUrl, taskName, ":skull: Job `"+taskName+"` - *failed*!\n"+"Result:\n```"+result+"\n"+err.Error()+"```\nExecution time: *"+endExecutionCommandTime.String()+"*")
		w.Write([]byte(err.Error()))
		return
	}

	//fmt.Println(new(TaskConfig).taskConfig(taskName).Vars)

	endExecutionCommandTime = time.Now().Sub(startExecutionCommandTime)
	mess, err = sendSlack(responseUrl, taskName, ":+1: Job `"+taskName+"` has been finished *successfully*!\n"+"Result:\n```"+result+"```\nExecution time: *"+endExecutionCommandTime.String()+"*")
	if err != nil {
		log.Printf("Slack error: %v", err)
	}
	if mess != "" {
		log.Printf("Slack message: %s", mess)
	}

	w.Write([]byte(result))
}
