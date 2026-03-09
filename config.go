package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/pelletier/go-toml/v2"
)

func errorMessageBuild(message string) error {
	fmt.Println("提示：可以通过 .env 文件填写 API_KEY、BASE_URL、MODEL_NAME 三个参数")
	return fmt.Errorf("%s", message+"没有填写")
}

// 获取配置文件 - toml
func getConfigValue() (RemoteAPIConfig, error) {
	config := RemoteAPIConfig{}
	configPath, err := getConfigRootDir("llm.toml")
	if err != nil {
		return config, fmt.Errorf("定位配置文件路径错误:%w", err)
	}
	tomlConfigBody, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Println("未能成功读取预期在 " + configPath + " 的配置文件，尝试读取环境变量……")
		initNewTomlFileIfNeed(err, config)
		return tryReadEnv()
	}
	err = toml.Unmarshal(tomlConfigBody, &config)
	if err != nil {
		return config, err
	}
	configCheckResult := checkValue(config)
	if configCheckResult != nil {
		return config, configCheckResult
	}
	return config, nil
}

func initNewTomlFile(config RemoteAPIConfig) error {
	configRootDir, err := getConfigRootDir("")
	if err != nil {
		log.Fatalln("定位配置文件路径错误")
		return err
	}
	configFilePath := filepath.Join(configRootDir, "llm.toml")
	_, err = os.Stat(configFilePath)
	if err == nil {
		// 文件存在，提示并返回
		log.Printf("配置文件已经存在，不做任何修改: %s", configFilePath)
		return nil
	}
	if !errors.Is(err, syscall.ENOENT) {
		log.Fatalln("检查配置文件存在性失败", err)
		return err
	}
	pathPerm := os.FileMode(0755)
	if err := os.MkdirAll(configRootDir, pathPerm); err != nil {
		return err
	}
	tomlFileBody, err := toml.Marshal(config)
	if err != nil {
		return err
	}
	filePerm := os.FileMode(0644)
	if err := os.WriteFile(configFilePath, tomlFileBody, filePerm); err != nil {
		return err
	}
	fmt.Println("配置文件完成初始化，请打开并填写文件 ", configFilePath)
	return nil
}

// 判断错误是 syscall.ENOENT (文件不存在) 就尝试创建文件
func initNewTomlFileIfNeed(err error, cfg RemoteAPIConfig) error {
	if !errors.Is(err, syscall.ENOENT) {
		return nil
	}
	return initNewTomlFile(cfg)
}

// 回退 - 使用系统变量
func tryReadEnv() (RemoteAPIConfig, error) {
	config := RemoteAPIConfig{}
	if err := godotenv.Load(); err != nil {
		return config, err
	}
	config.API_KEY = os.Getenv("API_KEY")
	config.MODEL_NAME = os.Getenv("MODEL_NAME")
	config.BASE_URL = os.Getenv("BASE_URL")
	configCheckResult := checkValue(config)
	if configCheckResult != nil {
		return config, configCheckResult
	}
	return config, nil
}

// 检查必填数据
func checkValue(cfg RemoteAPIConfig) error {
	if cfg.API_KEY == "" {
		return errorMessageBuild("API_KEY")
	}
	if cfg.BASE_URL == "" {
		return errorMessageBuild("BASE_URL")
	}
	if cfg.MODEL_NAME == "" {
		return errorMessageBuild("MODEL_NAME")
	}
	return nil
}
