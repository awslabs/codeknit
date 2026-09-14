---
title: Soporte de Lenguajes y Fidelidad de la API
description: Capacidades de extracción probadas, selección de analizadores y limitaciones conocidas de los lenguajes.
---

<!-- Maintained in docs/reference/language-support.md; copied verbatim by docs/gen. -->

Codeknit extrae sintaxis sin ejecutar un compilador, evaluar macros o cargar las dependencias de un proyecto. La matriz a continuación describe construcciones probadas, no la conformidad completa del lenguaje o la resolución de nombres garantizada entre archivos.

Cada lenguaje tiene una afirmación exacta de parámetros básicos y firma de retorno en `internal/plugin/api_fidelity_test.go`. La cobertura de declaraciones también se verifica en `internal/plugin/coverage_test.go` y en los paquetes de prueba individuales de cada lenguaje.

| Lenguaje   | Extensiones                                  | Línea base de firmas                                                                                           | Cobertura de declaraciones y contenedores                                                                         |
| ---------- | ------------------------------------------- | ------------------------------------------------------------------------------------------------------------ | ---------------------------------------------------------------------------------------------------------- |
| C          | `.c`, `.h`                                  | Parámetros nombrados y tipos de retorno en definiciones y prototipos                                              | Structs, unions, enums, typedefs, macros; declaraciones dentro de guards de encabezado                                |
| C++        | `.cpp`, `.cc`, `.cxx`, `.hpp`, `.hxx`, `.h` | Parámetros nombrados y tipos de retorno en funciones, prototipos y declaraciones de métodos                          | Clases, structs, métodos, namespaces, templates; guards de encabezado y wrappers `extern "C"`                  |
| C#         | `.cs`                                       | Parámetros y tipos de retorno nombrados y explícitos                                                             | Clases, structs, interfaces, métodos, campos, delegados, namespaces                                       |
| Go         | `.go`                                       | Parámetros agrupados, sin nombre, variádicos, tipados como función y genéricos; resultados nombrados/múltiples; receptores genéricos | Campos de struct con tipos y tags, campos embebidos, firmas de métodos de interfaz, parámetros de tipo, paquetes |
| Java       | `.java`                                     | Parámetros y tipos de retorno nombrados y explícitos                                                             | Clases, interfaces, constructores, métodos, campos, clases anidadas, paquetes                               |
| JavaScript | `.js`, `.jsx`, `.mjs`, `.cjs`               | Parámetros nombrados; sin tipo de retorno inferido                                                                    | Funciones, arrow functions, clases, campos, métodos, declaraciones que contienen JSX                          |
| PHP        | `.php`                                      | Parámetros y tipos de retorno nombrados y explícitos                                                             | Clases, interfaces, traits, enums, métodos, propiedades, namespaces                                        |
| Python     | `.py`, `.pyi`                               | Anotaciones, valores predeterminados, `*args`, `**kwargs`, separadores posicionales/palabra clave, anotaciones de retorno                | Funciones, clases, clases anidadas, métodos decorados; receptores bound convencionales omitidos                |
| Ruby       | `.rb`                                       | Parámetros nombrados; sin tipo de retorno inferido                                                                    | Clases, módulos, métodos de instancia/singleton, constantes                                                    |
| Rust       | `.rs`                                       | Patrones tipados, `self`/`&self`/`&mut self`, bindings mutables, tipos de retorno explícitos                          | Módulos inline y anidados, campos de struct nombrados, métodos de traits y métodos de implementación                   |
| Scala      | `.scala`, `.sc`                             | Parámetros y tipos de retorno nombrados y explícitos                                                             | Clases, traits, objetos, métodos, enums, paquetes                                                         |
| TypeScript | `.ts`, `.tsx`, `.mts`, `.cts`               | Parámetros tipados nombrados y anotaciones de retorno explícitas                                                       | Funciones, arrow functions, interfaces, clases, campos, namespaces; TSX usa su propia gramática              |

## Selección de lenguaje para encabezados

Todos los comandos CLI de procesamiento de código fuente aceptan `--header-language c|cpp`. La elección se aplica solo a los archivos `.h`. Otras extensiones conservan sus analizadores habituales.

```bash
codeknit parse ./src --header-language cpp
codeknit fingerprint ./src --header-language c
codeknit graph analyze ./src --header-language cpp
```

El valor predeterminado es `c`, preservando la extracción de macros, typedefs y unions de C. Use `cpp` para encabezados C++ nombrados `.h`; los archivos nombrados `.hpp` y `.hxx` ya usan C++. La elección es explícita porque ambas gramáticas pueden aceptar parte de la sintaxis del otro lenguaje sin extraer su estructura completa.

Los guards de encabezado y las ramas condicionales se recorren sin evaluar sus condiciones. Los resultados pueden incluir, por lo tanto, declaraciones de ramas mutuamente excluyentes. La expansión de macros, la resolución de includes y las flags de compilación no se aplican. La TUI usa la selección predeterminada de C para los archivos `.h`.

## Preservación de la forma de la API

Por ejemplo, en Go:

```go
func Pair(a, b int, rest ...string) (string, error)
```

produce:

```text
Pair(a: int, b: int, rest: ...string) -> (string, error)
```

Los métodos de interfaz de Go y los campos nombrados de Go/Rust aparecen como símbolos tanto en el modo normal como en el de fingerprinting. Los métodos y campos conservan su tipo propietario. Los módulos anidados de Rust conservan nombres calificados como `service.storage.read`, incluso cuando otro módulo declara una función con el mismo nombre.

Rust permite campos y métodos con el mismo nombre. Los nombres de campos con ámbito llevan un sufijo `#field` internamente para que sus IDs y relaciones de contención permanezcan distintos; los nombres de visualización y las firmas conservan el nombre original del campo.

Las firmas de Python conservan anotaciones, valores predeterminados, splats y separadores de convención de llamada. Un parámetro convencional `self` o `cls` se omite en un método bound, pero se conserva en una función libre o método estático.

Las firmas multilínea conservan los saltos de línea del código fuente en JSON. SKT escapa los saltos de línea como `\n` y `\r` para que cada símbolo permanezca en una sola línea física de salida.

## Resolución de relaciones

Las relaciones conservan la calificación del receptor y la propiedad léxica. Las importaciones relativas y las importaciones renombradas de JavaScript/TypeScript y Python proporcionan contexto de destino. Las declaraciones de paquetes de Go y Java pueden resolverse entre archivos en el mismo directorio. Una importación no coincidente nunca retrocede a una declaración no relacionada, y los candidatos competidores permanecen ambiguos.

Las vinculaciones de receptores locales tienen prioridad sobre las importaciones de namespaces, incluso cuando un parámetro oculta un namespace importado. Las importaciones de paquetes de Python consideran las declaraciones en el módulo `__init__` del paquete, no en módulos descendientes arbitrarios. Las reexportaciones que no pueden establecerse desde ese módulo permanecen sin resolver.

La extracción de callable/function, callable/method y callable/constructor en Java, C++, C#, Scala y TypeScript conserva los rangos de bytes de las declaraciones para que los cuerpos sobrecargados puedan tener llamadores y ámbitos de alias distintos, incluyendo declaraciones en la misma línea. La selección de un destino sobrecargado sigue siendo conservadora. Las relaciones de sobreescritura explícita en Java, C++, C#, Scala y TypeScript buscan la declaración heredada más cercana; las bases faltantes, ciclos o candidatos competidores conservan la incertidumbre en lugar de producir enlaces al propio método.

Este es un análisis basado en sintaxis, no un compilador o grafo de llamadas en tiempo de ejecución. Los especificadores de paquetes JavaScript simples, las reexportaciones de paquetes, las configuraciones de compilación, los tipos de receptores complejos y el despacho dinámico pueden permanecer sin resolver. Un nombre único en otro lugar del repositorio no es evidencia suficiente de una dependencia.

JSON conserva las relaciones inciertas con `resolution: "unresolved"` o `"ambiguous"` e incluye los IDs de los candidatos cuando están disponibles. Estas relaciones no tienen un short ID de destino.

SKT usa la misma notación de relación para todas las relaciones, con un marcador de resolución opcional:

```text
S8 --references--> S5
S10 --references[unresolved]--> string, bool
S11 --calls[ambiguous]--> Helper [candidates=S12, S15]
```

Los puntos finales conocidos y los candidatos usan short IDs, incluyendo los puntos finales de relaciones no resueltas. Los destinos inciertos sin símbolo son nombres relativos al archivo fuente. Los nombres literales que parecen short IDs, y otras identidades complejas, usan IDs completos entre comillas para evitar ambigüedades. Las relaciones con la misma fuente, tipo, resolución y conjunto de candidatos comparten una lista de destinos. El lector de SKT conserva la resolución y los candidatos, incluyendo referencias a símbolos en chunks de salida posteriores. La salida minificada codifica el tipo de relación mientras mantiene legible el marcador de resolución.

Los includes de C/C++ tienen símbolos explícitos `meta/file` y `meta/include`. Las grafías de los encabezados aparecen una vez en la lista de símbolos; las relaciones usan sus short IDs:

```text
[symbols]
## src/main.c
S1 meta/file L1-L4 main.c
S2 meta/include L1-L1 "config.h" {resolution=unresolved}
S3 meta/include L2-L2 "utils.h" {resolution=unresolved}
S4 meta/include L3-L3 <stdio.h> {resolution=unresolved}

[edges]
S1 --references[unresolved]--> S2, S3, S4
```

Un símbolo de include describe la directiva escrita, no una declaración de encabezado resuelta. Los includes repetidos dentro de un archivo comparten un punto final. Los metadatos de archivo/include están excluidos de las clasificaciones de dependencias y los hallazgos de dead code; agregar IDs no afirma que se haya encontrado un encabezado a través de la configuración de búsqueda de includes del compilador.

Las métricas de dependencia y los enlaces del grafo HTML usan solo relaciones resueltas. Los informes muestran recuentos de relaciones excluidas, y el grafo HTML muestra un recuento de incertidumbre. Fan-in y fan-out cuentan vecinos distintos. El análisis de herencia considera todas las bases e interfaces; la herencia cíclica no tiene una profundidad reportada. La alcanzabilidad comienza desde callable/function o callable/method nombrados `main`, `Main` o `init` y sigue las relaciones de dependencia. Alcanzar un miembro hace relevante su contenedor sin hacer alcanzables a todos sus hermanos. Sin puntos de entrada reconocidos, los hallazgos de alcanzabilidad se omiten.

Los argumentos de callback y los callable/function devueltos son referencias, no prueba de invocación. La resolución de alias conserva la identidad del archivo, función y objeto; las asignaciones conflictivas permanecen inciertas. Los hallazgos del grafo son candidatos para revisión: los llamadores externos y el comportamiento dinámico pueden cambiar el resultado. Las reglas de capa basadas en directorios y los pesos de propagación de cambios son heurísticas, no probabilidades medidas.

## Límites conocidos

- Los declaradores complejos de punteros, referencias, arrays y punteros a función en C/C++ pueden tener firmas incompletas. Las declaraciones generadas por macros no se expanden.
- El manejo de constructores de TypeScript/JavaScript y algunas formas de parámetros opcionales, rest o desestructurados están incompletos.
- Las restricciones avanzadas de genéricos y las formas de tipo fuera de los casos probados varían según el lenguaje. Los tipos de retorno no se infieren.
- Las macros de Rust no se expanden. Las declaraciones de módulos fuera de línea se registran; sus archivos deben incluirse en el escaneo. Los campos de tuplas, los detalles de variantes de enums y los elementos asociados tienen cobertura parcial.
- La extracción de sintaxis no resuelve el despacho dinámico, las sobrecargas, las dependencias importadas o la semántica de configuración de compilación con la precisión de un compilador.
- Los errores de sintaxis parciales aún pueden omitir declaraciones. Inspeccione las advertencias reportadas y verifique el código fuente antes de aplicar una refactorización.

Una extracción más completa cambia los recuentos de símbolos y los short IDs. Regenera la salida de skeleton existente al actualizar en lugar de mezclarla con salida más antigua.