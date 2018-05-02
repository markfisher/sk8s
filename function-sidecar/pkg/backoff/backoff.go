/*
 * Copyright 2018 the original author or authors.
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
package backoff

import (
	"errors"
	"time"
)

type Backoff struct {
	maxRetries, retries, multiplier int
	duration                        time.Duration
}

func NewBackoff(duration time.Duration, maxRetries int, multiplier int) (*Backoff, error) {
	if maxRetries <= 0 {
		return nil, errors.New("'maxRetries' must be > 0")
	}
	if multiplier <= 0 {
		return nil, errors.New("'multiplier' must be > 0")
	}
	if duration <= 0 {
		return nil, errors.New("'duration' must be > 0")
	}

	return &Backoff{
		maxRetries: maxRetries,
		multiplier: multiplier,
		duration:   duration,
	}, nil
}

func (b *Backoff) Backoff() bool {
	// Back off a bit to give the invoker time to come back
	// (if we support windowing or polling this logic will be more complex)
	if b.retries > 0 {
		b.duration = b.duration * time.Duration(b.multiplier)
	}

	b.retries++
	if b.retries <= b.maxRetries {
		time.Sleep(b.duration)
		return true
	}
	return false
}
