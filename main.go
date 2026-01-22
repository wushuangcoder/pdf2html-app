package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

func main() {
	// 设置Gin模式为发布模式，提高性能
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()

	r.GET("/pdf-to-html", func(c *gin.Context) {
		// 从查询参数中获取pdf_url
		pdfUrl := c.Query("pdf_url")
		if pdfUrl == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "pdf_url parameter is required"})
			return
		}

		// 获取信号量，限制并发请求数
		semaphore <- struct{}{}
		defer func() {
			<-semaphore
		}()

		// 创建临时目录 - 在系统临时目录内，确保隔离性
		tempDir, err := os.MkdirTemp("", "pdf2html-")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create temp directory: %v", err)})
			return
		}
		defer os.RemoveAll(tempDir)

		// 下载PDF文件
		pdfPath := filepath.Join(tempDir, "input.pdf")
		if err := downloadFile(pdfPath, pdfUrl); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to download PDF: %v", err)})
			return
		}

		// 使用pdf2htmlEX转换PDF为HTML，启用分页功能
		baseHtmlPath := filepath.Join(tempDir, "output.html")
		if err := convertPdfToHtml(pdfPath, baseHtmlPath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to convert PDF to HTML: %v", err)})
			return
		}

		// 提取分页HTML内容
		htmlPages, err := extractHtmlPages(tempDir)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to extract HTML pages: %v", err)})
			return
		}

		c.JSON(http.StatusOK, gin.H{"pages": htmlPages})
	})

	r.Run(":8080")
}

// 下载文件
func downloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// 使用pdf2htmlEX转换PDF为HTML
func convertPdfToHtml(pdfPath, htmlPath string) error {
	// 获取当前工作目录
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	// 打印调试信息
	fmt.Printf("Current working directory: %s\n", currentDir)
	fmt.Printf("PDF Path: %s\n", pdfPath)
	fmt.Printf("HTML Path: %s\n", htmlPath)

	// 使用优化的pdf2htmlEX选项
	// --split-pages 1 - 将页面分割为单独的文件
	// --page-filename "page-%d.html" - 分页文件命名模板
	// --dest-dir <tempDir> - 指定输出目录为临时目录
	// --embed "all" - 嵌入所有资源（CSS, 字体, 图片, JavaScript, 大纲）
	// --optimize-text 1 - 优化文本渲染，减少HTML元素数量
	// --correct-text-visibility 2 - 处理部分遮挡的文本
	// --process-nontext 1 - 渲染图形
	// --process-outline 1 - 显示大纲
	// --printing 1 - 支持打印
	// --quiet 1 - 静默模式，减少输出
	cmd := exec.Command(
		"pdf2htmlEX",
		"--split-pages", "1",
		"--page-filename", "page-%d.html",
		"--dest-dir", filepath.Dir(htmlPath),
		"--embed", "all",
		"--optimize-text", "1",
		"--correct-text-visibility", "2",
		"--process-nontext", "1",
		"--process-outline", "1",
		"--printing", "1",
		"--quiet", "1",
		pdfPath,
		htmlPath,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Command stdout: %s\n", stdout.String())
		fmt.Printf("Command stderr: %s\n", stderr.String())
		return fmt.Errorf("%w: %s", err, stderr.String())
	}

	fmt.Printf("Command succeeded, stdout: %s\n", stdout.String())
	return nil
}

// 页面信息结构体
type PageInfo struct {
	Page    int    `json:"page"`
	Content string `json:"content"`
}

// 并发控制信号量，限制最大并发数
var (
	semaphore = make(chan struct{}, 50) // 限制最大50个并发请求
	wg        sync.WaitGroup
)

// 提取分页HTML内容
func extractHtmlPages(tempDir string) ([]PageInfo, error) {
	// 打印调试信息
	fmt.Printf("Reading pages from directory: %s\n", tempDir)

	// 获取目录中的所有文件
	files, err := os.ReadDir(tempDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	// 打印目录中的所有文件
	fmt.Printf("All files in directory: ")
	var allFiles []string
	for _, file := range files {
		allFiles = append(allFiles, file.Name())
	}
	fmt.Printf("%v\n", allFiles)

	// 收集所有HTML文件，排除主文件output.html
	var pageFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".html") && file.Name() != "output.html" {
			pageFiles = append(pageFiles, file.Name())
		}
	}

	fmt.Printf("Found %d page files: %v\n", len(pageFiles), pageFiles)

	// 如果没有找到页面文件，返回错误
	if len(pageFiles) == 0 {
		return nil, fmt.Errorf("no page files found, pdf2htmlEX command splitting may not be supported with this version")
	}

	// 按照页码顺序读取文件
	var pages []PageInfo
	for i := 1; ; i++ {
		pageFile := fmt.Sprintf("page-%d.html", i)
		pagePath := filepath.Join(tempDir, pageFile)

		// 检查文件是否存在
		if _, err := os.Stat(pagePath); os.IsNotExist(err) {
			break // 没有更多页面文件了
		}

		// 读取页面内容
		content, err := os.ReadFile(pagePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read page file %s: %w", pageFile, err)
		}

		pageContent := strings.TrimSpace(string(content))
		pages = append(pages, PageInfo{
			Page:    i,
			Content: pageContent,
		})
		fmt.Printf("Read page %d, content length: %d\n", i, len(pageContent))
	}

	return pages, nil
}
