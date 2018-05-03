/*
 * Copyright 2017 the original author or authors.
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

package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"flag"
	"io"
	"strings"

	"github.com/Shopify/sarama"
	"github.com/bsm/sarama-cluster"
	"github.com/projectriff/riff/function-sidecar/pkg/backoff"
	"github.com/projectriff/riff/function-sidecar/pkg/carrier"
	dispatch "github.com/projectriff/riff/function-sidecar/pkg/dispatcher"
	"github.com/projectriff/riff/function-sidecar/pkg/dispatcher/grpc"
	"github.com/projectriff/riff/function-sidecar/pkg/dispatcher/http"
	"github.com/projectriff/riff/message-transport/pkg/transport"
	"github.com/projectriff/riff/message-transport/pkg/transport/metrics/kafka_over_kafka"
	"github.com/satori/go.uuid"
)

type stringSlice []string

func (sl *stringSlice) String() string {
	return fmt.Sprint(*sl)
}

func (sl *stringSlice) Set(value string) error {
	*sl = stringSlice(strings.Split(value, ","))
	return nil
}

var brokers stringSlice = []string{"localhost:9092"}
var inputs, outputs stringSlice
var group, protocol string
var exitOnComplete = false
var backoffDurationMs, backoffMultiplier, backoffMaxRetries, initialDelayMs, port int

func init() {
	flag.Var(&brokers, "brokers", "location of the Kafka server(s) to connect to")
	flag.Var(&inputs, "inputs", "kafka topic(s) to listen to, as input for the function")
	flag.Var(&outputs, "outputs", "kafka topic(s) to write to with results from the function")
	flag.StringVar(&group, "group", "", "kafka consumer group to act as")
	flag.StringVar(&protocol, "protocol", "", "dispatcher protocol to use. One of [http, grpc]")
	flag.IntVar(&port, "port", -1, "invoker port to call")
	flag.IntVar(&initialDelayMs, "initialDelay", 0, "time to wait (ms) for function container to initialize")
	flag.BoolVar(&exitOnComplete, "exitOnComplete", false, "flag to signal that the sidecar should exit when the output stream is closed")
	flag.IntVar(&backoffMaxRetries, "maxBackoffRetries", 3, "maximum number of times to retry connecting to the invoker")
	flag.IntVar(&backoffMultiplier, "backoffMultiplier", 2, "wait time increase for each retry")
	flag.IntVar(&backoffDurationMs, "backoffDuration", 1000, "base wait time (ms) to wait before retry")

}

func main() {

	flag.Parse()

	if len(inputs) > 1 {
		log.Fatalf("Only 1 input supported for now. See https://github.com/projectriff/riff/issues/184. Provided %v\n", inputs)
	} else if len(outputs) > 1 {
		log.Fatalf("Only 1 output supported for now. See https://github.com/projectriff/riff/issues/184. Provided %v\n", outputs)
	}

	backoffPtr, err := backoff.NewBackoff(time.Duration(backoffDurationMs)*time.Millisecond, backoffMaxRetries, backoffMultiplier)
	if err != nil {
		log.Fatalf("Error initializing backoff: %v\n", err)
	}

	input := inputs[0]
	var output string
	if len(outputs) > 0 {
		output = outputs[0]
	} else {
		output = ""
	}

	log.Printf("Sidecar for function '%v' (%v->%v) using dispatcher %v on port %v starting\n", group, input, output, protocol, port)
	var producer transport.Producer

	if output != "" {
		producer, err = kafka_over_kafka.NewMetricsEmittingProducer(brokers, uuid.NewV4().String())
		if err != nil {
			panic(err)
		}

		if prod, ok := producer.(io.Closer); ok {
			defer prod.Close()
		}
	}

	consumerConfig := makeConsumerConfig()
	consumerConfig.Consumer.Offsets.Initial = sarama.OffsetOldest

	pod := uuid.NewV4().String()

	consumer, err := kafka_over_kafka.NewMetricsEmittingConsumer(brokers, group, pod, []string{input}, consumerConfig)
	if err != nil {
		panic(err)
	}

	if consumer, ok := consumer.(io.Closer); ok {
		defer consumer.Close()
	}

	// trap SIGINT and SIGTERM to trigger a shutdown.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	if initialDelayMs > 0 {
		delay := time.Duration(initialDelayMs) * time.Millisecond
		log.Printf("Waiting for initial delay of %v\n", delay)
		time.Sleep(delay)
	}

LOOP:
	for {

		select {
		case <-signals:
			log.Println("Shutting Down...")
			break LOOP
		default:
		}

		log.Print("Creating dispatcher")
		dispatcher, err := createDispatcher(protocol)
		if err != nil {
			if !backoffOrExit(backoffPtr) {
				panic(err)
			} else {
				log.Printf("Error %v attempting to create dispatcher\n", err)
				continue LOOP
			}
		}
		if d, ok := dispatcher.(io.Closer); ok {
			log.Print("Deferring close()")
			defer d.Close()
		}

		carrier.Run(consumer, producer, dispatcher, output)

		select {
		case <-signals:
			log.Println("Shutting Down...")
		case <-dispatcher.Closed():
			log.Println("End of Stream ...")
		}
		break LOOP
	}

}

func backoffOrExit(backoff *backoff.Backoff) bool {
	if exitOnComplete {
		return false
	}
	// Back off a bit to give the invoker time to come back
	return backoff.Backoff()
}

func createDispatcher(protocol string) (dispatch.Dispatcher, error) {
	switch protocol {
	case "http":
		return dispatch.NewWrapper(http.NewHttpDispatcher(port))
	case "grpc":
		return grpc.NewGrpcDispatcher(port, grpcContextTimeout())
	default:
		panic("Unsupported Dispatcher " + protocol)
	}
}

/*
Since 'exitOnComplete' disables backoff, if set, we block for a long time and quit on a grpc timeout.
With backoff/retry enabled, if we expect the function to start quickly, we block for a short time on each attempt.
If 'initialDelayMs' is set, we expect the function to take longer, and set the grpc timeout to the same value.
For example, if we expect the function to start in about 5 seconds, we delay 5 seconds before attempting to start the
dispatcher. The grpc timeout will also be 5 seconds, so if it takes between 5 and 10 seconds, we're ready to rock without
having to backoff. This is more efficient than backing off in cases where the function takes longer than 5100 ms to start.
*/
func grpcContextTimeout() time.Duration {
	const (
		grpcDefaultContextTimeout        time.Duration = 100 * time.Millisecond
		grpcDefaultExitOnCompleteTimeout time.Duration = 60 * time.Second
	)

	if exitOnComplete {
		return grpcDefaultExitOnCompleteTimeout
	}
	delay := grpcDefaultContextTimeout
	if initialDelayMs > 0 {
		delay = time.Duration(initialDelayMs) * time.Millisecond
	}
	return delay
}

func makeConsumerConfig() *cluster.Config {
	consumerConfig := cluster.NewConfig()
	consumerConfig.Consumer.Return.Errors = true
	consumerConfig.Group.Return.Notifications = true
	return consumerConfig
}
