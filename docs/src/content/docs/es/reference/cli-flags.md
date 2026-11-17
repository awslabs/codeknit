---
title: Referencia de la CLI
description: Referencia completa de todos los comandos y opciones de codeknit.
---

Esta página lista cada comando de codeknit, sus argumentos y sus opciones. Para instrucciones orientadas a tareas, consulta la [guía del comando parse](../../guides/parse-command/), la [guía de comandos graph](../../guides/graph-commands/) y la [guía del comando fingerprint](../../guides/fingerprint-command/).

## Resumen de comandos

| Comando                                       | Propósito                                                        |
| --------------------------------------------- | -------------------------------------------------------------- |
| `codeknit`                                    | Inicia la interfaz interactiva de terminal.                     |
| `codeknit parse <input-path> [output-dir]`    | Extrae símbolos y relaciones en salida `.skt` o JSON.           |
| `codeknit graph show <input-path>`            | Genera un grafo HTML interactivo.                               |
| `codeknit graph analyze <input-path>`         | Ejecuta algoritmos de grafo estructural y escribe un informe `.skt`. |
| `codeknit graph hotspots <input-path>`        | Clasifica puntos calientes de cambios usando historial Git y estructura de grafo. |
| `codeknit fingerprint <input-path>`           | Detecta código duplicado y casi duplicado.                      |
| `codeknit completion <shell>`                 | Genera un script de autocompletado para shell.                  |

Cinco comandos procesan código fuente: `parse`, `graph show`, `graph analyze`, `graph hotspots` y `fingerprint`. Comparten las opciones descritas en [Opciones compartidas de procesamiento de código fuente](#opciones-compartidas-de-procesamiento-de-código-fuente).

## Opciones globales

| Opción            | Tipo | Valor por defecto | Descripción                  |
| ----------------- | ---- | ----------------- | ---------------------------- |
| `--version`       | bool | `false`           | Muestra información de versión. |
| `-h`, `--help`    | bool | `false`           | Muestra ayuda para el comando. |

## Opciones compartidas de procesamiento de código fuente

Estas opciones aparecen en la tabla de cada comando de procesamiento de código fuente.

| Opción                    | Tipo                | Valor por defecto         | Valores permitidos       | Descripción                                                                          |
| ----------------------- | ------------------- | ------------------------- | ------------------------ | ------------------------------------------------------------------------------------ |
| `--header-language`     | string              | `c`                       | `c`, `cpp`               | Parser utilizado para archivos `.h`.                                                 |
| `--include GLOB`        | string, repetible   | Todos los fuentes soportados | Patrón estilo gitignore | Incluye rutas de código fuente que coincidan. Múltiples patrones forman una unión.  |
| `--exclude GLOB`        | string, repetible   | Ninguno                   | Patrón estilo gitignore  | Excluye rutas o directorios que coincidan.                                           |
| `--max-file-size BYTES` | int                 | `10485760` (10 MiB)       | `0` o mayor              | Omite archivos de código fuente más grandes que este tamaño. `0` desactiva el límite. |
| `--strict`              | bool                | `false`                   | —                        | Falla antes de escribir la salida si ocurren diagnósticos de escaneo o análisis.     |

### `--header-language`

Aplica solo a archivos `.h`. Otras extensiones mantienen su parser habitual (`.hpp` y `.hxx` ya usan C++).

- `c` (por defecto) preserva la extracción de macros, typedefs y uniones de C.
- `cpp` selecciona el parser de C++. Úsalo para proyectos de C++ cuyos encabezados tienen extensión `.h`.

La interfaz de usuario (TUI) usa la selección predeterminada `c` para archivos `.h`.

```bash
codeknit parse ./src --header-language cpp
```

Para conocer las capacidades de extracción probadas y los límites conocidos por lenguaje, consulta [Soporte de Lenguajes y Fidelidad de API](../language-support/). Esa página describe construcciones probadas, no cobertura semántica completa del lenguaje o del compilador.

### Selección de código fuente

- Los patrones usan globs estilo gitignore separados por barras, relativos al directorio de entrada. Para una entrada de archivo único, los patrones coinciden con el nombre base del archivo.
- Entrecomilla los patrones para que el shell no los expanda.
- Las exclusiones tienen prioridad sobre las inclusiones.
- Las inclusiones no anulan las reglas de `.gitignore`, el filtrado de archivos de prueba o las exclusiones integradas.
- No se aceptan patrones de negación (`!patrón`) en `--include` o `--exclude`.
- Los escaneos de directorios respetan los archivos `.gitignore` ancestrales y anidados independientemente del directorio de trabajo actual.
- Seleccionar explícitamente un archivo de código fuente regular omite el filtrado de ignorar y de pruebas. Las inclusiones, exclusiones y límites de tamaño aún se aplican.
- Los archivos de prueba se excluyen a menos que se establezca `--collect-test`.

```bash
codeknit parse ./src --include '**/*.go' --exclude '**/generated/' --strict
```

### Diagnósticos y `--strict`

- Los enlaces simbólicos, enlaces simbólicos de directorio y archivos no regulares se omiten y se reportan.
- Las rutas no legibles, las lecturas fallidas de archivos de ignorar, los archivos de código fuente demasiado grandes y las advertencias de sintaxis hacen que un análisis sea incompleto.
- Por defecto, se conservan los resultados parciales utilizables. Los diagnósticos aparecen en stderr. `parse` también los registra en `errors` de JSON o en `[errors]` de SKT.
- Con `--strict`, un análisis incompleto es rechazado antes de escribir la salida.
- Una ejecución sin fuentes soportadas seleccionadas, o en la que falla la lectura de cada fuente, siempre termina sin éxito.
- La interfaz de usuario (TUI) aplica la misma política de enlaces simbólicos y el límite predeterminado de 10 MiB. Sus formularios no exponen las opciones de selección y estrictas.

Consulta [Fiabilidad del Análisis](../analysis-reliability/) para garantías exactas.

## `codeknit`

```bash
codeknit
```

Ejecutar `codeknit` sin subcomando inicia la interfaz interactiva de terminal. La interfaz ofrece los comandos `parse`, `graph show`, `graph analyze`, `graph hotspots` y `fingerprint` junto con sus opciones.

## `codeknit parse`

```bash
codeknit parse <input-path> [output-dir]
```

| Argumento      | Requerido | Valor por defecto | Descripción                                                              |
| -------------- | --------- | ----------------- | ------------------------------------------------------------------------ |
| `input-path`   | Sí        | —                 | Directorio de código fuente o archivo de código fuente único.            |
| `output-dir`   | No        | `./skeleton`      | Directorio de salida. No se usa cuando `--output-mode` es `inline`.     |

| Opción                    | Tipo                | Valor por defecto         | Valores permitidos                                   | Descripción                                                                                         |
| ----------------------- | ------------------- | ------------------------- | ---------------------------------------------------- | --------------------------------------------------------------------------------------------------- |
| `--output-mode`         | string              | `directory-flat`          | `inline`, `directory-flat`, `directory-tree`         | Dónde y cómo se escribe la salida.                                                                  |
| `--format`              | string              | `skt`                     | `skt`, `json`                                        | Formato de salida.                                                                                  |
| `--max-lines`           | int                 | `500`                     | Para `directory-tree`, 3 o mayor                     | Objetivo de líneas para salida plana. Límite estricto de fragmentos de código fuente para salida en árbol. |
| `--collect-test`        | bool                | `false`                   | —                                                    | Incluir archivos de prueba.                                                                         |
| `--minify`              | bool                | `false`                   | —                                                    | Habilita la compresión de salida basada en diccionario.                                            |
| `--edges`               | bool                | `false`                   | —                                                    | Incluir la sección `[edges]`. Desactivado por defecto para ahorrar tokens.                          |
| `--clean`               | bool                | `false`                   | —                                                    | Reemplazar la salida generada previamente, solo después de una generación exitosa.                  |
| `--workers`             | int                 | `0` (`NumCPU`)            | `0` o mayor                                          | Máximo de goroutines de análisis concurrentes.                                                      |
| `--verbose`             | bool                | `false`                   | —                                                    | Mostrar información de progreso.                                                                   |
| `--header-language`     | string              | `c`                       | `c`, `cpp`                                           | Parser utilizado para archivos `.h`.                                                               |
| `--include GLOB`        | string, repetible   | Todos los fuentes soportados | —                                                  | Incluir rutas de código fuente que coincidan.                                                      |
| `--exclude GLOB`        | string, repetible   | Ninguno                   | —                                                    | Excluir rutas o directorios que coincidan.                                                         |
| `--max-file-size BYTES` | int                 | `10485760`                | `0` desactiva el límite                              | Omitir archivos de código fuente más grandes.                                                       |
| `--strict`              | bool                | `false`                   | —                                                    | Falla antes de escribir la salida si ocurren diagnósticos de escaneo o análisis.                   |

### Modos de salida

| Valor             | Comportamiento                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------ |
| `inline`          | Escribe en stdout. No se usa directorio de salida.                                                  |
| `directory-flat`  | Escribe archivos divididos en el directorio de salida. Los bloques de código fuente permanecen intactos, por lo que `--max-lines` es un objetivo de división en lugar de un límite estricto. |
| `directory-tree`  | Refleja el árbol de código fuente. Los nombres conservan la extensión del código fuente (`foo.c.skt`, `foo.h.skt`). Los archivos divididos usan `foo.c_part1.skt`. |

### Comportamiento de la salida

- Sin `--clean`, la salida generada existente para el formato seleccionado en el directorio de salida causa un error.
- Con `--clean`, los archivos `.skt` obsoletos (formato SKT) o `codeknit.json` (formato JSON) se reemplazan solo después de una generación exitosa. Las entradas no relacionadas se conservan.
- En formato JSON, los modos de directorio escriben `codeknit.json`. JSON es un único documento y no se divide por conteo de líneas.
- `dict.skt` y `warnings.skt` son nombres de archivos auxiliares reservados.
- No uses un directorio de código fuente o uno de sus ancestros como directorio de salida.

Consulta [Modos de salida](../../guides/output-modes/) y la [referencia del formato de salida](../output-format/) para más detalles.

```bash
codeknit parse ./src
codeknit parse ./src ./out --output-mode directory-tree
codeknit parse ./src --output-mode inline --format json --edges
```

## `codeknit graph show`

```bash
codeknit graph show <input-path>
```

| Argumento      | Requerido | Descripción                              |
| -------------- | --------- | ---------------------------------------- |
| `input-path`   | Sí        | Directorio de código fuente o archivo de código fuente único. |

| Opción                    | Tipo                | Valor por defecto                            | Valores permitidos          | Descripción                                                       |
| ----------------------- | ------------------- | ------------------------------------------ | --------------------------- | ----------------------------------------------------------------- |
| `-o`, `--output`        | string              | `./skeleton/codeknit-graph.html`           | Ruta de archivo             | Archivo HTML de salida.                                           |
| `--collect-test`        | bool                | `false`                                    | —                           | Incluir archivos de prueba.                                       |
| `--workers`             | int                 | `0` (`NumCPU`)                             | `0` o mayor                 | Máximo de goroutines de análisis concurrentes.                    |
| `--verbose`             | bool                | `false`                                    | —                           | Mostrar información de progreso.                                 |
| `--header-language`     | string              | `c`                                        | `c`, `cpp`                  | Parser utilizado para archivos `.h`.                              |
| `--include GLOB`        | string, repetible   | Todos los fuentes soportados               | —                           | Incluir rutas de código fuente que coincidan.                     |
| `--exclude GLOB`        | string, repetible   | Ninguno                                    | —                           | Excluir rutas o directorios que coincidan.                        |
| `--max-file-size BYTES` | int                 | `10485760`                                 | `0` desactiva el límite     | Omitir archivos de código fuente más grandes.                     |
| `--strict`              | bool                | `false`                                    | —                           | Falla antes de escribir la salida si ocurren diagnósticos de escaneo o análisis. |

El comando escribe un archivo HTML autónomo y lo abre en el navegador predeterminado. El visor dibuja las relaciones resueltas y muestra un conteo de relaciones inciertas.

```bash
codeknit graph show ./src -o graph.html
```

## `codeknit graph analyze`

```bash
codeknit graph analyze <input-path>
```

| Argumento      | Requerido | Descripción                              |
| -------------- | --------- | ---------------------------------------- |
| `input-path`   | Sí        | Directorio de código fuente o archivo de código fuente único. |

| Opción                       | Tipo                | Valor por defecto                            | Valores permitidos          | Descripción                                                  |
| -------------------------- | ------------------- | ------------------------------------------ | --------------------------- | ------------------------------------------------------------ |
| `-o`, `--output`           | string              | `./skeleton/graph_analysis.skt`            | Ruta de archivo             | Archivo `.skt` de salida.                                    |
| `--collect-test`           | bool                | `false`                                    | —                           | Incluir archivos de prueba.                                  |
| `--workers`                | int                 | `0` (`NumCPU`)                             | `0` o mayor                 | Máximo de goroutines de análisis concurrentes.               |
| `--verbose`                | bool                | `false`                                    | —                           | Mostrar información de progreso.                            |
| `--fan-threshold`          | int                 | `10`                                       | —                           | Mínimo de fan-in o fan-out para marcar un símbolo como hub.  |
| `--god-threshold`          | int                 | `15`                                       | —                           | Mínimo de conteo de relaciones `contains` para marcar una god class o función. |
| `--max-inheritance-depth`  | int                 | `5`                                        | —                           | Marcar cadenas de herencia más profundas que este valor.     |
| `--top-n`                  | int                 | `30`                                       | `0` = sin límite            | Limitar las secciones de salida clasificadas.                |
| `--betweenness-threshold`  | float               | `0.001`                                    | —                           | Valor mínimo de centralidad de intermediación para reportar. |
| `--propagation-cutoff`     | float               | `0.05`                                     | —                           | Probabilidad mínima para continuar la propagación de cambios. |
| `--header-language`        | string              | `c`                                        | `c`, `cpp`                  | Parser utilizado para archivos `.h`.                         |
| `--include GLOB`           | string, repetible   | Todos los fuentes soportados               | —                           | Incluir rutas de código fuente que coincidan.                |
| `--exclude GLOB`           | string, repetible   | Ninguno                                    | —                           | Excluir rutas o directorios que coincidan.                   |
| `--max-file-size BYTES`    | int                 | `10485760`                                 | `0` desactiva el límite     | Omitir archivos de código fuente más grandes.                |
| `--strict`                 | bool                | `false`                                    | —                           | Falla antes de escribir la salida si ocurren diagnósticos de escaneo o análisis. |

Las métricas de dependencia excluyen relaciones inciertas y metadatos de archivo/inclusión. El informe registra los conteos de relaciones excluidas.

```bash
codeknit graph analyze ./src -o analysis.skt --fan-threshold 15 --top-n 50
```

## `codeknit graph hotspots`

```bash
codeknit graph hotspots <input-path>
```

| Argumento      | Requerido | Descripción                                                  |
| -------------- | --------- | ------------------------------------------------------------ |
| `input-path`   | Sí        | Directorio de código fuente dentro de un repositorio Git.    |

| Opción                       | Tipo                | Valor por defecto                    | Valores permitidos                              | Descripción                                                  |
| -------------------------- | ------------------- | ------------------------------------ | ----------------------------------------------- | ------------------------------------------------------------ |
| `-o`, `--output`           | string              | `./skeleton/hotspots.skt`            | Ruta de archivo                                 | Archivo de salida.                                           |
| `--format`                 | string              | `skt`                                | `skt`, `json`                                   | Formato de salida.                                           |
| `--since`                  | string              | `12mo`                               | Ventana como `180d`, `12mo` o `2y`              | Ventana de historial.                                        |
| `--max-commits`            | int                 | `2000`                               | —                                               | Máximo de commits a inspeccionar.                            |
| `--max-files-per-commit`   | int                 | `50`                                 | —                                               | Excluir commits que toquen más archivos que este valor.      |
| `--min-cochanges`          | int                 | `3`                                  | —                                               | Mínimo de commits compartidos para reportar acoplamiento temporal. |
| `--top-n`                  | int                 | `30`                                 | —                                               | Máximo de resultados por sección.                            |
| `--include-merges`         | bool                | `false`                              | —                                               | Incluir commits de fusión en las métricas de historial.      |
| `--collect-test`           | bool                | `false`                              | —                                               | Incluir archivos de prueba.                                  |
| `--workers`                | int                 | `0` (`NumCPU`)                       | `0` o mayor                                     | Máximo de goroutines de análisis concurrentes.               |
| `--verbose`                | bool                | `false`                              | —                                               | Mostrar información de progreso.                           |
| `--header-language`        | string              | `c`                                  | `c`, `cpp`                                      | Parser utilizado para archivos `.h`.                         |
| `--include GLOB`           | string, repetible   | Todos los fuentes soportados         | —                                               | Incluir rutas de código fuente que coincidan.                |
| `--exclude GLOB`           | string, repetible   | Ninguno                              | —                                               | Excluir rutas o directorios que coincidan.                   |
| `--max-file-size BYTES`    | int                 | `10485760`                           | `0` desactiva el límite                         | Omitir archivos de código fuente más grandes.                |
| `--strict`                 | bool                | `false`                              | —                                               | Falla antes de escribir la salida si ocurren diagnósticos de escaneo o análisis. |

El comando combina frecuencia de commits, cambios de líneas y actualidad con PageRank a nivel de archivo, fan-in transitivo y centralidad de intermediación. También reporta acoplamiento temporal entre archivos que cambian juntos. Los commits de fusión se excluyen por defecto.

```bash
codeknit graph hotspots ./src --since 2y --format json -o hotspots.json
```

## `codeknit fingerprint`

```bash
codeknit fingerprint <input-path>
```

| Argumento      | Requerido | Descripción                              |
| -------------- | --------- | ---------------------------------------- |
| `input-path`   | Sí        | Directorio de código fuente o archivo de código fuente único. |

| Opción                    | Tipo                | Valor por defecto                          | Valores permitidos          | Descripción                                                                                                      |
| ----------------------- | ------------------- | ----------------------------------------- | --------------------------- | ---------------------------------------------------------------------------------------------------------------- |
| `-o`, `--output`        | string              | `./skeleton/fingerprints.skt`             | Ruta de archivo             | Archivo `.skt` de salida.                                                                                        |
| `--min-similarity`      | int                 | `65`                                      | `0`–`100`                   | Porcentaje mínimo de similitud para reportar.                                                                    |
| `--max-similarity`      | int                 | `95`                                      | `0`–`100`                   | Porcentaje máximo de similitud para reportar.                                                                    |
| `--show-all`            | bool                | `false`                                   | —                           | También emitir la sección `[fingerprints]` con datos de tokens sin procesar.                                    |
| `--rerank`              | bool                | `false`                                   | —                           | Encontrar vecinos semánticos y reclasificar candidatos con embeddings de Ollama. Requiere `ollama serve` y `ollama pull qwen3-embedding:0.6b`. |
| `--model`               | string              | `qwen3-embedding:0.6b`                    | Nombre de modelo de Ollama  | Modelo de embedding de Ollama a usar con `--rerank`.                                                             |
| `--collect-test`        | bool                | `false`                                   | —                           | Incluir archivos de prueba.                                                                                      |
| `--workers`             | int                 | `0` (`NumCPU`)                            | `0` o mayor                 | Máximo de goroutines de análisis concurrentes.                                                                   |
| `--verbose`             | bool                | `false`                                   | —                           | Mostrar información de progreso.                                                                                 |
| `--header-language`     | string              | `c`                                       | `c`, `cpp`                  | Parser utilizado para archivos `.h`.                                                                             |
| `--include GLOB`        | string, repetible   | Todos los fuentes soportados              | —                           | Incluir rutas de código fuente que coincidan.                                                                    |
| `--exclude GLOB`        | string, repetible   | Ninguno                                   | —                           | Excluir rutas o directorios que coincidan.                                                                       |
| `--max-file-size BYTES` | int                 | `10485760`                                | `0` desactiva el límite     | Omitir archivos de código fuente más grandes.                                                                    |
| `--strict`              | bool                | `false`                                   | —                           | Falla antes de escribir la salida si ocurren diagnósticos de escaneo o análisis.                                |

### Rango de similitud

El rango predeterminado, 65–95%, excluye coincidencias del 100%. Para reportar coincidencias exactas de tokens normalizados, establece ambos límites en `100`:

```bash
codeknit fingerprint ./src --min-similarity 100 --max-similarity 100
```

Los candidatos de tokens normalizados exactos son completos. Con 500 o más símbolos elegibles, otros candidatos de casi duplicados usan recuperación aproximada de los 50 vecinos principales. `--rerank` puede eliminar un candidato estructuralmente exacto. Consulta [Fiabilidad del Análisis](../analysis-reliability/#garantías-de-detección-de-duplicados) para más detalles.

## `codeknit completion`

```bash
codeknit completion <shell>
```

| Argumento  | Requerido | Valores permitidos                         | Descripción                                   |
| --------- | --------- | ------------------------------------------ | --------------------------------------------- |
| `shell`   | Sí        | `bash`, `zsh`, `fish`, `powershell`        | Shell para el que generar un script de autocompletado. |

El script se escribe en stdout.

```bash
codeknit completion zsh >> ~/.zshrc
```

## Páginas relacionadas

- [Referencia del formato de salida](../output-format/)
- [Soporte de Lenguajes y Fidelidad de API](../language-support/)
- [Fiabilidad del Análisis](../analysis-reliability/)