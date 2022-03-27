package settings

import (
	"fmt"

	"github.com/spf13/viper"
)

// 全局配置
var Config = new(MysqlConfig)

type MysqlConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Dbname   string `mapstructure:"dbname"`
}

func Init(configFile string) (err error) {
	// 配置文件名称
	viper.SetConfigFile(configFile)
	// 配置文件路径
	viper.AddConfigPath(".")
	// 读取文件信息
	err = viper.ReadInConfig()
	if err != nil {
		fmt.Println("读取配置文件失败")
		return err
	}
	//把读取到的信息反序列化到 Conf 变量中
	if err := viper.Unmarshal(Config); err != nil {
		fmt.Printf("viper.Unmarshal failed: %v\n", err)
	}
	return nil
}
