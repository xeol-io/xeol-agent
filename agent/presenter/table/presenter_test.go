package table

import (
	"bytes"
	"flag"
	"testing"
	"time"

	"github.com/noqcks/xeol-agent/agent/inventory"

	"github.com/anchore/go-testutils"
	"github.com/sergi/go-diff/diffmatchpatch"
)

var update = flag.Bool("update", false, "update the *.golden files for json presenters")

func TestTablePresenter(t *testing.T) {
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
	}

	pres := NewPresenter(mockReport)

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

	dmp := diffmatchpatch.New()
	if diffs := dmp.DiffMain(string(expected), string(actual), true); len(diffs) > 1 {
		t.Errorf("mismatched output:\n%s\ndiffs:%d", dmp.DiffPrettyText(diffs), len(diffs))
	}
}

func TestEmptyTablePresenter(t *testing.T) {
	// Expected to have no output

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
