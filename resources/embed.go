package resources

import (
	"bytes"
	"embed"
	"fmt"
	"text/template"
)

//go:embed templates/config/*.yaml
//go:embed templates/dockerfile/*.Dockerfile
//go:embed templates/dockercompose/*.yaml
//go:embed templates/makefile/*.tmpl
//go:embed templates/scripts/*.sh
var Templates embed.FS

// GetTemplate 获取模板内容
func GetTemplate(path string) ([]byte, error) {
	return Templates.ReadFile("templates/" + path)
}

// MustGetTemplate 获取模板内容，失败时 panic
func MustGetTemplate(path string) []byte {
	data, err := GetTemplate(path)
	if err != nil {
		panic(fmt.Sprintf("failed to load template %s: %v", path, err))
	}
	return data
}

// ParseTemplate 解析模板并返回 *template.Template
func ParseTemplate(path string) (*template.Template, error) {
	data, err := GetTemplate(path)
	if err != nil {
		return nil, err
	}
	return template.New(path).Parse(string(data))
}

// ExecuteTemplate 解析并执行模板，返回渲染后的字符串
func ExecuteTemplate(path string, data interface{}) (string, error) {
	tmpl, err := ParseTemplate(path)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
