---
title: Comando Fingerprint
description: Rileva codice duplicato e quasi duplicato tra file e linguaggi utilizzando fuzzy hashing.
---

Il comando `codeknit fingerprint` rileva codice duplicato e quasi duplicato nella tua codebase utilizzando **Context-Triggered Piecewise Hashing (CTPH)**. Funziona tra file e persino tra linguaggi di programmazione normalizzando nomi di variabili, stringhe letterali e annotazioni di tipo prima di calcolare le impronte strutturali.

## Cosa fa

`codeknit fingerprint` analizza ogni funzione, metodo, variabile e tipo nella tua codebase e calcola un'**impronta strutturale normalizzata** basata su:

- Flusso di controllo (`if`, `for`, `while`, `switch`)
- Operazioni (`=`, `+`, `==`, `&&`, `||`)
- Chiamate, return, assegnazioni e creazione di oggetti
- Costrutti del linguaggio come `try/catch`, `yield`, `await`, `defer`

Questa normalizzazione significa che **copia-incolla rinominato**, **refactoring banali** e **logica equivalente in linguaggi diversi** possono ancora essere rilevati come duplicati.

L'algoritmo utilizza **CTPH** (una variante di rolling hash) per trovare in modo efficiente i quasi duplicati. Codice simile produce impronte simili, consentendo il matching fuzzy anche quando il codice è stato leggermente modificato.

## Utilizzo di base

```bash
codeknit fingerprint ./src
```

Questo comando:

- Analizza tutti i file sorgente in `./src`
- Calcola le impronte strutturali
- Invia i risultati a `./skeleton/fingerprints.skt`
- Segnala corrispondenze con similarità tra **65% e 95%** (intervallo predefinito)

## Flag

| Flag               | Default                       | Description                                                                                                                                                |
| ------------------ | ----------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `-o`, `--output`   | `./skeleton/fingerprints.skt` | Percorso del file `.skt` di output                                                                                                                         |
| `--min-similarity` | `65`                          | Percentuale minima di similarità da segnalare (0–100)                                                                                                      |
| `--max-similarity` | `95`                          | Percentuale massima di similarità da segnalare (0–100)                                                                                                     |
| `--show-all`       | `false`                       | Include la sezione `[fingerprints]` con i dati grezzi dei token                                                                                            |
| `--rerank`         | `false`                       | Trova vicini semantici e riordina i candidati utilizzando embeddings Ollama (richiede: `ollama serve` e `ollama pull qwen3-embedding:0.6b`) |
| `--model`          | `qwen3-embedding:0.6b`        | Modello di embedding Ollama da utilizzare con `--rerank`                                                                                                   |
| `--collect-test`   | `false`                       | Includi i file di test nell'analisi                                                                                                                        |
| `--workers`        | `NumCPU`                      | Numero massimo di goroutine di parsing concorrenti (0 = usa tutti i core della CPU)                                                                        |
| `--verbose`        | `false`                       | Stampa informazioni di avanzamento durante l'elaborazione                                                                                                  |

## Formato di output

L'output è un file `.skt` con le seguenti sezioni:

### `[duplicates]` (sempre presente)

Elenca coppie di simboli con similarità superiore alla soglia:

```skt
[duplicates]
similarity:96%  pkg/user.go::GetUser <-> pkg/admin.go::GetAdmin
similarity:88%  utils/str.go::TrimSpaces <-> lib/text.go::CleanString
```

Ogni riga mostra:

- Percentuale di similarità
- Simbolo sinistro (percorso del file, scope, nome)
- Simbolo destro (percorso del file, scope, nome)

### `[fingerprints]` (solo con `--show-all`)

Contiene i dati grezzi delle impronte per ogni simbolo:

```skt
[fingerprints]
validateToken  FP:3:a1b2c3...:d4e5f6...  tokens:8e0f1a2b...
```

Campi:

- Nome del simbolo
- `FP:<versione>:<hash1>:<hash2>` — impronta CTPH
- `tokens:<hex>` — flusso di token del corpo normalizzato

Questa sezione è utile per il debug o per la creazione di strumenti downstream.

## Pattern comuni

```bash
# Scansione predefinita
codeknit fingerprint /codeknit/it/src
```

```bash
# Trova solo duplicati esatti
codeknit fingerprint ./src --min-similarity 100
```

```bash
# Trova codice moderatamente simile (ad es. stesso algoritmo, nomi diversi)
codeknit fingerprint ./src --min-similarity 50 --max-similarity 80
```

```bash
# Utilizza il matching semantico per trovare candidati aggiuntivi e ridurre i falsi positivi
# Richiede: ollama serve && ollama pull qwen3-embedding:0.6b
codeknit fingerprint ./src --rerank
```

```bash
# Utilizza un modello di embedding diverso per il matching semantico
codeknit fingerprint ./src --rerank --model qwen3-embedding:4b
```

```bash
# Output completo dell'elenco delle impronte (per strumenti di analisi)
codeknit fingerprint ./src --show-all
```

```bash
# File di output personalizzato
codeknit fingerprint ./src -o duplicates.skt
```

## Scelta di un intervallo di similarità

| Intervallo | Guida                                                                                     |
| ---------- | ----------------------------------------------------------------------------------------- |
| 96–100%    | Duplicati strutturali esatti o quasi esatti. Quasi certamente copia-incolla.              |
| 85–95%     | Quasi duplicati. Solitamente copia-incolla con modifiche minori (ad es. variabili rinominate, logging aggiunto). |
| 65–84%     | Intervallo predefinito. Forte similarità strutturale. Buoni candidati per il refactoring. |
| 50–64%     | Similarità moderata. Stessa forma algoritmica ma dettagli diversi. Da rivedere manualmente. |
| < 50%      | Solitamente rumore. Duplicazione non significativa.                                       |

## Suggerimenti

- **Le impronte misurano la struttura, non il significato**: Un punteggio di similarità elevato significa che il codice _sembra_ simile, non che _fa_ la stessa cosa. Rivedi sempre entrambi i simboli.
- **Usa `--rerank` per il matching semantico**: Gli embeddings aggiungono vicini semantici che il retrieval strutturale può perdere e filtrano i candidati che non concordano semanticamente.
- **I corpi brevi vengono saltati**: I simboli con meno di 4 token normalizzati (ad es. semplici getter) vengono ignorati per evitare rumore.
- **Funziona il matching cross-language**: Costrutti equivalenti (ad es., una funzione Python e una funzione Go con la stessa logica) possono corrispondere, ma i pattern specifici del linguaggio possono produrre corrispondenze spurie a bassa similarità.
- **Una corrispondenza è un segnale, non una sentenza**: Tratta ogni corrispondenza come un invito a investigare — non come prova automatica di duplicazione.