package main

import (
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

type Post struct {
	Title       string `yaml:"title"`
	Date        string `yaml:"date"`
	Author      string `yaml:"author"`
	Body        string `yaml:"body"`
	Description string `yaml:"description"`
	Link        string 
}

// Función para copiar archivos
func copyAssets(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil { return err }
		relPath, _ := filepath.Rel(src, path)
		targetPath := filepath.Join(dst, relPath)
		if info.IsDir() { return os.MkdirAll(targetPath, info.Mode()) }
		srcFile, _ := os.Open(path)
		defer srcFile.Close()
		dstFile, _ := os.Create(targetPath)
		defer dstFile.Close()
		_, err = io.Copy(dstFile, srcFile)
		return err
	})
}

func slugify(s string) string {
	s = strings.ToLower(s)
	reg := regexp.MustCompile("[^a-z0-9]+")
	s = reg.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

func main() {
	// 1. Leer el config.yaml y sacar el BaseURL
	configRaw, _ := os.ReadFile("config.yaml")
	var config map[string]string
	yaml.Unmarshal(configRaw, &config)
	baseUrl := config["base_url"]

	files, _ := os.ReadDir("content")
	os.MkdirAll("public", 0755)
	copyAssets("assets", "public/assets")

	var allPosts []Post

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".yaml" {
			content, _ := os.ReadFile(filepath.Join("content", file.Name()))
			var post Post
			yaml.Unmarshal(content, &post)

			if post.Title != "" {
				post.Link = slugify(post.Title) + ".html"
			} else {
				post.Link = strings.TrimSuffix(file.Name(), ".yaml") + ".html"
			}
			allPosts = append(allPosts, post)

			// 2. Renderizar Post pasando un mapa simple
			tmplPost := template.Must(template.ParseFiles("templates/post.html"))
			outFile, _ := os.Create(filepath.Join("public", post.Link))
			tmplPost.Execute(outFile, map[string]any{
				"BaseURL": baseUrl,
				"Post":    post,
			})
			outFile.Close()
		}
	}

	// 3. Renderizar Index pasando un mapa simple
	tmplIndex := template.Must(template.ParseFiles("templates/index.html"))
	indexFile, _ := os.Create(filepath.Join("public", "index.html"))
	tmplIndex.Execute(indexFile, map[string]any{
		"BaseURL": baseUrl,
		"Posts":   allPosts,
	})
	indexFile.Close()

	fmt.Println("🚀 Blog generado con BaseURL:", baseUrl)
}