# go-kubernetes

This library provides a simplified interface on top of the official kubernetes
client libraries.

The design focus of this module is usability, not performance or type-safety.
We wanted to create a module that is closer to the use of kubectl based
scripts.

We use this library for tools that interact with the Kubernetes API,
especially when the inspected objects are of varying types.  
There's also support for building simple admission webhooks using the
[Gin](https://github.com/gin-gonic/gin) framework. This provides an
alternative solutions to writing admission controllers that is more
lightweight than the official SDK.

## Maintenance and PRs

This repository is in active development but is not our main focus.  
PRs are welcome, but will take some time to be reviewed.

## License

All files in the repository are subject to the [Apache 2.0 License](LICENSE)

## Builds and Releases

All commits to the main branch need to use [conventional commits](https://www.conventionalcommits.org/en/v1.0.0/).  
Releases will be generated automatically from these commits using [Release Please](https://github.com/googleapis/release-please).

### Required tools

All [required tools](flake.nix) can be installed locally via [nix](https://nixos.org/)
and are loaded on demand via [direnv](https://direnv.net/).  
On MacOS you can install nix via the installer from [determinate systems](https://determinate.systems/).

```shell
curl --proto '=https' --tlsv1.2 -sSf -L https://install.determinate.systems/nix | sh -s -- install
```

We provided a [justfile](https://github.com/casey/just) to generate the required `.envrc` file.
Run `just init-nix` to get started, or run the [script](hack/init-nix.sh) directly.

### Running unit-tests

After you have set up your environment, run unittests via `just test` or

```shell
go test ./...
```

### Running commit checks

We use pre-commit hooks to ensure code quality.
These hooks are run on PR creation.  
If you encounter issues reported during PR creation, please run the tests
locally until the issues are resolved.

You can use `just lint` to run the pre-commit hooks.  
Please note that this command requires [RequiredTools] to be installed

## Examples

We provide example code in the `cmd` directory.
This code is not part of the official library.

### List all namespaces

```golang
func ListAllNamespaces() {
  // Get the kubeconfig file path
  kubeConfigPath := os.Getenv("KUBECONFIG")
  if kubeConfigPath == "" {
      kubeConfigPath = os.ExpandEnv("$HOME/.kube/config")
  }

  // Create a new client
  client, err := kubernetes.NewClient(kubeConfigPath)
  if err != nil {
    log.Fatal(err)
  }

  // List all objects of type "namespace"
  namespaces, err := client.ListAllObjects(kubernetes.ResourceNamespace, "", "")
  if err != nil {
    log.Fatal(err)
  }

  // Print the names
  for _, namespace := range namespaces {
    fmt.Println(namespace.GetName())
  }
}

```
