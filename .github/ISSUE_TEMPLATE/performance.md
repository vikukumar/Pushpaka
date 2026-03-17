---
name: Performance Issue
about: Report performance problems or optimization opportunities
title: '[PERF] '
labels: 'performance'
assignees: ''
---

## Issue Type
- [ ] Slow API responses
- [ ] High memory/CPU usage
- [ ] Deployment takes too long
- [ ] Build process is slow
- [ ] Dashboard is unresponsive
- [ ] Database queries are slow
- [ ] Other (specify)

## Description
Clear description of the performance problem.

## Performance Metrics

**Current Performance:**
- **Metric:** [e.g., API response time]
- **Current Value:** [e.g., 5s]
- **Expected Value:** [e.g., <500ms]
- **Frequency:** [Always, Intermittent, Under load]

**Impact:**
- [ ] Affects local development
- [ ] Affects small deployments
- [ ] Affects large deployments
- [ ] Affects all deployments

## Steps to Reproduce
1. Configure system as follows...
2. Perform action...
3. Observe performance degradation

## Current Behavior
What currently happens?

```
Logs or output showing the issue
```

## Expected Behavior
How should this perform?

## Environment
- **OS:** [e.g., Linux, macOS]
- **CPU:** [e.g., 4 cores]
- **Memory:** [e.g., 8GB]
- **Pushpaka Version:** [v1.0.0]
- **Deployment:** [Docker/K8s/Single binary]
- **Scale:** [Number of projects, deployments, etc.]

## Profiling Data
If available, attach:
- CPU flame graphs
- Memory profiles
- Query execution plans
- Network traces

## Investigated Causes
Have you investigated what might be causing this?

## Potential Solutions
Any ideas on how to optimize this?

## Additional Context
Links to related issues, discussions, or external resources.

## Checklist
- [ ] I've confirmed this happens consistently
- [ ] I've searched for similar issues
- [ ] I've provided performance metrics
- [ ] I've described the environment
- [ ] I've attached profiling data (if possible)
