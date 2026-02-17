// Copyright 2026 Tamás Gulácsi.
//
// SPDX-License-Identifier: Apache-2.0

// Package gtrace helps receiving and sending
//
// Traceparent: Version-TraceID-SpanID-Flags
//
// headers with GRPC.
//
// https://github.com/grpc/grpc-go/blob/master/Documentation/grpc-metadata.md
package gtrace

import (
	"context"

	"github.com/UNO-SOFT/w3ctrace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const key = "traceparent"

// AppendTraceToContext adds the trace to the Client's outgoing Context, to be sent to the Server.
func AppendTraceToContext(ctx context.Context, tr w3ctrace.Trace) context.Context {
	if tr.IsValid() {
		return metadata.AppendToOutgoingContext(ctx, key, tr.String())
	}
	return ctx
}

// FromIncomingContext (ServerStream.Context()) reads the Trace in the Server, send by the Client.
func FromIncomingContext(ctx context.Context) w3ctrace.Trace {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		for _, s := range md.Get(key) {
			if s == "" {
				continue
			}
			if tr, err := w3ctrace.ParseString(s); err == nil {
				return tr
			}
		}
	}
	return w3ctrace.Trace{}
}

// ServerToClientUnary sends the Trace from the unary Server to the Client.
func ServerToClientUnary(ctx context.Context, tr w3ctrace.Trace) {
	if tr.IsValid() {
		grpc.SetHeader(ctx, metadata.Pairs(key, tr.String()))
	}
}

// ServerToClientStreaming sends the Trace from the streaming Server to the Client.
func ServerToClientStreaming(ss grpc.ServerStream, tr w3ctrace.Trace) {
	if tr.IsValid() {
		ss.SetHeader(metadata.Pairs(key, tr.String()))
	}
}
