# InfraTales | Enterprise Secure WebApp – Production 3-Tier Architecture

**Production-ready reference architecture for enterprise-grade secure 3-tier web architectures on AWS.**

This repository provides a complete, secure, and cost-aware implementation of a multi-tier secure web application using AWS CDK and Go. It is designed to solve deploying secure, scalable web applications with proper network isolation by deploying VPC with public/private subnets, EC2 instances, RDS database, Application Load Balancer with built-in resilience and observability.

![Architecture Diagram](diagrams/architecture.mmd)

[![CI](https://github.com/infratales/enterprise-3tier-secure-cdk-go/workflows/InfraTales%20CI/badge.svg)](https://github.com/infratales/enterprise-3tier-secure-cdk-go/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

---

## Table of Contents
- [Architecture](#architecture)
- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Cost Estimate](#cost-estimate)
- [Security Posture](#security-posture)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)

---

## Architecture

### The Problem
Many organizations struggle with securely deploying multi-tier web applications that follow AWS best practices for network isolation, encryption, and high availability.
*Example: "Running a high-traffic web application requires scaling, fault tolerance, and strict network isolation, which is complex to configure manually."*

### The Solution
This project deploys a production-ready 3-tier architecture with VPC isolation, encrypted RDS database, and Application Load Balancer across multiple availability zones.
*Example: "We implement a 3-tier architecture using VPC, private subnets for compute, and Multi-AZ RDS, managed entirely via IaC."*

### Key Decisions
- **Multi-AZ Deployment**: High availability and fault tolerance across availability zones
- **Private Subnets for Data Tier**: Enhanced security by isolating database from internet access

---

## Prerequisites

- **Go**: Version 1.21+
- **AWS CDK CLI**: Version 2.114.0+
- **AWS CLI**: Configured with appropriate permissions.
- **Docker**: (If applicable)

---

## Quick Start

For a detailed walkthrough, refer to the Quick Start section above.

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/infratales/<REPO_NAME>.git
    cd <REPO_NAME>
    ```

2.  **Install dependencies:**
    ```bash
    go mod download
    ```

3.  **Deploy:**
    ```bash
    cdk deploy --all
    ```

4.  **Verify:**
    Check the outputs for VPC ID, Load Balancer DNS, RDS Endpoint.

5.  **Teardown:**
    ```bash
    cdk destroy --all
    ```

---

## Configuration

This project uses environment variables and CDK context (e.g., environment variables, context files) for configuration.

| Variable | Description | Default | Required |
| :--- | :--- | :--- | :--- |
| `PROJECT_NAME` | Name of the project tag | `demo` | No |
| `ENVIRONMENT` | Target environment (dev/prod) | `dev` | No |
| `AWS_REGION` | Target AWS Region | `us-east-1` | No |
| `OFFICE_CIDR` | Allowed ingress CIDR | `0.0.0.0/0` | **Yes** |

---

## Cost Estimate

**Monthly Estimate (Approximate, in ₹)**

| Resource | Configuration | Est. Cost (₹) | Notes |
| :--- | :--- | :--- | :--- |
| **Compute** | t3.medium EC2 | ₹2,800 | 2 instances across AZs |
| **Database** | RDS MySQL Multi-AZ | ₹4,500 | db.t3.small with automated backups |
| **Network** | NAT Gateways / LB | ₹3,200 | NAT Gateway + ALB |
| **Total** | | **~₹10,500** | *Varies by traffic* |

> **Tip:** See [docs/cost.md](docs/cost.md) for detailed breakdown and saving tips.

---

## Security Posture

- **IAM:** Least privilege policies applied to all roles.
- **Network:** Resources deployed in private subnets; strict Security Groups.
- **Encryption:** KMS used for data at rest; TLS for transit.
- **Compliance:** CIS AWS Foundations Benchmark ready.

> See [docs/security.md](docs/security.md) for the full threat model.

---

## Troubleshooting

| Issue | Probable Cause | Fix |
| :--- | :--- | :--- |
| **Deployment Fails with Subnet Error** | Insufficient IP addresses in VPC CIDR | Expand VPC CIDR or reduce subnet mask |
| **Cannot Connect to RDS** | Security group not allowing EC2 traffic | Update RDS security group inbound rules |

> See [docs/troubleshooting.md](docs/troubleshooting.md) for more scenarios.

---

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

---

<div align="center">
  <b>InfraTales | Accelerating secure, cost-aware IaC delivery.</b>
  <br><br>
  <a href="https://infratales.com">Website</a> •
  <a href="https://infratales.com/projects">Projects</a> •
  <a href="https://infratales.com/premium">Premium</a> •
  <a href="https://infratales.com/newsletter">Newsletter</a>
</div>
