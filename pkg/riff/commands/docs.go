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
	"context"

	"github.com/knative/pkg/apis"
	"github.com/projectriff/riff/pkg/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

type DocsOptions struct {
	Directory string
}

func (opts *DocsOptions) Validate(ctx context.Context) *apis.FieldError {
	// TODO implement
	return nil
}

func NewDocsCommand(c *cli.Config) *cobra.Command {
	opts := &DocsOptions{}

	cmd := &cobra.Command{
		Use:     "docs",
		Short:   "<todo>",
		Example: "<todo>",
		Hidden:  true,
		Args:    cli.Args(),
		PreRunE: cli.ValidateOptions(opts),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := c.FileSystem.MkdirAll(opts.Directory, 0744); err != nil {
				return err
			}
			return doc.GenMarkdownTree(cmd.Root(), opts.Directory)
		},
	}

	cmd.Flags().StringVarP(&opts.Directory, "dir", "d", "docs", "the output directory for the docs.")

	return cmd
}
