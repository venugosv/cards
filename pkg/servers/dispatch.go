package servers

import (
	"context"
	"net/http"
	"strings"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// Use x/net/http2/h2c so we can have http2 cleartext connections. The default
// Go http server does not support it. We also cannot plug into the grpc
// http2 server.
// From: https://github.com/philips/grpc-gateway-example/issues/22#issuecomment-490733965
func grpcDispatcher(ctx context.Context, grpcHandler http.Handler, httpHandler http.Handler) http.Handler {
	hf := func(w http.ResponseWriter, r *http.Request) {
		// Set the request context
		// We cannot use the http.OpsServer BaseContext field due to an unexpected issue with istio (see relevant issue for details)
		req := r.WithContext(ctx)

		contentTypeHeader := r.Header.Get(contentType)

		if r.ProtoMajor == 2 && strings.HasPrefix(contentTypeHeader, grpcContentType) {
			logf.Debug(ctx, "%s: \"%s\", routing to gRPC server", contentType, contentTypeHeader)
			grpcHandler.ServeHTTP(w, req)
		} else {
			logf.Debug(ctx, "%s: \"%s\", routing to HTTP server", contentType, contentTypeHeader)
			httpHandler.ServeHTTP(w, req)
		}
	}
	return h2c.NewHandler(http.HandlerFunc(hf), &http2.Server{})
}
