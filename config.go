package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func errorMessageBuild(message string) error {
	fmt.Println("提示：可以通过 .env 文件填写 API_KEY、BASE_URL、MODEL_NAME 三个参数")
	return fmt.Errorf("%s", message+"没有填写")
}

// 获取配置文件
func getConfigValue() (RemoteAPIConfig, error) {

	config := RemoteAPIConfig{}
	err := godotenv.Load()
	if err != nil {
		return config, err
	}
	config.API_KEY = os.Getenv("API_KEY")
	if config.API_KEY == "" {
		return config, errorMessageBuild("API_KEY")
	}
	config.BASE_URL = os.Getenv("BASE_URL")
	if config.BASE_URL == "" {
		return config, errorMessageBuild("BASE_URL")
	}
	config.MODEL_NAME = os.Getenv("MODEL_NAME")
	if config.MODEL_NAME == "" {
		return config, errorMessageBuild("MODEL_NAME")
	}
	return config, nil
}
