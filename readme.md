<p align="center">
  <img src="logo.png" alt="Yamblg Logo" width="max-content">
</p>

**Yamblg** (leído como *"I am blog"*) es un repositorio plantilla diseñado para quienes quieren un blog personal profesional en tiempo récord, sin configurar motores pesados. La idea es simple: haces un fork, modificas el YAML y ya tienes tu blog.

<br>

<div align="center">

![Status](https://img.shields.io/badge/Status-Active-brightgreen?style=flat-square)
![Go Version](https://img.shields.io/badge/go-v1.25+-00ADD8?style=flat-square&logo=go&logoColor=white)
![Config](https://img.shields.io/badge/Config-YAML-red?style=flat-square&logo=yaml&logoColor=white)
![Version](https://img.shields.io/badge/Version-0.0.30-blue?style=flat-square)

</div>

## 🚀 ¿Qué es Yamblg?
No es un software que se instala, sino un punto de partida para tu identidad digital. El nombre juega con la fonética "I am blog" y la **"g"** final rinde tributo a **Go**, el lenguaje que procesa tu nuevo blog de forma casi instantánea.


## 🛠️ Cómo se usa
Tener tu blog es tan sencillo como seguir estos tres pasos:

1. **Fork:** Haz un fork de este repositorio a tu cuenta de GitHub.
2. **Configura:** Crea archivos del estilo `name.yaml` con tu información para tus entradas, dentro del directorio Content.
3. **Publica:** En los ajustes de tu repo, activa *GitHub Pages* y elige la opciones de **Github Actions**.

<br>

## ⚙️ Cómo funciona
Yamblg utiliza una arquitectura de sitio estático (SSG) de bajo consumo:
* **YAML como Fuente de Verdad:** No hay bases de datos. Todo el contenido y la configuración viven en archivos de texto plano fáciles de leer.
* **Go como Procesador:** El binario de Go toma tus archivos YAML y los transforma en HTML puro usando plantillas optimizadas.
* **Automatización Total:** Gracias a GitHub Actions, cada vez que editas un archivo desde la web o subes un cambio, el blog se reconstruye y se despliega solo.

<br>

Desarrollado para la simplicidad. Desarollado para la comunidad. 💖 
<br>
**¡Go publish!**
