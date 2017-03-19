package config

import (
	"flag"
	"io/ioutil"
	"os"

	"log"
	"github.com/naoina/toml"
)

// TomlConfig - config structure
type TomlConfig struct {
	General struct{
		Listen string
		Username string
		Password string
		CmdInterpreter string `toml:"cmd-interpreter"`
		CmdFlag string `toml:"cmd-flag"`
	}
}

var config = new(TomlConfig).readConfig()
var logger = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.LUTC|log.Lshortfile)

// GetConfig - get current TOML-config
func GetConfig() *TomlConfig {
	return config
}

func (t *TomlConfig) readConfig() *TomlConfig {
	confFlag := t.readFlags()
	f, err := os.Open(*confFlag)
	if err != nil {
		logger.Printf("Can't open file: %v", err)
	}

	defer f.Close()

	buf, err := ioutil.ReadAll(f)
	if err != nil {
		logger.Printf("IO read error: %v", err)
	}

	if err := toml.Unmarshal(buf, t); err != nil {
		logger.Printf("TOML error: %v", err)
	}

	return t
}

func (t *TomlConfig) readFlags() (confFlag *string) {
	confFlag = flag.String("conf", "meteor.conf", "meteor.conf")
	flag.Parse()
	return
}