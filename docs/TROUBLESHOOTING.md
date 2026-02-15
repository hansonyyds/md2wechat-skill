# 故障排查向导

> 遇到问题？按照下面的步骤一步步排查

---

## 问题分类

点击你的问题跳转到解决方案：

- [安装问题](#安装问题)
- [配置问题](#配置问题)
- [转换问题](#转换问题)
- [图片问题](#图片问题)
- [微信问题](#微信问题)

---

## 安装问题

### ❓ 下载后双击没反应

**Windows**：
1. 右键点击 `md2wechat.exe`
2. 选择「属性」
3. 点击「解除锁定」（如果有）
4. 再双击运行

**Mac**：
1. 打开「系统偏好设置」→「安全性与隐私」
2. 点击「仍要打开」

---

### ❓ 提示 "命令不存在" 或 "不是内部或外部命令"

**原因**：程序没有添加到系统 PATH

#### Windows 解决方法：

**方法 A：简单方式（推荐）**
1. 把 `md2wechat.exe` 复制到 `C:\Windows\System32\`
2. 重新打开命令提示符

**方法 B：添加到 PATH**
1. 搜索「环境变量」→「编辑系统环境变量」
2. 点击「环境变量」
3. 在「用户变量」中找到 `Path`，点击「编辑」
4. 点击「新建」，输入程序所在目录
5. 点击「确定」保存

#### Mac/Linux 解决方法：

```bash
# 把程序移动到系统目录
sudo mv md2wechat /usr/local/bin/

# 如果提示没有这个目录，先创建
sudo mkdir -p /usr/local/bin
sudo mv md2wechat /usr/local/bin/
```

---

### ❓ Windows 提示 "Windows 已保护你的电脑"

1. 点击「更多信息」
2. 点击「仍要运行」

---

## 配置问题

### ❓ 提示 "WECHAT_APPID is required"

**原因**：还没有配置微信凭证

**解决步骤**：

1. **获取微信凭证**
   - 打开 https://developers.weixin.qq.com/platform
   - 登录后选择公众号 → 开发接口管理
   - 复制 AppID
   - 点击「重置」获取 AppSecret

2. **创建配置文件**
   ```bash
   md2wechat config init
   ```

3. **编辑配置文件**
   - 用记事本打开 `md2wechat.yaml`
   - 填入你的 AppID 和 Secret
   - 保存

4. **验证配置**
   ```bash
   md2wechat config validate
   ```

---

### ❓ 配置文件在哪？

配置文件会在你运行 `md2wechat config init` 的目录下创建。

**查找配置文件**：
- Windows：当前文件夹，如 `C:\Users\你的用户名\md2wechat.yaml`
- Mac/Linux：当前文件夹，如 `~/md2wechat.yaml`

---

## 转换问题

### ❓ 转换结果是空的

**可能原因 1**：文件路径不对

```bash
# 错误示例（文件不存在）
md2wechat convert 文章.md

# 正确示例（使用正确的文件名）
md2wechat convert 我的文章.md

# 或者使用完整路径
md2wechat convert /Users/你的名字/Documents/文章.md
```

**可能原因 2**：文件编码不是 UTF-8

- 用记事本打开文件
- 点击「另存为」
- 编码选择「UTF-8」
- 保存

---

### ❓ 中文显示乱码

**解决方法**：确保文件是 UTF-8 编码

1. 用 VS Code 或记事本打开文件
2. 点击「另存为」
3. 编码选择「UTF-8」
4. 保存

---

### ❓ AI 模式报错

**原因**：AI 模式需要 API Key

**解决方法 A**：使用 API 模式（更简单）
```bash
md2wechat convert 文章.md --mode api
```

**解决方法 B**：配置 AI API Key
1. 编辑 `md2wechat.yaml`
2. 添加：
   ```yaml
   api:
     image_key: "你的_claude_api_key"
   ```

---

## 图片问题

### ❓ 图片没有显示

**原因**：需要加上 `--upload` 参数

```bash
# 错误（不会上传图片）
md2wechat convert 文章.md

# 正确（上传图片）
md2wechat convert 文章.md --upload
```

---

### ❓ 图片上传失败

**可能原因 1**：图片格式不支持

**支持的格式**：`.jpg`、`.png`、`.gif`、`.bmp`、`.webp`

**不支持**：`.heic`（iPhone 默认格式）、`.tiff`

**解决方法**：用手机相册打开图片，选择「导出」为 JPEG 格式

---

**可能原因 2**：图片太大

程序会自动压缩，但如果仍然失败：

1. 用图片编辑器缩小图片
2. 或在配置文件中调整：
   ```yaml
   image:
     max_width: 1280  # 缩小最大宽度
     max_size_mb: 2   # 缩小最大文件大小
   ```

---

### ❓ 在线图片下载失败

**原因**：网络问题或图片链接有问题

**解决方法**：
1. 先用浏览器打开图片链接，确认能访问
2. 把图片下载到本地，使用本地图片：
   ```bash
   # 原来：![图片](https://...)
   # 改为：![图片](./images/图片.jpg)
   ```

---

## 微信问题

### ❓ 提示 "access_token expired"

**原因**：微信凭证过期

**解决方法**：
```bash
# 1. 验证配置
md2wechat config validate

# 2. 等待几分钟后重试
md2wechat convert 文章.md --draft
```

---

### ❓ 草稿创建失败

**可能原因 1**：公众号未认证

- 认证后的公众号才能使用草稿 API

**可能原因 2**：内容包含敏感词

- 尝试简化内容后重试
- 或先保存为 JSON：
  ```bash
  md2wechat convert 文章.md --save-draft draft.json
  ```

**可能原因 3**：API 调用次数超限

- 等待几分钟后重试
- 或联系微信提高限额

---

### ❓ 代理配置失败

**现象**：配置代理后仍无法访问微信 API，或提示 "invalid ip"

**原因 1：代理 URL 格式错误**

```bash
# 检查日志是否有以下错误
⚠️  invalid wechat proxy url, using direct connection
```

**解决方法**：
- 密码中的 `%` 需要转义为 `%25`
- 例如：`http://user:pass%word@host:port` → `http://user:pass%25word@host:port`

**原因 2：代理服务器 IP 未在微信白名单**

```
errcode=40164, errmsg=invalid ip xxx.xxx.xxx.xxx, not in whitelist
```

**解决方法**：
1. 获取代理服务器的公网 IP：
   ```bash
   curl ifconfig.me
   ```

2. 添加到微信白名单：
   - 访问 [微信开发者平台](https://developers.weixin.qq.com/platform)
   - 选择公众号 → 开发接口管理 → IP白名单
   - 添加代理服务器 IP

**原因 3：代理服务器无法访问**

```bash
# 测试代理连通性
curl -x http://proxy.example.com:8080 https://api.weixin.qq.com
```

**解决方法**：
- 检查代理服务器是否正常运行
- 检查防火墙是否阻止代理连接
- 尝试使用不同的代理协议（http/https）

---

## 需要更多帮助？

### 收集诊断信息

遇到问题时，运行以下命令收集信息：

```bash
# 1. 检查版本
md2wechat --version

# 2. 验证配置
md2wechat config validate

# 3. 查看配置（不显示密码）
md2wechat config show
```

### 获取支持

1. 查看 [常见问题](FAQ.md)
2. 查看 [使用教程](USAGE.md)
3. 提交 Issue：https://github.com/geekjourneyx/md2wechat-skill/issues

提交问题时，请附上：
- 你的操作系统（Windows 10 / macOS 13 / Linux）
- 错误信息的截图
- 运行的完整命令
