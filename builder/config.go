package builder

import (
    "os"
	"fmt"
	"log"
    "strings"
    "path/filepath"
    "gopkg.in/yaml.v3"
	
	// Sistema de guardado
    "github.com/spf13/afero"

	"github.com/evanw/esbuild/pkg/api"
)

type Config struct {
    UserUrl   string `yaml:"userUrl"`
	BaseURL   string `yaml:"baseUrl"`
	SiteTitle string `yaml:"siteTitle"`
    Email     string `yaml:"email"`
	UseSectionPost struct {
		Active      bool   `yaml:"active"`
		LimitOfPost int    `yaml:"limitOfPost"`
	} `yaml:"useSectionPost"`
    UsePinned struct {
		Active      bool   `yaml:"active"`
	} `yaml:"usePinned"`
}

func LoadConfig() (Config, error){

	var config Config
    
    data, err := os.ReadFile("config.yaml")
    if err != nil {
        return config, err
    }
    
    err = yaml.Unmarshal(data, &config)
    
	return config, err
}

func ConfigYaml() error {
	contentDir := "./content"

	// Verificamos si la carpeta existe
	if _, err := os.Stat(contentDir); os.IsNotExist(err) {
		return fmt.Errorf("el directorio %s no existe", contentDir)
	}

	return filepath.Walk(contentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Filtrar solo archivos YAML
		if !info.IsDir() && (strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml")) {
			err := fillDateIfEmpty(path, info)
			if err != nil {
				fmt.Printf("Error procesando %s: %v\n", path, err)
			}
		}
		return nil
	})
}

// Función interna (no exportada) para la lógica de edición
func fillDateIfEmpty(path string, info os.FileInfo) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Usamos un mapa para mantener la estructura flexible
	var data map[string]interface{}
	if err := yaml.Unmarshal(content, &data); err != nil {
		return err
	}

	// Lógica de validación de fecha
	dateVal, exists := data["date"]
	if !exists || dateVal == "" || dateVal == nil {
		// info.ModTime() es agnóstico y representa la creación si el archivo es nuevo
		data["date"] = info.ModTime().Format("02-01-2006")

		// Serializar de nuevo a YAML
		newData, err := yaml.Marshal(&data)
		if err != nil {
			return err
		}

		// Sobrescribir el archivo con la fecha incluida
		return os.WriteFile(path, newData, 0644)
	}

	return nil
}

func MinifyCSS(fs afero.Fs) {
    result := api.Build(api.BuildOptions{
        EntryPoints: []string{"style/index.css"},
        Outdir:      "public/style",
        Bundle:      true,
        Write:       false,
		MinifyWhitespace:  true,
        MinifyIdentifiers: true,
        MinifySyntax:      true,
        Loader: map[string]api.Loader{
            ".css": api.LoaderCSS,
            ".ttf": api.LoaderCopy,
        },
        AssetNames: "[name]",
    })

    if len(result.Errors) > 0 {
        log.Fatalf("Error minificando CSS: %v", result.Errors)
    }

	if len(result.OutputFiles) == 0 {
        log.Println("⚠️ Ojo: esbuild no generó ningún archivo de salida.")
    }

   for _, file := range result.OutputFiles {
    // Intentamos limpiar la ruta absoluta
    // Si file.Path es "C:\Users\...\public\style\index.css"
    // Queremos que sea "public/style/index.css"
    
    // Obtenemos el directorio actual de trabajo
    cwd, _ := os.Getwd()
    
    // Intentamos obtener la ruta relativa
    relPath, err := filepath.Rel(cwd, file.Path)
    if err != nil {
        // Si falla, al menos intentamos quitar el prefijo manualmente
        // o tomamos el nombre del archivo.
        relPath = filepath.ToSlash(file.Path) 
    }

    // NORMALIZACIÓN CRÍTICA:
    // 1. Convertir \ a / (Windows a Web/FS)
    relPath = filepath.ToSlash(relPath)
    
    // 2. Si por algún motivo sigue teniendo "C:/", lo quitamos
    if len(relPath) > 2 && relPath[1] == ':' {
        // Esto quita "C:" del principio
        relPath = relPath[3:] 
    }
    
    // 3. Quitar posibles "/" iniciales para que afero no se confunda
    relPath = strings.TrimPrefix(relPath, "/")

    // Aseguramos que la carpeta exista en la memoria
    _ = fs.MkdirAll(filepath.Dir(relPath), 0755)

    // Escribimos en la memoria con la ruta LIMPIA
    err = afero.WriteFile(fs, relPath, file.Contents, 0644)
    
    if err != nil {
        log.Printf("❌ Error escribiendo: %v", err)
    }
}
}