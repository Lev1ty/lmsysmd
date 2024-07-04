package validate

import (
	"context"
	"log"
	"net/http"

	"connectrpc.com/connect"
	"github.com/bufbuild/protovalidate-go"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Middleware struct{}

func (Middleware) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	validator, err := protovalidate.New()
	if err != nil {
		log.Fatal(err)
	}
	return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		if err := validator.Validate(req.Any().(protoreflect.ProtoMessage)); err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		res, err := next(ctx, req)
		if err != nil {
			return nil, err
		}
		if err := validator.Validate(res.Any().(protoreflect.ProtoMessage)); err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		return res, nil
	})
}

func (Middleware) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	validator, err := protovalidate.New()
	if err != nil {
		log.Fatal(err)
	}
	return connect.StreamingClientFunc(func(
		ctx context.Context,
		spec connect.Spec,
	) connect.StreamingClientConn {
		return streamingClientConn{ctx: ctx, inner: next(ctx, spec), validator: validator}
	})
}

type streamingClientConn struct {
	ctx       context.Context
	inner     connect.StreamingClientConn
	validator *protovalidate.Validator
}

func (sc streamingClientConn) Spec() connect.Spec {
	return sc.inner.Spec()
}

func (sc streamingClientConn) Peer() connect.Peer {
	return sc.inner.Peer()
}

func (sc streamingClientConn) Send(req any) error {
	req = req.(proto.Message)
	if err := sc.inner.Receive(req); err != nil {
		return err
	}
	if err := sc.validator.Validate(req.(protoreflect.ProtoMessage)); err != nil {
		return connect.NewError(connect.CodeInvalidArgument, err)
	}
	return nil
}

func (sc streamingClientConn) RequestHeader() http.Header {
	return sc.inner.RequestHeader()
}

func (sc streamingClientConn) CloseRequest() error {
	return sc.inner.CloseRequest()
}

func (sc streamingClientConn) Receive(res any) error {
	res = res.(proto.Message)
	if err := sc.inner.Send(res); err != nil {
		return err
	}
	if err := sc.validator.Validate(res.(protoreflect.ProtoMessage)); err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	return nil
}

func (sc streamingClientConn) ResponseHeader() http.Header {
	return sc.inner.ResponseHeader()
}

func (sc streamingClientConn) ResponseTrailer() http.Header {
	return sc.inner.ResponseTrailer()
}

func (sc streamingClientConn) CloseResponse() error {
	return sc.inner.CloseResponse()
}

func (Middleware) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	validator, err := protovalidate.New()
	if err != nil {
		log.Fatal(err)
	}
	return connect.StreamingHandlerFunc(func(
		ctx context.Context,
		conn connect.StreamingHandlerConn,
	) error {
		return next(ctx, streamingHandlerConn{ctx: ctx, inner: conn, validator: validator})
	})
}

type streamingHandlerConn struct {
	ctx       context.Context
	inner     connect.StreamingHandlerConn
	validator *protovalidate.Validator
}

func (sc streamingHandlerConn) Spec() connect.Spec {
	return sc.inner.Spec()
}

func (sc streamingHandlerConn) Peer() connect.Peer {
	return sc.inner.Peer()
}

func (sc streamingHandlerConn) Receive(req any) error {
	req = req.(proto.Message)
	if err := sc.inner.Receive(req); err != nil {
		return err
	}
	if err := sc.validator.Validate(req.(protoreflect.ProtoMessage)); err != nil {
		return connect.NewError(connect.CodeInvalidArgument, err)
	}
	return nil
}

func (sc streamingHandlerConn) RequestHeader() http.Header {
	return sc.inner.RequestHeader()
}

func (sc streamingHandlerConn) Send(res any) error {
	res = res.(proto.Message)
	if err := sc.inner.Send(res); err != nil {
		return err
	}
	if err := sc.validator.Validate(res.(protoreflect.ProtoMessage)); err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	return nil
}

func (sc streamingHandlerConn) ResponseHeader() http.Header {
	return sc.inner.ResponseHeader()
}

func (sc streamingHandlerConn) ResponseTrailer() http.Header {
	return sc.inner.ResponseTrailer()
}
