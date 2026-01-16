package builder

import (          
    "os"            
    "path/filepath"
)

type RouteType int

const (
    RoutePublic RouteType = iota
    RoutePost                   
)

func CreateRoute(routeType RouteType, slug string, result RenderResult) error {
    var baseDir string

    switch routeType {
    case RoutePost:
        // Une el folderName del template ("post") con el slug del post
        baseDir = filepath.Join("public", result.FolderName, slug)
    case RoutePublic:
        // Para páginas raíz, si es "home", lo mandamos directo a public/
        if result.FolderName == "home" || result.FolderName == "index" {
            baseDir = "public"
        } else {
            baseDir = filepath.Join("public", result.FolderName)
        }
    }

    finalPath := filepath.Join(baseDir, "index.html")
    os.MkdirAll(baseDir, 0755)
    return os.WriteFile(finalPath, result.Content, 0644)
}