---
title: Comandos de grafo
description: Visualiza y analiza la estructura de tu base de código con algoritmos de grafo.
---

codeknit proporciona dos potentes comandos de grafo para ayudarte a entender y mejorar la estructura de tu base de código: `graph show` para visualización interactiva y `graph analyze` para análisis estructural automatizado.

## graph show

Genera una visualización interactiva de grafo en HTML de tu base de código.

```bash
codeknit graph show <input-path>
```

Este comando analiza tu base de código y produce un archivo HTML autónomo con una visualización interactiva de grafo. Los símbolos (funciones, clases, tipos) aparecen como nodos, y sus relaciones (llamadas, contiene, implementa) como relaciones. La visualización se abre automáticamente en tu navegador predeterminado.

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

Ejecuta algoritmos de grafo estructurales en tu base de código y emite un informe `.skt` legible por LLM que contiene insights sobre la calidad del código.

```bash
codeknit graph analyze <input-path>
```

Este comando detecta problemas comunes de calidad de código como dependencias cíclicas, símbolos hub, código muerto, god classes y cuellos de botella arquitectónicos.

### Algoritmos

El análisis incluye 22 algoritmos de grafo estructurales:

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
- Dependencias circulares de paquetes
- Detección de violaciones de capas
- Alcanzabilidad desde puntos de entrada
- Componentes débilmente conectados
- Peso de dependencia (fuerza de acoplamiento de paquetes)
- Distancia desde la Secuencia Principal (balance A+I)
- Detección de shotgun surgery
- Detección de feature envy
- Violaciones del principio de dependencias estables
- Violaciones del principio de segregación de interfaces
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

# Salida personalizada y umbrales
codeknit graph analyze ./myproject -o analysis.skt --fan-threshold 15

# Mostrar más resultados por sección
codeknit graph analyze ./myproject --top-n 50

# Incluir archivos de prueba
codeknit graph analyze ./src --collect-test
```
