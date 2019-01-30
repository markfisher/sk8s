## riff channel delete

Delete existing channels

### Synopsis

Delete existing channels

```
riff channel delete [flags]
```

### Examples

```
  riff channel delete tweets
  riff channel delete channel-1 channel-2
```

### Options

```
  -h, --help                  help for delete
  -n, --namespace namespace   the namespace of the channel
```

### Options inherited from parent commands

```
      --kubeconfig path   the path of a kubeconfig (default "~/.kube/config")
      --master address    the address of the Kubernetes API server; overrides any value in kubeconfig
```

### SEE ALSO

* [riff channel](riff_channel.md)	 - Interact with channel related resources

