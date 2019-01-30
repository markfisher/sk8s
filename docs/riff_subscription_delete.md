## riff subscription delete

Delete existing subscriptions

### Synopsis

Delete existing subscriptions

```
riff subscription delete [flags]
```

### Examples

```
  riff subscription delete my-subscription --namespace joseph-ns
  riff subscription delete my-subscription-1 my-subscription-2
```

### Options

```
  -h, --help               help for delete
  -n, --namespace string   the namespace of the subscription
```

### Options inherited from parent commands

```
      --kubeconfig path   the path of a kubeconfig (default "~/.kube/config")
      --master address    the address of the Kubernetes API server; overrides any value in kubeconfig
```

### SEE ALSO

* [riff subscription](riff_subscription.md)	 - Interact with subscription-related resources

