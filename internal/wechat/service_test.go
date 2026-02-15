package wechat

import (
	"net/http"
	"testing"

	"github.com/geekjourneyx/md2wechat-skill/internal/config"
	"go.uber.org/zap"
)

func TestCreateHTTPClient(t *testing.T) {
	tests := []struct {
		name        string
		proxy       string
		expectProxy bool
	}{
		{"no proxy", "", false},
		{"http proxy", "http://proxy.example.com:8080", true},
		{"with auth", "http://user:pass@proxy.example.com:8080", true},
		{"https proxy", "https://secure.proxy:443", true},
		{"invalid url", "://invalid", false}, // 降级为直连
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{WechatProxy: tt.proxy}
			svc := &Service{cfg: cfg, log: zap.NewNop()}
			client := svc.createHTTPClient()

			if tt.expectProxy {
				transport, ok := client.Transport.(*http.Transport)
				if !ok {
					t.Errorf("expected *http.Transport, got %T", client.Transport)
					return
				}
				if transport.Proxy == nil {
					t.Errorf("expected proxy to be set")
				}
			} else {
				if client.Transport != nil {
					t.Errorf("expected no transport (direct connection), got %T", client.Transport)
				}
			}
		})
	}
}
