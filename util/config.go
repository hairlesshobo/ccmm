package util

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/hairlesshobo/go-import-media/model"
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"
)

func LoadConfig() {
	readFile(&model.Config)
	readEnv(&model.Config)

	fmt.Printf("%+v\n", model.Config)
}

func processError(err error) {
	slog.Error(err.Error())
	os.Exit(2)
}

func readFile(cfg *model.ConfigModel) {
	f, err := os.Open("config.yml")
	if err != nil {
		processError(err)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)
	if err != nil {
		processError(err)
	}
}

func readEnv(cfg *model.ConfigModel) {
	err := envconfig.Process("", cfg)
	if err != nil {
		processError(err)
	}
}
