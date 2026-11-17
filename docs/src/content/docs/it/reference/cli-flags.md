---
title: Riferimento CLI
description: Riferimento completo per tutti i comandi e i flag di codeknit.
---

Questa pagina elenca ogni comando di codeknit, i suoi argomenti e i suoi flag. Per istruzioni orientate alle attività, consulta la [guida al comando parse](../../guides/parse-command/), la [guida ai comandi graph](../../guides/graph-commands/) e la [guida al comando fingerprint](../../guides/fingerprint-command/).

## Riepilogo dei comandi

| Comando                                       | Scopo                                                        |
| --------------------------------------------- | -------------------------------------------------------------- |
| `codeknit`                                    | Avvia l'interfaccia utente terminale interattiva.             |
| `codeknit parse <input-path> [output-dir]`    | Estrae simboli e relazioni in output `.skt` o JSON.           |
| `codeknit graph show <input-path>`            | Genera un grafo HTML interattivo.                              |
| `codeknit graph analyze <input-path>`         | Esegue algoritmi di analisi del grafo strutturale e scrive un report `.skt`. |
| `codeknit graph hotspots <input-path>`        | Classifica i punti caldi di modifica utilizzando la cronologia Git e la struttura del grafo. |
| `codeknit fingerprint <input-path>`           | Rileva codice duplicato e quasi duplicato.                     |
| `codeknit completion <shell>`                 | Genera uno script di completamento per la shell.               |

Cinque comandi elaborano il codice sorgente: `parse`, `graph show`, `graph analyze`, `graph hotspots` e `fingerprint`. Condividono i flag descritti in [Flag condivisi per l'elaborazione del sorgente](#flag-condivisi-per-lelaborazione-del-sorgente).

## Flag globali

| Flag              | Tipo | Default | Descrizione                  |
| ----------------- | ---- | ------- | ---------------------------- |
| `--version`       | bool | `false` | Stampa le informazioni sulla versione. |
| `-h`, `--help`    | bool | `false` | Mostra l'aiuto per il comando. |

## Flag condivisi per l'elaborazione del sorgente

Questi flag appaiono nella tabella di ogni comando di elaborazione del sorgente.

| Flag                    | Tipo                | Default               | Valori consentiti       | Descrizione                                                                          |
| ----------------------- | ------------------- | --------------------- | -------------------- | ------------------------------------------------------------------------------------ |
| `--header-language`     | string              | `c`                   | `c`, `cpp`           | Parser utilizzato per i file `.h`.                                                   |
| `--include GLOB`        | string, ripetibile  | Tutti i sorgenti supportati | Glob stile gitignore | Include i percorsi di sorgente corrispondenti. Più pattern formano un'unione.        |
| `--exclude GLOB`        | string, ripetibile  | Nessuno               | Glob stile gitignore | Esclude i percorsi o le directory corrispondenti.                                    |
| `--max-file-size BYTES` | int                 | `10485760` (10 MiB)   | `0` o maggiore       | Salta i file di sorgente più grandi di questa dimensione. `0` disabilita il limite.  |
| `--strict`              | bool                | `false`               | —                    | Termina con errore prima della scrittura dell'output se si verificano diagnostiche di scansione o parsing. |

### `--header-language`

Si applica solo ai file `.h`. Altre estensioni mantengono il loro parser abituale (`.hpp` e `.hxx` utilizzano già C++).

- `c` (predefinito) preserva l'estrazione di macro C, typedef e union.
- `cpp` seleziona il parser C++. Utilizzalo per progetti C++ i cui header sono denominati `.h`.

L'interfaccia utente terminale utilizza la selezione predefinita `c` per i file `.h`.

```bash
codeknit parse ./src --header-language cpp
```

Per le capacità di estrazione testate e i limiti noti per lingua, consulta [Supporto Linguistico e Fedeltà API](../language-support/). Quella pagina descrive i costrutti testati, non la copertura semantica completa del linguaggio o del compilatore.

### Selezione del sorgente

- I pattern utilizzano glob separati da slash, stile gitignore, relativi alla directory di input. Per un input a file singolo, i pattern corrispondono al basename del file.
- Racchiudi i pattern tra virgolette per evitare che la shell li espanda.
- Le esclusioni hanno la precedenza sulle inclusioni.
- Le inclusioni non sovrascrivono le regole `.gitignore`, il filtraggio dei file di test o le esclusioni integrate.
- I pattern di negazione (`!pattern`) non sono accettati in `--include` o `--exclude`.
- Le scansioni delle directory rispettano i file `.gitignore` antenati e nidificati indipendentemente dalla directory di lavoro corrente.
- La selezione esplicita di un file di sorgente regolare bypassa il filtraggio di ignore e test. Si applicano comunque i limiti di inclusione, esclusione e dimensione.
- I file di test sono esclusi a meno che non sia impostato `--collect-test`.

```bash
codeknit parse ./src --include '**/*.go' --exclude '**/generated/' --strict
```

### Diagnostiche e `--strict`

- I symlink, i symlink a directory e i file non regolari vengono saltati e segnalati.
- Percorsi illeggibili, letture fallite di file ignore, sorgenti di dimensioni eccessive e avvisi di sintassi rendono un'analisi incompleta.
- Per impostazione predefinita, vengono mantenuti risultati parziali utilizzabili. Le diagnostiche appaiono su stderr. `parse` le registra anche in `errors` JSON o in SKT `[errors]`.
- Con `--strict`, un'analisi incompleta viene rifiutata prima della scrittura dell'output.
- Un'esecuzione senza sorgenti supportati selezionati, o in cui ogni lettura del sorgente fallisce, termina sempre senza successo.
- L'interfaccia utente terminale applica la stessa politica per i symlink e il limite predefinito di 10 MiB. Le sue forme non espongono i flag di selezione e strict.

Consulta [Affidabilità dell'Analisi](../analysis-reliability/) per garanzie esatte.

## `codeknit`

```bash
codeknit
```

L'esecuzione di `codeknit` senza sottocomandi avvia l'interfaccia utente terminale interattiva. L'interfaccia offre i comandi `parse`, `graph show`, `graph analyze`, `graph hotspots` e `fingerprint` e le loro opzioni.

## `codeknit parse`

```bash
codeknit parse <input-path> [output-dir]
```

| Argomento      | Obbligatorio | Default      | Descrizione                                                              |
| -------------- | ------------ | ------------ | ------------------------------------------------------------------------ |
| `input-path`   | Sì           | —            | Directory di sorgente o singolo file di sorgente.                        |
| `output-dir`   | No           | `./skeleton` | Directory di output. Non utilizzata quando `--output-mode` è `inline`.  |

| Flag                    | Tipo                | Default            | Valori consentiti                                   | Descrizione                                                                                         |
| ----------------------- | ------------------- | ------------------ | ------------------------------------------------ | --------------------------------------------------------------------------------------------------- |
| `--output-mode`         | string              | `directory-flat`   | `inline`, `directory-flat`, `directory-tree`     | Dove e come viene scritto l'output.                                                                |
| `--format`              | string              | `skt`              | `skt`, `json`                                    | Formato di output.                                                                                  |
| `--max-lines`           | int                 | `500`              | Per `directory-tree`, 3 o maggiore               | Obiettivo di righe per l'output flat. Limite rigido per chunk di sorgente nell'output tree.         |
| `--collect-test`        | bool                | `false`            | —                                                | Include i file di test.                                                                             |
| `--minify`              | bool                | `false`            | —                                                | Abilita la compressione dell'output basata su dizionario.                                          |
| `--edges`               | bool                | `false`            | —                                                | Include la sezione `[edges]`. Disattivata per impostazione predefinita per risparmiare token.      |
| `--clean`               | bool                | `false`            | —                                                | Sostituisce l'output generato precedentemente, solo dopo una generazione riuscita.                  |
| `--workers`             | int                 | `0` (`NumCPU`)     | `0` o maggiore                                   | Massimo numero di goroutine di parsing concorrenti.                                                 |
| `--verbose`             | bool                | `false`            | —                                                | Stampa informazioni di avanzamento.                                                                |
| `--header-language`     | string              | `c`                | `c`, `cpp`                                       | Parser utilizzato per i file `.h`.                                                                 |
| `--include GLOB`        | string, ripetibile  | Tutti i sorgenti supportati | —                                             | Include i percorsi di sorgente corrispondenti.                                                     |
| `--exclude GLOB`        | string, ripetibile  | Nessuno            | —                                                | Esclude i percorsi o le directory corrispondenti.                                                  |
| `--max-file-size BYTES` | int                 | `10485760`         | `0` disabilita il limite                          | Salta i file di sorgente più grandi.                                                                |
| `--strict`              | bool                | `false`            | —                                                | Termina con errore prima della scrittura dell'output se si verificano diagnostiche di scansione o parsing. |

### Modalità di output

| Valore             | Comportamento                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------ |
| `inline`          | Scrive su stdout. Non viene utilizzata alcuna directory di output.                                                    |
| `directory-flat`  | Scrive file suddivisi nella directory di output. I blocchi di sorgente rimangono intatti, quindi `--max-lines` è un obiettivo di suddivisione piuttosto che un limite rigido. |
| `directory-tree`  | Rispecchia l'albero dei sorgenti. I nomi mantengono l'estensione del sorgente (`foo.c.skt`, `foo.h.skt`). I file suddivisi utilizzano `foo.c_part1.skt`. |

### Comportamento dell'output

- Senza `--clean`, l'output generato esistente per il formato selezionato nella directory di output causa un errore.
- Con `--clean`, i file `.skt` obsoleti (formato SKT) o `codeknit.json` (formato JSON) vengono sostituiti solo dopo una generazione riuscita. Le voci non correlate vengono preservate.
- Nel formato JSON, le modalità directory scrivono `codeknit.json`. JSON è un singolo documento e non viene suddiviso per conteggio delle righe.
- `dict.skt` e `warnings.skt` sono nomi di file ausiliari riservati.
- Non utilizzare una directory di sorgente o uno dei suoi antenati come directory di output.

Consulta [Modalità di output](../../guides/output-modes/) e il [riferimento del formato di output](../output-format/) per i dettagli.

```bash
codeknit parse ./src
codeknit parse ./src ./out --output-mode directory-tree
codeknit parse ./src --output-mode inline --format json --edges
```

## `codeknit graph show`

```bash
codeknit graph show <input-path>
```

| Argomento      | Obbligatorio | Descrizione                              |
| -------------- | ------------ | ---------------------------------------- |
| `input-path`   | Sì           | Directory di sorgente o singolo file di sorgente. |

| Flag                    | Tipo                | Default                            | Valori consentiti          | Descrizione                                                       |
| ----------------------- | ------------------- | ---------------------------------- | ----------------------- | ----------------------------------------------------------------- |
| `-o`, `--output`        | string              | `./skeleton/codeknit-graph.html`   | Percorso file            | File HTML di output.                                              |
| `--collect-test`        | bool                | `false`                            | —                       | Include i file di test.                                           |
| `--workers`             | int                 | `0` (`NumCPU`)                     | `0` o maggiore          | Massimo numero di goroutine di parsing concorrenti.               |
| `--verbose`             | bool                | `false`                            | —                       | Stampa informazioni di avanzamento.                               |
| `--header-language`     | string              | `c`                                | `c`, `cpp`              | Parser utilizzato per i file `.h`.                                |
| `--include GLOB`        | string, ripetibile  | Tutti i sorgenti supportati        | —                       | Include i percorsi di sorgente corrispondenti.                    |
| `--exclude GLOB`        | string, ripetibile  | Nessuno                            | —                       | Esclude i percorsi o le directory corrispondenti.                 |
| `--max-file-size BYTES` | int                 | `10485760`                         | `0` disabilita il limite | Salta i file di sorgente più grandi.                              |
| `--strict`              | bool                | `false`                            | —                       | Termina con errore prima della scrittura dell'output se si verificano diagnostiche di scansione o parsing. |

Il comando scrive un file HTML autonomo e lo apre nel browser predefinito. Il visualizzatore disegna le relazioni risolte e mostra un conteggio delle relazioni incerte.

```bash
codeknit graph show ./src -o graph.html
```

## `codeknit graph analyze`

```bash
codeknit graph analyze <input-path>
```

| Argomento      | Obbligatorio | Descrizione                              |
| -------------- | ------------ | ---------------------------------------- |
| `input-path`   | Sì           | Directory di sorgente o singolo file di sorgente. |

| Flag                       | Tipo                | Default                            | Valori consentiti          | Descrizione                                                  |
| -------------------------- | ------------------- | ---------------------------------- | ----------------------- | ------------------------------------------------------------ |
| `-o`, `--output`           | string              | `./skeleton/graph_analysis.skt`    | Percorso file            | File `.skt` di output.                                       |
| `--collect-test`           | bool                | `false`                            | —                       | Include i file di test.                                      |
| `--workers`                | int                 | `0` (`NumCPU`)                     | `0` o maggiore          | Massimo numero di goroutine di parsing concorrenti.          |
| `--verbose`                | bool                | `false`                            | —                       | Stampa informazioni di avanzamento.                          |
| `--fan-threshold`          | int                 | `10`                               | —                       | Soglia minima di fan-in o fan-out per segnalare un simbolo hub. |
| `--god-threshold`          | int                 | `15`                               | —                       | Conteggio minimo di archi `contains` per segnalare una god class o funzione. |
| `--max-inheritance-depth`  | int                 | `5`                                | —                       | Segnala catene di ereditarietà più profonde di questo valore. |
| `--top-n`                  | int                 | `30`                               | `0` = nessun limite      | Limita le sezioni di output classificate.                    |
| `--betweenness-threshold`  | float               | `0.001`                            | —                       | Valore minimo di betweenness centrality da riportare.        |
| `--propagation-cutoff`     | float               | `0.05`                             | —                       | Probabilità minima per continuare la propagazione delle modifiche. |
| `--header-language`        | string              | `c`                                | `c`, `cpp`              | Parser utilizzato per i file `.h`.                           |
| `--include GLOB`           | string, ripetibile  | Tutti i sorgenti supportati        | —                       | Include i percorsi di sorgente corrispondenti.               |
| `--exclude GLOB`           | string, ripetibile  | Nessuno                            | —                       | Esclude i percorsi o le directory corrispondenti.            |
| `--max-file-size BYTES`    | int                 | `10485760`                         | `0` disabilita il limite | Salta i file di sorgente più grandi.                          |
| `--strict`                 | bool                | `false`                            | —                       | Termina con errore prima della scrittura dell'output se si verificano diagnostiche di scansione o parsing. |

Le metriche di dipendenza escludono le relazioni incerte e i metadati di file/include. Il report registra i conteggi delle relazioni escluse.

```bash
codeknit graph analyze ./src -o analysis.skt --fan-threshold 15 --top-n 50
```

## `codeknit graph hotspots`

```bash
codeknit graph hotspots <input-path>
```

| Argomento      | Obbligatorio | Descrizione                                                  |
| -------------- | ------------ | ------------------------------------------------------------ |
| `input-path`   | Sì           | Directory di sorgente all'interno di un repository Git.      |

| Flag                       | Tipo                | Default                    | Valori consentiti                              | Descrizione                                                  |
| -------------------------- | ------------------- | -------------------------- | ------------------------------------------- | ------------------------------------------------------------ |
| `-o`, `--output`           | string              | `./skeleton/hotspots.skt`  | Percorso file                                | File di output.                                              |
| `--format`                 | string              | `skt`                      | `skt`, `json`                                | Formato di output.                                           |
| `--since`                  | string              | `12mo`                     | Finestra come `180d`, `12mo` o `2y`         | Finestra temporale.                                          |
| `--max-commits`            | int                 | `2000`                     | —                                           | Massimo numero di commit da ispezionare.                    |
| `--max-files-per-commit`   | int                 | `50`                       | —                                           | Esclude i commit che toccano più file di questo valore.      |
| `--min-cochanges`          | int                 | `3`                        | —                                           | Minimo numero di commit condivisi per riportare l'accoppiamento temporale. |
| `--top-n`                  | int                 | `30`                       | —                                           | Massimo numero di risultati per sezione.                     |
| `--include-merges`         | bool                | `false`                    | —                                           | Include i merge commit nelle metriche della cronologia.      |
| `--collect-test`           | bool                | `false`                    | —                                           | Include i file di test.                                      |
| `--workers`                | int                 | `0` (`NumCPU`)             | `0` o maggiore                              | Massimo numero di goroutine di parsing concorrenti.          |
| `--verbose`                | bool                | `false`                    | —                                           | Stampa informazioni di avanzamento.                          |
| `--header-language`        | string              | `c`                        | `c`, `cpp`                                  | Parser utilizzato per i file `.h`.                           |
| `--include GLOB`           | string, ripetibile  | Tutti i sorgenti supportati | —                                           | Include i percorsi di sorgente corrispondenti.               |
| `--exclude GLOB`           | string, ripetibile  | Nessuno                    | —                                           | Esclude i percorsi o le directory corrispondenti.            |
| `--max-file-size BYTES`    | int                 | `10485760`                 | `0` disabilita il limite                    | Salta i file di sorgente più grandi.                          |
| `--strict`                 | bool                | `false`                    | —                                           | Termina con errore prima della scrittura dell'output se si verificano diagnostiche di scansione o parsing. |

Il comando combina la frequenza dei commit, il churn delle righe e la recenza con PageRank a livello di file, fan-in transitivo e betweenness centrality. Riporta anche l'accoppiamento temporale tra file che cambiano insieme. I merge commit sono esclusi per impostazione predefinita.

```bash
codeknit graph hotspots ./src --since 2y --format json -o hotspots.json
```

## `codeknit fingerprint`

```bash
codeknit fingerprint <input-path>
```

| Argomento      | Obbligatorio | Descrizione                              |
| -------------- | ------------ | ---------------------------------------- |
| `input-path`   | Sì           | Directory di sorgente o singolo file di sorgente. |

| Flag                    | Tipo                | Default                          | Valori consentiti          | Descrizione                                                                                                      |
| ----------------------- | ------------------- | -------------------------------- | ----------------------- | ---------------------------------------------------------------------------------------------------------------- |
| `-o`, `--output`        | string              | `./skeleton/fingerprints.skt`    | Percorso file            | File `.skt` di output.                                                                                           |
| `--min-similarity`      | int                 | `65`                             | `0`–`100`               | Percentuale minima di similarità da riportare.                                                                   |
| `--max-similarity`      | int                 | `95`                             | `0`–`100`               | Percentuale massima di similarità da riportare.                                                                 |
| `--show-all`            | bool                | `false`                          | —                       | Emette anche la sezione `[fingerprints]` con i dati grezzi dei token.                                           |
| `--rerank`              | bool                | `false`                          | —                       | Trova vicini semantici e riclassifica i candidati con embeddings Ollama. Richiede `ollama serve` e `ollama pull qwen3-embedding:0.6b`. |
| `--model`               | string              | `qwen3-embedding:0.6b`           | Nome modello Ollama      | Modello di embedding Ollama da utilizzare con `--rerank`.                                                       |
| `--collect-test`        | bool                | `false`                          | —                       | Include i file di test.                                                                                          |
| `--workers`             | int                 | `0` (`NumCPU`)                   | `0` o maggiore          | Massimo numero di goroutine di parsing concorrenti.                                                              |
| `--verbose`             | bool                | `false`                          | —                       | Stampa informazioni di avanzamento.                                                                              |
| `--header-language`     | string              | `c`                              | `c`, `cpp`              | Parser utilizzato per i file `.h`.                                                                               |
| `--include GLOB`        | string, ripetibile  | Tutti i sorgenti supportati      | —                       | Include i percorsi di sorgente corrispondenti.                                                                   |
| `--exclude GLOB`        | string, ripetibile  | Nessuno                          | —                       | Esclude i percorsi o le directory corrispondenti.                                                                |
| `--max-file-size BYTES` | int                 | `10485760`                       | `0` disabilita il limite | Salta i file di sorgente più grandi.                                                                              |
| `--strict`              | bool                | `false`                          | —                       | Termina con errore prima della scrittura dell'output se si verificano diagnostiche di scansione o parsing.       |

### Intervallo di similarità

L'intervallo predefinito, 65–95%, esclude le corrispondenze al 100%. Per riportare le corrispondenze esatte dei token normalizzati, imposta entrambi i limiti a `100`:

```bash
codeknit fingerprint ./src --min-similarity 100 --max-similarity 100
```

I candidati con token normalizzati esatti sono completi. Con 500 o più simboli idonei, altri candidati quasi duplicati utilizzano il recupero approssimativo dei primi 50 vicini. `--rerank` può rimuovere un candidato strutturalmente esatto. Consulta [Affidabilità dell'Analisi](../analysis-reliability/#garanzie-di-rilevamento-dei-duplicati) per i dettagli.

## `codeknit completion`

```bash
codeknit completion <shell>
```

| Argomento | Obbligatorio | Valori consentiti                         | Descrizione                                   |
| --------- | ------------ | ----------------------------------------- | --------------------------------------------- |
| `shell`   | Sì           | `bash`, `zsh`, `fish`, `powershell`       | Shell per cui generare uno script di completamento. |

Lo script viene scritto su stdout.

```bash
codeknit completion zsh >> ~/.zshrc
```

## Pagine correlate

- [Riferimento del formato di output](../output-format/)
- [Supporto Linguistico e Fedeltà API](../language-support/)
- [Affidabilità dell'Analisi](../analysis-reliability/)