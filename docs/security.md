# Security Model

## Threat Model

| Threat | Mitigation | Status |
| :--- | :--- | :--- |
| **Unauthorized Access** | Least-privilege IAM, bastion-only SSH | ✅ Implemented |
| **Data Leakage** | KMS encryption, private subnets, VPC FlowLogs | ✅ Implemented |
| **Web Exploits** | WAF on CloudFront, security groups | ✅ Implemented |
| **Compliance Drift** | AWS Config continuous monitoring | ✅ Implemented |
| **Secrets Exposure** | Secrets Manager, no hardcoded values | ✅ Implemented |

## IAM Strategy
-   **Roles**: All AWS services use IAM roles, no access keys
-   **Policies**: Scoped to specific resources with ARN notation
-   **Lambda**: Inline policies for S3/DynamoDB access only
-   **EC2**: SSM Session Manager preferred over SSH where possible

## Network Security
-   **VPC**: Isolated VPC, public/private subnet separation
-   **Security Groups**: Bastion allows SSH (your CIDR), EC2 allows traffic from bastion only
-   **NACLs**: Default allow (stateless firewall optional, not implemented)
-   **Flow Logs**: Enabled to centralized S3 bucket for traffic analysis

## Data Protection
-   **At Rest**: All S3 buckets use customer KMS CMK, EBS volumes encrypted
-   **In Transit**: HTTPS enforced on CloudFront, TLS 1.2+ only
-   **Secrets**: Stored in Secrets Manager, auto-rotation enabled where applicable

## WAF Rules (CloudFront)
-   SQL Injection protection
-   XSS protection
-   Rate limiting (configurable)

---

<div align="center">
  <a href="https://infratales.com">InfraTales</a> •
  <a href="https://infratales.com/projects">Projects</a>
</div>
