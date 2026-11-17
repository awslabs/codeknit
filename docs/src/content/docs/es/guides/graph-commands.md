---
title: Comandos de grafo
description: Visualice y analice la estructura de su código con algoritmos de grafos.
---

Utilice los subcomandos `graph` para ver cómo se conecta su código, encontrar problemas estructurales y decidir qué archivos refactorizar primero. Esta guía asume que ha [instalado codeknit](../../installation/) y puede ejecutar [`codeknit parse`](../parse-command/).

| Tarea | Comando |
| --- | --- |
| Explorar el grafo visualmente | `codeknit graph show` |
| Encontrar problemas estructurales | `codeknit graph analyze` |
| Encontrar archivos volátiles e importantes | `codeknit graph hotspots` |

## Visualizar el grafo de código

`graph show` analiza la entrada y escribe un archivo HTML autónomo. Los símbolos (funciones, clases, tipos) son nodos. Las relaciones (llamadas, contiene, implementa) son aristas.

```bash
codeknit graph show ./myproject
```

El archivo se escribe en `./skeleton/codeknit-graph.html` y se abre en su navegador predeterminado.

Para elegir la ruta de salida o incluir archivos de prueba:

```bash
codeknit graph show ./myproject -o graph.html
codeknit graph show ./src --collect-test
```

| Flag | Valor predeterminado | Propósito |
| --- | --- | --- |
| `-o`, `--output` | `./skeleton/codeknit-graph.html` | Ruta del archivo HTML de salida |
| `--collect-test` | `false` | Incluir archivos de prueba |
| `--workers` | `0` (`NumCPU`) | Máximo de goroutines de análisis concurrentes |
| `--verbose` | `false` | Mostrar información de progreso |

### Inspeccionar relaciones no resueltas

El visor dibuja solo las relaciones resueltas. Muestra un conteo de relaciones inciertas (`unresolved` o `ambiguous`) pero no las dibuja ni sus enlaces candidatos.

Para ver aristas inciertas individuales, ejecute `parse` con `--edges`:

```bash
codeknit parse ./src --output-mode inline --edges
```

Agregue `--format json` para salida JSON. Consulte la [referencia de formato de salida](../../reference/output-format/) para ver cómo se escriben las aristas y los candidatos.

## Analizar estructura y encontrar problemas

`graph analyze` ejecuta 22 algoritmos de grafos y escribe un informe `.skt` legible por LLM.

```bash
codeknit graph analyze ./myproject
```

El informe se escribe en `./skeleton/graph_analysis.skt`. Léalo directamente o proporciónelo a un LLM junto con el código fuente relevante.

### Ajustar el informe

Aumente un umbral para ver menos hallazgos, o redúzcalo para ver más:

```bash
codeknit graph analyze ./myproject -o analysis.skt --fan-threshold 15
codeknit graph graph analyze ./myproject --top-n 50
codeknit graph analyze ./src --collect-test
```

`--top-n 0` elimina el límite en las secciones clasificadas.

| Flag | Valor predeterminado | Propósito |
| --- | --- | --- |
| `-o`, `--output` | `./skeleton/graph_analysis.skt` | Ruta del archivo `.skt` de salida |
| `--collect-test` | `false` | Incluir archivos de prueba |
| `--workers` | `0` (`NumCPU`) | Máximo de goroutines de análisis concurrentes |
| `--verbose` | `false` | Mostrar información de progreso |
| `--fan-threshold` | `10` | Mínimo de fan-in o fan-out para marcar un símbolo como hub |
| `--god-threshold` | `15` | Mínimo de conteo de aristas de tipo contains para marcar una god class o función |
| `--max-inheritance-depth` | `5` | Marcar cadenas de herencia más profundas que este valor |
| `--top-n` | `30` | Limitar secciones clasificadas; `0` significa sin límite |
| `--betweenness-threshold` | `0.001` | Mínimo de centralidad de intermediación para reportar |
| `--propagation-cutoff` | `0.05` | Puntuación heurística mínima para continuar la propagación de cambios |

### Saber qué hallazgo responde a qué pregunta

| Algoritmo | Úselo para encontrar |
| --- | --- |
| Dependencias cíclicas (SCC de Tarjan) | Grupos de símbolos que dependen entre sí en un bucle |
| Detección de hubs | Símbolos con alto acoplamiento fan-in o fan-out |
| Detección de huérfanos | Candidatos a código muerto |
| Detección de god class/function | Símbolos que contienen un número excesivo de hijos |
| Métrica de inestabilidad | Símbolos y paquetes que son frágiles o estables, usando Ce/(Ca+Ce) |
| Cadenas de herencia profundas | Jerarquías más profundas que `--max-inheritance-depth` |
| Centralidad de intermediación | Símbolos cuello de botella por los que pasan muchas rutas de dependencia |
| Puntos de articulación | Puntos únicos de fallo cuya eliminación divide el grafo |
| PageRank | Símbolos que son recursivamente importantes |
| Fan-in transitivo | Radio de impacto: todo lo que depende de un símbolo, directa o indirectamente |
| Simulación de propagación de cambios | Qué tan lejos es probable que se propague un cambio a un símbolo |
| Dependencias cíclicas de paquetes | Bucles de dependencia a nivel de paquete |
| Detección de violaciones de capa | Dependencias que cruzan capas en la dirección incorrecta |
| Alcanzabilidad desde puntos de entrada | Código que ningún punto de entrada alcanza |
| Componentes débilmente conectados | Clusters aislados de código |
| Peso de dependencia | Qué tan fuertemente están acoplados los paquetes |
| Distancia desde la Secuencia Principal | Paquetes cuya abstracción e inestabilidad están desequilibradas |
| Detección de cirugía de escopeta | Cambios que afectarían muchos lugares en un paquete |
| Detección de envidia de características | Símbolos que usan otro paquete más que el propio |
| Violaciones de dependencia estable | Paquetes estables que dependen de otros menos estables |
| Violaciones de segregación de interfaces | Interfaces cuyos miembros son mayormente no utilizados por sus usuarios |
| Profundidad de contención | Símbolos profundamente anidados |

### Leer los resultados correctamente

Estas reglas afectan qué tan confiable puede ser un hallazgo:

- **Las relaciones inciertas están excluidas.** Las métricas de dependencia usan solo aristas resueltas. El informe cuenta las relaciones `unresolved` y `ambiguous` que dejó fuera.
- **Se ignora la metadata de archivos e includes.** Los símbolos `meta/file` y `meta/include` no afectan las clasificaciones de dependencia ni los hallazgos de código muerto.
- **Fan-in y fan-out cuentan vecinos distintos.** La contención no cuenta como una dependencia.
- **La profundidad de herencia cubre todas las bases e interfaces.** No se asigna profundidad a jerarquías cíclicas.
- **La alcanzabilidad comienza en callables `main`, `Main` o `init`** y sigue aristas de dependencia. Un miembro alcanzable hace relevante a su contenedor, pero no a sus miembros hermanos. Si el análisis no encuentra ningún punto de entrada reconocido, se omiten los hallazgos de alcanzabilidad.
- **Los callbacks son referencias, no llamadas.** Un argumento de callback o un callable retornado no prueba que el callback se ejecute.
- **Las puntuaciones son heurísticas.** La propagación de cambios usa la ruta de dependencia más fuerte. Sus puntuaciones y las reglas de capa no son probabilidades medidas ni garantías del compilador.

### Actuar sobre los hallazgos de manera segura

Trate cada hallazgo como un candidato para revisión. Antes de cambiar el código, abra los símbolos que nombra el informe y verifique:

- Llamadores externos fuera del directorio escaneado.
- Despacho dinámico.
- Dependencias que no escaneó.

El análisis se basa en sintaxis, por lo que cualquiera de estos puede cambiar la conclusión. Consulte [resolución de relaciones en la referencia de soporte de lenguajes](../../reference/language-support/#relationship-resolution) para conocer las reglas y límites de resolución.

## Encontrar hotspots basados en historial

`graph hotspots` clasifica archivos que cambian con frecuencia y ocupan posiciones estructuralmente importantes. También informa sobre acoplamiento temporal: archivos que cambian repetidamente en los mismos commits. Ejecútelo en un directorio dentro de un repositorio Git.

```bash
codeknit graph hotspots ./myproject
```

El informe se escribe en `./skeleton/hotspots.skt`.

Combina dos grupos de métricas:

- **Historial:** frecuencia de commits, agitación de líneas añadidas y eliminadas, y actualidad.
- **Estructura:** PageRank a nivel de archivo, fan-in transitivo y centralidad de intermediación.

El informe incluye un nivel de confianza basado en el número de commits relevantes analizados. Trate las clasificaciones de un informe de baja confianza como tentativas.

### Cambiar la ventana de historial y los filtros

```bash
# Analizar dos años y escribir JSON
codeknit graph hotspots ./myproject --since 2y --format json -o hotspots.json

# Permitir commits más grandes y requerir acoplamiento temporal más fuerte
codeknit graph hotspots . --max-files-per-commit 100 --min-cochanges 5
```

Los commits de fusión se excluyen a menos que pase `--include-merges`. Los commits que tocan más archivos que `--max-files-per-commit` se omiten como commits masivos, por lo que los cambios generados o mecánicos no distorsionan la clasificación.

| Flag | Valor predeterminado | Propósito |
| --- | --- | --- |
| `-o`, `--output` | `./skeleton/hotspots.skt` | Ruta del archivo de salida |
| `--format` | `skt` | Formato de salida: `skt` o `json` |
| `--since` | `12mo` | Ventana de historial, como `180d`, `12mo` o `2y` |
| `--max-commits` | `2000` | Máximo de commits a inspeccionar |
| `--max-files-per-commit` | `50` | Excluir commits que cambien más archivos que este valor |
| `--min-cochanges` | `3` | Mínimo de commits compartidos para reportar acoplamiento temporal |
| `--top-n` | `30` | Máximo de resultados por sección |
| `--include-merges` | `false` | Incluir commits de fusión en las métricas de historial |
| `--collect-test` | `false` | Incluir archivos de prueba |
| `--workers` | `0` (`NumCPU`) | Máximo de goroutines de análisis concurrentes |
| `--verbose` | `false` | Mostrar información de progreso |

### Leer la salida JSON

Con `--format json`, cada hotspot tiene `score`, `history_score`, `structure_score`, `recency_score`, `pagerank`, `betweenness`, `commits`, `churn` y `transitive_fan_in`. Cada entrada de acoplamiento temporal tiene `left`, `right`, `strength` y `co_changes`. El nivel superior también registra `commits_visited`, `commits_analyzed`, `skipped_merges` y `skipped_bulk_commits`. Use estos conteos para verificar en cuánto historial se basa la clasificación.

## Controlar qué fuentes se escanean

Los cinco comandos de procesamiento de fuentes (`parse`, `fingerprint`, `graph show`, `graph analyze` y `graph hotspots`) aceptan las mismas banderas de selección:

| Flag | Valor predeterminado | Propósito |
| --- | --- | --- |
| `--include GLOB` | Todas las fuentes admitidas | Incluir rutas que coincidan; repetible |
| `--exclude GLOB` | Ninguno | Excluir rutas o directorios que coincidan; repetible. Las exclusiones prevalecen sobre las inclusiones |
| `--max-file-size BYTES` | `10485760` (10 MiB) | Omitir archivos más grandes; `0` desactiva el límite |
| `--strict` | `false` | Fallar antes de escribir la salida si el escaneo o análisis está incompleto |

Los patrones son globs estilo gitignore relativos al directorio de entrada. Escríbalos entre comillas para que su shell no los expanda. Las inclusiones no anulan las reglas `.gitignore` ni el filtrado de pruebas.

```bash
codeknit graph analyze ./src --include '**/*.go' --exclude '**/generated/' --strict
```

Se omiten los enlaces simbólicos y los archivos no regulares, y se reportan. Los problemas de escaneo y análisis aparecen en stderr. Sin `--strict`, codeknit conserva resultados parciales utilizables, así que revise stderr antes de confiar en un informe. Con `--strict`, un análisis incompleto se detiene antes de escribir cualquier salida. Una ejecución en la que falla la lectura de todas las fuentes siempre falla.

Para conocer las garantías exactas, consulte la [referencia de confiabilidad del análisis](../../reference/analysis-reliability/).

## Páginas relacionadas

- [Referencia de banderas CLI](../../reference/cli-flags/): todas las banderas para cada comando.
- [Referencia de formato de salida](../../reference/output-format/): sintaxis `.skt` y JSON.
- [Referencia de soporte de lenguajes](../../reference/language-support/): límites de extracción y resolución.
- [Guía del comando parse](../parse-command/): extraer símbolos y aristas para archivos individuales.