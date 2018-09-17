/*
 * Copyright 2018 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package commands_test

import (
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff/cmd/commands"
	. "github.com/spf13/cobra"
)

var _ = Describe("The cobra extensions", func() {

	Context("the broadcasting string value", func() {
		It("should panic when constructed with 0 pointers", func() {

			Expect(func() { commands.BroadcastStringValue("default") }).To(Panic())

		})

		It("should set the default value to all pointers", func() {
			var value1, value2 string

			commands.BroadcastStringValue("the-default", &value1, &value2)

			Expect(value1).To(Equal("the-default"))
			Expect(value2).To(Equal("the-default"))

		})

		It("should set the value to all pointers", func() {
			var value1, value2 string

			v := commands.BroadcastStringValue("the-default", &value1, &value2)

			v.Set("bar")

			Expect(value1).To(Equal("bar"))
			Expect(value2).To(Equal("bar"))

		})

		It("should not fail if an optional argument to validate is not provided", func() {
			command := &Command{
				Use: "some-command",
				Args: commands.ArgValidationConjunction(
					MaximumNArgs(1),
					commands.OptionalAtPosition(1, func(_ *Command, _ string) error {
						return errors.New("should not be called")
					}),
				),
				RunE: func(cmd *Command, args []string) error {
					return nil
				},
			}

			Expect(command.Execute()).NotTo(HaveOccurred())
		})

	})
})
