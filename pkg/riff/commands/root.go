/*
 * Copyright 2019 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package commands

import (
	"fmt"

	"github.com/projectriff/riff/pkg/cli"
	"github.com/spf13/cobra"
)

func NewRootCommand(c *cli.Config) *cobra.Command {
	cmd := NewRiffCommand(c)

	cmd.Use = c.Name
	if c.GitDirty == "" {
		cmd.Version = fmt.Sprintf("%s (%s)", c.Version, c.GitSha)
	} else {
		cmd.Version = fmt.Sprintf("%s (%s, with local modifications)", c.Version, c.GitSha)
	}
	cmd.DisableAutoGenTag = true

	cmd.PersistentFlags().StringVar(&c.ViperConfigFile, "config", "", "config file (default is $HOME/.riff.yaml)")
	cmd.PersistentFlags().StringVar(&c.KubeConfigFile, "kubeconfig", "", "kubectl config file (default is $HOME/.kube/config)")

	cmd.AddCommand(NewCompletionCommand(c))
	cmd.AddCommand(NewDocsCommand(c))

	return cmd
}
