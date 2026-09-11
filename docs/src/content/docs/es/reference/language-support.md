---
title: Soporte de Lenguajes y Fidelidad de la API
description: Capacidades de extracción probadas, selección de analizadores y limitaciones conocidas de los lenguajes.
---

<!-- Maintained in docs/reference/language-support.md; copied verbatim by docs/gen. -->

Codeknit extrae sintaxis sin ejecutar un compilador, evaluar macros o cargar las dependencias de un proyecto. La matriz a continuación describe constructos probados, no la conformidad completa del lenguaje ni la resolución garantizada de nombres entre archivos.

Cada lenguaje tiene una afirmación exacta de parámetros básicos y firma de retorno en `internal/plugin/api_fidelity_test.go`. La cobertura de declaraciones también se verifica en `internal/plugin/coverage_test.go` y en los paquetes de prueba individuales de cada lenguaje.

| Lenguaje   | Extensiones                                 | Línea base de firmas                                                                                                                 | Cobertura de declaraciones y contenedores                                                                        |
| ---------- | ------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------ | ---------------------------------------------------------------------------------------------------------------- |
| C          | `.c`, `.h`                                  | Parámetros nombrados y tipos de retorno en definiciones y prototipos                                                                 | Structs, unions, enums, typedefs, macros; declaraciones dentro de guards de encabezado                           |
| C++        | `.cpp`, `.cc`, `.cxx`, `.hpp`, `.hxx`, `.h` | Parámetros nombrados y tipos de retorno en funciones, prototipos y declaraciones de métodos                                          | Clases, structs, métodos, namespaces, templates; guards de encabezado y wrappers `extern "C"`                    |
| C#         | `.cs`                                       | Parámetros tipados nombrados y tipos de retorno explícitos                                                                           | Clases, structs, interfaces, métodos, campos, delegados, namespaces                                              |
| Go         | `.go`                                       | Parámetros agrupados, sin nombre, variádicos, tipados como función y genéricos; resultados nombrados/múltiples; receptores genéricos | Campos de struct con tipos y tags, campos embebidos, firmas de métodos de interfaz, parámetros de tipo, paquetes |
| Java       | `.java`                                     | Parámetros tipados nombrados y tipos de retorno explícitos                                                                           | Clases, interfaces, constructores, métodos, campos, clases anidadas, paquetes                                    |
| JavaScript | `.js`, `.jsx`, `.mjs`, `.cjs`               | Parámetros nombrados; sin tipo de retorno inferido                                                                                   | Funciones, arrow functions, clases, campos, métodos, declaraciones que contienen JSX                             |
| PHP        | `.php`                                      | Parámetros tipados nombrados y tipos de retorno explícitos                                                                           | Clases, interfaces, traits, enums, métodos, propiedades, namespaces                                              |
| Python     | `.py`, `.pyi`                               | Anotaciones, valores por defecto, `*args`, `**kwargs`, separadores posicionales/de palabra clave, anotaciones de retorno             | Funciones, clases, clases anidadas, métodos decorados; receptores bound convencionales omitidos                  |
| Ruby       | `.rb`                                       | Parámetros nombrados; sin tipo de retorno inferido                                                                                   | Clases, módulos, métodos de instancia/singleton, constantes                                                      |
| Rust       | `.rs`                                       | Patrones tipados, `self`/`&self`/`&mut self`, bindings mutables, tipos de retorno explícitos                                         | Módulos inline y anidados, campos de struct nombrados, métodos de traits y métodos de implementación             |
| Scala      | `.scala`, `.sc`                             | Parámetros tipados nombrados y tipos de retorno explícitos                                                                           | Clases, traits, objetos, métodos, enums, paquetes                                                                |
| TypeScript | `.ts`, `.tsx`, `.mts`, `.cts`               | Parámetros tipados nombrados y anotaciones de retorno explícitas                                                                     | Funciones, arrow functions, interfaces, clases, campos, namespaces; TSX usa su propia gramática                  |

## Selección de lenguaje para encabezados

Todos los comandos CLI de procesamiento de código fuente aceptan `--header-language c|cpp`. La elección aplica solo a archivos `.h`. Otras extensiones conservan sus analizadores habituales.

```bash
codeknit parse ./src --header-language cpp
codeknit fingerprint ./src --header-language c
codeknit graph analyze ./src --header-language cpp
```

El valor predeterminado es `c`, preservando la extracción de macros, typedefs y unions de C. Use `cpp` para encabezados C++ nombrados `.h`; los archivos nombrados `.hpp` y `.hxx` ya usan C++. La elección es explícita porque ambas gramáticas pueden aceptar parte de la sintaxis del otro lenguaje sin extraer su estructura completa.

Los guards de encabezado y las ramas condicionales se recorren sin evaluar sus condiciones. Los resultados pueden incluir, por lo tanto, declaraciones de ramas mutuamente excluyentes. La expansión de macros, la resolución de includes y las flags de compilación no se aplican. La TUI usa la selección predeterminada de C para archivos `.h`.

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

Rust permite campos y métodos con el mismo nombre. Los nombres de campos con ámbito llevan un sufijo `#field` internamente para que sus IDs y relaciones de contención permanezcan distintos; los nombres de visualización y firmas conservan el nombre original del campo.

Las firmas de Python conservan anotaciones, valores por defecto, splats y separadores de convención de llamada. Un parámetro convencional primero `self` o `cls` se omite en un método bound, pero se conserva en una función libre o método estático.

Las firmas multilínea conservan los saltos de línea del código fuente en JSON. SKT escapa los saltos de línea como `\n` y `\r` para que cada símbolo permanezca en una sola línea física de salida.

## Límites conocidos

- Los declaradores complejos de punteros, referencias, arrays y punteros a función en C/C++ pueden tener firmas incompletas. Las declaraciones generadas por macros no se expanden.
- El manejo de constructores en TypeScript/JavaScript y algunas formas de parámetros opcionales, rest o desestructurados están incompletos.
- Las restricciones avanzadas de genéricos y formas de tipo fuera de los casos probados varían según el lenguaje. Los tipos de retorno no se infieren.
- Las macros de Rust no se expanden. Las declaraciones de módulos fuera de línea se registran; sus archivos deben incluirse en el escaneo. Los campos de tuplas, detalles de variantes de enums y elementos asociados tienen cobertura parcial.
- La extracción de sintaxis no resuelve el despacho dinámico, las sobrecargas, las dependencias importadas o la semántica de configuración de compilación con precisión de compilador.
- Los errores de sintaxis parciales aún pueden omitir declaraciones. Inspeccione las advertencias reportadas y verifique el código fuente antes de aplicar una refactorización.

Cambios más completos en la extracción modifican el conteo de símbolos y los ShortIDs. Regenerar la salida de skeleton existente al actualizar en lugar de mezclarla con salida más antigua.
