---
title: Riferimento CLI
description: Riferimento completo per tutti i comandi e le opzioni di codeknit.
---

## codeknit

Avvia l'interfaccia utente terminale interattiva (TUI), che guida attraverso i comandi e le opzioni disponibili.

```bash
codeknit
```

## codeknit parse

Estrae informazioni strutturali dal codice sorgente in file `.skt` o JSON.

```bash
codeknit parse <input-path> [output-dir]
```

| Flag             | Tipo   | Default          | Descrizione                                                                            |
| ---------------- | ------ | ---------------- | -------------------------------------------------------------------------------------- |
| `--output-mode`  | string | `directory-flat` | Modalità di output: `inline`, `directory-flat` o `directory-tree`                      |
| `--format`       | string | `skt`            | Formato di output: `skt` o `json`                                                      |
| `--max-lines`    | int    | `500`            | Numero massimo di righe per file di output (si applica alle modalità `directory-flat` e `directory-tree`) |
| `--collect-test` | bool   | `false`          | Includi file di test nell'analisi                                                       |
| `--minify`       | bool   | `false`          | Abilita la minimizzazione dell'output basata su dizionario                             |
| `--edges`        | bool   | `false`          | Includi la sezione `[edges]` nell'output (disattivata per impostazione predefinita per risparmiare token) |
| `--clean`        | bool   | `false`          | Rimuovi i file `.skt` obsoleti dalla directory di output prima della scrittura         |
| `--workers`      | int    | `0` (NumCPU)     | Numero massimo di goroutine di parsing concorrenti                                     |
| `--verbose`      | bool   | `false`          | Mostra informazioni di avanzamento durante l'elaborazione                              |

La directory di output predefinita è `./skeleton` quando non specificata. In modalità `inline`, l'output viene scritto su stdout e non viene utilizzata alcuna directory. Con `--format json`, l'output della directory viene scritto come `codeknit.json`.

## codeknit graph show

Genera una visualizzazione interattiva del grafo in HTML della struttura della codebase.

```bash
codeknit graph show <input-path>
```

| Flag             | Tipo   | Default                          | Descrizione                                  |
| ---------------- | ------ | -------------------------------- | -------------------------------------------- |
| `-o`, `--output` | string | `./skeleton/codeknit-graph.html` | Percorso del file HTML di output             |
| `--collect-test` | bool   | `false`                          | Includi file di test nell'analisi            |
| `--workers`      | int    | `0` (NumCPU)                     | Numero massimo di goroutine di parsing concorrenti |
| `--verbose`      | bool   | `false`                          | Mostra informazioni di avanzamento durante l'elaborazione |

Il file HTML generato è autonomo e si apre automaticamente nel browser predefinito.

## codeknit graph analyze

Esegue algoritmi di analisi strutturale ed emette un report `.skt` leggibile da LLM.

```bash
codeknit graph analyze <input-path>
```

| Flag                      | Tipo    | Default                         | Descrizione                                                   |
| ------------------------- | ------- | ------------------------------- | ------------------------------------------------------------- |
| `-o`, `--output`          | string  | `./skeleton/graph_analysis.skt` | Percorso del file `.skt` di output                            |
| `--collect-test`          | bool    | `false`                         | Includi file di test nell'analisi                             |
| `--workers`               | int     | `0` (NumCPU)                    | Numero massimo di goroutine di parsing concorrenti            |
| `--verbose`               | bool    | `false`                         | Mostra informazioni di avanzamento durante l'elaborazione     |
| `--fan-threshold`         | int     | `10`                            | Fan-in o fan-out minimo per segnalare un simbolo hub          |
| `--god-threshold`         | int     | `15`                            | Conteggio minimo di archi contains per segnalare una god class/function |
| `--max-inheritance-depth` | int     | `5`                             | Segnala catene di ereditarietà più profonde di questo valore  |
| `--top-n`                 | int     | `30`                            | Limita le sezioni di output classificate; `0` significa nessun limite |
| `--betweenness-threshold` | float64 | `0.001`                         | Valore minimo di centralità betweenness da riportare          |
| `--propagation-cutoff`    | float64 | `0.05`                          | Probabilità minima per continuare la simulazione di propagazione delle modifiche |

## codeknit graph hotspots

Classifica i file utilizzando la cronologia Git e l'importanza strutturale, e riporta l'accoppiamento temporale tra file che cambiano ripetutamente insieme.

```bash
codeknit graph hotspots <input-path>
```

| Flag                     | Tipo   | Default                   | Descrizione                                      |
| ------------------------ | ------ | ------------------------- | ------------------------------------------------ |
| `-o`, `--output`         | string | `./skeleton/hotspots.skt` | Percorso del file di output                      |
| `--format`               | string | `skt`                     | Formato di output: `skt` o `json`                |
| `--since`                | string | `12mo`                    | Finestra temporale, ad esempio `180d`, `12mo` o `2y` |
| `--max-commits`          | int    | `2000`                    | Numero massimo di commit da ispezionare          |
| `--max-files-per-commit` | int    | `50`                      | Escludi i commit che modificano più file         |
| `--min-cochanges`        | int    | `3`                       | Numero minimo di commit condivisi per l'accoppiamento temporale |
| `--top-n`                | int    | `30`                      | Numero massimo di risultati per sezione del report |
| `--include-merges`       | bool   | `false`                   | Includi i commit di merge                        |
| `--collect-test`         | bool   | `false`                   | Includi file di test                             |
| `--workers`              | int    | `0` (NumCPU)              | Numero massimo di goroutine di parsing concorrenti |
| `--verbose`              | bool   | `false`                   | Mostra informazioni di avanzamento               |

## codeknit fingerprint

Rileva codice duplicato e quasi duplicato utilizzando fuzzy hashing.

```bash
codeknit fingerprint <input-path>
```

| Flag               | Tipo   | Default                       | Descrizione                                                                                                                  |
| ------------------ | ------ | ----------------------------- | ---------------------------------------------------------------------------------------------------------------------------- |
| `-o`, `--output`   | string | `./skeleton/fingerprints.skt` | Percorso del file `.skt` di output                                                                                           |
| `--min-similarity` | int    | `65`                          | Percentuale minima di similarità da riportare (0–100)                                                                        |
| `--max-similarity` | int    | `95`                          | Percentuale massima di similarità da riportare (0–100)                                                                       |
| `--show-all`       | bool   | `false`                       | Includi la sezione `[fingerprints]` con i dati grezzi dei token                                                              |
| `--rerank`         | bool   | `false`                       | Trova vicini semantici e riclassifica i candidati utilizzando embeddings Ollama (richiede `ollama serve` e `ollama pull qwen3-embedding:0.6b`) |
| `--model`          | string | `qwen3-embedding:0.6b`        | Modello di embedding Ollama da utilizzare con `--rerank`                                                                     |
| `--collect-test`   | bool   | `false`                       | Includi file di test nell'analisi                                                                                            |
| `--workers`        | int    | `0` (NumCPU)                  | Numero massimo di goroutine di parsing concorrenti                                                                           |
| `--verbose`        | bool   | `false`                       | Mostra informazioni di avanzamento durante l'elaborazione                                                                    |

## codeknit completion

Genera script di completamento della shell per le shell supportate.

```bash
codeknit completion <shell>
```

Shell supportate: `bash`, `zsh`, `fish`, `powershell`.

## Flag globali

| Flag           | Descrizione                       |
| -------------- | --------------------------------- |
| `--version`    | Mostra informazioni sulla versione |
| `--help`, `-h` | Mostra aiuto per il comando corrente |