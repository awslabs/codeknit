---
title: Instalación
description: Cómo instalar codeknit en tu sistema.
---

codeknit puede instalarse desde el código fuente. Los siguientes pasos te guiarán a través de la configuración de codeknit en tu sistema.

## Desde el código fuente

El método de instalación principal es compilar desde el código fuente. Necesitarás:

- Go 1.26+
- Un compilador C (requerido para tree-sitter a través de CGo)

Clona el repositorio y compila el binario:

```bash
git clone https://github.com/awslabs/codeknit.git
cd codeknit
make build
```

El binario compilado estará disponible en `./bin/codeknit`.

## Añadir a PATH

Para ejecutar `codeknit` desde cualquier directorio, añade la ubicación del binario a la variable PATH de tu sistema.

Para **bash** (`~/.bashrc`):

```bash
export PATH="$PATH:/ruta/a/codeknit"
```

Para **zsh** (`~/.zshrc`):

```bash
export PATH="$PATH:/ruta/a/codeknit"
```

Para **fish** (`~/.config/fish/config.fish`):

```fish
fish_add_path /ruta/a/codeknit
```

Después de actualizar la configuración de tu shell, recárgala ejecutando `source ~/.bashrc` (o `~/.zshrc`) o reinicia tu terminal.

## Completado de shell

codeknit soporta el autocompletado para shells populares. Instala los completados usando estos comandos:

Para **bash**:

```bash
codeknit completion bash >> ~/.bashrc
```

Para **zsh**:

```bash
codeknit completion zsh >> ~/.zshrc
```

Para **fish**:

```bash
codeknit completion fish > ~/.config/fish/completions/codeknit.fish
```

Para **PowerShell**:

```powershell
codeknit completion powershell >> $PROFILE
```

## Verificar instalación

Después de la instalación, verifica que codeknit esté configurado correctamente:

```bash
codeknit --version
```

## Configuración de desarrollo

Si estás contribuyendo a codeknit, ejecuta estos comandos adicionales:

Instala las dependencias de desarrollo:

```bash
make deps
```

Configura los hooks de git:

```bash
make setup
```

Ejecuta el conjunto de pruebas:

```bash
make test
```
