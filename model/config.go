package model

var Config ConfigModel

type ConfigModel struct {
	LiveDataDir string `yaml:"live_data_dir" envconfig:"LIVE_DATA_DIR"`
	LogLevel    int8   `yaml:"log_level" envconfig:"LOG_LEVEL"`
	// Server struct {
	// 	Port string `yaml:"port" envconfig:"SERVER_PORT"`
	// 	Host string `yaml:"host" envconfig:"SERVER_HOST"`
	// } `yaml:"server"`
	// Database struct {
	// 	Username string `yaml:"user" envconfig:"DB_USERNAME"`
	// 	Password string `yaml:"pass" envconfig:"DB_PASSWORD"`
	// } `yaml:"database"`
}
