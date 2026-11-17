---
title: Referencia del formato de salida
description: Referencia completa del formato de salida .skt utilizado por codeknit.
---

Esta página describe la gramática del formato de salida `.skt` y el equivalente en JSON. Para las reglas que deciden cómo se resuelven las relaciones, consulta [Soporte de Lenguajes y Fidelidad de API](../language-support/).

## Visión general

Un archivo `.skt` es un archivo de texto compacto compuesto por secciones. Cada sección comienza con una línea de encabezado entre corchetes.

| Sección        | Presente cuando                                       | Contenido                                          |
| -------------- | ---------------------------------------------------- | ------------------------------------------------- |
| `[symbols]`    | Siempre                                              | Definiciones de símbolos agrupadas por archivo fuente |
| `[edges]`      | `--edges` está establecido                           | Relaciones entre símbolos                         |
| `[errors]`     | Ocurrió un diagnóstico de escaneo o análisis         | Archivos que no pudieron ser procesados completamente |
| `[dict]`       | `--minify` está establecido                          | Diccionario que mapea códigos cortos a tokens repetidos |

En los modos de salida de directorio, `[errors]` y `[dict]` pueden escribirse en los archivos auxiliares `warnings.skt` y `dict.skt`. En el modo `inline` aparecen en la misma salida que las otras secciones.

`--format json` produce un documento JSON con la misma información. Consulta [Salida JSON](#json-output).

## [symbols]

Los símbolos están agrupados por archivo fuente. Cada grupo comienza con una línea de encabezado `## <ruta>`, seguida de una línea por símbolo.

```skt
[symbols]
## pkg/services/auth.go
S1 module/package L1-L1 services {}
S2 type/struct L5-L8 AuthService {exported}
```

### Línea de símbolo

```text
ShortID categoría/tipo Linicio-Lfin firma {propiedades}
```

| Campo          | Descripción                                                                                               |
| -------------- | --------------------------------------------------------------------------------------------------------- |
| `ShortID`      | Identificador secuencial: `S1`, `S2`, `S3`, y así sucesivamente. Las relaciones hacen referencia a los símbolos mediante este valor. |
| `categoría/tipo` | Clasificación del símbolo. Consulta [Categorías y tipos](#categorías-y-tipos).                            |
| `Linicio-Lfin`  | Primera y última línea del símbolo en el archivo fuente.                                                  |
| `firma`        | Nombre del símbolo, con parámetros y tipo de retorno para elementos invocables. Consulta [Formatos de firma](#formatos-de-firma). |
| `{propiedades}` | Metadatos entre llaves. Consulta [Propiedades](#propiedades).                                             |

Una línea de símbolo se mantiene en una sola línea física. Los saltos de línea dentro de una firma se escapan como `\n` y `\r`.

Los IDs de símbolos pueden cambiar cuando la extracción cambia. Regenera todos los fragmentos de salida juntos en lugar de mezclar fragmentos de diferentes ejecuciones. Esto aplica en particular al actualizar a la versión 0.5.0, ya que los IDs de símbolos y los resultados de resolución de relaciones pueden cambiar.

### Categorías y tipos

| Categoría   | Tipos                                      | Ejemplos                                           |
| ---------- | ------------------------------------------ | -------------------------------------------------- |
| `callable` | function, method, constructor              | `callable/function`, `callable/method`             |
| `type`     | class, interface, struct, enum             | `type/class`, `type/interface`                     |
| `value`    | variable, constant, field                  | `value/variable`, `value/constant`                 |
| `module`   | package, namespace                         | `module/package`                                   |
| `meta`     | type parameters, files, includes, metadata | `meta/type_parameter`, `meta/file`, `meta/include` |

### Formatos de firma

| Formato                        | Utilizado para                                  |
| ---------------------------- | ---------------------------------------------- |
| `nombre`                      | Tipos, valores y módulos                       |
| `nombre(parámetros)`          | Elementos invocables sin tipo de retorno       |
| `nombre(parámetros) -> tipoRetorno` | Elementos invocables con tipo de retorno |

Los parámetros se escriben como:

| Formato          | Utilizado para                  |
| ------------- | ------------------------------ |
| `nombre`        | Lenguajes sin tipos             |
| `nombre: tipo`  | Lenguajes con tipos             |

Cuando una referencia de tipo coincide con un símbolo conocido, el nombre del tipo se reemplaza con el ID corto de ese símbolo. Por ejemplo, `-> *S2` hace referencia a un puntero al símbolo `S2`.

### Propiedades

Las propiedades aparecen entre llaves después de la firma, separadas por comas. Las llaves vacías (`{}`) indican que el símbolo no tiene propiedades.

| Formato de propiedad        | Ejemplo                 |
| ------------------------- | ----------------------- |
| Bandera                   | `exported`, `async`, `static` |
| `clave=valor`             | `visibility=public`, `receiver=*AuthService`, `exported=false`, `resolution=unresolved` |

### Ejemplo

```skt
[symbols]
## pkg/services/auth.go
S1 module/package L1-L1 services {}
S2 type/struct L5-L8 AuthService {exported}
S3 callable/function L10-L12 NewAuthService(secret: string, ttl: int) -> *S2 {exported}
S4 callable/method L14-L19 Authenticate(token: string) {exported, receiver=*AuthService}
S5 callable/function L29-L31 verifyToken(token: string) -> bool {exported=false}
```

## [edges]

La sección `[edges]` lista las relaciones entre símbolos. Se escribe solo cuando se solicitan relaciones con `--edges`.

### Línea de relación

```text
IDOrigen --tipo--> IDDestino1, IDDestino2
IDOrigen --tipo[no resuelto]--> Destino1, Destino2
IDOrigen --tipo[ambiguo]--> Destino [candidatos=IDCandidato1, IDCandidato2]
```

| Parte                     | Descripción                                                                           |
| ------------------------ | ------------------------------------------------------------------------------------- |
| `IDOrigen`               | Extremo de origen.                                                                    |
| `tipo`                   | Tipo de relación. Consulta [Tipos de relación](#tipos-de-relación).                   |
| `[no resuelto]`, `[ambiguo]` | Marcador de resolución opcional. Consulta [Marcadores de resolución](#marcadores-de-resolución). |
| `IDDestino1, IDDestino2` | Uno o más destinos, separados por comas.                                              |
| `[candidatos=...]`       | Posibles destinos para una relación ambigua, como una lista separada por comas dentro de corchetes. |

Los destinos comparten una línea solo cuando el origen, el tipo, el marcador de resolución y el conjunto de candidatos son idénticos. Una definición referenciada por una relación puede aparecer en un archivo posterior o fragmento de salida. Carga todos los fragmentos antes de interpretar las relaciones.

### Marcadores de resolución

| Marcador         | Significado                                                                              |
| -------------- | ---------------------------------------------------------------------------------------- |
| Ninguno        | El analizador resolvió la relación.                                                      |
| `[no resuelto]` | El analizador no pudo establecer la dependencia.                                         |
| `[ambiguo]`    | Más de una interpretación sigue siendo posible. Los IDs de candidatos listan los posibles destinos. |

Un marcador faltante no es una garantía del compilador. Los IDs de candidatos identifican posibles destinos, no conexiones confirmadas. La incertidumbre siempre se escribe en esta gramática SKT, nunca como JSON incrustado.

### Formas de extremo

| Forma                      | Cuándo se usa                                                                                         | Ejemplo                     |
| ------------------------- | ----------------------------------------------------------------------------------------------------- | --------------------------- |
| ID corto                  | El extremo o candidato es un símbolo conocido. Esto incluye extremos de relaciones no resueltas y símbolos de inclusión. | `S4`, `S8`       |
| Nombre simple             | El destino no tiene símbolo. El nombre es relativo al archivo fuente.                                | `Refresh`, `string`         |
| ID completo entre comillas | La identidad es compleja, o un nombre literal se asemeja a un ID corto.                              | `"app.go::S2"`              |

Los IDs completos entre comillas escapan comillas, barras invertidas y saltos de línea. Un nombre literal como `S2` debe escribirse como un ID completo entre comillas para que no se lea como un ID corto.

```skt
S4 --references[unresolved]--> string, bool
S5 --calls[ambiguous]--> Helper [candidates=S8, S9]
S6 --references[unresolved]--> "app.go::S2"
```

### Tipos de relación

| Tipo         | Significado                                         |
| ------------ | --------------------------------------------------- |
| `calls`      | Invocación de función/método                        |
| `contains`   | La clase contiene método, el módulo contiene función |
| `inherits`   | La clase extiende otra                              |
| `implements` | La clase implementa interfaz                        |
| `overrides`  | El método sobrescribe al padre                      |
| `references` | El símbolo hace referencia a otro                   |
| `imports`    | El módulo importa otro                              |
| `decorates`  | Decorador aplicado al símbolo                       |

Los elementos invocables devueltos y los argumentos de callback se registran como `references`, no como un tipo separado.

### Símbolos de archivo e inclusión en C y C++

Los fuentes de C y C++ producen símbolos `meta/file` y `meta/include`, por lo que las relaciones de inclusión tienen IDs cortos. Cada ortografía de encabezado aparece una vez en la lista de símbolos, y las relaciones usan los IDs cortos. Las inclusiones repetidas dentro de un archivo comparten un extremo.

```skt
[symbols]
## src/main.c
S1 meta/file L1-L4 main.c
S2 meta/include L1-L1 "config.h" {resolution=unresolved}
S3 meta/include L2-L2 "utils.h" {resolution=unresolved}
S4 meta/include L3-L3 <stdio.h> {resolution=unresolved}

[edges]
S1 --references[unresolved]--> S2, S3, S4
```

Un símbolo de inclusión registra la directiva tal como fue escrita. No representa una declaración de encabezado resuelta.

### Efecto en el análisis y la visualización

- Las relaciones inciertas y los metadatos de archivo/inclusión no contribuyen a las métricas de dependencia. Los informes muestran el número de relaciones excluidas.
- El grafo HTML dibuja enlaces resueltos y muestra un recuento de relaciones inciertas. No dibuja relaciones inciertas individuales.

### Ejemplo

```skt
[edges]
S2 --contains--> S4
S4 --calls--> S5
S10 --inherits--> S2
S24 --implements--> S19
```

## [errors]

La sección `[errors]` lista los archivos que no fueron procesados completamente. Solo está presente cuando existen diagnósticos.

```text
- <ruta>: <mensaje>
```

```skt
[errors]
- path/to/broken.go: error de sintaxis en la línea 42
```

## [dict]

La sección `[dict]` aparece solo en la salida minimizada (`--minify`). Mapea códigos cortos a tokens repetidos.

| Parte       | Descripción                                                                 |
| ---------- | --------------------------------------------------------------------------- |
| `d0`, `d1`, … | Códigos del diccionario.                                                    |
| Línea de entrada | `- <código>: <token>`                                                       |

Para leer la salida minimizada, reemplaza cada código con su valor en el diccionario. Los códigos pueden reemplazar pares categoría/tipo y propiedades en `[symbols]`. También pueden reemplazar tipos de relación en `[edges]`. Los marcadores de incertidumbre permanecen legibles. Por ejemplo, `S1 --d2[unresolved]--> S3` es una relación no resuelta cuyo tipo es el valor de `d2`.

### Ejemplo minimizado

```skt
[dict]
- d0: async=false
- d1: callable/method
- d2: exported

[symbols]
## src/handler.py
S1 type/class L1-L6 Handler {}
S2 d1 L2-L3 __init__(name) {d0}
S3 d1 L5-L6 handle(request) {d0}

[edges]
S1 --contains--> S2, S3
```

Aquí `S2` es un `callable/method` con la propiedad `async=false`.

## Ejemplo completo

La siguiente salida contiene `[symbols]`, `[edges]` y `[errors]`. No está minimizada, por lo que no tiene sección `[dict]`; consulta [el ejemplo minimizado](#ejemplo-minimizado).

```skt
[symbols]
## pkg/services/auth.go
S1 module/package L1-L1 services {}
S2 type/struct L5-L8 AuthService {exported}
S3 callable/function L10-L12 NewAuthService(secret: string, ttl: int) -> *S2 {exported}
S4 callable/method L14-L19 Authenticate(token: string) {exported, receiver=*AuthService}
S5 callable/function L29-L31 verifyToken(token: string) -> bool {exported=false}
## pkg/util/hash.go
S6 callable/function L3-L5 Hash(data: string) -> string {exported}
## pkg/util/hash_alt.go
S7 callable/function L3-L5 Hash(data: string) -> string {exported}

[edges]
S1 --contains--> S2, S3, S4, S5
S2 --contains--> S4
S3 --references--> S2
S4 --calls--> S5
S4 --calls[unresolved]--> Refresh
S5 --calls[ambiguous]--> Hash [candidates=S6, S7]

[errors]
- pkg/legacy/broken.go: error de sintaxis en la línea 42
```

## Salida JSON

`--format json`