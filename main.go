package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

// 时间戳格式常量
const TimeFormat = "2006-01-02 15:04:05" /*
使用方法: pdf2htmlEX [选项] <input.pdf> [<output.html>]

	-f,--first-page <int>         要转换的第一页 (默认: 1)
	-l,--last-page <int>          要转换的最后一页 (默认: 2147483647)
	--zoom <fp>                   缩放比例
	--fit-width <fp>              将页面宽度适配到指定像素
	--fit-height <fp>             将页面高度适配到指定像素
	--use-cropbox <int>           使用CropBox而不是MediaBox (默认: 1)
	--dpi <fp>                    图形渲染分辨率 (DPI, 默认: 144)
	--embed <string>              指定要嵌入到输出中的元素
	--embed-css <int>             将CSS文件嵌入到输出中 (默认: 1)
	--embed-font <int>            将字体文件嵌入到输出中 (默认: 1)
	--embed-image <int>           将图片文件嵌入到输出中 (默认: 1)
	--embed-javascript <int>      将JavaScript文件嵌入到输出中 (默认: 1)
	--embed-outline <int>         将大纲嵌入到输出中 (默认: 1)
	--split-pages <int>           将页面分割为单独的文件 (默认: 0)
	--dest-dir <string>           指定输出目录 (默认: ".")
	--css-filename <string>       生成的CSS文件名 (默认: "")
	--page-filename <string>      分页文件命名模板 (默认: "")
	--outline-filename <string>   生成的大纲文件名 (默认: "")
	--process-nontext <int>       除文本外还渲染图形 (默认: 1)
	--process-outline <int>       在HTML中显示大纲 (默认: 1)
	--process-annotation <int>    在HTML中显示注释 (默认: 0)
	--process-form <int>          包含文本字段和单选按钮 (默认: 0)
	--printing <int>              启用打印支持 (默认: 1)
	--fallback <int>              以回退模式输出 (默认: 0)
	--tmp-file-size-limit <int>   临时文件使用的最大大小 (KB), -1表示无限制 (默认: -1)
	--embed-external-font <int>   嵌入本地匹配的外部字体 (默认: 1)
	--font-format <string>        嵌入字体的格式 (ttf,otf,woff,svg) (默认: "woff")
	--decompose-ligature <int>    分解连字，如 ﬁ -> fi (默认: 0)
	--turn-off-ligatures <int>    明确告诉浏览器不要使用连字 (默认: 0)
	--auto-hint <int>             对无提示的字体使用fontforge自动提示 (默认: 0)
	--external-hint-tool <string> 外部字体提示工具 (覆盖 --auto-hint) (默认: "")
	--stretch-narrow-glyph <int>  拉伸窄字形而不是填充它们 (默认: 0)
	--squeeze-wide-glyph <int>    缩小宽字形而不是截断它们 (默认: 1)
	--override-fstype <int>       清除TTF/OTF字体中的fstype位 (默认: 0)
	--process-type3 <int>         转换Type 3字体为Web格式 (实验性) (默认: 0)
	--heps <fp>                   合并文本的水平阈值，以像素为单位 (默认: 1)
	--veps <fp>                   合并文本的垂直阈值，以像素为单位 (默认: 1)
	--space-threshold <fp>        单词 break 阈值 (阈值 * em) (默认: 0.125)
	--font-size-multiplier <fp>   大于1的值会提高渲染精度 (默认: 4)
	--space-as-offset <int>       将空格字符视为偏移量 (默认: 0)
	--tounicode <int>             如何处理ToUnicode CMaps (0=自动, 1=强制, -1=忽略) (默认: 0)
	--optimize-text <int>         尝试减少用于文本的HTML元素数量 (默认: 0)
	--correct-text-visibility <int> 处理被遮挡的文本: 0: 不检查, 1: 处理完全被遮挡的文本, 2: 处理部分被遮挡的文本 (默认: 1)
	--covered-text-dpi <fp>       当启用 --correct-text-visibility=2 时，用于渲染被遮挡文本的DPI (默认: 300)
	--bg-format <string>          指定背景图片格式 (默认: "png")
	--svg-node-count-limit <int>  如果svg背景图片中的节点数超过此限制，则将此页面回退到位图背景；负值表示无限制 (默认: -1)
	--svg-embed-bitmap <int>      1: 在svg背景中嵌入位图；0: 尽可能将位图转储到外部文件 (默认: 1)
	-o,--owner-password <string>  所有者密码 (用于加密文件)
	-u,--user-password <string>   用户密码 (用于加密文件)
	--no-drm <int>                覆盖文档DRM设置 (默认: 0)
	--clean-tmp <int>             转换完成后删除临时文件 (默认: 1)
	--tmp-dir <string>            指定临时目录位置 (默认: "/tmp")
	--data-dir <string>           指定数据目录 (默认: "/usr/local/share/pdf2htmlEX")
	--poppler-data-dir <string>   指定poppler数据目录 (默认: "/usr/local/share/pdf2htmlEX/poppler")
	--debug <int>                 打印调试信息 (默认: 0)
	--proof <int>                 文本同时绘制在文本层和背景上以进行验证 (默认: 0)
	--quiet <int>                 安静执行操作 (默认: 0)
	-v,--version                  打印版权和版本信息
	-h,--help                     打印使用信息
*/
func main() {
	// 设置Gin模式为发布模式，提高性能
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()

	r.GET("/pdf-to-html", func(c *gin.Context) {
		// 记录请求开始
		startTime := time.Now()
		rawQuery := c.Request.URL.RawQuery
		rawPdfUrl := extractRawPdfUrl(rawQuery)

		if rawPdfUrl == "" {
			fmt.Printf("[%s] 错误: 缺少pdf_url参数\n", time.Now().Format(TimeFormat))
			c.JSON(http.StatusBadRequest, gin.H{"error": "缺少pdf_url参数"})
			return
		}

		// 获取信号量，限制并发请求数
		semaphore <- struct{}{}
		defer func() {
			<-semaphore
			fmt.Printf("[%s] 请求处理完成，耗时: %v\n", time.Now().Format(TimeFormat), time.Since(startTime))
		}()

		// 创建临时目录 - 在系统临时目录内，确保隔离性
		tempDir, err := os.MkdirTemp("", "pdf2html-")
		if err != nil {
			fmt.Printf("[%s] 创建临时目录失败: %v\n", time.Now().Format(TimeFormat), err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("创建临时目录失败: %v", err)})
			return
		}
		fmt.Printf("[%s] 创建临时目录成功: %s\n", time.Now().Format(TimeFormat), tempDir)
		defer os.RemoveAll(tempDir)

		// 下载PDF文件
		pdfPath := filepath.Join(tempDir, "input.pdf")
		fmt.Printf("[%s] 开始下载PDF文件: %s\n", time.Now().Format(TimeFormat), rawPdfUrl)
		if err := downloadFile(pdfPath, rawPdfUrl); err != nil {
			fmt.Printf("[%s] 下载PDF文件失败: %v\n", time.Now().Format(TimeFormat), err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("下载PDF文件失败: %v", err)})
			return
		}
		fmt.Printf("[%s] PDF文件下载成功: %s\n", time.Now().Format(TimeFormat), pdfPath)

		// 校验下载的文件是否为合法 PDF
		fmt.Printf("[%s] 验证PDF文件: %s\n", time.Now().Format(TimeFormat), pdfPath)
		if err := validatePDF(pdfPath); err != nil {
			fmt.Printf("[%s] 验证PDF文件失败: %v\n", time.Now().Format(TimeFormat), err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("下载的文件不是有效的PDF: %v", err)})
			return
		}
		fmt.Printf("[%s] PDF文件验证成功\n", time.Now().Format(TimeFormat))

		// 使用pdf2htmlEX转换PDF为HTML
		baseHtmlPath := filepath.Join(tempDir, "output.html")
		fmt.Printf("[%s] 开始PDF转HTML转换\n", time.Now().Format(TimeFormat))
		if err := convertPdfToHtml(pdfPath, baseHtmlPath); err != nil {
			fmt.Printf("[%s] PDF转HTML转换失败: %v\n", time.Now().Format(TimeFormat), err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("PDF转HTML转换失败: %v", err)})
			return
		}
		fmt.Printf("[%s] PDF转HTML转换成功\n", time.Now().Format(TimeFormat))

		// 直接返回转换后的HTML文件
		outputHtmlPath := filepath.Join(tempDir, "output.html")
		if _, err := os.Stat(outputHtmlPath); os.IsNotExist(err) {
			fmt.Printf("[%s] 错误: 未找到输出HTML文件: %s\n", time.Now().Format(TimeFormat), outputHtmlPath)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "未找到输出HTML文件"})
			return
		}
		fmt.Printf("[%s] 返回HTML文件: %s\n", time.Now().Format(TimeFormat), outputHtmlPath)

		// 设置响应头，指定文件类型为HTML
		c.Header("Content-Type", "text/html")
		// 返回文件
		c.File(outputHtmlPath)
	})

	r.Run(":8080")
}

// 下载文件
func downloadFile(filepath string, url string) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	// 👇 关键：伪装成 Chrome 浏览器
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/pdf,*/*;q=0.9")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	client := &http.Client{Timeout: 300 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// validatePDF 检查文件是否以 %PDF 开头
func validatePDF(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file for validation: %w", err)
	}
	defer f.Close()

	buf := make([]byte, 4)
	if _, err := io.ReadFull(f, buf); err != nil {
		return fmt.Errorf("failed to read file header: %w", err)
	}

	if !bytes.HasPrefix(buf, []byte("%PDF")) {
		// 读取前 200 字节用于错误诊断
		f.Seek(0, 0)
		fullBuf := make([]byte, 200)
		n, _ := io.ReadFull(f, fullBuf)
		return fmt.Errorf("file is not a valid PDF (starts with: %q)", string(fullBuf[:n]))
	}
	return nil
}

// extractRawPdfUrl 从原始查询字符串中提取pdf_url参数
func extractRawPdfUrl(rawQuery string) string {
	// 使用net/url包解析查询字符串，自动处理URL编码
	params, err := url.ParseQuery(rawQuery)
	if err != nil {
		fmt.Printf("[%s] 解析查询字符串失败: %v\n", time.Now().Format(TimeFormat), err)
		return ""
	}

	// 获取pdf_url参数值
	if pdfUrls, ok := params["pdf_url"]; ok && len(pdfUrls) > 0 {
		return pdfUrls[0]
	}
	return ""
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
