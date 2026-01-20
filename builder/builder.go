package builder

import (
	"fmt"
	"bytes"
	"regexp"
	"strings"
	"html/template"
	"path/filepath"

	"github.com/tdewolff/minify/v2"
    "github.com/tdewolff/minify/v2/html"
)

type Builder struct {
    baseTmpl *template.Template
	pages map[string]*template.Template
}

type RenderResult struct {
    FolderName string
    Content    []byte
}

func slugify(s string) string {
	s = strings.ToLower(s)
	reg := regexp.MustCompile("[^a-z0-9]+")
	s = reg.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

// Init de templates
func (b *Builder) InitTemplates() ([]string, error) {
    b.pages = make(map[string]*template.Template)
    var pageNames []string // Aquí guardaremos los nombres

    // 1. Cargar base y componentes (como ya lo tienes)
    files := []string{"layout/index.html"}
    components, _ := filepath.Glob("components/*.html")
    files = append(files, components...)
    
    var err error
    b.baseTmpl, err = template.ParseFiles(files...)
    if err != nil {
        return nil, err
    }

    // 2. Escanear la carpeta pages/ y cargar el mapa
    pagesFiles, _ := filepath.Glob("pages/*.html")
    for _, path := range pagesFiles {
        name := filepath.Base(path) // "home.html", "contacto.html", etc.
        
        t, err := b.baseTmpl.Clone()
		if err != nil { return nil, err }

        t, err = t.ParseFiles(path)
        if err != nil {
            return nil, err
        }
        
        b.pages[name] = t
        pageNames = append(pageNames, name) // Agregamos el nombre a la lista
    }

    return pageNames, nil
}

func (b *Builder) BuildPage(contentTemplate string, data any) (RenderResult, error) {

    // 1. Clonamos la base (Layout + Componentes) que ya está en memoria
    tmpl, err := b.baseTmpl.Clone()
    if err != nil {
        return RenderResult{}, fmt.Errorf("error clonando base: %w", err)
    }

    // 2. Añadimos SOLO el archivo de la página específica (ej: pages/post.html)
    tmpl, err = tmpl.ParseFiles(filepath.Join("pages", contentTemplate))
    if err != nil {
        return RenderResult{}, fmt.Errorf("error añadiendo página %s: %w", contentTemplate, err)
    }

    // 3. Extraemos el nombre para la carpeta (ej: "post.html" -> "post")
    folderName := strings.TrimSuffix(contentTemplate, ".html")

    // 4. Renderizamos al buffer
    var buf bytes.Buffer
    err = tmpl.ExecuteTemplate(&buf, "base", data)
    if err != nil {
        return RenderResult{}, fmt.Errorf("error ejecutando template: %w", err)
    }

	// --- BLOQUE DE MINIFICACIÓN ---
	m := minify.New()
    m.AddFunc("text/html", html.Minify) // Configuramos el minificador de HTML
	
    minified, err := m.Bytes("text/html", buf.Bytes())
    if err != nil {
        // Si falla la minificación, devolvemos el HTML normal por seguridad
        return RenderResult{
            FolderName: folderName,
            Content:    buf.Bytes(),
        }, nil
    }

	// Retorno minificado
    return RenderResult{
        FolderName: folderName,
        Content:    minified,
    }, nil
}

func (b *Builder) BuildPosts(baseUrl string, allPosts []Post, active bool) {
	
	// 3.2 Recorrer y renderizar los posts
	for i := range allPosts {
		// Apuntamos al post original para actualizar su Link permanentemente
		post := &allPosts[i]

		// Nombre de la carpeta dentro de post
		RouteNamePost := slugify(post.Title)
		
		post.Link = "post/" + RouteNamePost + "/"

		// Preparamos los datos para el template
		postData := map[string]any{
			"BaseURL": baseUrl,
			"Post": post, // Pasamos el puntero o el valor (*post)
            "ActivePinned": active,
		}
		
		// Generamos el archivo físico (ej: public/post/mi-titulo.html)
		PostResult, err := b.BuildPage("post.html", postData)
		if err != nil {
			fmt.Printf("Error renderizando post: %v\n", err)
			continue // Salta al siguiente post si este falla
		}
		CreateRoute(RoutePost, RouteNamePost, PostResult)
	}
}