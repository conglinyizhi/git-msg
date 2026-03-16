package main

type RemoteAPIConfig struct {
	API_KEY    string
	BASE_URL   string
	MODEL_NAME string
}

type CommandlineConfig struct {
	git  string
	loop int
	init bool
	ping bool
}

type ChanResult[T any] struct {
	data  T
	err   error
	index int
}

type ScoreMsg struct {
	msg   string
	score int
}
