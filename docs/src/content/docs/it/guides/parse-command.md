---
title: Comando Parse
description: Estrai informazioni strutturali dal codice sorgente in file .skt o JSON.
---

Il comando `codeknit parse` estrae informazioni strutturali dalla tua codebase — come funzioni, classi, metodi, variabili e le loro relazioni — e le restituisce in formato compatto `.skt` per impostazione predefinita. Usa JSON quando hai bisogno di output leggibile dalla macchina per script, integrazioni o strumenti downstream.

## Utilizzo di base

```bash
codeknit parse <input-path> [output-dir]
```

- **`<input-path>`**: Percorso della directory o del file che si desidera analizzare.
- **`[output-dir]`**: Directory di output opzionale. Se non specificata, il valore predefinito è `./skeleton`.

### Esempi

```bash
# Parse a project, output to default directory ./skeleton
codeknit parse ./src

# Parse and write to a custom output directory
codeknit parse ./src ./output

# Parse a single file and output to stdout
codeknit parse ./src/main.go --output-mode inline

# Emit machine-readable JSON to stdout
codeknit parse ./src --output-mode inline --format json
```

## Modalità di output

Usa `--output-mode` per controllare come viene strutturato l'output. Sono disponibili tre modalità:

| Modalità         | Descrizione                                                                                      | Ideale per                                                                |
| ---------------- | ------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------- |
| `directory-flat` | Scrive file `.skt` suddivisi (es. `map_001.skt`, `map_002.skt`) nella directory di output.       | ✅ **La maggior parte dei progetti** — modalità predefinita e consigliata |
| `directory-tree` | Rispecchia la struttura della directory sorgente, creando un file `.skt` per ogni file sorgente. | Navigare l'output insieme al codice sorgente                              |
| `inline`         | Invia tutto l'output a stdout.                                                                   | File singoli o piping verso altri strumenti                               |

> **Suggerimento**: Usa `directory-flat` come predefinito a meno che tu non stia lavorando con un singolo file. Evita `inline` per input di grandi dimensioni poiché può sovraccaricare le finestre di contesto.

## Flag

| Flag             | Predefinito      | Descrizione                                                                        |
| ---------------- | ---------------- | ---------------------------------------------------------------------------------- |
| `--output-mode`  | `directory-flat` | Modalità di output: `inline`, `directory-flat` o `directory-tree`                  |
| `--format`       | `skt`            | Formato di output: `skt` o `json`                                                  |
| `--max-lines`    | `500`            | Numero massimo di righe per file di output in modalità flat/tree                   |
| `--collect-test` | `false`          | Includi i file di test nell'analisi                                                |
| `--minify`       | `false`          | Abilita la compressione basata su dizionario per ridurre l'uso di token            |
| `--edges`        | `false`          | Includi la sezione `[edges]` con i dati delle relazioni (chiamate, contiene, ecc.) |
| `--clean`        | `false`          | Rimuovi i file `.skt` esistenti nella directory di output prima della scrittura    |
| `--workers`      | `NumCPU`         | Numero massimo di goroutine di parsing concorrenti (0 = usa tutti i core CPU)      |
| `--verbose`      | `false`          | Stampa informazioni di avanzamento e temporizzazione durante l'elaborazione        |

## Pattern comuni

```bash
# First run on a project
codeknit parse ./src
```

```bash
# Re-run and clean previous output
codeknit parse ./src --clean
```

```bash
# Parse a single file to stdout
codeknit parse ./src/main.go --output-mode inline
```

```bash
# Minify output for large codebases
codeknit parse ./src --minify
```

```bash
# Include relationship edges (e.g., for dependency analysis)
codeknit parse ./src --edges
```

```bash
# Emit JSON for another tool
codeknit parse ./src --output-mode inline --format json --edges
```

Esempio di output JSON:

```json
{
  "files": ["app.go"],
  "symbols": [
    {
      "id": "app.go::User",
      "short_id": "S1",
      "name": "User",
      "file": "app.go",
      "category": "type",
      "kind": "struct",
      "signature": "type User struct",
      "span": [3, 3]
    },
    {
      "id": "app.go::Save",
      "short_id": "S2",
      "name": "Save",
      "file": "app.go",
      "category": "callable",
      "kind": "function",
      "signature": "Save(u: S1)",
      "span": [5, 5]
    }
  ],
  "edges": [
    {
      "from": "app.go::Save",
      "from_short": "S2",
      "to": "app.go::User",
      "to_short": "S1",
      "kind": "references"
    }
  ]
}
```

```bash
# Mirror source tree structure in output
codeknit parse ./src --output-mode directory-tree
```

## Incertezza nelle relazioni

Con `--edges`, Codeknit preserva le relazioni anche quando la sintassi e lo scope non stabiliscono un target univoco:

```skt
S1 --calls--> S2
S1 --references[unresolved]--> string, bool
S1 --calls[ambiguous]--> Helper [candidates=S3, S4]
```

L'assenza di uno stato significa risolto dall'analizzatore. `unresolved` indica che la dipendenza non può essere stabilita; `ambiguous` indica che rimangono interpretazioni concorrenti possibili. I candidati sono possibili target, non connessioni confermate. Gli endpoint noti usano ShortID; i target non risolti senza simboli possono apparire come nomi.

JSON trasporta le stesse informazioni in `resolution` e `candidates`, e omette `to_short` per i target incerti. C/C++ include hanno simboli `meta/file` e `meta/include` quindi gli archi usano ShortID; questo registra le direttive scritte senza affermare la risoluzione a livello di compilatore dei header.

Le metriche di dipendenza escludono le relazioni incerte. Il grafo HTML mostra il loro conteggio senza disegnare singoli collegamenti incerti. Consulta il [riferimento del formato di output](../../reference/output-format/) per la grammatica e il [riferimento del supporto linguistico](../../reference/language-support/) per i limiti di risoluzione.

## Protezione dall'output obsoleto

Se la directory di output contiene già file `.skt` da una precedente esecuzione, `codeknit` rifiuterà di scrivere nuovo output per evitare di mescolare dati obsoleti e freschi.

Per sovrascrivere questo comportamento e pulire la directory di output prima della scrittura, usa il flag `--clean`:

```bash
codeknit parse ./src --clean
```

Questo garantisce un set di output fresco e coerente.

Quando si esegue l'aggiornamento alla versione `0.5.0`, rigenera l'intero set di output. Gli ID dei simboli e i risultati delle relazioni possono cambiare; non mescolare chunk da esecuzioni diverse.

## Suggerimenti

- ✅ **Usa `directory-flat` come predefinito** per la maggior parte dei progetti. Bilancia leggibilità e gestibilità.
- 🔍 Usa `--minify` su codebase di grandi dimensioni per ridurre l'uso di token tramite un dizionario condiviso (`dict.skt`).
- 🔗 La sezione `[edges]` è **esclusa per impostazione predefinita** per risparmiare token. Usa `--edges` quando hai bisogno dei dati delle relazioni come `calls`, `contains` o `inherits`.
- 🧾 Usa `--format json` quando uno script o un'integrazione necessita di dati strutturati invece di `.skt`.
- 🧹 Usa sempre `--clean` quando riesegui l'analisi sulla stessa directory di output.
- 📁 Usa `directory-tree` se desideri correlare i file `.skt` direttamente con i file sorgente nel tuo editor.
