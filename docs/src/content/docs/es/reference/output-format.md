---
title: Referencia del formato de salida
description: Referencia completa del formato de salida .skt utilizado por codeknit.
---

El formato `.skt` (skeleton) es un formato de texto compacto y legible por humanos utilizado por `codeknit` para representar la estructura de código extraída. Contiene símbolos, relaciones y metadatos en una forma minimalista adecuada para el consumo por LLM y el análisis estructural.

Un archivo `.skt` está dividido en secciones. Cada sección comienza con un encabezado entre corchetes. Las secciones pueden aparecer en cualquier orden, aunque `[symbols]` típicamente aparece primero.

## [symbols]

La sección `[symbols]` lista todos los símbolos extraídos agrupados por su archivo de origen. Cada archivo se introduce con un encabezado `##` seguido de la ruta del archivo.

### Formato de línea

Cada símbolo se representa en una sola línea con la siguiente estructura:

```
ShortID categoría/tipo Linicio-Lfin firma {propiedades}
```

### Campos

- **ShortID**: Un identificador secuencial asignado a cada símbolo (por ejemplo, `S1`, `S2`, `S3`). Se utiliza como referencia en relaciones y otras secciones.
- **categoría/tipo**: Un par separado por una barra que indica la categoría del símbolo y su tipo específico.
- **Linicio-Lfin**: El rango de líneas en el archivo de origen donde se define el símbolo (por ejemplo, `L10-L15`).
- **firma**: El nombre del símbolo y la información de tipo. El formato depende del símbolo:
  - `nombre` — para tipos, valores, módulos
  - `nombre(parámetros)` — para elementos invocables sin tipo de retorno
  - `nombre(parámetros) -> tipoRetorno` — para elementos invocables con tipo de retorno
- **{propiedades}**: Metadatos opcionales encerrados entre llaves. Varias propiedades se separan por comas.

### Parámetros

- En lenguajes sin tipos: `nombreParámetro`
- En lenguajes con tipos: `nombreParámetro: tipo`
- Las referencias de tipo que coinciden con símbolos conocidos se reemplazan con sus ShortIDs (por ejemplo, `config: S5` en lugar de `config: Config`).

### Propiedades

Propiedades comunes incluyen:

- `async`: `true` o `false`
- `exported`: `true` o `false`
- `static`: presente si el símbolo es estático
- `visibility=public|private|protected`
- `receiver=*NombreTipo`: para métodos, indica el tipo receptor

### Categorías y tipos de símbolos

| Categoría  | Tipos                          | Ejemplos                               |
| ---------- | ------------------------------ | -------------------------------------- |
| `callable` | function, method, constructor  | `callable/function`, `callable/method` |
| `type`     | class, interface, struct, enum | `type/class`, `type/interface`         |
| `value`    | variable, constant, field      | `value/variable`, `value/constant`     |
| `module`   | package, namespace             | `module/package`                       |
| `meta`     | type parameters, metadata      | `meta/type_parameter`                  |

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

La sección `[edges]` define relaciones entre símbolos utilizando sus ShortIDs.

### Formato de línea

```
IDOrigen --tipo--> IDDestino1, IDDestino2
```

Varios IDs de destino se separan por comas. Cada línea representa una relación direccional.

### Tipos de relaciones

| Tipo         | Significado                                    |
| ------------ | ---------------------------------------------- |
| `calls`      | invocación de función/método                   |
| `contains`   | clase contiene método, módulo contiene función |
| `inherits`   | clase extiende otra clase                      |
| `implements` | clase implementa interfaz                      |
| `overrides`  | método sobrescribe método padre                |
| `references` | símbolo hace referencia a otro símbolo         |
| `imports`    | módulo importa otro módulo                     |
| `decorates`  | decorador aplicado a un símbolo                |

### Ejemplo

```skt
[edges]
S2 --contains--> S4
S4 --calls--> S5
S10 --inherits--> S2
S24 --implements--> S19
```

## [errors]

La sección `[errors]` lista los archivos que no pudieron ser analizados completamente.

### Formato

Cada línea comienza con `-` seguido de la ruta del archivo y el mensaje de error:

```
- ruta/al/archivo.go: error de sintaxis en la línea 42
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
S3 --returns--> S2

[errors]
- utils/broken.go: syntax error at line 5
```
