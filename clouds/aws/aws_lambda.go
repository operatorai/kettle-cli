package aws

import (
	"fmt"

	"github.com/janeczku/go-spinner"
	"github.com/operatorai/operator/command"
	"github.com/operatorai/operator/config"
)

type AWSLambdaFunction struct{}

func (AWSLambdaFunction) Deploy(directory string, cfg *config.TemplateConfig) error {
	fmt.Println("🚢  Deploying ", cfg.Name, "as an AWS Lambda function")
	fmt.Println("⏭  Entry point: ", cfg.FunctionName, fmt.Sprintf("(%s)", cfg.Runtime))

	deploymentArchive, err := createDeploymentArchive(cfg)
	if err != nil {
		return err
	}

	var waitType string
	exists, err := lambdaFunctionExists(cfg.Name)
	if err != nil {
		return err
	}
	if exists {
		waitType, err = updateLambda(deploymentArchive, cfg)
		if err != nil {
			return err
		}
	} else {
		waitType, err = createLambda(deploymentArchive, cfg)
		if err != nil {
			return err
		}
	}
	return waitForLambda(waitType, cfg)
}

func lambdaFunctionExists(name string) (bool, error) {
	s := spinner.StartNew(fmt.Sprintf("Checking if: %s exists...", name))
	defer s.Stop()
	err := command.Execute("aws", []string{
		"lambda",
		"get-function",
		"--function-name",
		name,
	}, true)
	if err != nil {
		return false, err
	}
	return true, nil
}

func updateLambda(deploymentArchive string, cfg *config.TemplateConfig) (string, error) {
	err := command.Execute("aws", []string{
		"lambda",
		"update-function-code",
		"--function-name", cfg.Name,
		"--zip-file", fmt.Sprintf("fileb://%s", deploymentArchive),
	}, false)
	if err != nil {
		return "", err
	}
	return "function-updated", nil
}

func createLambda(deploymentArchive string, cfg *config.TemplateConfig) (string, error) {
	err := setExecutionRole(cfg)
	if err != nil {
		return "", err
	}
	err = command.Execute("aws", []string{
		"lambda",
		"create-function",
		"--function-name", cfg.Name,
		"--runtime", cfg.Runtime,
		"--role", cfg.RoleArn,
		"--handler", fmt.Sprintf("main.%s", cfg.FunctionName),
		"--package-type", "Zip",
		"--zip-file", fmt.Sprintf("fileb://%s", deploymentArchive),
	}, false)
	if err != nil {
		return "", err
	}

	err = setApiGatewayResource(cfg)
	if err != nil {
		return "", err
	}
	// Set the Lambda function as the destination for the POST method
	// Deploy the API
	// Grant invoke permission to the API
	// $ curl -X POST -d "{\"operation\":\"create\",\"tableName\":\"lambda-apigateway\",\"payload\":{\"Item\":{\"id\":\"1\",\"name\":\"Bob\"}}}" https://$API.execute-api.$REGION.amazonaws.com/prod/DynamoDBManager
	return "function-active", nil
}

func waitForLambda(waitType string, cfg *config.TemplateConfig) error {
	s := spinner.StartNew(fmt.Sprintf("Deploying. Waiting for: %s", waitType))
	defer s.Stop()
	return command.Execute("aws", []string{
		"lambda",
		"wait",
		waitType,
		"--function-name",
		cfg.Name,
	}, false)
}
