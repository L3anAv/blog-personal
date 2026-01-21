package builder

import (
    "os"
	"fmt"
    "strings"
    "path/filepath"
    "gopkg.in/yaml.v3"
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
		Method      string `yaml:"method"`
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

func MinifyCSS() {
    result := api.Build(api.BuildOptions{
        EntryPoints:       []string{"style/index.css"}, // Tu archivo principal
        Outfile:           "public/style/index.css",
        Bundle:            true,
        MinifyWhitespace:  true,
        MinifyIdentifiers: true,
        MinifySyntax:      true,
		Write:            true,
        Loader: map[string]api.Loader{
            ".css": api.LoaderCSS,
			".ttf": api.LoaderFile,
        },
    })

    if len(result.Errors) > 0 {
        fmt.Printf("Error minificando CSS: %v\n", result.Errors)
    }else {
        fmt.Println("✓ CSS minificado con éxito")
    }
}