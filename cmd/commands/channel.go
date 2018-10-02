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

package commands

import (
	"fmt"
	"github.com/knative/eventing/pkg/apis/channels/v1alpha1"
	"k8s.io/api/core/v1"

	"github.com/projectriff/riff/pkg/core"
	"github.com/spf13/cobra"
)

func Channel() *cobra.Command {
	return &cobra.Command{
		Use:   "channel",
		Short: "Interact with channel related resources",
	}
}

const (
	channelCreateNameIndex = iota
	channelCreateNumberOfArgs
)

const (
	channelListNumberOfArgs = iota
)

const (
	channelDeleteNameIndex = iota
	channelDeleteNumberOfArgs
)

var exactlyOneOfBusOrClusterBus = ExactlyOneOf("bus", "cluster-bus")

func ChannelCreate(fcTool *core.Client) *cobra.Command {
	options := core.CreateChannelOptions{}

	command := &cobra.Command{
		Use:   "create",
		Short: "Create a new channel on a bus or a cluster bus",
		Args: ArgValidationConjunction(
			cobra.ExactArgs(channelCreateNumberOfArgs),
			AtPosition(channelCreateNameIndex, ValidName())),
		Example: `  riff channel create tweets --bus kafka --namespace steve-ns
  riff channel create orders --cluster-bus global-rabbit`,
		PreRunE: FlagsValidatorAsCobraRunE(exactlyOneOfBusOrClusterBus),
		RunE: func(cmd *cobra.Command, args []string) error {
			channelName := args[channelCreateNameIndex]
			options.Name = channelName

			c, err := (*fcTool).CreateChannel(options)
			if err != nil {
				return err
			}

			if options.DryRun {
				marshaller := NewMarshaller(cmd.OutOrStdout())
				if err = marshaller.Marshal(c); err != nil {
					return err
				}
			} else {
				PrintSuccessfulCompletion(cmd)
			}

			return nil
		},
	}

	LabelArgs(command, "CHANNEL_NAME")

	command.Flags().StringVar(&options.Bus, "bus", "", "the `name` of the bus to create the channel in.")
	command.Flags().StringVar(&options.ClusterBus, "cluster-bus", "", "the `name` of the cluster bus to create the channel in.")
	command.Flags().StringVarP(&options.Namespace, "namespace", "n", "", "the `namespace` of the channel and any non-cluster bus")

	command.Flags().BoolVar(&options.DryRun, "dry-run", false, dryRunUsage)
	return command
}

func ChannelList(fcTool *core.Client) *cobra.Command {
	listChannelOptions := core.ListChannelOptions{}

	command := &cobra.Command{
		Use:   "list",
		Short: "List channels",
		Example: `  riff channel list
  riff channel list --namespace joseph-ns`,
		Args: cobra.ExactArgs(channelListNumberOfArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			channels, err := (*fcTool).ListChannels(listChannelOptions)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			Display(out, channelToInterfaceSlice(channels.Items), makeChannelExtractors())
			fmt.Fprintln(out)

			PrintSuccessfulCompletion(cmd)
			return nil
		},
	}

	command.Flags().StringVarP(&listChannelOptions.Namespace, "namespace", "n", "", "the `namespace` of the channels to be listed")

	return command
}

func ChannelDelete(fcTool *core.Client) *cobra.Command {
	options := core.DeleteChannelOptions{}

	command := &cobra.Command{
		Use:   "delete",
		Short: "Delete an existing channel",
		Args: ArgValidationConjunction(
			cobra.ExactArgs(channelDeleteNumberOfArgs),
			AtPosition(channelDeleteNameIndex, ValidName())),
		Example: `  riff channel delete tweets`,
		RunE: func(cmd *cobra.Command, args []string) error {
			options.Name = args[channelDeleteNameIndex]

			err := (*fcTool).DeleteChannel(options)
			if err != nil {
				return err
			}

			PrintSuccessfulCompletion(cmd)
			return nil
		},
	}

	LabelArgs(command, "CHANNEL_NAME")

	command.Flags().StringVarP(&options.Namespace, "namespace", "n", "", "the `namespace` of the channel")
	return command
}

func channelToInterfaceSlice(channels []v1alpha1.Channel) []interface{} {
	result := make([]interface{}, len(channels))
	for i := range channels {
		result[i] = channels[i]
	}
	return result
}

func makeChannelExtractors() []NamedExtractor {
	return []NamedExtractor{
		{
			name: "NAME",
			fn:   func(ch interface{}) string { return ch.(v1alpha1.Channel).Name },
		},
		{
			name: "STATUS",
			fn: func(ch interface{}) string {
				channel := ch.(v1alpha1.Channel)
				condition := channel.Status.GetCondition(v1alpha1.ChannelReady)
				if condition == nil {
					return "Unknown"
				} else {
					switch condition.Status {
					case v1.ConditionTrue:
						return "Running"
					case v1.ConditionFalse:
						return fmt.Sprintf("%s: %s", condition.Reason, condition.Message)
					default:
						return "Unknown"
					}
				}
			},
		},
	}
}
