---
title: Modos de salida
description: Elija el modo de salida adecuado para el tamaño de su proyecto y su flujo de trabajo.
---

codeknit admite tres modos de salida, controlados por la bandera `--output-mode`. Cada modo determina cómo se escribe en disco (o en stdout) la estructura de código extraída.

### directory-flat (predeterminado, recomendado)

- **Comportamiento**: Escribe archivos `.skt` divididos como `map_001.skt`, `map_002.skt`, etc.
- **Directorio de salida**: `./skeleton/` de forma predeterminada
- **División**: Los archivos se dividen cuando superan el límite de `--max-lines` (predeterminado: 500 líneas)
- **Caso de uso**: Mejor para la mayoría de los proyectos. Mantiene la salida organizada y legible limitando el tamaño del archivo. Puede leer solo los fragmentos relevantes para su tarea.
- **Minificación**: Cuando `--minify` está habilitado, también se genera un archivo `dict.skt` en el directorio de salida, que contiene las asignaciones de tokens para valores comprimidos.

Ejemplo:

```bash
codeknit parse ./src
# Salida: ./skeleton/map_001.skt, map_002.skt, ...
```

### directory-tree

- **Comportamiento**: Refleja exactamente la estructura del directorio de origen.
- **Directorio de salida**: `./skeleton/` de forma predeterminada
- **Asignación**: Se crea un archivo `.skt` por cada archivo de origen, en una ruta correspondiente.
- **Caso de uso**: Ideal cuando desea buscar rápidamente la estructura de un archivo específico. Útil para la navegación junto con el código base original.

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

### Tabla de decisión

| Modo             | Mejor para                                                | Ubicación de salida                                                    |
| ---------------- | --------------------------------------------------------- | ---------------------------------------------------------------------- |
| `directory-flat` | La mayoría de los proyectos (predeterminado, recomendado) | `./skeleton/map_001.skt`, `map_002.skt`, ...                           |
| `directory-tree` | Navegar por la salida junto con el código fuente          | `./skeleton/<ruta reflejada>.skt`                                      |
| `inline`         | Archivo individual, canalización a otra herramienta       | stdout — solo usar para archivos individuales o proyectos muy pequeños |

### Reglas generales

- **Si no está seguro** → use `directory-flat` (el predeterminado)
- **Inspección de un solo archivo** → `inline` es aceptable
- **Más de unos pocos archivos** → prefiera `directory-flat` o `directory-tree`
- **Bases de código grandes** → agregue `--minify` para reducir el uso de tokens
- **Volver a ejecutar en la misma salida** → use `--clean` para eliminar archivos `.skt` obsoletos

### Minificación

La bandera `--minify` habilita la compresión basada en diccionario de tokens repetidos (por ejemplo, claves de propiedades como `exported`, `async` o nombres de tipos comunes). Cuando está habilitada:

- Los valores repetidos se reemplazan con códigos cortos (`d0`, `d1`, `d2`, ...)
- Se escribe un archivo `dict.skt` en el directorio de salida, que asigna códigos a valores originales
- Reduce significativamente el tamaño de la salida para bases de código grandes
- Funciona en los modos `directory-flat` y `directory-tree`

Ejemplo de salida minificada:

```skt
[dict]
- d0: exported
- d1: callable/function

[symbols]
## src/main.py
S1 d1 L1-L5 main() {d0}
```
