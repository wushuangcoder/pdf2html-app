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