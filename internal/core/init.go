package core

import (
	"errors"
	"fmt"
	"gitmsg/embed"
	"gitmsg/internal/config"
	"gitmsg/internal/llm"
	"gitmsg/internal/types"
	"gitmsg/internal/utils"
	"log"
	"os"
	"path/filepath"
	"syscall"

	"golang.org/x/sync/errgroup"
)

func SubcommandInit() (int, error) {
	rootDir, err := utils.GetConfigRootDir("")
	if err != nil {
		return 1, fmt.Errorf("定位配置目录失败：%w", err)
	}
	var group errgroup.Group
	group.Go(func() error {
		return initConfigDir(rootDir)
	})
	group.Go(func() error {
		return initSkillDir(rootDir)
	})

	if err = group.Wait(); err != nil {
		return 1, err
	}
	return 0, nil
}

func subCommand_Ping(cfg types.Config) int {
	const testTitle = "测试"
	const testUser = "请返回且只返回OK"
	if str, err := llm.SendReqCore(testTitle, testUser, cfg, false); err != nil {
		fmt.Fprintln(os.Stdout, "测试失败：", err)
		return 1
	} else {
		fmt.Fprintln(os.Stdout, "测试通过，LLM API 回复：", str)
		return 0
	}
}

func initConfigDir(rootDir string) error {
	if err := os.MkdirAll(rootDir, 0755); err != nil {
		return fmt.Errorf("初始化环节，config(%s) 目录创建失败，原因：%w", rootDir, err)
	}
	if err := config.InitNewTomlFile(types.RemoteAPIConfig{}); err != nil {
		return fmt.Errorf("初始化环节，初始化 toml 文件失败，原因：%w", err)
	}
	return nil
}

func initSkillDir(rootDir string) error {
	const skillPath = "skill"
	if err := os.MkdirAll(filepath.Join(rootDir, skillPath), 0755); err != nil {
		return fmt.Errorf("初始化：Skill 目录创建失败，原因：%w", err)
	}
	skillFiles, err := embed.SkillFilesEmbed.ReadDir(skillPath)
	if err != nil {
		return fmt.Errorf("初始化：Skill 目录读取嵌入文件失败，原因：%w", err)
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
		data, err := embed.SkillFilesEmbed.ReadFile(filepath.Join("skill", skill.Name()))
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
