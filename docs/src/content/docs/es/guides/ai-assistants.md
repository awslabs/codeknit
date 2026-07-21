---
title: Uso con asistentes de IA
description: Configura codeknit como una habilidad para Kiro, Claude Code y otros asistentes de codificación con IA.
---

codeknit incluye habilidades listas para usar que enseñan a los asistentes de codificación con IA cómo utilizarlo de manera efectiva. Estas habilidades permiten a los asistentes extraer la estructura de código, detectar duplicados y realizar análisis estructurales sin necesidad de indicaciones manuales.

## Descripción general de habilidades

codeknit proporciona dos habilidades:

- **`codeknit-parse`**: Enseña a los asistentes a extraer la estructura de código (funciones, clases, métodos, variables) y relaciones (llamadas, herencia, contención) en archivos `.skt`.
- **`codeknit-fingerprint`**: Enseña a los asistentes a detectar código duplicado y casi duplicado utilizando *fuzzy hashing*.

Cada habilidad incluye documentación que el asistente lee bajo demanda para comprender el uso, las *flags*, los formatos de salida y los flujos de trabajo.

## Instalación

Utiliza el ayudante de instalación para copiar los directorios de habilidades a la carpeta de habilidades de tu asistente. El instalador descarga solo los archivos de habilidades incluidos, por lo que no necesitas clonar el repositorio.

Instala para **Codex**, **Kiro** y **Claude Code**:

```bash
curl -fsSL https://raw.githubusercontent.com/awslabs/codeknit/main/scripts/install-skills.sh | bash
```

Instala para un asistente específico:

```bash
curl -fsSL https://raw.githubusercontent.com/awslabs/codeknit/main/scripts/install-skills.sh | bash -s -- --assistant codex
curl -fsSL https://raw.githubusercontent.com/awslabs/codeknit/main/scripts/install-skills.sh | bash -s -- --assistant kiro
curl -fsSL https://raw.githubusercontent.com/awslabs/codeknit/main/scripts/install-skills.sh | bash -s -- --assistant claude
```

Desde una copia local, puedes usar los ayudantes de Makefile:

```bash
make skills-install-dry-run
make skills-install
```

El instalador omite los directorios de habilidades existentes de forma predeterminada. Para reemplazarlos, añade `--force`:

```bash
curl -fsSL https://raw.githubusercontent.com/awslabs/codeknit/main/scripts/install-skills.sh | bash -s -- --assistant all --force
```

Después de la instalación, el asistente sabrá automáticamente cómo invocar los comandos de codeknit, seleccionar las *flags* apropiadas e interpretar la salida `.skt`.

## Qué enseña cada habilidad

### codeknit-parse

La habilidad `codeknit-parse` enseña a los asistentes a:

- Ejecutar `codeknit parse` con las *flags* apropiadas para diferentes escenarios
- Elegir el modo de salida correcto:
  - `directory-flat` (predeterminado) para la mayoría de los proyectos
  - `inline` para archivos individuales o entradas pequeñas
  - `directory-tree` para reflejar la estructura del código fuente
- Leer e interpretar archivos de salida `.skt`, incluyendo las secciones `[symbols]`, `[edges]` y opcionalmente `[dict]`
- Utilizar datos estructurales para refactorización, mapeo de dependencias y revisión de código
- Ejecutar `codeknit graph analyze` para obtener información más profunda sobre la calidad del código (dependencias cíclicas, símbolos *hub*, *god classes*, etc.)

### codeknit-fingerprint

La habilidad `codeknit-fingerprint` enseña a los asistentes a:

- Usar `codeknit fingerprint` para la detección de duplicados, auditorías DRY e identificación de refactorizaciones
- Seleccionar rangos de similitud apropiados (`--min-similarity`, `--max-similarity`)
- Leer la sección `[duplicates]` para identificar código casi duplicado
- Entender que los *fingerprints* miden la forma estructural, no la intención semántica
- Usar `--rerank` con incrustaciones de Ollama para reducir falsos positivos cuando sea necesario

## Ejemplos de flujos de trabajo

### Análisis estructural

1. Pide al asistente que analice la estructura de tu base de código
2. Ejecuta `codeknit parse ./src` y lee los archivos `.skt` resultantes
3. Responde preguntas estructurales: dependencias, cadenas de llamadas, código muerto
4. Para obtener información más profunda, ejecuta `codeknit graph analyze ./src` e interpreta el informe

```skt
[symbols]
## src/service.go
S1 type/struct L5-L8 AuthService {}
S2 callable/method L10-L15 Authenticate(token: string) {receiver=*AuthService}

[edges]
S1 --contains--> S2
```

### Detección de duplicados

1. Pide al asistente que encuentre código duplicado
2. Ejecuta `codeknit fingerprint ./src`
3. Lee la sección `[duplicates]` en la salida
4. Investiga los pares marcados y propone consolidaciones

```skt
[duplicates]
S1, S2: 87% similitud
S3, S4: 76% similitud
```

## Consejos

- **Siempre lee los archivos `.skt`, no el código fuente sin procesar, para preguntas estructurales** — contienen la estructura extraída en un formato compacto y confiable
- Usa `codeknit graph analyze` para descubrir problemas de calidad de código como dependencias cíclicas, símbolos *hub* y cadenas de herencia profundas
- Ejecuta `codeknit fingerprint` antes de grandes refactorizaciones para identificar código copiado y pegado que debería consolidarse
- El formato `.skt` está diseñado para ser eficiente en tokens, lo que lo hace ideal para ventanas de contexto de LLM
- Usa `--minify` para reducir aún más el uso de tokens al procesar bases de código grandes