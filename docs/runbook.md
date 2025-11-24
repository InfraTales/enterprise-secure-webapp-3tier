# Operations Runbook

## Alerts & Thresholds

| Alert Name | Threshold | Severity | Action |
| :--- | :--- | :--- | :--- |
| **Lambda Errors** | > 5% error rate | High | Check CloudWatch Logs, rollback if needed |
| **NAT Gateway Packet Drop** | > 1000/min | Critical | Check bandwidth limits, scale |
| **WAF Block Rate** | > 100/min | Medium | Investigate source IPs, adjust rules |
| **Config Non-Compliant** | Any resource | High | Review Config dashboard, remediate |

## Common Tasks

### 1. Rotate Bastion SSH Keys
**Trigger**: Scheduled rotation (90 days) or compromise
**Steps**:
1.  Generate new key pair: `aws ec2 create-key-pair --key-name webapp-bastion-new`
2.  Update EC2 instance with new authorized_keys
3.  Test access with new key
4.  Revoke old key: `aws ec2 delete-key-pair --key-name webapp-bastion-old`

### 2. Scale Lambda Concurrency
**Trigger**: Throttling errors in CloudWatch
**Steps**:
1.  Request limit increase: AWS Support ticket
2.  OR implement SQS buffering for burst traffic
3.  Monitor reserved concurrency usage

### 3. Update WAF Rules
**Trigger**: New attack pattern detected
**Steps**:
1.  Test rule in CloudFormation: `cdk diff`
2.  Deploy: `cdk deploy`
3.  Monitor WAF metrics for false positives

---

<div align="center">
  <a href="https://infratales.com">InfraTales</a> •
  <a href="https://infratales.com/projects">Projects</a>
</div>
