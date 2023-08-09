# kubectl-history

🚀 *Time-travel through your cluster* 🕰️

## Installation

```bash
go install github.com/timebertt/kubectl-history@latest
```

## Usage

`kubectl-history` is a [kubectl plugin](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/) and can be invoked as `kubectl history`.

### `kubectl history get`

```bash
$ k history get deploy demo
NAME              REVISION   AGE
demo-75554748b7   1          2m
demo-6b6f8d8b5f   2          43s
demo-7d7cf9bc6    3          5s
```

### `kubectl history diff`

```bash
$ k history diff deploy demo
comparing revisions 2 and 3 of deployment.apps/demo
--- /var/folders/d8/x7ty7dh12sg7vrq374x52pk80000gq/T/deployment.apps_demo-903550842/2-demo-6b6f8d8b5f.yaml	2023-08-09 08:33:47
+++ /var/folders/d8/x7ty7dh12sg7vrq374x52pk80000gq/T/deployment.apps_demo-903550842/3-demo-7d7cf9bc6.yaml	2023-08-09 08:33:47
@@ -4,7 +4,7 @@
     app: demo
 spec:
   containers:
-  - image: nginx:1.24
+  - image: nginx:1.25-alpine
     imagePullPolicy: IfNotPresent
     name: nginx
     resources: {}
```

The `kubectl history diff` command uses `diff -u -N` to compare revisions by default.
It also respects the `KUBECTL_EXTERNAL_DIFF` environment variable like the `kubectl diff` command.
To get a nicer diff view, you can use one of these:

```bash
# add color to the diff output
k history diff deploy demo | colordiff
# specify an external diff programm
KUBECTL_EXTERNAL_DIFF="colordiff -u" k history diff deploy demo
# show diff in VS Code
KUBECTL_EXTERNAL_DIFF="code --diff --wait" k history diff deploy demo
```
