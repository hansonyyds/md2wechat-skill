# 新手入门指南（5分钟上手）

> **你不需要懂编程！** 按照下面的步骤操作即可。

---

## 🚀 Claude Code 用户（最简单）

如果你使用 **Claude Code**，只需两步：

### 第一步：安装插件

在 Claude Code 中运行：

```bash
/plugin marketplace add hansonyyds/md2wechat-skill
/plugin install md2wechat@hansonyyds-md2wechat-skill
```

### 第二步：开始使用

直接和 Claude 对话：

```
请用秋日暖光主题将 article.md 转换为微信公众号格式
```

```
帮我把这篇技术文章转换后上传到微信草稿箱
```

完成！🎉

---

*如果你不使用 Claude Code，请继续阅读下面的内容。*

---

## 第一步：安装软件

### 选择你的系统，点击下载

| 你的系统 | 下载链接 | 安装位置 |
|----------|----------|----------|
| Windows 10/11 | [下载 .exe](https://github.com/hansonyyds/md2wechat-skill/releases/latest/download/md2wechat-windows-amd64.exe) | 任意文件夹或 `C:\Windows\System32\` |
| Mac (Intel芯片) | [下载](https://github.com/hansonyyds/md2wechat-skill/releases/latest/download/md2wechat-darwin-amd64) | `/usr/local/bin/` 或 `~/.local/bin/` |
| Mac (M1/M2芯片) | [下载](https://github.com/hansonyyds/md2wechat-skill/releases/latest/download/md2wechat-darwin-arm64) | `/usr/local/bin/` 或 `~/.local/bin/` |
| Linux | [下载](https://github.com/hansonyyds/md2wechat-skill/releases/latest/download/md2wechat-linux-amd64) | `/usr/local/bin/` 或 `~/.local/bin/` |

---

### 安装步骤（图文说明）

#### Windows 用户

1. 下载 `md2wechat-windows-amd64.exe`
2. 可以重命名为 `md2wechat.exe`（方便输入）
3. **方法 A（推荐）**：直接放到你想放的文件夹，用时打开 CMD 切换到那个文件夹
4. **方法 B（全局可用）**：复制到 `C:\Windows\System32\`
5. 测试：打开「命令提示符」或「PowerShell」，输入 `md2wechat --help`

#### Mac / Linux 用户

**方法一：一键安装（最简单）**

```bash
# 复制这条命令，粘贴到终端，回车
curl -fsSL https://raw.githubusercontent.com/hansonyyds/md2wechat-skill/main/scripts/install.sh | bash
```

**方法二：手动安装**

```bash
# 1. 下载
curl -Lo md2wechat https://github.com/hansonyyds/md2wechat-skill/releases/latest/download/md2wechat-linux-amd64

# 2. 添加执行权限
chmod +x md2wechat

# 3. 移动到系统目录
sudo mv md2wechat /usr/local/bin/

# 4. 测试
md2wechat --help
```

**方法三：用户目录安装（无需 sudo）**

```bash
# 1. 创建用户 bin 目录
mkdir -p ~/.local/bin

# 2. 下载到用户目录
curl -Lo ~/.local/bin/md2wechat https://github.com/hansonyyds/md2wechat-skill/releases/latest/download/md2wechat-linux-amd64

# 3. 添加执行权限
chmod +x ~/.local/bin/md2wechat

# 4. 添加到 PATH（只需一次）
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc   # 如果你用 bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc    # 如果你用 zsh
source ~/.bashrc   # 或 source ~/.zshrc

# 5. 测试
md2wechat --help
```

---

### 验证安装

输入以下命令，如果看到帮助信息，说明安装成功：

```bash
md2wechat --help
```

---

## 第二步：配置微信（只需1次）

### 2.1 获取微信公众号密码

1. 用浏览器打开：https://mp.weixin.qq.com
2. 登录你的公众号
3. 点击左上角「**设置与开发**」→「**基本配置**」
4. 复制这两个信息：
   - **开发者ID(AppID)**：类似 `wx1234567890abcdef`
   - **开发者密码(AppSecret)**：点击「重置」获取

### 2.2 生成配置文件

打开「**终端**」（Mac/Linux）或「**命令提示符**」（Windows）：

```bash
# 输入这个命令，回车
md2wechat config init
```

这会创建一个 `md2wechat.yaml` 文件，用记事本打开它。

### 2.3 填写你的信息

用记事本打开 `md2wechat.yaml`，修改这两行：

```yaml
wechat:
  appid: "wx1234567890abcdef"    # ← 粘贴你的 AppID
  secret: "your_secret_here"      # ← 粘贴你的 Secret
```

### 2.4 （可选）配置代理

如果访问微信 API 需要通过代理服务器，可以在配置文件中添加：

```yaml
api:
  wechat_proxy: "http://proxy.example.com:8080"  # 你的代理地址
```

或使用环境变量：
```bash
export WECHAT_PROXY="http://proxy.example.com:8080"
```

> 💡 **提示**：代理仅在调用微信 API 时使用。详细说明见 [FAQ.md](docs/FAQ.md#q21-如何配置微信-api-代理)

保存文件，完成！

---

## 第三步：开始使用

### 3.1 准备你的文章

你的文章用 Markdown 格式写，保存为 `我的文章.md`

**什么是 Markdown？**
- 一种简单的写作格式
- 用 `#` 表示标题
- 用 `![图片](地址)` 插入图片
- [Markdown 教程](https://www.markdownguide.org/zh-cn/basic-syntax/)

### 3.2 转换文章

```bash
# 预览效果（先看看怎么样）
md2wechat convert 我的文章.md --preview

# 满意后，直接发送到微信草稿箱
md2wechat convert 我的文章.md --draft
```

### 3.3 在微信中查看

1. 登录微信公众号后台
2. 点击「**新的创作**」→「**草稿箱**」
3. 你的文章已经在那里了！
4. 编辑后发表即可

---

## 常用命令一览

| 你想做什么 | 输入这个命令 |
|------------|--------------|
| 预览文章 | `md2wechat convert 文章.md --preview` |
| 发送到草稿箱 | `md2wechat convert 文章.md --draft` |
| 使用精美主题 | `md2wechat convert 文章.md --mode ai --theme autumn-warm` |
| 查看配置 | `md2wechat config show` |
| 检查配置是否正确 | `md2wechat config validate` |

---

## 精美主题推荐

| 命令 | 效果 |
|------|------|
| `--theme autumn-warm` | 🟠 秋日暖光（温暖治愈） |
| `--theme spring-fresh` | 🟢 春日清新（生机盎然） |
| `--theme ocean-calm` | 🔵 深海静谧（理性专业） |

**用法示例**：
```bash
md2wechat convert 我的文章.md --mode ai --theme autumn-warm --draft
```

---

## 遇到问题？

### 问题 1：提示 "命令不存在"

**Windows**：把下载的 `md2wechat.exe` 放到 `C:\Windows\System32\` 文件夹

**Mac/Linux**：
```bash
# 给文件执行权限
chmod +x md2wechat

# 移动到系统目录
sudo mv md2wechat /usr/local/bin/
```

### 问题 2：提示 "WECHAT_APPID is required"

说明你还没配置，回到「第二步：配置微信」

### 问题 3：图片没有上传

需要加 `--upload` 参数：
```bash
md2wechat convert 文章.md --upload --draft
```

---

## 完整示例

假设你有一篇文章叫 `产品发布.md`：

```bash
# 第一步：预览效果
md2wechat convert 产品发布.md --mode ai --theme autumn-warm --preview

# 第二步：满意后，上传图片并发送到草稿箱
md2wechat convert 产品发布.md --mode ai --theme autumn-warm --upload --draft
```

就这么简单！

---

## 下一步

- 查看 [使用教程](docs/USAGE.md) 了解更多功能
- 查看 [常见问题](docs/FAQ.md) 解决更多问题
