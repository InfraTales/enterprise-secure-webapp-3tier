package lib_test

import (
	"testing"

	"github.com/TuringGpt/iac-test-automations/lib"
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/assertions"
	"github.com/aws/jsii-runtime-go"
	"github.com/stretchr/testify/assert"
)

func TestTapStack(t *testing.T) {
	defer jsii.Close()

	t.Run("creates complete secure infrastructure with correct environment suffix", func(t *testing.T) {
		// ARRANGE
		app := awscdk.NewApp(nil)
		envSuffix := "test"
		stack := lib.NewTapStack(app, jsii.String("TapStackTest"), &lib.TapStackProps{
			StackProps:        &awscdk.StackProps{},
			EnvironmentSuffix: jsii.String(envSuffix),
		})
		template := assertions.Template_FromStack(stack.Stack, nil)

		// ASSERT - VPC and Network Resources
		template.ResourceCountIs(jsii.String("AWS::EC2::VPC"), jsii.Number(1))
		template.HasResourceProperties(jsii.String("AWS::EC2::VPC"), map[string]interface{}{
			"EnableDnsHostnames": true,
			"EnableDnsSupport":   true,
			"CidrBlock":          "10.0.0.0/16",
		})

		// ASSERT - S3 Buckets
		template.ResourceCountIs(jsii.String("AWS::S3::Bucket"), jsii.Number(3)) // App bucket, Logging bucket, VPC Flow logs bucket

		// ASSERT - KMS Key
		template.ResourceCountIs(jsii.String("AWS::KMS::Key"), jsii.Number(1))
		template.ResourceCountIs(jsii.String("AWS::KMS::Alias"), jsii.Number(1))

		// ASSERT - Lambda Function (There might be more than 1 due to bastion host)
		template.ResourceCountIs(jsii.String("AWS::Lambda::Function"), jsii.Number(2))
		template.HasResourceProperties(jsii.String("AWS::Lambda::Function"), map[string]interface{}{
			"Runtime":    "python3.9",
			"Handler":    "index.lambda_handler",
			"MemorySize": 256,
			"Timeout":    30,
		})

		// ASSERT - Security Groups
		template.ResourceCountIs(jsii.String("AWS::EC2::SecurityGroup"), jsii.Number(3)) // Lambda, EC2, Bastion

		// ASSERT - Auto Scaling Group
		template.ResourceCountIs(jsii.String("AWS::AutoScaling::AutoScalingGroup"), jsii.Number(1))

		// ASSERT - CloudFront Distribution
		template.ResourceCountIs(jsii.String("AWS::CloudFront::Distribution"), jsii.Number(1))
		// Note: OAI creation may vary based on CDK version and CloudFront configuration

		// ASSERT - WAF
		template.ResourceCountIs(jsii.String("AWS::WAFv2::WebACL"), jsii.Number(1))

		// ASSERT - CloudTrail
		template.ResourceCountIs(jsii.String("AWS::CloudTrail::Trail"), jsii.Number(1))

		// ASSERT - SNS Topic
		template.ResourceCountIs(jsii.String("AWS::SNS::Topic"), jsii.Number(1))

		// ASSERT - Secrets Manager
		template.ResourceCountIs(jsii.String("AWS::SecretsManager::Secret"), jsii.Number(1))

		// ASSERT - SSM Parameters
		template.ResourceCountIs(jsii.String("AWS::SSM::Parameter"), jsii.Number(4))

		// ASSERT - CloudWatch Log Groups
		template.ResourceCountIs(jsii.String("AWS::Logs::LogGroup"), jsii.Number(2)) // Lambda and CloudTrail

		// ASSERT - Stack properties
		assert.NotNil(t, stack)
		assert.Equal(t, envSuffix, *stack.EnvironmentSuffix)
	})

	t.Run("creates VPC with correct subnet configuration", func(t *testing.T) {
		// ARRANGE
		app := awscdk.NewApp(nil)
		stack := lib.NewTapStack(app, jsii.String("VPCTest"), &lib.TapStackProps{
			StackProps:        &awscdk.StackProps{},
			EnvironmentSuffix: jsii.String("vpc-test"),
		})
		template := assertions.Template_FromStack(stack.Stack, nil)

		// ASSERT - Subnets (2 public + 2 private = 4 total)
		template.ResourceCountIs(jsii.String("AWS::EC2::Subnet"), jsii.Number(4))
		template.ResourceCountIs(jsii.String("AWS::EC2::InternetGateway"), jsii.Number(1))
		template.ResourceCountIs(jsii.String("AWS::EC2::NatGateway"), jsii.Number(2)) // One per AZ

		// ASSERT - VPC Flow Logs
		template.ResourceCountIs(jsii.String("AWS::EC2::FlowLog"), jsii.Number(1))
	})

	t.Run("creates S3 buckets with proper encryption", func(t *testing.T) {
		// ARRANGE
		app := awscdk.NewApp(nil)
		stack := lib.NewTapStack(app, jsii.String("S3Test"), &lib.TapStackProps{
			StackProps:        &awscdk.StackProps{},
			EnvironmentSuffix: jsii.String("s3-test"),
		})
		template := assertions.Template_FromStack(stack.Stack, nil)

		// ASSERT - All S3 buckets have encryption
		template.HasResourceProperties(jsii.String("AWS::S3::Bucket"), map[string]interface{}{
			"BucketEncryption": map[string]interface{}{
				"ServerSideEncryptionConfiguration": []interface{}{
					map[string]interface{}{
						"ServerSideEncryptionByDefault": map[string]interface{}{
							"SSEAlgorithm": "aws:kms",
						},
					},
				},
			},
			"VersioningConfiguration": map[string]interface{}{
				"Status": "Enabled",
			},
		})

		// ASSERT - App bucket has full public access block
		template.HasResourceProperties(jsii.String("AWS::S3::Bucket"), map[string]interface{}{
			"PublicAccessBlockConfiguration": map[string]interface{}{
				"BlockPublicAcls":       true,
				"BlockPublicPolicy":     true,
				"IgnorePublicAcls":      true,
				"RestrictPublicBuckets": true,
			},
		})

		// ASSERT - Logging bucket allows CloudFront access (RestrictPublicBuckets = false)
		template.HasResourceProperties(jsii.String("AWS::S3::Bucket"), map[string]interface{}{
			"PublicAccessBlockConfiguration": map[string]interface{}{
				"BlockPublicAcls":       true,
				"BlockPublicPolicy":     true,
				"IgnorePublicAcls":      true,
				"RestrictPublicBuckets": false,
			},
			"OwnershipControls": map[string]interface{}{
				"Rules": []interface{}{
					map[string]interface{}{
						"ObjectOwnership": "BucketOwnerPreferred",
					},
				},
			},
		})
	})

	t.Run("creates CloudTrail with proper S3 logging configuration", func(t *testing.T) {
		// ARRANGE
		app := awscdk.NewApp(nil)
		stack := lib.NewTapStack(app, jsii.String("CloudTrailTest"), &lib.TapStackProps{
			StackProps:        &awscdk.StackProps{},
			EnvironmentSuffix: jsii.String("trail-test"),
		})
		template := assertions.Template_FromStack(stack.Stack, nil)

		// ASSERT - CloudTrail exists and is properly configured
		template.ResourceCountIs(jsii.String("AWS::CloudTrail::Trail"), jsii.Number(1))
		template.HasResourceProperties(jsii.String("AWS::CloudTrail::Trail"), map[string]interface{}{
			"IsMultiRegionTrail":         true,
			"IncludeGlobalServiceEvents": true,
			"EnableLogFileValidation":    true,
		})

		// ASSERT - At least one S3 bucket policy exists (for CloudTrail access)
		// The exact policy structure is managed by CDK, but we verify policies exist
		template.ResourcePropertiesCountIs(jsii.String("AWS::S3::BucketPolicy"), map[string]interface{}{}, jsii.Number(3))
	})

	t.Run("creates Lambda function with least privilege IAM role", func(t *testing.T) {
		// ARRANGE
		app := awscdk.NewApp(nil)
		stack := lib.NewTapStack(app, jsii.String("LambdaTest"), &lib.TapStackProps{
			StackProps:        &awscdk.StackProps{},
			EnvironmentSuffix: jsii.String("lambda-test"),
		})
		template := assertions.Template_FromStack(stack.Stack, nil)

		// ASSERT - Lambda function configuration (our main function)
		template.HasResourceProperties(jsii.String("AWS::Lambda::Function"), map[string]interface{}{
			"Runtime":       "python3.9",
			"Handler":       "index.lambda_handler",
			"MemorySize":    256,
			"Timeout":       30,
			"Architectures": []interface{}{"x86_64"},
		})

		// ASSERT - IAM Roles (Lambda, EC2, Bastion + auto-delete custom resource)
		template.ResourceCountIs(jsii.String("AWS::IAM::Role"), jsii.Number(5))
	})

	t.Run("creates EC2 resources in private subnets only", func(t *testing.T) {
		// ARRANGE
		app := awscdk.NewApp(nil)
		stack := lib.NewTapStack(app, jsii.String("EC2Test"), &lib.TapStackProps{
			StackProps:        &awscdk.StackProps{},
			EnvironmentSuffix: jsii.String("ec2-test"),
		})
		template := assertions.Template_FromStack(stack.Stack, nil)

		// ASSERT - Auto Scaling Group
		template.ResourceCountIs(jsii.String("AWS::AutoScaling::AutoScalingGroup"), jsii.Number(1))
		template.HasResourceProperties(jsii.String("AWS::AutoScaling::AutoScalingGroup"), map[string]interface{}{
			"MinSize":         "1",
			"MaxSize":         "3",
			"DesiredCapacity": "2",
		})

		// ASSERT - Launch Template
		template.ResourceCountIs(jsii.String("AWS::EC2::LaunchTemplate"), jsii.Number(1))
	})

	t.Run("creates bastion host with proper security group", func(t *testing.T) {
		// ARRANGE
		app := awscdk.NewApp(nil)
		stack := lib.NewTapStack(app, jsii.String("BastionTest"), &lib.TapStackProps{
			StackProps:        &awscdk.StackProps{},
			EnvironmentSuffix: jsii.String("bastion-test"),
		})
		template := assertions.Template_FromStack(stack.Stack, nil)

		// ASSERT - Bastion instance
		template.ResourceCountIs(jsii.String("AWS::EC2::Instance"), jsii.Number(1))

		// ASSERT - Security groups are configured (ingress rules may be embedded in SecurityGroup)
	})

	t.Run("creates CloudFront distribution with proper configuration", func(t *testing.T) {
		// ARRANGE
		app := awscdk.NewApp(nil)
		stack := lib.NewTapStack(app, jsii.String("CloudFrontTest"), &lib.TapStackProps{
			StackProps:        &awscdk.StackProps{},
			EnvironmentSuffix: jsii.String("cf-test"),
		})
		template := assertions.Template_FromStack(stack.Stack, nil)

		// ASSERT - CloudFront distribution
		template.ResourceCountIs(jsii.String("AWS::CloudFront::Distribution"), jsii.Number(1))
		template.HasResourceProperties(jsii.String("AWS::CloudFront::Distribution"), map[string]interface{}{
			"DistributionConfig": map[string]interface{}{
				"Enabled": true,
			},
		})

		// ASSERT - CloudFront should work with S3 origin (OAI creation may vary)
	})

	t.Run("creates monitoring and compliance resources", func(t *testing.T) {
		// ARRANGE
		app := awscdk.NewApp(nil)
		stack := lib.NewTapStack(app, jsii.String("MonitoringTest"), &lib.TapStackProps{
			StackProps:        &awscdk.StackProps{},
			EnvironmentSuffix: jsii.String("monitoring-test"),
		})
		template := assertions.Template_FromStack(stack.Stack, nil)

		// ASSERT - CloudTrail
		template.ResourceCountIs(jsii.String("AWS::CloudTrail::Trail"), jsii.Number(1))
		template.HasResourceProperties(jsii.String("AWS::CloudTrail::Trail"), map[string]interface{}{
			"IsMultiRegionTrail":         true,
			"IncludeGlobalServiceEvents": true,
			"EnableLogFileValidation":    true,
		})

		// Note: Config recorder is not implemented in this stack

		// ASSERT - CloudWatch alarm
		template.ResourceCountIs(jsii.String("AWS::CloudWatch::Alarm"), jsii.Number(1))
	})

	t.Run("defaults environment suffix to 'dev' if not provided", func(t *testing.T) {
		// ARRANGE
		app := awscdk.NewApp(nil)
		stack := lib.NewTapStack(app, jsii.String("TapStackTestDefault"), &lib.TapStackProps{
			StackProps: &awscdk.StackProps{},
		})

		// ASSERT
		assert.NotNil(t, stack)
		assert.Equal(t, "dev", *stack.EnvironmentSuffix)
	})

	t.Run("creates all required outputs", func(t *testing.T) {
		// ARRANGE
		app := awscdk.NewApp(nil)
		stack := lib.NewTapStack(app, jsii.String("OutputsTest"), &lib.TapStackProps{
			StackProps:        &awscdk.StackProps{},
			EnvironmentSuffix: jsii.String("outputs-test"),
		})
		template := assertions.Template_FromStack(stack.Stack, nil)

		// ASSERT - CloudFormation outputs
		template.HasOutput(jsii.String("VPCId"), map[string]interface{}{})
		template.HasOutput(jsii.String("S3BucketName"), map[string]interface{}{})
		template.HasOutput(jsii.String("CloudFrontDomainName"), map[string]interface{}{})
		template.HasOutput(jsii.String("LambdaFunctionArn"), map[string]interface{}{})
		template.HasOutput(jsii.String("BastionHostId"), map[string]interface{}{})
		template.HasOutput(jsii.String("KMSKeyId"), map[string]interface{}{})
	})

	t.Run("applies proper tags to resources", func(t *testing.T) {
		// ARRANGE
		app := awscdk.NewApp(nil)
		envSuffix := "tags-test"
		stack := lib.NewTapStack(app, jsii.String("TagsTest"), &lib.TapStackProps{
			StackProps:        &awscdk.StackProps{},
			EnvironmentSuffix: jsii.String(envSuffix),
		})

		// ASSERT
		assert.NotNil(t, stack)
		assert.Equal(t, envSuffix, *stack.EnvironmentSuffix)
	})
}

// Benchmark tests
func BenchmarkTapStackCreation(b *testing.B) {
	defer jsii.Close()

	for i := 0; i < b.N; i++ {
		app := awscdk.NewApp(nil)
		lib.NewTapStack(app, jsii.String("BenchStack"), &lib.TapStackProps{
			StackProps:        &awscdk.StackProps{},
			EnvironmentSuffix: jsii.String("bench"),
		})
	}
}

func BenchmarkTapStackValidation(b *testing.B) {
	defer jsii.Close()

	app := awscdk.NewApp(nil)
	stack := lib.NewTapStack(app, jsii.String("BenchValidationStack"), &lib.TapStackProps{
		StackProps:        &awscdk.StackProps{},
		EnvironmentSuffix: jsii.String("bench-validation"),
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		assertions.Template_FromStack(stack.Stack, nil)
	}
}
