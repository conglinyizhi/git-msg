package main

import (
	"context"
	"embed"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/erikgeiser/promptkit/confirmation"
	"github.com/erikgeiser/promptkit/selection"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
)

const appName = "git-msg"

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

func sendReqCore(sys, user string, cfg RemoteAPIConfig) (string, error) {
	chatModel, err := openai.NewChatModel(context.Background(), &openai.ChatModelConfig{
		Model:   cfg.MODEL_NAME,
		APIKey:  cfg.API_KEY,
		BaseURL: cfg.BASE_URL,
	})
	if err != nil {
		log.Fatalln("创建 chat model 失败")
		return "", err
	}
	sr, err := chatModel.Stream(context.Background(), []*schema.Message{
		schema.SystemMessage(sys),
		schema.UserMessage(user),
	})
	if err != nil {
		log.Fatalln("创建 stream 失败")
		return "", err
	}
	return reportStream(sr)
}

func reportStream(sr *schema.StreamReader[*schema.Message]) (string, error) {
	defer sr.Close()
	var strBuff strings.Builder
	for {
		message, err := sr.Recv()
		if err == io.EOF { // 流式输出结束
			fmt.Println()
			return strBuff.String(), nil
		}
		if err != nil {
			fmt.Println()
			log.Fatalf("recv failed: %v", err)
			return strBuff.String(), err
		}
		fmt.Print(message.Content)
		strBuff.WriteString(message.Content)
	}
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
	const showGitStatusItem = "查看仓库状态"
	const exitSelectPromptItem = "退出"
	if isNeedAdd {
		selectPrompt := selection.New("本次操作使用暂存区外的文件差异，先添加到暂存区然后提交吗？", []string{"Yes", "No", showGitStatusItem, exitSelectPromptItem})
		selectPrompt.PageSize = 2
		for {
			spResult, err := selectPrompt.RunPrompt()
			if err != nil {
				log.Fatalln("交互命令出现异常")
				return err
			}
			if spResult == "Yes" {
				goAdd = true
				break
			}
			if spResult == exitSelectPromptItem {
				return fmt.Errorf("用户选择退出")
			}
			if spResult == showGitStatusItem {
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

	commitMessage, err := sendReqCore(getPromptMain(), diff, config)
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

func subcommand_Init() int {
	rootDir, err := getConfigRootDir("")
	if err != nil {
		log.Fatalln("定位配置目录失败：", err)
		return 1
	}
	initConfigDir(rootDir)
	initSkillDir(rootDir)
	return 0
}

func subCommand_Ping(cfg RemoteAPIConfig) int {
	if str, err := sendReqCore("test", "you must replay: OK.(DO NOT MORE TEXT)", cfg); err != nil {
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
	if err := os.MkdirAll(filepath.Join(rootDir, "skill"), 0755); err != nil {
		return err
	}
	skillFiles, err := skillFilesEmbed.ReadDir("skill")
	if err != nil {
		return err
	}
	for _, skill := range skillFiles {
		if skill.IsDir() {
			continue
		}
		data, err := skillFilesEmbed.ReadFile(filepath.Join("skill", skill.Name()))
		targetSkillFilePath := filepath.Join(rootDir, "skill", skill.Name())
		if err != nil {
			fmt.Println("提取" + targetSkillFilePath + "失败（读取文件出错），原因：" + err.Error())
			continue
		}
		if err := os.WriteFile(targetSkillFilePath, data, 0644); err != nil {
			fmt.Println("提取" + targetSkillFilePath + "失败（写入文件出错），原因：" + err.Error())
			continue
		}
		fmt.Println("成功提取" + targetSkillFilePath)
	}
	return nil
}
