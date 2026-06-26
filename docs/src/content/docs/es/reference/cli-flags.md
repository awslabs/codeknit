---
title: Referencia de la CLI
description: Referencia completa de todos los comandos y opciones de codeknit.
---

## codeknit

Inicia la interfaz de usuario interactiva de terminal (TUI), que te guía a través de los comandos y opciones disponibles.

```bash
codeknit
```

## codeknit parse

Extrae información estructural del código fuente a archivos `.skt` o JSON.

```bash
codeknit parse <input-path> [output-dir]
```

| Flag             | Tipo   | Valor predeterminado | Descripción                                                                                     |
| ---------------- | ------ | -------------------- | ----------------------------------------------------------------------------------------------- |
| `--output-mode`  | string | `directory-flat`     | Modo de salida: `inline`, `directory-flat` o `directory-tree`                                   |
| `--format`       | string | `skt`                | Formato de salida: `skt` o `json`                                                              |
| `--max-lines`    | int    | `500`                | Número máximo de líneas por archivo de salida (aplica a los modos `directory-flat` y `directory-tree`) |
| `--collect-test` | bool   | `false`              | Incluir archivos de prueba en el análisis                                                       |
| `--minify`       | bool   | `false`              | Habilitar la minificación de salida basada en diccionario                                       |
| `--edges`        | bool   | `false`              | Incluir la sección `[edges]` en la salida (desactivado por defecto para ahorrar tokens)         |
| `--clean`        | bool   | `false`              | Eliminar archivos `.skt` obsoletos del directorio de salida antes de escribir                   |
| `--workers`      | int    | `0` (NumCPU)         | Número máximo de goroutines de análisis concurrentes                                            |
| `--verbose`      | bool   | `false`              | Mostrar información de progreso durante el procesamiento                                        |

El directorio de salida predeterminado es `./skeleton` cuando no se especifica. En el modo `inline`, la salida se escribe en stdout y no se utiliza ningún directorio. Con `--format json`, la salida del directorio se escribe como `codeknit.json`.

## codeknit graph show

Genera una visualización interactiva en HTML de la estructura del código base.

```bash
codeknit graph show <input-path>
```

| Flag             | Tipo   | Valor predeterminado            | Descripción                                  |
| ---------------- | ------ | ------------------------------- | -------------------------------------------- |
| `-o`, `--output` | string | `./skeleton/codeknit-graph.html` | Ruta del archivo HTML de salida             |
| `--collect-test` | bool   | `false`                         | Incluir archivos de prueba en el análisis   |
| `--workers`      | int    | `0` (NumCPU)                    | Número máximo de goroutines de análisis concurrentes |
| `--verbose`      | bool   | `false`                         | Mostrar información de progreso durante el procesamiento |

El archivo HTML generado es autónomo y se abre automáticamente en tu navegador predeterminado.

## codeknit graph analyze

Ejecuta algoritmos de análisis estructural y emite un informe `.skt` legible por LLM.

```bash
codeknit graph analyze <input-path>
```

| Flag                      | Tipo    | Valor predeterminado           | Descripción                                                   |
| ------------------------- | ------- | ------------------------------ | ------------------------------------------------------------- |
| `-o`, `--output`          | string  | `./skeleton/graph_analysis.skt` | Ruta del archivo `.skt` de salida                             |
| `--collect-test`          | bool    | `false`                        | Incluir archivos de prueba en el análisis                     |
| `--workers`               | int     | `0` (NumCPU)                   | Número máximo de goroutines de análisis concurrentes          |
| `--verbose`               | bool    | `false`                        | Mostrar información de progreso durante el procesamiento      |
| `--fan-threshold`         | int     | `10`                           | Número mínimo de fan-in o fan-out para marcar un símbolo hub  |
| `--god-threshold`         | int     | `15`                           | Número mínimo de relaciones contains para marcar una god class/function |
| `--max-inheritance-depth` | int     | `5`                            | Marcar cadenas de herencia más profundas que este valor       |
| `--top-n`                 | int     | `30`                           | Limitar las secciones de salida clasificadas; `0` significa sin límite |
| `--betweenness-threshold` | float64 | `0.001`                        | Valor mínimo de centralidad de intermediación para informar   |
| `--propagation-cutoff`    | float64 | `0.05`                         | Probabilidad mínima para continuar la simulación de propagación de cambios |

## codeknit fingerprint

Detecta código duplicado y casi duplicado utilizando fuzzy hashing.

```bash
codeknit fingerprint <input-path>
```

| Flag               | Tipo   | Valor predeterminado         | Descripción                                                                                                                  |
| ------------------ | ------ | ---------------------------- | ---------------------------------------------------------------------------------------------------------------------------- |
| `-o`, `--output`   | string | `./skeleton/fingerprints.skt` | Ruta del archivo `.skt` de salida                                                                                           |
| `--min-similarity` | int    | `65`                         | Porcentaje mínimo de similitud para informar (0–100)                                                                         |
| `--max-similarity` | int    | `95`                         | Porcentaje máximo de similitud para informar (0–100)                                                                         |
| `--show-all`       | bool   | `false`                      | Incluir la sección `[fingerprints]` con datos de tokens sin procesar                                                         |
| `--rerank`         | bool   | `false`                      | Reclasificar candidatos CTPH utilizando embeddings semánticos mediante Ollama (requiere `ollama serve` y `ollama pull qwen3-embedding:0.6b`) |
| `--model`          | string | `qwen3-embedding:0.6b`       | Modelo de embedding de Ollama a utilizar con `--rerank`                                                                     |
| `--collect-test`   | bool   | `false`                      | Incluir archivos de prueba en el análisis                                                                                    |
| `--workers`        | int    | `0` (NumCPU)                 | Número máximo de goroutines de análisis concurrentes                                                                         |
| `--verbose`        | bool   | `false`                      | Mostrar información de progreso durante el procesamiento                                                                     |

## codeknit completion

Genera scripts de autocompletado para shells compatibles.

```bash
codeknit completion <shell>
```

Shells compatibles: `bash`, `zsh`, `fish`, `powershell`.

## Opciones globales

| Flag           | Descripción                       |
| -------------- | --------------------------------- |
| `--version`    | Mostrar información de la versión |
| `--help`, `-h` | Mostrar ayuda para el comando actual |