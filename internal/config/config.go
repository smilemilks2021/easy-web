package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type CaptureHeader struct {
	Header    string   `mapstructure:"header"`
	CacheKeys []string `mapstructure:"cache_keys"`
}

type MultiStepExtract struct {
	Source   string `mapstructure:"source"`   // cookie|header|json_response|localStorage
	Key      string `mapstructure:"path"`
	Variable string `mapstructure:"variable"`
	Final    bool   `mapstructure:"final"`
}

type MultiStepStep struct {
	ID      string             `mapstructure:"id"`
	Type    string             `mapstructure:"type"` // browser_capture|http_request
	URL     string             `mapstructure:"url"`
	Method  string             `mapstructure:"method"`
	Headers map[string]string  `mapstructure:"headers"`
	Extract []MultiStepExtract `mapstructure:"extract"`
}

type MultiStepAuth struct {
	Description  string          `mapstructure:"description"`
	URLPattern   string          `mapstructure:"url_pattern"`
	OutputHeader string          `mapstructure:"output_header"`
	Steps        []MultiStepStep `mapstructure:"steps"`
}

type Config struct {
	Mode           string `mapstructure:"mode"`
	Port           int    `mapstructure:"port"`
	DebugPort      int    `mapstructure:"debug_port"`
	AutoClose      bool   `mapstructure:"auto_close"`
	NoReuseProfile bool   `mapstructure:"no_reuse_profile"`
	Domains        struct {
		JWTCheck         []string `mapstructure:"jwt_check"`
		JWTCookies       []string `mapstructure:"jwt_cookies"`
		LocalStorageKeys []string `mapstructure:"localStorage_keys"`
	} `mapstructure:"domains"`
	CaptureHeaders map[string]CaptureHeader `mapstructure:"capture_headers"`
	MultiStepAuth  map[string]MultiStepAuth `mapstructure:"multi_step_auth"`
}

var C Config

func Init() {
	viper.SetConfigName(".easy-web")
	viper.SetConfigType("yaml")
	home, _ := os.UserHomeDir()
	viper.AddConfigPath(home)
	viper.SetDefault("mode", "auto")
	viper.SetDefault("port", 8080)
	viper.SetDefault("debug_port", 9222)
	viper.SetDefault("auto_close", true)
	_ = viper.ReadInConfig()
	_ = viper.Unmarshal(&C)
}

func CacheDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".easy-web", "cookies")
}

func ChromiumDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".easy-web", "chromium")
}

func ProfileDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".easy-web", "chrome-data")
}
