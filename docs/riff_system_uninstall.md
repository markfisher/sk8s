## riff system uninstall

Remove riff and Knative system components

### Synopsis

Remove riff and Knative system components

Use the '--istio' flag to also remove Istio components.'


```
riff system uninstall [flags]
```

### Examples

```
  riff system uninstall
```

### Options

```
      --force        force the removal of components without getting any prompts
  -h, --help         help for uninstall
      --istio        include Istio components and the istio-system namespace in the removal
      --monitoring   include Prometheus and Grafana components and the monitoring namespace in the removal
      --tracing      remove Zipkin tracing components
```

### Options inherited from parent commands

```
      --kubeconfig path   the path of a kubeconfig (default "~/.kube/config")
      --master address    the address of the Kubernetes API server; overrides any value in kubeconfig
```

### SEE ALSO

* [riff system](riff_system.md)	 - Manage system related resources

