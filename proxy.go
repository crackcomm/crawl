package crawl

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
)

var proxyKey = "crawl_proxy"

// WithProxy - Sets proxies in context metadata.
func WithProxy(ctx context.Context, addrs ...string) context.Context {
	if md, ok := metadata.FromContext(ctx); ok {
		md[proxyKey] = append(md[proxyKey], addrs...)
		return ctx
	}
	return metadata.NewContext(ctx, metadata.MD{proxyKey: addrs})
}

// ProxyFromContext - Returns proxy from context metadata.
func ProxyFromContext(ctx context.Context) (addrs []string, ok bool) {
	md, ok := metadata.FromContext(ctx)
	if ok {
		addrs = md[proxyKey]
	}
	return
}
