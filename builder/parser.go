package builder

import (
    "fmt"
    "os"
	"strings"
    "path/filepath"

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

func LoadPosts() ([]Post, error) {

	directoryPath := "content"

    files, err := os.ReadDir(directoryPath)
    if err != nil {
        return nil, err
    }

    var posts []Post

    for _, file := range files {
        if filepath.Ext(file.Name()) == ".yaml" || filepath.Ext(file.Name()) == ".yml" {
            path := filepath.Join(directoryPath, file.Name())
            
            content, err := os.ReadFile(path)
            if err != nil {
                continue 
            }

            var post Post
            if err := yaml.Unmarshal(content, &post); err != nil {
                return nil, fmt.Errorf("error parseando %s: %v", file.Name(), err)
            }
            
			if post.Title == "" {
                post.Title = strings.TrimSuffix(file.Name(), ".yaml")
            }

            posts = append(posts, post)
        }
    }

    return posts, nil
}