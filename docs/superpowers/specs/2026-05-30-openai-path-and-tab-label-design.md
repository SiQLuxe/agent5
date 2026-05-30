# OpenAI 请求路径修复 & Tab 标签自动更新设计

**日期:** 2026-05-30
**状态:** 草稿

## 目标

1. 修复 `[models.openai]` 请求路径：给 `base_url` 添加 `/v1` 前缀
2. AI 回复完成后，用用户消息内容自动更新 tab 标签

## 改动

### 1. 配置：给 OpenAI base URL 添加 /v1

**文件:** `configs/config.toml`

将 `[models.openai]` 的 `base_url` 从 `http://localhost:8080` 改为 `http://localhost:8080/v1`。

请求路径从 `http://localhost:8080/chat/completions` 变为 `http://localhost:8080/v1/chat/completions`，符合 OpenAI 兼容 API 规范。

### 2. AI 回复后自动更新 Tab 标签

**文件:** `internal/ui/app.go`（`sendMessage()` 方法内）

AI 流式回复完成后，调用 `GenerateLabel()` 获取标签并更新到 tab 栏。

**流程:**
1. 用户发送消息 → AI 开始流式回复（标签仍为 "New Session"）
2. AI 回复完成 → `QueueUpdateDraw` 执行收尾逻辑
3. `isLoading = false` 之后，调用 `sessionPtr.GenerateLabel()`
4. 如果返回的 label 不是 "New Session"，调用 `a.tabDock.UpdateTab(a.activeSession, label)` 更新 tab
5. `GenerateLabel` 已有截断逻辑（20 字 + `...`）和斜杠命令检测，无需额外修改

**效果:** 发送 `say hi` → AI 回复完成后，tab 标签从 "New Session" 变为 "say hi"。

**注意:** 仅第一次 AI 回复后会更新标签。后续发送消息不会覆盖已有标签（`GenerateLabel` 判断 label 已存在则跳过）。

## 验证

- 运行 `go test ./internal/ui/...` — 现有标签/tab 测试必须通过
- 构建并启动应用，发送 `say hi`，观察请求成功且回复完成后 tab 标签更新
