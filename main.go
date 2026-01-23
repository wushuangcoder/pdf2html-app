package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gin-gonic/gin"
) /*
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

		// 直接返回转换后的HTML文件
		outputHtmlPath := filepath.Join(tempDir, "output.html")
		if _, err := os.Stat(outputHtmlPath); os.IsNotExist(err) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find output HTML file"})
			return
		}

		// 设置响应头，指定文件类型为HTML
		c.Header("Content-Type", "text/html")
		// 返回文件
		c.File(outputHtmlPath)
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

	cmd := exec.Command(
		"pdf2htmlEX",
		"--dest-dir", filepath.Dir(htmlPath),
		"--embed-css", "1",
		"--embed-font", "1",
		"--embed-image", "1",
		"--embed-javascript", "1",
		"--embed-outline", "1",
		"--css-filename", "",
		"--outline-filename", "",
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

// 并发控制信号量，限制最大并发数
var (
	semaphore = make(chan struct{}, 50) // 限制最大50个并发请求
)
