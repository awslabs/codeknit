---
title: Modos de salida
description: Elige el modo de salida adecuado para el tamaño de tu proyecto y tu flujo de trabajo.
---

`codeknit` admite tres modos de salida, controlados por la bandera `--output-mode`. Cada modo determina cómo se escribe en disco (o en stdout) la estructura de código extraída.

El modo de salida es independiente del formato de salida. El formato predeterminado es `.skt`; pasa `--format json` para emitir el mismo resultado del análisis en JSON legible por máquina. En los modos de directorio, JSON se escribe en `codeknit.json`. En el modo `inline`, JSON se escribe en stdout.

### directory-flat (predeterminado, recomendado)

- **Comportamiento**: Escribe archivos `.skt` divididos como `map_001.skt`, `map_002.skt`, etc.
- **Directorio de salida**: `./skeleton/` de forma predeterminada
- **División**: Los archivos se dividen cuando superan el límite `--max-lines` (predeterminado: 500 líneas)
- **Caso de uso**: Mejor para la mayoría de los proyectos. Mantiene la salida organizada y legible limitando el tamaño del archivo. Puedes leer solo los fragmentos relevantes para tu tarea.
- **Minificación**: Cuando `--minify` está habilitado, también se genera un archivo `dict.skt` en el directorio de salida, que contiene las asignaciones de tokens para valores comprimidos.

Ejemplo:

```bash
codeknit parse ./src
# Salida: ./skeleton/map_001.skt, map_002.skt, ...
```

### directory-tree

- **Comportamiento**: Refleja exactamente la estructura del directorio fuente.
- **Directorio de salida**: `./skeleton/` de forma predeterminada
- **Asignación**: Se crea un archivo `.skt` por cada archivo fuente, en una ruta correspondiente.
- **Caso de uso**: Ideal cuando deseas buscar rápidamente la estructura de un archivo específico. Útil para navegar junto con el código base original.

Ejemplo:

```bash
codeknit parse ./src --output-mode directory-tree
# Salida: ./skeleton/src/handler.skt, ./skeleton/pkg/db.skt, etc.
```

### inline

- **Comportamiento**: Vuelca toda la salida a stdout.
- **Directorio de salida**: Ninguno creado
- **Caso de uso**: Solo recomendado para archivos individuales o proyectos muy pequeños (menos de 5 archivos). Útil cuando se canaliza la salida a otra herramienta o se inspecciona un archivo individual de manera interactiva.

Ejemplo:

```bash
codeknit parse ./src/main.go --output-mode inline
# Salida: impresa directamente en la terminal
```

### Formato JSON

- **Comportamiento**: Emite un único documento JSON que contiene `files`, `symbols`, `edges` opcionales y `errors` opcionales.
- **Ubicación de salida**: `codeknit.json` en los modos de directorio, o stdout en el modo `inline`.
- **Caso de uso**: Mejor para scripts, integraciones de editores, verificaciones de CI y herramientas que necesitan datos estructurados.

Ejemplo:

```bash
codeknit parse ./src --output-mode inline --format json --edges
```

Salida de ejemplo:

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

### Tabla de decisión

| Modo             | Mejor para                                | Ubicación de salida                                     |
| ---------------- | ----------------------------------------- | ------------------------------------------------------- |
| `directory-flat` | La mayoría de los proyectos (predeterminado, recomendado) | `./skeleton/map_001.skt`, `map_002.skt`, ...            |
| `directory-tree` | Navegar por la salida junto con el código fuente | `./skeleton/<ruta reflejada>.skt`                       |
| `inline`         | Archivo individual, canalización a otra herramienta | stdout — solo usar para archivos individuales o proyectos muy pequeños |

| Formato | Mejor para                           | Salida                                                   |
| ------ | ------------------------------------ | -------------------------------------------------------- |
| `skt`  | Contexto de LLM e inspección humana | Archivos `.skt` o stdout                                 |
| `json` | Scripts e integración estructurada  | `codeknit.json` en modos de directorio, o stdout en `inline` |

### Reglas generales

- **Si no estás seguro** → usa `directory-flat` (el predeterminado)
- **Inspección de un solo archivo** → `inline` es aceptable
- **Más de unos pocos archivos** → prefiere `directory-flat` o `directory-tree`
- **Bases de código grandes** → añade `--minify` para reducir el uso de tokens
- **Volver a ejecutar en la misma salida** → usa `--clean` para eliminar archivos `.skt` obsoletos

### Minificación

La bandera `--minify` habilita la compresión basada en diccionario de tokens repetidos (por ejemplo, claves de propiedades como `exported`, `async` o nombres de tipos comunes). Cuando está habilitada:

- Los valores repetidos se reemplazan con códigos cortos (`d0`, `d1`, `d2`, ...)
- Se escribe un archivo `dict.skt` en el directorio de salida, que asigna códigos a valores originales
- Reduce significativamente el tamaño de la salida para bases de código grandes
- Funciona tanto en el modo `directory-flat` como en `directory-tree`

Ejemplo de salida minificada:

```skt
[dict]
- d0: exported
- d1: callable/function

[symbols]
## src/main.py
S1 d1 L1-L5 main() {d0}
```