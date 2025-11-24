# Cost Analysis

## Monthly Estimate (us-east-1, ₹)

| Resource | Configuration | Est. Cost (₹) | Notes |
| :--- | :--- | :--- | :--- |
| **NAT Gateways** | 2x (Multi-AZ) | ₹7,000 | $0.045/hr × 2 × 730hrs |
| **CloudFront** | 100GB transfer | ₹1,500 | Traffic-dependent |
| **Lambda** | 1M invokes, 128MB | ₹300 | Compute + requests |
| **EC2 Bastion** | t3.micro on-demand | ₹750 | Can use Spot for 70% savings |
| **S3 Storage** | 100GB Standard | ₹250 | + KMS ~₹100 |
| **AWS Config** | 5 rules | ₹600 | ₹120/rule |
| **WAF** | 1 WebACL, 3 rules | ₹1,800 | ₹600 WebACL + ₹400/rule |
| **Secrets Manager** | 5 secrets | ₹250 | ₹50/secret |
| **VPC Flow Logs** | S3 storage | ₹200 | Variable |
| **CloudWatch** | Logs + metrics | ₹400 | Log ingestion |
| **Total** | | **~₹13,050** | Can vary ±30% based on traffic |

## Cost Saving Tips

### Development Environment
1.  **Single AZ**: Use 1 NAT Gateway instead of 2 → Save ₹3,500/month
2.  **Spot Instances**: Use Spot for bastion → Save ₹525/month (70%)
3.  **Disable Config**: Turn off AWS Config in dev → Save ₹600/month
4.  **Simplify WAF**: Use 1 rule vs 3 → Save ₹800/month

### Production Environment
1.  **Reserved NAT**: Use NAT Gateway savings plan → Save 20-30%
2.  **S3 Lifecycle**: Move old logs to Glacier after 90 days → Save 50% on old data
3.  **CloudWatch Insights**: Use log filtering to reduce ingestion → Variable savings
4.  **Right-Size Lambda**: Monitor memory usage, reduce if possible → Up to 50% savings

---

<div align="center">
  <a href="https://infratales.com">InfraTales</a> •
  <a href="https://infratales.com/projects">Projects</a>
</div>
