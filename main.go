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

	"github.com/erikgeiser/promptkit/confirmation"
	"github.com/erikgeiser/promptkit/selection"
)

const appName = "git-msg"

// 主函数
func main() {
	cmdConfig := parseCommandLineExData()
	if cmdConfig.init {
		os.Exit(subcommand_Init())
	}
	config, err := getConfigValue()
	if cmdConfig.ping {
		os.Exit(subCommand_Ping(config))
	}

	if err != nil {
		log.Fatalln("获取大模型配置信息失败：", err)
		os.Exit(1)
	}

	diff, isNeedAddCommand, err := getDiff(cmdConfig)
	if err != nil {
		log.Fatalln("获取差异信息失败，原因：", err)
		os.Exit(1)
	}

	commitMessage, err := sendReqCore(getPromptMain(), diff, config, true)
	if err != nil {
		log.Fatalln("调用远程大模型失败，原因：", err)
		os.Exit(1)
	}
	err = callcmd(cmdConfig, commitMessage, isNeedAddCommand)
	if err != nil {
		log.Fatalln("运行指令的过程中出现错误，详情：", err)
		afterRemoteCallRollback(commitMessage)
		os.Exit(1)
	}
}

// 当调用 LLM 接口后程序后处理报错时回退
func afterRemoteCallRollback(msg string) {
	tmpFilePath := filepath.Join(os.TempDir(), "git-commit-latest.txt")
	err := os.WriteFile(tmpFilePath, []byte(msg), 0666)
	if err != nil {
		log.Fatalln("[回退]失败，原因：", err)
		return
	}
	fmt.Println("[回退]大模型输出结果将保存到", tmpFilePath)
}

func callcmd(cmd CommandlineConfig, commitMessage string, isNeedAdd bool) error {
	// 询问用户是否提交，如果需要，则提交
	goCommit, err := confirmation.New("一切准备就绪，发起提交吗?", confirmation.Yes).RunPrompt()
	if err != nil {
		log.Fatalln("交互命令出现异常")
		return err
	}
	if !goCommit {
		return fmt.Errorf("用户取消提交")
	}
	// 是否需要添加文件到暂存区
	var goAdd = false
	const YesItem = "是(Yes)"
	const NoItem = "否(No)"
	const showGitStatusItem = "查看仓库状态"
	const exitSelectPromptItem = "退出"
	if isNeedAdd {
		selectPrompt := selection.New("本次操作使用暂存区外的文件差异，先添加到暂存区然后提交吗？", []string{YesItem, NoItem, showGitStatusItem, exitSelectPromptItem})
		selectPrompt.PageSize = 2
	loop:
		for {
			spResult, err := selectPrompt.RunPrompt()
			if err != nil {
				log.Fatalln("交互命令出现异常")
				return err
			}
			switch spResult {
			case YesItem:
				goAdd = true
				break loop
			case NoItem:
				goAdd = false
				break loop
			case exitSelectPromptItem:
				return fmt.Errorf("用户选择退出")
			case showGitStatusItem:
				cmdResult, err := getStatus(cmd)
				if err != nil {
					return err
				}
				printCommandOutput(cmdResult, "status -sb")
				fmt.Println("*咳咳*，所以……")
			}
		}
	}

	if goAdd {
		stdout, err := runGitAdd(cmd)
		if err != nil {
			return err
		}
		printCommandOutput(stdout, "add")
	}
	if goCommit {
		stdout, err := runGitCommit(cmd, commitMessage)
		if err != nil {
			return err
		}
		printCommandOutput(stdout, "commit -m")
	}
	return nil
}

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

func subCommand_Ping(cfg RemoteAPIConfig) int {
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

//go:embed skill/*
var skillFilesEmbed embed.FS

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
