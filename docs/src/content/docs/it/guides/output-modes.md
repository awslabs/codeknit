---
title: Modalità di output
description: Scegli la modalità di output giusta per le dimensioni del tuo progetto e il tuo flusso di lavoro.
---

codeknit supporta tre modalità di output, controllate dal flag `--output-mode`. Ogni modalità determina come la struttura del codice estratta viene scritta su disco (o su stdout).

### directory-flat (predefinita, consigliata)

- **Comportamento**: Scrive file `.skt` suddivisi come `map_001.skt`, `map_002.skt`, ecc.
- **Directory di output**: `./skeleton/` per impostazione predefinita
- **Suddivisione**: I file vengono suddivisi quando superano il limite `--max-lines` (predefinito: 500 righe)
- **Caso d'uso**: Migliore per la maggior parte dei progetti. Mantiene l'output organizzato e leggibile limitando la dimensione dei file. Puoi leggere solo i chunk rilevanti per il tuo compito.
- **Minificazione**: Quando `--minify` è abilitato, viene generato anche un file `dict.skt` nella directory di output, contenente le mappature dei token per i valori compressi.

Esempio:

```bash
codeknit parse ./src
# Output: ./skeleton/map_001.skt, map_002.skt, ...
```

### directory-tree

- **Comportamento**: Rispecchia esattamente la struttura della directory sorgente.
- **Directory di output**: `./skeleton/` per impostazione predefinita
- **Mappatura**: Viene creato un file `.skt` per ogni file sorgente, nello stesso percorso corrispondente.
- **Caso d'uso**: Ideale quando si desidera cercare rapidamente la struttura di un file specifico. Utile per la navigazione insieme al codice sorgente originale.

Esempio:

```bash
codeknit parse ./src --output-mode directory-tree
# Output: ./skeleton/src/handler.skt, ./skeleton/pkg/db.skt, ecc.
```

### inline

- **Comportamento**: Stampa tutto l'output su stdout.
- **Directory di output**: Nessuna creata
- **Caso d'uso**: Consigliato solo per file singoli o progetti molto piccoli (meno di 5 file). Utile quando si invia l'output a un altro strumento o si ispeziona un singolo file in modo interattivo.

Esempio:

```bash
codeknit parse ./src/main.go --output-mode inline
# Output: stampato direttamente sul terminale
```

### Tabella di decisione

| Modalità         | Migliore per                                             | Posizione dell'output                                         |
| ---------------- | -------------------------------------------------------- | ------------------------------------------------------------- |
| `directory-flat` | La maggior parte dei progetti (predefinita, consigliata) | `./skeleton/map_001.skt`, `map_002.skt`, ...                  |
| `directory-tree` | Navigazione dell'output insieme al codice sorgente       | `./skeleton/<percorso speculare>.skt`                         |
| `inline`         | File singolo, invio a un altro strumento                 | stdout — usare solo per file singoli o progetti molto piccoli |

### Regole pratiche

- **In caso di dubbio** → usa `directory-flat` (predefinita)
- **Ispezione di un singolo file** → `inline` è accettabile
- **Più di qualche file** → preferisci `directory-flat` o `directory-tree`
- **Codebase di grandi dimensioni** → aggiungi `--minify` per ridurre l'uso di token
- **Riesecuzione sullo stesso output** → usa `--clean` per rimuovere i file `.skt` obsoleti

### Minificazione

Il flag `--minify` abilita la compressione basata su dizionario dei token ripetuti (ad esempio, chiavi di proprietà come `exported`, `async` o nomi di tipi comuni). Quando abilitato:

- I valori ripetuti vengono sostituiti con codici brevi (`d0`, `d1`, `d2`, ...)
- Un file `dict.skt` viene scritto nella directory di output, mappando i codici ai valori originali
- Riduce significativamente la dimensione dell'output per codebase di grandi dimensioni
- Funziona sia in modalità `directory-flat` che `directory-tree`

Esempio di output minificato:

```skt
[dict]
- d0: exported
- d1: callable/function

[symbols]
## src/main.py
S1 d1 L1-L5 main() {d0}
```
