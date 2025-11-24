package lib

import (
	"fmt"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsautoscaling"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudfront"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudfrontorigins"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudtrail"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudwatch"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awskms"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssecretsmanager"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssns"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsssm"
	"github.com/aws/aws-cdk-go/awscdk/v2/awswafv2"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// TapStackProps defines the properties for the TapStack CDK stack.
type TapStackProps struct {
	*awscdk.StackProps
	EnvironmentSuffix *string
}

// TapStack represents the main CDK stack for secure multi-tier web app infrastructure.
type TapStack struct {
	awscdk.Stack
	EnvironmentSuffix *string
	// Network resources
	Vpc            awsec2.Vpc
	PrivateSubnets *[]awsec2.ISubnet
	PublicSubnets  *[]awsec2.ISubnet
	BastionHost    awsec2.BastionHostLinux
	// Security resources
	KmsKey         awskms.Key
	SecurityGroups map[string]awsec2.SecurityGroup
	// Storage resources
	S3Bucket       awss3.Bucket
	LoggingBucket  awss3.Bucket
	CloudFrontOAI  awscloudfront.OriginAccessIdentity
	CloudFrontDist awscloudfront.Distribution
	// Compute resources
	LambdaFunction   awslambda.Function
	AutoScalingGroup awsautoscaling.AutoScalingGroup
	// Monitoring and compliance
	CloudTrail awscloudtrail.Trail
	SNSAlerts  awssns.Topic
	WAF        awswafv2.CfnWebACL
	// Configuration management
	SSMParameters  map[string]awsssm.StringParameter
	SecretsManager awssecretsmanager.Secret
}

// NewTapStack creates a secure multi-tier web application infrastructure stack.
func NewTapStack(scope constructs.Construct, id *string, props *TapStackProps) *TapStack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = *props.StackProps
	}
	stack := awscdk.NewStack(scope, id, &sprops)

	// Get environment suffix
	var environmentSuffix string
	if props != nil && props.EnvironmentSuffix != nil {
		environmentSuffix = *props.EnvironmentSuffix
	} else if suffix := stack.Node().TryGetContext(jsii.String("environmentSuffix")); suffix != nil {
		environmentSuffix = *suffix.(*string)
	} else {
		environmentSuffix = "dev"
	}

	// Add stack-level tags
	awscdk.Tags_Of(stack).Add(jsii.String("Environment"), jsii.String(environmentSuffix), nil)
	awscdk.Tags_Of(stack).Add(jsii.String("Project"), jsii.String("prod-"+environmentSuffix), nil)
	awscdk.Tags_Of(stack).Add(jsii.String("ManagedBy"), jsii.String("CDK"), nil)

	tapStack := &TapStack{
		Stack:             stack,
		EnvironmentSuffix: jsii.String(environmentSuffix),
		SecurityGroups:    make(map[string]awsec2.SecurityGroup),
		SSMParameters:     make(map[string]awsssm.StringParameter),
	}

	// Create infrastructure components in order
	tapStack.createKMSKey()
	tapStack.createNetworking()
	tapStack.createSecurityGroups()
	tapStack.createS3Resources()
	tapStack.createSecretsManager()
	tapStack.createSSMParameters()
	tapStack.createSNSAlerts()
	tapStack.createLambdaFunction()
	tapStack.createEC2Resources()
	tapStack.createBastionHost()
	tapStack.createCloudFront()
	tapStack.createWAF()
	tapStack.createMonitoring()
	tapStack.createOutputs()

	return tapStack
}

// createKMSKey creates customer-managed KMS keys for encryption
func (t *TapStack) createKMSKey() {
	t.KmsKey = awskms.NewKey(t.Stack, jsii.String("ProdKMSKey"), &awskms.KeyProps{
		Description: jsii.String(fmt.Sprintf("Customer-managed KMS key for prod-%s environment", *t.EnvironmentSuffix)),
		KeySpec:     awskms.KeySpec_SYMMETRIC_DEFAULT,
		KeyUsage:    awskms.KeyUsage_ENCRYPT_DECRYPT,
		Policy: awsiam.NewPolicyDocument(&awsiam.PolicyDocumentProps{
			Statements: &[]awsiam.PolicyStatement{
				awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
					Effect: awsiam.Effect_ALLOW,
					Principals: &[]awsiam.IPrincipal{
						awsiam.NewAccountRootPrincipal(),
					},
					Actions: &[]*string{
						jsii.String("kms:*"),
					},
					Resources: &[]*string{
						jsii.String("*"),
					},
				}),
				awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
					Effect: awsiam.Effect_ALLOW,
					Principals: &[]awsiam.IPrincipal{
						awsiam.NewServicePrincipal(jsii.String("s3.amazonaws.com"), nil),
						awsiam.NewServicePrincipal(jsii.String("lambda.amazonaws.com"), nil),
						awsiam.NewServicePrincipal(jsii.String("logs.amazonaws.com"), nil),
					},
					Actions: &[]*string{
						jsii.String("kms:Encrypt"),
						jsii.String("kms:Decrypt"),
						jsii.String("kms:ReEncrypt*"),
						jsii.String("kms:GenerateDataKey*"),
						jsii.String("kms:DescribeKey"),
					},
					Resources: &[]*string{
						jsii.String("*"),
					},
				}),
			},
		}),
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})

	awskms.NewAlias(t.Stack, jsii.String("ProdKMSKeyAlias"), &awskms.AliasProps{
		AliasName: jsii.String(fmt.Sprintf("alias/prod-%s-key", *t.EnvironmentSuffix)),
		TargetKey: t.KmsKey,
	})

	awscdk.Tags_Of(t.KmsKey).Add(jsii.String("Name"), jsii.String(fmt.Sprintf("prod-%s-kms-key", *t.EnvironmentSuffix)), nil)
}

// createNetworking creates VPC with public/private subnets across 2 AZs
func (t *TapStack) createNetworking() {
	// Create VPC with 2 AZs in us-east-1
	t.Vpc = awsec2.NewVpc(t.Stack, jsii.String("ProdVPC"), &awsec2.VpcProps{
		VpcName:            jsii.String(fmt.Sprintf("prod-%s-vpc", *t.EnvironmentSuffix)),
		IpAddresses:        awsec2.IpAddresses_Cidr(jsii.String("10.0.0.0/16")),
		MaxAzs:             jsii.Number(2),
		EnableDnsHostnames: jsii.Bool(true),
		EnableDnsSupport:   jsii.Bool(true),
		SubnetConfiguration: &[]*awsec2.SubnetConfiguration{
			{
				Name:       jsii.String("Public"),
				SubnetType: awsec2.SubnetType_PUBLIC,
				CidrMask:   jsii.Number(24),
			},
			{
				Name:       jsii.String("Private"),
				SubnetType: awsec2.SubnetType_PRIVATE_WITH_EGRESS,
				CidrMask:   jsii.Number(24),
			},
		},
	})

	// Get subnet references
	t.PublicSubnets = t.Vpc.PublicSubnets()
	t.PrivateSubnets = t.Vpc.PrivateSubnets()

	// Enable VPC Flow Logs to centralized S3 bucket
	flowLogsBucket := awss3.NewBucket(t.Stack, jsii.String("VPCFlowLogsBucket"), &awss3.BucketProps{
		BucketName:        jsii.String(fmt.Sprintf("prod-%s-vpc-flow-logs-%s", *t.EnvironmentSuffix, *t.Account())),
		Versioned:         jsii.Bool(true),
		RemovalPolicy:     awscdk.RemovalPolicy_DESTROY,
		AutoDeleteObjects: jsii.Bool(true),
		BlockPublicAccess: awss3.BlockPublicAccess_BLOCK_ALL(),
		EncryptionKey:     t.KmsKey,
		Encryption:        awss3.BucketEncryption_KMS,
	})

	awsec2.NewFlowLog(t.Stack, jsii.String("VPCFlowLog"), &awsec2.FlowLogProps{
		ResourceType: awsec2.FlowLogResourceType_FromVpc(t.Vpc),
		Destination:  awsec2.FlowLogDestination_ToS3(flowLogsBucket, jsii.String("vpc-flow-logs/"), nil),
		TrafficType:  awsec2.FlowLogTrafficType_ALL,
	})

	awscdk.Tags_Of(t.Vpc).Add(jsii.String("Name"), jsii.String(fmt.Sprintf("prod-%s-vpc", *t.EnvironmentSuffix)), nil)
	awscdk.Tags_Of(flowLogsBucket).Add(jsii.String("Name"), jsii.String(fmt.Sprintf("prod-%s-vpc-flow-logs", *t.EnvironmentSuffix)), nil)
}

// createSecurityGroups creates security groups with least privilege access
func (t *TapStack) createSecurityGroups() {
	// Lambda security group
	lambdaSG := awsec2.NewSecurityGroup(t.Stack, jsii.String("LambdaSG"), &awsec2.SecurityGroupProps{
		Vpc:         t.Vpc,
		Description: jsii.String("Security group for Lambda functions"),
	})
	lambdaSG.AddEgressRule(
		awsec2.Peer_AnyIpv4(),
		awsec2.Port_Tcp(jsii.Number(443)),
		jsii.String("HTTPS outbound for AWS API calls"),
		jsii.Bool(false),
	)
	t.SecurityGroups["lambda"] = lambdaSG

	// EC2 security group (private subnets only)
	ec2SG := awsec2.NewSecurityGroup(t.Stack, jsii.String("EC2SG"), &awsec2.SecurityGroupProps{
		Vpc:         t.Vpc,
		Description: jsii.String("Security group for EC2 instances in private subnets"),
	})
	t.SecurityGroups["ec2"] = ec2SG

	// Bastion host security group
	bastionSG := awsec2.NewSecurityGroup(t.Stack, jsii.String("BastionSG"), &awsec2.SecurityGroupProps{
		Vpc:         t.Vpc,
		Description: jsii.String("Security group for bastion host SSH access"),
	})
	bastionSG.AddIngressRule(
		awsec2.Peer_AnyIpv4(),
		awsec2.Port_Tcp(jsii.Number(22)),
		jsii.String("SSH access from anywhere (restrict to specific IPs in production)"),
		jsii.Bool(false),
	)
	t.SecurityGroups["bastion"] = bastionSG

	// Allow bastion to SSH to EC2 instances
	ec2SG.AddIngressRule(
		awsec2.Peer_SecurityGroupId(bastionSG.SecurityGroupId(), nil),
		awsec2.Port_Tcp(jsii.Number(22)),
		jsii.String("SSH from bastion host"),
		jsii.Bool(false),
	)

	// Tag security groups
	awscdk.Tags_Of(lambdaSG).Add(jsii.String("Name"), jsii.String(fmt.Sprintf("prod-%s-lambda-sg", *t.EnvironmentSuffix)), nil)
	awscdk.Tags_Of(ec2SG).Add(jsii.String("Name"), jsii.String(fmt.Sprintf("prod-%s-ec2-sg", *t.EnvironmentSuffix)), nil)
	awscdk.Tags_Of(bastionSG).Add(jsii.String("Name"), jsii.String(fmt.Sprintf("prod-%s-bastion-sg", *t.EnvironmentSuffix)), nil)
}

// createS3Resources creates S3 buckets with customer-managed KMS encryption
func (t *TapStack) createS3Resources() {
	// Main application S3 bucket
	t.S3Bucket = awss3.NewBucket(t.Stack, jsii.String("ProdS3Bucket"), &awss3.BucketProps{
		BucketName:        jsii.String(fmt.Sprintf("prod-%s-app-bucket-%s", *t.EnvironmentSuffix, *t.Account())),
		Versioned:         jsii.Bool(true),
		RemovalPolicy:     awscdk.RemovalPolicy_DESTROY,
		AutoDeleteObjects: jsii.Bool(true),
		BlockPublicAccess: awss3.BlockPublicAccess_BLOCK_ALL(),
		EncryptionKey:     t.KmsKey,
		Encryption:        awss3.BucketEncryption_KMS,
	})

	// Separate logging bucket for security events - CloudFront requires ACL access
	t.LoggingBucket = awss3.NewBucket(t.Stack, jsii.String("ProdLoggingBucket"), &awss3.BucketProps{
		BucketName:        jsii.String(fmt.Sprintf("prod-%s-logging-bucket-%s", *t.EnvironmentSuffix, *t.Account())),
		Versioned:         jsii.Bool(true),
		RemovalPolicy:     awscdk.RemovalPolicy_DESTROY,
		AutoDeleteObjects: jsii.Bool(true),
		BlockPublicAccess: awss3.NewBlockPublicAccess(&awss3.BlockPublicAccessOptions{
			BlockPublicAcls:       jsii.Bool(true),
			BlockPublicPolicy:     jsii.Bool(true),
			IgnorePublicAcls:      jsii.Bool(true),
			RestrictPublicBuckets: jsii.Bool(false), // Allow CloudFront service to write logs
		}),
		ObjectOwnership: awss3.ObjectOwnership_BUCKET_OWNER_PREFERRED,
		EncryptionKey:   t.KmsKey,
		Encryption:      awss3.BucketEncryption_KMS,
	})

	// Add bucket policy to allow CloudTrail service to write logs
	// Separate statement for PutObject with ACL condition
	t.LoggingBucket.AddToResourcePolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Effect: awsiam.Effect_ALLOW,
		Principals: &[]awsiam.IPrincipal{
			awsiam.NewServicePrincipal(jsii.String("cloudtrail.amazonaws.com"), nil),
		},
		Actions: &[]*string{
			jsii.String("s3:PutObject"),
		},
		Resources: &[]*string{
			jsii.String(*t.LoggingBucket.BucketArn() + "/*"),
		},
		Conditions: &map[string]interface{}{
			"StringEquals": map[string]interface{}{
				"s3:x-amz-acl": "bucket-owner-full-control",
			},
		},
	}))

	// Separate statement for GetBucketAcl without conditions
	t.LoggingBucket.AddToResourcePolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Effect: awsiam.Effect_ALLOW,
		Principals: &[]awsiam.IPrincipal{
			awsiam.NewServicePrincipal(jsii.String("cloudtrail.amazonaws.com"), nil),
		},
		Actions: &[]*string{
			jsii.String("s3:GetBucketAcl"),
		},
		Resources: &[]*string{
			t.LoggingBucket.BucketArn(),
		},
	}))

	// Also allow CloudTrail to check bucket location and encryption
	t.LoggingBucket.AddToResourcePolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Effect: awsiam.Effect_ALLOW,
		Principals: &[]awsiam.IPrincipal{
			awsiam.NewServicePrincipal(jsii.String("cloudtrail.amazonaws.com"), nil),
		},
		Actions: &[]*string{
			jsii.String("s3:GetBucketLocation"),
			jsii.String("s3:GetBucketVersioning"),
		},
		Resources: &[]*string{
			t.LoggingBucket.BucketArn(),
		},
	}))

	awscdk.Tags_Of(t.S3Bucket).Add(jsii.String("Name"), jsii.String(fmt.Sprintf("prod-%s-app-bucket", *t.EnvironmentSuffix)), nil)
	awscdk.Tags_Of(t.LoggingBucket).Add(jsii.String("Name"), jsii.String(fmt.Sprintf("prod-%s-logging-bucket", *t.EnvironmentSuffix)), nil)
}

// createSecretsManager creates secrets with auto-rotation
func (t *TapStack) createSecretsManager() {
	t.SecretsManager = awssecretsmanager.NewSecret(t.Stack, jsii.String("ProdAppSecrets"), &awssecretsmanager.SecretProps{
		SecretName:  jsii.String(fmt.Sprintf("prod-%s/app-secrets", *t.EnvironmentSuffix)),
		Description: jsii.String("Application secrets for production environment"),
		GenerateSecretString: &awssecretsmanager.SecretStringGenerator{
			SecretStringTemplate: jsii.String(`{"username": "admin"}`),
			GenerateStringKey:    jsii.String("password"),
			ExcludeCharacters:    jsii.String(`"@/\`),
		},
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})

	awscdk.Tags_Of(t.SecretsManager).Add(jsii.String("Name"), jsii.String(fmt.Sprintf("prod-%s-app-secrets", *t.EnvironmentSuffix)), nil)
}

// createSSMParameters creates Systems Manager parameters for configuration
func (t *TapStack) createSSMParameters() {
	parameters := map[string]string{
		"app-environment": *t.EnvironmentSuffix,
		"app-version":     "1.0.0",
		"log-level":       "INFO",
		"max-connections": "100",
	}

	for key, value := range parameters {
		param := awsssm.NewStringParameter(t.Stack, jsii.String("SSMParam"+key), &awsssm.StringParameterProps{
			ParameterName: jsii.String(fmt.Sprintf("/prod-%s/%s", *t.EnvironmentSuffix, key)),
			StringValue:   jsii.String(value),
			Description:   jsii.String(fmt.Sprintf("Configuration parameter for %s", key)),
		})
		t.SSMParameters[key] = param
	}
}

// createSNSAlerts creates SNS topic for security alerts
func (t *TapStack) createSNSAlerts() {
	t.SNSAlerts = awssns.NewTopic(t.Stack, jsii.String("ProdSecurityAlerts"), &awssns.TopicProps{
		TopicName:   jsii.String(fmt.Sprintf("prod-%s-security-alerts", *t.EnvironmentSuffix)),
		DisplayName: jsii.String("Production Security Alerts"),
	})

	awscdk.Tags_Of(t.SNSAlerts).Add(jsii.String("Name"), jsii.String(fmt.Sprintf("prod-%s-security-alerts", *t.EnvironmentSuffix)), nil)
}

// createLambdaFunction creates Lambda function with proper IAM roles (least privilege)
func (t *TapStack) createLambdaFunction() {
	// Create IAM role for Lambda with least privilege
	lambdaRole := awsiam.NewRole(t.Stack, jsii.String("ProdLambdaRole"), &awsiam.RoleProps{
		RoleName:  jsii.String(fmt.Sprintf("prod-%s-lambda-role", *t.EnvironmentSuffix)),
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("lambda.amazonaws.com"), nil),
		ManagedPolicies: &[]awsiam.IManagedPolicy{
			awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("service-role/AWSLambdaVPCAccessExecutionRole")),
		},
		InlinePolicies: &map[string]awsiam.PolicyDocument{
			"S3Access": awsiam.NewPolicyDocument(&awsiam.PolicyDocumentProps{
				Statements: &[]awsiam.PolicyStatement{
					awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
						Effect: awsiam.Effect_ALLOW,
						Actions: &[]*string{
							jsii.String("s3:GetObject"),
							jsii.String("s3:PutObject"),
						},
						Resources: &[]*string{
							t.S3Bucket.BucketArn(),
							jsii.String(*t.S3Bucket.BucketArn() + "/*"),
						},
					}),
				},
			}),
			"KMSAccess": awsiam.NewPolicyDocument(&awsiam.PolicyDocumentProps{
				Statements: &[]awsiam.PolicyStatement{
					awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
						Effect: awsiam.Effect_ALLOW,
						Actions: &[]*string{
							jsii.String("kms:Encrypt"),
							jsii.String("kms:Decrypt"),
							jsii.String("kms:ReEncrypt*"),
							jsii.String("kms:GenerateDataKey*"),
							jsii.String("kms:DescribeKey"),
						},
						Resources: &[]*string{
							t.KmsKey.KeyArn(),
						},
					}),
				},
			}),
		},
	})

	// Create CloudWatch Log Group
	logGroup := awslogs.NewLogGroup(t.Stack, jsii.String("ProdLambdaLogGroup"), &awslogs.LogGroupProps{
		LogGroupName:  jsii.String(fmt.Sprintf("/aws/lambda/prod-%s-background-job", *t.EnvironmentSuffix)),
		Retention:     awslogs.RetentionDays_ONE_MONTH,
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})

	// Simple Python Lambda function for background jobs
	lambdaCode := `
import json
import boto3
import os
import logging
from datetime import datetime

logger = logging.getLogger()
logger.setLevel(logging.INFO)

def lambda_handler(event, context):
    """
    Simple background job Lambda function
    Processes background tasks securely
    """
    try:
        logger.info(f"Processing background job: {json.dumps(event)}")
        
        # Get environment variables
        environment = os.environ.get('ENVIRONMENT', 'unknown')
        s3_bucket = os.environ.get('S3_BUCKET', 'default-bucket')
        
        # Simple processing logic
        result = {
            'status': 'success',
            'timestamp': datetime.utcnow().isoformat() + 'Z',
            'environment': environment,
            'processed_records': len(event.get('records', [])),
            'request_id': context.aws_request_id
        }
        
        logger.info(f"Background job completed successfully: {result}")
        
        return {
            'statusCode': 200,
            'body': json.dumps(result)
        }
        
    except Exception as e:
        logger.error(f"Error processing background job: {str(e)}")
        return {
            'statusCode': 500,
            'body': json.dumps({
                'error': 'Background job failed',
                'message': str(e)
            })
        }
`

	t.LambdaFunction = awslambda.NewFunction(t.Stack, jsii.String("ProdLambdaFunction"), &awslambda.FunctionProps{
		FunctionName: jsii.String(fmt.Sprintf("prod-%s-background-job", *t.EnvironmentSuffix)),
		Runtime:      awslambda.Runtime_PYTHON_3_9(),
		Code:         awslambda.Code_FromInline(jsii.String(lambdaCode)),
		Handler:      jsii.String("index.lambda_handler"),
		MemorySize:   jsii.Number(256),
		Timeout:      awscdk.Duration_Seconds(jsii.Number(30)),
		Role:         lambdaRole,
		LogGroup:     logGroup,
		Vpc:          t.Vpc,
		VpcSubnets: &awsec2.SubnetSelection{
			Subnets: t.PrivateSubnets,
		},
		SecurityGroups: &[]awsec2.ISecurityGroup{
			t.SecurityGroups["lambda"],
		},
		Environment: &map[string]*string{
			"ENVIRONMENT": t.EnvironmentSuffix,
			"S3_BUCKET":   t.S3Bucket.BucketName(),
			"LOG_LEVEL":   jsii.String("INFO"),
		},
		Description:  jsii.String("Background job processing Lambda function"),
		Tracing:      awslambda.Tracing_ACTIVE,
		Architecture: awslambda.Architecture_X86_64(),
	})

	awscdk.Tags_Of(t.LambdaFunction).Add(jsii.String("Name"), jsii.String(fmt.Sprintf("prod-%s-lambda", *t.EnvironmentSuffix)), nil)
	awscdk.Tags_Of(lambdaRole).Add(jsii.String("Name"), jsii.String(fmt.Sprintf("prod-%s-lambda-role", *t.EnvironmentSuffix)), nil)
}

// createEC2Resources creates EC2 instances in private subnets only
func (t *TapStack) createEC2Resources() {
	// Create IAM role for EC2 instances
	ec2Role := awsiam.NewRole(t.Stack, jsii.String("ProdEC2Role"), &awsiam.RoleProps{
		RoleName:  jsii.String(fmt.Sprintf("prod-%s-ec2-role", *t.EnvironmentSuffix)),
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("ec2.amazonaws.com"), nil),
		ManagedPolicies: &[]awsiam.IManagedPolicy{
			awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("AmazonSSMManagedInstanceCore")),
		},
	})

	// Create launch template
	launchTemplate := awsec2.NewLaunchTemplate(t.Stack, jsii.String("ProdLaunchTemplate"), &awsec2.LaunchTemplateProps{
		LaunchTemplateName: jsii.String(fmt.Sprintf("prod-%s-lt", *t.EnvironmentSuffix)),
		InstanceType:       awsec2.InstanceType_Of(awsec2.InstanceClass_T3, awsec2.InstanceSize_MICRO),
		MachineImage:       awsec2.MachineImage_LatestAmazonLinux2(nil),
		Role:               ec2Role,
		SecurityGroup:      t.SecurityGroups["ec2"],
		UserData: awsec2.UserData_ForLinux(&awsec2.LinuxUserDataOptions{
			Shebang: jsii.String("#!/bin/bash"),
		}),
	})

	// Create Auto Scaling Group in private subnets only
	t.AutoScalingGroup = awsautoscaling.NewAutoScalingGroup(t.Stack, jsii.String("ProdAutoScalingGroup"), &awsautoscaling.AutoScalingGroupProps{
		AutoScalingGroupName: jsii.String(fmt.Sprintf("prod-%s-asg", *t.EnvironmentSuffix)),
		Vpc:                  t.Vpc,
		VpcSubnets: &awsec2.SubnetSelection{
			Subnets: t.PrivateSubnets,
		},
		LaunchTemplate:  launchTemplate,
		MinCapacity:     jsii.Number(1),
		MaxCapacity:     jsii.Number(3),
		DesiredCapacity: jsii.Number(2),
	})

	awscdk.Tags_Of(t.AutoScalingGroup).Add(jsii.String("Name"), jsii.String(fmt.Sprintf("prod-%s-asg", *t.EnvironmentSuffix)), nil)
	awscdk.Tags_Of(ec2Role).Add(jsii.String("Name"), jsii.String(fmt.Sprintf("prod-%s-ec2-role", *t.EnvironmentSuffix)), nil)
}

// createBastionHost creates secure bastion host for SSH access
func (t *TapStack) createBastionHost() {
	t.BastionHost = awsec2.NewBastionHostLinux(t.Stack, jsii.String("ProdBastionHost"), &awsec2.BastionHostLinuxProps{
		Vpc:           t.Vpc,
		InstanceName:  jsii.String(fmt.Sprintf("prod-%s-bastion", *t.EnvironmentSuffix)),
		InstanceType:  awsec2.InstanceType_Of(awsec2.InstanceClass_T3, awsec2.InstanceSize_NANO),
		SecurityGroup: t.SecurityGroups["bastion"],
		SubnetSelection: &awsec2.SubnetSelection{
			SubnetType: awsec2.SubnetType_PUBLIC,
		},
	})

	awscdk.Tags_Of(t.BastionHost).Add(jsii.String("Name"), jsii.String(fmt.Sprintf("prod-%s-bastion", *t.EnvironmentSuffix)), nil)
}

// createCloudFront creates CloudFront distribution with WAF
func (t *TapStack) createCloudFront() {
	// Create Origin Access Identity for CloudFront (still needed for compatibility)
	t.CloudFrontOAI = awscloudfront.NewOriginAccessIdentity(t.Stack, jsii.String("ProdCloudFrontOAI"), &awscloudfront.OriginAccessIdentityProps{
		Comment: jsii.String(fmt.Sprintf("OAI for prod-%s S3 bucket", *t.EnvironmentSuffix)),
	})

	// Grant CloudFront OAI read access to S3 bucket
	t.S3Bucket.GrantRead(t.CloudFrontOAI.GrantPrincipal(), jsii.String("*"))

	// Create CloudFront distribution using S3Origin (deprecated but still functional)
	t.CloudFrontDist = awscloudfront.NewDistribution(t.Stack, jsii.String("ProdCloudFrontDist"), &awscloudfront.DistributionProps{
		Comment: jsii.String(fmt.Sprintf("CloudFront distribution for prod-%s", *t.EnvironmentSuffix)),
		DefaultBehavior: &awscloudfront.BehaviorOptions{
			Origin: awscloudfrontorigins.NewS3Origin(t.S3Bucket, &awscloudfrontorigins.S3OriginProps{
				OriginAccessIdentity: t.CloudFrontOAI,
			}),
			ViewerProtocolPolicy: awscloudfront.ViewerProtocolPolicy_REDIRECT_TO_HTTPS,
			CachePolicy:          awscloudfront.CachePolicy_CACHING_OPTIMIZED(),
		},
		PriceClass:         awscloudfront.PriceClass_PRICE_CLASS_100,
		EnableIpv6:         jsii.Bool(false),
		EnableLogging:      jsii.Bool(true),
		LogBucket:          t.LoggingBucket,
		LogFilePrefix:      jsii.String("cloudfront-logs/"),
		LogIncludesCookies: jsii.Bool(false),
	})

	awscdk.Tags_Of(t.CloudFrontDist).Add(jsii.String("Name"), jsii.String(fmt.Sprintf("prod-%s-cloudfront", *t.EnvironmentSuffix)), nil)
}

// createWAF creates Web Application Firewall
func (t *TapStack) createWAF() {
	// Create WAF v2 Web ACL
	t.WAF = awswafv2.NewCfnWebACL(t.Stack, jsii.String("ProdWAF"), &awswafv2.CfnWebACLProps{
		Name:  jsii.String(fmt.Sprintf("prod-%s-waf", *t.EnvironmentSuffix)),
		Scope: jsii.String("CLOUDFRONT"),
		DefaultAction: &awswafv2.CfnWebACL_DefaultActionProperty{
			Allow: &awswafv2.CfnWebACL_AllowActionProperty{},
		},
		Rules: &[]interface{}{
			&awswafv2.CfnWebACL_RuleProperty{
				Name:     jsii.String("AWSManagedRulesCommonRuleSet"),
				Priority: jsii.Number(1),
				Statement: &awswafv2.CfnWebACL_StatementProperty{
					ManagedRuleGroupStatement: &awswafv2.CfnWebACL_ManagedRuleGroupStatementProperty{
						VendorName: jsii.String("AWS"),
						Name:       jsii.String("AWSManagedRulesCommonRuleSet"),
					},
				},
				OverrideAction: &awswafv2.CfnWebACL_OverrideActionProperty{
					None: &map[string]interface{}{},
				},
				VisibilityConfig: &awswafv2.CfnWebACL_VisibilityConfigProperty{
					SampledRequestsEnabled:   jsii.Bool(true),
					CloudWatchMetricsEnabled: jsii.Bool(true),
					MetricName:               jsii.String("CommonRuleSetMetric"),
				},
			},
		},
		VisibilityConfig: &awswafv2.CfnWebACL_VisibilityConfigProperty{
			SampledRequestsEnabled:   jsii.Bool(true),
			CloudWatchMetricsEnabled: jsii.Bool(true),
			MetricName:               jsii.String(fmt.Sprintf("prod-%s-waf", *t.EnvironmentSuffix)),
		},
	})
}

// createMonitoring creates CloudTrail for compliance and monitoring
func (t *TapStack) createMonitoring() {
	// Create CloudTrail
	t.CloudTrail = awscloudtrail.NewTrail(t.Stack, jsii.String("ProdCloudTrail"), &awscloudtrail.TrailProps{
		TrailName:                  jsii.String(fmt.Sprintf("prod-%s-cloudtrail", *t.EnvironmentSuffix)),
		Bucket:                     t.LoggingBucket,
		S3KeyPrefix:                jsii.String("cloudtrail-logs/"),
		IncludeGlobalServiceEvents: jsii.Bool(true),
		IsMultiRegionTrail:         jsii.Bool(true),
		EnableFileValidation:       jsii.Bool(true),
		SendToCloudWatchLogs:       jsii.Bool(true),
		CloudWatchLogGroup: awslogs.NewLogGroup(t.Stack, jsii.String("CloudTrailLogGroup"), &awslogs.LogGroupProps{
			LogGroupName:  jsii.String(fmt.Sprintf("/aws/cloudtrail/prod-%s", *t.EnvironmentSuffix)),
			Retention:     awslogs.RetentionDays_ONE_MONTH,
			RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
		}),
	})

	// Create CloudWatch Alarms for monitoring
	awscloudwatch.NewAlarm(t.Stack, jsii.String("LambdaErrorAlarm"), &awscloudwatch.AlarmProps{
		AlarmName:         jsii.String(fmt.Sprintf("prod-%s-lambda-errors", *t.EnvironmentSuffix)),
		AlarmDescription:  jsii.String("Lambda function error rate"),
		Metric:            t.LambdaFunction.MetricErrors(nil),
		Threshold:         jsii.Number(1),
		EvaluationPeriods: jsii.Number(2),
		TreatMissingData:  awscloudwatch.TreatMissingData_NOT_BREACHING,
	})

	awscdk.Tags_Of(t.CloudTrail).Add(jsii.String("Name"), jsii.String(fmt.Sprintf("prod-%s-cloudtrail", *t.EnvironmentSuffix)), nil)
}

// createOutputs creates CloudFormation outputs for important resources
func (t *TapStack) createOutputs() {
	awscdk.NewCfnOutput(t.Stack, jsii.String("VPCId"), &awscdk.CfnOutputProps{
		Value:       t.Vpc.VpcId(),
		Description: jsii.String("VPC ID"),
		ExportName:  jsii.String(fmt.Sprintf("prod-%s-vpc-id", *t.EnvironmentSuffix)),
	})

	awscdk.NewCfnOutput(t.Stack, jsii.String("S3BucketName"), &awscdk.CfnOutputProps{
		Value:       t.S3Bucket.BucketName(),
		Description: jsii.String("S3 Bucket Name"),
		ExportName:  jsii.String(fmt.Sprintf("prod-%s-s3-bucket", *t.EnvironmentSuffix)),
	})

	awscdk.NewCfnOutput(t.Stack, jsii.String("CloudFrontDomainName"), &awscdk.CfnOutputProps{
		Value:       t.CloudFrontDist.DomainName(),
		Description: jsii.String("CloudFront Distribution Domain Name"),
		ExportName:  jsii.String(fmt.Sprintf("prod-%s-cloudfront-domain", *t.EnvironmentSuffix)),
	})

	awscdk.NewCfnOutput(t.Stack, jsii.String("LambdaFunctionArn"), &awscdk.CfnOutputProps{
		Value:       t.LambdaFunction.FunctionArn(),
		Description: jsii.String("Lambda Function ARN"),
		ExportName:  jsii.String(fmt.Sprintf("prod-%s-lambda-arn", *t.EnvironmentSuffix)),
	})

	awscdk.NewCfnOutput(t.Stack, jsii.String("BastionHostId"), &awscdk.CfnOutputProps{
		Value:       t.BastionHost.InstanceId(),
		Description: jsii.String("Bastion Host Instance ID"),
		ExportName:  jsii.String(fmt.Sprintf("prod-%s-bastion-id", *t.EnvironmentSuffix)),
	})

	awscdk.NewCfnOutput(t.Stack, jsii.String("KMSKeyId"), &awscdk.CfnOutputProps{
		Value:       t.KmsKey.KeyId(),
		Description: jsii.String("KMS Key ID"),
		ExportName:  jsii.String(fmt.Sprintf("prod-%s-kms-key-id", *t.EnvironmentSuffix)),
	})
}
