---
title: Instalación
description: Cómo instalar codeknit en tu sistema.
---

Esta guía instala codeknit desde el código fuente, lo añade a tu `PATH`, habilita las completaciones de shell y verifica que funcione.

:::tip[¿Usas un asistente de codificación con IA?]
Los usuarios de Claude Code, Codex y Kiro pueden instalar el plugin de codeknit en su lugar. El plugin descarga el binario la primera vez que se ejecuta. Consulta [Usar codeknit con asistentes de IA](../guides/ai-assistants/).
:::

## Instalar desde el código fuente

**Requisitos previos**

- Go 1.26 o posterior
- Un compilador C. codeknit usa CGo para compilar sus analizadores tree-sitter.

**Pasos**

```bash
git clone https://github.com/awslabs/codeknit.git
cd codeknit
make build
```

El binario se guarda en `./bin/codeknit`.

### Usar un binario precompilado en su lugar

Los binarios precompilados para Linux, macOS y Windows (amd64 y arm64) están disponibles en la [página de versiones](https://github.com/awslabs/codeknit/releases). Descarga el archivo para tu plataforma, extráelo y mueve el binario a tu `PATH`. Cada versión incluye un archivo `checksums.txt` con el que puedes verificar la descarga.

## Añadir codeknit a tu PATH

Omite esta sección si ya moviste el binario a un directorio que esté en tu `PATH`, como `/usr/local/bin`.

Reemplaza `/path/to/codeknit` con el directorio que contiene el binario. Si compilaste desde el código fuente, este es el directorio `bin` dentro del repositorio clonado.

Añade la línea correspondiente a tu shell en su archivo de configuración:

```bash
# bash (~/.bashrc)
export PATH="$PATH:/path/to/codeknit"

# zsh (~/.zshrc)
export PATH="$PATH:/path/to/codeknit"

# fish (~/.config/fish/config.fish)
fish_add_path /path/to/codeknit
```

Recarga tu shell o carga el archivo que editaste:

```bash
source ~/.bashrc   # bash
source ~/.zshrc    # zsh
```

Para fish, abre una nueva terminal.

## Habilitar completaciones de shell

codeknit genera scripts de completación para bash, zsh, fish y PowerShell. Ejecuta el comando correspondiente a tu shell una vez:

```bash
# bash
codeknit completion bash >> ~/.bashrc

# zsh
codeknit completion zsh >> ~/.zshrc

# fish
codeknit completion fish > ~/.config/fish/completions/codeknit.fish

# PowerShell
codeknit completion powershell >> $PROFILE
```

Los comandos para bash, zsh y PowerShell añaden el script a tu perfil. Ejecutarlos de nuevo lo añadirá una segunda vez, así que ejecuta cada uno solo una vez.

Abre un nuevo shell o carga tu perfil para activar las completaciones.

## Verificar la instalación

```bash
codeknit --version
```

Si el comando muestra una versión, codeknit está instalado. Si tu shell reporta `command not found`, verifica que el directorio en tu línea `PATH` contenga el binario `codeknit`, luego recarga tu shell.

## Configurar un entorno de desarrollo

Haz esto solo si planeas contribuir a codeknit. Ejecuta estos comandos desde la raíz del repositorio clonado:

```bash
make deps    # instala dependencias
make setup   # configura git hooks
make test    # ejecuta la suite de pruebas
```

`make lint` ejecuta el linter y el formateador.

## Próximos pasos

- Ejecuta tu primer comando en [Introducción](../getting-started/).
- Explora todas las opciones en la [referencia de flags de la CLI](../reference/cli-flags/).