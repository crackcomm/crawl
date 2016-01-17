package crawl

import (
	"testing"

	"golang.org/x/net/context"
)

// TestProxyFromContext -
func TestProxyFromContext(t *testing.T) {
	ctx := WithProxy(context.Background(), "a", "b")
	addrs, ok := ProxyFromContext(ctx)
	if !ok {
		t.Fail()
	}
	if addrs[0] != "a" || addrs[1] != "b" {
		t.Fail()
	}
}
