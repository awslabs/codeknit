---
title: Modos de salida
description: Elige el modo de salida adecuado para el tamaño de tu proyecto y flujo de trabajo.
---

codeknit separa dos decisiones: dónde va la salida (el **modo de salida**) y cómo se codifica (el **formato de salida**). Esta página explica por qué existe cada opción, qué coste tiene y cuándo es adecuada. Para la gramática `.skt`, consulta la [referencia del formato de salida](../../reference/output-format/). Para los comandos, consulta la [guía del comando parse](../parse-command/).

## Dos elecciones independientes

La bandera `--output-mode` selecciona uno de tres modos: `directory-flat`, `directory-tree` o `inline`. La bandera `--format` selecciona `skt` (el valor predeterminado) o `json`. Cualquier modo puede combinarse con cualquier formato.

Las dos elecciones son independientes porque responden a preguntas diferentes. El modo decide cómo se divide y localiza la salida en el disco, lo que importa a una persona o a un LLM que la lea. El formato decide quién la consume: un LLM o un lector humano, o un script.

## directory-flat

`directory-flat` es el valor predeterminado. Escribe archivos de fragmentos numerados, `map_001.skt`, `map_002.skt`, y así sucesivamente, en `./skeleton/` a menos que proporciones otro directorio de salida.

**Por qué existe.** Un código base grande produce más estructura de la que cabe en una ventana de contexto. La división en fragmentos permite a un lector abrir solo los archivos que necesita en lugar de cargar todo.

**Cómo funciona la división.** codeknit mantiene intacto el bloque de cada archivo fuente y comienza un nuevo fragmento entre bloques una vez que un fragmento alcanza aproximadamente `--max-lines` (predeterminado 500). El límite es un objetivo, no un límite estricto: un archivo fuente con un bloque muy grande puede producir un fragmento más largo que 500 líneas. Mantener los bloques intactos significa que un símbolo nunca se separa de sus hermanos en el mismo archivo.

**Compensación.** Los fragmentos no están organizados por ruta de origen. Para encontrar la estructura de un archivo, buscas en los fragmentos su encabezado `## source-path`. Si a menudo buscas archivos específicos, `directory-tree` se ajusta mejor.

**Minificación.** Con `--minify`, el diccionario se escribe en `dict.skt` junto a los fragmentos. Consulta [Minificación](#minificación).

## directory-tree

`directory-tree` refleja tu árbol de origen bajo el directorio de salida. Cada archivo fuente obtiene un archivo `.skt` en la ruta correspondiente.

**Nomenclatura.** codeknit añade `.skt` al nombre completo del archivo fuente en lugar de reemplazar la extensión. `src/foo.c` se convierte en `src/foo.c.skt`, y `src/foo.h` se convierte en `src/foo.h.skt`. Sin esta regla, ambos se asignarían a `foo.skt` y colisionarían. Cuando la salida de un archivo se divide, las partes se nombran `foo.c_part1.skt`, `foo.c_part2.skt`, y así sucesivamente. Las rutas de salida en conflicto, incluidas las colisiones solo por mayúsculas, fallan visiblemente en lugar de sobrescribirse entre sí.

**La división difiere del modo flat.** En el modo tree, `--max-lines` es un límite estricto en cada fragmento de origen, con un mínimo de 3. Cada fragmento de símbolo repite los encabezados `[symbols]` y `## source-path` para que pueda leerse solo. Los fragmentos solo de relaciones llevan `[edges]`. El límite se aplica a los fragmentos de origen, no a archivos auxiliares como `dict.skt` y `warnings.skt`.

**Compensación.** La salida en árbol es fácil de navegar junto al origen, pero produce muchos archivos pequeños. Un proyecto grande puede generar miles de ellos, y un lector que quiera una visión general amplia debe abrir muchos archivos.

## inline

`inline` escribe todo en stdout. No crea un directorio de salida ni un área de preparación.

**Por qué existe.** Es adecuado para dos casos: inspeccionar un solo archivo rápidamente y canalizar la salida a otra herramienta. Es el único modo en el que no se escribe nada en el disco.

**Compensación.** Nada se divide. Todo el resultado llega al terminal o al contexto del consumidor de una vez, por lo que abruma a un lector en cualquier cosa más allá de un puñado de archivos. Trátalo como adecuado para un solo archivo o un proyecto muy pequeño, aproximadamente menos de cinco archivos. Debido a que stdout no puede retrocederse, un consumidor fallido puede ver una salida parcial. Los modos de directorio no tienen este problema.

## Formatos de salida

| Formato | Mejor para | Salida |
| --- | --- | --- |
| `skt` | Contexto de LLM e inspección humana | Archivos `.skt` o stdout |
| `json` | Scripts e integración estructurada | `codeknit.json` en modos de directorio, o stdout en modo inline |

**skt** es compacto y orientado a líneas. Cada símbolo ocupa una línea física, con saltos de línea en las firmas escapados, por lo que un lector puede cargar parte de la salida sin romper su estructura. Por eso la división en fragmentos funciona para `.skt`.

**json** existe para consumidores que necesitan campos tipados en lugar de texto para leer. Es un solo documento con `files`, `symbols` y, cuando se incluyen relaciones, `edges`. Debido a que es un solo documento, no se divide por conteo de líneas. En los modos de directorio, `--max-lines` por lo tanto no tiene efecto en JSON, y un proyecto grande produce un solo `codeknit.json` grande.

Un ejemplo compacto:

```json
{
  "files": ["app.go"],
  "symbols": [
    {"id": "app.go::User", "short_id": "S1", "name": "User", "file": "app.go",
     "category": "type", "kind": "struct", "signature": "type User struct", "span": [3, 3]},
    {"id": "app.go::Save", "short_id": "S2", "name": "Save", "file": "app.go",
     "category": "callable", "kind": "function", "signature": "Save(u: S1)", "span": [5, 5]}
  ],
  "edges": [
    {"from": "app.go::Save", "from_short": "S2", "to": "app.go::User",
     "to_short": "S1", "kind": "references"}
  ]
}
```

## Elegir un modo

| Modo | Mejor para | Salida |
| --- | --- | --- |
| `directory-flat` | La mayoría de proyectos (predeterminado) | `./skeleton/map_001.skt`, `map_002.skt`, ... |
| `directory-tree` | Navegar junto al origen | `./skeleton/<ruta reflejada>.skt` |
| `inline` | Archivo único, canalización | stdout |

Reglas generales:

- **Si no estás seguro**, usa `directory-flat`, el valor predeterminado.
- **Un solo archivo**: `inline` está bien.
- **Cualquier cosa más allá de unos pocos archivos**: usa `directory-flat` o `directory-tree`.
- **Códigos base grandes**: añade `--minify`.
- **Volver a ejecutar en el mismo directorio**: añade `--clean`.

## Preparación, re-ejecuciones y `--clean`

Los modos de directorio no escriben en el directorio de salida a medida que avanzan. codeknit genera la salida en un directorio de preparación hermano y la publica solo después de que la generación tenga éxito. Una ejecución fallida deja tu salida anterior en su lugar.

De forma predeterminada, codeknit se niega a escribir cuando ya existe salida generada para el formato seleccionado. La razón es que los IDs de símbolos (`S1`, `S2`, ...) pueden cambiar cada vez que cambia la extracción. Mezclar fragmentos de diferentes ejecuciones daría relaciones que apuntan a símbolos incorrectos. `--clean` opta por el reemplazo: después de una generación exitosa, los archivos `.skt` obsoletos (en modo SKT) o `codeknit.json` (en modo JSON) son reemplazados. Las entradas no relacionadas en el directorio se conservan. Si una de ellas colisiona con una ruta generada, la ejecución falla en lugar de sobrescribirla.

La publicación toma un bloqueo específico del directorio, hace una copia de seguridad de la salida anterior y la restaura si el reemplazo falla. Si el proceso se interrumpe entre los dos cambios de nombre de directorio, la siguiente ejecución restaura la copia de seguridad antes de generar nada. Esto se recupera de fallos del proceso. No hace que la actualización del directorio sea instantánea: hay un breve intervalo entre los cambios de nombre, por lo que un lector que necesite acceso ininterrumpido debe coordinarse con la generación. Tampoco protege contra fallos de almacenamiento o cortes de energía.

No elijas un directorio fuente, o un ancestro de él, como directorio de salida. Para las garantías completas de publicación, consulta la [referencia de fiabilidad del análisis](../../reference/analysis-reliability/).

## Minificación

`--minify` aplica compresión basada en diccionario. codeknit reemplaza tokens repetidos, como `callable/method` o `exported`, con códigos cortos (`d0`, `d1`, ...). La asignación se almacena en una sección `[dict]`, que en los modos de directorio es el archivo `dict.skt`.

El ahorro proviene de la repetición. La mayoría de las líneas de símbolos comparten las mismas categorías, tipos y propiedades, por lo que un código base grande repite estas cadenas miles de veces. El coste es la indirección: un lector debe expandir los códigos usando el diccionario, y un solo fragmento ya no es autónomo. El diccionario es una salida auxiliar separada, por lo que no cuenta para `--max-lines`.

La salida minificada codifica el tipo de relación pero mantiene legibles los marcadores de resolución, por lo que las relaciones `unresolved` y `ambiguous` siguen siendo distinguibles. Para proyectos pequeños, los ahorros son modestos y el paso adicional rara vez vale la pena. Para proyectos grandes, los ahorros en tokens pueden ser significativos.

## Páginas relacionadas

- [Guía del comando parse](../parse-command/): los comandos para cada modo.
- [Referencia del formato de salida](../../reference/output-format/): la gramática `.skt` y JSON.
- [Referencia de banderas CLI](../../reference/cli-flags/): cada bandera y valor predeterminado.
- [Referencia de fiabilidad del análisis](../../reference/analysis-reliability/): nomenclatura de salida y garantías de publicación.