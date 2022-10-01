package grpclogging

import (
	"context"
	"path"
	"time"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"
	"github.com/anzx/pkg/log"
	"github.com/anzx/pkg/log/fabriclog"

	"github.com/anzx/fabric-cards/pkg/middleware/grpclogging/ctxlg"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

var (
	// SystemField is used in every log statement made through grpc_zap. Can be overwritten before any initialization code.
	SystemField = "grpc"
	// ServerField is used in every server-side log statement made through grpc_zap.Can be overwritten before initialization.
	ServerField = "server"
)

// UnaryServerInterceptor returns a new unary server interceptors that adds zap.Logger to the context.
func UnaryServerInterceptor(opts ...Option) grpc.UnaryServerInterceptor {
	o := evaluateServerOpt(opts)
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		startTime := time.Now()
		newCtx := injectLogger(ctx, info.FullMethod, startTime)

		resp, err := handler(newCtx, req)
		if !o.shouldLog(info.FullMethod, err) {
			if err != nil {
				logf.Err(ctx, err)
			}
			return resp, err
		}

		code := o.codeFunc(err)
		logCall(newCtx, o, "finished unary call with code "+code.String(), code, startTime, err)

		if err != nil {
			logf.Err(ctx, err)
		}
		return resp, err
	}
}

func injectLogger(ctx context.Context, fullMethodString string, start time.Time) context.Context {
	f := ctxlg.TagsToFields(ctx)
	f["grpc.start_time"] = start.Format(time.RFC3339)
	if d, ok := ctx.Deadline(); ok {
		f["grpc.request.deadline"] = d.Format(time.RFC3339)
	}
	for k, v := range serverCallFields(fullMethodString) {
		f[k] = v
	}

	injectLog := ctxlg.CtxLogger{Fields: f}

	return ctxlg.ToContext(ctx, &injectLog)
}

func serverCallFields(fullMethodString string) map[string]interface{} {
	service := path.Dir(fullMethodString)[1:]
	method := path.Base(fullMethodString)
	return map[string]interface{}{
		"system":       SystemField,
		"span.kind":    ServerField,
		"grpc.service": service,
		"grpc.method":  method,
	}
}

func logCall(ctx context.Context, options *options, msg string, code codes.Code, startTime time.Time, err error) {
	extractedLogger := ctxlg.Extract(ctx)
	level := options.levelFunc(code)
	args := []log.Attribute{
		log.Str("grpc.code", code.String()),
	}

	args = append(args, fromMapToAttributes(extractedLogger.Fields)...)
	args = append(args, options.durationFunc(startTime))

	switch level {
	case fabriclog.DebugLevel:
		log.Debug(ctx, msg, args...)
	case fabriclog.InfoLevel:
		log.Info(ctx, msg, args...)
	case fabriclog.ErrorLevel:
		log.Error(ctx, err, msg, args...)
	default:
		log.Info(ctx, msg, args...)
	}
}
