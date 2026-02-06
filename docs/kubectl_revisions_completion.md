## kubectl revisions completion

Setup shell completion

### Synopsis

The completion command outputs a script which makes the revisions plugin's completion available to kubectl's completion
(supported in kubectl v1.26+), see https://github.com/kubernetes/kubernetes/pull/105867 and
https://github.com/kubernetes/sample-cli-plugin#shell-completion.

This script needs to be installed as an executable file in PATH named kubectl_complete-revisions. E.g., you could
install it in krew's binary directory. This is not supported natively yet, but can be done manually as follows
(see https://github.com/kubernetes-sigs/krew/issues/812):
```
SCRIPT="${KREW_ROOT:-$HOME/.krew}/bin/kubectl_complete-revisions"; kubectl revisions completion > "$SCRIPT" && chmod +x "$SCRIPT"
```

If you don't use krew, you can install the script next to the binary itself as follows:
```
SCRIPT="$(dirname "$(which kubectl-revisions)")/kubectl_complete-revisions"; kubectl revisions completion > "$SCRIPT" && chmod +x "$SCRIPT"
```

Alternatively, you can also use https://github.com/marckhouzam/kubectl-plugin_completion to generate completion
scripts for this plugin along with other kubectl plugins that support it.


```
kubectl revisions completion
```

### Options

```
  -h, --help   help for completion
```

### Options inherited from parent commands

```
      --as string                      Username to impersonate for the operation. User could be a regular user or a service account in a namespace.
      --as-group stringArray           Group to impersonate for the operation, this flag can be repeated to specify multiple groups.
      --as-uid string                  UID to impersonate for the operation.
      --as-user-extra stringArray      User extras to impersonate for the operation, this flag can be repeated to specify multiple values for the same key.
      --cache-dir string               Default cache directory (default "$HOME/.kube/cache")
      --certificate-authority string   Path to a cert file for the certificate authority
      --client-certificate string      Path to a client certificate file for TLS
      --client-key string              Path to a client key file for TLS
      --cluster string                 The name of the kubeconfig cluster to use
      --context string                 The name of the kubeconfig context to use
      --disable-compression            If true, opt-out of response compression for all requests to the server
      --insecure-skip-tls-verify       If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
      --kubeconfig string              Path to the kubeconfig file to use for CLI requests.
      --log-flush-frequency duration   Maximum number of seconds between log flushes (default 5s)
  -n, --namespace string               If present, the namespace scope for this CLI request
      --password string                Password for basic authentication to the API server
      --request-timeout string         The length of time to wait before giving up on a single server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests. (default "0")
  -s, --server string                  The address and port of the Kubernetes API server
      --tls-server-name string         Server name to use for server certificate validation. If it is not provided, the hostname used to contact the server is used
      --token string                   Bearer token for authentication to the API server
      --user string                    The name of the kubeconfig user to use
      --username string                Username for basic authentication to the API server
  -v, --v Level                        number for the log level verbosity
      --vmodule moduleSpec             comma-separated list of pattern=N settings for file-filtered logging (only works for the default text log format)
```

### SEE ALSO

* [kubectl revisions](kubectl_revisions.md)	 - Time-travel through your workload revision history

