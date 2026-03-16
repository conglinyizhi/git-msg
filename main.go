package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/erikgeiser/promptkit/confirmation"
	"github.com/erikgeiser/promptkit/selection"
)

const appName = "git-msg"

// 主函数
func main() {
	isFoundElementToString := func(isFound bool) string {
		if isFound {
			return "分数增加:"
		} else {
			return "新增条目:"
		}
	}
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
	dataChan := make(chan ChanResult[string], cmdConfig.loop)

	for i := 0; i < cmdConfig.loop; i++ {
		go func(routineId int) {
			defer reqWaitGroup.Done()
			commitMessage, err := sendReqCore(pText, diff, config, false)
			if err != nil {
				dataChan <- ChanResult[string]{data: "", err: err, index: routineId}
			} else {
				dataChan <- ChanResult[string]{data: commitMessage, err: nil, index: routineId}
			}
		}(i)
	}

	// 等待用 go routine
	go func() {
		reqWaitGroup.Wait()
		close(dataChan)
	}()
	routineIndex := 0
	var messageListScore []ScoreMsg
	lengthStringSize := len(strconv.Itoa(cmdConfig.loop))
	for data := range dataChan {
		routineIndex++
		if data.err != nil {
			log.Println("好像哪儿出问题了：", data.err)
			continue
		}
		// 如果找到相通条目，分数+1
		isFoundElement := false
		for index, el := range messageListScore {
			if el.msg == data.data {
				messageListScore[index].score++
				isFoundElement = true
				break
			}
		}
		// 如果没有找到相同项目，新建条目，分数=1
		if !isFoundElement {
			messageListScore = append(messageListScore, ScoreMsg{score: 1, msg: data.data})
		}
		indexLength := len(strconv.Itoa(routineIndex))
		nowIndexString := strings.Join([]string{strings.Repeat("0", lengthStringSize-indexLength), strconv.Itoa(routineIndex)}, "")
		fmt.Println("完成", nowIndexString, "/", cmdConfig.loop, "全部|", isFoundElementToString(isFoundElement), data.data)

	}

	var messageList []string
	// 卸载分数外壳，同时以分数排序
	slices.SortFunc(messageListScore, func(a, b ScoreMsg) int {
		return b.score - a.score
	})
	for _, obj := range messageListScore {
		messageList = append(messageList, obj.msg)
	}

	msg, err := selectPrompt(messageList)
	if err != nil {
		afterRemoteCallRollback(messageList)
		log.Fatalln("选择提交消息时出现了问题，详情：", err)
	}
	err = callcmd(cmdConfig, msg, isNeedAddCommand)
	if err != nil {
		afterRemoteCallRollback(messageList)
		log.Fatalln("运行指令的过程中出现错误，详情：", err)
	}
}

func selectPrompt(list []string) (string, error) {
	const allNo = "都不行"
	selectPrompt := selection.New("请选择提示词", append(list, allNo))
	selectPrompt.PageSize = len(list) + 1
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
