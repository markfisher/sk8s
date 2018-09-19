## riff function build

Trigger a revision build for a function resource

### Synopsis

Trigger a revision build for a function resource

```
riff function build [flags]
```

### Examples

```
  riff function build square
```

### Options

```
  -h, --help                  help for build
  -l, --local-path string     path to local source to build the image from
  -n, --namespace namespace   the namespace of the function
  -v, --verbose               print details of command progress
  -w, --wait                  wait until the created resource reaches either a successful or an error state (automatic with --verbose)
```

### Options inherited from parent commands

```
      --kubeconfig path   the path of a kubeconfig (default "~/.kube/config")
      --master address    the address of the Kubernetes API server; overrides any value in kubeconfig
```

### SEE ALSO

* [riff function](riff_function.md)	 - Interact with function related resources

