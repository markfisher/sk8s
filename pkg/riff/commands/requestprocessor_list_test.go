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

package commands_test

import (
	duckv1alpha1 "github.com/knative/pkg/apis/duck/v1alpha1"
	"github.com/projectriff/riff/pkg/cli"
	"github.com/projectriff/riff/pkg/riff/commands"
	"github.com/projectriff/riff/pkg/testing"
	requestv1alpha1 "github.com/projectriff/system/pkg/apis/request/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestRequestProcessorListOptions(t *testing.T) {
	table := testing.OptionsTable{
		{
			Name: "invalid list",
			Options: &commands.RequestProcessorListOptions{
				ListOptions: testing.InvalidListOptions,
			},
			ExpectFieldError: testing.InvalidListOptionsFieldError,
		},
		{
			Name: "valid list",
			Options: &commands.RequestProcessorListOptions{
				ListOptions: testing.ValidListOptions,
			},
			ShouldValidate: true,
		},
	}

	table.Run(t)
}

func TestRequestProcessorListCommand(t *testing.T) {
	requestprocessorsName := "test-requestprocessors"
	requestprocessorOtherName := "test-other-requestprocessors"
	defaultNamespace := "default"
	otherNamespace := "other-namespace"

	table := testing.CommandTable{
		{
			Name: "invalid args",
			Args: []string{},
			Prepare: func(t *testing.T, c *cli.Config) error {
				// disable default namespace
				c.Client.(*testing.FakeClient).Namespace = ""
				return nil
			},
			ShouldError: true,
		},
		{
			Name:         "empty",
			Args:         []string{},
			ExpectOutput: "No request processors found.\n",
		},
		{
			Name: "lists an item",
			Args: []string{},
			GivenObjects: []runtime.Object{
				&requestv1alpha1.RequestProcessor{
					ObjectMeta: metav1.ObjectMeta{
						Name:      requestprocessorsName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectOutput: `
NAME                     TYPE        REF         DOMAIN    READY       AGE
test-requestprocessors   <unknown>   <unknown>   <empty>   <unknown>   <unknown>
`,
		},
		{
			Name: "filters by namespace",
			Args: []string{cli.NamespaceFlagName, otherNamespace},
			GivenObjects: []runtime.Object{
				&requestv1alpha1.RequestProcessor{
					ObjectMeta: metav1.ObjectMeta{
						Name:      requestprocessorsName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectOutput: "No request processors found.\n",
		},
		{
			Name: "all namespace",
			Args: []string{cli.AllNamespacesFlagName},
			GivenObjects: []runtime.Object{
				&requestv1alpha1.RequestProcessor{
					ObjectMeta: metav1.ObjectMeta{
						Name:      requestprocessorsName,
						Namespace: defaultNamespace,
					},
				},
				&requestv1alpha1.RequestProcessor{
					ObjectMeta: metav1.ObjectMeta{
						Name:      requestprocessorOtherName,
						Namespace: otherNamespace,
					},
				},
			},
			ExpectOutput: `
NAMESPACE         NAME                           TYPE        REF         DOMAIN    READY       AGE
default           test-requestprocessors         <unknown>   <unknown>   <empty>   <unknown>   <unknown>
other-namespace   test-other-requestprocessors   <unknown>   <unknown>   <empty>   <unknown>   <unknown>
`,
		},
		{
			Name: "table populates all columns",
			Args: []string{},
			GivenObjects: []runtime.Object{
				&requestv1alpha1.RequestProcessor{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "img",
						Namespace: defaultNamespace,
					},
					Spec: requestv1alpha1.RequestProcessorSpec{
						Template: &corev1.PodSpec{
							Containers: []corev1.Container{
								{Image: "projectriff/upper"},
							},
						},
					},
					Status: requestv1alpha1.RequestProcessorStatus{
						Status: duckv1alpha1.Status{
							Conditions: []duckv1alpha1.Condition{
								{Type: requestv1alpha1.RequestProcessorConditionReady, Status: "True"},
							},
						},
						Domain: "image.default.example.com",
					},
				},
				&requestv1alpha1.RequestProcessor{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "app",
						Namespace: defaultNamespace,
					},
					Spec: requestv1alpha1.RequestProcessorSpec{
						Build: &requestv1alpha1.Build{ApplicationRef: "petclinic"},
					},
					Status: requestv1alpha1.RequestProcessorStatus{
						Status: duckv1alpha1.Status{
							Conditions: []duckv1alpha1.Condition{
								{Type: requestv1alpha1.RequestProcessorConditionReady, Status: "True"},
							},
						},
						Domain: "app.default.example.com",
					},
				},
				&requestv1alpha1.RequestProcessor{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "func",
						Namespace: defaultNamespace,
					},
					Spec: requestv1alpha1.RequestProcessorSpec{
						Build: &requestv1alpha1.Build{FunctionRef: "square"},
					},
					Status: requestv1alpha1.RequestProcessorStatus{
						Status: duckv1alpha1.Status{
							Conditions: []duckv1alpha1.Condition{
								{Type: requestv1alpha1.RequestProcessorConditionReady, Status: "True"},
							},
						},
						Domain: "func.default.example.com",
					},
				},
			},
			ExpectOutput: `
NAME   TYPE          REF                 DOMAIN                      READY   AGE
app    application   petclinic           app.default.example.com     True    <unknown>
func   function      square              func.default.example.com    True    <unknown>
img    image         projectriff/upper   image.default.example.com   True    <unknown>
`,
		},
		{
			Name: "list error",
			Args: []string{},
			WithReactors: []testing.ReactionFunc{
				testing.InduceFailure("list", "requestprocessors"),
			},
			ShouldError: true,
		},
	}

	table.Run(t, commands.NewRequestProcessorListCommand)
}
