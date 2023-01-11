package inventory

import (
	"fmt"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

const (
	defaultIgnoreNotRunning = true
	defaultMissingTagPolicy = "digest"
	defualtDummyTag         = "UNKNOWN"
)

func logout(actual, expected ReportItem, t *testing.T) {
	t.Log("")
	t.Log("Actual")
	for _, container := range actual.Containers {
		t.Logf("  %#v", container)
	}
	t.Log("")
	t.Log("Expected")
	for _, container := range expected.Containers {
		t.Logf("  %#v", container)
	}
	t.Log("")
}

func equivalent(left, right ReportItem) error {
	if left.Namespace != right.Namespace {
		return fmt.Errorf("Namespaces do not match %s != %s", left.Namespace, right.Namespace)
	}

	if len(left.Containers) != len(right.Containers) {
		return fmt.Errorf("Mismatch in number of containers %d != %d", len(left.Containers), len(right.Containers))
	}

	tmap := make(map[string]struct{})
	for _, container := range right.Containers {
		// key := fmt.Sprintf("%s@%s", image.Tag, image.RepoDigest)
		key := container.Name
		tmap[key] = struct{}{}
	}

	for _, container := range left.Containers {
		// key := fmt.Sprintf("%s@%s", image.Tag, image.RepoDigest)
		key := container.Name
		_, exists := tmap[key]
		if !exists {
			return fmt.Errorf("Actual key %s not found in expected results", key)
		}
	}
	return nil
}

// TODO(benji): add more tests here

// Test out NewReportItem with an empty list of pods
func TestNewReportItemEmptyPodList(t *testing.T) {
	namespace := "default"
	mockPods := []v1.Pod{}
	mockDeployments := []appsv1.Deployment{}
	actual := NewReportItem(mockPods, mockDeployments, namespace, defaultIgnoreNotRunning, defaultMissingTagPolicy, defualtDummyTag)

	expected := ReportItem{
		Namespace:  namespace,
		Containers: []Container{},
	}
	err := equivalent(actual, expected)
	if err != nil {
		logout(actual, expected, t)
		t.Error(err)
	}
}
