package util

import (
	"log/slog"
	"os"
	"path"
	"path/filepath"

	"github.com/hairlesshobo/go-import-media/model"
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"
)

func LoadConfig() {
	readFile(&model.Config)
	readEnv(&model.Config)
}

func processError(err error) {
	slog.Error(err.Error())
	os.Exit(2)
}

func readFile(cfg *model.ConfigModel) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	configPath := os.Getenv("CONFIG_FILE")

	if configPath == "" {
		binPath, _ := os.Executable()
		binDir := filepath.Dir(binPath)
		sidecarPath := path.Join(binDir, "config.yml")

		if FileExists(sidecarPath) {
			configPath = sidecarPath
		} else {
			homeDir, _ := os.UserHomeDir()
			homeDotConfigPath := path.Join(homeDir, ".config", "go-import-media.yml")

			if FileExists(homeDotConfigPath) {
				configPath = homeDotConfigPath
			}
		}
	}

	if configPath == "" {
		logger.Error("No config file found")
		os.Exit(1)
	}

	if !FileExists(configPath) {
		logger.Error("The specified config file does not exist: " + configPath)
		os.Exit(1)
	}

	logger.Info("Reading config from " + configPath)

	f, err := os.Open(configPath)
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
