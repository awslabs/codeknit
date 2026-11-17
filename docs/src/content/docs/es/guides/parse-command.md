---
title: Comando Parse
description: Extraer información estructural del código fuente en archivos .skt o JSON.
---

`codeknit parse` lee archivos de código fuente y escribe sus símbolos (y, opcionalmente, las relaciones entre ellos) como archivos `.skt` o JSON. Esta guía cubre las situaciones con las que es más probable que se encuentre. Para ver la lista completa de flags compartidos por todos los comandos, consulte la [referencia de flags de CLI](../../reference/cli-flags/). Para ver la gramática de salida, consulte la [referencia del formato de salida](../../reference/output-format/).

## Analizar un proyecto por primera vez

```bash
codeknit parse ./src
```

codeknit escribe archivos `.skt` divididos en `./skeleton/`: `map_001.skt`, `map_002.skt`, y así sucesivamente. Abra los fragmentos de forma selectiva en lugar de cargar todo de una vez.

Para escribir en otro lugar, pase el directorio de salida como segundo argumento:

```bash
codeknit parse ./src ./output
```

No utilice el directorio fuente, ni uno de sus ancestros, como directorio de salida.

## Elegir un modo de salida

Seleccione un modo con `--output-mode`.

| Modo             | Úselo cuando                                    | Resultado                                                                         |
| ---------------- | ---------------------------------------------- | ------------------------------------------------------------------------------ |
| `directory-flat` | Analizar la mayoría de proyectos (predeterminado) | `map_001.skt`, `map_002.skt`, ...                                              |
| `directory-tree` | Quiere buscar la estructura de un archivo       | Un `.skt` por archivo fuente, reflejando el árbol fuente                          |
| `inline`         | Analizar un solo archivo o canalizar a otra herramienta | Salida en stdout; no se utiliza un directorio de salida                                  |

En el modo árbol, los nombres de archivo conservan la extensión fuente: `src/foo.c` se convierte en `src/foo.c.skt`, y `src/foo.h` se convierte en `src/foo.h.skt`. Los archivos grandes se dividen en `foo.c_part1.skt`, `foo.c_part2.skt`, y así sucesivamente.

```bash
codeknit parse ./src --output-mode directory-tree
```

Evite `inline` en un proyecto completo. Escribe todo en stdout y puede saturar su terminal o la ventana de contexto de un LLM.

Para más información sobre cómo elegir un modo, consulte [Modos de salida](../output-modes/).

## Volver a ejecutar en el mismo proyecto

codeknit se niega a escribir en un directorio que ya contenga salida generada para el formato seleccionado. Esto evita que se mezclen fragmentos obsoletos y nuevos. Agregue `--clean` para reemplazar la salida anterior:

```bash
codeknit parse ./src --clean
```

`--clean` no elimina nada de antemano. codeknit genera la nueva salida en un directorio de staging y reemplaza los archivos generados antiguos solo después de que la generación tenga éxito. Si la generación falla, o el proceso se interrumpe, se mantiene o restaura la salida anterior en la siguiente ejecución. Los archivos en el directorio de salida que codeknit no generó se conservan. Si uno de ellos colisiona con una ruta generada, la ejecución falla.

En el modo JSON, se aplican las mismas reglas a `codeknit.json`.

**Peligro: fragmentos mezclados.** Los IDs de símbolos (`S1`, `S2`, ...) y los resultados de relaciones pueden cambiar entre ejecuciones y entre versiones. Después de actualizar a 0.5.0, regenere todo el conjunto de salida con `--clean`. Nunca combine fragmentos de diferentes ejecuciones.

## Analizar un solo archivo

Apunte `parse` al archivo y use el modo `inline`:

```bash
codeknit parse ./src/main.go --output-mode inline
```

Cuando selecciona explícitamente un archivo fuente regular, codeknit omite el filtrado de archivos ignorados y archivos de prueba para ese archivo. Los patrones de inclusión/exclusión y el límite de tamaño aún se aplican.

## Obtener JSON para scripts

```bash
codeknit parse ./src --output-mode inline --format json --edges
```

El documento JSON tiene los campos de nivel superior `files`, `symbols`, `edges` (omitido cuando no hay ninguno), y `errors` (omitido cuando no hay ninguno). Un pequeño ejemplo:

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

Para escribir el JSON en disco en su lugar, elimine `--output-mode inline`. En los modos de directorio, el JSON va a `codeknit.json`. JSON es un documento único y no se divide por conteo de líneas.

## Incluir relaciones

De forma predeterminada, la salida solo tiene la sección `[symbols]`, lo que ahorra tokens. Agregue `--edges` cuando necesite saber qué llama, contiene, hereda o hace referencia a qué:

```bash
codeknit parse ./src --edges
```

### Leer relaciones inciertas

Codeknit analiza la sintaxis sin un compilador, por lo que algunas relaciones no pueden establecerse. Cada relación indica su grado de certeza:

```skt
S8 --references--> S5
S10 --references[unresolved]--> string, bool
S11 --calls[ambiguous]--> Helper [candidates=S12, S15]
```

- `--kind-->` es una relación resuelta.
- `--kind[unresolved]-->` significa que codeknit no pudo establecer el destino.
- `--kind[ambiguous]-->` significa que más de un destino sigue siendo posible. Se enumeran los IDs de candidatos cuando están disponibles.

Los candidatos son posibles destinos, no dependencias confirmadas. Los puntos finales conocidos usan ShortID. Un destino no resuelto sin símbolo coincidente puede aparecer como un nombre, relativo al archivo fuente.

En JSON, las relaciones inciertas llevan `resolution` (`"unresolved"` o `"ambiguous"`) y, cuando están disponibles, `candidates` como IDs de símbolos completos. Omiten `to_short`.

Para C y C++, las directivas `#include` se convierten en símbolos `meta/file` y `meta/include`, por lo que sus relaciones tienen ShortID. Esto registra la directiva tal como está escrita. No significa que codeknit haya encontrado el encabezado.

Las métricas de grafo excluyen las relaciones inciertas e informan cuántas fueron excluidas. El visor de grafos HTML muestra un conteo de relaciones inciertas y no las dibuja. Para inspeccionar relaciones inciertas individuales, use `--edges` en `parse`. Consulte [Comandos de grafo](../graph-commands/) para esos informes.

La gramática de relaciones está en la [referencia del formato de salida](../../reference/output-format/). Las reglas y límites de resolución están en la [referencia de soporte de lenguajes](../../reference/language-support/#relationship-resolution).

## Reducir el tamaño de la salida

Para comprimir tokens repetidos, agregue `--minify`:

```bash
codeknit parse ./src --minify
```

Esto escribe un diccionario en `dict.skt` (o una sección `[dict]` en línea). Reemplace códigos como `d0` y `d1` con sus valores de diccionario al leer. Los tipos de relaciones también están codificados, pero las etiquetas de resolución como `[unresolved]` permanecen legibles.

Para controlar el tamaño del fragmento, use `--max-lines` (predeterminado `500`):

- En el modo `directory-flat`, es un objetivo. Un bloque fuente nunca se divide, por lo que un fragmento puede excederlo en un bloque.
- En el modo `directory-tree`, es un límite estricto para cada fragmento fuente. El mínimo es `3`. Cada fragmento repite sus encabezados `[symbols]` y `## source-path`.

## Analizar encabezados C++ nombrados `.h`

Codeknit analiza los archivos `.h` como C de forma predeterminada. Para los encabezados C++ con esa extensión, establezca `--header-language cpp`:

```bash
codeknit parse ./src --header-language cpp
```

El flag afecta solo a los archivos `.h`. Los archivos nombrados `.hpp` y `.hxx` ya se analizan como C++, y otras extensiones mantienen sus analizadores habituales.

## Controlar qué archivos se escanean

Los cinco comandos de procesamiento de código fuente aceptan estos flags. Están listados en su totalidad en la [referencia de flags de CLI](../../reference/cli-flags/).

```bash
codeknit parse ./src --include '**/*.go' --exclude '**/generated/' --strict
```

- `--include GLOB` y `--exclude GLOB` son patrones repetibles al estilo gitignore relativos al directorio de entrada. Escríbalos entre comillas para que su shell no los expanda. Los excludes tienen prioridad. Los includes no anulan `.gitignore`, el filtrado de archivos de prueba o las exclusiones integradas.
- `--max-file-size BYTES` omite archivos más grandes. El valor predeterminado es `10485760` (10 MiB). `0` desactiva el límite.
- `--strict` falla antes de escribir cualquier salida si el escaneo o el análisis están incompletos. El valor predeterminado es `false`.

Codeknit omite y reporta enlaces simbólicos y archivos no regulares. Las reglas `.gitignore` de los ancestros se aplican independientemente de su directorio de trabajo.

**Peligro: resultados parciales.** Las rutas no legibles, archivos demasiado grandes, enlaces simbólicos omitidos y advertencias de sintaxis hacen que un análisis esté incompleto. Sin `--strict`, codeknit mantiene los resultados utilizables y reporta los problemas:

- en stderr,
- en el campo `errors` de JSON, o
- en la sección SKT `[errors]` (en línea, o en `warnings.skt` en los modos de directorio).

Una entrada sin fuentes admitidas seleccionadas, o una ejecución en la que falla la lectura de cada fuente, siempre termina sin éxito. Para garantías exactas, consulte [Fiabilidad del análisis](../../reference/analysis-reliability/).

## Flags de parseo

Estos flags son específicos de `parse`. Los flags compartidos están en la [referencia de flags de CLI](../../reference/cli-flags/).

| Flag               | Predeterminado          | Descripción                                                                                       |
| ------------------ | ---------------- | ------------------------------------------------------------------------------------------------- |
| `--output-mode`    | `directory-flat` | Modo de salida: `inline`, `directory-flat`, o `directory-tree`.                                     |
| `--format`         | `skt`            | Formato de salida: `skt` o `json`.                                                                   |
| `--header-language`| `c`              | Analizador para archivos `.h`: `c` o `cpp`. Otras extensiones no se modifican.                              |
| `--max-lines`      | `500`            | Objetivo de líneas para salida plana; límite estricto (mínimo 3) para cada fragmento fuente en árbol.                   |
| `--collect-test`   | `false`          | Incluir archivos de prueba.                                                                               |
| `--minify`         | `false`          | Usar compresión basada en diccionario.                                                                 |
| `--edges`          | `false`          | Incluir la sección `[edges]` (desactivada de forma predeterminada para ahorrar tokens).                                    |
| `--clean`          | `false`          | Reemplazar la salida generada previamente, solo después de que la generación tenga éxito.                                   |
| `--workers`        | `NumCPU`         | Máximo de goroutines de análisis concurrentes.                                                            |
| `--verbose`        | `false`          | Mostrar información de progreso.                                                                       |

## Consejos

- Use `directory-flat` para la mayoría de los proyectos y lea solo los fragmentos que necesite.
- Use `--minify` en bases de código grandes para ahorrar tokens.
- Pase `--edges` solo cuando necesite datos de relaciones.
- Revise la [matriz de capacidades probadas y límites conocidos](../../reference/language-support/) antes de confiar en la extracción para un lenguaje específico.

## Véase también

- [Modos de salida](../output-modes/)
- [Referencia del formato de salida](../../reference/output-format/)
- [Fiabilidad del análisis](../../reference/analysis-reliability/)
- [Comandos de grafo](../graph-commands/)