---
name: deploy-service
description: Deploys service to production
agent: code-reviewer
model: anthropic/claude-sonnet-4-20250514
context: fork
subtask: false
argument-hint: "Target environment (staging|production)"
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
