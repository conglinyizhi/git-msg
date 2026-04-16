package core

import (
	"fmt"
	"gitmsg/internal/config"
	"gitmsg/internal/git"
	"gitmsg/internal/llm"
	"gitmsg/internal/tui"
	"gitmsg/internal/types"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/google/uuid"
)

// 主函数
func CommitMain(cmd *types.CommandlineConfig) {

	getCLIConfig := func() (types.Config, error) {
		config, err := config.GetConfigValue()
		if err != nil {
			return types.Config{}, err
		}
		return types.Config{
			Api: config,
			Cmd: *cmd,
		}, nil
	}
	ctxConfig, err := getCLIConfig()
	if err != nil {
		log.Panic(err)
	}
	if ctxConfig.Cmd.Init {
		exitCode, err := subcommand_Init()
		if err != nil {
			log.Fatalln(err)
		}
		os.Exit(exitCode)
	}
	if ctxConfig.Cmd.Ping {
		os.Exit(subCommand_Ping(ctxConfig))
	}
	diff, isNeedAddCommand, err := git.GetDiff(ctxConfig.Cmd)
	if err != nil {
		log.Fatalln("获取差异信息失败，原因：", err)
	}

	messageListScore := startRoutine(ctxConfig, diff)

	var messageList []string
	// 卸载分数外壳，同时以分数排序
	slices.SortFunc(messageListScore, func(a, b types.ScoreMsg) int {
		return b.Score - a.Score
	})
	for _, obj := range messageListScore {
		messageList = append(messageList, obj.Msg)
	}

	msg, err := tui.SelectPrompt(messageList)
	if err != nil {
		if rollbackErr := afterRemoteCallRollback(messageList); rollbackErr != nil {
			log.Println("回退操作失败:", rollbackErr)
		}
		log.Fatalln("选择提交消息时出现了问题，详情：", err)
	}
	err = tui.CallCmd(ctxConfig, msg, isNeedAddCommand)
	if err != nil {
		if rollbackErr := afterRemoteCallRollback(messageList); rollbackErr != nil {
			log.Println("回退操作失败:", rollbackErr)
		}
		log.Fatalln("运行指令的过程中出现错误，详情：", err)
	}
}

func startRoutine(ctxConfig types.Config, diff string) []types.ScoreMsg {
	isFoundElementToString := func(isFound bool, nowIndexString string, loop int, msg string) string {
		var str strings.Builder
		str.WriteString("完成 ")
		str.WriteString(nowIndexString)
		str.WriteString(" / ")
		str.WriteString(strconv.Itoa(loop))
		str.WriteString(" 全部| ")
		if isFound {
			str.WriteString(" 分数增加: ")
		} else {
			str.WriteString(" 新增条目: ")
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
	pText := llm.GetPromptMain()

	var reqWaitGroup sync.WaitGroup
	reqWaitGroup.Add(ctxConfig.Cmd.Loop)
	dataChan := make(chan types.ChanResult[string], ctxConfig.Cmd.Loop)

	for i := 0; i < ctxConfig.Cmd.Loop; i++ {
		go func(routineId int) {
			defer reqWaitGroup.Done()
			commitMessage, err := llm.SendReqCore(pText, diff, ctxConfig, false)
			if err != nil {
				dataChan <- types.ChanResult[string]{Data: "", Err: err, Index: routineId}
			} else {
				dataChan <- types.ChanResult[string]{Data: commitMessage, Err: nil, Index: routineId}
			}
		}(i)
	}

	// 等待用 go routine
	go func() {
		reqWaitGroup.Wait()
		close(dataChan)
	}()
	routineIndex := 0
	var messageListScore []types.ScoreMsg
	lengthStringSize := len(strconv.Itoa(ctxConfig.Cmd.Loop))
	for data := range dataChan {
		routineIndex++
		commitMessageText := data.Data
		if data.Err != nil {
			log.Println("好像哪儿出问题了：", data.Err)
			continue
		}
		// 如果找到相通条目，分数+1
		isFoundElement := false
		for index, el := range messageListScore {
			if el.Msg == commitMessageText {
				messageListScore[index].Score++
				isFoundElement = true
				break
			}
		}
		// 如果没有找到相同项目，新建条目，分数=1
		if !isFoundElement {
			messageListScore = append(messageListScore, types.ScoreMsg{Score: 1, Msg: commitMessageText})
		}
		nowIndexString := padLeadingZero(routineIndex, lengthStringSize)
		fmt.Println(isFoundElementToString(isFoundElement, nowIndexString, ctxConfig.Cmd.Loop, commitMessageText))
	}
	return messageListScore
}

// 当调用 LLM 接口后程序后处理报错时回退
func afterRemoteCallRollback(msg []string) error {
	readySaveString := strings.Join(msg, "\n")
	uuid, err := uuid.NewRandom()
	if err != nil {
		return fmt.Errorf("[回退]生成UUID失败，原因：%w\n遗言：\n%s", err, readySaveString)
	}
	tmpFilePath := filepath.Join(os.TempDir(), "git-commit_"+uuid.String()+".txt")
	err = os.WriteFile(tmpFilePath, []byte(readySaveString), 0666)
	if err != nil {
		return fmt.Errorf("[回退]写入文件失败，原因：%w\n遗言：\n%s", err, readySaveString)
	}
	fmt.Println("[回退]大模型输出结果将保存到", tmpFilePath)
	return nil
}
