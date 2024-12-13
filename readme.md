# Git Diff 分析器与 LLM

这是一个简单的 Go 应用程序，使用大型语言模型（LLM）来分析 Git 仓库中的更改。它会获取当前仓库的 diff，并将其发送到 LLM API 进行分析。LLM 随后会提供更改的简要总结。

## 功能

- **Git Diff 获取**：自动获取当前 Git 仓库的 diff。
- **LLM 集成**：将 diff 发送到 LLM API 进行分析。
- **环境变量支持**：使用 `.env` 文件安全地加载 API 令牌。

## 先决条件

- Go（1.16 或更高版本）
- Git
- LLM 服务的 API 令牌（例如，`BIGMODEL_TOKEN`）

## 安装

1. 克隆仓库：

   ```bash
   git clone ....
   cd
   ```

2. 安装依赖项：

   ```bash
   go mod tidy
   ```

3. 在项目根目录中创建一个 `.env` 文件，并添加你的 API 令牌：

   ```env
   BIGMODEL_TOKEN=your_api_token_here
   ```

## 使用

1. 运行应用程序：

   ```bash
   go run main.go
   ```

2. 应用程序将：
   - 获取当前仓库的 Git diff。
   - 将 diff 发送到 LLM API 进行分析。
   - 打印 LLM 的响应，其中包含更改的简要总结。

## 示例输出

```json
{
  "model": "glm-4-flash",
  "messages": [
    {
      "role": "system",
      "content": "帮我分析一下这个工程有什么改动，简单说说。"
    },
    {
      "role": "user",
      "content": "diff --git a/main.go b/main.go\nindex 1234567..89abcdef 100644\n--- a/main.go\n+++ b/main.go\n@@ -10,6 +10,7 @@ import (\n\n func main() {\n\t// Some code changes here\n+    // Added a new line\n }"
    }
  ],
  "type": "text"
}
```

```json
{
  "response": "项目在 main.go 文件中添加了一行新代码。该行是注释，似乎是一个新功能或备注。"
}
```

## 配置

- **LLM_API**：LLM 服务的 API 端点。当前设置为 `https://open.bigmodel.cn/api/paas/v4/chat/completions`。
- **BIGMODEL_TOKEN**：你的 API 令牌，应存储在 `.env` 文件中。

## 贡献

欢迎贡献！请随时提交 Pull Request。
