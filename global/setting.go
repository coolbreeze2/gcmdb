package global

import (
	"goTool/pkg/setting"
	"log"
)

var (
	ClientSetting *setting.ClientSettingS
	ServerSetting *setting.ServerSettingS
)

func init() {
	err := setupSetting()
	if err != nil {
		log.Fatalf("init.setupSetting err: %v", err)
	}
}

func setupSetting() error {
	setting, err := setting.NewSetting()
	if err != nil {
		return err
	}
	if err = setting.ReadSection("Client", &ClientSetting); err != nil {
		return err
	}
	if err = setting.ReadSection("Server", &ServerSetting); err != nil {
		return err
	}
	return nil
}
