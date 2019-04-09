/*
 * Copyright 2018-2019 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package core

import (
	"io"

	build_cs "github.com/knative/build/pkg/client/clientset/versioned"
	eventing "github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	eventing_cs "github.com/knative/eventing/pkg/client/clientset/versioned"
	duckv1alpha1 "github.com/knative/pkg/apis/duck/v1alpha1"
	serving "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	serving_cs "github.com/knative/serving/pkg/client/clientset/versioned"
	"github.com/projectriff/riff/pkg/kubectl"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Client interface {
	CreateFunction(builder Builder, options CreateFunctionOptions, log io.Writer) (*serving.Service, *serving.Revision, *corev1.PersistentVolumeClaim, error)
	UpdateFunction(builder Builder, options UpdateFunctionOptions, log io.Writer) error
	BuildFunction(builder Builder, options BuildFunctionOptions, log io.Writer) error

	CreateSubscription(options CreateSubscriptionOptions) (*eventing.Subscription, error)
	DeleteSubscription(options DeleteSubscriptionOptions) error
	ListSubscriptions(options ListSubscriptionsOptions) (*eventing.SubscriptionList, error)

	ListChannels(options ListChannelOptions) (*eventing.ChannelList, error)
	CreateChannel(options CreateChannelOptions) (*eventing.Channel, error)
	DeleteChannel(options DeleteChannelOptions) error

	ListServices(options ListServiceOptions) (*serving.ServiceList, error)
	CreateService(options CreateOrUpdateServiceOptions) (*serving.Service, error)
	UpdateService(options CreateOrUpdateServiceOptions) (*serving.Service, error)
	DeleteService(options DeleteServiceOptions) error
	ServiceStatus(options ServiceStatusOptions) (*duckv1alpha1.Condition, error)
	ServiceCoordinates(options ServiceInvokeOptions) (ingressIP string, hostName string, err error)

	SystemInstall(manifests map[string]*Manifest, options SystemInstallOptions) (bool, error)
	SystemUninstall(options SystemUninstallOptions) (bool, error)

	NamespaceInit(manifests map[string]*Manifest, options NamespaceInitOptions) error
	NamespaceCleanup(options NamespaceCleanupOptions) error

	// helpers
	FetchPackConfig() (*PackConfig, error)
	DefaultBuildImagePrefix(namespace string) (string, error)
	SetDefaultBuildImagePrefix(namespace, prefix string) error
}

type Builder interface {
	Build(repoName string, options BuildOptions, log io.Writer) error
}

type client struct {
	kubeClient   kubernetes.Interface
	eventing     eventing_cs.Interface
	serving      serving_cs.Interface
	build        build_cs.Interface
	clientConfig clientcmd.ClientConfig
	kubeCtl      kubectl.KubeCtl
}

func NewClient(clientConfig clientcmd.ClientConfig, kubeClient kubernetes.Interface, eventing eventing_cs.Interface, serving serving_cs.Interface, build build_cs.Interface, kubeCtl kubectl.KubeCtl) Client {
	return &client{
		clientConfig: clientConfig,
		kubeClient:   kubeClient,
		eventing:     eventing,
		serving:      serving,
		build:        build,
		kubeCtl:      kubeCtl,
	}
}
