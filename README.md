# git-msg

[English](README.md) | [简体中文](README.zh_cn.md)

[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/conglinyizhi/git-msg)

`git-msg` is a Git helper tool.

Its main function is to automatically generate commit messages that conform to the [Conventional Commits](https://www.conventionalcommits.org/) specification, complete with appropriate emojis, by analyzing the current repository's diff and calling a large model API. It supports interactive confirmation, staging area management, and custom prompt templates.

According to the library's documentation, the interactive interface may have issues on Windows.

## ✨ Features

- Automatically retrieves diffs from the Git working directory or staging area.
- Calls a large model (compatible with OpenAI API format) to generate commit messages.
- Generated commit message format: `<emoji> <type>(<scope>): <description>` (e.g., `🔬 test(cli.go): add new command line feature -t`).
- Interactively asks whether to add unstaged files and whether to commit.
- Supports custom prompts (via Markdown files in the `skill/` directory).
- Can specify a custom Git command (e.g., `yadm`).

## 📦 Installation

### Build from Source

Ensure Go 1.18+ is installed, then execute:

```bash
git clone https://github.com/conglinyizhi/git-msg.git
cd git-msg
go build -o git-msg
```

Place the generated `git-msg` binary in your `PATH`, e.g., `~/.local/bin`.

```bash
install -Dm755 git-msg ~/.local/bin/
```

## ⚙️ Configuration

`git-msg` requires configuration of the large model's API information. There are two ways to configure it (in order of priority, highest first):

1. **Configuration File** (recommended):  
   The tool creates a `llm.toml` file in the system's standard configuration directory:
   - Linux: `~/.config/git-msg/llm.toml`
   - macOS: `~/Library/Preferences/git-msg/llm.toml`
   - Windows: `%LOCALAPPDATA%\git-msg\Config\llm.toml`

   You can also use the `--init` flag to quickly create the configuration file, which will also copy the built-in skills to the designated skill directory.

   Example file content:

   ```toml
   API_KEY = "your-api-key"
   BASE_URL = "https://api.openai.com/v1/chat/completions"
   MODEL_NAME = "gpt-3.5-turbo"
   ```

2. **Environment Variables**:
   - `API_KEY`: API key
   - `BASE_URL`: API endpoint
   - `MODEL_NAME`: model name

   For example:

   ```bash
   export API_KEY="your-api-key"
   export BASE_URL="https://api.openai.com/v1/chat/completions"
   export MODEL_NAME="gpt-3.5-turbo"
   ```

## 🚀 Usage

Simply run the tool inside a Git repository:

```bash
git-msg
```

The tool will automatically perform the following steps:

1. Read the configuration (file or environment variables).
2. Retrieve the Git diff (prioritizes the staging area; if empty, falls back to the working directory).
3. Call the large model to generate a commit message (printing the generation process in real-time).
4. Interactively ask:
   - If there are unstaged files, it will ask whether to add them, view the status, or exit.
   - Finally, it will ask whether to execute `git commit`.
5. If the user confirms, it executes the corresponding Git command and outputs the result.

### Command Line Options

- `-g, --git <command>`: Specify the Git command path or alias (default `git`). Useful for replacing it with other Git-compatible tools like `yadm`.
- `-l, --loop <number>`: Send multiple requests simultaneously (default 1) to generate multiple commit messages for selection, making it easier to choose the most appropriate one.
- `--init`: Initialize the environment required for the current version of git-msg (creates configuration directories and default files).
- `--ping`: Test the configured large model API with a minimal prompt to verify connectivity.

Example:

```bash
git-msg -g yadm
git-msg --init
git-msg -l 3  # Generate 3 commit messages simultaneously for selection
```

## 🧠 Custom Prompts

The tool uses Markdown files in the `skill/` directory as system prompts. You can place multiple `.md` files; during runtime, you will be prompted to select one. If only one file exists, it will be used automatically.

Example directory structure:

```
./skill/
  ├── default.md
  ├── detailed.md
  └── emoji-only.md
```

### Skill File Format

Skill files use Markdown format. **Note: The header section of the file (such as YAML frontmatter) is currently not processed specially; all content in the file is embedded directly into the system prompt.**

It is recommended to keep skill file content in plain text prompt format.

## 🤝 Contributing

Issues and Pull Requests are welcome. Please ensure code style is consistent and tests pass.

## 📄 License

MIT License

## Acknowledgements

This tool uses the following excellent open-source libraries:

- [go-toml](https://github.com/pelletier/go-toml) (MIT)
- [pflag](https://github.com/spf13/pflag) (BSD-3)
- [promptkit](https://github.com/erikgeiser/promptkit) (MIT)
- [cloudwego/eino](https://github.com/cloudwego/eino) (Apache-2.0)
