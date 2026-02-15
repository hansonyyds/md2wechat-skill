# HTTP 代理支持功能设计文档

**日期**: 2026-02-15
**状态**: 设计完成
**作者**: AI 设计会话

---

## 目录

- [需求概述](#需求概述)
- [设计原则](#设计原则)
- [架构设计](#架构设计)
- [配置层变更](#配置层变更)
- [服务层变更](#服务层变更)
- [错误处理与日志](#错误处理与日志)
- [测试策略](#测试策略)
- [文档更新](#文档更新)

---

## 需求概述

### 功能目标

为 md2wechat 项目添加 HTTP 代理支持，允许通过代理服务器访问微信公众号 API。

### 需求汇总

| 需求项 | 选择 |
|--------|------|
| 代理范围 | 仅微信 API |
| 配置方式 | 环境变量 + 配置文件（环境变量优先） |
| 代理协议 | HTTP/HTTPS |
| 代理认证 | 支持 `http://user:pass@host:port` |
| 无代理行为 | 直连 |

---

## 设计原则

1. **最小侵入**：在现有架构基础上添加代理支持，不破坏现有功能
2. **单一变更点**：代理配置只在 `Config` 层处理
3. **透明传递**：`WeChatService` 通过配置获取代理，无需关心来源
4. **Go 标准库**：使用 `net/http.ProxyURL`，零额外依赖

---

## 架构设计

### 架构变更点

```
┌─────────────────────────────────────────────────────────┐
│                      CLI Layer                          │
└─────────────────────────────────────────────────────────┘
                        │
┌───────────────────────┴───────────────────────────────┐
│              Config (配置层 - 唯一变更)                  │
│  - 新增 WechatProxy 字段                                │
│  - 优先级: ENV > YAML > 默认值                          │
└───────────────────────┬───────────────────────────────┘
                        │
┌───────────────────────┴───────────────────────────────┐
│            WeChat Service (服务层 - 轻微变更)           │
│  - 创建 HTTP Client 时配置 Proxy                        │
│  - 使用 http.ProxyURL 解析代理 URL                      │
│  - 通过 SDK 的 SetHTTPClient() 完整覆盖所有 API 调用    │
└─────────────────────────────────────────────────────────┘
```

### 覆盖范围确认

| 功能 | 调用方式 | 覆盖情况 |
|------|----------|----------|
| **素材上传** `UploadMaterial` | SDK: `mat.AddMaterial()` | ✅ 通过 SDK 客户端覆盖 |
| **草稿创建** `CreateDraft` | SDK: `dm.AddDraft()` | ✅ 通过 SDK 客户端覆盖 |
| **小绿书草稿** `CreateNewspicDraft` | 直接 HTTP | ✅ 使用自定义客户端 |
| **下载文件** `DownloadFile` | 直接 HTTP | ✅ 使用自定义客户端 |

---

## 配置层变更

### Config 结构体变更

**文件**: `internal/config/config.go`

```go
// Config 应用配置
type Config struct {
    // ... 现有字段 ...

    // 微信 API 代理配置（新增）
    WechatProxy string `json:"wechat_proxy" yaml:"wechat_proxy" env:"WECHAT_PROXY"`
}
```

### 配置文件格式

**YAML** (`~/.md2wechat.yaml` 或 `md2wechat.yaml`):
```yaml
api:
  wechat_proxy: "http://proxy.example.com:8080"
  http_timeout: 30
```

**JSON** (`md2wechat.json`):
```json
{
  "api": {
    "wechat_proxy": "http://user:pass@proxy.example.com:8080",
    "http_timeout": 30
  }
}
```

### 环境变量

```bash
export WECHAT_PROXY="http://proxy.example.com:8080"
```

### 配置加载优先级

```
WECHAT_PROXY (环境变量) > 配置文件 > "" (默认直连)
```

### loadFromEnv 函数新增

**文件**: `internal/config/config.go`

```go
func loadFromEnv(cfg *Config) {
    // ... 现有代码 ...

    // 新增代理配置
    if v := os.Getenv("WECHAT_PROXY"); v != "" {
        cfg.WechatProxy = v
    }
}
```

### loadFromYAML 函数新增

**文件**: `internal/config/config.go`

```go
func loadFromYAML(cfg *Config, data []byte) error {
    // ... 现有解析逻辑 ...

    // 新增代理配置映射
    if cf.API.WechatProxy != "" {
        cfg.WechatProxy = cf.API.WechatProxy
    }

    return nil
}
```

### configFile 结构体新增

```go
type configFile struct {
    Wechat struct {
        AppID  string `json:"appid" yaml:"appid"`
        Secret string `json:"secret" yaml:"secret"`
    } `json:"wechat" yaml:"wechat"`

    API struct {
        // ... 现有字段 ...
        WechatProxy string `json:"wechat_proxy" yaml:"wechat_proxy"` // 新增
    } `json:"api" yaml:"api"`

    // ... 其他字段 ...
}
```

---

## 服务层变更

### HTTP 客户端创建函数

**文件**: `internal/wechat/service.go`

```go
import (
    "net/http"
    "net/url"
)

// createHTTPClient 创建 HTTP 客户端，根据配置决定是否使用代理
func (s *Service) createHTTPClient() *http.Client {
    client := &http.Client{
        Timeout: 60 * time.Second,
    }

    // 如果配置了代理，设置 Proxy
    if s.cfg.WechatProxy != "" {
        if proxyURL, err := url.Parse(s.cfg.WechatProxy); err == nil {
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

### 初始化时设置 SDK 的 HTTP 客户端

**文件**: `internal/wechat/service.go`

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

### DownloadFile 函数变更

**文件**: `internal/wechat/service.go`

```go
// 原代码 (第 193-196 行)
client := &http.Client{
    Timeout: 60 * time.Second,
}

// 变更为
client := s.createHTTPClient()
```

### CreateNewspicDraft 函数变更

**文件**: `internal/wechat/service.go`

```go
// 原代码 (第 314 行)
httpResp, err := http.Post(apiURL, "application/json", bytes.NewReader(reqBody))

// 变更为
client := s.createHTTPClient()
httpResp, err := client.Post(apiURL, "application/json", bytes.NewReader(reqBody))
```

---

## 错误处理与日志

### 代理 URL 解析失败处理

当 `WechatProxy` 配置的 URL 格式不正确时：

```go
if proxyURL, err := url.Parse(s.cfg.WechatProxy); err == nil {
    client.Transport = &http.Transport{
        Proxy: http.ProxyURL(proxyURL),
    }
    s.log.Info("using wechat proxy", zap.String("proxy", s.cfg.WechatProxy))
} else {
    s.log.Warn("invalid wechat proxy url, using direct connection",
        zap.String("proxy", s.cfg.WechatProxy),
        zap.Error(err))
}
```

**策略**: 解析失败时记录警告日志，**降级为直连**，不阻塞程序运行。

### 代理连接失败处理

代理服务器不可用时，Go 的 `http.Transport` 会：
- 第一次请求超时（使用配置的 `Timeout`）
- 返回错误

**策略**: 由上层的业务逻辑（如 `UploadMaterialWithRetry`）处理重试。

### 日志输出示例

启动时显示代理状态：

```bash
# 有代理
✅ 使用配置文件: ~/.md2wechat.yaml
ℹ️  using wechat proxy: http://proxy.example.com:8080

# 代理解析失败
✅ 使用配置文件: ~/.md2wechat.yaml
⚠️  invalid wechat proxy url: "://invalid", using direct connection

# 无代理配置
✅ 使用配置文件: ~/.md2wechat.yaml
(直连，无额外日志)
```

---

## 测试策略

### 单元测试

**文件**: `internal/wechat/service_test.go`

```go
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
                if !ok || transport.Proxy == nil {
                    t.Errorf("expected proxy transport")
                }
            } else {
                if client.Transport != nil {
                    t.Errorf("expected no transport (direct connection)")
                }
            }
        })
    }
}
```

### 集成测试

**测试场景**:
1. 无代理配置 → 直连成功
2. 代理 URL 格式错误 → 降级为直连
3. 代理服务器不可达 → 超时重试

**注意**: 集成测试需要 mock 代理服务器，或在 CI 环境中跳过。

### 手动验证

```bash
# 1. 设置无效代理（验证降级）
export WECHAT_PROXY="://invalid-url"
md2wechat upload_image test.jpg
# 预期：警告日志，直连尝试

# 2. 设置有效代理
export WECHAT_PROXY="http://proxy.example.com:8080"
md2wechat upload_image test.jpg
# 预期：显示使用代理的日志

# 3. 验证配置文件
echo "api:
  wechat_proxy: http://proxy.example.com:8080" > md2wechat.yaml
md2wechat upload_image test.jpg
```

---

## 文档更新

### 1. 配置文档 (`docs/CONFIG.md`)

#### 新增环境变量说明

| 环境变量 | 对应配置项 | 说明 |
|----------|-----------|------|
| `WECHAT_PROXY` | `api.wechat_proxy` | 微信 API 代理地址（可选） |

#### 新增配置文件示例

```yaml
api:
  wechat_proxy: "http://proxy.example.com:8080"  # 可选
  http_timeout: 30
```

### 2. FAQ 新增条目 (`docs/FAQ.md`)

```markdown
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
```

### 3. OpenClaw 配置 (`docs/OPENCLAW.md`)

#### 新增代理配置示例

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

### 4. CHANGELOG.md

```markdown
## [x.y.z] - YYYY-MM-DD

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
```

---

## 实现检查清单

- [ ] 修改 `internal/config/config.go`
  - [ ] Config 结构体添加 `WechatProxy` 字段
  - [ ] `configFile` 结构体添加 `API.WechatProxy` 字段
  - [ ] `loadFromYAML` 添加代理配置映射
  - [ ] `loadFromJSON` 添加代理配置映射
  - [ ] `loadFromEnv` 添加环境变量读取
  - [ ] `SaveConfig` 添加代理配置保存

- [ ] 修改 `internal/wechat/service.go`
  - [ ] 新增 `createHTTPClient()` 方法
  - [ ] 修改 `NewService()` 调用 `wc.SetHTTPClient()`
  - [ ] 修改 `DownloadFile()` 使用自定义客户端
  - [ ] 修改 `CreateNewspicDraft()` 使用自定义客户端

- [ ] 添加单元测试
  - [ ] `TestCreateHTTPClient()` 测试各种代理配置

- [ ] 更新文档
  - [ ] `docs/CONFIG.md` 添加代理配置说明
  - [ ] `docs/FAQ.md` 添加 Q21
  - [ ] `docs/OPENCLAW.md` 添加代理配置示例
  - [ ] `CHANGELOG.md` 添加版本更新说明

---

## 总结

本设计方案通过最小侵入的方式为 md2wechat 项目添加了 HTTP 代理支持。核心变更点在配置层和服务层，利用 `silenceper/wechat` SDK 的 `SetHTTPClient()` 方法完整覆盖所有微信 API 调用。

| 设计项 | 内容 |
|--------|------|
| **代理范围** | 仅微信 API（通过 SDK 完整覆盖） |
| **配置方式** | 环境变量 + 配置文件（环境变量优先） |
| **代理协议** | HTTP/HTTPS |
| **代理认证** | 支持 `http://user:pass@host:port` |
| **无代理行为** | 直连 |
| **错误处理** | URL 解析失败降级为直连，记录警告日志 |
| **测试策略** | 单元测试 + 集成测试 + 手动验证 |
| **文档更新** | CONFIG.md、FAQ.md、OPENCLAW.md、CHANGELOG.md |
