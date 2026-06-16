---
title: Comando Parse
description: Extraer información estructural del código fuente a archivos .skt.
---

El comando `codeknit parse` extrae información estructural de tu base de código —como funciones, clases, métodos, variables y sus relaciones— y la exporta en un formato compacto `.skt` diseñado para un consumo eficiente por parte de LLMs y herramientas de análisis.

## Uso básico

```bash
codeknit parse <input-path> [output-dir]
```

- **`<input-path>`**: Ruta al directorio o archivo que deseas analizar.
- **`[output-dir]`**: Directorio de salida opcional. Si no se proporciona, el valor predeterminado es `./skeleton`.

### Ejemplos

```bash
# Analizar un proyecto, salida en el directorio predeterminado ./skeleton
codeknit parse ./src

# Analizar y escribir en un directorio de salida personalizado
codeknit parse ./src ./output

# Analizar un solo archivo y mostrar la salida en stdout
codeknit parse ./src/main.go --output-mode inline
```

## Modos de salida

Usa `--output-mode` para controlar cómo se estructura la salida. Tres modos están disponibles:

| Modo             | Descripción                                                                                      | Mejor para                                                         |
| ---------------- | ------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------ |
| `directory-flat` | Escribe archivos `.skt` divididos (ej. `map_001.skt`, `map_002.skt`) en el directorio de salida. | ✅ **La mayoría de proyectos** — modo predeterminado y recomendado |
| `directory-tree` | Refleja la estructura del directorio fuente, creando un archivo `.skt` por cada archivo fuente.  | Navegar la salida junto al código fuente                           |
| `inline`         | Envía toda la salida a stdout.                                                                   | Archivos individuales o para redirigir a otras herramientas        |

> **Consejo**: Usa `directory-flat` como predeterminado a menos que trabajes con un solo archivo. Evita `inline` para entradas grandes, ya que puede saturar las ventanas de contexto.

## Flags

| Flag             | Valor predeterminado | Descripción                                                                              |
| ---------------- | -------------------- | ---------------------------------------------------------------------------------------- |
| `--output-mode`  | `directory-flat`     | Formato de salida: `inline`, `directory-flat` o `directory-tree`                         |
| `--max-lines`    | `500`                | Número máximo de líneas por archivo de salida en modos flat/tree                         |
| `--collect-test` | `false`              | Incluir archivos de prueba en el análisis                                                |
| `--minify`       | `false`              | Habilitar compresión basada en diccionario para reducir el uso de tokens                 |
| `--edges`        | `false`              | Incluir la sección `[edges]` con datos de relaciones (llamadas, contiene, etc.)          |
| `--clean`        | `false`              | Eliminar archivos `.skt` existentes en el directorio de salida antes de escribir         |
| `--workers`      | `NumCPU`             | Número máximo de goroutines de análisis concurrentes (0 = usar todos los núcleos de CPU) |
| `--verbose`      | `false`              | Mostrar información de progreso y tiempo durante el procesamiento                        |

## Patrones comunes

```bash
# Primera ejecución en un proyecto
codeknit parse ./src
```

```bash
# Volver a ejecutar y limpiar la salida anterior
codeknit parse ./src --clean
```

```bash
# Analizar un solo archivo a stdout
codeknit parse ./src/main.go --output-mode inline
```

```bash
# Minimizar la salida para bases de código grandes
codeknit parse ./src --minify
```

```bash
# Incluir relaciones (ej. para análisis de dependencias)
codeknit parse ./src --edges
```

```bash
# Reflejar la estructura del árbol fuente en la salida
codeknit parse ./src --output-mode directory-tree
```

## Protección contra salida obsoleta

Si el directorio de salida ya contiene archivos `.skt` de una ejecución anterior, `codeknit` se negará a escribir nueva salida para evitar mezclar datos obsoletos con los nuevos.

Para sobrescribir este comportamiento y limpiar el directorio de salida antes de escribir, usa el flag `--clean`:

```bash
codeknit parse ./src --clean
```

Esto garantiza un conjunto de salida fresco y consistente.

## Consejos

- ✅ **Usa `directory-flat` como predeterminado** para la mayoría de proyectos. Ofrece un equilibrio entre legibilidad y facilidad de gestión.
- 🔍 Usa `--minify` en bases de código grandes para reducir el uso de tokens mediante un diccionario compartido (`dict.skt`).
- 🔗 La sección `[edges]` está **excluida por defecto** para ahorrar tokens. Usa `--edges` cuando necesites datos de relaciones como `calls`, `contains` o `inherits`.
- 🧹 Usa siempre `--clean` al volver a ejecutar en el mismo directorio de salida.
- 📁 Usa `directory-tree` si deseas correlacionar archivos `.skt` directamente con los archivos fuente en tu editor.
