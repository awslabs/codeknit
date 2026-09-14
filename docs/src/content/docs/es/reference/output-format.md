---
title: Referencia del formato de salida
description: Referencia completa del formato de salida .skt utilizado por codeknit.
---

El formato `.skt` (skeleton) es un formato de texto compacto y legible por humanos utilizado por `codeknit` para representar la **estructura de código** extraída. Contiene símbolos, **relaciones** y metadatos en una forma mínima adecuada para el consumo por LLM y el **análisis de grafo**.

Un archivo `.skt` está dividido en secciones. Cada sección comienza con un encabezado entre corchetes. Las secciones pueden aparecer en cualquier orden, aunque `[symbols]` típicamente aparece primero.

## [symbols]

La sección `[symbols]` lista todos los símbolos extraídos agrupados por su archivo de origen. Cada archivo se introduce con un encabezado `##` seguido de la ruta del archivo.

### Formato de línea

Cada símbolo se representa en una sola línea con la siguiente estructura:

```
ShortID category/kind Lstart-Lend signature {properties}
```

### Campos

- **ShortID**: Un identificador secuencial asignado a cada símbolo (por ejemplo, `S1`, `S2`, `S3`). Se utiliza como referencia en **relaciones** y otras secciones.
- **categoría/tipo**: Un par separado por barra que indica la categoría del símbolo y su tipo específico.
- **Linicio-Lfin**: El **rango de líneas** en el archivo de origen donde se define el símbolo (por ejemplo, `L10-L15`).
- **firma**: El nombre del símbolo y la información de tipo. El formato depende del símbolo:
  - `name` — para tipos, valores, módulos
  - `name(params)` — para **callables** sin tipo de retorno
  - `name(params) -> returnType` — para **callables** con tipo de retorno
- **{propiedades}**: Metadatos opcionales encerrados entre llaves. Varias propiedades se separan por comas.

### Parámetros

- En lenguajes sin tipos: `paramName`
- En lenguajes con tipos: `paramName: type`
- Las referencias de tipo que coinciden con símbolos conocidos se reemplazan con sus ShortIDs (por ejemplo, `config: S5` en lugar de `config: Config`).

### Propiedades

Propiedades comunes incluyen:

- `async`: `true` o `false`
- `exported`: `true` o `false`
- `static`: presente si el símbolo es estático
- `visibility=public|private|protected`
- `receiver=*TypeName`: para métodos, indica el tipo receptor

### Categorías y tipos de símbolos

| Categoría   | Tipos                                      | Ejemplos                                           |
| ----------- | ------------------------------------------ | -------------------------------------------------- |
| `callable`  | function, method, constructor              | `callable/function`, `callable/method`             |
| `type`      | class, interface, struct, enum             | `type/class`, `type/interface`                     |
| `value`     | variable, constant, field                  | `value/variable`, `value/constant`                 |
| `module`    | package, namespace                         | `module/package`                                   |
| `meta`      | type parameters, files, includes, metadata | `meta/type_parameter`, `meta/file`, `meta/include` |

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

La sección `[edges]` registra **relaciones** cuando se habilita la bandera `--edges`. Los puntos finales conocidos utilizan sus ShortIDs; los objetivos que no pueden resolverse a un símbolo pueden aparecer como nombres.

### Formato de línea

```
FromID --kind--> ToID1, ToID2
FromID --kind[unresolved]--> Target1, Target2
FromID --kind[ambiguous]--> Target [candidates=CandidateID1, CandidateID2]
```

La omisión del estado significa que el analizador resolvió la **relación**. `unresolved` significa que no pudo establecer la dependencia; `ambiguous` significa que quedan interpretaciones competidoras posibles. Los IDs de candidatos identifican posibles objetivos, no conexiones confirmadas. La resolución se basa en sintaxis y ámbito, no en una garantía del compilador.

Los objetivos se agrupan solo cuando su origen, tipo, resolución y conjunto de candidatos coinciden. Los nombres sin símbolos son relativos al archivo de origen. Las identidades complejas utilizan IDs completos entre comillas con comillas, barras invertidas y saltos de línea escapados. Un nombre literal que se asemeje a un ShortID debe citarse como un ID completo, como `"app.go::S2"`. Las definiciones de símbolos pueden ocurrir en archivos posteriores o fragmentos de salida.

```skt
S4 --references[unresolved]--> string, bool
S5 --calls[ambiguous]--> Helper [candidates=S8, S9]
S6 --references[unresolved]--> "app.go::S2"
```

### Puntos finales de archivo e include

Los includes de C/C++ tienen símbolos de metadatos explícitos. Las escrituras de encabezados aparecen una vez en la lista de símbolos, y las **relaciones** utilizan ShortIDs aunque la resolución del encabezado permanezca sin resolver:

```skt
[symbols]
## src/main.c
S1 meta/file L1-L3 main.c
S2 meta/include L1-L1 "config.h" {resolution=unresolved}
S3 meta/include L2-L2 <stdio.h> {resolution=unresolved}

[edges]
S1 --references[unresolved]--> S2, S3
```

Un símbolo de include registra la directiva escrita, no una declaración de encabezado resuelta. Los includes repetidos dentro de un archivo comparten un punto final. Los metadatos de archivo/include se excluyen de las clasificaciones de dependencia y los hallazgos de **dead code**.

Las **relaciones** inciertas se conservan en la salida del análisis pero se excluyen de las métricas de dependencia. Los informes de grafo muestran recuentos de **relaciones** excluidas; el visor HTML muestra su recuento sin dibujar **relaciones** inciertas individuales.

Al actualizar a `0.5.0`, regenera todos los fragmentos de salida juntos. Los IDs de símbolos y los resultados de **relaciones** pueden cambiar, por lo que los fragmentos de diferentes ejecuciones no deben mezclarse.

### Tipos de relaciones

| Tipo         | Significado                                         |
| ------------ | --------------------------------------------------- |
| `calls`      | invocación de función/método                        |
| `contains`   | clase contiene método, módulo contiene función      |
| `inherits`   | clase extiende otra clase                           |
| `implements` | clase implementa interfaz                           |
| `overrides`  | método sobrescribe método padre                     |
| `references` | símbolo hace referencia a otro símbolo              |
| `imports`    | módulo importa otro módulo                          |
| `decorates`  | decorador aplicado a un símbolo                     |

### Ejemplo

```skt
[edges]
S2 --contains--> S4
S4 --calls--> S5
S10 --inherits--> S2
S24 --implements--> S19
```

## [errors]

La sección `[errors]` lista los archivos que no pudieron analizarse completamente.

### Formato

Cada línea comienza con `-` seguido de la ruta del archivo y el mensaje de error:

```
- path/to/file.go: syntax error at line 42
```

### Ejemplo

```skt
[errors]
- src/broken.go: unexpected token at line 10
- tests/corner_case.py: unterminated string literal
```

## [dict]

La sección `[dict]` aparece solo cuando se utiliza la bandera `--minify`. Mapea códigos cortos del diccionario a tokens de cadena repetidos para reducir el tamaño de la salida.

### Formato

Cada línea mapea un código de diccionario (`d0`, `d1`, etc.) a su valor expandido:

```
- d0: async=false
- d1: callable/method
- d2: exported
```

En el resto del archivo, estos códigos reemplazan sus valores completos.

Los tipos de **relaciones** también pueden utilizar códigos de diccionario mientras que las etiquetas de estado permanecen legibles: `S1 --d2[unresolved]--> S3`. Expande `d2` utilizando el diccionario y conserva el estado unresolved.

### Ejemplo

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

## Ejemplo completo

```skt
[dict]
- d0: exported
- d1: callable/function

[symbols]
## main.go
S1 module/package L1-L1 main {}
S2 type/struct L5-L8 Server {d0}
S3 d1 L10-L12 NewServer(addr: string) -> *S2 {d0}
S4 callable/method L14-L20 Serve() {d0, receiver=*Server}
S5 callable/function L22-L25 handleError(err: error) -> bool {}

[edges]
S2 --contains--> S4
S4 --calls--> S5
S3 --references--> S2

[errors]
- utils/broken.go: syntax error at line 5
```

## Relaciones en JSON

Utiliza JSON cuando un script o integración necesite una salida estructurada:

```bash
codeknit parse ./src --output-mode inline --format json --edges
```

Cada **relación** incluye `from`, `to` y `kind`. Los IDs de origen conocidos aparecen en `from_short`; los IDs de destino resueltos aparecen en `to_short`. Las **relaciones** inciertas añaden `resolution` y, cuando están disponibles, `candidates` que contienen IDs completos de símbolos. Omiten `to_short`, incluso cuando el destino tiene un símbolo de metadatos de include.

Por ejemplo, una llamada ambigua puede representarse como:

```json
{
  "from": "app.py::run",
  "from_short": "S1",
  "to": "app.py::Helper",
  "kind": "calls",
  "resolution": "ambiguous",
  "candidates": ["helpers/a.py::Helper", "helpers/b.py::Helper"]
}
```

JSON y SKT preservan la misma incertidumbre con diferentes codificaciones. Consulta [Soporte de Lenguajes y Fidelidad de API](../language-support/) para conocer las reglas de resolución y los límites conocidos.