# PDF转HTML在线转换工具

这是一个基于Go和Gin开发的在线PDF转HTML应用，使用pdf2htmlEX工具实现转换功能。

## 功能特点

- 支持从网络URL转换PDF文件为HTML
- 分页返回HTML内容，每个页面包含页码和HTML内容
- 使用Docker容器化，环境统一
- 转换完成后自动清理临时文件，不占用空间

## 使用方法

### 构建和运行

```bash
docker build -t pdf2html-app .
docker run -d -p 8080:8080 pdf2html-app
```

### API调用

```bash
curl "http://localhost:8080/pdf-to-html?pdf_url=https://example.com/path/to/your.pdf"
```

### 响应格式

```json
{
  "pages": [
    {
      "page": 1,
      "content": "<html>...</html>"
    },
    {
      "page": 2,
      "content": "<html>...</html>"
    }
  ]
}
```

## 技术栈

- **Go**: 编程语言
- **Gin**: Web框架
- **pdf2htmlEX**: PDF转HTML工具
- **Docker**: 容器化平台

## 项目结构

```
pdf2html/
├── main.go          // 主应用代码
├── go.mod           // Go模块定义
├── go.sum           // 依赖锁定文件
├── Dockerfile       // Docker构建配置
└── docker-compose.yml // Docker Compose配置
```


以下是 `pdf2htmlEX` 工具的中文使用说明（基于你提供的英文帮助信息整理）：

---

# pdf2htmlEX 使用说明

**基本用法：**
```bash
pdf2htmlEX [选项] <输入文件.pdf> [<输出文件.html>]
```

将 PDF 文档转换为 HTML 格式，保留文本可选、搜索、缩放等特性，并支持嵌入字体、图片、CSS 等资源。

---

## 常用选项说明

### 页面范围控制
- `-f, --first-page <整数>`  
  起始转换页码（默认：1）
- `-l, --last-page <整数>`  
  结束转换页码（默认：2147483647，即全部页面）

### 输出尺寸与分辨率
- `--zoom <浮点数>`  
  设置缩放比例（例如 1.5 表示放大 1.5 倍）
- `--fit-width <像素值>`  
  将页面宽度适配到指定像素
- `--fit-height <像素值>`  
  将页面高度适配到指定像素
- `--dpi <浮点数>`  
  图形渲染分辨率（DPI，默认：144）

### 资源嵌入控制
- `--embed <字符串>`  
  指定要嵌入到 HTML 中的资源类型，使用以下字母组合：
    - `c`：CSS 样式
    - `f`：字体（Font）
    - `i`：图片（Image）
    - `j`：JavaScript 脚本
    - `o`：文档大纲（Outline）  
      示例：`--embed cfijo` 表示嵌入所有资源。


- `--embed-css <0/1>`  
  是否嵌入 CSS（默认：1）
- `--embed-font <0/1>`  
  是否嵌入字体（默认：1）
- `--embed-image <0/1>`  
  是否嵌入图片（默认：1）
- `--embed-javascript <0/1>`  
  是否嵌入 JavaScript（默认：1）
- `--embed-outline <0/1>`  
  是否嵌入文档大纲（默认：1）

### 输出结构
- `--split-pages <0/1>`  
  是否将每页输出为单独的 HTML 文件（默认：0）
- `--dest-dir <路径>`  
  指定输出目录（默认：当前目录）
- `--page-filename <模板>`  
  分页输出时的文件名模板，例如 `page-%d.html`（`%d` 会被替换为页码）
- `--css-filename <文件名>`  
  自定义生成的 CSS 文件名
- `--outline-filename <文件名>`  
  自定义大纲文件名

### 内容处理
- `--process-nontext <0/1>`  
  是否渲染非文本内容（如图形、背景，默认：1）
- `--process-outline <0/1>`  
  是否在 HTML 中显示文档大纲（书签，默认：1）
- `--process-annotation <0/1>`  
  是否显示注释（默认：0）
- `--process-form <0/1>`  
  是否包含表单字段（如文本框、单选按钮，默认：0）
- `--printing <0/1>`  
  启用打印支持（默认：1）

### 文本优化
- `--optimize-text <0/1>`  
  尝试减少用于文本的 HTML 元素数量（默认：0）
- `--correct-text-visibility <0/1/2>`  
  处理被遮挡的文本：
    - `0`：不检查
    - `1`：处理完全被遮挡的文本
    - `2`：处理部分被遮挡的文本（默认：1）
- `--covered-text-dpi <浮点数>`  
  当启用 `--correct-text-visibility=2` 时，用于渲染被遮挡文本的 DPI（默认：300）

### 字体与排版
- `--font-format <格式>`  
  嵌入字体的格式（可选：ttf / otf / woff / svg，默认：woff）
- `--decompose-ligature <0/1>`  
  拆分连字（如 “ﬁ” → “fi”，默认：0）
- `--auto-hint <0/1>`  
  对无提示信息的字体自动添加 hint（需 FontForge， 默认：0）

### 其他
- `--quiet <0/1>`  
  静默模式，不输出日志（默认：0）
- `--tmp-dir <路径>`  
  指定临时文件目录（默认：`/tmp`）
- `--clean-tmp <0/1>`  
  转换完成后是否删除临时文件（默认：1）
- `-o, --owner-password <密码>`  
  PDF 所有者密码（用于解密）
- `-u, --user-password <密码>`  
  PDF 用户密码（用于解密）

### 帮助与版本
- `-h, --help`  
  显示帮助信息
- `-v, --version`  
  显示版本和版权信息

---

## 使用示例

### 基础转换（嵌入所有资源，单页输出）
```bash
pdf2htmlEX --embed cfijo document.pdf output.html
```

### 分页输出到指定目录
```bash
pdf2htmlEX \
  --split-pages 1 \
  --page-filename "page-%d.html" \
  --dest-dir ./html_output \
  --embed cfi \
  document.pdf
```

> 此命令会生成 `./html_output/page-1.html`, `page-2.html` 等，并嵌入 CSS、字体和图片。

---

## 注意事项
- 确保输出目录存在且有写权限。
- 若 PDF 受密码保护，请使用 `-o` 或 `-u` 提供密码。
--- 

如需进一步调试，可暂时关闭 `--quiet 1` 查看详细日志。