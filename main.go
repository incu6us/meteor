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
)

var conf = config.GetConfig()

func main() {

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/api/task/run/{taskName}", Run)

	log.Println("Start listening on 8080")
	log.Fatal(http.ListenAndServe(conf.General.Listen, router))
}

func Run(w http.ResponseWriter, r *http.Request){
	//w.Header().Set("Content-Type", "application/text")
	w.WriteHeader(http.StatusOK)

	vars := mux.Vars(r)

	//var msg = make(chan string)

	defer r.Body.Close()

	taskName := vars["taskName"]

	//go func() {

		w.Write([]byte(executeTask(taskName)))
	//}()

}

func executeTask(task_id string) string {
	//var msg = make(chan string)
	var err error
	var scriptFile *os.File

	defer scriptFile.Close()

	executeCmd := func(cmdStr string) interface{} {
		var cmdOut []byte

		cmd := exec.Command(conf.General.CmdInterpreter, conf.General.CmdFlag, cmdStr)

		if cmdOut, err = cmd.Output(); err != nil {
			log.Printf("!!! Error to execute line: %v", err)
			//msg <- fmt.Sprintf("!!! Error to execute line: %v", err)
			return fmt.Sprintf("!!! Error to execute line: %v", err)
		}

		log.Printf("--- Output: %s", cmdOut)
		//msg <- fmt.Sprintf("--- Output: %s", cmdOut)
		if cmdOut != nil{
			return fmt.Sprintf("--- Output: %s", cmdOut)
		}
		return nil
	}

	if scriptFile, err = os.Open("./"+conf.General.TaskDir+"/" + task_id + "/script.sh"); err != nil {
		//msg <- fmt.Sprintf("Error to open script file: %v", err)
		return fmt.Sprintf("Error to open script file: %v", err)
	}

	scanner := bufio.NewScanner(scriptFile)
	scanner.Split(bufio.ScanLines)

	var buf bytes.Buffer
	for scanner.Scan() {
		str := scanner.Text()
		if str != ""{
			log.Printf("--- Running: %s\n", str)
			//msg <- fmt.Sprintf("--- Running: %s\n", str)
			buf.WriteString(fmt.Sprintf("--- Running: %s\n", str))
			if output := executeCmd(str); output != nil {
				buf.WriteString(output.(string)+"\n\n")
			}
		}
	}

	return buf.String()
}
