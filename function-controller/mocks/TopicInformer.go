// Code generated by mockery v1.0.0
package mocks

import cache "k8s.io/client-go/tools/cache"
import mock "github.com/stretchr/testify/mock"
import projectriffv1 "github.com/projectriff/riff/kubernetes-crds/pkg/client/listers/projectriff/v1alpha1"

// TopicInformer is an autogenerated mock type for the TopicInformer type
type TopicInformer struct {
	mock.Mock
}

// Informer provides a mock function with given fields:
func (_m *TopicInformer) Informer() cache.SharedIndexInformer {
	ret := _m.Called()

	var r0 cache.SharedIndexInformer
	if rf, ok := ret.Get(0).(func() cache.SharedIndexInformer); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(cache.SharedIndexInformer)
		}
	}

	return r0
}

// Lister provides a mock function with given fields:
func (_m *TopicInformer) Lister() projectriffv1.TopicLister {
	ret := _m.Called()

	var r0 projectriffv1.TopicLister
	if rf, ok := ret.Get(0).(func() projectriffv1.TopicLister); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(projectriffv1.TopicLister)
		}
	}

	return r0
}
