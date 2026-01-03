package main

import (
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"gopkg.in/yaml.v3"
)

type Post struct {
	Title  string `yaml:"title"`
	Date   string `yaml:"date"`
	Author string `yaml:"author"`
	Body   string `yaml:"body"`
	Link   string // Campo extra para el índice
}

// Principal
func main() {
	files, _ := os.ReadDir("content")
	os.MkdirAll("public", 0755)

	var allPosts []Post

	// 1. Procesar cada archivo YAML
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".yaml" {
			content, _ := os.ReadFile(filepath.Join("content", file.Name()))
			var post Post
			yaml.Unmarshal(content, &post)

			// Definir el nombre del archivo de salida
			fileName := strings.TrimSuffix(file.Name(), ".yaml") + ".html"
			post.Link = fileName // Guardamos el nombre para el índice
			allPosts = append(allPosts, post)

			// Renderizar el post individual
			tmplPost := template.Must(template.ParseFiles("templates/post.html"))
			outFile, _ := os.Create(filepath.Join("public", fileName))
			tmplPost.Execute(outFile, post)
			outFile.Close()
		}
	}

	// 2. Generar el index.html (La lista de todos los posts)
	tmplIndex := template.Must(template.ParseFiles("templates/index.html"))
	indexFile, _ := os.Create(filepath.Join("public", "index.html"))
	tmplIndex.Execute(indexFile, allPosts) // Pasamos la lista de todos los posts
	indexFile.Close()
}