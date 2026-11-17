---
title: Modalità di output
description: Scegli la modalità di output giusta per le dimensioni del tuo progetto e il flusso di lavoro.
---

codeknit separa due decisioni: dove va l'output (la **modalità di output**) e come viene codificato (il **formato di output**). Questa pagina spiega perché esiste ogni opzione, il suo costo e quando è adatta. Per la grammatica `.skt`, consulta il [riferimento del formato di output](../../reference/output-format/). Per i comandi, consulta la [guida del comando parse](../parse-command/).

## Due scelte indipendenti

Il flag `--output-mode` seleziona una tra tre modalità: `directory-flat`, `directory-tree` o `inline`. Il flag `--format` seleziona `skt` (il valore predefinito) o `json`. Qualsiasi modalità può essere combinata con uno qualsiasi dei formati.

Le due scelte sono separate perché rispondono a domande diverse. La modalità decide come l'output viene suddiviso e localizzato sul disco, il che è importante per una persona o un LLM che lo legge. Il formato decide chi lo consuma: un LLM o un lettore umano, oppure uno script.

## directory-flat

`directory-flat` è la modalità predefinita. Scrive file di chunk numerati, `map_001.skt`, `map_002.skt`, e così via, nella directory `./skeleton/` a meno che non si specifichi un'altra directory di output.

**Perché esiste.** Una codebase di grandi dimensioni produce più struttura di quanta ne possa entrare in una singola finestra di contesto. La suddivisione in chunk consente a un lettore di aprire solo i file di cui ha bisogno invece di caricare tutto.

**Come funziona la suddivisione.** codeknit mantiene intatto il blocco di ogni file sorgente e inizia un nuovo chunk tra i blocchi una volta che un chunk raggiunge circa `--max-lines` (predefinito 500). Il limite è un obiettivo, non un vincolo rigido: un file sorgente con un blocco molto grande può produrre un chunk più lungo di 500 righe. Mantenere i blocchi intatti significa che un simbolo non viene mai separato dai suoi fratelli nello stesso file.

**Compromesso.** I chunk non sono organizzati per percorso sorgente. Per trovare la struttura di un file, si cercano i chunk per l'intestazione `## source-path`. Se si cercano spesso file specifici, `directory-tree` è più adatto.

**Minificazione.** Con `--minify`, il dizionario viene scritto in `dict.skt` accanto ai chunk. Vedi [Minificazione](#minificazione).

## directory-tree

`directory-tree` rispecchia la struttura della directory sorgente sotto la directory di output. Ogni file sorgente ottiene un file `.skt` nel percorso corrispondente.

**Nomenclatura.** codeknit aggiunge `.skt` al nome completo del file sorgente invece di sostituire l'estensione. `src/foo.c` diventa `src/foo.c.skt`, e `src/foo.h` diventa `src/foo.h.skt`. Senza questa regola, entrambi verrebbero mappati a `foo.skt` e colliderebbero. Quando l'output di un file viene suddiviso, le parti sono nominate `foo.c_part1.skt`, `foo.c_part2.skt`, e così via. I percorsi di output in conflitto, comprese le collisioni solo per maiuscole/minuscole, falliscono visibilmente invece di sovrascriversi a vicenda.

**La suddivisione differisce dalla modalità flat.** Nella modalità tree, `--max-lines` è un limite rigido su ogni chunk sorgente, con un minimo di 3. Ogni chunk di simboli ripete le intestazioni `[symbols]` e `## source-path` in modo che possa essere letto da solo. I chunk solo archi portano `[edges]`. Il limite si applica ai chunk sorgente, non ai file ausiliari come `dict.skt` e `warnings.skt`.

**Compromesso.** L'output tree è facile da navigare insieme al sorgente, ma produce molti file piccoli. Un progetto di grandi dimensioni può generarne migliaia, e un lettore che vuole una panoramica ampia deve aprire molti file.

## inline

`inline` scrive tutto su stdout. Non crea una directory di output né un'area di staging.

**Perché esiste.** È adatto a due casi: ispezionare rapidamente un singolo file e inviare l'output a un altro strumento tramite pipe. È l'unica modalità in cui nulla viene scritto su disco.

**Compromesso.** Nulla viene suddiviso. L'intero risultato finisce nel terminale o nel contesto del consumatore in una volta sola, quindi sovraccarica un lettore su qualsiasi cosa oltre a una manciata di file. Trattala come appropriata per un singolo file o un progetto molto piccolo, circa meno di cinque file. Poiché stdout non può essere riportato indietro, un consumatore che fallisce può vedere un output parziale. Le modalità directory non hanno questo problema.

## Formati di output

| Formato | Ideale per | Output |
| --- | --- | --- |
| `skt` | Contesto LLM e ispezione umana | File `.skt` o stdout |
| `json` | Script e integrazione strutturata | `codeknit.json` nelle modalità directory, o stdout nella modalità inline |

**skt** è compatto e orientato alle righe. Ogni simbolo si trova su una riga fisica, con i salti di riga nelle firme escapati, quindi un lettore può caricare parte dell'output senza rompere la sua struttura. Questo è il motivo per cui la suddivisione in chunk funziona per `.skt`.

**json** esiste per i consumatori che necessitano di campi tipizzati piuttosto che di testo da leggere. È un singolo documento con `files`, `symbols` e, quando sono inclusi gli archi, `edges`. Poiché è un unico documento, non viene suddiviso per conteggio delle righe. Nelle modalità directory `--max-lines` quindi non ha effetto su JSON, e un progetto di grandi dimensioni produce un unico grande `codeknit.json`.

Un esempio compatto:

```json
{
  "files": ["app.go"],
  "symbols": [
    {"id": "app.go::User", "short_id": "S1", "name": "User", "file": "app.go",
     "category": "type", "kind": "struct", "signature": "type User struct", "span": [3, 3]},
    {"id": "app.go::Save", "short_id": "S2", "name": "Save", "file": "app.go",
     "category": "callable", "kind": "function", "signature": "Save(u: S1)", "span": [5, 5]}
  ],
  "edges": [
    {"from": "app.go::Save", "from_short": "S2", "to": "app.go::User",
     "to_short": "S1", "kind": "references"}
  ]
}
```

## Scegliere una modalità

| Modalità | Ideale per | Output |
| --- | --- | --- |
| `directory-flat` | La maggior parte dei progetti (predefinita) | `./skeleton/map_001.skt`, `map_002.skt`, ... |
| `directory-tree` | Navigare insieme al sorgente | `./skeleton/<percorso rispecchiato>.skt` |
| `inline` | Singolo file, piping | stdout |

Regole pratiche:

- **In caso di dubbio**, usa `directory-flat`, la modalità predefinita.
- **Un singolo file**: `inline` va bene.
- **Qualsiasi cosa oltre a pochi file**: usa `directory-flat` o `directory-tree`.
- **Codebase di grandi dimensioni**: aggiungi `--minify`.
- **Esecuzione ripetuta nella stessa directory**: aggiungi `--clean`.

## Staging, riesecuzioni e `--clean`

Le modalità directory non scrivono nella directory di output mentre procedono. codeknit genera l'output in una directory di staging sorella e lo pubblica solo dopo che la generazione ha avuto successo. Un'esecuzione fallita lascia quindi il tuo output precedente inalterato.

Per impostazione predefinita, codeknit rifiuta di scrivere quando esiste già un output generato per il formato selezionato. Il motivo è che gli ID dei simboli (`S1`, `S2`, ...) possono cambiare ogni volta che cambia l'estrazione. Mescolare chunk di esecuzioni diverse darebbe archi che puntano a simboli sbagliati. `--clean` accetta la sostituzione: dopo una generazione riuscita, i file `.skt` obsoleti (in modalità SKT) o `codeknit.json` (in modalità JSON) vengono sostituiti. Le voci non correlate nella directory vengono preservate. Se una di esse collide con un percorso generato, l'esecuzione fallisce invece di sovrascriverla.

La pubblicazione prende un lock specifico per la directory, esegue il backup dell'output precedente e lo ripristina se la sostituzione fallisce. Se il processo viene interrotto tra i due rinomini di directory, la prossima esecuzione ripristina il backup prima di generare qualsiasi cosa. Questo si riprende da un fallimento del processo. Non rende l'aggiornamento della directory istantaneo: c'è un breve intervallo tra i rinomini, quindi un lettore che necessita di accesso ininterrotto dovrebbe coordinarsi con la generazione. Inoltre, non protegge da fallimenti di archiviazione o perdita di alimentazione.

Non scegliere una directory sorgente, o un suo antenato, come directory di output. Per le garanzie complete di pubblicazione, consulta il [riferimento sull'affidabilità dell'analisi](../../reference/analysis-reliability/).

## Minificazione

`--minify` applica la compressione basata su dizionario. codeknit sostituisce token ripetuti, come `callable/method` o `exported`, con codici brevi (`d0`, `d1`, ...). La mappatura è memorizzata in una sezione `[dict]`, che nelle modalità directory è il file `dict.skt`.

Il risparmio deriva dalla ripetizione. La maggior parte delle righe dei simboli condivide le stesse categorie, tipi e proprietà, quindi una codebase di grandi dimensioni ripete queste stringhe migliaia di volte. Il costo è l'indirezione: un lettore deve espandere i codici usando il dizionario, e un singolo chunk non è più autosufficiente. Il dizionario è un output ausiliario separato, quindi non conta rispetto a `--max-lines`.

L'output minificato codifica il tipo di relazione ma mantiene leggibili i marcatori di risoluzione, quindi gli archi `unresolved` e `ambiguous` rimangono distinguibili. Per progetti piccoli i risparmi sono modesti e il passaggio aggiuntivo raramente vale la pena. Per quelli grandi, i risparmi in token possono essere significativi.

## Pagine correlate

- [Guida al comando parse](../parse-command/): i comandi per ogni modalità.
- [Riferimento del formato di output](../../reference/output-format/): la grammatica `.skt` e JSON.
- [Riferimento dei flag CLI](../../reference/cli-flags/): ogni flag e valore predefinito.
- [Riferimento sull'affidabilità dell'analisi](../../reference/analysis-reliability/): denominazione dell'output e garanzie di pubblicazione.