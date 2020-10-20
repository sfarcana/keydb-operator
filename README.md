# KeyDB Operator
KeyDB Operator creates/configures/manages keydb services atop Kubernetes.

## 1. Installation
Ref: https://v0-19-x.sdk.operatorframework.io/docs/golang/installation/#additional-prerequisites

## 2. How to build
Ref: https://v0-19-x.sdk.operatorframework.io/docs/golang/quickstart/

- Generate CRD manifests:
```bash
make manifest
```

- After modifying the file has suffix `_types.go`, we always run command to update the generated code for that resource type:
```bash
make generate
```

- Before running the operator, the CRD must be registered with the Kubernetes apiserver:
```bash
make install
```
## 3. How to test
We test on minikube

- Install CRD in your minikube, require do steps `make manifest` and `make generate`:
```bash
make install
```

- At local, we run command:
```bash
make run ENABLE_WEBHOOKS=false
```

and finally, we will see like that, it's ready
```bash
$ make run ENABLE_WEBHOOKS=false
go: creating new go.mod: module tmp
go: found sigs.k8s.io/controller-tools/cmd/controller-gen in sigs.k8s.io/controller-tools v0.3.0
/Users/lap01439/go/bin/controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."
go fmt ./...
go vet ./...
/Users/lap01439/go/bin/controller-gen "crd:trivialVersions=true" rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases
go run ./main.go
2020-10-20T16:15:50.860+0700	INFO	controller-runtime.metrics	metrics server is starting to listen	{"addr": ":8080"}
2020-10-20T16:15:50.861+0700	INFO	setup	starting manager
2020-10-20T16:15:50.861+0700	INFO	controller-runtime.manager	starting metrics server	{"path": "/metrics"}
2020-10-20T16:15:50.861+0700	INFO	controller-runtime.controller	Starting EventSource	{"controller": "keydb", "source": "kind source: /, Kind="}
2020-10-20T16:15:50.962+0700	INFO	controller-runtime.controller	Starting EventSource	{"controller": "keydb", "source": "kind source: /, Kind="}
2020-10-20T16:15:51.065+0700	INFO	controller-runtime.controller	Starting Controller	{"controller": "keydb"}
2020-10-20T16:15:51.065+0700	INFO	controller-runtime.controller	Starting workers	{"controller": "keydb", "worker count": 2}


```

## 4. Contribute

## 5. License
