# git-msg

[English](README.md) | [简体中文](README.zh_cn.md)

`git-msg` is a Git helper tool.

Its main function is to automatically generate commit messages that conform to the [Conventional Commits](https://www.conventionalcommits.org/) specification, complete with appropriate emojis, by analyzing the current repository's diff and calling a large model API. It supports interactive confirmation, staging area management, and custom prompt templates.

According to the library's documentation, the interactive interface may have issues on Windows.

## ✨ Features

- Automatically retrieves diffs from the Git working directory or staging area.
- Calls a large model (compatible with OpenAI API format) to generate commit messages.
- Generated commit message format: `<emoji> <type>(<scope>): <description>` (e.g., `🔬 test(cli.go): add new command line feature -t`).
- Interactively asks whether to add unstaged files and whether to commit.
- Supports custom prompts (via Markdown files in the `skills/` directory).
- Can specify a custom Git command (e.g., `yadm`).

## 📦 Installation

### Build from Source

Ensure Go 1.18+ is installed, then execute:

```bash
git clone https://github.com/yourusername/git-msg.git
cd git-msg
go build -o git-msg
```

Place the generated `git-msg` binary in your `PATH`, e.g., `/usr/local/bin`.

### Download Pre-built Binaries

Visit the [Releases](https://github.com/yourusername/git-msg/releases) page to download a pre-compiled binary for your platform.

## ⚙️ Configuration

`git-msg` requires configuration of the large model's API information. There are two ways to configure it (in order of priority, highest first):

1. **Configuration File** (recommended):  
   The tool creates a `llm.toml` file in the system's standard configuration directory:
   - Linux: `~/.config/git-msg/llm.toml`
   - macOS: `~/Library/Preferences/git-msg/llm.toml`
   - Windows: `%LOCALAPPDATA%\git-msg\Config\llm.toml`

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

Example:

```bash
git-msg -g yadm
```

## 🧠 Custom Prompts

The tool uses Markdown files in the `skills/` directory as system prompts. You can place multiple `.md` files; during runtime, you will be prompted to select one. If only one file exists, it will be used automatically.

Example directory structure:

```
./skills/
  ├── default.md
  ├── detailed.md
  └── emoji-only.md
```

Each file should contain the system prompt content. The default prompt can be found in the `defaultPrompt` constant in the code.

## 📝 Example

Assume you have modified `cli.go` in a Git repository and added tests. Running `git-msg`:

```bash
$ git-msg
Retrieving diff information...
Calling remote large model...
🔬 test(cli.go): add new command line feature -t to specify a category.

Detected differences outside the staging area. Do you want to add them to the staging area?
> Yes
  No
  View repository status
  Exit

After selecting Yes:
add executed successfully, no output.
Everything is ready. Proceed with commit? [Y/n] y
commit -m the generated message
---
[main 1a2b3c4] 🔬 test(cli.go): add new command line feature -t to specify a category.
 1 file changed, 10 insertions(+)
```

## 🤝 Contributing

Issues and Pull Requests are welcome. Please ensure code style is consistent and tests pass.

## 📄 License

MIT License

## Acknowledgements

This tool uses the following excellent open-source libraries:

- [go-toml](https://github.com/pelletier/go-toml) (MIT)
- [pflag](https://github.com/spf13/pflag) (BSD-3)
- [promptkit](https://github.com/erikgeiser/promptkit) (MIT)
