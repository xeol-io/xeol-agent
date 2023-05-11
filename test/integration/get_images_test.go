package integration

import (
	"strings"
	"testing"

	agent "github.com/xeol-io/xeol-agent/agent"
	"github.com/xeol-io/xeol-agent/cmd"
)

const IntegrationTestNamespace = "xeol-agent-integration-test"
const IntegrationTestImageTag = "nginx:latest"

// Assumes that the hello-world helm chart in ./fixtures was installed (basic nginx container)
func TestGetImageResults(t *testing.T) {
	cmd.InitAppConfig()
	report, err := agent.GetInventoryReport(cmd.GetAppConfig())
	if err != nil {
		t.Fatalf("failed to get image results: %v", err)
	}

	if report.ServerVersionMetadata == nil {
		t.Errorf("Failed to include Server Version Metadata in report")
	}

	if report.Timestamp == "" {
		t.Errorf("Failed to include Timestamp in report")
	}

	foundIntegrationTestNamespace := false
	for _, item := range report.Results {
		if item.Namespace != IntegrationTestNamespace {
			continue
		} else {
			foundIntegrationTestNamespace = true
			foundIntegrationTestImage := false
			for _, container := range item.Containers {
				if !strings.Contains(container.Image.Tag, IntegrationTestImageTag) {
					continue
				} else {
					foundIntegrationTestImage = true
					if container.Image.RepoDigest == "" {
						t.Logf("Image Found, but no digest located: %v", container.Image)
					}
				}
			}
			if !foundIntegrationTestImage {
				t.Errorf("failed to locate integration test image")
			}
		}
	}
	if !foundIntegrationTestNamespace {
		t.Errorf("failed to locate integration test namespace")
	}
}
