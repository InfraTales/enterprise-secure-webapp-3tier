# Quick Start Guide

**Deploy in 10 minutes.**

## Prerequisites
1.  **Go 1.21+** installed (`go version`)
2.  **Node.js 20+** installed (`node -v`)
3.  **AWS CDK CLI** installed (`npm install -g aws-cdk`)
4.  **AWS Credentials** configured (`aws configure`)

## 1. Clone & Configure
```bash
git clone https://github.com/infratales/webapp-multitier-secure-cdk-go.git
cd webapp-multitier-secure-cdk-go

# Set environment variables
export AWS_REGION=us-east-1
export OFFICE_CIDR=$(curl -s ifconfig.me)/32
export VPC_FLOW_LOGS_BUCKET=my-centralized-logs
```

## 2. Bootstrap CDK
```bash
cdk bootstrap aws://ACCOUNT-ID/us-east-1
```

## 3. Build & Deploy
```bash
# CDK will use the Go binary
cdk deploy --all --require-approval never
```

## 4. Verify
Check CloudFormation outputs:
```bash
aws cloudformation describe-stacks \
  --stack-name TapStack \
  --query 'Stacks[0].Outputs'
```

## 5. Cleanup
```bash
cdk destroy --all --force
```

---

<div align="center">
  <a href="https://infratales.com">InfraTales</a> •
  <a href="https://infratales.com/projects">Projects</a>
</div>
