# kubectl-revisions

🚀 *Time-travel through your workload revision history* 🕰️

## About

`kubectl-revisions` is a [kubectl plugin](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/) and can be invoked as `kubectl revisions`, or for short `k revisions`.

The `revisions` plugin allows you to go back in time in the history of revisions and answers common questions like "Why was my Deployment rolled?"

It gives more output than `kubectl rollout history` and is easier to use than `kubectl get replicaset` or `kubectl get controllerrevision`.

## Installation

Using [krew](https://krew.sigs.k8s.io/) (recommended):

```bash
kubectl krew install revisions
```

Using go:

```bash
go install github.com/timebertt/kubectl-revisions@latest
```

Learn how to set up shell completion:

```bash
kubectl revisions completion -h
```

## Usage

### `k revisions get` / `k revisions list`

Get the revision history of a workload resource (`Deployment`, `StatefulSet`, or `DaemonSet`).

The history is based on the `ReplicaSets`/`ControllerRevisions` still in the system. I.e., the history is limited by the
configured `revisionHistoryLimit`.

By default, all revisions are printed as a list. If the `--revision` flag is given, the selected revision is printed
instead.

```bash
$ k revisions get deploy nginx -owide
NAME               REVISION   AGE   CONTAINERS   IMAGES
nginx-77b4fdf86c   1          22m   nginx        nginx
nginx-7bf8c77b5b   2          21m   nginx        nginx:latest
nginx-7bb88f5ff4   3          20m   nginx        nginx:1.24

$ k revisions get deploy nginx -r -1 -oyaml
apiVersion: apps/v1
kind: ReplicaSet
metadata:
  name: nginx-7bb88f5ff4
...
```

This is similar to using `k get replicaset` or `k get controllerrevision`, but allows easy selection of the relevant objects and returns a sorted list.
This is also similar to `k rollout history`, but doesn't only print revision numbers.

### `k revisions diff` / `k revisions why`

Compare multiple revisions of a workload resource (`Deployment`, `StatefulSet`, or `DaemonSet`).
A.k.a., "Why was my Deployment rolled?"

The history is based on the `ReplicaSets`/`ControllerRevisions` still in the system. I.e., the history is limited by the
configured `revisionHistoryLimit`.

By default, the latest two revisions are compared. The `--revision` flag allows selecting the revisions to compare.

```bash
$ k revisions diff deploy nginx
comparing revisions 2 and 3 of deployment.apps/nginx
--- /var/folders/d8/x7ty7dh12sg7vrq374x52pk80000gq/T/deployment.apps_nginx-2577026088/2-nginx-7bf8c77b5b.yaml	2024-05-22 23:16:51
+++ /var/folders/d8/x7ty7dh12sg7vrq374x52pk80000gq/T/deployment.apps_nginx-2577026088/3-nginx-7bb88f5ff4.yaml	2024-05-22 23:16:51
@@ -7,7 +7,7 @@
     app: nginx
 spec:
   containers:
-  - image: nginx:latest
+  - image: nginx:1.24
     imagePullPolicy: Always
     name: nginx
     resources: {}
```

The `k revisions diff` command uses `diff -u -N` to compare revisions by default.
It also respects the `KUBECTL_EXTERNAL_DIFF` environment variable like the `kubectl diff` command.
To get a nicer diff output, you can use one of these:

```bash
# Use a colored external diff program
export KUBECTL_EXTERNAL_DIFF="colordiff -u"
# Use dyff as a rich diff program
export KUBECTL_EXTERNAL_DIFF="dyff between --omit-header"
# Show diff in VS Code
export KUBECTL_EXTERNAL_DIFF="code --diff --wait"
```
