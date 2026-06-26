---
title: Comando Parse
description: Estrae informazioni strutturali dal codice sorgente in file .skt o JSON.
---

Il comando `codeknit parse` estrae informazioni strutturali dalla tua codebase — come funzioni, classi, metodi, variabili e le loro relazioni — e le emette in formato compatto `.skt` per impostazione predefinita. Usa JSON quando hai bisogno di output leggibile da macchine per script, integrazioni o strumenti downstream.

## Utilizzo di base

```bash
codeknit parse <input-path> [output-dir]
```

- **`<input-path>`**: Percorso della directory o del file che si desidera analizzare.
- **`[output-dir]`**: Directory di output opzionale. Se non specificata, il valore predefinito è `./skeleton`.

### Esempi

```bash
# Analizza un progetto, output nella directory predefinita ./skeleton
codeknit parse ./src

# Analizza e scrive in una directory di output personalizzata
codeknit parse ./src ./output

# Analizza un singolo file e invia l'output a stdout
codeknit parse ./src/main.go --output-mode inline

# Emette JSON leggibile da macchine su stdout
codeknit parse ./src --output-mode inline --format json
```

## Modalità di output

Usa `--output-mode` per controllare come viene strutturato l'output. Sono disponibili tre modalità:

| Modalità            | Descrizione                                                                              | Ideale per                                            |
| ------------------- | ---------------------------------------------------------------------------------------- | ----------------------------------------------------- |
| `directory-flat`    | Scrive file `.skt` suddivisi (es. `map_001.skt`, `map_002.skt`) nella directory di output. | ✅ **La maggior parte dei progetti** — modalità predefinita e consigliata |
| `directory-tree`    | Rispecchia la struttura della directory sorgente, creando un file `.skt` per ogni file sorgente. | Navigare l'output insieme al codice sorgente             |
| `inline`            | Invia tutto l'output a stdout.                                                           | Singoli file o piping verso altri strumenti               |

> **Suggerimento**: Usa `directory-flat` come predefinito a meno che tu non stia lavorando con un singolo file. Evita `inline` per input di grandi dimensioni poiché può sovraccaricare le finestre di contesto.

## Flag

| Flag               | Predefinito       | Descrizione                                                                  |
| ------------------ | ----------------- | ---------------------------------------------------------------------------- |
| `--output-mode`    | `directory-flat`  | Modalità di output: `inline`, `directory-flat` o `directory-tree`            |
| `--format`         | `skt`             | Formato di output: `skt` o `json`                                            |
| `--max-lines`      | `500`             | Numero massimo di righe per file di output in modalità flat/tree             |
| `--collect-test`   | `false`           | Includi i file di test nell'analisi                                          |
| `--minify`         | `false`           | Abilita la compressione basata su dizionario per ridurre l'uso di token      |
| `--edges`          | `false`           | Includi la sezione `[edges]` con i dati delle relazioni (chiamate, contiene, ecc.) |
| `--clean`          | `false`           | Rimuovi i file `.skt` esistenti nella directory di output prima della scrittura |
| `--workers`        | `NumCPU`          | Numero massimo di goroutine di parsing concorrenti (0 = usa tutti i core CPU) |
| `--verbose`        | `false`           | Stampa informazioni di avanzamento e temporizzazione durante l'elaborazione  |

## Pattern comuni

```bash
# Prima esecuzione su un progetto
codeknit parse ./src
```

```bash
# Riesegui e pulisci l'output precedente
codeknit parse ./src --clean
```

```bash
# Analizza un singolo file su stdout
codeknit parse ./src/main.go --output-mode inline
```

```bash
# Minimizza l'output per codebase di grandi dimensioni
codeknit parse ./src --minify
```

```bash
# Includi gli archi delle relazioni (es. per l'analisi delle dipendenze)
codeknit parse ./src --edges
```

```bash
# Emetti JSON per un altro strumento
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
# Rispecchia la struttura della directory sorgente nell'output
codeknit parse ./src --output-mode directory-tree
```

## Protezione contro output obsoleto

Se la directory di output contiene già file `.skt` da una precedente esecuzione, `codeknit` rifiuterà di scrivere nuovo output per evitare di mescolare dati obsoleti e freschi.

Per sovrascrivere questo comportamento e pulire la directory di output prima della scrittura, usa il flag `--clean`:

```bash
codeknit parse ./src --clean
```

Questo garantisce un set di output fresco e coerente.

## Suggerimenti

- ✅ **Usa `directory-flat` come predefinito** per la maggior parte dei progetti. Offre un buon equilibrio tra leggibilità e gestibilità.
- 🔍 Usa `--minify` su codebase di grandi dimensioni per ridurre l'uso di token tramite un dizionario condiviso (`dict.skt`).
- 🔗 La sezione `[edges]` è **esclusa per impostazione predefinita** per risparmiare token. Usa `--edges` quando hai bisogno di dati sulle relazioni come `calls`, `contains` o `inherits`.
- 🧾 Usa `--format json` quando uno script o un'integrazione necessita di dati strutturati invece di `.skt`.
- 🧹 Usa sempre `--clean` quando riesegui l'analisi sulla stessa directory di output.
- 📁 Usa `directory-tree` se desideri correlare i file `.skt` direttamente con i file sorgente nel tuo editor.