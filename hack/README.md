## Getting Started For Developers

- [Download CDI](#download-cdi)
- [Lint, Test, Build](#lint-test-build)
  - [Make Targets](#make-targets)
  - [Make Variables](#make-variables)
  - [Execute Standard Environment Functional Tests](#execute-standard-environment-functional-tests)
  - [Execute Alternative Environment Functional Tests](#execute-alternative-environment-functional-tests)
- [Submit PRs](#submit-prs)
- [Releases](#releases)
- [Vendoring Dependencies](#vendoring-dependencies)
- [S3-compatible client setup:](#s3-compatible-client-setup)
  - [AWS S3 cli](#aws-s3-cli)
  - [Minio cli](#minio-cli)

### Download CDI

To download the source directly, simply

`$ go get -u kubevirt.io/containerized-data-importer`

### Lint, Test, Build

GnuMake is used to drive a set of scripts that handle linting, testing, compiling, and containerizing.  Executing the scripts directly is not supported at present.

    NOTE: Standard builds require a running Docker daemon!

The standard workflow is performed inside a helper container to normalize the build and test environment for all devs.  Building in the host environment is supported by the Makefile, but is not recommended.

    Docker builds may be disabled by setting DOCKER=0; e.g.
    $ make all DOCKER=0

`$ make all` executes the full workflow.  For granular control of the workflow, several Make targets are defined:

#### Make Targets

- `all`: cleans up previous build artifacts, compiles all CDI packages and builds containers
- `apidocs`: generate client-go code (same as 'make generate') and swagger docs.
- `build-functest`: build the functional tests (content of tests/ subdirectory).
- `bazel-build`: build all the Go binaries used.
- `bazel-build-images`: build all the container images used (for both CDI and functional tests).
- `bazel-generate`: generate BUILD files for Bazel.
- `bazel-push-images`: push the built container images to the registry defined in DOCKER_PREFIX
- `builder-push`: Build and push the builder container image, declared in docker/builder/Dockerfile.
- `clean`: cleans up previous build artifacts
- `cluster-up`: start a default Kubernetes or Open Shift cluster. set KUBEVIRT_PROVIDER environment variable to either 'k8s-1.18' or 'os-3.11.0' to select the type of cluster. set KUBEVIRT_NUM_NODES to something higher than 1 to have more than one node.
- `cluster-down`: stop the cluster, doing a make cluster-down && make cluster-up will basically restart the cluster into an empty fresh state.
- `cluster-down-purge`: cluster-down and cleanup all cached images from docker registry. Accepts [make variables](#make-variables) DOCKER_PREFIX. Removes all images of the specified repository. If not specified removes localhost repository of current cluster instance.
- `cluster-sync`: builds the controller/importer/cloner, and pushes it into a running cluster. The cluster must be up before running a cluster sync. Also generates a manifest and applies it to the running cluster after pushing the images to it.
- `deps-update`: runs 'go mod tidy' and 'go mod vendor'
- `format`: execute 'shfmt', 'goimports', and 'go vet' on all CDI packages.  Writes back to the source files.
- `generate`: generate client-go deepcopy functions, clientset, listers and informers.
- `generate-verify`: generate client-go deepcopy functions, clientset, listers and informers and validate codegen.
- `gomod-update`: Update vendored Go code in vendor/ subdirectory.
- `goveralls`: run code coverage tracking system.
- `manifests`: generate a cdi-controller and operator manifests in '_out/manifests/'.  Accepts [make variables]\(#make-variables\) DOCKER_TAG, DOCKER_PREFIX, VERBOSITY, PULL_POLICY, CSV_VERSION, QUAY_REPOSITORY, QUAY_NAMESPACE
- `openshift-ci-image-push`: Build and push the OpenShift CI build+test container image, declared in hack/ci/Dockerfile.ci
- `push`: compiles, builds, and pushes to the repo passed in 'DOCKER_PREFIX=<my repo>'
- `release-description`: generate a release announcement detailing changes between 2 commits (typically tags).  Expects 'RELREF' and 'PREREF' to be set
- `test`: execute all tests (_NOTE:_ 'WHAT' is expected to match the go cli pattern for paths e.g. './pkg/...'.  This differs slightly from rest of the 'make' targets)
    - `test-unit`: Run unit tests.
    - `test-lint`: Run golangci-lint against src files
    - `test-functional`: Run functional tests (in tests/ subdirectory).
- `vet`: lint all CDI packages


#### Make Variables

Several variables are provided to alter the targets of the above `Makefile` recipes.

These may be passed to a target as `$ make VARIABLE=value target`

- `WHAT`:  The path from the repository root to a target directory (e.g. `make test WHAT=pkg/importer`)
- `DOCKER_PREFIX`: (default: kubevirt) Set repo globally for image and manifest creation
- `DOCKER_TAG`: (default: latest) Set global version tags for image and manifest creation
- `VERBOSITY`: (default: 1) Set global log level verbosity
- `PULL_POLICY`: (default: IfNotPresent) Set global CDI pull policy
- `TEST_ARGS`: A variable containing a list of additional ginkgo flags to be passed to functional tests. The string "--test-args=" must prefix the variable value. For example:

             `make TEST_ARGS="--test-args=-ginkgo.no-color=true" test-functional >& foo`.

  Note: the following extra flags are not supported in TEST_ARGS: -kubeurl, -cdi-namespace, -kubeconfig, -kubectl-path
since these flags are overridden by the _hack/build/run-functional-tests.sh_ script.
To change the default settings for these values the KUBE_URL, CDI_NAMESPACE, KUBECONFIG, and KUBECTL variables, respectively, must be set.
- `RELREF`: Required by `release-description`. Must be a commit or tag.  Should be the more recent than `PREREF`
- `PREREF`: Required by `release-description`. Must also be a commit or tag.  Should be the later than `RELREF`

#### Execute Standard Environment Functional Tests

If using a standard bare-metal/local laptop rhel/kvm environment where nested
virtualization is supported then the standard *kubevirtci framework* can be used.

Environment Variables and Supported Values

| Env Variable       | Default       | Additional Values           |
|--------------------|---------------|-----------------------------|
|KUBEVIRT_PROVIDER   | k8s-1.18      | k8s-1.17, os-3.11.0-crio,   |
|KUBEVIRT_STORAGE*   | none          | ceph, hpp, nfs              |
|KUBEVIRT_PROVIDER_EXTRA_ARGS |      |                             |
|NUM_NODES           | 1             | 2-5                         |

To Run Standard *cluster-up/kubevirtci* Tests
```
 # make cluster-up
 # make cluster-sync
 # make test-functional
```

To run specific functional tests, you can leverage ginkgo command line options as follows:
```
# make TEST_ARGS="--test-args=-ginkgo.focus=<test_suite_name>" test-functional
```
E.g. to run the tests in transport_test.go:
```
# make TEST_ARGS="--test-args=-ginkgo.focus=Transport" test-functional
```

Clean Up
```
 # make cluster-down
```

Clean Up with docker container cache cleanup
To cleanup all container images from local registry and to free a considerable amount of disk space. Note: caveat - cluser-sync will take longer since will have to fetch all the data again 
```
 # make cluster-down-purge
``` 
#### Execute Alternative Environment Functional Tests

If running in a non-standard environment such as Mac or Cloud where the *kubevirtci framework* is
not supported, then you can use the following example to run Functional Tests.

1. Stand-up a Kubernetes cluster (local-up-cluster.sh/kubeadm/minikube/etc...)

2. Clone or get the kubevirt/containerized-data-importer repo

3. Run the CDI controller manifests

   - To generate latest manifests
   ```
   # make manifests 
   ```
   *To customize environment variables see [make targets](#make-targets)*

   - Run the generated latest manifests
     There are two options to deploy cdi directly via cdi-controller.yaml or to deploy it via operator
   ##### Direct deployment
   ```
     #kubectl create -f ./_out/manifests/cdi-controller.yaml
     
     namespace/cdi created
     customresourcedefinition.apiextensions.k8s.io/datavolumes.cdi.kubevirt.io created
     customresourcedefinition.apiextensions.k8s.io/cdiconfigs.cdi.kubevirt.io created
     clusterrole.rbac.authorization.k8s.io/cdi created
     clusterrolebinding.rbac.authorization.k8s.io/cdi-sa created
     clusterrole.rbac.authorization.k8s.io/cdi-apiserver created
     clusterrolebinding.rbac.authorization.k8s.io/cdi-apiserver created
     clusterrolebinding.rbac.authorization.k8s.io/cdi-apiserver-auth-delegator created
     serviceaccount/cdi-sa created
     deployment.apps/cdi-deployment created
     serviceaccount/cdi-apiserver created
     rolebinding.rbac.authorization.k8s.io/cdi-apiserver created
     role.rbac.authorization.k8s.io/cdi-apiserver created
     rolebinding.rbac.authorization.k8s.io/cdi-extension-apiserver-authentication created
     role.rbac.authorization.k8s.io/cdi-extension-apiserver-authentication created
     service/cdi-api created
     deployment.apps/cdi-apiserver created
     service/cdi-uploadproxy created
     deployment.apps/cdi-uploadproxy created

   ```
   ##### Deployment via operator
   ```
     #./cluster-up/kubectl.sh apply -f "./_out/manifests/release/cdi-operator.yaml" 
     namespace/cdi created
     customresourcedefinition.apiextensions.k8s.io/cdis.cdi.kubevirt.io created
     configmap/cdi-operator-leader-election-helper created
     clusterrole.rbac.authorization.k8s.io/cdi.kubevirt.io:operator created
     serviceaccount/cdi-operator created
     clusterrole.rbac.authorization.k8s.io/cdi-operator-cluster-permissions created
     clusterrolebinding.rbac.authorization.k8s.io/cdi-operator created
     deployment.apps/cdi-operator created

     #./cluster-up/kubectl.sh apply -f "./_out/manifests/release/cdi-cr.yaml"
     cdi.cdi.kubevirt.io/cdi created

   ```

4. Build and run the func test servers
   In order to run fucntional tests the below servers have to be run
   - *host-file-server* is required by the functional tests and provides an
     endpoint server for image files and s3 buckets
   - *registry-server* is required by the functional tests and provides an endpoint server for container images. 
     Note: for this server to run the follwoing setting is required in each cluster node 
     ``` systemctl -w user.max_user_namespaces=1024 ```


   Build and Push to registry 
   ```
   # DOCKER_PREFIX=<repo> DOCKER_TAG=<tag> make docker-functest-images
   ```
   Generate manifests
   ```
   # DOCKER_PREFIX=<repo> DOCKER_TAG=<docker tag> PULL_POLICY=<pull policy> VERBOSITY=<verbosity> make manifests 
   ```
   Run servers
   ```
   # ./cluster-up/kubectl.sh apply -f ./_out/manifests/bad-webserver.yaml
   # ./cluster-up/kubectl.sh apply -f ./_out/manifests/test-proxy.yaml
   # ./cluster-up/kubectl.sh apply -f ./_out/manifests/file-host.yaml
   # ./cluster-up/kubectl.sh apply -f ./_out/manifests/registry-host.yaml
   # ./cluster-up/kubectl.sh apply -f ./_out/manifests/imageio.yaml
   ```

5. Run the tests
```
 # make test-functional
```

6. If you encounter test errors and are following the above steps try:
```
 # make clean && make docker
```
redeploy the manifests above, and re-run the tests.

### Submit PRs

All PRs should originate from forks of kubevirt.io/containerized-data-importer.  Work should not be done directly in the upstream repository.  Open new working branches from main/HEAD of your forked repository and push them to your remote repo.  Then submit PRs of the working branch against the upstream main branch.

### Releases

Release practices are described in the [release doc](/doc/releases.md).

### Vendoring Dependencies

This project uses `go modules` as it's dependency manager.  At present, all project dependencies are vendored; using `go mod` is unnecessary in the normal work flow.

`go modules` automatically scans and vendors in dependencies during the build process, you can also manually trigger go modules by running 'make dep-update'.

### S3-compatible client setup:

#### AWS S3 cli
$HOME/.aws/credentials
```
[default]
aws_access_key_id = <your-access-key>
aws_secret_access_key = <your-secret>
```

#### Minio cli

$HOME/.mc/config.json:
```
{
        "version": "8",
        "hosts": {
                "s3": {
                        "url": "https://s3.amazonaws.com",
                        "accessKey": "<your-access-key>",
                        "secretKey": "<your-secret>",
                        "api": "S3v4"
                }
        }
}
```
