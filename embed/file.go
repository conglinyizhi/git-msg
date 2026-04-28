package embed

import "embed"

//go:embed skill/*
var SkillFilesEmbed embed.FS

//go:embed prompt.md
var DefaultPrompt string
