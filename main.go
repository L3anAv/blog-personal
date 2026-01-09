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
	Fijado      bool   `yaml:"fijado"` // Lee la propiedad del YAML
	Link        string
}

// Función para copiar archivos (assets)
func copyRoute(src, dst string) error {
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

// Función generalizada y reutilizable
func renderPage(outputFile string, contentTemplate string, data any) {
    // Definimos los archivos base que siempre se usan
    files := []string{
        "templates/layout.html",
        "templates/banner.html", // Agregamos el banner aquí
        filepath.Join("templates", contentTemplate),
    }

    // Parseamos todos los archivos juntos
    tmpl, err := template.ParseFiles(files...)
    if err != nil {
        fmt.Printf("Error parseando templates para %s: %v\n", outputFile, err)
        return
    }

    f, err := os.Create(filepath.Join("public", outputFile))
    if err != nil {
        fmt.Printf("Error creando archivo %s: %v\n", outputFile, err)
        return
    }
    defer f.Close()

    // Ejecutamos el bloque principal (asumiendo que layout.html define "base")
    err = tmpl.ExecuteTemplate(f, "base", data)
    if err != nil {
        fmt.Printf("Error ejecutando template %s: %v\n", outputFile, err)
    }
}

func main() {
	
	// 1. Configuración inicial
	configRaw, _ := os.ReadFile("config.yaml")
	var config map[string]string
	yaml.Unmarshal(configRaw, &config)
	baseUrl := config["base_url"]

	files, _ := os.ReadDir("content")
	os.MkdirAll("public", 0755)

	// Copiar Archivos
	copyRoute("assets", "public/assets")
	copyRoute("style", "public/style")

	var allPosts []Post

	// 2. Procesar archivos YAML
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".yaml" {
			content, _ := os.ReadFile(filepath.Join("content", file.Name()))
			var post Post
			yaml.Unmarshal(content, &post)

			// Generar link basado en título o nombre de archivo
			if post.Title != "" {
				post.Link = slugify(post.Title) + ".html"
			} else {
				post.Link = strings.TrimSuffix(file.Name(), ".yaml") + ".html"
			}

			// Guardamos todos los posts en un único array
			allPosts = append(allPosts, post)

			// Renderizar la página individual del Post
			tmplPost := template.Must(template.ParseFiles("templates/post.html"))
			outFile, _ := os.Create(filepath.Join("public", post.Link))
			tmplPost.Execute(outFile, map[string]any{
				"BaseURL": baseUrl,
				"Post":    post,
			})
			outFile.Close()
		}
	}

	limite := 5
	if len(allPosts) < 5 {
		limite = len(allPosts)
	}
	
	data := map[string]any{
			"BaseURL": baseUrl,
			"Posts":   allPosts,
			"Latest":  allPosts[:limite],
	}

	// Renderizamos ambas páginas usando el layout
	renderPage("index.html", "index.html", data)
	renderPage("lista-de-posteos.html", "lista-de-posteos.html", data)

	fmt.Println("🚀 Blog generado con éxito")
	fmt.Printf("Total de posts procesados: %d\n", len(allPosts))
}