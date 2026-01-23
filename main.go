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

/*
Usage: pdf2htmlEX [options] <input.pdf> [<output.html>]

	-f,--first-page <int>         first page to convert (default: 1)
	-l,--last-page <int>          last page to convert (default: 2147483647)
	--zoom <fp>                   zoom ratio
	--fit-width <fp>              fit width to <fp> pixels
	--fit-height <fp>             fit height to <fp> pixels
	--use-cropbox <int>           use CropBox instead of MediaBox (default: 1)
	--dpi <fp>                    Resolution for graphics in DPI (default: 144)
	--embed <string>              specify which elements should be embedded into output
	--embed-css <int>             embed CSS files into output (default: 1)
	--embed-font <int>            embed font files into output (default: 1)
	--embed-image <int>           embed image files into output (default: 1)
	--embed-javascript <int>      embed JavaScript files into output (default: 1)
	--embed-outline <int>         embed outlines into output (default: 1)
	--split-pages <int>           split pages into separate files (default: 0)
	--dest-dir <string>           specify destination directory (default: ".")
	--css-filename <string>       filename of the generated css file (default: "")
	--page-filename <string>      filename template for split pages  (default: "")
	--outline-filename <string>   filename of the generated outline file (default: "")
	--process-nontext <int>       render graphics in addition to text (default: 1)
	--process-outline <int>       show outline in HTML (default: 1)
	--process-annotation <int>    show annotation in HTML (default: 0)
	--process-form <int>          include text fields and radio buttons (default: 0)
	--printing <int>              enable printing support (default: 1)
	--fallback <int>              output in fallback mode (default: 0)
	--tmp-file-size-limit <int>   Maximum size (in KB) used by temporary files, -1 for no limit (default: -1)
	--embed-external-font <int>   embed local match for external fonts (default: 1)
	--font-format <string>        suffix for embedded font files (ttf,otf,woff,svg) (default: "woff")
	--decompose-ligature <int>    decompose ligatures, such as ﬁ -> fi (default: 0)
	--turn-off-ligatures <int>    explicitly tell browsers not to use ligatures (default: 0)
	--auto-hint <int>             use fontforge autohint on fonts without hints (default: 0)
	--external-hint-tool <string> external tool for hinting fonts (overrides --auto-hint) (default: "")
	--stretch-narrow-glyph <int>  stretch narrow glyphs instead of padding them (default: 0)
	--squeeze-wide-glyph <int>    shrink wide glyphs instead of truncating them (default: 1)
	--override-fstype <int>       clear the fstype bits in TTF/OTF fonts (default: 0)
	--process-type3 <int>         convert Type 3 fonts for web (experimental) (default: 0)
	--heps <fp>                   horizontal threshold for merging text, in pixels (default: 1)
	--veps <fp>                   vertical threshold for merging text, in pixels (default: 1)
	--space-threshold <fp>        word break threshold (threshold * em) (default: 0.125)
	--font-size-multiplier <fp>   a value greater than 1 increases the rendering accuracy (default: 4)
	--space-as-offset <int>       treat space characters as offsets (default: 0)
	--tounicode <int>             how to handle ToUnicode CMaps (0=auto, 1=force, -1=ignore) (default: 0)
	--optimize-text <int>         try to reduce the number of HTML elements used for text (default: 0)
	--correct-text-visibility <int> 0: Don't do text visibility checks. 1: Fully occluded text handled. 2: Partially occluded text handled (default: 1)
	--covered-text-dpi <fp>       Rendering DPI to use if correct-text-visibility == 2 and there is partially covered text on the page (default: 300)
	--bg-format <string>          specify background image format (default: "png")
	--svg-node-count-limit <int>  if node count in a svg background image exceeds this limit, fall back this page to bitmap background; negative value means no limit (default: -1)
	--svg-embed-bitmap <int>      1: embed bitmaps in svg background; 0: dump bitmaps to external files if possible (default: 1)
	-o,--owner-password <string>  owner password (for encrypted files)
	-u,--user-password <string>   user password (for encrypted files)
	--no-drm <int>                override document DRM settings (default: 0)
	--clean-tmp <int>             remove temporary files after conversion (default: 1)
	--tmp-dir <string>            specify the location of temporary directory (default: "/tmp")
	--data-dir <string>           specify data directory (default: "/usr/local/share/pdf2htmlEX")
	--poppler-data-dir <string>   specify poppler data directory (default: "/usr/local/share/pdf2htmlEX/poppler")
	--debug <int>                 print debugging information (default: 0)
	--proof <int>                 texts are drawn on both text layer and background for proof (default: 0)
	--quiet <int>                 perform operations quietly (default: 0)
	-v,--version                  print copyright and version info
	-h,--help                     print usage information
*/
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
		"--embed", "cfijo",
		"--optimize-text", "1",
		"--correct-text-visibility", "2",
		"--process-nontext", "1",
		"--process-outline", "1",
		"--printing", "1",
		"--quiet", "1",
		pdfPath,
		filepath.Base(htmlPath),
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
