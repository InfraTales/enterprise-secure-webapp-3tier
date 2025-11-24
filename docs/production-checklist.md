# Production Readiness Checklist

Use this checklist before deploying to production.

## Security

- [ ] **IAM Policies:** All roles follow least privilege principle
- [ ] **Encryption:** Data encrypted at rest (KMS) and in transit (TLS)
- [ ] **Secrets Management:** No hardcoded credentials
- [ ] **Network Isolation:** Resources in private subnets where appropriate
- [ ] **Security Groups:** Ingress restricted to known sources
- [ ] **MFA:** Enabled for privileged accounts
- [ ] **CloudTrail:** Enabled and logging to secure S3 bucket
- [ ] **Security Scan:** Passed Gitleaks/Trivy scan
- [ ] **Penetration Test:** Completed (if required)

## Reliability

- [ ] **Multi-AZ:** Critical resources deployed across AZs
- [ ] **Auto Scaling:** Configured for compute resources
- [ ] **Health Checks:** Load balancer health checks configured
- [ ] **RTO/RPO:** Defined and documented
- [ ] **Backup Strategy:** Automated backups enabled
- [ ] **Disaster Recovery:** DR plan documented and tested

## Performance

- [ ] **Load Testing:** Completed under expected traffic
- [ ] **Resource Sizing:** Instances right-sized for workload
- [ ] **Caching:** Implemented where appropriate
- [ ] **Database Indexing:** Optimized for queries
- [ ] **CDN:** Configured for static assets (if applicable)

## Cost Optimization

- [ ] **Budget Alerts:** AWS Budget configured
- [ ] **Cost Anomaly Detection:** Enabled
- [ ] **Reserved Instances:** Evaluated for steady-state workloads
- [ ] **Right-Sizing:** Instances match actual usage
- [ ] **Lifecycle Policies:** S3 transitions configured
- [ ] **NAT Optimization:** Evaluated alternatives (VPC endpoints)
- [ ] **Development Toggles:** Cost-saving features for non-prod

## Monitoring & Observability

- [ ] **CloudWatch Alarms:** Critical metrics monitored
- [ ] **SNS Notifications:** Alerts routed to team
- [ ] **Log Aggregation:** CloudWatch Logs or equivalent
- [ ] **Dashboards:** Created for key metrics
- [ ] **Distributed Tracing:** X-Ray enabled (if applicable)
- [ ] **Synthetic Monitoring:** Health check from external location

## Compliance

- [ ] **Data Residency:** Data stored in approved regions
- [ ] **Audit Logging:** Comprehensive audit trail
- [ ] **Access Control:** Role-based access implemented
- [ ] **Retention Policies:** Comply with regulations
- [ ] **Privacy:** PII handling documented

## Operational

- [ ] **Runbook:** Created in `docs/runbook.md`
- [ ] **Troubleshooting Guide:** Created in `docs/troubleshooting.md`
- [ ] **On-Call Rotation:** Defined
- [ ] **Rollback Plan:** Documented and tested
- [ ] **Change Management:** Process defined
- [ ] **Documentation:** All docs up to date
- [ ] **Diagrams:** Architecture diagrams accurate

## Testing

- [ ] **Unit Tests:** 80%+ coverage
- [ ] **Integration Tests:** Pass on deployed stack
- [ ] **E2E Tests:** Critical paths validated
- [ ] **Smoke Tests:** Post-deployment validation
- [ ] **Performance Tests:** Under load
- [ ] **Chaos Engineering:** Failure scenarios tested (optional)

## Deployment

- [ ] **CI/CD Pipeline:** Fully automated
- [ ] **Blue/Green Deployment:** Configured (if applicable)
- [ ] **Canary Deployment:** Configured (if applicable)
- [ ] **Feature Flags:** Implemented for risky features
- [ ] **Database Migrations:** Tested and reversible

## Final Sign-Off

| Role | Name | Date | Signature |
| :--- | :--- | :--- | :--- |
| **Tech Lead** | | | |
| **Security** | | | |
| **DevOps** | | | |
| **Product** | | | |

---

**Deployment Date:** ___________

**Post-Deployment Notes:**
- [ ] Smoke tests passed
- [ ] Monitoring confirmed working
- [ ] No critical alarms

---

<div align="center">
  <a href="https://infratales.com">InfraTales</a>
</div>
