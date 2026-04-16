package types

const AppName = "git-msg"

type RemoteAPIConfig struct {
	API_KEY    string
	BASE_URL   string
	MODEL_NAME string
}

type CommandlineConfig struct {
	Git  string
	Loop int
	Init bool
	Ping bool
	Note string
}

type Config struct {
	Cmd CommandlineConfig
	Api RemoteAPIConfig
}

type ChanResult[T any] struct {
	Data  T
	Err   error
	Index int
}

type ScoreMsg struct {
	Msg   string
	Score int
}
