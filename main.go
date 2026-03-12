package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

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
	}

	diff, isNeedAddCommand, err := getDiff(cmdConfig)
	if err != nil {
		log.Fatalln("获取差异信息失败，原因：", err)
	}

	pText := getPromptMain()

	var reqWaitGroup sync.WaitGroup
	reqWaitGroup.Add(cmdConfig.loop)
	dataChan := make(chan CommitChan, cmdConfig.loop)

	var messageList []string

	for i := 0; i < cmdConfig.loop; i++ {
		go func() {
			defer reqWaitGroup.Done()
			routineId := i
			commitMessage, err := sendReqCore(pText, diff, config, false)
			if err != nil {
				dataChan <- CommitChan{data: "", err: err, index: routineId}
			} else {
				dataChan <- CommitChan{data: commitMessage, err: nil, index: routineId}
			}
		}()
	}

	// 等待用 go routine
	go func() {
		reqWaitGroup.Wait()
		close(dataChan)
	}()
	routineIndex := 1
	for data := range dataChan {
		if data.err != nil {
			log.Println("好像哪儿出问题了：", data.err)
		} else {
			fmt.Println("完成", routineIndex, "/", cmdConfig.loop, "全部|新增：", data.data)
			messageList = append(messageList, data.data)
		}
		routineIndex++
	}
	println(isNeedAddCommand)

	msg, err := selectPrompt(messageList)

	err = callcmd(cmdConfig, msg, isNeedAddCommand)
	if err != nil {
		afterRemoteCallRollback(messageList)
		log.Fatalln("运行指令的过程中出现错误，详情：", err)
	}
}

func selectPrompt(list []string) (string, error) {
	const allNo = "都不行"
	selectPrompt := selection.New("请选择提示词", append(list, allNo))
	selectPrompt.PageSize = len(list)
	commitMsg, err := selectPrompt.RunPrompt()
	if err != nil {
		return "", err
	}
	if commitMsg == allNo {
		return "", fmt.Errorf("用户取消提交")
	}
	return commitMsg, nil
}

// 当调用 LLM 接口后程序后处理报错时回退
func afterRemoteCallRollback(msg []string) {
	readySaveString := strings.Join(msg, "\n")
	tmpFilePath := filepath.Join(os.TempDir(), "git-commit-latest.txt")
	err := os.WriteFile(tmpFilePath, []byte(readySaveString), 0666)
	if err != nil {
		log.Fatalln("[回退]失败，原因：", err, "\n遗言：\n", readySaveString)
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
