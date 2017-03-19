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

func main() {

	log.Printf("ROOT PATH: %s", APP_PATH)

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/api/task/run/{taskName}", Run)

	log.Println("Start listening on 8080")
	if err := http.ListenAndServe(conf.General.Listen, router); err != nil {
		log.Panicln(err)
	} else {
		log.Printf("Start listening on %s", conf.General.Listen)
	}
}

func Run(w http.ResponseWriter, r *http.Request) {
	//w.Header().Set("Content-Type", "application/text")
	w.WriteHeader(http.StatusOK)

	vars := mux.Vars(r)

	//var msg = make(chan string)

	defer r.Body.Close()

	taskName := vars["taskName"]

	w.Write([]byte(executeTask(taskName)))

}

func executeTask(taskName string) string {
	taskWorkspace := WORKSPACE + string(filepath.Separator) + taskName

	var globalVars = make(map[string]string)
	globalVars["$WORKSPACE"] = taskWorkspace
	globalVars["$TASKSPACE"] = TASK_DIR+string(filepath.Separator)+taskName

	if exists(taskWorkspace) == true {
		log.Printf("Task is already running. Workspace: %s - is busy. Wait a while", taskWorkspace)
		return fmt.Sprintf("Task is already running. Workspace: %s - is busy. Wait a while", taskWorkspace)
	}

	//var msg = make(chan string)
	var err error
	var scriptFile *os.File

	defer scriptFile.Close()

	if err = os.MkdirAll(taskWorkspace, 0777); err != nil {
		log.Println(err)
	}
	os.Chdir(taskWorkspace)

	executeCmd := func(cmdStr string) interface{} {
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
			return fmt.Sprintf("!!! Error to execute line: %v\n%s", err, cmdOut)
		}

		log.Printf("--- Output: %s", cmdOut)
		//msg <- fmt.Sprintf("--- Output: %s", cmdOut)
		if cmdOut != nil {
			return fmt.Sprintf("--- Output: %s", cmdOut)
		}
		return nil
	}

	if scriptFile, err = os.Open(
		TASK_DIR + string(filepath.Separator) + taskName + string(filepath.Separator) + "script",
	); err != nil {
		//msg <- fmt.Sprintf("Error to open script file: %v", err)
		return fmt.Sprintf("Error to open script file: %v", err)
	}

	scanner := bufio.NewScanner(scriptFile)
	scanner.Split(bufio.ScanLines)

	var buf bytes.Buffer
	log.Printf("Running a script: %s", taskName)
	for scanner.Scan() {
		str := scanner.Text()
		if str != "" && !strings.HasPrefix(str, "#") {
			log.Printf("--- Running: %s\n", str)
			//msg <- fmt.Sprintf("--- Running: %s\n", str)
			buf.WriteString(fmt.Sprintf("--- Running: %s\n", str))
			if output := executeCmd(str); output != nil {
				buf.WriteString(output.(string) + "\n\n")
			}
		}
	}

	cleanTaskWorkspace(taskWorkspace)

	log.Println()
	log.Println()

	return buf.String()
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
		log.Println("1", err)
		return true
	}

	if os.IsNotExist(err) {
		log.Println("2", err)
		return false
	}

	return false
}
