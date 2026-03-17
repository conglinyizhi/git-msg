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

	getCLIConfig := func() Config {
		ctxConfig := parseCommandLineExData()
		config, err := getConfigValue()
		if err != nil {
			log.Panic(err)
		}
		return Config{
			api: config,
			cmd: ctxConfig,
		}
	}
	ctxConfig := getCLIConfig()
	if ctxConfig.cmd.init {
		os.Exit(subcommand_Init())
	}
	if ctxConfig.cmd.ping {
		os.Exit(subCommand_Ping(ctxConfig))
	}
	diff, isNeedAddCommand, err := getDiff(ctxConfig.cmd)
	if err != nil {
		log.Fatalln("获取差异信息失败，原因：", err)
	}

	messageListScore := newFunction(ctxConfig, diff)

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
	err = callcmd(ctxConfig, msg, isNeedAddCommand)
	if err != nil {
		afterRemoteCallRollback(messageList)
		log.Fatalln("运行指令的过程中出现错误，详情：", err)
	}
}

func newFunction(ctxConfig Config, diff string) []ScoreMsg {
	isFoundElementToString := func(isFound bool, nowIndexString string, loop int, msg string) string {
		var str strings.Builder
		str.WriteString("完成")
		str.WriteString(nowIndexString)
		str.WriteString("/")
		str.WriteString(strconv.Itoa(loop))
		str.WriteString("全部|")
		if isFound {
			str.WriteString("分数增加:")
		} else {
			str.WriteString("新增条目:")
		}
		str.WriteString(msg)
		return str.String()
	}
	padLeadingZero := func(routineIndex int, lengthStringSize int) string {
		var str strings.Builder
		indexLength := len(strconv.Itoa(routineIndex))
		str.WriteString(strings.Repeat("0", lengthStringSize-indexLength))
		str.WriteString(strconv.Itoa(routineIndex))
		return str.String()
	}
	pText := getPromptMain()

	var reqWaitGroup sync.WaitGroup
	reqWaitGroup.Add(ctxConfig.cmd.loop)
	dataChan := make(chan ChanResult[string], ctxConfig.cmd.loop)

	for i := 0; i < ctxConfig.cmd.loop; i++ {
		go func(routineId int) {
			defer reqWaitGroup.Done()
			commitMessage, err := sendReqCore(pText, diff, ctxConfig, false)
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
	lengthStringSize := len(strconv.Itoa(ctxConfig.cmd.loop))
	for data := range dataChan {
		routineIndex++
		commitMessageText := data.data
		if data.err != nil {
			log.Println("好像哪儿出问题了：", data.err)
			continue
		}
		// 如果找到相通条目，分数+1
		isFoundElement := false
		for index, el := range messageListScore {
			if el.msg == commitMessageText {
				messageListScore[index].score++
				isFoundElement = true
				break
			}
		}
		// 如果没有找到相同项目，新建条目，分数=1
		if !isFoundElement {
			messageListScore = append(messageListScore, ScoreMsg{score: 1, msg: commitMessageText})
		}
		nowIndexString := padLeadingZero(routineIndex, lengthStringSize)
		fmt.Println(isFoundElementToString(isFoundElement, nowIndexString, ctxConfig.cmd.loop, commitMessageText))
	}
	return messageListScore
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

func callcmd(cfg Config, commitMessage string, isNeedAdd bool) error {
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
				cmdResult, err := getStatus(cfg.cmd)
				if err != nil {
					return err
				}
				printCommandOutput(cmdResult, "status -sb")
				fmt.Println("*咳咳*，所以……")
			}
		}
	}

	if goAdd {
		stdout, err := runGitAdd(cfg.cmd)
		if err != nil {
			return err
		}
		printCommandOutput(stdout, "add")
	}
	if goCommit {
		stdout, err := runGitCommit(cfg.cmd, commitMessage)
		if err != nil {
			return err
		}
		printCommandOutput(stdout, "commit -m")
	}
	return nil
}
