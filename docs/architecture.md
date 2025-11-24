# Architecture Deep Dive

## Overview
This project implements a secure, multi-tier web application infrastructure across 2 Availability Zones in us-east-1 using AWS CDK with Go.

## High-Level Design
```
Internet → Cloud Front (WAF) → S3 (Static Assets)
         ↓
      ALB/API Gateway → Lambda (Compute) → RDS (Data)
         ↓
      EC2 (Private Subnets) → Secrets Manager
```

### Components
1.  **VPC**: Custom VPC with public/private subnets, NAT Gateways, VPC Flow Logs→S3
2.  **Compute**: Lambda functions (background jobs), EC2 instances (private only), Bastion host (SSH access)
3.  **CDN**: CloudFront distribution with OAI, WAF WebACL protection
4.  **Storage**: S3 buckets with customer KMS keys, separate logging bucket
5.  **Security**: AWS Config (compliance), SNS (alerts), least-privilege IAM
6.  **Config**: Systems Manager Parameter Store, Secrets Manager with auto-rotation

## Design Decisions

### 1. Go for CDK
-   **Context**: Team proficiency, type safety requirements
-   **Decision**: Use CDK with Go instead of CloudFormation YAML
-   **Tradeoffs**:
    -   *Pros*: Type safety, IDE support, reusable components, actual programming constructs
    -   *Cons*: Learning curve for team unfamiliar with CDK

### 2. Customer-Managed KMS Keys
-   **Context**: Compliance requirements for full encryption control
-   **Decision**: Deploy custom KMS CMKs for all encryption needs
-   **Tradeoffs**:
    -   *Pros*: Full audit trail, key rotation control, compliance ready
    -   *Cons*: Additional cost (~₹100/key/month), complexity in key management

### 3. Multi-AZ Deployment
-   **Context**: High availability requirements
-   **Decision**: Deploy NAT Gateways, subnets across 2 AZs
-   **Tradeoffs**:
    -   *Pros*: Fault tolerance, AZ failure resilience
    -   *Cons*: ~2x cost for NAT Gateways

## Scalability
-   **Horizontal Scaling**: Lambda auto-scales, EC2 can be placed in ASG
-   **Limits**: NAT Gateway bandwidth (45 Gbps per AZ), Lambda concurrency (1000 default)

## Failure Modes
| Failure Scenario | System Behavior | Recovery |
| :--- | :--- | :--- |
| **AZ Failure** | Traffic routes to healthy AZ | Automatic (Multi-AZ) |
| **Lambda Timeout** | Retry with exponential backoff | Automatic (3 retries) |
| **Bastion Compromise** | Revoke keys, rotate secrets | Manual (runbook) |

---

<div align="center">
  <a href="https://infratales.com">InfraTales</a> •
  <a href="https://infratales.com/projects">Projects</a>
</div>
