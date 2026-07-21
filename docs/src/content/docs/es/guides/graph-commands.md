---
title: Comandos de Grafo
description: Visualice y analice la estructura de su base de código con algoritmos de grafo.
---

codeknit proporciona comandos de grafo para visualizar la estructura, ejecutar análisis automatizados y combinar el grafo de dependencias actual con el historial de cambios de Git.

## graph show

Genera una visualización interactiva de grafo en HTML de su base de código.

```bash
codeknit graph show <input-path>
```

Este comando analiza su base de código y produce un archivo HTML autónomo con una visualización interactiva de grafo. Los símbolos (funciones, clases, tipos) aparecen como nodos, y sus relaciones (llamadas, contiene, implementa) como relaciones. La visualización se abre automáticamente en su navegador predeterminado.

### Flags

| Flag             | Default                          | Description                                  |
| ---------------- | -------------------------------- | -------------------------------------------- |
| `-o`, `--output` | `./skeleton/codeknit-graph.html` | Ruta del archivo HTML de salida              |
| `--collect-test` | `false`                          | Incluir archivos de prueba en el análisis    |
| `--workers`      | `NumCPU`                         | Máximo de goroutines de análisis concurrentes|
| `--verbose`      | `false`                          | Mostrar información de progreso durante el procesamiento |

### Ejemplos

```skt
# Generar visualización predeterminada
codeknit graph show ./myproject

# Archivo de salida personalizado
codeknit graph show ./myproject -o graph.html

# Incluir archivos de prueba
codeknit graph show ./src --collect-test
```

## graph analyze

Ejecuta algoritmos de grafo estructural en su base de código y emite un informe `.skt` legible por LLM que contiene insights sobre la calidad del código.

```bash
codeknit graph analyze <input-path>
```

Este comando detecta problemas comunes de calidad de código como dependencias cíclicas, símbolos hub, código muerto, god classes y cuellos de botella arquitectónicos.

### Algoritmos

El análisis incluye 22 algoritmos de grafo estructural:

- Dependencias cíclicas (SCC de Tarjan)
- Detección de hubs (alto acoplamiento fan-in/fan-out)
- Detección de huérfanos (candidatos a código muerto)
- Detección de god class/function (exceso de hijos)
- Métrica de inestabilidad (Ce/(Ca+Ce) de Robert C. Martin)
- Cadenas de herencia profundas
- Centralidad de intermediación (detección de cuellos de botella)
- Puntos de articulación (puntos únicos de fallo)
- PageRank (importancia recursiva)
- Fan-in transitivo (radio de impacto)
- Simulación de propagación de cambios
- Dependencias cíclicas de paquetes
- Detección de violaciones de capas
- Alcanzabilidad desde puntos de entrada
- Componentes débilmente conectados
- Peso de dependencia (fuerza de acoplamiento de paquetes)
- Distancia desde la Secuencia Principal (balance A+I)
- Detección de cirugía de escopeta
- Detección de envidia de características
- Violaciones de dependencia estable
- Violaciones de segregación de interfaces
- Profundidad de contención

### Flags

| Flag                      | Default                         | Description                                              |
| ------------------------- | ------------------------------- | -------------------------------------------------------- |
| `-o`, `--output`          | `./skeleton/graph_analysis.skt` | Ruta del archivo `.skt` de salida                        |
| `--collect-test`          | `false`                         | Incluir archivos de prueba en el análisis                |
| `--workers`               | `NumCPU`                        | Máximo de goroutines de análisis concurrentes            |
| `--verbose`               | `false`                         | Mostrar información de progreso durante el procesamiento |
| `--fan-threshold`         | `10`                            | Mínimo fan-in o fan-out para marcar un símbolo hub       |
| `--god-threshold`         | `15`                            | Mínimo conteo de relaciones contiene para marcar una god class/function |
| `--max-inheritance-depth` | `5`                             | Marcar cadenas de herencia más profundas que esto        |
| `--top-n`                 | `30`                            | Limitar secciones de salida clasificadas; 0 = sin límite |
| `--betweenness-threshold` | `0.001`                         | Valor mínimo de centralidad de intermediación para reportar |
| `--propagation-cutoff`    | `0.05`                          | Probabilidad mínima para continuar la propagación de cambios |

### Ejemplos

```skt
# Ejecutar análisis estructural con valores predeterminados
codeknit graph analyze ./myproject

# Salida y umbrales personalizados
codeknit graph analyze ./myproject -o analysis.skt --fan-threshold 15

# Mostrar más resultados por sección
codeknit graph analyze ./myproject --top-n 50

# Incluir archivos de prueba
codeknit graph analyze ./src --collect-test
```

## graph hotspots

Clasifica los archivos que son tanto frecuentemente modificados como estructuralmente importantes:

```bash
codeknit graph hotspots <input-path>
```

La puntuación combina frecuencia de commits, cambios en líneas y actualidad con PageRank a nivel de archivo, fan-in transitivo y centralidad de intermediación. El informe también identifica acoplamiento temporal entre archivos que se modifican repetidamente en los mismos commits.

Los commits de fusión se excluyen de forma predeterminada. También se excluyen los commits que modifican más de 50 archivos para que los cambios generados, vendidos o mecánicos a granel no distorsionen los resultados.

### Flags

| Flag                     | Default                   | Description                                      |
| ------------------------ | ------------------------- | ------------------------------------------------ |
| `-o`, `--output`         | `./skeleton/hotspots.skt` | Ruta del archivo de salida                       |
| `--format`               | `skt`                     | Formato de salida: `skt` o `json`                |
| `--since`                | `12mo`                    | Ventana de historial, como `180d`, `12mo` o `2y` |
| `--max-commits`          | `2000`                    | Máximo de commits a inspeccionar                 |
| `--max-files-per-commit` | `50`                      | Excluir commits que modifiquen más archivos      |
| `--min-cochanges`        | `3`                       | Mínimo de commits compartidos para acoplamiento temporal |
| `--top-n`                | `30`                      | Máximo de resultados por sección del informe     |
| `--include-merges`       | `false`                   | Incluir commits de fusión                        |
| `--collect-test`         | `false`                   | Incluir archivos de prueba                       |
| `--workers`              | `NumCPU`                  | Máximo de goroutines de análisis concurrentes    |
| `--verbose`              | `false`                   | Mostrar información de progreso                  |

### Ejemplos

```bash
# Analizar los últimos 12 meses
codeknit graph hotspots ./myproject

# Analizar dos años y emitir JSON
codeknit graph hotspots ./myproject --since 2y --format json -o hotspots.json

# Incluir commits más grandes y requerir acoplamiento más fuerte
codeknit graph hotspots . --max-files-per-commit 100 --min-cochanges 5
```