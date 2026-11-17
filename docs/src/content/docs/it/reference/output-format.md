---
title: Riferimento del formato di output
description: Riferimento completo per il formato di output .skt utilizzato da codeknit.
---

Questa pagina descrive la grammatica dell'output `.skt` e l'output JSON equivalente. Per le regole che decidono come vengono risolte le relazioni, consulta [Supporto linguistico e fedeltà dell'API](../language-support/).

## Panoramica

Un file `.skt` è un file di testo compatto composto da sezioni. Ogni sezione inizia con una riga di intestazione tra parentesi quadre.

| Sezione      | Presente quando                                      | Contenuti                                          |
| ------------ | ---------------------------------------------------- | -------------------------------------------------- |
| `[symbols]`  | Sempre                                               | Definizioni dei simboli raggruppate per file sorgente |
| `[edges]`    | `--edges` è impostato                                | Relazioni tra simboli                              |
| `[errors]`   | Si è verificato un errore di scansione o parsing     | File che non sono stati elaborati completamente    |
| `[dict]`     | `--minify` è impostato                               | Dizionario che mappa codici brevi a token ripetuti |

Nelle modalità di output per directory, `[errors]` e `[dict]` possono essere scritti nei file ausiliari `warnings.skt` e `dict.skt`. In modalità `inline` appaiono nello stesso output delle altre sezioni.

`--format json` produce un documento JSON con le stesse informazioni. Vedi [Output JSON](#json-output).

## [symbols]

I simboli sono raggruppati per file sorgente. Ogni gruppo inizia con una riga di intestazione `## <percorso>`, seguita da una riga per simbolo.

```skt
[symbols]
## pkg/services/auth.go
S1 module/package L1-L1 services {}
S2 type/struct L5-L8 AuthService {exported}
```

### Riga del simbolo

```text
ShortID categoria/tipo Linizio-Lfine firma {proprietà}
```

| Campo          | Descrizione                                                                                               |
| -------------- | --------------------------------------------------------------------------------------------------------- |
| `ShortID`      | Identificatore sequenziale: `S1`, `S2`, `S3`, e così via. Gli archi fanno riferimento ai simboli tramite questo valore. |
| `categoria/tipo` | Classificazione del simbolo. Vedi [Categorie e tipi](#categorie-e-tipi).                                  |
| `Linizio-Lfine` | Prima e ultima riga del simbolo nel file sorgente.                                                       |
| `firma`        | Nome del simbolo, con parametri e tipo di ritorno per i callable. Vedi [Formati della firma](#formati-della-firma). |
| `{proprietà}`  | Metadati tra parentesi. Vedi [Proprietà](#proprietà).                                                    |

Una riga del simbolo rimane su una singola riga fisica. I salti di riga all'interno di una firma sono escapati come `\n` e `\r`.

Gli ID dei simboli possono cambiare quando cambia l'estrazione. Rigenera tutti i chunk di output insieme invece di mescolare chunk provenienti da esecuzioni diverse. Questo si applica in particolare quando si esegue l'aggiornamento alla versione 0.5.0, perché gli ID dei simboli e i risultati della risoluzione delle relazioni possono cambiare.

### Categorie e tipi

| Categoria   | Tipi                                      | Esempi                                           |
| ---------- | ------------------------------------------ | -------------------------------------------------- |
| `callable` | function, method, constructor              | `callable/function`, `callable/method`             |
| `type`     | class, interface, struct, enum             | `type/class`, `type/interface`                     |
| `value`    | variable, constant, field                  | `value/variable`, `value/constant`                 |
| `module`   | package, namespace                         | `module/package`                                   |
| `meta`     | type parameters, files, includes, metadata | `meta/type_parameter`, `meta/file`, `meta/include` |

### Formati della firma

| Forma                         | Utilizzato per                                  |
| ---------------------------- | ----------------------------------------------- |
| `nome`                       | Tipi, valori e moduli                           |
| `nome(parametri)`            | Callable senza tipo di ritorno                  |
| `nome(parametri) -> tipoRitorno` | Callable con tipo di ritorno                |

I parametri sono scritti come:

| Forma          | Utilizzato per                  |
| ------------- | ------------------------------- |
| `nome`        | Linguaggi senza tipi            |
| `nome: tipo`  | Linguaggi con tipi              |

Quando un riferimento a un tipo corrisponde a un simbolo noto, il nome del tipo viene sostituito con l'ID breve di quel simbolo. Ad esempio, `-> *S2` fa riferimento a un puntatore al simbolo `S2`.

### Proprietà

Le proprietà appaiono tra parentesi dopo la firma, separate da virgole. Parentesi vuote (`{}`) indicano che il simbolo non ha proprietà.

| Forma della proprietà        | Esempio                 |
| ---------------------------- | ----------------------- |
| Flag                         | `exported`, `async`, `static` |
| `chiave=valore`              | `visibility=public`, `receiver=*AuthService`, `exported=false`, `resolution=unresolved` |

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

La sezione `[edges]` elenca le relazioni tra i simboli. Viene scritta solo quando gli archi sono richiesti con `--edges`.

### Riga dell'arco

```text
FromID --tipo--> ToID1, ToID2
FromID --tipo[unresolved]--> Target1, Target2
FromID --tipo[ambiguous]--> Target [candidates=CandidateID1, CandidateID2]
```

| Parte                     | Descrizione                                                                           |
| ------------------------ | ------------------------------------------------------------------------------------- |
| `FromID`                 | Estremo sorgente.                                                                     |
| `tipo`                   | Tipo di relazione. Vedi [Tipi di arco](#tipi-di-arco).                                |
| `[unresolved]`, `[ambiguous]` | Marcatore di risoluzione opzionale. Vedi [Marcatori di risoluzione](#marcatori-di-risoluzione). |
| `ToID1, ToID2`           | Uno o più target, separati da virgole.                                                |
| `[candidates=...]`       | Possibili target per un arco ambiguo, come elenco separato da virgole tra parentesi.  |

I target condividono una riga solo quando la sorgente, il tipo, il marcatore di risoluzione e l'insieme dei candidati sono identici. Una definizione referenziata da un arco può trovarsi in un file successivo o in un chunk di output diverso. Carica tutti i chunk prima di interpretare gli archi.

### Marcatori di risoluzione

| Marcatore         | Significato                                                                              |
| -------------- | ---------------------------------------------------------------------------------------- |
| Nessuno        | L'analizzatore ha risolto la relazione.                                                  |
| `[unresolved]` | L'analizzatore non è riuscito a stabilire la dipendenza.                                 |
| `[ambiguous]`  | Rimane più di un'interpretazione possibile. Gli ID dei candidati elencano i possibili target. |

Un marcatore mancante non è una garanzia del compilatore. Gli ID dei candidati identificano possibili target, non connessioni confermate. L'incertezza è sempre scritta in questa grammatica SKT, mai come JSON incorporato.

### Forme degli estremi

| Forma                      | Quando utilizzato                                                                                         | Esempio                     |
| ------------------------- | --------------------------------------------------------------------------------------------------------- | --------------------------- |
| Short ID                  | L'estremo o il candidato è un simbolo noto. Questo include gli estremi degli archi non risolti e i simboli di inclusione. | `S4`, `S8`       |
| Nome semplice             | Il target non ha un simbolo. Il nome è relativo al file sorgente.                                        | `Refresh`, `string`         |
| ID completo tra virgolette | L'identità è complessa, o un nome letterale assomiglia a un ID breve.                                    | `"app.go::S2"`              |

Gli ID completi tra virgolette escapano virgolette, backslash e salti di riga. Un nome letterale come `S2` deve essere scritto come ID completo tra virgolette in modo che non venga letto come un ID breve.

```skt
S4 --references[unresolved]--> string, bool
S5 --calls[ambiguous]--> Helper [candidates=S8, S9]
S6 --references[unresolved]--> "app.go::S2"
```

### Tipi di arco

| Tipo         | Significato                                         |
| ------------ | --------------------------------------------------- |
| `calls`      | invocazione di funzione/metodo                      |
| `contains`   | classe contiene metodo, modulo contiene funzione   |
| `inherits`   | classe estende un'altra                             |
| `implements` | classe implementa interfaccia                       |
| `overrides`  | metodo sovrascrive genitore                         |
| `references` | simbolo fa riferimento a un altro                   |
| `imports`    | modulo importa un altro                             |
| `decorates`  | decoratore applicato a simbolo                      |

I callable restituiti e gli argomenti di callback sono registrati come `references`, non come tipo separato.

### Simboli di file e inclusione in C e C++

I sorgenti C e C++ producono simboli `meta/file` e `meta/include`, quindi le relazioni di inclusione hanno ID brevi. Ogni ortografia di header appare una volta nell'elenco dei simboli, e gli archi utilizzano gli ID brevi. Le inclusioni ripetute all'interno di un file condividono un estremo.

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

Un simbolo di inclusione registra la direttiva come scritta. Non rappresenta una dichiarazione di header risolta.

### Effetto sull'analisi e la visualizzazione

- Gli archi incerti e i metadati di file/inclusione non contribuiscono alle metriche di dipendenza. I report mostrano il numero di relazioni escluse.
- Il grafo HTML disegna i link risolti e visualizza un conteggio delle relazioni incerte. Non disegna singoli archi incerti.

### Esempio

```skt
[edges]
S2 --contains--> S4
S4 --calls--> S5
S10 --inherits--> S2
S24 --implements--> S19
```

## [errors]

La sezione `[errors]` elenca i file che non sono stati elaborati completamente. È presente solo quando esistono diagnostiche.

```text
- <percorso>: <messaggio>
```

```skt
[errors]
- path/to/broken.go: syntax error at line 42
```

## [dict]

La sezione `[dict]` appare solo nell'output minimizzato (`--minify`). Mappa codici brevi a token ripetuti.

| Parte       | Descrizione                                                                 |
| ---------- | --------------------------------------------------------------------------- |
| `d0`, `d1`, … | Codici del dizionario.                                                      |
| Riga di voce | `- <codice>: <token>`                                                       |

Per leggere l'output minimizzato, sostituisci ogni codice con il suo valore nel dizionario. I codici possono sostituire coppie categoria/tipo e proprietà in `[symbols]`. Possono anche sostituire tipi di relazione in `[edges]`. I marcatori di incertezza rimangono leggibili. Ad esempio, `S1 --d2[unresolved]--> S3` è un arco non risolto il cui tipo è il valore di `d2`.

### Esempio minimizzato

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

Qui `S2` è un `callable/method` con la proprietà `async=false`.

## Esempio completo

Il seguente output contiene `[symbols]`, `[edges]` e `[errors]`. Non è minimizzato, quindi non ha una sezione `[dict]`; vedi [l'esempio minimizzato](#esempio-minimizzato).

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
- pkg/legacy/broken.go: syntax error at line 42
```

## Output JSON

`--format json`