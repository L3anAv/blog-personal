<p align="center">
  <img src="./assets/logo.png" alt="Yamblg Logo" width="350">
</p>

# Yamblg
> **Fork. YAML. Go Publish.**

![Status](https://img.shields.io/badge/Status-Active-brightgreen)
![Made with Go](https://img.shields.io/badge/Made%20with-Go-00ADD8?logo=go)
![Configured with YAML](https://img.shields.io/badge/Config-YAML-CB171E?logo=yaml&logoColor=white)
![Version](https://img.shields.io/badge/Version-1.0.0-blue)

**Yamblg** (leído como *"I am blog"*) es un repositorio plantilla diseñado para quienes quieren un blog personal profesional en tiempo récord, sin configurar motores pesados. La idea es simple: haces un fork, modificas el YAML y ya tienes tu blog.

---

## 🚀 ¿Qué es Yamblg?
Es un **Starter Template** minimalista. No es un software que se instala, sino un punto de partida para tu identidad digital. El nombre juega con la fonética "I am blog" y la **"g"** final rinde tributo a **Go**, el lenguaje que procesa tu sitio con velocidad instantánea.

## 🛠️ Cómo se usa
Tener tu blog es tan sencillo como seguir estos tres pasos:

1. **Fork:** Haz un fork de este repositorio a tu cuenta de GitHub.
2. **Configura:** Edita el archivo `config.yaml` con tu información personal (nombre, bio, redes).
3. **Publica:** En los ajustes de tu repo, activa *GitHub Pages* apuntando a la rama `main` (o la carpeta `/docs`) y ¡listo! Tu blog estará online.

## ⚙️ Cómo funciona
Yamblg utiliza una arquitectura de sitio estático (SSG) de bajo consumo:
* **YAML como Fuente de Verdad:** No hay bases de datos. Todo el contenido y la configuración viven en archivos de texto plano fáciles de leer.
* **Go como Procesador:** El binario de Go toma tus archivos YAML y los transforma en HTML puro usando plantillas optimizadas.
* **Automatización Total:** Gracias a GitHub Actions, cada vez que editas un archivo desde la web o subes un cambio, el blog se reconstruye y se despliega solo.

---

Desarrollado para la simplicidad. 
**¡Go publish!**