package inventory

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/anchore/kai/internal/log"
	appsv1 "k8s.io/api/apps/v1"

	v1 "k8s.io/api/core/v1"
)

type Container struct {
	Name  string `json:"name,omitempty"`
	Image Image  `json:"image,omitempty"`
	Pod   Pod    `json:"pod,omitempty"`
}

type Image struct {
	Tag        string `json:"tag,omitempty"`
	RepoDigest string `json:"repoDigest,omitempty"`
}
type Pod struct {
	Name       string     `json:"name,omitempty"`
	Deployment Deployment `json:"deployment,omitempty"`
}

type Deployment struct {
	Name string `json:"name,omitempty"`
}

// ReportItem represents a namespace and all it's unique images
type ReportItem struct {
	Namespace  string      `json:"namespace,omitempty"`
	Containers []Container `json:"containers"`
}

// NewReportItem parses a list of pods into a ReportItem full of unique images
func NewReportItem(pods []v1.Pod, deployments []appsv1.Deployment, namespace string, ignoreNotRunning bool, missingTagPolicy string, dummyTag string) ReportItem {
	reportItem := ReportItem{
		Namespace: namespace,
	}

	var containers []Container
	for _, pod := range pods {
		// Check for non-running
		if ignoreNotRunning && pod.Status.Phase != "Running" {
			continue
		}
		containers = append(containers, processContainer(pod, missingTagPolicy, dummyTag, deployments)...)
	}

	reportItem.Containers = containers
	return reportItem
}

// fillContainerDetails grabs all the relevant fields out of a pod object so
// they can be used to parse out the image details for all the containers in
// a pod. Return details as an mapped array of strings using the container name
// as the map key and the fields as an array of strings so they can be iterated
func fillContainerDetails(pod v1.Pod) map[string][]string {
	details := make(map[string][]string)

	// grab init images
	for _, c := range pod.Spec.InitContainers {
		details[c.Name] = append(details[c.Name], c.Image)
	}

	for _, c := range pod.Status.InitContainerStatuses {
		details[c.Name] = append(details[c.Name], c.Image, c.ImageID)
	}

	// grab regular images
	for _, c := range pod.Spec.Containers {
		details[c.Name] = append(details[c.Name], c.Image)
	}

	for _, c := range pod.Status.ContainerStatuses {
		details[c.Name] = append(details[c.Name], c.Image, c.ImageID)
	}

	return details
}

// image is an intermediate struct for parsing out image details from
// a list of containers
type image struct {
	repo   string
	tag    string
	digest string
}

// // Compile the regexes used for parsing once so they can be reused without having to recompile
var digestRegex = regexp.MustCompile(`@(sha[[:digit:]]{3}:[[:alnum:]]{32,})`)
var tagRegex = regexp.MustCompile(`:[\w][\w.-]{0,127}$`)

// extractImageDetails extracts the repo, tag, and digest of an image out of the fields
// grabbed from the pod.
func (img *image) extractImageDetails(s string) {
	if img.digest != "" && img.tag != "" && img.repo != "" {
		return
	}

	// Attempt to grab the digest out of the string
	// Set repo to the initial string. If there's no digest to parse then we can assume
	// it's just a repo and tag
	repo := s
	digest := ""

	// Look for something like:
	//  k3d-registry.localhost:5000/redis:4@sha256:5bd4fe08813b057df2ae55003a75c39d80a4aea9f1a0fbc0fbd7024edf555786
	digestresult := digestRegex.FindStringSubmatchIndex(repo)
	if len(digestresult) > 0 {
		i := digestresult[0]
		digest = repo[i+1:] // sha256:5bd4fe08813b057df2ae55003a75c39d80a4aea9f1a0fbc0fbd7024edf555786
		repo = repo[:i]     // k3d-registry.localhost:5000/redis:4
	}

	// Attempt to split the repo and tag
	tag := ""

	// repo contains something like
	//  k3d-registry.localhost:5000/redis:4
	tagresult := tagRegex.FindStringSubmatchIndex(repo)
	if len(tagresult) > 0 {
		i := tagresult[0]
		tag = repo[i+1:] // 4
		repo = repo[:i]  // k3d-registry.localhost:5000/redis
	}

	// Only fill if the field hasn't been successfully parsed yet
	if img.digest == "" {
		img.digest = digest
	}

	if img.tag == "" {
		img.tag = tag
	}

	if img.repo == "" {
		img.repo = repo
	}
}

func (img *image) handleMissingTag(missingTagPolicy string, dummyTag string) {
	switch missingTagPolicy {
	case "digest":
		tag := strings.Split(img.digest, ":")
		img.tag = tag[len(tag)-1]
	case "insert":
		img.tag = dummyTag
	}
}

// processContainers takes in a pod object and will return a list of unique
// ReportImage structures from the containers inside the pod
//
// IMPORTANT: Ensures unique images inside a pod
func processContainer(pod v1.Pod, missingTagPolicy string, dummyTag string, deployments []appsv1.Deployment) []Container {
	containerset := fillContainerDetails(pod)
	var containers []Container
	var deployment Deployment
	for _, d := range deployments {
		if IsMapSubset(pod.Labels, d.Spec.Selector.MatchLabels) {
			deployment = Deployment{
				Name: d.Name,
			}
		}
	}

	for containerName, containerData := range containerset {
		img := image{
			repo:   "",
			tag:    "",
			digest: "",
		}

		for _, imgs := range containerData {
			img.extractImageDetails(imgs)
			if img.tag == "" {
				if missingTagPolicy == "drop" {
					log.Debugf("Dropping %s %s due to missing tag policy of 'drop'", img.repo, img.digest)
					continue
				}
				img.handleMissingTag(missingTagPolicy, dummyTag)
			}
		}

		containers = append(containers, Container{
			Name: containerName,
			Image: Image{
				Tag:        fmt.Sprintf("%s:%s", img.repo, img.tag),
				RepoDigest: img.digest,
			},
			Pod: Pod{
				Name:       pod.Name,
				Deployment: deployment,
			},
		})
	}

	return containers
}

// IsMapSubset returns true if sub is a subset of m.
func IsMapSubset[K, V comparable](m, sub map[K]V) bool {
	if len(sub) > len(m) {
		return false
	}
	for k, vsub := range sub {
		if vm, found := m[k]; !found || vm != vsub {
			return false
		}
	}
	return true
}
