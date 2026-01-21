package builder

import (
    "os"
	"fmt"
    "log"
    "time"
	"bytes"
	"regexp"
	"strings"
	"html/template"
	"path/filepath"

    // RSS
    "github.com/snabb/sitemap"
    "github.com/gorilla/feeds"

    // Minificacion
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

func GenerateSitemap(posts []Post, UrlUser string, BaseUrl string) {
    sm := sitemap.New()
    
    // Añadir la home
    sm.Add(&sitemap.URL{
        Loc:      UrlUser + BaseUrl,
        Priority: 1.0,
    })
    
    layout := "2006-01-02"
   
    // Añadir tus posts
    for _, p := range posts {
        
        t, _ := time.Parse(layout, p.Date)

        sm.Add(&sitemap.URL{
            Loc:        p.FullLink,
            LastMod:    &t, // Usa tu fecha de creación
            ChangeFreq: sitemap.Weekly,
        })
    }

    f, _ := os.Create("public/sitemap.xml")
    sm.WriteTo(f)
}

func GenerateRSS(posts []Post, UrlUser string, baseUrl string, descrp string, author string, email string) {

    feed := &feeds.Feed{
        Title:       descrp,
        Link:        &feeds.Link{Href: UrlUser + baseUrl + "/index.xml"},
        Description: descrp,
        Author: &feeds.Author{Name: author, Email: email},
        Created:     time.Now(),
    }

    layout := "2006-01-02"

    for _, p := range posts {
        // Parseamos el string a objeto time.Time
        t, err := time.Parse(layout, p.Date)
        if err != nil {
            // Si el YAML no tiene fecha o el formato falla, usamos la hora actual
            t = time.Now() 
        }

        item := &feeds.Item{
            Title:       p.Title,
            Link:        &feeds.Link{Href: p.FullLink},
            Id: p.FullLink,
            Description: p.Description,
            Author:      &feeds.Author{Name: p.Author, Email: p.Email},
            Created:     t, // <--- QUITA EL & AQUÍ. RSS usa valor, no puntero.
        }
        feed.Items = append(feed.Items, item)
    }

    f, err := os.Create("public/index.xml")
    if err != nil {
        log.Fatal(err)
    }
    defer f.Close()
    
    feed.WriteRss(f)
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

func (b *Builder) BuildPosts(baseUrl string, allPosts []Post, active bool, userUrl string, emailDir string) {
	
	// 3.2 Recorrer y renderizar los posts
	for i := range allPosts {
		// Apuntamos al post original para actualizar su Link permanentemente
		post := &allPosts[i]

		// Nombre de la carpeta dentro de post
		RouteNamePost := slugify(post.Title)
        
        //Definiendo rutas
        post.UrlUser = userUrl
		post.Link = "post/" + RouteNamePost + "/"
        post.FullLink = userUrl + baseUrl + "/" + post.Link

        //Email
        post.Email = emailDir

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