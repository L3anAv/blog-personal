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
	Fijado      bool   `yaml:"fijado"`
	Link        string
}

// Función para copiar archivos (assets, estilos, etc.)
func copyRoute(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, _ := filepath.Rel(src, path)
		targetPath := filepath.Join(dst, relPath)
		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}
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

// Renderiza una página asegurando que existan las subcarpetas necesarias
func renderPage(outputFile string, contentTemplate string, data any) {
	// 1. Configurar archivos base
	files := []string{"layout/index.html"}

	// 2. Cargar componentes dinámicamente
	components, err := filepath.Glob("components/*.html")
	if err != nil {
		fmt.Printf("Error buscando componentes: %v\n", err)
		return
	}
	files = append(files, components...)
	files = append(files, filepath.Join("pages", contentTemplate))

	// 3. Parsear templates
	tmpl, err := template.ParseFiles(files...)
	if err != nil {
		fmt.Printf("Error parseando templates para %s: %v\n", outputFile, err)
		return
	}

	// 4. Lógica de directorios: public/ + ruta solicitada
	fullOutputPath := filepath.Join("public", outputFile)
	outputDir := filepath.Dir(fullOutputPath)

	// Crea public/ y cualquier subcarpeta (como public/post) si no existen
	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		fmt.Printf("Error creando directorios: %v\n", err)
		return
	}

	// 5. Crear archivo y ejecutar template
	f, err := os.Create(fullOutputPath)
	if err != nil {
		fmt.Printf("Error creando archivo %s: %v\n", fullOutputPath, err)
		return
	}
	defer f.Close()

	// Se asume que el bloque principal en tus .html se llama "base"
	err = tmpl.ExecuteTemplate(f, "base", data)
	if err != nil {
		fmt.Printf("Error ejecutando %s: %v\n", outputFile, err)
	}
}

func main() {
	// 1. Cargar configuración
	configRaw, err := os.ReadFile("config.yaml")
	if err != nil {
		fmt.Println("Error: No se encontró config.yaml")
		return
	}
	var config map[string]string
	yaml.Unmarshal(configRaw, &config)
	baseUrl := config["base_url"]

	// 2. Limpieza y preparación inicial
	os.RemoveAll("public") // Opcional: limpia antes de generar
	os.MkdirAll("public", 0755)

	// Copiar archivos estáticos
	copyRoute("assets", "public/assets")
	copyRoute("style", "public/style")

	// 3. Procesar contenidos
	files, _ := os.ReadDir("content")
	var allPosts []Post

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".yaml" {
			content, _ := os.ReadFile(filepath.Join("content", file.Name()))
			var post Post
			yaml.Unmarshal(content, &post)

			// Determinar nombre del archivo .html
			var fileName string
			if post.Title != "" {
				fileName = slugify(post.Title) + ".html"
			} else {
				fileName = strings.TrimSuffix(file.Name(), ".yaml") + ".html"
			}

			// IMPORTANTE: Definimos la ruta relativa para el enlace y el archivo
			// Esto hará que renderPage lo guarde en public/post/
			post.Link = "post/" + fileName

			allPosts = append(allPosts, post)

			// Renderizado individual del post
			postData := map[string]any{
				"BaseURL": baseUrl,
				"Post":    post,
			}
			renderPage(post.Link, "post.html", postData)
		}
	}

	// 4. Renderizar páginas globales (Index y Lista)
	limite := 5
	if len(allPosts) < 5 {
		limite = len(allPosts)
	}

	data := map[string]any{
		"BaseURL": baseUrl,
		"Posts":   allPosts,
		"Latest":  allPosts[:limite],
	}

	// Estas se guardan en public/ (raíz)
	renderPage("index.html", "home.html", data)
	renderPage("lista-de-posteos.html", "lista-de-posteos.html", data)

	fmt.Println("🚀 Sitio generado con éxito en /public")
	fmt.Printf("📂 Posts en: public/post/\n")
	fmt.Printf("📄 Páginas en: public/\n")
}