---
title: Comando Fingerprint
description: Detectar código duplicado y casi duplicado en archivos y lenguajes utilizando hashing difuso.
---

El comando `codeknit fingerprint` detecta código duplicado y casi duplicado en tu base de código utilizando **Context-Triggered Piecewise Hashing (CTPH)**. Funciona entre archivos e incluso entre lenguajes de programación normalizando nombres de variables, literales de cadena y anotaciones de tipo antes de calcular las huellas estructurales normalizadas.

## Qué hace

`codeknit fingerprint` analiza cada función, método, variable y tipo en tu base de código y calcula una **huella estructural normalizada** basada en:

- Flujo de control (`if`, `for`, `while`, `switch`)
- Operaciones (`=`, `+`, `==`, `&&`, `||`)
- Llamadas, retornos, asignaciones y creación de objetos
- Constructores del lenguaje como `try/catch`, `yield`, `await`, `defer`

Esta normalización significa que **copiar y pegar con cambios de nombre**, **refactorizaciones triviales** y **lógica equivalente en diferentes lenguajes** aún pueden ser detectados como duplicados.

El algoritmo utiliza **CTPH** (una variante de hash rodante) para encontrar casi duplicados de manera eficiente. Código similar produce huellas similares, permitiendo coincidencias difusas incluso cuando el código ha sido ligeramente modificado.

## Uso básico

```bash
codeknit fingerprint ./src
```

Este comando:

- Analiza todos los archivos fuente en `./src`
- Calcula huellas estructurales
- Genera resultados en `./skeleton/fingerprints.skt`
- Reporta coincidencias con similitud entre **65% y 95%** (rango predeterminado)

## Flags

| Flag               | Valor predeterminado            | Descripción                                                                                                                                                |
| ------------------ | ------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `-o`, `--output`   | `./skeleton/fingerprints.skt`   | Ruta del archivo `.skt` de salida                                                                                                                          |
| `--min-similarity` | `65`                            | Porcentaje mínimo de similitud para reportar (0–100)                                                                                                        |
| `--max-similarity` | `95`                            | Porcentaje máximo de similitud para reportar (0–100)                                                                                                        |
| `--show-all`       | `false`                         | Incluir la sección `[fingerprints]` con datos de tokens sin procesar                                                                                        |
| `--rerank`         | `false`                         | Reordenar candidatos CTPH utilizando embeddings semánticos vía Ollama para eliminar falsos positivos (requiere: `ollama serve` y `ollama pull qwen3-embedding:0.6b`) |
| `--model`          | `qwen3-embedding:0.6b`          | Modelo de embedding de Ollama a utilizar con `--rerank`                                                                                                    |
| `--collect-test`   | `false`                         | Incluir archivos de prueba en el análisis                                                                                                                  |
| `--workers`        | `NumCPU`                        | Número máximo de goroutines de análisis concurrentes (0 = usar todos los núcleos de CPU)                                                                    |
| `--verbose`        | `false`                         | Mostrar información de progreso durante el procesamiento                                                                                                   |

## Formato de salida

La salida es un archivo `.skt` con las siguientes secciones:

### `[duplicates]` (siempre presente)

Lista pares de símbolos con similitud por encima del umbral:

```skt
[duplicates]
similarity:96%  pkg/user.go::GetUser <-> pkg/admin.go::GetAdmin
similarity:88%  utils/str.go::TrimSpaces <-> lib/text.go::CleanString
```

Cada línea muestra:

- Porcentaje de similitud
- Símbolo izquierdo (ruta del archivo, ámbito, nombre)
- Símbolo derecho (ruta del archivo, ámbito, nombre)

### `[fingerprints]` (solo con `--show-all`)

Contiene datos de huella sin procesar para cada símbolo:

```skt
[fingerprints]
validateToken  FP:3:a1b2c3...:d4e5f6...  tokens:8e0f1a2b...
```

Campos:

- Nombre del símbolo
- `FP:<versión>:<hash1>:<hash2>` — huella CTPH
- `tokens:<hex>` — flujo de tokens del cuerpo normalizado

Esta sección es útil para depuración o construcción de herramientas downstream.

## Patrones comunes

```bash
# Análisis predeterminado
codeknit fingerprint /codeknit/es/src
```

```bash
# Encontrar solo duplicados exactos
codeknit fingerprint /codeknit/es/src --min-similarity 100
```

```bash
# Encontrar código moderadamente similar (ej. mismo algoritmo, nombres diferentes)
codeknit fingerprint /codeknit/es/src --min-similarity 50 --max-similarity 80
```

```bash
# Usar reordenamiento semántico para reducir falsos positivos
# Requiere: ollama serve && ollama pull qwen3-embedding:0.6b
codeknit fingerprint /codeknit/es/src --rerank
```

```bash
# Usar un modelo de embedding diferente para reordenamiento
codeknit fingerprint /codeknit/es/src --rerank --model qwen3-embedding:4b
```

```bash
# Generar listado completo de huellas (para herramientas de análisis)
codeknit fingerprint /codeknit/es/src --show-all
```

```bash
# Archivo de salida personalizado
codeknit fingerprint /codeknit/es/src -o duplicates.skt
```

## Elección de rango de similitud

| Rango    | Guía                                                                                     |
| -------- | ---------------------------------------------------------------------------------------- |
| 96–100%  | Duplicados estructurales exactos o casi exactos. Casi con certeza copiar y pegar.        |
| 85–95%   | Casi duplicados. Generalmente copiar y pegar con ediciones menores (ej. variables renombradas, logging añadido). |
| 65–84%   | Rango predeterminado. Fuerte similitud estructural. Buenos candidatos para refactorización. |
| 50–64%   | Similitud moderada. Misma forma algorítmica pero con detalles diferentes. Revisar manualmente. |
| < 50%    | Generalmente ruido. No es duplicación significativa.                                     |

## Consejos

- **Las huellas miden estructura, no significado**: Una puntuación alta de similitud significa que el código _se ve_ similar, no que _hace_ lo mismo. Siempre revisa ambos símbolos.
- **Usa `--rerank` para resultados ruidosos**: Si obtienes muchos falsos positivos, habilita el reordenamiento semántico para filtrar coincidencias utilizando embeddings.
- **Se omiten cuerpos cortos**: Los símbolos con menos de 4 tokens normalizados (ej. getters simples) se ignoran para evitar ruido.
- **Funciona la coincidencia entre lenguajes**: Constructores equivalentes (ej. una función en Python y una función en Go con la misma lógica) pueden coincidir, pero los patrones específicos del lenguaje pueden producir coincidencias espurias de baja similitud.
- **Una coincidencia es una señal, no un veredicto**: Trata cada coincidencia como un indicio para investigar — no como prueba automática de duplicación.