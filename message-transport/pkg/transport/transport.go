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

// Package transport provides interfaces for sending and receiving messages.
package transport

import (
	"github.com/projectriff/riff/message-transport/pkg/message"
)

//go:generate mockery -name=Producer -output mocktransport -outpkg mocktransport

// Producer is an interface for sending messages to arbitrary topics.
// If io.Closer is implemented it will be called when the Producer is no longer needed.
type Producer interface {
	// Send sends a message to a topic.
	Send(topic string, message message.Message) error

	// Errors returns a channel which receives errors resulting asynchronously from sending messages.
	Errors() <-chan error
}

//go:generate mockery -name=Consumer -output mocktransport -outpkg mocktransport

// Consumer is an interface for receiving messages, along with their topics, from a fixed, implementation-defined set
// of topics.
// If io.Closer is implemented it will be called when the Consumer is no longer needed.
type Consumer interface {
	// Receive returns a message along with the topic from which the message was received.
	Receive() (message.Message, string, error)
}

//go:generate mockery -name=Inspector -output mocktransport -outpkg mocktransport

// Inspector is an interface for inspecting the transport.
type Inspector interface {
	// QueueLength returns the queue length of the given topic from the perspective of the given function.
	QueueLength(topic string, function string) (int64, error)
}
