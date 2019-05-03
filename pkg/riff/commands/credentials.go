package commands

import (
	"fmt"
	"github.com/projectriff/riff/pkg/core"
	"github.com/projectriff/riff/pkg/core/tasks"
	"github.com/projectriff/riff/pkg/env"
	"github.com/spf13/cobra"
	"k8s.io/api/core/v1"
	"strings"
)

const (
	credentialsSetNumberOfArgs = iota
)

const (
	credentialsListNumberOfArgs = iota
)

const (
	credentialsDeleteSecretNameStartIndex = iota
	credentialsDeleteMinNumberOfArgs
)

func Credentials() *cobra.Command {
	return &cobra.Command{
		Use:   "credentials",
		Short: "Interact with credentials related resources",
	}
}

func CredentialsSet(c *core.Client) *cobra.Command {
	options := core.SetCredentialsOptions{}

	command := &cobra.Command{
		Use:     "set",
		Short:   "create or update secret and bind it to the " + env.Cli.Name + " service account (created if not found)",
		Example: `  ` + env.Cli.Name + ` credentials set --secret mysecret --namespace default --docker-hub johndoe`,
		Args:    cobra.ExactArgs(credentialsSetNumberOfArgs),
		PreRunE: FlagsValidatorAsCobraRunE(
			FlagsValidationConjunction(
				FlagsDependency(Set("namespace"), ValidDnsSubdomain("namespace")),
				NotBlank("secret"),
				AtMostOneOf("gcr", "docker-hub", "registry-user"),
				FlagsDependency(Set("image-prefix"), NotBlank("image-prefix")),
				FlagsDependency(Set("image-prefix"), IsTrue("enable-image-prefix")),
				FlagsDependency(Set("registry-user"), NotBlank("registry")),
				FlagsDependency(Set("registry"),
					NotBlank("registry-user"),
					SupportedRegistryProtocol(func() string {
						return options.Registry
					}))),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			if options.Registry != "" && !strings.Contains(options.Registry, "://") {
				options.Registry = fmt.Sprintf("https://%s", options.Registry)
			}

			if err := (*c).SetCredentials(options); err != nil {
				return err
			}
			PrintSuccessfulCompletion(cmd)
			return nil
		},
	}

	command.Flags().StringVar(&options.NamespaceName, "namespace", "", "the `namespace` of the credentials to be added")
	command.Flags().StringVarP(&options.SecretName, "secret", "s", "push-credentials", "the name of a `secret` containing credentials for the image registry")
	command.Flags().StringVar(&options.GcrTokenPath, "gcr", "", "path to a file containing Google Container Registry credentials")
	command.Flags().StringVar(&options.DockerHubId, "docker-hub", "", "Docker ID for authenticating with Docker Hub; password will be read from stdin")
	command.Flags().StringVar(&options.Registry, "registry", "", "registry server host, scheme must be \"http\" or \"https\" (default \"https\")")
	command.Flags().StringVar(&options.RegistryUser, "registry-user", "", "registry username; password will be read from stdin")
	command.Flags().BoolVar(&options.EnableImagePrefix, "enable-image-prefix", false, "allow image prefix creation/update")
	command.Flags().StringVar(&options.ImagePrefix, "image-prefix", "", "image prefix to use for commands that would otherwise require an --image argument (needs --enable-image-prefix). If not set but --enable-image-prefix is, this value will be derived for Docker Hub and GCR")

	return command
}

func CredentialsList(c *core.Client) *cobra.Command {
	options := core.ListCredentialsOptions{}

	command := &cobra.Command{
		Use:   "list",
		Short: "List credentials resources",
		Example: `  ` + env.Cli.Name + ` credentials list
  ` + env.Cli.Name + ` credentials list --namespace joseph-ns`,
		Args: cobra.ExactArgs(credentialsListNumberOfArgs),
		PreRunE: FlagsValidatorAsCobraRunE(
			FlagsDependency(Set("namespace"), ValidDnsSubdomain("namespace")),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			credentialsList, err := (*c).ListCredentials(options)
			if err != nil {
				return err
			}
			Display(cmd.OutOrStdout(), secretToInterfaceSlice(credentialsList.Items), makeSecretExtractors())
			return nil
		},
	}

	command.Flags().StringVarP(&options.NamespaceName, "namespace", "n", "", "the `namespace` of the credentials to be listed")
	return command
}

type DeleteCredentialsCliOptions struct {
	NamespaceName string
}

func CredentialsDelete(c *core.Client) *cobra.Command {
	cliOptions := DeleteCredentialsCliOptions{}

	command := &cobra.Command{
		Use:   "delete",
		Short: "Delete specified credentials",
		Example: `  ` + env.Cli.Name + ` credentials delete secret1 secret2
  ` + env.Cli.Name + ` credentials delete --namespace joseph-ns secret`,
		Args: ArgValidationConjunction(
			cobra.MinimumNArgs(credentialsDeleteMinNumberOfArgs),
			StartingAtPosition(credentialsDeleteSecretNameStartIndex, ArgNotBlank("secret")),
		),
		PreRunE: FlagsValidatorAsCobraRunE(
			FlagsDependency(Set("namespace"), ValidDnsSubdomain("namespace")),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			names := args[credentialsDeleteSecretNameStartIndex:]
			results := tasks.ApplyInParallel(names, func(name string) error {
				options := core.DeleteCredentialsOptions{NamespaceName: cliOptions.NamespaceName, Name: name}
				return (*c).DeleteCredentials(options)
			})
			err := tasks.MergeResults(results, func(result tasks.CorrelatedResult) string {
				err := result.Error
				if err == nil {
					return ""
				}
				return fmt.Sprintf("Unable to delete credentials %s: %v", result.Input, err)
			})
			if err != nil {
				return err
			}

			PrintSuccessfulCompletion(cmd)
			return nil
		},
	}
	command.Flags().StringVarP(&cliOptions.NamespaceName, "namespace", "n", "", "the `namespace` of the credentials to be deleted")
	return command
}

func secretToInterfaceSlice(items []v1.Secret) []interface{} {
	result := make([]interface{}, len(items))
	for i := range items {
		result[i] = items[i]
	}
	return result
}

func makeSecretExtractors() []NamedExtractor {
	return []NamedExtractor{
		{
			name: "NAME",
			fn:   func(s interface{}) string { return s.(v1.Secret).Name },
		},
	}
}