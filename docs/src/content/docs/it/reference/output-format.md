---
title: Riferimento del formato di output
description: Riferimento completo per il formato di output .skt utilizzato da codeknit.
---

Il formato `.skt` (skeleton) è un formato di testo compatto e leggibile dall'uomo utilizzato da `codeknit` per rappresentare la struttura del codice estratta. Contiene simboli, relazioni e metadati in una forma minimale adatta al consumo da parte di LLM e all'analisi strutturale.

Un file `.skt` è diviso in sezioni. Ogni sezione inizia con un'intestazione tra parentesi quadre. Le sezioni possono apparire in qualsiasi ordine, anche se `[symbols]` tipicamente viene per prima.

## [symbols]

La sezione `[symbols]` elenca tutti i simboli estratti raggruppati per file sorgente. Ogni file è introdotto con un'intestazione `##` seguita dal percorso del file.

### Formato della riga

Ogni simbolo è rappresentato su una singola riga con la seguente struttura:

```
ShortID categoria/tipo Linizio-Lfine firma {proprietà}
```

### Campi

- **ShortID**: Un identificatore sequenziale assegnato a ogni simbolo (ad es., `S1`, `S2`, `S3`). Utilizzato come riferimento negli archi e in altre sezioni.
- **categoria/tipo**: Una coppia separata da slash che indica la categoria del simbolo e il tipo specifico.
- **Linizio-Lfine**: L'intervallo di righe nel file sorgente in cui il simbolo è definito (ad es., `L10-L15`).
- **firma**: Il nome del simbolo e le informazioni sul tipo. Il formato dipende dal simbolo:
  - `nome` — per tipi, valori, moduli
  - `nome(parametri)` — per callable senza tipo di ritorno
  - `nome(parametri) -> tipoRitorno` — per callable con tipo di ritorno
- **{proprietà}**: Metadati opzionali racchiusi tra parentesi graffe. Più proprietà sono separate da virgole.

### Parametri

- In linguaggi non tipizzati: `nomeParametro`
- In linguaggi tipizzati: `nomeParametro: tipo`
- I riferimenti ai tipi che corrispondono a simboli noti sono sostituiti con i loro ShortID (ad es., `config: S5` invece di `config: Config`).

### Proprietà

Proprietà comuni includono:

- `async`: `true` o `false`
- `exported`: `true` o `false`
- `static`: presente se il simbolo è statico
- `visibility=public|private|protected`
- `receiver=*NomeTipo`: per i metodi, indica il tipo del receiver

### Categorie e tipi di simboli

| Categoria  | Tipi                           | Esempi                                 |
| ---------- | ------------------------------ | -------------------------------------- |
| `callable` | function, method, constructor  | `callable/function`, `callable/method` |
| `type`     | class, interface, struct, enum | `type/class`, `type/interface`         |
| `value`    | variable, constant, field      | `value/variable`, `value/constant`     |
| `module`   | package, namespace             | `module/package`                       |
| `meta`     | type parameters, metadata      | `meta/type_parameter`                  |

### Esempio

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

La sezione `[edges]` definisce le relazioni tra i simboli utilizzando i loro ShortID.

### Formato della riga

```
IDOrigine --tipo--> IDDestinazione1, IDDestinazione2
```

Più ID di destinazione sono separati da virgole. Ogni riga rappresenta una relazione direzionale.

### Tipi di arco

| Tipo         | Significato                                      |
| ------------ | ------------------------------------------------ |
| `calls`      | invocazione di funzione/metodo                   |
| `contains`   | classe contiene metodo, modulo contiene funzione |
| `inherits`   | classe estende un'altra classe                   |
| `implements` | classe implementa interfaccia                    |
| `overrides`  | metodo sovrascrive metodo padre                  |
| `references` | simbolo fa riferimento a un altro simbolo        |
| `imports`    | modulo importa un altro modulo                   |
| `decorates`  | decoratore applicato a un simbolo                |

### Esempio

```skt
[edges]
S2 --contains--> S4
S4 --calls--> S5
S10 --inherits--> S2
S24 --implements--> S19
```

## [errors]

La sezione `[errors]` elenca i file che non sono stati analizzati completamente.

### Formato

Ogni riga inizia con `-` seguito dal percorso del file e dal messaggio di errore:

```
- percorso/al/file.go: errore di sintassi alla riga 42
```

### Esempio

```skt
[errors]
- src/broken.go: unexpected token at line 10
- tests/corner_case.py: unterminated string literal
```

## [dict]

La sezione `[dict]` appare solo quando viene utilizzato il flag `--minify`. Mappa codici brevi del dizionario a token stringa ripetuti per ridurre la dimensione dell'output.

### Formato

Ogni riga mappa un codice del dizionario (`d0`, `d1`, ecc.) al suo valore espanso:

```
- d0: async=false
- d1: callable/method
- d2: exported
```

Nel resto del file, questi codici sostituiscono i loro valori completi.

### Esempio

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

## Esempio completo

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
