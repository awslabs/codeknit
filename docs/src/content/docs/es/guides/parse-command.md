---
title: Comando Parse
description: Extraer información estructural del código fuente en archivos .skt o JSON.
---

El comando `codeknit parse` extrae información estructural de tu base de código —como funciones, clases, métodos, variables y sus relaciones— y la emite en formato compacto `.skt` de forma predeterminada. Usa JSON cuando necesites salida legible por máquina para scripts, integraciones o herramientas posteriores.

## Uso básico

```bash
codeknit parse <input-path> [output-dir]
```

- **`<input-path>`**: Ruta al directorio o archivo que deseas analizar.
- **`[output-dir]`**: Directorio de salida opcional. Si no se proporciona, el valor predeterminado es `./skeleton`.

### Ejemplos

```bash
# Parse a project, output to default directory ./skeleton
codeknit parse ./src

# Parse and write to a custom output directory
codeknit parse ./src ./output

# Parse a single file and output to stdout
codeknit parse ./src/main.go --output-mode inline

# Emit machine-readable JSON to stdout
codeknit parse ./src --output-mode inline --format json
```

## Modos de salida

Usa `--output-mode` para controlar cómo se estructura la salida. Tres modos están disponibles:

| Modo             | Descripción                                                                                      | Mejor para                                                         |
| ---------------- | ------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------ |
| `directory-flat` | Escribe archivos `.skt` divididos (ej. `map_001.skt`, `map_002.skt`) en el directorio de salida. | ✅ **La mayoría de proyectos** — modo predeterminado y recomendado |
| `directory-tree` | Refleja la estructura del directorio fuente, creando un archivo `.skt` por archivo fuente.       | Navegar la salida junto al código fuente                           |
| `inline`         | Vuelca toda la salida a stdout.                                                                  | Archivos individuales o canalización a otras herramientas          |

> **Consejo**: Usa `directory-flat` de forma predeterminada a menos que estés trabajando con un solo archivo. Evita `inline` para entradas grandes, ya que puede saturar las ventanas de contexto.

## Flags

| Flag             | Predeterminado   | Descripción                                                                             |
| ---------------- | ---------------- | --------------------------------------------------------------------------------------- |
| `--output-mode`  | `directory-flat` | Modo de salida: `inline`, `directory-flat` o `directory-tree`                           |
| `--format`       | `skt`            | Formato de salida: `skt` o `json`                                                       |
| `--max-lines`    | `500`            | Número máximo de líneas por archivo de salida en modos flat/tree                        |
| `--collect-test` | `false`          | Incluir archivos de prueba en el análisis                                               |
| `--minify`       | `false`          | Habilitar compresión basada en diccionario para reducir el uso de tokens                |
| `--edges`        | `false`          | Incluir la sección `[edges]` con datos de relación (llamadas, contiene, etc.)           |
| `--clean`        | `false`          | Eliminar archivos `.skt` existentes en el directorio de salida antes de escribir        |
| `--workers`      | `NumCPU`         | Número máximo de gorutinas de análisis concurrentes (0 = usar todos los núcleos de CPU) |
| `--verbose`      | `false`          | Mostrar información de progreso y tiempo durante el procesamiento                       |

## Patrones comunes

```bash
# First run on a project
codeknit parse ./src
```

```bash
# Re-run and clean previous output
codeknit parse ./src --clean
```

```bash
# Parse a single file to stdout
codeknit parse ./src/main.go --output-mode inline
```

```bash
# Minify output for large codebases
codeknit parse ./src --minify
```

```bash
# Include relationship edges (e.g., for dependency analysis)
codeknit parse ./src --edges
```

```bash
# Emit JSON for another tool
codeknit parse ./src --output-mode inline --format json --edges
```

Ejemplo de salida JSON:

```json
{
  "files": ["app.go"],
  "symbols": [
    {
      "id": "app.go::User",
      "short_id": "S1",
      "name": "User",
      "file": "app.go",
      "category": "type",
      "kind": "struct",
      "signature": "type User struct",
      "span": [3, 3]
    },
    {
      "id": "app.go::Save",
      "short_id": "S2",
      "name": "Save",
      "file": "app.go",
      "category": "callable",
      "kind": "function",
      "signature": "Save(u: S1)",
      "span": [5, 5]
    }
  ],
  "edges": [
    {
      "from": "app.go::Save",
      "from_short": "S2",
      "to": "app.go::User",
      "to_short": "S1",
      "kind": "references"
    }
  ]
}
```

```bash
# Mirror source tree structure in output
codeknit parse ./src --output-mode directory-tree
```

## Incertidumbre en las relaciones

Con `--edges`, Codeknit preserva las relaciones incluso cuando la sintaxis y el ámbito no establecen un objetivo único:

```skt
S1 --calls--> S2
S1 --references[unresolved]--> string, bool
S1 --calls[ambiguous]--> Helper [candidates=S3, S4]
```

La omisión de un estado significa que fue resuelto por el analizador. `unresolved` significa que la dependencia no pudo establecerse; `ambiguous` significa que quedan interpretaciones competidoras posibles. Los candidatos son posibles objetivos, no conexiones confirmadas. Los puntos finales conocidos usan ShortID; los objetivos no resueltos sin símbolos pueden aparecer como nombres.

JSON lleva la misma información en `resolution` y `candidates`, y omite `to_short` para objetivos inciertos. C/C++ incluye símbolos `meta/file` y `meta/include` para que las relaciones usen ShortID; esto registra directivas escritas sin afirmar la resolución de encabezados a nivel de compilador.

Las métricas de dependencia excluyen relaciones inciertas. El grafo HTML muestra su conteo sin dibujar enlaces individuales inciertos. Consulta la [referencia del formato de salida](../../reference/output-format/) para la gramática y la [referencia de soporte de lenguajes](../../reference/language-support/) para los límites de resolución.

## Protección contra salida obsoleta

Si el directorio de salida ya contiene archivos `.skt` de una ejecución anterior, `codeknit` se negará a escribir nueva salida para evitar mezclar datos obsoletos y frescos.

Para sobrescribir este comportamiento y limpiar el directorio de salida antes de escribir, usa el flag `--clean`:

```bash
codeknit parse ./src --clean
```

Esto garantiza un conjunto de salida fresco y consistente.

Al actualizar a `0.5.0`, regenera todo el conjunto de salida. Los IDs de símbolos y los resultados de relaciones pueden cambiar; no mezcles fragmentos de diferentes ejecuciones.

## Consejos

- ✅ **Usa `directory-flat` de forma predeterminada** para la mayoría de proyectos. Equilibra legibilidad y manejabilidad.
- 🔍 Usa `--minify` en bases de código grandes para reducir el uso de tokens mediante un diccionario compartido (`dict.skt`).
- 🔗 La sección `[edges]` está **excluida de forma predeterminada** para ahorrar tokens. Usa `--edges` cuando necesites datos de relación como `calls`, `contains` o `inherits`.
- 🧾 Usa `--format json` cuando un script o integración necesite datos estructurados en lugar de `.skt`.
- 🧹 Usa siempre `--clean` al volver a ejecutar en el mismo directorio de salida.
- 📁 Usa `directory-tree` si deseas correlacionar archivos `.skt` directamente con archivos fuente en tu editor.
