package builder

import (
    "os"

    "gopkg.in/yaml.v3"
)

type Config struct {
	BaseURL   string `yaml:"baseUrl"`
	SiteTitle string `yaml:"siteTitle"`
	UsePinned struct {
		Active      bool   `yaml:"active"`
		LimitOfPost int    `yaml:"limitOfPost"`
		Method      string `yaml:"method"`
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

	/*
	// viejo
	configRaw, err := os.ReadFile("config.yaml")
	if err != nil {
		fmt.Println("Error: No se encontró config.yaml")
		return
	}

	var config map[string]string
	yaml.Unmarshal(configRaw, &config)
	baseUrl := config["baseUrl"]
	*/
}