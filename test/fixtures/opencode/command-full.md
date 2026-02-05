---
name: deploy-service
description: Deploys service to production
tools:
  - bash
execution:
  context: fork
  subtask: false
  agent: code-reviewer
arguments:
  hint: "Target environment (staging|production)"
model: anthropic/claude-sonnet-4-20250514
targets:
  claude-code:
    disable-model-invocation: false
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
