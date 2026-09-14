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
ShortID category/kind Lstart-Lend signature {properties}
```

### Campi

- **ShortID**: Un identificatore sequenziale assegnato a ogni simbolo (es. `S1`, `S2`, `S3`). Utilizzato come riferimento negli archi e in altre sezioni.
- **categoria/tipo**: Una coppia separata da slash che indica la categoria del simbolo e il tipo specifico.
- **Linizio-Lfine**: L'intervallo di righe nel file sorgente dove il simbolo è definito (es. `L10-L15`).
- **firma**: Il nome del simbolo e le informazioni sul tipo. Il formato dipende dal simbolo:
  - `name` — per tipi, valori, moduli
  - `name(params)` — per callable senza tipo di ritorno
  - `name(params) -> returnType` — per callable con tipo di ritorno
- **{proprietà}**: Metadati opzionali racchiusi tra parentesi graffe. Più proprietà sono separate da virgole.

### Parametri

- In linguaggi non tipizzati: `paramName`
- In linguaggi tipizzati: `paramName: type`
- I riferimenti ai tipi che corrispondono a simboli noti sono sostituiti con i loro ShortID (es. `config: S5` invece di `config: Config`).

### Proprietà

Proprietà comuni includono:

- `async`: `true` o `false`
- `exported`: `true` o `false`
- `static`: presente se il simbolo è statico
- `visibility=public|private|protected`
- `receiver=*TypeName`: per i metodi, indica il tipo del receiver

### Categorie e tipi di simboli

| Categoria   | Tipi                                      | Esempi                                           |
| ---------- | ------------------------------------------ | -------------------------------------------------- |
| `callable` | function, method, constructor              | `callable/function`, `callable/method`             |
| `type`     | class, interface, struct, enum             | `type/class`, `type/interface`                     |
| `value`    | variable, constant, field                  | `value/variable`, `value/constant`                 |
| `module`   | package, namespace                         | `module/package`                                   |
| `meta`     | type parameters, files, includes, metadata | `meta/type_parameter`, `meta/file`, `meta/include` |

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

La sezione `[edges]` registra le relazioni quando `--edges` è abilitato. Gli endpoint noti utilizzano i loro ShortID; i target che non possono essere risolti a un simbolo possono apparire come nomi.

### Formato della riga

```
FromID --kind--> ToID1, ToID2
FromID --kind[unresolved]--> Target1, Target2
FromID --kind[ambiguous]--> Target [candidates=CandidateID1, CandidateID2]
```

Un stato omesso significa che l'analizzatore ha risolto la relazione. `unresolved` significa che non è stato possibile stabilire la dipendenza; `ambiguous` significa che rimangono possibili interpretazioni concorrenti. Gli ID dei candidati identificano possibili target, non connessioni confermate. La risoluzione si basa sulla sintassi e sullo scope, non su una garanzia del compilatore.

I target sono raggruppati solo quando la loro sorgente, tipo, risoluzione e insieme di candidati corrispondono. I nomi senza simboli sono relativi al file sorgente. Le identità complesse utilizzano ID completi tra virgolette con virgolette, backslash e interruzioni di riga escapati. Un nome letterale che assomiglia a un ShortID deve essere quotato come ID completo, come `"app.go::S2"`. Le definizioni dei simboli possono verificarsi in file successivi o chunk di output.

```skt
S4 --references[unresolved]--> string, bool
S5 --calls[ambiguous]--> Helper [candidates=S8, S9]
S6 --references[unresolved]--> "app.go::S2"
```

### Endpoint di file e include

Le include C/C++ hanno simboli di metadati espliciti. Le spelling delle intestazioni appaiono una volta nell'elenco dei simboli, e gli archi utilizzano ShortID anche se la risoluzione dell'intestazione rimane non risolta:

```skt
[symbols]
## src/main.c
S1 meta/file L1-L3 main.c
S2 meta/include L1-L1 "config.h" {resolution=unresolved}
S3 meta/include L2-L2 <stdio.h> {resolution=unresolved}

[edges]
S1 --references[unresolved]--> S2, S3
```

Un simbolo di include registra la direttiva scritta, non una dichiarazione di intestazione risolta. Le include ripetute all'interno di un file condividono un endpoint. I metadati di file/include sono esclusi dalle classifiche di dipendenza e dalle scoperte di dead code.

Le relazioni incerte sono preservate nell'output di parsing ma escluse dalle metriche di dipendenza. I report del grafo mostrano i conteggi delle relazioni escluse; il visualizzatore HTML mostra il loro conteggio senza disegnare i singoli archi incerti.

Quando si esegue l'aggiornamento a `0.5.0`, rigenerare tutti i chunk di output insieme. Gli ID dei simboli e i risultati delle relazioni possono cambiare, quindi i chunk di diverse esecuzioni non devono essere mischiati.

### Tipi di arco

| Tipo         | Significato                                         |
| ------------ | ----------------------------------------------- |
| `calls`      | invocazione di funzione/metodo                      |
| `contains`   | classe contiene metodo, modulo contiene funzione |
| `inherits`   | classe estende un'altra classe                     |
| `implements` | classe implementa interfaccia                      |
| `overrides`  | metodo sovrascrive metodo genitore                 |
| `references` | simbolo fa riferimento a un altro simbolo          |
| `imports`    | modulo importa un altro modulo                     |
| `decorates`  | decoratore applicato a un simbolo                  |

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
- path/to/file.go: syntax error at line 42
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

I tipi di relazione possono anche utilizzare codici del dizionario mentre le etichette di stato rimangono leggibili: `S1 --d2[unresolved]--> S3`. Espandere `d2` utilizzando il dizionario e mantenere lo stato unresolved.

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
S3 --references--> S2

[errors]
- utils/broken.go: syntax error at line 5
```

## Relazioni JSON

Utilizzare JSON quando uno script o un'integrazione necessita di output strutturato:

```bash
codeknit parse ./src --output-mode inline --format json --edges
```

Ogni arco include `from`, `to` e `kind`. Gli ID sorgente noti appaiono in `from_short`; gli ID target risolti appaiono in `to_short`. Gli archi incerti aggiungono `resolution` e, quando disponibili, `candidates` contenenti gli ID completi dei simboli. Omettono `to_short`, anche quando il target ha un simbolo di metadati di include.

Ad esempio, una chiamata ambigua può essere rappresentata come:

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

JSON e SKT preservano la stessa incertezza con codifiche diverse. Vedere [Supporto linguistico e fedeltà dell'API](../language-support/) per le regole di risoluzione e i limiti noti.