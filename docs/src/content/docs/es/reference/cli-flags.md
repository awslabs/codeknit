---
title: Referencia de CLI
description: Referencia completa para todos los comandos y opciones de codeknit.
---

## codeknit

Inicia la interfaz de usuario de terminal interactiva (TUI), que te guía a través de los comandos y opciones disponibles.

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
| `--max-lines`    | int    | `500`                | Máximo de líneas por archivo de salida (aplica a los modos `directory-flat` y `directory-tree`) |
| `--collect-test` | bool   | `false`              | Incluir archivos de prueba en el análisis                                                       |
| `--minify`       | bool   | `false`              | Habilitar la minificación de salida basada en diccionario                                       |
| `--edges`        | bool   | `false`              | Incluir la sección `[edges]` en la salida (desactivado por defecto para ahorrar tokens)         |
| `--clean`        | bool   | `false`              | Eliminar archivos `.skt` obsoletos del directorio de salida antes de escribir                   |
| `--workers`      | int    | `0` (NumCPU)         | Máximo de goroutines de análisis concurrentes                                                   |
| `--verbose`      | bool   | `false`              | Mostrar información de progreso durante el procesamiento                                        |

El directorio de salida predeterminado es `./skeleton` cuando no se especifica. En el modo `inline`, la salida se escribe en stdout y no se utiliza ningún directorio. Con `--format json`, la salida en directorio se escribe como `codeknit.json`.

## codeknit graph show

Genera una visualización interactiva de grafo en HTML de la estructura del código base.

```bash
codeknit graph show <input-path>
```

| Flag             | Tipo   | Valor predeterminado               | Descripción                                      |
| ---------------- | ------ | ---------------------------------- | ------------------------------------------------ |
| `-o`, `--output` | string | `./skeleton/codeknit-graph.html`   | Ruta del archivo HTML de salida                 |
| `--collect-test` | bool   | `false`                            | Incluir archivos de prueba en el análisis       |
| `--workers`      | int    | `0` (NumCPU)                       | Máximo de goroutines de análisis concurrentes   |
| `--verbose`      | bool   | `false`                            | Mostrar información de progreso durante el procesamiento |

El archivo HTML generado es autónomo y se abre automáticamente en tu navegador predeterminado.

## codeknit graph analyze

Ejecuta algoritmos de análisis estructural y emite un informe `.skt` legible por LLM.

```bash
codeknit graph analyze <input-path>
```

| Flag                      | Tipo    | Valor predeterminado            | Descripción                                                                 |
| ------------------------- | ------- | ------------------------------- | --------------------------------------------------------------------------- |
| `-o`, `--output`          | string  | `./skeleton/graph_analysis.skt` | Ruta del archivo `.skt` de salida                                          |
| `--collect-test`          | bool    | `false`                         | Incluir archivos de prueba en el análisis                                   |
| `--workers`               | int     | `0` (NumCPU)                    | Máximo de goroutines de análisis concurrentes                               |
| `--verbose`               | bool    | `false`                         | Mostrar información de progreso durante el procesamiento                   |
| `--fan-threshold`         | int     | `10`                            | Mínimo de fan-in o fan-out para marcar un símbolo como centro               |
| `--god-threshold`         | int     | `15`                            | Mínimo de relaciones de tipo "contiene" para marcar una god class/function  |
| `--max-inheritance-depth` | int     | `5`                             | Marcar cadenas de herencia más profundas que este valor                     |
| `--top-n`                 | int     | `30`                            | Limitar las secciones de salida clasificadas; `0` significa sin límite      |
| `--betweenness-threshold` | float64 | `0.001`                         | Valor mínimo de centralidad de intermediación para informar                |
| `--propagation-cutoff`    | float64 | `0.05`                          | Probabilidad mínima para continuar la simulación de propagación de cambios |

## codeknit graph hotspots

Clasifica archivos utilizando el historial de Git y la importancia estructural, e informa sobre el acoplamiento temporal entre archivos que cambian juntos repetidamente.

```bash
codeknit graph hotspots <input-path>
```

| Flag                     | Tipo   | Valor predeterminado       | Descripción                                                      |
| ------------------------ | ------ | -------------------------- | ---------------------------------------------------------------- |
| `-o`, `--output`         | string | `./skeleton/hotspots.skt`  | Ruta del archivo de salida                                       |
| `--format`               | string | `skt`                      | Formato de salida: `skt` o `json`                                |
| `--since`                | string | `12mo`                     | Ventana de historial, como `180d`, `12mo` o `2y`                 |
| `--max-commits`          | int    | `2000`                     | Máximo de commits a inspeccionar                                  |
| `--max-files-per-commit` | int    | `50`                       | Excluir commits que cambian más archivos                          |
| `--min-cochanges`        | int    | `3`                        | Mínimo de commits compartidos para el acoplamiento temporal      |
| `--top-n`                | int    | `30`                       | Máximo de resultados por sección del informe                      |
| `--include-merges`       | bool   | `false`                    | Incluir commits de fusión                                         |
| `--collect-test`         | bool   | `false`                    | Incluir archivos de prueba                                        |
| `--workers`              | int    | `0` (NumCPU)               | Máximo de goroutines de análisis concurrentes                     |
| `--verbose`              | bool   | `false`                    | Mostrar información de progreso                                   |

## codeknit fingerprint

Detecta código duplicado y casi duplicado utilizando fuzzy hashing.

```bash
codeknit fingerprint <input-path>
```

| Flag               | Tipo   | Valor predeterminado           | Descripción                                                                                                                  |
| ------------------ | ------ | ------------------------------ | ---------------------------------------------------------------------------------------------------------------------------- |
| `-o`, `--output`   | string | `./skeleton/fingerprints.skt`  | Ruta del archivo `.skt` de salida                                                                                            |
| `--min-similarity` | int    | `65`                           | Porcentaje mínimo de similitud para informar (0–100)                                                                         |
| `--max-similarity` | int    | `95`                           | Porcentaje máximo de similitud para informar (0–100)                                                                         |
| `--show-all`       | bool   | `false`                        | Incluir la sección `[fingerprints]` con datos de tokens sin procesar                                                         |
| `--rerank`         | bool   | `false`                        | Reclasificar candidatos CTPH utilizando embeddings semánticos vía Ollama (requiere `ollama serve` y `ollama pull qwen3-embedding:0.6b`) |
| `--model`          | string | `qwen3-embedding:0.6b`         | Modelo de embedding de Ollama a utilizar con `--rerank`                                                                     |
| `--collect-test`   | bool   | `false`                        | Incluir archivos de prueba en el análisis                                                                                    |
| `--workers`        | int    | `0` (NumCPU)                   | Máximo de goroutines de análisis concurrentes                                                                                |
| `--verbose`        | bool   | `false`                        | Mostrar información de progreso durante el procesamiento                                                                     |

## codeknit completion

Genera scripts de autocompletado para shells soportados.

```bash
codeknit completion <shell>
```

Shells soportados: `bash`, `zsh`, `fish`, `powershell`.

## Opciones globales

| Flag           | Descripción                       |
| -------------- | --------------------------------- |
| `--version`    | Muestra información de la versión |
| `--help`, `-h` | Muestra ayuda para el comando actual |