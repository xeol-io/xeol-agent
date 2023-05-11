package json

import (
	"bytes"
	"flag"
	"testing"
	"time"

	"github.com/xeol-io/xeol-agent/agent/inventory"
	"k8s.io/apimachinery/pkg/version"

	"github.com/anchore/go-testutils"
	"github.com/sergi/go-diff/diffmatchpatch"
)

var update = flag.Bool("update", false, "update the *.golden files for json presenters")

func TestJsonPresenter(t *testing.T) {
	var buffer bytes.Buffer

	var item1 = inventory.ReportItem{
		Namespace: "docker",
		Containers: []inventory.Container{
			{
				Name: "docker",
				Image: inventory.Image{
					Tag:        "docker/kube-compose-controller:v0.4.25-alpha1",
					RepoDigest: "sha256:6ad2d6a2cc1909fbc477f64e3292c16b88db31eb83458f420eb223f119f3dffd",
				},
				Pod: inventory.Pod{
					Name: "docker",
					Deployment: inventory.Deployment{
						Name: "docker",
					},
				},
			},
		},
	}

	var item2 = inventory.ReportItem{
		Namespace: "kube-system",
		Containers: []inventory.Container{
			{
				Name: "kube-system",
				Image: inventory.Image{
					Tag:        "docker/kube-compose-controller:v0.4.25-alpha1",
					RepoDigest: "sha256:6ad2d6a2cc1909fbc477f64e3292c16b88db31eb83458f420eb223f119f3dffd",
				},
				Pod: inventory.Pod{
					Name: "kube-system",
					Deployment: inventory.Deployment{
						Name: "kube-system",
					},
				},
			},
		},
	}

	var testTime = time.Date(2020, time.September, 18, 11, 00, 49, 0, time.UTC)
	var mockReport = inventory.Report{
		Timestamp: testTime.Format(time.RFC3339),
		Results:   []inventory.ReportItem{item1, item2},
		ServerVersionMetadata: &version.Info{
			Major:        "1",
			Minor:        "16+",
			GitVersion:   "v1.16.6-beta.0",
			GitCommit:    "e7f962ba86f4ce7033828210ca3556393c377bcc",
			GitTreeState: "clean",
			BuildDate:    "2020-01-15T08:18:29Z",
			GoVersion:    "go1.13.5",
			Compiler:     "gc",
			Platform:     "linux/amd64",
		},
		ClusterName:   "docker-desktop",
		InventoryType: "kubernetes",
	}

	pres := NewPresenter(mockReport)

	// run presenter
	if err := pres.Present(&buffer); err != nil {
		t.Fatal(err)
	}
	actual := buffer.Bytes()
	if *update {
		testutils.UpdateGoldenFileContents(t, actual)
	}

	var expected = testutils.GetGoldenFileContents(t)

	if !bytes.Equal(expected, actual) {
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(string(expected), string(actual), true)
		t.Errorf("mismatched output:\n%s", dmp.DiffPrettyText(diffs))
	}
}

func TestEmptyJsonPresenter(t *testing.T) {
	// Expected to have an empty JSON object back
	var buffer bytes.Buffer

	pres := NewPresenter(inventory.Report{})

	// run presenter
	err := pres.Present(&buffer)
	if err != nil {
		t.Fatal(err)
	}
	actual := buffer.Bytes()
	if *update {
		testutils.UpdateGoldenFileContents(t, actual)
	}

	var expected = testutils.GetGoldenFileContents(t)

	if !bytes.Equal(expected, actual) {
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(string(expected), string(actual), true)
		t.Errorf("mismatched output:\n%s", dmp.DiffPrettyText(diffs))
	}
}

func TestNoResultsJsonPresenter(t *testing.T) {
	// Expected to have an empty JSON object back
	var buffer bytes.Buffer

	var testTime = time.Date(2020, time.September, 18, 11, 00, 49, 0, time.UTC)
	pres := NewPresenter(inventory.Report{
		Timestamp:     testTime.Format(time.RFC3339),
		Results:       []inventory.ReportItem{},
		ClusterName:   "docker-desktop",
		InventoryType: "kubernetes",
	})

	// run presenter
	err := pres.Present(&buffer)
	if err != nil {
		t.Fatal(err)
	}
	actual := buffer.Bytes()
	if *update {
		testutils.UpdateGoldenFileContents(t, actual)
	}

	var expected = testutils.GetGoldenFileContents(t)

	if !bytes.Equal(expected, actual) {
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(string(expected), string(actual), true)
		t.Errorf("mismatched output:\n%s", dmp.DiffPrettyText(diffs))
	}
}
