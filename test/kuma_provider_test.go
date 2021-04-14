package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/docker"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

func TestKumaProvider(t *testing.T) {
	terraformOptions := &terraform.Options{
		// Set the path to the Terraform code that will be tested.
		TerraformBinary: "terraform",
		TerraformDir:    "../examples",
	}
	dockerOptions := &docker.Options{
		WorkingDir: "../",
	}

	// Clean up resources at the end of the test (kuma resources, then containers).
	defer docker.RunDockerCompose(t, dockerOptions, "down")
	defer terraform.Destroy(t, terraformOptions)

	// docker.RunDockerCompose(t, dockerOptions, "up", "-d")
	// Run "terraform init" and "terraform apply". Fail the test if there are any errors.
	terraform.InitAndApply(t, terraformOptions)

	permissionName := terraform.Output(t, terraformOptions, "test_permission_name")
	retryName := terraform.Output(t, terraformOptions, "test_retry_name")
	circuitBreakerName := terraform.Output(t, terraformOptions, "test_circuit_breaker_name")

	assert.Equal(t, "test_permission", permissionName)
	assert.Equal(t, "test_retry", retryName)
	assert.Equal(t, "test_circuit_breaker", circuitBreakerName)


}
