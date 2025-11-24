package main

import (
	"os"

	"github.com/TuringGpt/iac-test-automations/lib"
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/jsii-runtime-go"
)

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	// Get environment suffix from context (set by CI/CD pipeline) or use 'dev' as default
	var environmentSuffix string
	if suffix := app.Node().TryGetContext(jsii.String("environmentSuffix")); suffix != nil {
		if suffixStr, ok := suffix.(string); ok {
			environmentSuffix = suffixStr
		} else {
			environmentSuffix = "dev"
		}
	} else {
		environmentSuffix = "dev"
	}

	stackName := "TapStack" + environmentSuffix

	repositoryName := getEnv("REPOSITORY", "unknown")
	commitAuthor := getEnv("COMMIT_AUTHOR", "unknown")

	// Apply tags to all stacks in this app
	awscdk.Tags_Of(app).Add(jsii.String("Environment"), jsii.String(environmentSuffix), nil)
	awscdk.Tags_Of(app).Add(jsii.String("Repository"), jsii.String(repositoryName), nil)
	awscdk.Tags_Of(app).Add(jsii.String("Author"), jsii.String(commitAuthor), nil)

	// Create TapStackProps
	var env *awscdk.Environment
	account := getEnv("CDK_DEFAULT_ACCOUNT", "")
	region := getEnv("CDK_DEFAULT_REGION", "")

	// Only set environment if both account and region are provided
	if account != "" && region != "" {
		env = &awscdk.Environment{
			Account: jsii.String(account),
			Region:  jsii.String(region),
		}
	}

	props := &lib.TapStackProps{
		StackProps: &awscdk.StackProps{
			Env: env,
		},
		EnvironmentSuffix: jsii.String(environmentSuffix),
	}

	// Initialize the stack with proper parameters
	lib.NewTapStack(app, jsii.String(stackName), props)

	app.Synth(nil)
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
