package builder

import (
    "os"
	"fmt"
    "gopkg.in/yaml.v3"
	"github.com/evanw/esbuild/pkg/api"
)

type Config struct {
	BaseURL   string `yaml:"baseUrl"`
	SiteTitle string `yaml:"siteTitle"`
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