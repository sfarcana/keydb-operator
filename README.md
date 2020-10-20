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
## 3. Contribute

## 4. License
