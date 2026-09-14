---
title: Comandos de Grafo
description: Visualice y analice la estructura de su base de código con algoritmos de grafo.
---

codeknit proporciona comandos de grafo para visualizar la estructura, ejecutar
análisis automatizados y combinar el grafo de dependencias actual con el historial de cambios de Git.

## graph show

Genera una visualización interactiva de grafo en HTML de su base de código.

```bash
codeknit graph show <input-path>
```

Este comando analiza su base de código y produce un archivo HTML autónomo con una visualización interactiva de grafo. Los símbolos (funciones, clases, tipos) aparecen como nodos, y sus relaciones (llamadas, contiene, implementa) como relaciones. La visualización se abre automáticamente en su navegador predeterminado.

El visor dibuja relaciones resueltas. Muestra un conteo de incertidumbre pero no dibuja relaciones individuales no resueltas o ambiguas ni sus enlaces candidatos. Inspeccione esas relaciones con la salida de parse:

```bash
codeknit parse ./src --output-mode inline --edges
codeknit parse ./src --output-mode inline --format json --edges
```

### Flags

| Flag             | Default                          | Description                                  |
| ---------------- | -------------------------------- | -------------------------------------------- |
| `-o`, `--output` | `./skeleton/codeknit-graph.html` | Ruta del archivo HTML de salida             |
| `--collect-test` | `false`                          | Incluir archivos de prueba en el análisis   |
| `--workers`      | `NumCPU`                         | Máximo de goroutines de análisis concurrentes|
| `--verbose`      | `false`                          | Mostrar información de progreso durante el procesamiento |

### Ejemplos

```bash
# Generate default visualization
codeknit graph show ./myproject

# Custom output file
codeknit graph show ./myproject -o graph.html

# Include test files
codeknit graph show ./src --collect-test
```

## graph analyze

Ejecuta algoritmos de grafo estructurales en su base de código y emite un informe `.skt` legible por LLM que contiene información sobre la calidad del código.

```bash
codeknit graph analyze <input-path>
```

Este comando reporta candidatos para revisión, incluyendo dependencias cíclicas, símbolos hub, código muerto, god classes y cuellos de botella arquitectónicos.

### Algoritmos

El análisis incluye 22 algoritmos de grafo estructurales:

- Dependencias cíclicas (SCC de Tarjan)
- Detección de hubs (alto acoplamiento fan-in/fan-out)
- Detección de huérfanos (candidatos a código muerto)
- Detección de god class/function (excesivos hijos)
- Métrica de inestabilidad (Ce/(Ca+Ce) de Robert C. Martin)
- Cadenas de herencia profundas
- Centralidad de intermediación (detección de cuellos de botella)
- Puntos de articulación (puntos únicos de fallo)
- PageRank (importancia recursiva)
- Fan-in transitivo (radio de impacto)
- Simulación de propagación de cambios
- Dependencias cíclicas de paquetes
- Detección de violaciones de capa
- Alcanzabilidad desde puntos de entrada
- Componentes débilmente conectados
- Peso de dependencia (fuerza de acoplamiento de paquetes)
- Distancia desde la Secuencia Principal (equilibrio A+I)
- Detección de cirugía de escopeta
- Detección de envidia de características
- Violaciones de dependencia estable
- Violaciones de segregación de interfaces
- Profundidad de contención

### Interpretación del informe

Las métricas de dependencia utilizan relaciones resueltas. El informe cuenta las relaciones no resueltas y ambiguas excluidas. Los metadatos de archivos/inclusiones se excluyen de los rankings de dependencias y hallazgos de código muerto.

- Fan-in y fan-out cuentan vecinos distintos. Las relaciones repetidas no inflan estos conteos, y la contención no es una dependencia.
- La profundidad de herencia considera todas las bases e interfaces. Las jerarquías cíclicas no tienen asignada una profundidad.
- La alcanzabilidad comienza desde callables nombrados `main`, `Main` o `init` y sigue las relaciones de dependencia. Alcanzar un miembro hace que su contenedor sea relevante sin marcar todos los hermanos como alcanzables. Sin un punto de entrada reconocido, los hallazgos de alcanzabilidad se omiten.
- Los argumentos de callback y los callables devueltos son referencias, no prueba de que el callback se ejecute.
- La propagación de cambios utiliza el camino de dependencia más fuerte. Sus puntuaciones y reglas de capa basadas en directorios son heurísticas, no probabilidades medidas.

Los llamadores externos, el despacho dinámico y las dependencias fuera del escaneo pueden cambiar las conclusiones. Consulte [Soporte de Lenguajes y Fidelidad de API](../../reference/language-support/) para conocer las reglas y límites de resolución.

### Flags

| Flag                      | Default                         | Description                                              |
| ------------------------- | ------------------------------- | -------------------------------------------------------- |
| `-o`, `--output`          | `./skeleton/graph_analysis.skt` | Ruta del archivo `.skt` de salida                        |
| `--collect-test`          | `false`                         | Incluir archivos de prueba en el análisis                |
| `--workers`               | `NumCPU`                        | Máximo de goroutines de análisis concurrentes            |
| `--verbose`               | `false`                         | Mostrar información de progreso durante el procesamiento |
| `--fan-threshold`         | `10`                            | Mínimo de fan-in o fan-out para marcar un símbolo hub    |
| `--god-threshold`         | `15`                            | Mínimo de conteo de relaciones contiene para marcar una god class/function |
| `--max-inheritance-depth` | `5`                             | Marcar cadenas de herencia más profundas que esto        |
| `--top-n`                 | `30`                            | Limitar las secciones de salida clasificadas; 0 = sin límite |
| `--betweenness-threshold` | `0.001`                         | Valor mínimo de centralidad de intermediación para reportar |
| `--propagation-cutoff`    | `0.05`                          | Puntuación heurística mínima para continuar la propagación de cambios |

### Ejemplos

```bash
# Run structural analysis with defaults
codeknit graph analyze ./myproject

# Custom output and thresholds
codeknit graph analyze ./myproject -o analysis.skt --fan-threshold 15

# Show more results per section
codeknit graph analyze ./myproject --top-n 50

# Remove the cap on ranked sections
codeknit graph analyze ./myproject --top-n 0

# Include test files
codeknit graph analyze ./src --collect-test
```

## graph hotspots

Clasifica archivos que son tanto frecuentemente modificados como estructuralmente importantes:

```bash
codeknit graph hotspots <input-path>
```

La puntuación combina frecuencia de commits, cambios en líneas y actualidad con PageRank a nivel de archivo, fan-in transitivo y centralidad de intermediación. El informe también identifica acoplamiento temporal entre archivos que cambian repetidamente en los mismos commits.

Los commits de fusión están excluidos por defecto. También se excluyen los commits que cambian más de 50 archivos para que los cambios generados, vendidos o mecánicos a granel no distorsionen los resultados.

### Flags

| Flag                     | Default                   | Description                                     |
| ------------------------ | ------------------------- | ----------------------------------------------- |
| `-o`, `--output`         | `./skeleton/hotspots.skt` | Ruta del archivo de salida                      |
| `--format`               | `skt`                     | Formato de salida: `skt` o `json`              |
| `--since`                | `12mo`                    | Ventana de historial, como `180d`, `12mo`, o `2y` |
| `--max-commits`          | `2000`                    | Máximo de commits a inspeccionar                |
| `--max-files-per-commit` | `50`                      | Excluir commits que cambien más archivos        |
| `--min-cochanges`        | `3`                       | Mínimo de commits compartidos para acoplamiento temporal |
| `--top-n`                | `30`                      | Máximo de resultados por sección del informe    |
| `--include-merges`       | `false`                   | Incluir commits de fusión                       |
| `--collect-test`         | `false`                   | Incluir archivos de prueba                      |
| `--workers`              | `NumCPU`                  | Máximo de goroutines de análisis concurrentes   |
| `--verbose`              | `false`                   | Mostrar información de progreso                 |

### Ejemplos

```bash
# Analyze the last 12 months
codeknit graph hotspots ./myproject

# Analyze two years and emit JSON
codeknit graph hotspots ./myproject --since 2y --format json -o hotspots.json

# Include larger commits and require stronger coupling
codeknit graph hotspots . --max-files-per-commit 100 --min-cochanges 5
```