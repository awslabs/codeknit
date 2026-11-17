---
title: Primeros pasos
description: Comienza a usar codeknit en menos de 5 minutos.
---

En este tutorial compilarás codeknit desde el código fuente, ejecutarás tu primer análisis y leerás el resultado `.skt` que genera. Analizarás un paquete pequeño del repositorio de codeknit, por lo que no necesitarás un proyecto propio.

## Requisitos previos

- **Go 1.26 o posterior.** Ejecuta `go version` para verificar.
- **Un compilador C.** codeknit utiliza tree-sitter, que necesita CGo.
- **Git** y **Make**.

## 1. Compilar codeknit

Clona el repositorio y compila:

```bash
git clone https://github.com/awslabs/codeknit.git
cd codeknit
make build
```

Deberías ver que la compilación finaliza sin errores y un binario en `./bin/codeknit`.

## 2. Agregar codeknit a tu PATH

Desde el directorio `codeknit`, agrega `bin` a tu `PATH` usando la línea correspondiente a tu shell.

Para bash o zsh:

```bash
export PATH="$PATH:$(pwd)/bin"
```

Para fish:

```fish
fish_add_path (pwd)/bin
```

Esto aplica a tu sesión actual de bash o zsh. Para mantenerlo en sesiones futuras, agrega la misma línea a `~/.bashrc` o `~/.zshrc`, reemplazando `$(pwd)/bin` con la ruta absoluta al directorio `bin`.

## 3. Verificar la instalación

```bash
codeknit --version
```

Deberías ver la versión de codeknit impresa.

## 4. Ejecutar tu primer análisis

Permanece en el directorio `codeknit` y analiza el paquete `internal/console`:

```bash
codeknit parse ./internal/console --edges
```

codeknit escanea el directorio, analiza cada archivo fuente compatible y escribe archivos `.skt`. Un *archivo `.skt`* es un archivo de texto compacto que lista los símbolos en tu código y las relaciones entre ellos. La bandera `--edges` agrega las relaciones. Estas se omiten por defecto para ahorrar espacio.

Deberías ver un resumen breve cuando el comando finalice y un nuevo directorio `./skeleton` en tu directorio actual.

Lista el resultado:

```bash
ls ./skeleton
```

Deberías ver un archivo como `map_001.skt`.

## 5. Leer el resultado

Abre el archivo:

```bash
cat ./skeleton/map_001.skt
```

Deberías ver primero una sección `[symbols]`. Un *símbolo* es un elemento nombrado en tu código, como una función, tipo o campo. Los símbolos se agrupan bajo un encabezado `## <archivo fuente>`. Tu resultado incluye líneas como estas:

```skt
[symbols]
## <path to>/console.go
S1 module/package L5-L5 console
S2 type/struct L17-L25 Console {exported}
...
S10 callable/function L28-L41 New() -> *S2 {exported}
...
S12 callable/method L49-L49 Error(msg: string) {exported, receiver=*Console}
```

Cada línea tiene las mismas partes:

| Parte | Ejemplo | Significado |
| --- | --- | --- |
| ShortID | `S10` | Una etiqueta corta para el símbolo, usada en otras partes del resultado |
| Categoría y tipo | `callable/function` | Qué tipo de símbolo es |
| Rango de líneas | `L28-L41` | Las líneas que cubre el símbolo en el archivo fuente |
| Firma | `New() -> *S2` | El nombre, parámetros y tipo de retorno |
| Propiedades | `{exported}` | Detalles adicionales entre llaves |

En `New() -> *S2`, `S2` es el ShortID del tipo `Console`. codeknit reemplaza los nombres de tipos con ShortIDs cuando conoce el símbolo. Los números de línea pueden diferir ligeramente si tu copia es más reciente.

Desplázate hacia abajo hasta la sección `[edges]`. Una *relación* es una conexión entre dos símbolos. Cada línea de relación nombra el símbolo de origen, el tipo de relación y el símbolo o símbolos de destino:

```skt
[edges]
S2 --contains--> S4
S4 --calls--> S5
```

Este ejemplo muestra la estructura de una línea de relación. Significa que "`S2` contiene a `S4`" y "`S4` llama a `S5`". Puedes buscar cada ShortID en `[symbols]` para ver a qué se refiere.

## Lo que lograste

Compilaste codeknit, lo agregaste a tu `PATH`, ejecutaste `codeknit parse` y leíste las secciones `[symbols]` y `[edges]` del resultado `.skt`.

## Próximos pasos

- Analiza tu propio proyecto, elige un diseño de salida diferente y vuelve a ejecutar de forma segura: consulta la [guía del comando parse](../guides/parse-command/).
- Visualiza y analiza la estructura de tu código: consulta la [guía de comandos de grafo](../guides/graph-commands/).
- Encuentra código duplicado y casi duplicado: consulta la [guía del comando fingerprint](../guides/fingerprint-command/).
- Consulta cada parte del formato `.skt`: revisa la [referencia del formato de salida](../reference/output-format/).
- Consulta cada bandera de la línea de comandos: revisa la [referencia de banderas CLI](../reference/cli-flags/).