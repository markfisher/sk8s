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
	"fmt"

	"github.com/projectriff/riff/pkg/cli"
	"github.com/spf13/cobra"
)

type ProcessorUpdateOptions struct {
	Namespace string
}

func (opts *ProcessorUpdateOptions) Validate(ctx context.Context) *cli.FieldError {
	// TODO implement
	return nil
}

func (opts *ProcessorUpdateOptions) Exec(ctx context.Context, c *cli.Config) error {
	return fmt.Errorf("not implemented")
}

func NewProcessorUpdateCommand(c *cli.Config) *cobra.Command {
	opts := &ProcessorUpdateOptions{}

	cmd := &cobra.Command{
		Use:     "update",
		Short:   "<todo>",
		Example: "<todo>",
		Args:    cli.Args(),
		PreRunE: cli.ValidateOptions(opts),
		RunE:    cli.ExecOptions(c, opts),
	}

	cli.NamespaceFlag(cmd, c, &opts.Namespace)

	return cmd
}
