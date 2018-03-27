/*
 * Copyright 2018-Present the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package autoscaler

import (
	"github.com/projectriff/riff/message-transport/pkg/transport/metrics"
	"fmt"
	"sync"
	"time"
	"math"
	"log"
	"io"
	"github.com/projectriff/riff/message-transport/pkg/transport"
)

//go:generate mockery -name=AutoScaler -output mockautoscaler -outpkg mockautoscaler

type AutoScaler interface {
	// Set maximum replica count policy.
	SetMaxReplicasPolicy(func(function FunctionId) int)

	// Set delay scale down policy.
	SetDelayScaleDownPolicy(func(function FunctionId) time.Duration)

	// Run starts the autoscaler receiving and sampling metrics.
	Run()

	// Close stops the autoscaler receiving and sampling metrics.
	io.Closer

	// InformFunctionReplicas is used to tell the autoscaler the actual number of replicas there are for a given
	// function. The function is not necessarily being monitored by the autoscaler.
	InformFunctionReplicas(function FunctionId, replicas int)

	// StartMonitoring starts monitoring metrics for the given topic and function.
	StartMonitoring(topic string, function FunctionId) error

	// StopMonitoring stops monitoring metrics for the given topic and function.
	StopMonitoring(topic string, function FunctionId) error

	// Propose proposes the number of replicas for functions that are being monitored.
	Propose() map[FunctionId]int
}

// FunctionId identifies a function
// TODO: support namespaces.
type FunctionId struct {
	Function string
}

// Go does not provide a MaxInt, so we have to calculate it.
const MaxUint = ^uint(0)
const MaxInt = int(MaxUint >> 1)

// NewAutoScaler constructs an autoscaler instance using the given metrics receiver and the given transport inspector.
func NewAutoScaler(metricsReceiver metrics.MetricsReceiver, transportInspector transport.Inspector) *autoScaler {
	return &autoScaler{
		mutex:               &sync.Mutex{},
		metricsReceiver:     metricsReceiver,
		transportInspector:  transportInspector,
		totals:              make(map[string]map[FunctionId]*metricsTotals),
		scalers:             make(map[FunctionId]scaler),
		replicas:            make(map[FunctionId]int),
		maxReplicas:         func(function FunctionId) int { return MaxInt },
		delayScaleDown:      func(function FunctionId) time.Duration { return time.Duration(0) },
		stop:                make(chan struct{}),
		accumulatingStopped: make(chan struct{}),
	}
}

func (a *autoScaler) SetMaxReplicasPolicy(maxReplicas func(function FunctionId) int) {
	a.maxReplicas = maxReplicas
}

func (a *autoScaler) SetDelayScaleDownPolicy(delayScaleDown func(function FunctionId) time.Duration) {
	a.delayScaleDown = delayScaleDown
}

func (a *autoScaler) Run() {
	a.mutex.Lock() // fail Run if a.mutex is nil
	defer a.mutex.Unlock()

	go a.receiveLoop()
}

type autoScaler struct {
	mutex                      *sync.Mutex // nil when autoScaler is closed
	metricsReceiver            metrics.MetricsReceiver
	transportInspector         transport.Inspector
	totals                     map[string]map[FunctionId]*metricsTotals
	scalers                    map[FunctionId]scaler
	replicas                   map[FunctionId]int // tracks all functions, including those which are not being monitored
	maxReplicas                func(function FunctionId) int
	delayScaleDown             func(function FunctionId) time.Duration
	stop                       chan struct{}
	accumulatingStopped        chan struct{}
}

// metrics counts the number of messages transmitted to a Subscription's topic and received by the Subscription.
type metricsTotals struct {
	transmitCount int32
	receiveCount  int32
}

func (a *autoScaler) Propose() map[FunctionId]int {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	proposals := make(map[FunctionId]int)
	for _, funcTotals := range a.totals {
		for fn, mt := range funcTotals {
			if _, ok := proposals[fn]; ok {
				panic("Functions with multiple input topics are not supported")
			}
			proposals[fn] = a.scalers[fn](mt)
			// proposals[fn] = max(proposals[fn], a.scalers[fn](mt)) might help multiple input topics

			// Zero the sampled metrics for the next interval
			funcTotals[fn] = &metricsTotals{}
		}
	}

	return proposals
}

func (a *autoScaler) emptyQueue(funcId FunctionId) (bool, int64) {
	for topic, funcTotals := range a.totals {
		if _, ok := funcTotals[funcId]; ok {
			queueLen, err := a.transportInspector.QueueLength(topic, funcId.Function)
			if err != nil {
				log.Printf("Failed to obtain queue length (and will assume it is positive): %v", err)
				return false, -1
			}
			if queueLen > 0 {
				return false, queueLen
			}
		}
	}
	return true, 0
}

func (a *autoScaler) StartMonitoring(topic string, fn FunctionId) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	funcTotals, ok := a.totals[topic]
	if !ok {
		funcTotals = make(map[FunctionId]*metricsTotals)
		a.totals[topic] = funcTotals
	}

	_, ok = funcTotals[fn]
	if ok {
		return fmt.Errorf("Already monitoring topic %s and function %s", topic, fn)
	}

	funcTotals[fn] = &metricsTotals{}

	a.scalers[fn] = decorate(a.metricsScaler(fn), a.limitScalingUp(fn), a.limitScalingDown(fn), a.delay(fn))

	return nil
}

func (a *autoScaler) delay(fn FunctionId) adjuster {
	p := NewDelayer(func() time.Duration {
		return a.delayScaleDown(fn);
	})

	return func(proposal int) int {
		return p.Delay(proposal).Get()
	}
}

func (a *autoScaler) StopMonitoring(topic string, function FunctionId) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	funcTotals, ok := a.totals[topic]
	if !ok {
		return fmt.Errorf("Not monitoring topic %s and function %s", topic, function)
	}

	_, ok = funcTotals[function]
	if !ok {
		return fmt.Errorf("Not monitoring topic %s and function %s", topic, function)
	}

	delete(funcTotals, function)

	// Avoid leaking memory.
	if len(funcTotals) == 0 {
		delete(a.totals, topic)
	}
	delete(a.scalers, function)

	return nil
}

func (a *autoScaler) receiveLoop() {
	producerMetrics := a.metricsReceiver.ProducerMetrics()
	consumerMetrics := a.metricsReceiver.ConsumerMetrics()
	for {
		select {
		case pm, ok := <-producerMetrics:
			if ok { // ok should always be true
				a.receiveProducerMetric(pm)
			}

		case cm, ok := <-consumerMetrics:
			if ok { // ok should always be true
				a.receiveConsumerMetric(cm)
			}

		case <-a.stop:
			if receiver, ok := a.metricsReceiver.(io.Closer); ok {
				err := receiver.Close()
				if err != nil {
					log.Printf("Error closing metrics receiver: %v", err)
				}
			}
			close(a.accumulatingStopped)
			return
		}
	}
}

func (a *autoScaler) receiveConsumerMetric(cm metrics.ConsumerAggregateMetric) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	funcTotals, ok := a.totals[cm.Topic]
	if ok {
		mt, ok := funcTotals[FunctionId{cm.ConsumerGroup}]
		if ok {
			mt.receiveCount += cm.Count
		}
	}
}

func (a *autoScaler) receiveProducerMetric(pm metrics.ProducerAggregateMetric) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	funcTotals, ok := a.totals[pm.Topic]
	if ok {
		for _, mt := range funcTotals {
			mt.transmitCount += pm.Count
		}
	}
}

type scaler func(*metricsTotals) int
type adjuster func(int) int

func decorate(s scaler, adjusters ...adjuster) scaler {
	for _, a := range adjusters {
		s = compose(a, s)
	}

	return s
}

func compose(a adjuster, s scaler) scaler {
	return func(mt *metricsTotals) int {
		return a(s(mt))
	}
}

func (a *autoScaler) metricsScaler(fn FunctionId) scaler {
	return func(mt *metricsTotals) int {
		var proposedReplicas int
		if mt.receiveCount == 0 {
			if mt.transmitCount == 0 {
				proposedReplicas = 0
			} else {
				proposedReplicas = 1 // arbitrary value
			}
		} else {
			proposedReplicas = int(math.Ceil(float64(a.replicas[fn]) * float64(mt.transmitCount) / float64(mt.receiveCount)))
		}
		return proposedReplicas
	}
}

func (a *autoScaler) limitScalingUp(fn FunctionId) adjuster {
	return func(proposedReplicas int) int {
		maxReplicas := a.maxReplicas(fn)
		possibleChange := proposedReplicas != a.replicas[fn]
		if proposedReplicas > maxReplicas {
			if possibleChange {
				log.Printf("Proposing %v should have maxReplicas (%d) instead of %d replicas", fn, maxReplicas, proposedReplicas)
			}
			proposedReplicas = maxReplicas
		}
		return proposedReplicas
	}
}

func (a *autoScaler) limitScalingDown(fn FunctionId) adjuster {
	return func(proposedReplicas int) int {
		// If zero replicas are proposed *and* there is already at least one replica, check the queue of work to the function.
		// The queue length is not allowed to initiate scaling up from 0 to 1 as that would confuse rate-based autoscaling.
		if proposedReplicas == 0 && a.replicas[fn] != 0 {
			empty, length := a.emptyQueue(fn)
			if !empty {
				// There may be work to do, so propose 1 replica instead.
				log.Printf("Ignoring proposal to scale %v to 0 replicas since queue length is %d", fn.Function, length)
				proposedReplicas = 1
			}
		}
		return proposedReplicas
	}
}

func (a *autoScaler) InformFunctionReplicas(function FunctionId, replicas int) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.replicas[function] = replicas
}

func (a *autoScaler) Close() error {
	a.mutex.Lock()
	close(a.stop)
	<-a.accumulatingStopped
	a.mutex = nil // ensure autoScaler can no longer be used
	return nil
}
