package main

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/pelletier/go-toml/v2"
)

func errorMessageBuild(message string) error {
	fmt.Println("提示：可以通过 .env 文件填写 API_KEY、BASE_URL、MODEL_NAME 三个参数")
	return fmt.Errorf("%s", message+"没有填写")
}

// 获取系统规范的软件配置目录
func getSystemSoftwareConfigRootDir(fp string) (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	switch runtime.GOOS {
	case "windows":
		if shadowHome := os.Getenv("LOCALAPPDATA"); len(shadowHome) < 1 {
			return filepath.Join(usr.HomeDir, appName, "Config", fp), nil
		} else {
			return filepath.Join(shadowHome, appName, "Config", fp), nil
		}
	case "darwin":
		return filepath.Join(usr.HomeDir, "Library", "Preferences", appName, fp), nil
	}
	// Linux 和其他系统
	return filepath.Join(usr.HomeDir, ".config", appName, fp), nil
}

// 获取配置文件 - toml
func getConfigValue() (RemoteAPIConfig, error) {
	config := RemoteAPIConfig{}
	configPath, err := getSystemSoftwareConfigRootDir("llm.toml")
	if err != nil {
		return config, fmt.Errorf("定位配置文件路径错误:%w", err)
	}
	tomlConfigBody, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Printf("未能成功读取预期在 " + configPath + " 的配置文件，尝试读取环境变量……")
		initNewTomlFile(err, config)
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

// 判断错误是 syscall.ENOENT (文件不存在) 就尝试创建文件
func initNewTomlFile(err error, config RemoteAPIConfig) error {
	if !errors.Is(err, syscall.ENOENT) {
		return nil
	}
	configRootDir, err := getSystemSoftwareConfigRootDir("")
	if err != nil {
		return fmt.Errorf("定位配置文件路径错误:%w", err)
	}
	pathPerm := os.FileMode(0755)
	if err := os.MkdirAll(configRootDir, pathPerm); err != nil {
		return err
	}
	tomlFileBody, err := toml.Marshal(config)
	if err != nil {
		return err
	}
	configFilePath := filepath.Join(configRootDir, "llm.toml")

	filePerm := os.FileMode(0644)
	if err := os.WriteFile(configFilePath, tomlFileBody, filePerm); err != nil {
		return err
	}
	return nil
}

// 回退 - 使用系统变量
func tryReadEnv() (RemoteAPIConfig, error) {
	config := RemoteAPIConfig{}
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
