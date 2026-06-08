go build -o agent.exe ./cmd/agent
if ($?) { Write-Output "Built: agent.exe" }
