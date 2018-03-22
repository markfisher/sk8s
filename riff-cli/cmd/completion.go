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
	"github.com/spf13/cobra"
	"os"
	"github.com/projectriff/riff/riff-cli/cmd/utils"
)

func Completion(rootCmd *cobra.Command) *cobra.Command {

	var completionCmd = &cobra.Command{
		Use:       "completion [bash|zsh]",
		Short:     "Generate shell completion scripts",
		Long:      "Generate shell completion scripts",
		Example: 	`
To install completion for bash, assuming you have bash-completion installed:

    riff completion bash > /etc/bash_completion.d/riff     

or wherever your bash_completion.d is, for example $(brew --prefix)/etc/bash_completion.d if using homebrew.

    
Completion for zsh is a work in progress`,
		Hidden:    true,
		Args:      utils.And(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		ValidArgs: []string{"bash", "zsh"},
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return rootCmd.GenBashCompletion(os.Stdout)
			case "zsh":
				return rootCmd.GenZshCompletion(os.Stdout)
			}
			panic("Can't be reached")
		},
	}

	return completionCmd

}

