---
title: Supporto Linguistico e Fedeltà dell'API
description: Capacità di estrazione testate, selezione del parser e limitazioni note delle lingue.
---

<!-- Maintained in docs/reference/language-support.md; copied verbatim by docs/gen. -->

Codeknit estrae la sintassi senza eseguire un compilatore, valutare macro o caricare le dipendenze di un progetto. La matrice seguente descrive i costrutti testati, non la conformità completa del linguaggio o la risoluzione garantita dei nomi tra file.

Ogni lingua ha un'asserzione esatta dei parametri di base/firma di ritorno in `internal/plugin/api_fidelity_test.go`. La copertura delle dichiarazioni è verificata anche in `internal/plugin/coverage_test.go` e nei pacchetti di test delle singole lingue.

| Lingua      | Estensioni                                  | Baseline della firma                                                                                         | Copertura delle dichiarazioni e dei contenitori                                                            |
| ----------- | ------------------------------------------- | ------------------------------------------------------------------------------------------------------------ | ---------------------------------------------------------------------------------------------------------- |
| C           | `.c`, `.h`                                  | Parametri nominati e tipi di ritorno nelle definizioni e nei prototipi                                      | Struct, union, enum, typedef, macro; dichiarazioni all'interno di guardie di intestazione                 |
| C++         | `.cpp`, `.cc`, `.cxx`, `.hpp`, `.hxx`, `.h` | Parametri nominati e tipi di ritorno in funzioni, prototipi e dichiarazioni di metodi                       | Classi, struct, metodi, namespace, template; guardie di intestazione e wrapper `extern "C"`               |
| C#          | `.cs`                                       | Parametri tipizzati nominati e tipi di ritorno espliciti                                                    | Classi, struct, interfacce, metodi, campi, delegati, namespace                                            |
| Go          | `.go`                                       | Parametri raggruppati, non nominati, variadici, tipizzati come funzioni e generici; risultati multipli/nominati; ricevitori generici | Campi di struct con tipi e tag, campi embedded, firme dei metodi delle interfacce, parametri di tipo, pacchetti |
| Java        | `.java`                                     | Parametri tipizzati nominati e tipi di ritorno espliciti                                                    | Classi, interfacce, costruttori, metodi, campi, classi nidificate, pacchetti                              |
| JavaScript  | `.js`, `.jsx`, `.mjs`, `.cjs`               | Parametri nominati; nessun tipo di ritorno inferito                                                         | Funzioni, funzioni freccia, classi, campi, metodi, dichiarazioni contenenti JSX                            |
| PHP         | `.php`                                      | Parametri tipizzati nominati e tipi di ritorno espliciti                                                    | Classi, interfacce, trait, enum, metodi, proprietà, namespace                                              |
| Python      | `.py`, `.pyi`                               | Annotazioni, valori predefiniti, `*args`, `**kwargs`, separatori posizionali/keyword, annotazioni di ritorno | Funzioni, classi, classi nidificate, metodi decorati; ricevitori bound convenzionali omessi                |
| Ruby        | `.rb`                                       | Parametri nominati; nessun tipo di ritorno inferito                                                         | Classi, moduli, metodi di istanza/singleton, costanti                                                      |
| Rust        | `.rs`                                       | Pattern tipizzati, `self`/`&self`/`&mut self`, binding mutabili, tipi di ritorno espliciti                  | Moduli inline e nidificati, campi di struct nominati, metodi di trait e metodi di implementazione         |
| Scala       | `.scala`, `.sc`                             | Parametri tipizzati nominati e tipi di ritorno espliciti                                                    | Classi, trait, oggetti, metodi, enum, pacchetti                                                            |
| TypeScript  | `.ts`, `.tsx`, `.mts`, `.cts`               | Parametri tipizzati nominati e annotazioni di ritorno esplicite                                             | Funzioni, funzioni freccia, interfacce, classi, campi, namespace; TSX utilizza la propria grammatica      |

## Selezione della lingua per le intestazioni

Tutti i comandi CLI di elaborazione del codice sorgente accettano `--header-language c|cpp`. La scelta si applica solo ai file `.h`. Altre estensioni mantengono i loro parser usuali.

```bash
codeknit parse ./src --header-language cpp
codeknit fingerprint ./src --header-language c
codeknit graph analyze ./src --header-language cpp
```

Il valore predefinito è `c`, preservando l'estrazione di macro, typedef e union C. Utilizzare `cpp` per le intestazioni C++ denominate `.h`; i file denominati `.hpp` e `.hxx` utilizzano già C++. La scelta è esplicita perché entrambe le grammatiche possono accettare parte della sintassi dell'altra lingua senza estrarne l'intera struttura.

Le guardie di intestazione e i rami condizionali vengono attraversati senza valutare le loro condizioni. I risultati possono quindi includere dichiarazioni da rami mutuamente esclusivi. L'espansione delle macro, la risoluzione degli include e i flag di compilazione non vengono applicati. La TUI utilizza la selezione C predefinita per i file `.h`.

## Preservazione della forma dell'API

Ad esempio, in Go:

```go
func Pair(a, b int, rest ...string) (string, error)
```

produce:

```text
Pair(a: int, b: int, rest: ...string) -> (string, error)
```

I metodi delle interfacce Go e i campi nominati di Go/Rust appaiono come simboli sia in modalità normale che di fingerprint. I metodi e i campi mantengono il loro tipo proprietario. I moduli Rust nidificati mantengono nomi qualificati come `service.storage.read`, anche quando un altro modulo dichiara una funzione con lo stesso nome.

Rust consente campi e metodi con lo stesso nome. I nomi dei campi con scope portano un suffisso `#field` internamente in modo che i loro ID e gli archi di contenimento rimangano distinti; i nomi visualizzati e le firme mantengono il nome originale del campo.

Le firme Python mantengono annotazioni, valori predefiniti, splat e separatori di convenzione di chiamata. Un parametro convenzionale primo `self` o `cls` viene omesso da un metodo bound, ma preservato in una funzione libera o in un metodo statico.

Le firme multilinea mantengono le interruzioni di riga del sorgente in JSON. SKT sfugge le interruzioni di riga come `\n` e `\r` in modo che ogni simbolo rimanga su una riga fisica di output.

## Risoluzione delle relazioni

Le relazioni mantengono la qualificazione del ricevitore e la proprietà lessicale. Import relativi e import rinominati di JavaScript/TypeScript e Python forniscono il contesto di destinazione. Le dichiarazioni dei pacchetti Go e Java possono risolvere tra file nella stessa directory. Un import non corrispondente non ricade mai su una dichiarazione non correlata, e i candidati concorrenti rimangono ambigui.

I binding dei ricevitori locali hanno la precedenza sugli import dei namespace, anche quando un parametro oscura un namespace importato. Gli import dei pacchetti Python considerano le dichiarazioni nel modulo `__init__` del pacchetto, non moduli discendenti arbitrari. Le riesportazioni che non possono essere stabilite da quel modulo rimangono irrisolte.

L'estrazione dei chiamabili in Java, C++, C#, Scala e TypeScript mantiene gli intervalli di byte delle dichiarazioni in modo che i corpi sovraccaricati possano avere chiamanti distinti e ambiti di alias, comprese le dichiarazioni sulla stessa riga. La selezione di un target sovraccaricato rimane conservativa. Le relazioni di override esplicite in Java, C++, C#, Scala e TypeScript cercano la dichiarazione ereditata più vicina; basi mancanti, cicli o candidati concorrenti mantengono l'incertezza invece di produrre collegamenti a se stessi.

Questa è un'analisi basata sulla sintassi, non un compilatore o un grafo delle chiamate a runtime. Specificatori di pacchetti JavaScript nudi, riesportazioni di pacchetti, configurazioni di build, tipi di ricevitori complessi e dispatch dinamico possono rimanere irrisolti. Un nome univoco altrove nel repository non è una prova sufficiente di una dipendenza.

JSON preserva gli archi incerti con `resolution: "unresolved"` o `"ambiguous"` e include gli ID dei candidati quando disponibili. Questi archi non hanno un ID breve di destinazione.

SKT utilizza la stessa notazione degli archi per tutte le relazioni, con un marcatore di risoluzione opzionale:

```text
S8 --references--> S5
S10 --references[unresolved]--> string, bool
S11 --calls[ambiguous]--> Helper [candidates=S12, S15]
```

Gli endpoint noti e i candidati utilizzano ID brevi, compresi gli endpoint delle relazioni irrisolte. Bersagli incerti senza un simbolo sono nomi relativi al file sorgente. Nomi letterali che sembrano ID brevi, e altre identità complesse, utilizzano ID completi tra virgolette per evitare ambiguità. Gli archi con la stessa sorgente, tipo, risoluzione e insieme di candidati condividono un elenco di destinazioni. Il lettore SKT preserva la risoluzione e i candidati, comprese le referenze a simboli in chunk di output successivi. L'output minimizzato codifica il tipo di relazione mantenendo il marcatore di risoluzione leggibile.

Include C/C++ hanno simboli espliciti `meta/file` e `meta/include`. Le grafie delle intestazioni appaiono una volta nell'elenco dei simboli; gli archi utilizzano i loro ID brevi:

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

Un simbolo di include descrive la direttiva scritta, non una dichiarazione di intestazione risolta. Include ripetuti all'interno di un file condividono un endpoint. I metadati di file/include sono esclusi dalle classifiche delle dipendenze e dai risultati di dead code; l'aggiunta di ID non afferma che un'intestazione sia stata trovata attraverso la configurazione di ricerca include del compilatore.

Le metriche delle dipendenze e i collegamenti del grafo HTML utilizzano solo archi risolti. I report mostrano i conteggi delle relazioni escluse e il grafo HTML visualizza un conteggio di incertezza. Fan-in e fan-out contano vicini distinti. L'analisi dell'ereditarietà considera tutte le basi e le interfacce; l'ereditarietà ciclica non ha una profondità riportata. La raggiungibilità parte da chiamabili nominati `main`, `Main` o `init` e segue gli archi delle dipendenze. Raggiungere un membro rende il suo contenitore rilevante senza rendere raggiungibile ogni fratello. Senza punti di ingresso riconosciuti, i risultati di raggiungibilità sono omessi.

Gli argomenti di callback e i chiamabili restituiti sono riferimenti, non prove di invocazione. La risoluzione degli alias mantiene l'identità del file, della funzione e dell'oggetto; assegnazioni contrastanti rimangono incerte. I risultati del grafo sono candidati per la revisione: chiamanti esterni e comportamento dinamico possono cambiare il risultato. Le regole di livello basate su directory e i pesi di propagazione delle modifiche sono euristiche, non probabilità misurate.

## Limiti noti

- Dichiaratori complessi di puntatori, riferimenti, array e puntatori a funzione in C/C++ possono ancora avere firme incomplete. Le dichiarazioni generate da macro non vengono espanse.
- La gestione dei costruttori in TypeScript/JavaScript e alcune forme di parametri opzionali, rest o destrutturati sono incomplete.
- Vincoli generici avanzati e forme di tipo al di fuori dei casi testati variano a seconda della lingua. I tipi di ritorno non vengono inferiti.
- Le macro Rust non vengono espanse. Le dichiarazioni di moduli out-of-line sono registrate; i loro file devono essere inclusi nella scansione. I campi delle tuple, i dettagli delle varianti enum e gli elementi associati hanno una copertura parziale.
- L'estrazione della sintassi non risolve il dispatch dinamico, i sovraccarichi, le dipendenze importate o la semantica della configurazione di build con la precisione del compilatore.
- Errori di sintassi parziali possono ancora omettere dichiarazioni. Ispezionare gli avvisi riportati e verificare il sorgente prima di applicare un refactoring.

Un'estrazione più completa modifica i conteggi dei simboli e gli ID brevi. Rigenerare l'output dello skeleton esistente quando si esegue l'aggiornamento invece di mescolarlo con output più vecchi.