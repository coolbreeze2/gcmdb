package setting

import (
	"os"
	"runtime"

	"github.com/spf13/viper"
)

type Setting struct {
	vp *viper.Viper
}

func NewSetting() (*Setting, error) {
	configDirPath := userHomeDir() + string(os.PathSeparator) + ".cmdb"
	cofigName := "config"
	configType := "yaml"
	configFilePath := configDirPath + string(os.PathSeparator) + cofigName + "." + configType
	if err := makeDefaultConfigDir(configDirPath); err != nil {
		return nil, err
	}
	if err := genDefaultConfigFile(configFilePath); err != nil {
		return nil, err
	}
	vp := viper.New()
	vp.SetConfigName(cofigName)
	vp.AddConfigPath(configDirPath)
	vp.SetConfigType(configType)
	if err := vp.ReadInConfig(); err != nil {
		return nil, err
	}

	// 环境变量优先级高于配置文件
	viper.AutomaticEnv()

	return &Setting{vp}, nil
}

func makeDefaultConfigDir(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.Mkdir(path, os.ModeDir)
	}
	return nil
}

func genDefaultConfigFile(path string) error {
	defaultConfig := []byte(`CLIENT:
  CMDB_API_URL: http://127.0.0.1:8080/api/v1`)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.WriteFile(path, defaultConfig, 0644)
	}
	return nil
}

func userHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}
