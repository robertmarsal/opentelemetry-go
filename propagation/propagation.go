// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package propagation // import "go.opentelemetry.io/otel/propagation"

import (
	"context"
	"net/http"
	"strings"
)

// TextMapCarrier is the storage medium used by a TextMapPropagator.
type TextMapCarrier interface {
	// DO NOT CHANGE: any modification will not be backwards compatible and
	// must never be done outside of a new major release.

	// Get returns the value associated with the passed key.
	Get(key string) string
	// DO NOT CHANGE: any modification will not be backwards compatible and
	// must never be done outside of a new major release.

	// Set stores the key-value pair.
	Set(key string, value string)
	// DO NOT CHANGE: any modification will not be backwards compatible and
	// must never be done outside of a new major release.

	// Keys lists the keys stored in this carrier.
	Keys() []string
	// DO NOT CHANGE: any modification will not be backwards compatible and
	// must never be done outside of a new major release.
}

// MapCarrier is a TextMapCarrier that uses a map held in memory as a storage
// medium for propagated key-value pairs.
type MapCarrier map[string]string

// Compile time check that MapCarrier implements the TextMapCarrier.
var _ TextMapCarrier = MapCarrier{}

// Get returns the value associated with the passed key.
func (c MapCarrier) Get(key string) string {
	return c[key]
}

// Set stores the key-value pair.
func (c MapCarrier) Set(key, value string) {
	c[key] = value
}

// Keys lists the keys stored in this carrier.
func (c MapCarrier) Keys() []string {
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	return keys
}

// HeaderCarrier adapts http.Header to satisfy the TextMapCarrier interface.
type HeaderCarrier http.Header

// Get returns the first value associated with the passed key. The key is
// case-sensitive and will only return the value that matches exactly. If there
// are no values associated with the key, Get returns "".
func (hc HeaderCarrier) Get(key string) string {
	// Use the underlying map to avoid normalising the key and thereby
	// changing its value.
	if hc == nil {
		return ""
	}
	value, ok := http.Header(hc)[key]
	if !ok || len(value) == 0 {
		return ""
	}

	return value[0]
}

// Set sets the header entries associated with key to the single element value.
// It replaces any existing values associated with key.
//
// The key is case-sensitive, it is used exactly as passed.
func (hc HeaderCarrier) Set(key string, value string) {
	// Use the underlying map to set the value to avoid normalising the key and
	// thereby changing its value.
	http.Header(hc)[strings.ToLower(key)] = []string{value}
}

// Keys lists the keys stored in this carrier.
func (hc HeaderCarrier) Keys() []string {
	keys := make([]string, 0, len(hc))
	for k := range hc {
		keys = append(keys, k)
	}
	return keys
}

// TextMapPropagator propagates cross-cutting concerns as key-value text
// pairs within a carrier that travels in-band across process boundaries.
type TextMapPropagator interface {
	// DO NOT CHANGE: any modification will not be backwards compatible and
	// must never be done outside of a new major release.

	// Inject set cross-cutting concerns from the Context into the carrier.
	Inject(ctx context.Context, carrier TextMapCarrier)
	// DO NOT CHANGE: any modification will not be backwards compatible and
	// must never be done outside of a new major release.

	// Extract reads cross-cutting concerns from the carrier into a Context.
	Extract(ctx context.Context, carrier TextMapCarrier) context.Context
	// DO NOT CHANGE: any modification will not be backwards compatible and
	// must never be done outside of a new major release.

	// Fields returns the keys whose values are set with Inject.
	Fields() []string
	// DO NOT CHANGE: any modification will not be backwards compatible and
	// must never be done outside of a new major release.
}

type compositeTextMapPropagator []TextMapPropagator

func (p compositeTextMapPropagator) Inject(ctx context.Context, carrier TextMapCarrier) {
	for _, i := range p {
		i.Inject(ctx, carrier)
	}
}

func (p compositeTextMapPropagator) Extract(ctx context.Context, carrier TextMapCarrier) context.Context {
	for _, i := range p {
		ctx = i.Extract(ctx, carrier)
	}
	return ctx
}

func (p compositeTextMapPropagator) Fields() []string {
	unique := make(map[string]struct{})
	for _, i := range p {
		for _, k := range i.Fields() {
			unique[k] = struct{}{}
		}
	}

	fields := make([]string, 0, len(unique))
	for k := range unique {
		fields = append(fields, k)
	}
	return fields
}

// NewCompositeTextMapPropagator returns a unified TextMapPropagator from the
// group of passed TextMapPropagator. This allows different cross-cutting
// concerns to be propagates in a unified manner.
//
// The returned TextMapPropagator will inject and extract cross-cutting
// concerns in the order the TextMapPropagators were provided. Additionally,
// the Fields method will return a de-duplicated slice of the keys that are
// set with the Inject method.
func NewCompositeTextMapPropagator(p ...TextMapPropagator) TextMapPropagator {
	return compositeTextMapPropagator(p)
}
