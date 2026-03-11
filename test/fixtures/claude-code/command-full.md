---
name: deploy-service
description: Deploys service to production
tools:
  - Bash
execution:
  context: fork
  subtask: false
  agent: code-reviewer
arguments:
  hint: "Target environment (staging|production)"
model: claude-sonnet-4-5-20250929
---
Deployment workflow for $ARGUMENTS environment:

1. Build the service:
```bash
go build -o service .
```

2. Run health checks:
```bash
./service --health-check
```

3. Deploy to $ARGUMENTS:
```bash
kubectl apply -f k8s/
```
