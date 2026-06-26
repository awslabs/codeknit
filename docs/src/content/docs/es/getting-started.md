---
title: Primeros pasos
description: Comienza a usar codeknit en menos de 5 minutos.
---

# Primeros pasos

Comienza a usar codeknit en menos de 5 minutos.

## 1. Requisitos previos

Necesitarás:

- Go 1.26+
- Un compilador C (CGo es requerido para tree-sitter)

## 2. Instalación desde el código fuente

```bash
git clone https://github.com/awslabs/codeknit.git
cd codeknit
make build
# El binario se encuentra en ./bin/codeknit
```

## 3. Agregar al PATH

Agrega el binario a la variable PATH de tu shell:

```bash
# bash (~/.bashrc)
export PATH="$PATH:$(pwd)/bin"

# zsh (~/.zshrc)
export PATH="$PATH:$(pwd)/bin"

# fish (~/.config/fish/config.fish)
fish_add_path $(pwd)/bin
```

Recarga tu shell o ejecuta `source ~/.bashrc` (o `~/.zshrc`) para que el cambio surta efecto.

## 4. Verificar la instalación

Verifica que codeknit esté funcionando:

```bash
codeknit --version
```

## 5. Primer análisis

Ejecuta tu primer análisis en un código base:

```bash
codeknit parse ./myproject
```

Este comando:

- Analiza todos los archivos fuente en `./myproject`
- Extrae información estructural (funciones, clases, relaciones)
- Escribe archivos `.skt` divididos en `./skeleton/` (directorio de salida predeterminado)

Si vuelves a ejecutar este comando, usa `--clean` para eliminar la salida anterior:

```bash
codeknit parse ./myproject --clean
```

## 6. Leyendo la salida

Los archivos `.skt` contienen información estructurada del código. Aquí tienes un pequeño ejemplo:

```skt
[symbols]
## src/main.go
S1 module/package L1-L1 main {}
S2 type/struct L5-L8 Server {exported}
S3 callable/function L10-L12 NewServer(addr: string) -> *S2 {exported}
S4 callable/method L14-L19 Start() {receiver=*Server}

[edges]
S2 --contains--> S4
S3 --returns--> S2
```

Secciones clave:

- `[symbols]`: Definiciones agrupadas por archivo, mostrando nombre, rango de líneas y metadatos
- `[edges]`: Relaciones como `contains`, `calls`, `inherits` o `returns`

## 7. Próximos pasos

Ahora que has ejecutado tu primer análisis:

- Aprende más sobre el comando parse: [Guía del comando parse](/codeknit/es/guides/parse-command/)
- Explora el análisis de estructura: [Guía de comandos de grafo](/codeknit/es/guides/graph-commands/)
- Comprende la detección de duplicados: [Guía del comando fingerprint](/codeknit/es/guides/fingerprint-command/)
- Lee el formato de salida completo: [Referencia del formato de salida](/codeknit/es/reference/output-format/)
- Consulta todas las banderas disponibles: [Referencia de banderas CLI](/codeknit/es/reference/cli-flags/)