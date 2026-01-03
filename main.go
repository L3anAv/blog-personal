package main

import (
	"fmt" // Agregado
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
	Link        string // Campo extra para el índice
}

func copyAssets(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, _ := filepath.Rel(src, path)
		targetPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		dstFile, err := os.Create(targetPath)
		if err != nil {
			return err
		}
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
	files, _ := os.ReadDir("content")

	os.MkdirAll("public", 0755)

	// Copiar los assets
	err := copyAssets("assets", "public/assets")
	if err != nil {
		fmt.Println("Error copiando assets:", err)
	} // <-- ESTA LLAVE FALTABA

	var allPosts []Post

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".yaml" {
			content, _ := os.ReadFile(filepath.Join("content", file.Name()))
			var post Post
			yaml.Unmarshal(content, &post)

			var nameBase string
			if post.Title != "" {
				nameBase = slugify(post.Title) + ".html"
			} else {
				nameBase = strings.TrimSuffix(file.Name(), ".yaml") + ".html"
			}

			post.Link = nameBase 
			allPosts = append(allPosts, post)

			tmplPost := template.Must(template.ParseFiles("templates/post.html"))
			outFile, _ := os.Create(filepath.Join("public", post.Link))
			tmplPost.Execute(outFile, post)
			outFile.Close()
		}
	}

	tmplIndex := template.Must(template.ParseFiles("templates/index.html"))
	indexFile, _ := os.Create(filepath.Join("public", "index.html"))
	tmplIndex.Execute(indexFile, allPosts)
	indexFile.Close()
    
    fmt.Println("🚀 Blog generado con éxito en /public")
}