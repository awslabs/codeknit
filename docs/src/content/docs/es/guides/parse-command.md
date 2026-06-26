---
title: Comando Parse
description: Extraer información estructural del código fuente en archivos .skt o JSON.
---

El comando `codeknit parse` extrae información estructural de tu base de código —como funciones, clases, métodos, variables y sus relaciones— y la emite en formato compacto `.skt` de forma predeterminada. Usa JSON cuando necesites salida legible por máquina para scripts, integraciones o herramientas downstream.

## Uso básico

```bash
codeknit parse <input-path> [output-dir]
```

- **`<input-path>`**: Ruta al directorio o archivo que deseas analizar.
- **`[output-dir]`**: Directorio de salida opcional. Si no se proporciona, el valor predeterminado es `./skeleton`.

### Ejemplos

```bash
# Analizar un proyecto, salida al directorio predeterminado ./skeleton
codeknit parse ./src

# Analizar y escribir en un directorio de salida personalizado
codeknit parse ./src ./output

# Analizar un solo archivo y emitir a stdout
codeknit parse ./src/main.go --output-mode inline

# Emitir JSON legible por máquina a stdout
codeknit parse ./src --output-mode inline --format json
```

## Modos de salida

Usa `--output-mode` para controlar cómo se estructura la salida. Tres modos están disponibles:

| Modo               | Descripción                                                                              | Mejor para                                            |
| ------------------ | ---------------------------------------------------------------------------------------- | ----------------------------------------------------- |
| `directory-flat`   | Escribe archivos `.skt` divididos (ej. `map_001.skt`, `map_002.skt`) en el directorio de salida. | ✅ **La mayoría de proyectos** — modo predeterminado y recomendado |
| `directory-tree`   | Refleja la estructura del directorio fuente, creando un archivo `.skt` por cada archivo fuente. | Navegar por la salida junto al código fuente             |
| `inline`           | Envía toda la salida a stdout.                                                              | Archivos individuales o para canalizar a otras herramientas               |

> **Consejo**: Usa `directory-flat` de forma predeterminada a menos que estés trabajando con un solo archivo. Evita `inline` para entradas grandes, ya que puede saturar las ventanas de contexto.

## Flags

| Flag               | Valor predeterminado | Descripción                                                                  |
| ------------------ | -------------------- | ---------------------------------------------------------------------------- |
| `--output-mode`    | `directory-flat`     | **Modo de salida**: `inline`, `directory-flat` o `directory-tree`                 |
| `--format`         | `skt`                | Formato de salida: `skt` o `json`                                               |
| `--max-lines`      | `500`                | Número máximo de líneas por archivo de salida en modos flat/tree                             |
| `--collect-test`   | `false`              | Incluir archivos de prueba en el análisis                                               |
| `--minify`         | `false`              | Habilitar compresión basada en diccionario para reducir el uso de tokens                    |
| `--edges`          | `false`              | Incluir la sección `[edges]` con datos de relaciones (llamadas, contiene, etc.) |
| `--clean`          | `false`              | Eliminar archivos `.skt` existentes en el directorio de salida antes de escribir          |
| `--workers`        | `NumCPU`             | Número máximo de goroutines de análisis concurrentes (0 = usar todos los núcleos de CPU)      |
| `--verbose`        | `false`              | Mostrar información de progreso y tiempo durante el procesamiento                      |

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
# Minificar la salida para bases de código grandes
codeknit parse ./src --minify
```

```bash
# Incluir relaciones (ej. para análisis de dependencias)
codeknit parse ./src --edges
```

```bash
# Emitir JSON para otra herramienta
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
# Reflejar la estructura del árbol fuente en la salida
codeknit parse ./src --output-mode directory-tree
```

## Protección contra salida obsoleta

Si el directorio de salida ya contiene archivos `.skt` de una ejecución anterior, `codeknit` se negará a escribir nueva salida para evitar mezclar datos obsoletos y frescos.

Para sobrescribir este comportamiento y limpiar el directorio de salida antes de escribir, usa el flag `--clean`:

```bash
codeknit parse ./src --clean
```

Esto garantiza un conjunto de salida fresco y consistente.

## Consejos

- ✅ **Usa `directory-flat` de forma predeterminada** para la mayoría de proyectos. Equilibra legibilidad y facilidad de gestión.
- 🔍 Usa `--minify` en bases de código grandes para reducir el uso de tokens mediante un diccionario compartido (`dict.skt`).
- 🔗 La sección `[edges]` está **excluida de forma predeterminada** para ahorrar tokens. Usa `--edges` cuando necesites datos de relaciones como `calls`, `contains` o `inherits`.
- 🧾 Usa `--format json` cuando un script o integración necesite datos estructurados en lugar de `.skt`.
- 🧹 Usa siempre `--clean` al volver a ejecutar en el mismo directorio de salida.
- 📁 Usa `directory-tree` si deseas correlacionar archivos `.skt` directamente con archivos fuente en tu editor.