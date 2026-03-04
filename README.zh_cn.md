# git-msg

[English](README.md) | [简体中文](README.zh_cn.md)

`git-msg` 是一个 Git 辅助工具

当前的主要功能是通过分析当前仓库的差异（diff），调用大模型 API 自动生成符合[约定式提交](https://www.conventionalcommits.org/)规范的提交信息，并附带适当的 emoji。它支持交互式确认、暂存区管理，以及自定义提示词模板。

根据依赖库的说明，交互界面可能在 Windows 上出现问题

## ✨ 特性

- 自动获取 Git 工作区或暂存区的差异
- 调用大模型（兼容 OpenAI API 格式）生成提交信息
- 生成的提交信息格式：`<emoji> <type>(<scope>): <description>`（例如 `🔬 test(cli.go): 新增了命令行功能 -t 参数`）
- 交互式询问是否添加未暂存文件、是否提交
- 支持自定义提示词（通过 `skills/` 目录下的 Markdown 文件）
- 可指定自定义 Git 命令（如 `yadm`）

## 📦 安装

### 从源码编译

确保已安装 Go 1.18+，然后执行：

```bash
git clone https://github.com/yourusername/git-msg.git
cd git-msg
go build -o git-msg
```

将生成的 `git-msg` 二进制文件放到 `PATH` 下，例如 `/usr/local/bin`。

### 直接下载

前往 [Releases](https://github.com/yourusername/git-msg/releases) 页面下载对应平台的预编译二进制文件。

## ⚙️ 配置

`git-msg` 需要配置大模型的 API 信息。配置方式有两种（按优先级从高到低）：

1. **配置文件**（推荐）：  
   工具会在系统的标准配置目录下创建 `llm.toml` 文件：
   - Linux: `~/.config/git-msg/llm.toml`
   - macOS: `~/Library/Preferences/git-msg/llm.toml`
   - Windows: `%LOCALAPPDATA%\git-msg\Config\llm.toml`

   文件内容示例：

   ```toml
   API_KEY = "your-api-key"
   BASE_URL = "https://api.openai.com/v1/chat/completions"
   MODEL_NAME = "gpt-3.5-turbo"
   ```

2. **环境变量**：
   - `API_KEY`：API 密钥
   - `BASE_URL`：API 端点
   - `MODEL_NAME`：模型名称

   例如：

   ```bash
   export API_KEY="your-api-key"
   export BASE_URL="https://api.openai.com/v1/chat/completions"
   export MODEL_NAME="gpt-3.5-turbo"
   ```

## 🚀 使用方法

在 Git 仓库目录下直接运行：

```bash
git-msg
```

工具会自动执行以下步骤：

1. 读取配置（文件或环境变量）
2. 获取 Git 差异（优先暂存区，若为空则尝试工作区）
3. 调用大模型生成提交信息（并实时打印生成过程）
4. 交互式询问：
   - 如果存在未暂存的文件，会询问是否添加、查看状态或退出
   - 最后询问是否执行 `git commit`
5. 若用户确认，执行相应的 Git 命令并输出结果

### 命令行选项

- `-g, --git <command>`：指定 Git 命令路径或别名（默认 `git`），可用于替代为 `yadm` 等其他 Git 兼容工具。

示例：

```bash
git-msg -g yadm
```

## 🧠 自定义提示词

工具使用 `skills/` 目录下的 Markdown 文件作为 system prompt。你可以放置多个 `.md` 文件，运行时将让你选择使用哪一个。如果只有一个文件，则自动使用。

目录结构示例：

```
./skills/
  ├── default.md
  ├── detailed.md
  └── emoji-only.md
```

每个文件应包含 system prompt 内容。默认 prompt 见代码中的 `defaultPrompt` 常量。

## 📝 示例

假设你在某个 Git 仓库修改了 `cli.go` 并添加了测试，运行 `git-msg`：

```bash
$ git-msg
获取差异信息...
调用远程大模型...
🔬 test(cli.go): 新增了命令行功能 -t 参数，用于指定具体分类。

检测到暂存区外的文件差异，是否需要添加到暂存区？
> Yes
  No
  查看仓库状态
  退出

选择 Yes 后：
add 执行完成，空输出
一切准备就绪，发起提交吗? [Y/n] y
commit -m 输出的内容
---
[main 1a2b3c4] 🔬 test(cli.go): 新增了命令行功能 -t 参数，用于指定具体分类。
 1 file changed, 10 insertions(+)
```

## 🤝 贡献

欢迎提交 Issue 和 Pull Request。请确保代码格式规范，并通过测试。

## 📄 许可证

MIT License

## 致谢

本工具使用了以下优秀的开源库：

- [go-toml](https://github.com/pelletier/go-toml) (MIT)
- [pflag](https://github.com/spf13/pflag) (BSD-3)
- [promptkit](https://github.com/erikgeiser/promptkit) (MIT)
