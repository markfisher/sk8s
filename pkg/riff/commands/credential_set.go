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

	"github.com/projectriff/riff/pkg/riff"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CredentialSetOptions struct {
	Namespace string
	Name      string
}

func NewCredentialSetCommand(c *riff.Config) *cobra.Command {
	opt := &CredentialSetOptions{}

	cmd := &cobra.Command{
		Use: "set",
		RunE: func(cmd *cobra.Command, args []string) error {
			secret, err := c.Core().Secrets(opt.Namespace).Get(opt.Name, metav1.GetOptions{})
			if err != nil {
				if !apierrs.IsNotFound(err) {
					return err
				}

				_, err = c.Core().Secrets(opt.Namespace).Create(&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      opt.Name,
						Namespace: opt.Namespace,
						Labels: map[string]string{
							// TODO get label from riff system
							"projectriff.io/credential": "",
						},
					},
					// TODO define secret data
					StringData: map[string]string{},
				})
				return err
			}

			// ensure we are not mutating a non-riff secret
			if _, ok := secret.Labels["projectriff.io/credential"]; !ok {
				return fmt.Errorf("credential %q exists, but is not owned by riff", opt.Name)
			}

			// TODO mutate secret
			_, err = c.Core().Secrets(opt.Namespace).Update(secret)

			return err
		},
	}

	riff.Args(cmd, riff.NameArg(&opt.Name))

	riff.NamespaceFlag(cmd, c, &opt.Namespace)

	return cmd
}
