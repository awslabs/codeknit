---
title: Supporto per i linguaggi e fedeltà dell'API
description: Capacità di estrazione testate, selezione del parser e limitazioni note dei linguaggi.
---

<!-- Mantenuto in docs/reference/language-support.md; copiato letteralmente da docs/gen. -->

Codeknit estrae la sintassi senza eseguire un compilatore, valutare macro o caricare le dipendenze di un progetto. La matrice seguente descrive i costrutti testati, non la conformità completa del linguaggio o la risoluzione garantita dei nomi tra file.

Ogni linguaggio ha un'asserzione esatta dei parametri di base/firma di ritorno in `internal/plugin/api_fidelity_test.go`. La copertura delle dichiarazioni è verificata anche in `internal/plugin/coverage_test.go` e nei singoli pacchetti di test per il linguaggio.

| Linguaggio | Estensioni                                  | Baseline della firma                                                                                                                 | Copertura delle dichiarazioni e dei contenitori                                                               |
| ---------- | ------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------- |
| C          | `.c`, `.h`                                  | Parametri nominati e tipi di ritorno nelle definizioni e nei prototipi                                                               | Struct, union, enum, typedef, macro; dichiarazioni all'interno di guardie di intestazione                     |
| C++        | `.cpp`, `.cc`, `.cxx`, `.hpp`, `.hxx`, `.h` | Parametri nominati e tipi di ritorno in funzioni, prototipi e dichiarazioni di metodi                                                | Classi, struct, metodi, namespace, template; guardie di intestazione e wrapper `extern "C"`                   |
| C#         | `.cs`                                       | Parametri tipizzati nominati e tipi di ritorno espliciti                                                                             | Classi, struct, interfacce, metodi, campi, delegati, namespace                                                |
| Go         | `.go`                                       | Parametri raggruppati, non nominati, variadici, tipizzati come funzioni e generici; risultati multipli/nominati; ricevitori generici | Campi di struct con tipi e tag, campi embedded, firme dei metodi di interfaccia, parametri di tipo, pacchetti |
| Java       | `.java`                                     | Parametri tipizzati nominati e tipi di ritorno espliciti                                                                             | Classi, interfacce, costruttori, metodi, campi, classi nidificate, pacchetti                                  |
| JavaScript | `.js`, `.jsx`, `.mjs`, `.cjs`               | Parametri nominati; nessun tipo di ritorno inferito                                                                                  | Funzioni, arrow function, classi, campi, metodi, dichiarazioni contenenti JSX                                 |
| PHP        | `.php`                                      | Parametri tipizzati nominati e tipi di ritorno espliciti                                                                             | Classi, interfacce, trait, enum, metodi, proprietà, namespace                                                 |
| Python     | `.py`, `.pyi`                               | Annotazioni, valori predefiniti, `*args`, `**kwargs`, separatori posizionali/per parola chiave, annotazioni di ritorno               | Funzioni, classi, classi nidificate, metodi decorati; ricevitori bound convenzionali omessi                   |
| Ruby       | `.rb`                                       | Parametri nominati; nessun tipo di ritorno inferito                                                                                  | Classi, moduli, metodi di istanza/singleton, costanti                                                         |
| Rust       | `.rs`                                       | Pattern tipizzati, `self`/`&self`/`&mut self`, binding mutabili, tipi di ritorno espliciti                                           | Moduli inline e nidificati, campi nominati di struct, metodi di trait e metodi di implementazione             |
| Scala      | `.scala`, `.sc`                             | Parametri tipizzati nominati e tipi di ritorno espliciti                                                                             | Classi, trait, oggetti, metodi, enum, pacchetti                                                               |
| TypeScript | `.ts`, `.tsx`, `.mts`, `.cts`               | Parametri tipizzati nominati e annotazioni di ritorno esplicite                                                                      | Funzioni, arrow function, interfacce, classi, campi, namespace; TSX utilizza la propria grammatica            |

## Selezione del linguaggio per le intestazioni

Tutti i comandi CLI di elaborazione del codice sorgente accettano `--header-language c|cpp`. La scelta si applica solo ai file `.h`. Altre estensioni mantengono i loro parser usuali.

```bash
codeknit parse ./src --header-language cpp
codeknit fingerprint ./src --header-language c
codeknit graph analyze ./src --header-language cpp
```

Il valore predefinito è `c`, che preserva l'estrazione di macro, typedef e union in C. Utilizzare `cpp` per le intestazioni C++ denominate `.h`; i file denominati `.hpp` e `.hxx` utilizzano già C++. La scelta è esplicita perché entrambe le grammatiche possono accettare parte della sintassi dell'altro linguaggio senza estrarne l'intera struttura.

Le guardie di intestazione e i rami condizionali sono attraversati senza valutare le loro condizioni. I risultati possono quindi includere dichiarazioni da rami mutuamente esclusivi. L'espansione delle macro, la risoluzione degli include e i flag di compilazione non sono applicati. La TUI utilizza la selezione predefinita C per i file `.h`.

## Preservare la forma dell'API

Ad esempio, in Go:

```go
func Pair(a, b int, rest ...string) (string, error)
```

produce:

```text
Pair(a: int, b: int, rest: ...string) -> (string, error)
```

I metodi di interfaccia di Go e i campi nominati di Go/Rust appaiono come simboli sia in modalità normale che di fingerprint. Metodi e campi mantengono il loro tipo proprietario. I moduli Rust nidificati mantengono nomi qualificati come `service.storage.read`, anche quando un altro modulo dichiara una funzione con lo stesso nome.

Rust consente campi e metodi con lo stesso nome. I nomi dei campi con scope portano un suffisso `#field` internamente in modo che i loro ID e gli archi di contenimento rimangano distinti; i nomi visualizzati e le firme mantengono il nome originale del campo.

Le firme Python mantengono annotazioni, valori predefiniti, splat e separatori di convenzione di chiamata. Un parametro convenzionale primo `self` o `cls` è omesso da un metodo bound, ma preservato in una funzione libera o in un metodo statico.

Le firme multilinea mantengono le interruzioni di riga del sorgente in JSON. SKT esegue l'escape delle interruzioni di riga come `\n` e `\r` in modo che ogni simbolo rimanga su una riga fisica di output.

## Limiti noti

- I dichiaratori complessi di puntatori, riferimenti, array e puntatori a funzione in C/C++ possono ancora avere firme incomplete. Le dichiarazioni generate da macro non sono espanse.
- La gestione dei costruttori in TypeScript/JavaScript e alcune forme di parametri opzionali, rest o destrutturati sono incomplete.
- I vincoli avanzati dei generici e le forme di tipo al di fuori dei casi testati variano a seconda del linguaggio. I tipi di ritorno non sono inferiti.
- Le macro Rust non sono espanse. Le dichiarazioni di moduli out-of-line sono registrate; i loro file devono essere inclusi nella scansione. I campi delle tuple, i dettagli delle varianti di enum e gli elementi associati hanno una copertura parziale.
- L'estrazione della sintassi non risolve il dispatch dinamico, i sovraccarichi, le dipendenze importate o la semantica della configurazione di compilazione con la precisione del compilatore.
- Errori di sintassi parziali possono ancora omettere dichiarazioni. Ispezionare gli avvisi riportati e verificare il sorgente prima di applicare un refactoring.

Una modifica più completa dell'estrazione cambia il conteggio dei simboli e gli ShortID. Rigenerare l'output dello skeleton esistente durante l'aggiornamento invece di mescolarlo con l'output precedente.
