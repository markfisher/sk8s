## riff logs

Display the logs for a running function

### Synopsis

Display the logs for a running function

```
riff logs [flags]
```

### Examples

```

    riff logs -n myfunc -t

will tail the logs from the 'sidecar' container for the function 'myfunc'


```

### Options

```
  -c, --container string   the name of the function container (sidecar or main) (default "sidecar")
  -h, --help               help for logs
  -n, --name string        the name of the function
      --namespace string   the namespace used for the deployed resources
  -t, --tail               tail the logs
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.riff.yaml)
```

### SEE ALSO

* [riff](riff.md)	 - Commands for creating and managing function resources

