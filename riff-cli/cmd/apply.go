/*
 * Copyright 2018 the original author or authors.
 *
 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

package cmd

import (
	"fmt"
	"github.com/projectriff/riff/riff-cli/cmd/utils"
	"github.com/projectriff/riff/riff-cli/pkg/functions"
	"github.com/projectriff/riff/riff-cli/pkg/kubectl"
	"github.com/projectriff/riff/riff-cli/pkg/osutils"
	"github.com/spf13/cobra"
)

type ApplyOptions struct {
	FilePath  string
	Namespace string
	DryRun    bool
}

func Apply(realKubeCtl kubectl.KubeCtl, dryRunKubeCtl kubectl.KubeCtl) (*cobra.Command, *ApplyOptions) {

	applyOptions := ApplyOptions{}

	var applyCmd = &cobra.Command{
		Use:   "apply",
		Short: "Apply function resource definitions to the cluster",
		Long:  `Apply the resource definition[s] included in the path. A resource will be created if it doesn't exist yet.`,
		Example: `  riff apply -f some/function/path
  riff apply -f some/function/path/some.yaml`,
		Args: utils.AliasFlagToSoleArg("filepath"),


		RunE: func(cmd *cobra.Command, args []string) error {
			kubectlClient := realKubeCtl
			if applyOptions.DryRun {
				kubectlClient = dryRunKubeCtl
			}
			cmd.SilenceUsage = true
			return apply(cmd, applyOptions, kubectlClient)
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			applyOptions.Namespace = utils.GetStringValueWithOverride("namespace", *cmd.Flags())

			return validateApplyOptions(&applyOptions)
		},
	}

	applyCmd.Flags().BoolVar(&applyOptions.DryRun, "dry-run", false, "print generated function artifacts content to stdout only")
	applyCmd.Flags().StringVarP(&applyOptions.FilePath, "filepath", "f", "", "path or directory used for the function resources (defaults to the current directory)")
	applyCmd.Flags().StringVar(&applyOptions.Namespace, "namespace", "", "the namespace used for the deployed resources (defaults to kubectl's default)")

	return applyCmd, &applyOptions
}

func apply(cmd *cobra.Command, opts ApplyOptions, kubectlClient kubectl.KubeCtl) error {
	abs, err := functions.AbsPath(opts.FilePath)
	if err != nil {
		cmd.SilenceUsage = true
		return err
	}

	cmdArgs := []string{"apply"}
	if opts.Namespace != "" {
		cmdArgs = append(cmdArgs, "--namespace", opts.Namespace)
	}

	var message string

	if osutils.IsDirectory(abs) {
		message = fmt.Sprintf("Applying resources in %v\n\n", opts.FilePath)
		resourceDefinitionPaths, err := osutils.FindRiffResourceDefinitionPaths(abs)
		if err != nil {
			return err
		}
		for _, resourceDefinitionPath := range resourceDefinitionPaths {
			cmdArgs = append(cmdArgs, "-f", resourceDefinitionPath)
		}
	} else {
		message = fmt.Sprintf("Applying resource %v\n\n", opts.FilePath)
		cmdArgs = append(cmdArgs, "-f", abs)
	}

	fmt.Print(message)
	output, err := kubectlClient.Exec(cmdArgs)
	if err != nil {
		cmd.SilenceUsage = true
		return err
	}
	fmt.Printf("%v\n", output)
	return nil
}

func validateApplyOptions(options *ApplyOptions) error {
	return validateFilepath(&options.FilePath)
}
