# Troubleshooting Guide

## Scenario 1: CDK Deploy Fails with "Insufficient Permissions"
**Symptom**: `cdk deploy` fails with IAM permission errors.  
**Possible Causes**:
-   Missing CloudFormation permissions
-   Missing service-linked role creation permission
-   Account-level resource limits (VPC, EIP, etc.)

**Fix**:  
1. Check IAM policy has `cloudformation:*`, `iam:CreateRole`, `ec2:AllocateAddress`
2. Review CloudFormation stack events: `aws cloudformation describe-stack-events --stack-name <stack-name>`
3. Request limit increase if hitting service quotas

## Scenario 2: Cannot SSH to Bastion Host
**Symptom**: SSH connection timeout when trying to access bastion.  
**Possible Causes**:
-   Security group not allowing your IP
-   Bastion host in wrong subnet (should be public)
-   Key pair mismatch

**Fix**:  
1. Update `OFFICE_CIDR` environment variable to your current IP: `export OFFICE_CIDR=$(curl -s ifconfig.me)/32`
2. Redeploy: `cdk deploy`
3. Verify security group ingress: `aws ec2 describe-security-groups --group-ids <bastion-sg-id>`

## Scenario 3: Lambda Function Timing Out
**Symptom**: Lambda logs show frequent timeouts.  
**Possible Causes**:
-   VPC Lambda without NAT (cannot reach internet)
-   Insufficient memory allocation
-   Slow downstream dependency (RDS, API)

**Fix**:  
1. Check Lambda is in private subnet with NAT Gateway egress
2. Increase timeout: Update CDK stack, redeploy
3. Add CloudWatch Insights query to identify slow operations
4. Consider provisioned concurrency for cold start issues

---

<div align="center">
  <a href="https://infratales.com">InfraTales</a> •
  <a href="https://infratales.com/projects">Projects</a>
</div>
