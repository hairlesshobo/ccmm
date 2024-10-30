package model

var Config ConfigModel

type ConfigModel struct {
	LiveDataDir           string `yaml:"live_data_dir" envconfig:"LIVE_DATA_DIR"`
	LogLevel              int8   `yaml:"log_level" envconfig:"LOG_LEVEL"`
	UseSudo               string `yaml:"use_sudo" envconfig:"USE_SUDO"`
	ListenAddress         string `yaml:"listen_address" envconfig:"LISTEN_ADDRESS"`
	ListenPort            int32  `yaml:"listen_port" envconfig:"LISTEN_PORT"`
	ForceDryRun           bool   `yaml:"force_dry_run" envconfig:"FORCE_DRY_RUN"`
	DisableAutoProcessing bool   `yaml:"disable_auto_processing" envconfig:"DISABLE_AUTO_PROCESSING"`
	// Server struct {
	// 	Port string `yaml:"port" envconfig:"SERVER_PORT"`
	// 	Host string `yaml:"host" envconfig:"SERVER_HOST"`
	// } `yaml:"server"`
	// Database struct {
	// 	Username string `yaml:"user" envconfig:"DB_USERNAME"`
	// 	Password string `yaml:"pass" envconfig:"DB_PASSWORD"`
	// } `yaml:"database"`
}
