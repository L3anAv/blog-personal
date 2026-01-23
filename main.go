package main

import (
	"fmt"
	"log"
	"net/http"
	"yamblg/builder"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var (
	upgrader  = websocket.Upgrader{ CheckOrigin: func(r *http.Request) bool { return true } }
	clientes  = make(map[*websocket.Conn]bool)
	notificar = make(chan bool)
)

func main() {
	var rootCmd = &cobra.Command{Use: "yamblg"}

	var buildCmd = &cobra.Command{
		Use:   "build",
		Short: "Producción",
		Run: func(cmd *cobra.Command, args []string) {
			fs := afero.NewOsFs()
			builder.RunBuild(fs, false)
		},
	}

	var serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "Desarrollo con Live Reload",
		Run: func(cmd *cobra.Command, args []string) {
			memFs := afero.NewMemMapFs()
			sourceFs := afero.NewOsFs()

			builder.RunBuild(memFs, true)

			// Canal de comunicación para el reload
			go func() {
				for {
					<-notificar
					for c := range clientes {
						c.WriteMessage(websocket.TextMessage, []byte("reload"))
					}
				}
			}()

			go iniciarWatcher(sourceFs, memFs)
			iniciarServidor(memFs)
		},
	}

	rootCmd.AddCommand(buildCmd, serveCmd)
	rootCmd.Execute()
}

func iniciarWatcher(sourceFs afero.Fs, memFs afero.Fs) {
	watcher, _ := fsnotify.NewWatcher()
	defer watcher.Close()

	dirs := []string{"content", "pages", "layout", "styles"}
	for _, d := range dirs { _ = watcher.Add(d) }

	for {
		select {
		case event := <-watcher.Events:
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove) != 0 {
				log.Printf("♻️  Cambio en %s. Actualizando...", event.Name)
				builder.RunBuild(memFs, true)
				notificar <- true
			}
		}
	}
}

func iniciarServidor(memFs afero.Fs) {
	// 1. Creamos un sub-sistema de archivos que apunte directamente a "public"
	// Así, para el servidor, la raíz "/" será la carpeta "public" en RAM.
	publicDir := afero.NewBasePathFs(memFs, "public")
	httpFs := afero.NewHttpFs(publicDir)
	fileserver := http.FileServer(httpFs.Dir("/"))

	// Endpoint del WebSocket (igual que antes)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, _ := upgrader.Upgrade(w, r, nil)
		clientes[conn] = true
		defer func() { delete(clientes, conn); conn.Close() }()
		for { if _, _, err := conn.ReadMessage(); err != nil { break } }
	})

	// Manejador principal
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Desactivar caché para que el Live Reload sea instantáneo
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		
		// Servir el archivo
		fileserver.ServeHTTP(w, r)
	})

	fmt.Println("🌍 Yamblg Dev Server: http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}