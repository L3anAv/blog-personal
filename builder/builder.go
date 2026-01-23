package builder

import (
    "os"
    "io"
	"fmt"
    "log"
    "time"
	"bytes"
	"regexp"
    "strconv"
	"strings"
	"html/template"
	"path/filepath"

    // Sistema de guardado
    "github.com/spf13/afero"
    
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

// Función para copiar archivos (assets, estilos, etc.)
func copyRoute(fs afero.Fs,src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, _ := filepath.Rel(src, path)
		targetPath := filepath.Join(dst, relPath)
		if info.IsDir() {
			return fs.MkdirAll(targetPath, info.Mode())
		}
		srcFile, _ := os.Open(path)
		defer srcFile.Close()
		dstFile, _ := fs.Create(targetPath)
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

func RunBuild(fs afero.Fs, isDev bool) {
    cfg, err := LoadConfig()
    if err != nil {
        log.Fatal(err)
    }

    if isDev {
        cfg.BaseURL = "/"
    }

    err = ConfigYaml()
    if err != nil {
        log.Fatal(err)
    }

    b := &Builder{}
    paginasDetectadas, err := b.InitTemplates()
    if err != nil {
        log.Fatal(err)
    }

    allPosts, err := LoadPosts()
    if err != nil {
        log.Fatalf("Error cargando posts: %v", err)
    }
    
    limitePosts := min(len(allPosts), cfg.UseSectionPost.LimitOfPost)
    
    if !isDev {
    fs.RemoveAll("public")
    }
    
    fs.MkdirAll("public", 0755)
    fs.MkdirAll("public/style", 0755)
    
    MinifyCSS(fs)
    
    // Solo generar archivos de producción si no es MemMapFS
    if !isDev {
        origen, _ := os.Open("robots.txt") 
        if origen != nil {
            defer origen.Close()
            destino, _ := fs.Create("public/robots.txt")
            if destino != nil {
                defer destino.Close()
                io.Copy(destino, origen)
            }
        }
        GenerateSitemap(allPosts, cfg.UserUrl, cfg.BaseURL)
        GenerateRSS(allPosts, cfg.UserUrl, cfg.BaseURL, cfg.SiteTitle, allPosts[0].Author, cfg.Email)
    }

    copyRoute(fs, "assets", "public/assets")

    // Nota: Deberías pasar isDev a BuildPosts si quieres Live Reload en los artículos individuales
    b.BuildPosts(fs, cfg.BaseURL, allPosts, cfg.UsePinned.Active, cfg.UserUrl, cfg.Email)
    
    PagesData := map[string]any{
        "BaseURL":      cfg.BaseURL,
        "Title":         cfg.SiteTitle,
        "Posts":         allPosts,
        "ActiveLasted":  cfg.UseSectionPost.Active,
        "ActivePinned":  cfg.UsePinned.Active,
        "Latest":        allPosts[:limitePosts],
        "CantPost":      strconv.Itoa(limitePosts),
    }

    for _, nombreArchivo := range paginasDetectadas {
        if nombreArchivo == "post.html" {
            continue 
        }

        result, err := b.BuildPage(nombreArchivo, PagesData)
        if err != nil {
            log.Printf("Error en %s: %v", nombreArchivo, err)
            continue
        }

        // --- INYECCIÓN LIVE RELOAD ---
        if isDev {
            script := `
        <script>
        const ws = new WebSocket("ws://" + window.location.host + "/ws");
        ws.onmessage = (e) => { if (e.data === "reload") window.location.reload(); };
        </script>`

            contentStr := string(result.Content)
            
            if strings.Contains(strings.ToLower(contentStr), "</body>") {
                // Si existe, reemplazamos normal
                newContent := strings.Replace(contentStr, "</body>", script+"</body>", 1)
                result.Content = []byte(newContent)
            } else {
                // Si NO existe (por la minificación), lo pegamos al final
                result.Content = append(result.Content, []byte(script)...)
                fmt.Println("⚡ Etiqueta </body> no encontrada (posible minificación). Inyectando al final del archivo.")
            }
        }

        err = CreateRoute(fs, RoutePublic, "", result)
        if err != nil {
            log.Fatal(err)
        }
        
        fmt.Printf("✓ Página generada: %s\n", result.FolderName)
    }

    fmt.Println("🚀 Sitio generado con éxito")
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
        Link:        &feeds.Link{Href: UrlUser + baseUrl},
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

    // 1. Generamos el contenido XML en un string usando ToRss()
    // Esto es necesario para poder editar las etiquetas que faltan
    atomString, err := feed.ToAtom()
    if err != nil {
        log.Fatal("Error generando el string RSS:", err)
    }
    
    // 2. Definimos la URL de auto-referencia
    fullFeedURL := UrlUser + baseUrl + "/index.xml"
    
    // 3. Preparamos la etiqueta de auto-referencia obligatoria para Atom
    // Se debe colocar dentro del bloque principal <feed>
    atomSelfLink := fmt.Sprintf("\n  <link href=\"%s\" rel=\"self\" type=\"application/atom+xml\"></link>", fullFeedURL)

    // 4. Inyectamos la etiqueta justo después del subtítulo o el título
    // Buscamos el tag <subtitle> para pegar el link debajo
    if strings.Contains(atomString, "</subtitle>") {
        atomString = strings.Replace(atomString, "</subtitle>", "</subtitle>"+atomSelfLink, 1)
    } else {
        // Si no hay subtítulo, lo ponemos debajo del <title>
        atomString = strings.Replace(atomString, "</title>", "</title>"+atomSelfLink, 1)
    }
    
    // 5. Ahora guardamos el string final en el archivo físico
    err = os.WriteFile("public/index.xml", []byte(atomString), 0644)
    if err != nil {
        log.Fatal("Error escribiendo el archivo index.xml:", err)
    }
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
	
    HTMLminified, err := m.Bytes("text/html", buf.Bytes())
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
        Content:    HTMLminified,
    }, nil
}

func (b *Builder) BuildPosts(fs afero.Fs,baseUrl string, allPosts []Post, active bool, userUrl string, emailDir string) {
	
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
		CreateRoute(fs,RoutePost, RouteNamePost, PostResult)
	}
}