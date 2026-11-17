---
title: Comando Fingerprint
description: Detectar código duplicado y casi duplicado en archivos y lenguajes utilizando fuzzy hashing.
---

`codeknit fingerprint` calcula un Context-Triggered Piecewise Hash (CTPH) sobre los tokens normalizados del cuerpo de cada función, método, variable y tipo. Los nombres de variables, literales de cadena y anotaciones de tipo se normalizan antes de aplicar el hash. Por lo tanto, las copias renombradas y los ports a otros lenguajes aún pueden coincidir.

Esta guía muestra cómo encontrar duplicados, ajustar el rango de similitud, reducir falsos positivos y decidir qué refactorizar.

## Antes de empezar

- `codeknit` está instalado. Consulta [Instalación](../../installation/).
- Tienes un directorio fuente para analizar, como `./src`.

## Encontrar duplicados en un proyecto

Ejecuta el comando en el directorio que deseas revisar:

```bash
codeknit fingerprint ./src
```

El comando escribe `./skeleton/fingerprints.skt` e informa pares con similitud del 65% al 95%.

Abre el archivo. La sección `[duplicates]` siempre está presente, con las coincidencias más fuertes primero:

```skt
[duplicates]
# similarity range: 65%-95%
similarity:93%  pkg/user.go::GetUser <-> pkg/admin.go::GetAdmin
similarity:88%  src/auth.ts::validateToken <-> src/session.ts::validateSession
similarity:81%  lib/math.py::compute_tax <-> lib/math.py::compute_vat
```

Cada línea tiene el formato `similarity:<N>%  <file>::<símbolo> <-> <file>::<símbolo>`. Si no hay nada en el rango, la sección contiene solo `# no duplicates found`.

Para la gramática completa de salida, consulta la [referencia de formato de salida](../../reference/output-format/).

## Encontrar solo duplicados exactos

El rango predeterminado llega hasta el 95%, por lo que se excluyen las coincidencias del 100%. Para listar coincidencias exactas, establece ambos límites en 100:

```bash
codeknit fingerprint ./src --min-similarity 100 --max-similarity 100
```

"Exacto" significa que las secuencias de tokens normalizados completas son idénticas dentro de la misma categoría de símbolo. No significa que el texto fuente sea idéntico, ni que los dos símbolos sean equivalentes probados.

## Ampliar el rango para encontrar coincidencias más flexibles

Reduce el mínimo para incluir código que comparta una forma algorítmica pero difiera en detalles:

```bash
codeknit fingerprint ./src --min-similarity 50 --max-similarity 80
```

Los resultados por debajo del 50% suelen ser ruido.

## Reducir falsos positivos con reranking semántico

Si los resultados estructurales son ruidosos, reranquea los candidatos con embeddings de un servidor Ollama local.

1. Inicia Ollama y descarga el modelo de embedding predeterminado:

   ```bash
   ollama serve
   ollama pull qwen3-embedding:0.6b
   ```

2. Ejecuta el comando con `--rerank`:

   ```bash
   codeknit fingerprint ./src --rerank
   ```

Para usar otro modelo de embedding, pasa `--model`:

```bash
codeknit fingerprint ./src --rerank --model qwen3-embedding:4b
```

El reranking añade candidatos semánticos y filtra según una puntuación combinada. Reduce falsos positivos pero no los elimina todos. También puede eliminar un candidato que sea estructuralmente exacto, y el rango de similitud sigue aplicándose. Los embeddings se calculan a partir de la misma instantánea de código fuente analizada utilizada para la extracción.

## Inspeccionar las huellas digitales sin procesar

Para ver la huella digital de cada símbolo, añade `--show-all`:

```bash
codeknit fingerprint ./src --show-all
```

Esto añade una sección `[fingerprints]`, agrupada por archivo fuente, una línea por símbolo:

```skt
[fingerprints]
## src/auth.ts
validateToken  FP:3:a1b2c3...:d4e5f6...  tokens:8e0f1a2b...
```

El formato de la línea es `<nombre-símbolo>  FP:<blocksize>:<hash1>:<hash2>  tokens:<hex>`. `tokens` es la secuencia de tokens normalizados codificada en hexadecimal.

## Escribir el informe en otra ubicación

```bash
codeknit fingerprint ./src -o duplicates.skt
```

## Controlar qué archivos se analizan

Los archivos de prueba se omiten de forma predeterminada. Inclúyelos con `--collect-test`:

```bash
codeknit fingerprint ./src --collect-test
```

Para limitar el análisis, usa las opciones de selección de fuentes. Añade comillas a los patrones para que tu shell no los expanda:

```bash
codeknit fingerprint ./src --include '**/*.ts' --exclude '**/generated/'
```

Otros controles de selección:

- `--max-file-size BYTES` omite archivos más grandes. El valor predeterminado es `10485760` (10 MiB). `0` desactiva el límite.
- `--strict` falla antes de escribir la salida si el análisis o el parseo están incompletos.
- `--header-language c|cpp` selecciona el parser para archivos `.h` ambiguos. Consulta [Soporte de lenguajes](../../reference/language-support/#selección-de-lenguaje-de-encabezado).

Los enlaces simbólicos omitidos, archivos demasiado grandes y advertencias de sintaxis aparecen en stderr. Los includes no anulan `.gitignore` ni el filtrado de pruebas, y los excludes tienen prioridad sobre los includes. Consulta la [referencia de confiabilidad del análisis](../../reference/analysis-reliability/) para garantías exactas.

## Elegir un rango de similitud

| Rango     | Qué suele revelar                                      |
| --------- | ------------------------------------------------------------- |
| `96-100%` | Duplicados estructurales exactos o casi exactos; copia-pega clara   |
| `85-95%`  | Casi duplicados; generalmente copia-pega con ediciones triviales        |
| `65-84%`  | Rango predeterminado; fuerte similitud estructural, a menudo candidatos para refactorizar |
| `50-64%`  | Misma forma algorítmica con detalles diferentes                 |
| Menos del 50% | Generalmente ruido                                                 |

Para ver las dos primeras filas, aumenta `--max-similarity` a `100`.

## Decidir qué refactorizar

1. Comienza con el rango predeterminado y lee los pares principales.
2. Abre ambos símbolos de cada par y compáralos.
3. Trata una coincidencia como una señal, no como una prueba. Las huellas digitales miden la forma estructural, no los nombres de identificadores ni la intención. La misma lógica escrita de forma iterativa y recursiva obtiene una puntuación baja, y código no relacionado con la misma forma puede obtener una puntuación alta.
4. Los pares en el rango del 85–95% suelen ser los candidatos más seguros para consolidar. Los pares en el rango del 65–84% a menudo requieren una revisión más detallada.
5. Si la lista es ruidosa, reduce el rango o añade `--rerank`.

Las coincidencias entre lenguajes funcionan para constructos equivalentes. Las características específicas del lenguaje pueden producir coincidencias espurias de baja similitud.

Los porcentajes de similitud son una clasificación heurística, no una distancia de edición. La referencia de calibración etiquetada del proyecto informa precisión y recall por separado de las puntuaciones de similitud. Es una verificación de regresión, no una estimación de precisión en tu código.

## Proyectos grandes

- Con menos de 500 símbolos elegibles, codeknit puntúa cada par de la misma categoría.
- Con 500 o más, los grupos completos de tokens normalizados exactos siempre se conservan. Otros pares usan aproximadamente los 50 vecinos principales por símbolo, por lo que los resultados de casi duplicados pueden estar incompletos.
- Para duplicados exactos, usa el comando 100/100 anterior. No se ve afectado por la aproximación.
- El ordenamiento es determinista para entradas y puntuaciones estables.

## Solución de problemas

| Síntoma | Causa y solución |
| ------- | ------------- |
| `# no duplicates found` | Ningún par cae en el rango. Reduce `--min-similarity` o aumenta `--max-similarity`. |
| Falta una función corta conocida | Se omiten los símbolos con menos de 4 tokens normalizados. |
| Faltan copias exactas | El máximo predeterminado es 95%. Usa `--min-similarity 100 --max-similarity 100`. Con `--rerank`, el filtro semántico también puede eliminarlos. |
| `--rerank` falla | Verifica que `ollama serve` esté en ejecución y que el modelo esté descargado. |
| Resultados parciales con advertencias en stderr | Corrige los archivos reportados o añade `--strict` para fallar en lugar de publicar salida incompleta. |

## Opciones utilizadas en esta guía

| Opción               | Valor predeterminado                       | Descripción                                                                 |
| ------------------ | ----------------------------- | --------------------------------------------------------------------------- |
| `-o`, `--output`   | `./skeleton/fingerprints.skt` | Ruta del archivo `.skt` de salida.                                                    |
| `--min-similarity` | `65`                          | Porcentaje mínimo de similitud para informar (0–100).                            |
| `--max-similarity` | `95`                          | Porcentaje máximo de similitud para informar (0–100).                            |
| `--show-all`       | `false`                       | También emite la sección `[fingerprints]` con datos de tokens sin procesar.                 |
| `--rerank`         | `false`                       | Reranquea candidatos con embeddings semánticos vía Ollama. Requiere `ollama serve` y `ollama pull qwen3-embedding:0.6b`. |
| `--model`          | `qwen3-embedding:0.6b`        | Modelo de embedding de Ollama para usar con `--rerank`.                              |
| `--collect-test`   | `false`                       | Incluir archivos de prueba en el análisis.                                             |
| `--workers`        | `NumCPU`                      | Máximo de goroutines de parseo concurrentes.                                      |
| `--verbose`        | `false`                       | Mostrar información de progreso.                                             |

Para la lista completa, incluyendo `--include`, `--exclude`, `--max-file-size`, `--strict` y `--header-language`, consulta la [referencia de opciones de CLI](../../reference/cli-flags/).

## Relacionado

- [Comando Parse](../parse-command/)
- [Comandos de grafo](../graph-commands/)