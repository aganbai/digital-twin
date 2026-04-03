package knowledge

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ledongthuc/pdf"
	"github.com/nguyenthenguyen/docx"
)

// FileParser 文件解析器
type FileParser struct{}

// NewFileParser 创建文件解析器
func NewFileParser() *FileParser {
	return &FileParser{}
}

// Parse 解析文件内容，返回纯文本
// 根据文件后缀名分发到不同的解析器
func (p *FileParser) Parse(filePath string) (string, error) {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("文件不存在: %s", filePath)
	}

	// 根据后缀名分发解析
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".txt", ".md":
		return p.parsePlainText(filePath)
	case ".pdf":
		return p.parsePDF(filePath)
	case ".docx":
		return p.parseDOCX(filePath)
	case ".pptx":
		return p.parsePPTX(filePath)
	case ".ppt":
		// PPT (老格式) 需要额外的库支持，暂不支持
		return "", fmt.Errorf("暂不支持 PPT 格式，请转换为 PPTX 格式后上传")
	default:
		return "", fmt.Errorf("不支持的文件格式: %s", ext)
	}
}

// SupportedFormats 返回支持的文件格式列表
func (p *FileParser) SupportedFormats() []string {
	return []string{".pdf", ".docx", ".txt", ".md", ".pptx"}
}

// parsePlainText 解析纯文本文件（TXT/MD）
func (p *FileParser) parsePlainText(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("读取文件失败: %w", err)
	}
	return string(data), nil
}

// parsePDF 解析 PDF 文件
// 使用 github.com/ledongthuc/pdf 库提取文本
func (p *FileParser) parsePDF(filePath string) (string, error) {
	f, r, err := pdf.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("打开 PDF 文件失败: %w", err)
	}
	defer f.Close()

	totalPages := r.NumPage()
	if totalPages == 0 {
		return "", nil
	}

	var sb strings.Builder
	for i := 1; i <= totalPages; i++ {
		page := r.Page(i)
		if page.V.IsNull() {
			// 空页面，跳过
			continue
		}

		text, err := page.GetPlainText(nil)
		if err != nil {
			// 单页解析错误不中断，继续处理其他页面
			continue
		}

		if text != "" {
			if sb.Len() > 0 {
				sb.WriteString("\n")
			}
			sb.WriteString(text)
		}
	}

	return sb.String(), nil
}

// parseDOCX 解析 DOCX 文件
// 使用 github.com/nguyenthenguyen/docx 库提取文本
func (p *FileParser) parseDOCX(filePath string) (string, error) {
	r, err := docx.ReadDocxFile(filePath)
	if err != nil {
		return "", fmt.Errorf("打开 DOCX 文件失败: %w", err)
	}
	defer r.Close()

	doc := r.Editable()
	content := doc.GetContent()

	// 提取纯文本：去除 XML 标签
	text := stripXMLTags(content)

	return text, nil
}

// stripXMLTags 去除 XML 标签，提取纯文本
func stripXMLTags(s string) string {
	var sb strings.Builder
	inTag := false
	for _, ch := range s {
		if ch == '<' {
			inTag = true
			continue
		}
		if ch == '>' {
			inTag = false
			continue
		}
		if !inTag {
			sb.WriteRune(ch)
		}
	}
	return strings.TrimSpace(sb.String())
}

// parsePPTX 解析 PPTX 文件
// PPTX 是一个 ZIP 压缩包，内部包含 XML 格式的幻灯片内容
func (p *FileParser) parsePPTX(filePath string) (string, error) {
	// 打开 ZIP 文件
	r, err := zip.OpenReader(filePath)
	if err != nil {
		return "", fmt.Errorf("打开 PPTX 文件失败: %w", err)
	}
	defer r.Close()

	var sb strings.Builder

	// 遍历 ZIP 内的文件，查找幻灯片 XML
	for _, f := range r.File {
		// 幻灯片文件路径格式: ppt/slides/slideN.xml
		if strings.HasPrefix(f.Name, "ppt/slides/slide") && strings.HasSuffix(f.Name, ".xml") {
			rc, err := f.Open()
			if err != nil {
				continue // 单个幻灯片解析失败不中断
			}

			data, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				continue
			}

			// 解析 XML 提取文本
			text := extractPPTXText(string(data))
			if text != "" {
				if sb.Len() > 0 {
					sb.WriteString("\n\n") // 幻灯片之间用空行分隔
				}
				sb.WriteString(text)
			}
		}
	}

	if sb.Len() == 0 {
		return "", fmt.Errorf("PPTX 文件内容为空")
	}

	return sb.String(), nil
}

// extractPPTXText 从 PPTX 幻灯片 XML 中提取文本
func extractPPTXText(xmlContent string) string {
	var sb strings.Builder

	// 使用简单的 XML 解码器提取文本
	decoder := xml.NewDecoder(strings.NewReader(xmlContent))
	for {
		token, err := decoder.Token()
		if err != nil {
			break
		}

		switch t := token.(type) {
		case xml.CharData:
			text := strings.TrimSpace(string(t))
			if text != "" {
				if sb.Len() > 0 {
					sb.WriteString(" ")
				}
				sb.WriteString(text)
			}
		}
	}

	return sb.String()
}
