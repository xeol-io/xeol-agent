# xeol-agent

[![Go Report Card](https://goreportcard.com/badge/github.com/noqcks/xeol-agent)](https://goreportcard.com/report/github.com/noqcks/xeol-agent)
[![GitHub release](https://img.shields.io/github/release/noqcks/xeol-agent.svg)](https://github.com/noqcks/xeol-agent/releases/latest)
[![License: Apache-2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/noqcks/xeol-agent/blob/main/LICENSE)

The xeol-agent poll the Kubernetes API on an interval to retrieve inventory data about the
the cluster and reports back to xeol.io.

It can be run inside a cluster (under a Service Account) or outside (via any provided Kubeconfig).

## Getting Started

[Install the binary](#installation) or Download the [Docker image](https://hub.docker.com/repository/docker/noqcks/xeol-agent)

## Installation

xeol-agent can be installed via the Helm Chart
### Helm Chart

xeol-agent runs as a read-only service account in the cluster it's deployed to.

In order to report the inventory to xeol.io, xeol-agent does require an api key for your xeol.io organization.
xeol-agent's helm chart automatically creates a kubernetes secret for the xeol.io apiKey
based on the values file you use, Ex.:

```yaml
xeolAgent:
  xeol:
    apiKey: foobar
```

It will set the following environment variable based on this: `XEOL_AGENT_API_KEY=foobar`.

If you don't want to store your xeol ApiKey in the values file, you can create your own secret to do this:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: xeol-agent-api-key
type: Opaque
stringData:
  XEOL_AGENT_API_KEY: foobar
```

and then provide it to the helm chart via the values file:

```yaml
xeolAgent:
  existingSecret: xeol-agent-api-key
```

xeol-agent's helm chart is part of the [charts.xeol.io](https://charts.xeol.io) repo. You can install it via:

```sh
helm repo add xeol https://charts.xeol.io
helm install <release-name> -f <values.yaml> noqcks/xeol-agent
```

A basic values file can always be found [here](https://github.com/noqcks/xeol-charts/tree/master/stable/xeol-agent/values.yaml)

## Configuration
```yaml
# same as -o ; the output format (options: table, json)
output: "json"

# same as -q ; suppress all output (except for the inventory results)
quiet: false

log:
  # use structured logging
  structured: false

  # the log level; note: detailed logging suppress the ETUI
  level: "debug"

  # location to write the log file (default is not to have a log file)
  file: "./xeol-agent.log"

# enable/disable checking for application updates on startup
check-for-app-update: true

kubeconfig:
  path:
  cluster: docker-desktop
  cluster-cert:
  server:  # ex. https://kubernetes.docker.internal:6443
  user:
    type:  # valid: [private_key, token]
    client-cert:
    private-key:
    token:
```

### Namespace selection

Configure which namespaces xeol-agent should search.

* `include` section
  * A list of explicit strings that will detail the list of namespaces to capture image data from.
  * If left as an empty list `[]` all namespaces will be searched
  * Example:

```yaml
namespace-selectors:
  include:
  - default
  - kube-system
  - prod-app
```

* `exclude` section
  * A list of explicit strings and/or regex patterns for namespaces to be excluded.
  * A regex is determined if the string does not match standard DNS name requirements.
  * Example:

```yaml
namespace-selectors:
  exclude:
  - default
  - ^kube-*
  - ^prod-*
```

```yaml
# Which namespaces to search or exclude.
namespace-selectors:
  # Namespaces to include as explicit strings, not regex
  # NOTE: Will search ALL namespaces if left as an empty array
  include: []

  # List of namespaces to exclude, can use explicit strings and/or regexes.
  # For example
  #
  # list:
  # - default
  # - ^kube-*
  #
  # Will exclude the default, kube-system, and kube-public namespaces
  exclude: []
```

### Kubernetes API Parameters

This section will allow users to tune the way xeol-agent interacts with the kubernetes API server.

```yaml
# Kubernetes API configuration parameters (should not need tuning)
kubernetes:
  # Sets the request timeout for kubernetes API requests
  request-timeout-seconds: 60

  # Sets the number of objects to iteratively return when listing resources
  request-batch-size: 100

  # Worker pool size for collecting pods from namespaces. Adjust this if the api-server gets overwhelmed
  worker-pool-size: 100
```

### xeol-agent mode of operation

```yaml
# Can be one of adhoc, periodic (defaults to adhoc)
mode: adhoc

# Only respected if mode is periodic
polling-interval-seconds: 300
```

### Missing Tag Policy

There are cases where images in Kubernetes do not have an associated tag - for
example when an image is deployed using the digest.

```sh
kubectl run python --image=python@sha256:f0a210a37565286ecaaac0529a6749917e8ea58d3dfc72c84acfbfbe1a64a20a
```

xeol.io will use the image digest to process an image but it still requires a tag to be
associated with the image. The `missing-tag-policy` lets you configure the best way to handle the
missing tag edge case in your environment.

**digest** will use the image digest as a dummy tag.
```json
{
  "tag": "alpine:4ed1812024ed78962a34727137627e8854a3b414d19e2c35a1dc727a47e16fba",
  "repoDigest": "sha256:4ed1812024ed78962a34727137627e8854a3b414d19e2c35a1dc727a47e16fba"
}
```

**insert** will use a dummy tag configured by `missing-tag-policy.tag`
```json
{
  "tag": "alpine:UNKNOWN",
  "repoDigest": "sha256:4ed1812024ed78962a34727137627e8854a3b414d19e2c35a1dc727a47e16fba"
}
```

**drop** will simply ignore the images that don't have tags.


```yaml
# Handle cases where a tag is missing. For example - images designated by digest
missing-tag-policy:
  # One of the following options [digest, insert, drop]. Default is 'digest'
  #
  # [digest] will use the image's digest as a dummy tag.
  #
  # [insert] will insert a default tag in as a dummy tag. The dummy tag is
  #          customizable under missing-tag-policy.tag
  #
  # [drop] will drop images that do not have tags associated with them. Not
  #        recommended.
  policy: digest

  # Dummy tag to use. Only applicable if policy is 'insert'. Defaults to UNKNOWN
  tag: UNKNOWN
```

### Ignore images that are not yet in a Running state

```yaml
# Ignore images out of pods that are not in a Running state
ignore-not-running: true
```

### xeol.io API configuration

Use this section to configure the xeol.io API endpoint

```yaml
xeol:
  api-key: $XEOL_API_KEY
  http:
    insecure: true
    timeout-seconds: 10
```

## Configuration Changes (v0.2.2 -> v0.3.0)

There are a few configurations that were changed from v0.2.2 to v0.3.0

#### `kubernetes-request-timeout-seconds`

The request timeout for the kubernetes API was changed from

```yaml
kubernetes-request-timeout-seconds: 60
```

to

```yaml
kubernetes:
  request-timeout-seconds: 60
```

xeol-agent will still honor the old configuration. It will prefer the old configuration
parameter until it is removed from the config entirely. It is safe to remove the
old configuration in favor of the new config.

#### `namespaces`

The namespace configuration was changed from

```yaml
namespaces:
- all
```

to

```yaml
namespace-selectors:
  include: []
  exclude: []
```

`namespace-selectors` was added to eventually replace `namespaces` to allow for both
include and exclude configs. The old `namespaces` array will be honored if
`namespace-selectors.include` is empty. It is safe to remove `namespaces` entirely
in favor of `namespace-selectors`

## Developing
### Build
**Note:** This will drop the binary in the `./snapshot/` directory

**On Mac**
```sh
make mac-binary
```

**On Linux**
```sh
make linux-binary
```

### Testing

The Makefile has testing built into it. For unit tests simply run

```sh
make unit
```

### Docker
To build a docker image, you'll need to provide a kubeconfig.

Note: Docker build requires files to be within the docker build context

```sh
docker build -t localhost/xeol-agent:latest --build-arg KUBECONFIG=./kubeconfig .
```

## Releasing
To create a release of xeol-agent, a tag needs to be created that points to a commit in `main`
that we want to release. This tag shall be a semver prefixed with a `v`, e.g. `v0.2.7`.
This will trigger a GitHub Action that will create the release.

After the release has been successfully created, make sure to specify the updated version
in both Enterprise and the xeol-agent Helm Chart in
[xeol-charts](https://github.com/noqcks/xeol-charts).
