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

package utils

import (
	"github.com/spf13/cobra"
)

// CommandChain returns a composite command that runs the provided commands one after the other.
// For each run kind, the variants that can error (ie runE vs run) are preferred if defined.
// postRun variations are run in reversed order
func CommandChain(commands ... *cobra.Command) *cobra.Command {

	runE := func(cmd *cobra.Command, args []string) error {
		for _, command := range commands {
			if command.RunE != nil {
				err := command.RunE(cmd, args)
				if err != nil {
					return err
				}
			} else {
				command.Run(cmd, args)
			}
		}
		return nil
	}
	run := func(cmd *cobra.Command, args []string) {
		runE(cmd, args)
	}

	preRunE := func(cmd *cobra.Command, args []string) error {
		for _, command := range commands {
			if command.PreRunE != nil {
				err := command.PreRunE(cmd, args)
				if err != nil {
					return err
				}
			} else if command.PreRun != nil {
				command.PreRun(cmd, args)
			}
		}
		return nil
	}
	preRun := func(cmd *cobra.Command, args []string) {
		preRunE(cmd, args)
	}

	persistentPreRunE := func(cmd *cobra.Command, args []string) error {
	outer:
		for _, command := range commands {
			for p := command; p != nil; p = p.Parent() {
				if p.PersistentPreRunE != nil {
					if err := p.PersistentPreRunE(cmd, args); err != nil {
						return err
					}
					break outer
				} else if p.PersistentPreRun != nil {
					p.PersistentPreRun(cmd, args)
					break outer
				}
			}
		}
		return nil
	}
	persistentPreRun := func(cmd *cobra.Command, args []string) {
		persistentPreRunE(cmd, args)
	}


	postRunE := func(cmd *cobra.Command, args []string) error {
		for i := len(commands) ; i >= 0 ; i-- {
			command := commands[i]
			if command.PostRunE != nil {
				err := command.PostRunE(cmd, args)
				if err != nil {
					return err
				}
			} else if command.PostRun != nil {
				command.PostRun(cmd, args)
			}
		}
		return nil
	}
	postRun := func(cmd *cobra.Command, args []string) {
		postRunE(cmd, args)
	}

	persistentPostRunE := func(cmd *cobra.Command, args []string) error {
	outer:
		for i := len(commands) ; i >= 0 ; i-- {
			command := commands[i]
			for p := command; p != nil; p = p.Parent() {
				if p.PersistentPreRunE != nil {
					if err := p.PersistentPreRunE(cmd, args); err != nil {
						return err
					}
					break outer
				} else if p.PersistentPreRun != nil {
					p.PersistentPreRun(cmd, args)
					break outer
				}
			}
		}
		return nil
	}
	persistentPostRun := func(cmd *cobra.Command, args []string) {
		persistentPostRunE(cmd, args)
	}

	var chain = &cobra.Command{
		Run:                run,
		RunE:               runE,
		PreRun:             preRun,
		PreRunE:            preRunE,
		PostRun:            postRun,
		PostRunE:           postRunE,
		PersistentPreRun:   persistentPreRun,
		PersistentPreRunE:  persistentPreRunE,
		PersistentPostRun:  persistentPostRun,
		PersistentPostRunE: persistentPostRunE,
	}

	// Merge flags from all delegate commands
	for _, c := range commands {
		chain.Flags().AddFlagSet(c.Flags())
	}
	return chain
}

