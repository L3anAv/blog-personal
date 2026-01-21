package main

import (
	"os"
	"io"
	"fmt"
	"log"
	"strconv"
	"path/filepath"
	
	"yamblg/builder"
)

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

func main() {

// Obtener datos para crear .htmls
	cfg, err := builder.LoadConfig()
	if err != nil {
        log.Fatal(err)
    }

//Configuracion de yaml`s
	err = builder.ConfigYaml()
	if err != nil {
        log.Fatal(err)
    }

	// Instancia
	b := &builder.Builder{}
	
	// Inicializaciones
	paginasDetectadas, err := b.InitTemplates()
    if err != nil {
        log.Fatal(err)
    }

	allPosts, err := builder.LoadPosts()
	if err != nil {
		log.Fatalf("Error cargando posts: %v", err)
	}
	
	limitePosts := min(len(allPosts), cfg.UseSectionPost.LimitOfPost)
	
//  Limpieza y preparación
	os.RemoveAll("public")
	os.MkdirAll("public", 0755)
	os.MkdirAll("public/style", 0755)

	//Copiar robots.txt
	origen, _ := os.Open("robots.txt") 
	defer origen.Close()

	destino, _ := os.Create("public/robots.txt")
	defer destino.Close()

	io.Copy(destino, origen)

	// Antes de copiar
	copyRoute("assets", "public/assets")
	
	//Minificar CSS
	builder.MinifyCSS()
	
	//  Procesar Posts del blog
	b.BuildPosts(cfg.BaseURL, allPosts, cfg.UsePinned.Active, cfg.UserUrl, cfg.Email)
	
	// Crear rss y sitemap
	builder.GenerateSitemap(allPosts,cfg.UserUrl,cfg.BaseURL)
	builder.GenerateRSS(allPosts,cfg.UserUrl,cfg.BaseURL, cfg.SiteTitle, allPosts[0].Author, cfg.Email)

//  Renderizado de .html globales
	
	// 4.1 Pasando data para los tmpl
	PagesData := map[string]any{
		"BaseURL": cfg.BaseURL,
		"Title": cfg.SiteTitle,
		"Posts":   allPosts,
		"ActiveLasted": cfg.UseSectionPost.Active,
		"ActivePinned": cfg.UsePinned.Active,
		"Latest":  allPosts[:limitePosts],
		"CantPost": strconv.Itoa(limitePosts),
	}

	// 4.2 Creando los .html
	for _, nombreArchivo := range paginasDetectadas {
        
        // Caso especial: La plantilla de posts no se genera sola aquí
        // porque necesita datos (la lista de artículos).
        if nombreArchivo == "post.html" {
            continue 
        }

        // Renderizamos (usando el nombre como llave del mapa)
        result, err := b.BuildPage(nombreArchivo, PagesData)
        if err != nil {
            log.Printf("Error en %s: %v", nombreArchivo, err)
            continue
        }

        // Creamos la ruta en la raíz de public/
        err = builder.CreateRoute(builder.RoutePublic, "", result)
        if err != nil {
            log.Fatal(err)
        }
        
        fmt.Printf("✓ Página generada: %s\n", result.FolderName)
    }

// Logs de terminal para verificar
	fmt.Println("🚀 Sitio generado con éxito en /public")
	fmt.Printf("📂 Posts en: public/post/\n")
	fmt.Printf("📄 Páginas en: public/\n")
}