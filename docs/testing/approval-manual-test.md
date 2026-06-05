# Manual Test: Agent File Write Approval Modal

## Prerequisites

- Real LLM configured in `configs/config.toml` (e.g., `default_client = "openai"` with API key)
- Set `sandbox_dir = ""` (empty) in `configs/config.toml` to let OS auto-detect the sandbox path
- Project builds successfully: `go build ./cmd/agent`

## Steps

1. Run the TUI:
   ```bash
   go run ./cmd/agent
   ```

2. In the chat panel, send:
   ```
   write a hello world python program to hello.py
   ```

3. Observe that the agent processes the request through the ReAct loop and attempts to call `write_file`.

4. An approval modal overlay should appear showing:
   - File path: `hello.py` (resolved within sandbox)
   - Diff preview: new file with `print("hello")`
   - Prompt: `Approve? (y/n/d)`

5. Press `y` to approve — verify the file is written
6. Press `n` to reject — verify the file is NOT written and agent reports rejection
7. Press `d` to view full diff details

## Verification

After approving, check the sandbox directory for the file:

```bash
# Windows: %TEMP%\agent-tui\sandbox\
# Linux: /tmp/agent-tui/sandbox/
ls hello.py
cat hello.py  # should print "hello"
```

## Expected Results

| Action | Result |
|--------|--------|
| Approve (y) | File written, agent continues |
| Reject (n) | File not written, agent reports error |
| Diff (d) | Full diff shown in modal |
