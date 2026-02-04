# json-server-controller
Kubernetes controller that deploys JSON Server from a JsonServer custom resource, wiring a ConfigMap, Deployment, and Service.

## Description
This project provides a JsonServer CRD and controller. Each JsonServer resource creates:
- a ConfigMap holding `db.json`
- a Deployment running JSON Server
- a Service exposing port 3000 inside the cluster

## Getting Started

### Prerequisites
- go version v1.24.6+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.

### To Deploy on a local Kind cluster
**Build the json-server image used by the workload:**

```sh
docker build -t json-server-local:latest -f build/json-server/Dockerfile .
```

**Build the controller image:**

```sh
make docker-build IMG=controller:latest
```

**Load both images into Kind (cluster name `json-server`):**

```sh
kind load docker-image json-server-local:latest --name json-server
kind load docker-image controller:latest --name json-server
```

**Install the CRDs into the cluster:**

```sh
make install
```

**Deploy the controller manager using the local image:**

```sh
make deploy IMG=controller:latest
```

**Create a JsonServer instance:**

```sh
kubectl apply -f config/manifests/jsonserver.yaml
```

**Port-forward the service and verify:**

```sh
kubectl port-forward -n default svc/app-my-server 3000:3000
curl http://localhost:3000/people
```

**Scale the JsonServer via kubectl:**

```sh
kubectl scale jsonserver/app-my-server --replicas=3
kubectl get jsonserver app-my-server -o jsonpath='{.spec.replicas}'
```

**Troubleshooting:**
- If `http://localhost:3000/people` returns 404, ensure the controller is up to date and has reconciled the Deployment (it must mount the ConfigMap `db.json` at `/data/db.json`).
- If port-forward fails, try a different local port: `kubectl port-forward -n default svc/app-my-server 3001:3000`.

### To Deploy on a remote cluster
**Build and push your image to the location specified by `IMG`:**

```sh
make docker-build docker-push IMG=<some-registry>/json-server-controller:tag
```

**NOTE:** This image ought to be published in the personal registry you specified.
And it is required to have access to pull the image from the working environment.
Make sure you have the proper permission to the registry if the above commands don’t work.

**Install the CRDs into the cluster:**

```sh
make install
```

**Deploy the Manager to the cluster with the image specified by `IMG`:**

```sh
make deploy IMG=<some-registry>/json-server-controller:tag
```

> **NOTE**: If you encounter RBAC errors, you may need to grant yourself cluster-admin
privileges or be logged in as admin.

**Create instances of your solution**
Apply the provided manifest:

```sh
kubectl apply -f config/manifests/jsonserver.yaml
```

> **NOTE**: For remote clusters, ensure `spec.image` points to a publicly accessible JSON Server image (or one in your registry).

>**NOTE**: Ensure that the samples has default values to test it out.

## GitOps (local cluster)
This repo includes a GitOps Kustomize entry point at [config/gitops/kustomization.yaml](config/gitops/kustomization.yaml) and an Argo CD Application manifest at [config/gitops/argocd-application.yaml](config/gitops/argocd-application.yaml).

1. Install Argo CD in your local cluster (Kind or similar).
2. Update `spec.source.repoURL` in the Application manifest to your repo URL.
3. Apply the Application:

```sh
kubectl apply -n argocd -f config/gitops/argocd-application.yaml
```

Argo CD will sync the controller and the JsonServer instance from `config/gitops/`.

**Access the Argo CD UI:**

```sh
kubectl -n argocd port-forward svc/argocd-server 8080:443
```

Open https://localhost:8080 and log in with:

```sh
username: admin
password: $(kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath='{.data.password}' | base64 -d)
```

> **NOTE**: The GitOps Kustomization pins images to `ttl.sh`. The TTL is encoded in the tag (e.g., `:1h`, `:2h`), so ensure fresh images are pushed before syncing.

## CI
GitHub Actions workflow is defined in [.github/workflows/ci.yml](.github/workflows/ci.yml). It runs tests and builds/pushes the controller and JSON Server images to `ttl.sh` with short-lived tags (TTL encoded in the tag).

## Testing (envtest)
This project uses controller-runtime's envtest for integration tests. The repository includes a Makefile target that installs the control-plane binaries into `./bin/k8s` so tests can run without sudo.

**Install envtest binaries:**

```sh
make install-kubebuilder
```

**Run tests:**

```sh
make test
```

Notes:
- If `make install-kubebuilder` fails, check network/proxy settings (it downloads official Kubernetes release artifacts).
- The integration tests rely on binaries in `bin/k8s`; do not move them after download.

**Manual test manifests:**
Apply the sample manifests under [config/manifests](config/manifests):

```sh
kubectl apply -k config/manifests
```

This creates a sample JsonServer resource plus the supporting ConfigMap/Deployment/Service for manual testing.

## Development notes
- After modifying API types or markers, run `make manifests generate` and re-install CRDs with `make install`.

### To Uninstall
**Delete the instances (CRs) from the cluster:**

```sh
kubectl delete -f config/manifests/jsonserver.yaml
```

**Delete the APIs(CRDs) from the cluster:**

```sh
make uninstall
```

**UnDeploy the controller from the cluster:**

```sh
make undeploy
```

## Project Distribution

Following the options to release and provide this solution to the users.

### By providing a bundle with all YAML files

1. Build the installer for the image built and published in the registry:

```sh
make build-installer IMG=<some-registry>/json-server-controller:tag
```

**NOTE:** The makefile target mentioned above generates an 'install.yaml'
file in the dist directory. This file contains all the resources built
with Kustomize, which are necessary to install this project without its
dependencies.

2. Using the installer

Users can just run 'kubectl apply -f <URL for YAML BUNDLE>' to install
the project, i.e.:

```sh
kubectl apply -f https://raw.githubusercontent.com/<org>/json-server-controller/<tag or branch>/dist/install.yaml
```

### By providing a Helm Chart

1. Build the chart using the optional helm plugin

```sh
kubebuilder edit --plugins=helm/v2-alpha
```

2. See that a chart was generated under 'dist/chart', and users
can obtain this solution from there.

**NOTE:** If you change the project, you need to update the Helm Chart
using the same command above to sync the latest changes. Furthermore,
if you create webhooks, you need to use the above command with
the '--force' flag and manually ensure that any custom configuration
previously added to 'dist/chart/values.yaml' or 'dist/chart/manager/manager.yaml'
is manually re-applied afterwards.

## Contributing
Issues and pull requests are welcome. Please keep changes focused and add/update tests where appropriate.

**NOTE:** Run `make help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

