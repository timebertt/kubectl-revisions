## kubectl revisions diff

Compare multiple revisions of a workload resource

### Synopsis

Compare multiple revisions of a workload resource (Deployment, StatefulSet, or DaemonSet).
A.k.a., "Why was my Deployment rolled?"

The history is based on the ReplicaSets/ControllerRevisions still in the system. I.e., the history is limited by the
configured revisionHistoryLimit.

By default, the latest two revisions are compared. The --revision flag allows selecting the revisions to compare.

The `KUBECTL_EXTERNAL_DIFF` environment variable can be used to select your own diff command. Users can use external
commands with params too, e.g.: `KUBECTL_EXTERNAL_DIFF="colordiff -N -u"`

By default, the `diff` command available in your path will be run with the `-u` (unified diff) and `-N` (treat absent
files as empty) options.

```
kubectl revisions diff (TYPE[.VERSION][.GROUP] NAME | TYPE[.VERSION][.GROUP]/NAME) [flags]
```

### Examples

```
# Find out why the nginx Deployment was rolled: compare the latest two revisions
kubectl revisions diff deploy nginx

# Compare the first and third revision
kubectl revisions diff deploy nginx --revision=1,3

# Compare the previous revision and the revision before that
kubectl revisions diff deploy nginx --revision=-2

# Use a colored external diff program
KUBECTL_EXTERNAL_DIFF="colordiff -u" kubectl revisions diff deploy nginx

# Use dyff as a rich diff program
KUBECTL_EXTERNAL_DIFF="dyff between --omit-header" kubectl revisions diff deploy nginx

# Show diff in VS Code
KUBECTL_EXTERNAL_DIFF="code --diff --wait" kubectl revisions diff deploy nginx

```

### Options

```
      --allow-missing-template-keys   If true, ignore any errors in templates when a field or map key is missing in the template. Only applies to golang and jsonpath output formats. (default true)
  -h, --help                          help for diff
  -o, --output string                 Output format. One of: (json, yaml, go-template, go-template-file, template, templatefile, jsonpath, jsonpath-as-json, jsonpath-file). See golang template [https://golang.org/pkg/text/template/#pkg-overview] and jsonpath template [https://kubernetes.io/docs/reference/kubectl/jsonpath/]. (default "yaml")
  -r, --revision int64Slice           Compare the specified revision with its predecessor. Specify -1 for the latest revision, -2 for the one before the latest, etc.
                                      If given twice, compare the specified two revisions. If not given, compare the latest two revisions. (default [])
      --show-managed-fields           If true, keep the managedFields when printing objects in JSON or YAML format.
      --template string               Template string or path to template file to use when -o=go-template, -o=go-template-file. The template format is golang templates [http://golang.org/pkg/text/template/#pkg-overview].
      --template-only                 If false, print the full revision object (e.g., ReplicaSet) instead of only the pod template. (default true)
```

### Options inherited from parent commands

```
      --as string                      Username to impersonate for the operation. User could be a regular user or a service account in a namespace.
      --as-group stringArray           Group to impersonate for the operation, this flag can be repeated to specify multiple groups.
      --as-uid string                  UID to impersonate for the operation.
      --cache-dir string               Default cache directory (default "$HOME/.kube/cache")
      --certificate-authority string   Path to a cert file for the certificate authority
      --client-certificate string      Path to a client certificate file for TLS
      --client-key string              Path to a client key file for TLS
      --cluster string                 The name of the kubeconfig cluster to use
      --context string                 The name of the kubeconfig context to use
      --disable-compression            If true, opt-out of response compression for all requests to the server
      --insecure-skip-tls-verify       If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
      --kubeconfig string              Path to the kubeconfig file to use for CLI requests.
  -n, --namespace string               If present, the namespace scope for this CLI request
      --password string                Password for basic authentication to the API server
      --request-timeout string         The length of time to wait before giving up on a single server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests. (default "0")
  -s, --server string                  The address and port of the Kubernetes API server
      --tls-server-name string         Server name to use for server certificate validation. If it is not provided, the hostname used to contact the server is used
      --token string                   Bearer token for authentication to the API server
      --user string                    The name of the kubeconfig user to use
      --username string                Username for basic authentication to the API server
```

### SEE ALSO

* [kubectl revisions](kubectl_revisions.md)	 - Time-travel through your workload revision history

