# HTTP 代理支持功能实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 为 md2wechat 项目添加 HTTP 代理支持，允许通过代理服务器访问微信公众号 API

**Architecture:** 通过配置层添加 `WechatProxy` 字段，在服务层创建带代理的 HTTP 客户端，利用 silenceper/wechat SDK 的 `SetHTTPClient()` 方法完整覆盖所有微信 API 调用。

**Tech Stack:** Go 1.24+, silenceper/wechat v2.1.9, net/http, net/url

---

## Task 1: 添加配置层支持 - Config 结构体

**Files:**
- Modify: `internal/config/config.go:14-44`

**Step 1: 添加 WechatProxy 字段到 Config 结构体**

在 `Config` 结构体的 `HTTPTimeout` 字段后添加新字段：

```go
// Config 应用配置
type Config struct {
	// 微信公众号配置
	WechatAppID  string `json:"wechat_appid" yaml:"wechat_appid" env:"WECHAT_APPID"`
	WechatSecret string `json:"wechat_secret" yaml:"wechat_secret" env:"WECHAT_SECRET"`

	// md2wechat.cn API 配置
	MD2WechatAPIKey      string `json:"md2wechat_api_key" yaml:"md2wechat_api_key" env:"MD2WECHAT_API_KEY"`
	MD2WechatBaseURL     string `json:"md2wechat_base_url" yaml:"md2wechat_base_url" env:"MD2WECHAT_BASE_URL"`
	DefaultConvertMode   string `json:"default_convert_mode" yaml:"default_convert_mode" env:"CONVERT_MODE"`
	DefaultTheme         string `json:"default_theme" yaml:"default_theme" env:"DEFAULT_THEME"`
	DefaultBackgroundType string `json:"default_background_type" yaml:"default_background_type" env:"DEFAULT_BACKGROUND_TYPE"` // default/grid/none

	// 图片生成 API 配置
	ImageProvider string `json:"image_provider" yaml:"image_provider" env:"IMAGE_PROVIDER"`
	ImageAPIKey   string `json:"image_api_key" yaml:"image_api_key" env:"IMAGE_API_KEY"`
	ImageAPIBase  string `json:"image_api_base" yaml:"image_api_base" env:"IMAGE_API_BASE"`
	ImageModel    string `json:"image_model" yaml:"image_model" env:"IMAGE_MODEL"`
	ImageSize     string `json:"image_size" yaml:"image_size" env:"IMAGE_SIZE"`

	// 图片处理配置
	CompressImages bool  `json:"compress_images" yaml:"compress_images" env:"COMPRESS_IMAGES"`
	MaxImageWidth  int   `json:"max_image_width" yaml:"max_image_width" env:"MAX_IMAGE_WIDTH"`
	MaxImageSize   int64 `json:"max_image_size" yaml:"max_image_size" env:"MAX_IMAGE_SIZE"`

	// 超时配置
	HTTPTimeout int `json:"http_timeout" yaml:"http_timeout" env:"HTTP_TIMEOUT"`

	// 微信 API 代理配置
	WechatProxy string `json:"wechat_proxy" yaml:"wechat_proxy" env:"WECHAT_PROXY"`

	// 配置文件路径（用于追踪）
	configFile string
}
```

**Step 2: 保存文件**

**Step 3: 运行验证**

```bash
go build ./cmd/md2wechat
```

预期：编译成功

**Step 4: 提交**

```bash
git add internal/config/config.go
git commit -m "feat(config): add WechatProxy field to Config struct"
```

---

## Task 2: 添加配置文件结构支持 - configFile 结构体

**Files:**
- Modify: `internal/config/config.go:46-72`

**Step 1: 在 API 配置块中添加 WechatProxy 字段**

```go
// ConfigFile 配置文件结构（YAML/JSON）
type configFile struct {
	Wechat struct {
		AppID  string `json:"appid" yaml:"appid"`
		Secret string `json:"secret" yaml:"secret"`
	} `json:"wechat" yaml:"wechat"`

	API struct {
		MD2WechatKey    string `json:"md2wechat_key" yaml:"md2wechat_key"`
		MD2WechatBaseURL  string `json:"md2wechat_base_url" yaml:"md2wechat_base_url"`
		ImageKey           string `json:"image_key" yaml:"image_key"`
		ImageBaseURL       string `json:"image_base_url" yaml:"image_base_url"`
		ImageProvider      string `json:"image_provider" yaml:"image_provider"`
		ImageModel         string `json:"image_model" yaml:"image_model"`
		ImageSize          string `json:"image_size" yaml:"image_size"`
		ConvertMode        string `json:"convert_mode" yaml:"convert_mode"`
		DefaultTheme       string `json:"default_theme" yaml:"default_theme"`
		BackgroundType     string `json:"background_type" yaml:"background_type"`
		HTTPTimeout        int    `json:"http_timeout" yaml:"http_timeout"`
		WechatProxy        string `json:"wechat_proxy" yaml:"wechat_proxy"`
	} `json:"api" yaml:"api"`

	Image struct {
		Compress bool `json:"compress" yaml:"compress"`
		MaxWidth int  `json:"max_width" yaml:"max_width"`
		MaxSize  int  `json:"max_size_mb" yaml:"max_size_mb"`
	} `json:"image" yaml:"image"`
}
```

**Step 2: 保存文件**

**Step 3: 运行验证**

```bash
go build ./cmd/md2wechat
```

预期：编译成功

**Step 4: 提交**

```bash
git add internal/config/config.go
git commit -m "feat(config): add WechatProxy field to configFile struct"
```

---

## Task 3: 实现 loadFromYAML 代理配置映射

**Files:**
- Modify: `internal/config/config.go:184-240`

**Step 1: 在 loadFromYAML 函数中添加代理配置映射**

在 `if cf.API.HTTPTimeout > 0` 块之后添加：

```go
func loadFromYAML(cfg *Config, data []byte) error {
	var cf configFile
	if err := yaml.Unmarshal(data, &cf); err != nil {
		return fmt.Errorf("parse yaml: %w", err)
	}

	// 映射到 Config
	if cf.Wechat.AppID != "" {
		cfg.WechatAppID = cf.Wechat.AppID
	}
	if cf.Wechat.Secret != "" {
		cfg.WechatSecret = cf.Wechat.Secret
	}
	if cf.API.MD2WechatKey != "" {
		cfg.MD2WechatAPIKey = cf.API.MD2WechatKey
	}
	if cf.API.MD2WechatBaseURL != "" {
		cfg.MD2WechatBaseURL = cf.API.MD2WechatBaseURL
	}
	if cf.API.ImageKey != "" {
		cfg.ImageAPIKey = cf.API.ImageKey
	}
	if cf.API.ImageBaseURL != "" {
		cfg.ImageAPIBase = cf.API.ImageBaseURL
	}
	if cf.API.ImageProvider != "" {
		cfg.ImageProvider = cf.API.ImageProvider
	}
	if cf.API.ImageModel != "" {
		cfg.ImageModel = cf.API.ImageModel
	}
	if cf.API.ImageSize != "" {
		cfg.ImageSize = cf.API.ImageSize
	}
	if cf.API.ConvertMode != "" {
		cfg.DefaultConvertMode = cf.API.ConvertMode
	}
	if cf.API.DefaultTheme != "" {
		cfg.DefaultTheme = cf.API.DefaultTheme
	}
	if cf.API.BackgroundType != "" {
		cfg.DefaultBackgroundType = cf.API.BackgroundType
	}
	if cf.API.HTTPTimeout > 0 {
		cfg.HTTPTimeout = cf.API.HTTPTimeout
	}
	if cf.API.WechatProxy != "" {
		cfg.WechatProxy = cf.API.WechatProxy
	}
	cfg.CompressImages = cf.Image.Compress
	if cf.Image.MaxWidth > 0 {
		cfg.MaxImageWidth = cf.Image.MaxWidth
	}
	if cf.Image.MaxSize > 0 {
		cfg.MaxImageSize = int64(cf.Image.MaxSize) * 1024 * 1024
	}

	return nil
}
```

**Step 2: 保存文件**

**Step 3: 运行验证**

```bash
go build ./cmd/md2wechat
```

预期：编译成功

**Step 4: 提交**

```bash
git add internal/config/config.go
git commit -m "feat(config): add WechatProxy mapping in loadFromYAML"
```

---

## Task 4: 实现 loadFromJSON 代理配置映射

**Files:**
- Modify: `internal/config/config.go:242-298`

**Step 1: 在 loadFromJSON 函数中添加代理配置映射**

在 `if cf.API.HTTPTimeout > 0` 块之后添加：

```go
func loadFromJSON(cfg *Config, data []byte) error {
	var cf configFile
	if err := json.Unmarshal(data, &cf); err != nil {
		return fmt.Errorf("parse json: %w", err)
	}

	// 映射到 Config（与 loadFromYAML 相同的逻辑）
	if cf.Wechat.AppID != "" {
		cfg.WechatAppID = cf.Wechat.AppID
	}
	if cf.Wechat.Secret != "" {
		cfg.WechatSecret = cf.Wechat.Secret
	}
	if cf.API.MD2WechatKey != "" {
		cfg.MD2WechatAPIKey = cf.API.MD2WechatKey
	}
	if cf.API.MD2WechatBaseURL != "" {
		cfg.MD2WechatBaseURL = cf.API.MD2WechatBaseURL
	}
	if cf.API.ImageKey != "" {
		cfg.ImageAPIKey = cf.API.ImageKey
	}
	if cf.API.ImageBaseURL != "" {
		cfg.ImageAPIBase = cf.API.ImageBaseURL
	}
	if cf.API.ImageProvider != "" {
		cfg.ImageProvider = cf.API.ImageProvider
	}
	if cf.API.ImageModel != "" {
		cfg.ImageModel = cf.API.ImageModel
	}
	if cf.API.ImageSize != "" {
		cfg.ImageSize = cf.API.ImageSize
	}
	if cf.API.ConvertMode != "" {
		cfg.DefaultConvertMode = cf.API.ConvertMode
	}
	if cf.API.DefaultTheme != "" {
		cfg.DefaultTheme = cf.API.DefaultTheme
	}
	if cf.API.BackgroundType != "" {
		cfg.DefaultBackgroundType = cf.API.BackgroundType
	}
	if cf.API.HTTPTimeout > 0 {
		cfg.HTTPTimeout = cf.API.HTTPTimeout
	}
	if cf.API.WechatProxy != "" {
		cfg.WechatProxy = cf.API.WechatProxy
	}
	cfg.CompressImages = cf.Image.Compress
	if cf.Image.MaxWidth > 0 {
		cfg.MaxImageWidth = cf.Image.MaxWidth
	}
	if cf.Image.MaxSize > 0 {
		cfg.MaxImageSize = int64(cf.Image.MaxSize) * 1024 * 1024
	}

	return nil
}
```

**Step 2: 保存文件**

**Step 3: 运行验证**

```bash
go build ./cmd/md2wechat
```

预期：编译成功

**Step 4: 提交**

```bash
git add internal/config/config.go
git commit -m "feat(config): add WechatProxy mapping in loadFromJSON"
```

---

## Task 5: 实现 loadFromEnv 代理配置读取

**Files:**
- Modify: `internal/config/config.go:300-350`

**Step 1: 在 loadFromEnv 函数中添加代理配置读取**

在文件末尾、函数结束前添加：

```go
func loadFromEnv(cfg *Config) {
	if v := os.Getenv("WECHAT_APPID"); v != "" {
		cfg.WechatAppID = v
	}
	if v := os.Getenv("WECHAT_SECRET"); v != "" {
		cfg.WechatSecret = v
	}
	if v := os.Getenv("MD2WECHAT_API_KEY"); v != "" {
		cfg.MD2WechatAPIKey = v
	}
	if v := os.Getenv("MD2WECHAT_BASE_URL"); v != "" {
		cfg.MD2WechatBaseURL = v
	}
	if v := os.Getenv("CONVERT_MODE"); v != "" {
		cfg.DefaultConvertMode = v
	}
	if v := os.Getenv("DEFAULT_THEME"); v != "" {
		cfg.DefaultTheme = v
	}
	if v := os.Getenv("DEFAULT_BACKGROUND_TYPE"); v != "" {
		cfg.DefaultBackgroundType = v
	}
	if v := os.Getenv("IMAGE_API_KEY"); v != "" {
		cfg.ImageAPIKey = v
	}
	if v := os.Getenv("IMAGE_API_BASE"); v != "" {
		cfg.ImageAPIBase = v
	}
	if v := os.Getenv("IMAGE_PROVIDER"); v != "" {
		cfg.ImageProvider = v
	}
	if v := os.Getenv("IMAGE_MODEL"); v != "" {
		cfg.ImageModel = v
	}
	if v := os.Getenv("IMAGE_SIZE"); v != "" {
		cfg.ImageSize = v
	}
	if v := os.Getenv("COMPRESS_IMAGES"); v != "" {
		cfg.CompressImages = getEnvBool("COMPRESS_IMAGES", true)
	}
	if v := os.Getenv("MAX_IMAGE_WIDTH"); v != "" {
		cfg.MaxImageWidth = getEnvInt("MAX_IMAGE_WIDTH", cfg.MaxImageWidth)
	}
	if v := os.Getenv("MAX_IMAGE_SIZE"); v != "" {
		cfg.MaxImageSize = int64(getEnvInt("MAX_IMAGE_SIZE", int(cfg.MaxImageSize)))
	}
	if v := os.Getenv("HTTP_TIMEOUT"); v != "" {
		cfg.HTTPTimeout = getEnvInt("HTTP_TIMEOUT", cfg.HTTPTimeout)
	}
	if v := os.Getenv("WECHAT_PROXY"); v != "" {
		cfg.WechatProxy = v
	}
}
```

**Step 2: 保存文件**

**Step 3: 运行验证**

```bash
go build ./cmd/md2wechat
```

预期：编译成功

**Step 4: 提交**

```bash
git add internal/config/config.go
git commit -m "feat(config): add WECHAT_PROXY environment variable support"
```

---

## Task 6: 实现 SaveConfig 代理配置保存

**Files:**
- Modify: `internal/config/config.go:449-497`

**Step 1: 在 SaveConfig 函数中添加代理配置保存**

在 `cf.API.HTTPTimeout = cfg.HTTPTimeout` 之后添加：

```go
func SaveConfig(path string, cfg *Config) error {
	ext := strings.ToLower(filepath.Ext(path))

	cf := configFile{}
	cf.Wechat.AppID = cfg.WechatAppID
	cf.Wechat.Secret = cfg.WechatSecret
	cf.API.MD2WechatKey = cfg.MD2WechatAPIKey
	cf.API.MD2WechatBaseURL = cfg.MD2WechatBaseURL
	cf.API.ImageKey = cfg.ImageAPIKey
	cf.API.ImageBaseURL = cfg.ImageAPIBase
	cf.API.ImageProvider = cfg.ImageProvider
	cf.API.ImageModel = cfg.ImageModel
	cf.API.ImageSize = cfg.ImageSize
	cf.API.ConvertMode = cfg.DefaultConvertMode
	cf.API.DefaultTheme = cfg.DefaultTheme
	cf.API.BackgroundType = cfg.DefaultBackgroundType
	cf.API.HTTPTimeout = cfg.HTTPTimeout
	cf.API.WechatProxy = cfg.WechatProxy
	cf.Image.Compress = cfg.CompressImages
	cf.Image.MaxWidth = cfg.MaxImageWidth
	cf.Image.MaxSize = int(cfg.MaxImageSize / 1024 / 1024)

	var data []byte
	var err error

	if ext == ".json" {
		data, err = json.MarshalIndent(cf, "", "  ")
	} else {
		data, err = yaml.Marshal(cf)
	}

	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	// 确保目录存在
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("create directory: %w", err)
		}
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	return nil
}
```

**Step 2: 保存文件**

**Step 3: 运行验证**

```bash
go build ./cmd/md2wechat
```

预期：编译成功

**Step 4: 提交**

```bash
git add internal/config/config.go
git commit -m "feat(config): add WechatProxy to SaveConfig function"
```

---

## Task 7: 实现 ToMap 代理配置显示

**Files:**
- Modify: `internal/config/config.go:425-447`

**Step 1: 在 ToMap 函数中添加代理配置**

在 `http_timeout` 字段之后添加：

```go
func (c *Config) ToMap(maskSecret bool) map[string]any {
	result := map[string]any{
		"wechat_appid":           c.WechatAppID,
		"wechat_secret":          maskIf(c.WechatSecret, maskSecret),
		"default_convert_mode":   c.DefaultConvertMode,
		"default_theme":          c.DefaultTheme,
		"default_background_type": c.DefaultBackgroundType,
		"md2wechat_api_key":      maskIf(c.MD2WechatAPIKey, maskSecret),
		"md2wechat_base_url":     c.MD2WechatBaseURL,
		"image_provider":         c.ImageProvider,
		"image_api_key":          maskIf(c.ImageAPIKey, maskSecret),
		"image_api_base":         c.ImageAPIBase,
		"image_model":            c.ImageModel,
		"image_size":             c.ImageSize,
		"compress_images":        c.CompressImages,
		"max_image_width":        c.MaxImageWidth,
		"max_image_size_mb":      c.MaxImageSize / 1024 / 1024,
		"http_timeout":           c.HTTPTimeout,
		"wechat_proxy":           c.WechatProxy,
		"config_file":            c.configFile,
	}
	return result
}
```

**Step 2: 保存文件**

**Step 3: 运行验证**

```bash
go build ./cmd/md2wechat
```

预期：编译成功

**Step 4: 提交**

```bash
git add internal/config/config.go
git commit -m "feat(config): add WechatProxy to ToMap function"
```

---

## Task 8: 添加 wechat 服务层导入

**Files:**
- Modify: `internal/wechat/service.go:1-24`

**Step 1: 添加 net/url 导入**

在现有导入后添加：

```go
package wechat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	neturl "net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
	"url" "net/url"

	"github.com/hansonyyds/md2wechat-skill/internal/config"
	"github.com/silenceper/wechat/v2"
	wechatcache "github.com/silenceper/wechat/v2/cache"
	"github.com/silenceper/wechat/v2/officialaccount"
	wechatconfig "github.com/silenceper/wechat/v2/officialaccount/config"
	"github.com/silenceper/wechat/v2/officialaccount/draft"
	"github.com/silenceper/wechat/v2/officialaccount/material"
	"go.uber.org/zap"
)
```

**Step 2: 保存文件**

**Step 3: 运行验证**

```bash
go build ./cmd/md2wechat
```

预期：编译成功

**Step 4: 提交**

```bash
git add internal/wechat/service.go
git commit -m "feat(wechat): add net/url import for proxy support"
```

---

## Task 9: 实现 createHTTPClient 方法

**Files:**
- Modify: `internal/wechat/service.go` (在 NewService 函数后添加)

**Step 1: 在 NewService 函数后添加 createHTTPClient 方法**

```go
// NewService 创建微信服务
func NewService(cfg *config.Config, log *zap.Logger) *Service {
	return &Service{
		cfg: cfg,
		log: log,
		wc:  wechat.NewWechat(),
	}
}

// createHTTPClient 创建 HTTP 客户端，根据配置决定是否使用代理
func (s *Service) createHTTPClient() *http.Client {
	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	// 如果配置了代理，设置 Proxy
	if s.cfg.WechatProxy != "" {
		if proxyURL, err := neturl.Parse(s.cfg.WechatProxy); err == nil {
			client.Transport = &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			}
			s.log.Info("using wechat proxy", zap.String("proxy", s.cfg.WechatProxy))
		} else {
			s.log.Warn("invalid wechat proxy url, using direct connection",
				zap.String("proxy", s.cfg.WechatProxy),
				zap.Error(err))
		}
	}

	return client
}
```

**Step 2: 保存文件**

**Step 3: 运行验证**

```bash
go build ./cmd/md2wechat
```

预期：编译成功

**Step 4: 提交**

```bash
git add internal/wechat/service.go
git commit -m "feat(wechat): add createHTTPClient method with proxy support"
```

---

## Task 10: 修改 NewService 使用自定义 HTTP 客户端

**Files:**
- Modify: `internal/wechat/service.go:33-40`

**Step 1: 修改 NewService 函数以设置 SDK 的 HTTP 客户端**

```go
// NewService 创建微信服务
func NewService(cfg *config.Config, log *zap.Logger) *Service {
	wc := wechat.NewWechat()

	// 设置自定义 HTTP 客户端（支持代理）
	svc := &Service{
		cfg: cfg,
		log: log,
		wc:  wc,
	}

	wc.SetHTTPClient(svc.createHTTPClient())

	return svc
}
```

**Step 2: 保存文件**

**Step 3: 运行验证**

```bash
go build ./cmd/md2wechat
```

预期：编译成功

**Step 4: 提交**

```bash
git add internal/wechat/service.go
git commit -m "feat(wechat): configure SDK to use custom HTTP client"
```

---

## Task 11: 修改 DownloadFile 使用自定义 HTTP 客户端

**Files:**
- Modify: `internal/wechat/service.go:178-232`

**Step 1: 将 DownloadFile 中的 HTTP 客户端创建改为使用 createHTTPClient**

找到：
```go
	// HTTP URL - 下载文件
	url := urlOrPath

	// 创建 HTTP 客户端
	client := &http.Client{
		Timeout: 60 * time.Second,
	}
```

修改为：
```go
	// HTTP URL - 下载文件
	url := urlOrPath

	// 创建 HTTP 客户端（支持代理）
	client := s.createHTTPClient()
```

**Step 2: 保存文件**

**Step 3: 运行验证**

```bash
go build ./cmd/md2wechat
```

预期：编译成功

**Step 4: 提交**

```bash
git add internal/wechat/service.go
git commit -m "feat(wechat): use custom HTTP client in DownloadFile"
```

---

## Task 12: 修改 CreateNewspicDraft 使用自定义 HTTP 客户端

**Files:**
- Modify: `internal/wechat/service.go:293-348`

**Step 1: 将 CreateNewspicDraft 中的 http.Post 改为使用自定义客户端**

找到：
```go
	// 调用微信 API
	apiURL := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/draft/add?access_token=%s", accessToken)

	httpResp, err := http.Post(apiURL, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("call wechat api: %w", err)
	}
	defer httpResp.Body.Close()
```

修改为：
```go
	// 调用微信 API
	apiURL := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/draft/add?access_token=%s", accessToken)

	client := s.createHTTPClient()
	httpResp, err := client.Post(apiURL, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("call wechat api: %w", err)
	}
	defer httpResp.Body.Close()
```

**Step 2: 保存文件**

**Step 3: 运行验证**

```bash
go build ./cmd/md2wechat
```

预期：编译成功

**Step 4: 提交**

```bash
git add internal/wechat/service.go
git commit -m "feat(wechat): use custom HTTP client in CreateNewspicDraft"
```

---

## Task 13: 编写单元测试

**Files:**
- Create: `internal/wechat/service_test.go`

**Step 1: 创建测试文件**

```go
package wechat

import (
	"net/http"
	"testing"

	"github.com/hansonyyds/md2wechat-skill/internal/config"
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
```

**Step 2: 运行测试**

```bash
go test ./internal/wechat -v
```

预期：所有测试通过

**Step 3: 提交**

```bash
git add internal/wechat/service_test.go
git commit -m "test(wechat): add unit tests for createHTTPClient"
```

---

## Task 14: 更新 CONFIG.md 文档

**Files:**
- Modify: `docs/CONFIG.md`

**Step 1: 在环境变量表中添加 WECHAT_PROXY**

找到环境变量表格，添加新行：

```markdown
| 环境变量 | 对应配置项 | 说明 |
|----------|-----------|------|
| `WECHAT_APPID` | `wechat.appid` | 微信 AppID |
| `WECHAT_SECRET` | `wechat.secret` | 微信 Secret |
| `WECHAT_PROXY` | `api.wechat_proxy` | 微信 API 代理地址（可选） |
| `MD2WECHAT_API_KEY` | `api.md2wechat_key` | md2wechat.cn API Key |
```

**Step 2: 在配置文件示例中添加代理配置**

找到 YAML 配置示例，添加代理字段：

```yaml
# 微信公众号配置
wechat:
  appid: "wx1234567890abcdef"
  secret: "your_secret_here"

# API 配置
api:
  md2wechat_key: "your_md2wechat_key"
  wechat_proxy: "http://proxy.example.com:8080"  # 可选
  http_timeout: 30
```

**Step 3: 保存文件**

**Step 4: 验证 Markdown 格式**

```bash
# 如有 markdownlint 可运行
# markdownlint docs/CONFIG.md
```

**Step 5: 提交**

```bash
git add docs/CONFIG.md
git commit -m "docs(CONFIG): add WECHAT_PROXY configuration documentation"
```

---

## Task 15: 更新 FAQ.md 文档

**Files:**
- Modify: `docs/FAQ.md`

**Step 1: 在文档末尾添加新问题**

在 "## 仍然无法解决？" 之前添加：

```markdown
---

### Q21: 如何配置微信 API 代理？

**原因**: 访问微信 API 需要通过代理服务器

**解决方案 A**: 使用环境变量

```bash
export WECHAT_PROXY="http://proxy.example.com:8080"
```

**解决方案 B**: 使用配置文件

```yaml
# md2wechat.yaml
api:
  wechat_proxy: "http://user:pass@proxy.example.com:8080"
```

**支持的格式**:
- HTTP: `http://host:port`
- 带认证: `http://user:pass@host:port`
- HTTPS: `https://host:port`

**优先级**: 环境变量 > 配置文件 > 直连

**验证配置**:

```bash
# 查看当前配置
md2wechat config show

# 测试代理连接
md2wechat upload_image test.jpg
```

---

```

**Step 2: 保存文件**

**Step 3: 验证 Markdown 格式

**Step 4: 提交**

```bash
git add docs/FAQ.md
git commit -m "docs(FAQ): add Q21 about HTTP proxy configuration"
```

---

## Task 16: 更新 OPENCLAW.md 文档

**Files:**
- Modify: `docs/OPENCLAW.md`

**Step 1: 在配置示例中添加代理配置**

找到 OpenClaw 配置文件示例，添加 `WECHAT_PROXY`：

```json
{
  "skills": {
    "entries": {
      "md2wechat": {
        "enabled": true,
        "env": {
          "WECHAT_APPID": "你的AppID",
          "WECHAT_SECRET": "你的Secret",
          "WECHAT_PROXY": "http://proxy.example.com:8080"
        }
      }
    }
  }
}
```

**Step 2: 在环境变量表中添加 WECHAT_PROXY**

找到环境变量表格，添加新行：

```markdown
| 环境变量 | 必需 | 说明 | 获取方式 |
|---------|------|------|---------|
| `WECHAT_APPID` | 草稿上传时 | 微信公众号 AppID | [微信开发者平台](https://developers.weixin.qq.com/platform) → 开发接口管理 |
| `WECHAT_SECRET` | 草稿上传时 | 微信公众号 Secret | 同上，点击"重置"获取 |
| `WECHAT_PROXY` | 可选 | 微信 API 代理地址 | 你的代理服务器地址 |
| `IMAGE_API_KEY` | AI 图片时 | 图片生成 API Key | 见 [图片服务配置](IMAGE_PROVISIONERS.md) |
```

**Step 3: 保存文件**

**Step 4: 提交**

```bash
git add docs/OPENCLAW.md
git commit -m "docs(OPENCLAW): add WECHAT_PROXY configuration example"
```

---

## Task 17: 更新 CHANGELOG.md

**Files:**
- Modify: `CHANGELOG.md`

**Step 1: 在文件顶部添加新版本**

```markdown
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **HTTP Proxy**: 支持微信 API 代理配置
  - 环境变量: `WECHAT_PROXY`
  - 配置文件: `api.wechat_proxy`
  - 支持 HTTP/HTTPS 代理和认证

### Technical Details
- **Modified Files**:
  - `internal/config/config.go` - 新增 WechatProxy 字段及加载逻辑
  - `internal/wechat/service.go` - 代理支持的 HTTP 客户端创建
  - `docs/CONFIG.md`, `docs/FAQ.md`, `docs/OPENCLAW.md` - 文档更新

---

## [1.11.0] - 2026-02-15
```

**Step 2: 保存文件**

**Step 3: 提交**

```bash
git add CHANGELOG.md
git commit -m "docs(CHANGELOG): add HTTP proxy feature to Unreleased section"
```

---

## Task 18: 运行完整测试验证

**Step 1: 运行所有测试**

```bash
go test ./... -v
```

预期：所有测试通过

**Step 2: 编译验证**

```bash
go build ./cmd/md2wechat
```

预期：编译成功

**Step 3: 功能测试（可选，如已配置微信凭证）**

```bash
# 测试无代理配置
./md2wechat config show

# 测试环境变量
export WECHAT_PROXY="http://proxy.example.com:8080"
./md2wechat config show
```

**Step 4: 提交**

```bash
# 如有测试文件修改
git add internal/wechat/service_test.go
git commit -m "test: verify all tests pass with proxy support"
```

---

## Task 19: 推送到远程分支

**Step 1: 推送 feature 分支到远程**

```bash
git push -u origin feature/http-proxy-support
```

**Step 2: 验证远程分支**

```bash
git branch -r | grep http-proxy
```

预期：看到 `origin/feature/http-proxy-support`

**Step 3: 创建 Pull Request（可选）**

```bash
gh pr create --title "feat: add HTTP proxy support for WeChat API" --body "Implement HTTP proxy support for WeChat API calls. See design doc: docs/plans/2026-02-15-http-proxy-support-design.md"
```

---

## 实现检查清单

完成以上所有任务后，确认：

- [ ] Config 结构体添加 `WechatProxy` 字段
- [ ] configFile 结构体添加 `API.WechatProxy` 字段
- [ ] loadFromYAML 添加代理配置映射
- [ ] loadFromJSON 添加代理配置映射
- [ ] loadFromEnv 添加环境变量读取
- [ ] SaveConfig 添加代理配置保存
- [ ] ToMap 添加代理配置显示
- [ ] wechat service 添加 net/url 导入
- [ ] 实现 createHTTPClient 方法
- [ ] NewService 设置 SDK HTTP 客户端
- [ ] DownloadFile 使用自定义客户端
- [ ] CreateNewspicDraft 使用自定义客户端
- [ ] 单元测试通过
- [ ] CONFIG.md 文档更新
- [ ] FAQ.md 文档更新
- [ ] OPENCLAW.md 文档更新
- [ ] CHANGELOG.md 文档更新
- [ ] 所有测试通过
- [ ] 代码已推送到远程

---

**完成后**: 所有功能已实现，文档已更新，测试通过，代码已推送到 `feature/http-proxy-support` 分支，准备合并到 main。
