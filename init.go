package main

import (
	"embed"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"syscall"
)

//go:embed skill/*
var skillFilesEmbed embed.FS

func subcommand_Init() int {
	rootDir, err := getConfigRootDir("")
	if err != nil {
		log.Fatalln("定位配置目录失败：", err)
		return 1
	}
	errChan := make(chan error, 2)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		if err := initConfigDir(rootDir); err != nil {
			errChan <- fmt.Errorf("初始化配置目录失败：%w", err)
		}
	}()
	go func() {
		defer wg.Done()
		if err := initSkillDir(rootDir); err != nil {
			errChan <- fmt.Errorf("初始化技能目录失败：%w", err)
		}
	}()
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// 遍历通道，处理所有错误
	for err := range errChan {
		log.Println(err) // 可以根据需求选择是否终止程序
	}
	return 0
}

func subCommand_Ping(cfg Config) int {
	const testTitle = "测试"
	const testUser = "请返回且只返回OK"
	if str, err := sendReqCore(testTitle, testUser, cfg, false); err != nil {
		fmt.Fprintln(os.Stdout, "测试失败：", err)
		return 1
	} else {
		fmt.Fprintln(os.Stdout, "测试通过，LLM API 回复：", str)
		return 0
	}
}

func initConfigDir(rootDir string) error {
	if err := os.MkdirAll(filepath.Join(rootDir), 0755); err != nil {
		return err
	}
	if err := initNewTomlFile(RemoteAPIConfig{}); err != nil {
		return err
	}
	return nil
}

func initSkillDir(rootDir string) error {
	const skillPath = "skill"
	if err := os.MkdirAll(filepath.Join(rootDir, skillPath), 0755); err != nil {
		return err
	}
	skillFiles, err := skillFilesEmbed.ReadDir(skillPath)
	if err != nil {
		return err
	}
	for _, skill := range skillFiles {
		copyTarget := filepath.Join(rootDir, skillPath, skill.Name())
		_, err := os.Stat(copyTarget)
		if err == nil {
			log.Println("技能文件", copyTarget, "已经存在，略过")
			continue
		}
		if !errors.Is(err, syscall.ENOENT) {
			log.Println("无法确定", copyTarget, "是否存在，略过")
			continue
		}
		if skill.IsDir() {
			continue
		}
		data, err := skillFilesEmbed.ReadFile(filepath.Join("skill", skill.Name()))
		if err != nil {
			log.Println("提取" + copyTarget + "失败（读取文件出错），原因：" + err.Error())
			continue
		}
		if err := os.WriteFile(copyTarget, data, 0644); err != nil {
			log.Println("提取" + copyTarget + "失败（写入文件出错），原因：" + err.Error())
			continue
		}
		log.Println("成功提取" + copyTarget)
	}
	return nil
}
