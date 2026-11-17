---
title: Fiabilidad del análisis
description: Selección de fuentes, análisis incompletos, publicación de resultados y garantías de detección de duplicados.
---

<!-- Maintained in docs/reference/analysis-reliability.md; copied verbatim by docs/gen. -->

Estos comportamientos describen los cambios no publicados en Go/CLI posteriores a la versión 0.5.0.

## Selección de fuentes y diagnósticos

Los cinco comandos de procesamiento de fuentes (`parse`, `fingerprint`, `graph show`, `graph analyze` y `graph hotspots`) aceptan:

| Flag                    | Default               | Meaning                                                             |
| ----------------------- | --------------------- | ------------------------------------------------------------------- |
| `--include GLOB`        | Todas las fuentes admitidas | Incluir rutas de fuentes que coincidan; repetir para seleccionar varios patrones. |
| `--exclude GLOB`        | Ninguno               | Excluir rutas o directorios que coincidan; repetir para varios patrones. |
| `--max-file-size BYTES` | `10485760` (10 MiB)   | Omitir archivos de fuente más grandes; `0` desactiva este límite.   |
| `--strict`              | `false`               | Fallar antes de emitir resultados si ocurren diagnósticos de escaneo o análisis. |

Los patrones utilizan globs estilo gitignore separados por barras, relativos al directorio de entrada seleccionado. Para una entrada de archivo único, los patrones coinciden con su nombre base. Entrecomille los patrones para que el shell no los expanda. Las inclusiones forman una unión; las exclusiones tienen prioridad. Las inclusiones no anulan `.gitignore`, el filtrado de pruebas o las exclusiones integradas. Los patrones de negación (`!patrón`) no son aceptados en estas flags; utilice los controles de inclusión y exclusión por separado.

```bash
codeknit parse ./src --include '**/*.go' --exclude '**/generated/' --strict
codeknit fingerprint ./src --max-file-size 2097152 --include '**/*.ts'
codeknit parse ./src/app.go --output-mode inline --format json
```

Los escaneos de directorios respetan los archivos `.gitignore` ancestrales y anidados relativos a su ubicación en el proyecto, independientemente del directorio de trabajo del proceso. El directorio `.git` más cercano o el marcador de worktree establece el límite del repositorio. Fuera de un repositorio, los archivos ignore ancestrales aún se aplican. Un subdirectorio seleccionado bajo un directorio ignorado permanece ignorado. La selección explícita de un archivo de fuente regular omite ignore y el filtrado de pruebas, pero aún aplica los límites de inclusión/exclusión y tamaño.

Se omiten y reportan los enlaces simbólicos de fuentes, enlaces simbólicos de directorios y fuentes no regulares. Las lecturas se limitan a la raíz del escaneo, y el límite de tamaño se aplica nuevamente durante la lectura, de modo que un archivo que crece después de su descubrimiento no pueda eludirlo. Las exclusiones integradas de directorios incluyen `.git`, `.hg`, `.svn`, `node_modules`, cachés de Python y entornos virtuales comunes de Python. Los archivos de prueba permanecen excluidos a menos que se establezca `--collect-test`.

Los directorios/archivos no legibles, las lecturas fallidas de archivos ignore, los enlaces simbólicos omitidos, las fuentes demasiado grandes y las advertencias de sintaxis hacen que un análisis sea incompleto. De forma predeterminada, se conservan los resultados parciales utilizables y los diagnósticos aparecen en stderr. El JSON de análisis incluye registros de ruta/razón en `errors`; SKT incluye `[errors]` en línea o en `warnings.skt`. Los comandos de grafo/huella digital también reportan diagnósticos en stderr. Los archivos ignorados o filtrados deliberadamente no son advertencias. `--strict` rechaza resultados incompletos antes de la publicación de salida. Una entrada sin fuentes admitidas seleccionadas, o una ejecución en la que falla la lectura de cada fuente, siempre finaliza sin éxito.

La TUI utiliza la misma política de enlaces simbólicos y el límite predeterminado de 10 MiB. Sus formularios aún no exponen las nuevas flags de selección y strict.

## Nombres de salida y publicación

La salida en árbol añade `.skt` al nombre completo de la fuente: `src/foo.c` se convierte en `src/foo.c.skt`, mientras que `src/foo.h` se convierte en `src/foo.h.skt`. La salida dividida utiliza `src/foo.c_part1.skt`, `src/foo.c_part2.skt`, y así sucesivamente. Los archivos de diccionario y diagnósticos conservan los nombres reservados `dict.skt` y `warnings.skt`. Las rutas de salida en conflicto, incluidas las colisiones solo de mayúsculas/minúsculas, fallan de manera visible.

El árbol SKT requiere `--max-lines` de al menos 3. Cada fragmento de símbolo repite sus encabezados `[symbols]` y `## source-path`. Los fragmentos solo de relaciones llevan `[edges]`. Cada fragmento de fuente respeta el límite; los diccionarios y los archivos de diagnóstico son salidas auxiliares separadas. El SKT plano mantiene los bloques de fuente intactos, por lo que su límite de líneas es un objetivo de división en lugar de un límite estricto para un bloque de fuente grande. JSON es un único documento y no se divide por conteo de líneas.

La generación de directorios se completa en un directorio de staging hermano. Sin `--clean`, la salida generada existente para el formato seleccionado causa un error. Con `--clean`, los archivos `.skt` obsoletos (modo SKT) o `codeknit.json` (modo JSON) se reemplazan solo después de una generación exitosa. Se conservan las entradas no relacionadas; una colisión con una causa un error. No utilice un directorio de fuente o su ancestro como directorio de salida.

La publicación mantiene un bloqueo específico del directorio, hace una copia de seguridad de la salida anterior y la restaura si el reemplazo falla. Si un proceso termina entre los cambios de nombre de directorio, la siguiente invocación restaura la copia de seguridad antes de generar nada. El archivo `.NAME.codeknit.lock` hermano permanece intencionalmente para coordinar futuros publicadores; no elimine un bloqueo activo o una copia de seguridad de recuperación. Hay un breve intervalo entre los dos cambios de nombre de directorio, por lo que los lectores que requieren acceso ininterrumpido deben coordinarse con la generación. Esto proporciona recuperación ante fallos de proceso, no una garantía contra fallos de almacenamiento o cortes de energía.

Los informes individuales se escriben y sincronizan en un hermano temporal antes de reemplazar el destino. Los archivos existentes no se truncan durante la generación. La salida estándar en línea no se puede revertir si su consumidor falla.

## Garantías de detección de duplicados

La coincidencia exacta significa flujos de tokens normalizados completos idénticos en la misma categoría de símbolo entre declaraciones elegibles para fingerprinting. No significa texto fuente idéntico o equivalencia semántica probada. Los grupos de candidatos exactos utilizan los flujos de tokens en sí, en lugar de la igualdad de hash difuso.

Por debajo de 500 símbolos elegibles, la detección estructural compara todos los pares de la misma categoría. A partir de 500 o más, todos los pares de tokens exactos se conservan independientemente de la recuperación de candidatos, y otros pares utilizan aproximadamente los 50 principales vecinos estructurales por símbolo. Por lo tanto, la recuperación de casi duplicados no es exhaustiva en entradas grandes. La recuperación aún escanea todos los vectores; no hace que el algoritmo general sea subcuadrático.

El filtrado por rango de similitud aún se aplica. El rango predeterminado es 65–95%, que excluye coincidencias del 100%. Para incluir coincidencias exactas, utilice:

```bash
codeknit fingerprint ./src --min-similarity 100 --max-similarity 100
```

La opción `--rerank` opcional añade candidatos semánticos, aplica acuerdo coseno y filtra según la puntuación combinada; puede eliminar un candidato estructuralmente exacto. El texto de embedding proviene de la misma instantánea de fuente limitada utilizada para la extracción. Las entradas estables y las puntuaciones fijas producen un orden determinista, incluidas las igualdades; las respuestas del modelo en sí son entradas externas.

`internal/emitter/calibration_test.go` evalúa un pequeño conjunto de fuentes etiquetadas con un umbral del 70%, reportando verdaderos/falsos positivos y negativos, precisión y recall por separado de las puntuaciones de similitud. El conjunto de prueba tiene un par exacto, dos pares de casi clones y tres pares no relacionados, ejercitados por debajo y a través del límite de 500 símbolos. Esta es una verificación de regresión, no una estimación representativa de la precisión en el mundo real. Los presupuestos de calibración y rendimiento multilingües más grandes siguen siendo trabajo en la hoja de ruta.