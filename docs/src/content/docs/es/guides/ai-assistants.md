---
title: Uso con asistentes de IA
description: Instala el plugin codeknit para Claude Code, Codex y Kiro.
---

Instala el plugin codeknit en Claude Code, Codex o Kiro, manténlo actualizado y pregunta al asistente sobre estructura de código y código duplicado. El plugin contiene una habilidad, `codeknit`. La habilidad indica al asistente cuándo ejecutar `codeknit parse`, `codeknit graph analyze` y `codeknit fingerprint`, y cómo leer la salida `.skt`.

No necesitas instalar el binario `codeknit` tú mismo. La primera vez que se ejecuta la habilidad, ejecuta su script incluido `scripts/install-codeknit.sh`, que realiza lo siguiente:

1. Sale inmediatamente si `codeknit` ya está en tu `PATH` o en `~/.local/bin`.
2. De lo contrario, descarga la última versión de GitHub para tu sistema operativo y arquitectura.
3. Verifica la descarga con el archivo `checksums.txt` de la versión.
4. Instala el binario en `~/.local/bin` e imprime su ruta absoluta en la última línea de salida.

Si `~/.local/bin` no está en tu `PATH`, el asistente llama a codeknit mediante esa ruta completa durante el resto de la sesión.

Para controlar la instalación, establece estas variables de entorno:

| Variable | Efecto |
| --- | --- |
| `CODEKNIT_VERSION` | Fija una versión, por ejemplo `0.6.0`. |
| `CODEKNIT_INSTALL_DIR` | Cambia la ubicación de instalación. |

Para instalar el binario tú mismo, consulta [Instalación](../../installation/).

## Instala el plugin

Elige la sección correspondiente a tu asistente.

### Claude Code

Ejecuta estos comandos en tu shell. Dentro de una sesión de Claude Code, ejecuta los mismos comandos como `/plugin ...`.

```bash
claude plugin marketplace add awslabs/codeknit --sparse .claude-plugin skills
claude plugin install codeknit@codeknit
```

### Codex

```bash
codex plugin marketplace add awslabs/codeknit --sparse .agents --sparse .codex-plugin --sparse skills
codex plugin add codeknit@codeknit
```

### Kiro IDE

1. Abre el panel Powers.
2. Elige **Add Custom Power**, luego **Import power from GitHub**.
3. Ingresa `https://github.com/awslabs/codeknit`.

### Kiro CLI

La CLI de Kiro instala desde una ruta local, así que clona el repositorio primero:

```bash
git clone --depth 1 --filter=blob:none --sparse https://github.com/awslabs/codeknit ~/.codeknit-power
git -C ~/.codeknit-power sparse-checkout set skills
kiro-cli powers install ~/.codeknit-power
```

### Acerca de `--sparse`

La opción `--sparse` es opcional. Limita la clonación a los manifiestos del plugin y `skills/` en lugar de todo el árbol de código fuente.

### Instala la habilidad manualmente

Si tu asistente lee Agent Skills pero no tiene soporte para plugins, copia `skills/codeknit/` en el directorio de habilidades de tu asistente:

- `~/.claude/skills/`
- `~/.codex/skills/`
- `~/.kiro/skills/`

## Actualiza el plugin

| Asistente | Comando o acción |
| --- | --- |
| Claude Code | `claude plugin marketplace update codeknit` |
| Codex | `codex plugin marketplace upgrade` |
| Kiro IDE | Elige **Check for updates** en el panel Powers. |
| Kiro CLI | Ejecuta `git -C ~/.codeknit-power pull`, luego vuelve a ejecutar `kiro-cli powers install ~/.codeknit-power`. |

Actualizar el plugin no actualiza un binario que ya esté instalado. Para cambiar la versión del binario, establece `CODEKNIT_VERSION` como se describió anteriormente.

Si actualizas codeknit y cambia la extracción, los IDs de símbolos también pueden cambiar. Pide al asistente que regenere toda la salida en lugar de mezclar fragmentos de diferentes ejecuciones.

## Lo que la habilidad enseña al asistente

La habilidad proporciona al asistente estas reglas, para que no tengas que repetirlas en cada solicitud:

- **Elige el comando según la pregunta.**
  - `parse` responde qué existe, dónde está y qué llama, hereda o contiene qué.
  - `graph analyze` encuentra ciclos, hubs, código muerto, god classes y cuellos de botella.
  - `fingerprint` encuentra código copiado y pegado y candidatos para refactorización.
- **Usa `directory-flat` como modo de salida predeterminado.** El asistente escribe archivos fragmentados en `./skeleton/` y lee solo los fragmentos que necesita. Usa el modo de salida `inline` solo para un único archivo o un puñado de archivos, porque `inline` vuelca todo en la ventana de contexto.
- **Lee la salida `.skt`, no el código fuente sin procesar, para preguntas estructurales.** El asistente abre archivos de código fuente solo para inspeccionar o cambiar detalles de implementación, utilizando los rangos de líneas de la salida. El archivo `references/output-format.md` de la habilidad contiene el formato completo. Consulta también la [referencia del formato de salida](../../reference/output-format/).
- **Elige rangos de similitud y trata las coincidencias como señales.** Las huellas digitales comparan la forma estructural, no la intención. El asistente lee ambos símbolos de un par antes de proponer una refactorización.

## Solicita un análisis estructural

1. Pide al asistente que analice la estructura de tu base de código. Por ejemplo: "Mapea la estructura de `./src` y dime qué módulos dependen del paquete auth."
2. El asistente se asegura de que el binario esté instalado, luego ejecuta `codeknit parse` con las relaciones habilitadas y lee la salida `.skt`.
3. Haz preguntas de seguimiento sobre dependencias, cadenas de llamadas o código muerto. El asistente responde a partir de la salida.
4. Para un análisis más profundo, solicita una revisión estructural. Por ejemplo: "Encuentra dependencias cíclicas y god classes en `./src`." El asistente ejecuta `codeknit graph analyze` y lee el informe.

Las relaciones marcadas como `[unresolved]` o `[ambiguous]` en la salida son inciertas, no dependencias confirmadas. Trata las respuestas que se basan en ellas como candidatos para verificar.

Para conocer los comandos detrás de estas solicitudes, consulta [la guía del comando parse](../parse-command/) y [la guía de los comandos graph](../graph-commands/).

## Solicita detección de duplicados

1. Pide al asistente que encuentre código duplicado. Por ejemplo: "Encuentra funciones copiadas y pegadas en `./src`."
2. El asistente ejecuta `codeknit fingerprint` y lee la sección `[duplicates]` de `./skeleton/fingerprints.skt`.
3. El asistente abre ambos símbolos de cada par marcado, luego propone una refactorización.
4. Si los resultados son ruidosos, pídele que ajuste el rango de similitud o que agregue un reranking semántico. El reranking requiere un servidor local de Ollama con un modelo de embeddings descargado.

Para conocer las opciones detrás de estas solicitudes, consulta [la guía del comando fingerprint](../fingerprint-command/).

## Consejos

- Recuerda al asistente que lea los archivos `.skt`, no el código fuente sin procesar, para preguntas estructurales. Esto mantiene el contexto pequeño, ya que el formato es compacto y contiene símbolos, relaciones y rangos de líneas.
- Usa `graph analyze` para obtener información sobre la calidad del código. Los hallazgos son candidatos para revisión, porque los llamadores externos y el despacho dinámico pueden cambiar el resultado.
- Ejecuta fingerprint antes de una refactorización grande para detectar código copiado y pegado que deba consolidarse.
- Si el asistente vuelve a ejecutar `parse` en un directorio que ya contiene salida `.skt`, codeknit se niega a escribir. Pídele que pase `--clean` para reemplazar la salida anterior.
- Si la descarga del binario falla (sin conexión o plataforma no compatible), obtén una versión manualmente desde la [página de versiones](https://github.com/awslabs/codeknit/releases), o consulta [Instalación](../../installation/).
- Si una ejecución informa resultados parciales, revisa stderr y la sección `[errors]`. Consulta [fiabilidad del análisis](../../reference/analysis-reliability/) para más detalles.