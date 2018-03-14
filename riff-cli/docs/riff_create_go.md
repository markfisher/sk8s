## riff create go

Create a Go function

### Synopsis


Generate the function based on a shared '.so' library file specified as the filename
and exported symbol name specified as the handler.

For example, type:

    riff init go -i words -l go -n rot13 --handler=Encode

to generate the required Dockerfile and resource definitions using sensible defaults.

```
riff create go [flags]
```

### Options

```
  -a, --artifact string       path to the function artifact, source code or jar file
      --dry-run               print generated function artifacts content to stdout only
  -f, --filepath string       path or directory used for the function resources (defaults to the current directory)
      --force                 overwrite existing functions artifacts
      --handler string        the name of the function handler
  -h, --help                  help for go
  -i, --input string          the name of the input topic (defaults to function name)
  -n, --name string           the name of the function (defaults to the name of the current directory)
      --namespace string      the namespace used for the deployed resources (defaults to kubectl's default)
  -o, --output string         the name of the output topic (optional)
      --push                  push the image to Docker registry
      --riff-version string   the version of riff to use when building containers (default "latest")
  -u, --useraccount string    the Docker user account to be used for the image repository (default "current OS user")
  -v, --version string        the version of the function image (default "0.0.1")
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.riff.yaml)
```

### SEE ALSO
* [riff create](riff_create.md)	 - Create a function

