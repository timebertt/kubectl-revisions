# kubectl-history

🚀 *Time-travel through your cluster* 🕰️

## About

`kubectl-history` is a [kubectl plugin](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/) and can be invoked as `kubectl history`, or for short `k history`.

The history plugin allows you to go back in time in the history of rollouts and answers common questions like "Why was my Deployment rolled?"

It gives more output than `kubectl rollout history` and is easier to use than `kubectl get replicaset` or `kubectl get controllerrevision`.

## Installation

```bash
go install github.com/timebertt/kubectl-revisions@latest
```

## Usage

### `k history get` / `k history list`

Get the rollout history of a workload resource (`Deployment`, `StatefulSet`, or `DaemonSet`).

The history is based on the `ReplicaSets`/`ControllerRevisions` still in the system. I.e., the history is limited by the
configured `revisionHistoryLimit`.

By default, all revisions are printed as a list. If the `--revision` flag is given, the selected revision is printed
instead.

```bash
$ k history get deploy nginx -owide
NAME               REVISION   AGE   CONTAINERS   IMAGES
nginx-77b4fdf86c   1          22m   nginx        nginx
nginx-7bf8c77b5b   2          21m   nginx        nginx:latest
nginx-7bb88f5ff4   3          20m   nginx        nginx:1.24

$ k history get deploy nginx -r -1 -oyaml
apiVersion: apps/v1
kind: ReplicaSet
metadata:
  name: nginx-7bb88f5ff4
...
```

This is similar to using `k get replicaset` or `k get controllerrevision`, but allows easy selection of the relevant objects and returns a sorted list.
This is also similar to `k rollout history`, but doesn't only print revision numbers.

### `k history diff` / `k history why`

Compare multiple revisions of a workload resource (`Deployment`, `StatefulSet`, or `DaemonSet`).
A.k.a., "Why was my Deployment rolled?"

The history is based on the `ReplicaSets`/`ControllerRevisions` still in the system. I.e., the history is limited by the
configured `revisionHistoryLimit`.

By default, the latest two revisions are compared. The `--revision` flag allows selecting the revisions to compare.

```bash
$ k history diff deploy nginx
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

The `k history diff` command uses `diff -u -N` to compare revisions by default.
It also respects the `KUBECTL_EXTERNAL_DIFF` environment variable like the `kubectl diff` command.
To get a nicer diff view, you can use one of these:

```bash
# Add color to the diff output
k history diff deploy nginx | colordiff
# Specify an external diff programm
KUBECTL_EXTERNAL_DIFF="colordiff -u" k history diff deploy nginx
# Show diff in VS Code
KUBECTL_EXTERNAL_DIFF="code --diff --wait" k history diff deploy nginx
```
