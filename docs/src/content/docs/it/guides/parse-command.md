---
title: Comando Parse
description: Estrai informazioni strutturali dal codice sorgente in file .skt o JSON.
---

`codeknit parse` legge i file sorgente e scrive i loro simboli (e, opzionalmente, le relazioni tra essi) come file `.skt` o JSON. Questa guida copre le situazioni che è più probabile incontrare. Per l'elenco completo dei flag condivisi da tutti i comandi, consulta la [riferimento dei flag CLI](../../reference/cli-flags/). Per la grammatica dell'output, consulta il [riferimento del formato di output](../../reference/output-format/).

## Analizza un progetto per la prima volta

```bash
codeknit parse ./src
```

codeknit scrive file `.skt` suddivisi in `./skeleton/`: `map_001.skt`, `map_002.skt`, e così via. Apri i chunk selettivamente invece di caricare tutto in una volta.

Per scrivere altrove, passa la directory di output come secondo argomento:

```bash
codeknit parse ./src ./output
```

Non utilizzare la directory sorgente, o una delle sue directory antenate, come directory di output.

## Scegli una modalità di output

Seleziona una modalità con `--output-mode`.

| Modalità          | Usala quando                                    | Risultato                                                                         |
| ----------------- | ---------------------------------------------- | ------------------------------------------------------------------------------ |
| `directory-flat`  | Analisi della maggior parte dei progetti (predefinita) | `map_001.skt`, `map_002.skt`, ...                                              |
| `directory-tree`  | Vuoi cercare la struttura di un singolo file   | Un file `.skt` per ogni file sorgente, rispecchiando l'albero delle sorgenti    |
| `inline`          | Analisi di un singolo file, o piping verso un altro strumento | Output su stdout; nessuna directory di output viene utilizzata                  |

In modalità tree i nomi dei file mantengono l'estensione sorgente: `src/foo.c` diventa `src/foo.c.skt`, e `src/foo.h` diventa `src/foo.h.skt`. I file di grandi dimensioni sono suddivisi in `foo.c_part1.skt`, `foo.c_part2.skt`, e così via.

```bash
codeknit parse ./src --output-mode directory-tree
```

Evita `inline` su un intero progetto. Scrive tutto su stdout e può inondare il terminale o la finestra di contesto di un LLM.

Per ulteriori informazioni sulla scelta di una modalità, consulta [Modalità di output](../output-modes/).

## Esegui nuovamente sullo stesso progetto

codeknit rifiuta di scrivere in una directory che contiene già output generato per il formato selezionato. Questo impedisce la miscelazione di chunk obsoleti e nuovi. Aggiungi `--clean` per sostituire l'output precedente:

```bash
codeknit parse ./src --clean
```

`--clean` non elimina nulla in anticipo. codeknit genera il nuovo output in una directory di staging e sostituisce i vecchi file generati solo dopo che la generazione ha successo. Se la generazione fallisce, o il processo viene interrotto, l'output precedente viene mantenuto o ripristinato al successivo avvio. I file nella directory di output che codeknit non ha generato vengono preservati. Se uno di essi entra in conflitto con un percorso generato, l'esecuzione fallisce.

In modalità JSON, le stesse regole si applicano a `codeknit.json`.

**Insidia: chunk misti.** Gli ID dei simboli (`S1`, `S2`, ...) e i risultati delle relazioni possono cambiare tra un'esecuzione e l'altra e tra versioni diverse. Dopo l'aggiornamento alla 0.5.0, rigenera l'intero set di output con `--clean`. Non combinare mai chunk provenienti da esecuzioni diverse.

## Analizza un singolo file

Punta `parse` al file e usa la modalità `inline`:

```bash
codeknit parse ./src/main.go --output-mode inline
```

Quando selezioni esplicitamente un file sorgente regolare, codeknit salta il filtraggio dei file di ignore e dei file di test per esso. I pattern di inclusione/esclusione e il limite di dimensione si applicano ancora.

## Ottieni JSON per gli script

```bash
codeknit parse ./src --output-mode inline --format json --edges
```

Il documento JSON ha i campi di primo livello `files`, `symbols`, `edges` (omessi quando non ci sono), e `errors` (omessi quando non ci sono). Un piccolo esempio:

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

Per scrivere il JSON su disco invece che su stdout, ometti `--output-mode inline`. Nelle modalità directory, il JSON viene scritto in `codeknit.json`. JSON è un singolo documento e non è suddiviso per conteggio delle righe.

## Includi le relazioni

Per impostazione predefinita, l'output contiene solo la sezione `[symbols]`, il che fa risparmiare token. Aggiungi `--edges` quando hai bisogno di sapere cosa chiama, contiene, eredita o fa riferimento a cosa:

```bash
codeknit parse ./src --edges
```

### Leggi relazioni incerte

Codeknit analizza la sintassi senza un compilatore, quindi alcune relazioni non possono essere stabilite. Ogni arco indica quanto è certo:

```skt
S8 --references--> S5
S10 --references[unresolved]--> string, bool
S11 --calls[ambiguous]--> Helper [candidates=S12, S15]
```

- `--kind-->` è una relazione risolta.
- `--kind[unresolved]-->` significa che codeknit non è riuscito a stabilire il target.
- `--kind[ambiguous]-->` significa che più di un target rimane possibile. Gli ID dei candidati sono elencati quando disponibili.

I candidati sono possibili target, non dipendenze confermate. Gli endpoint noti utilizzano ShortID. Un target non risolto senza un simbolo corrispondente può apparire come un nome, relativo al file sorgente.

In JSON, gli archi incerti portano `resolution` (`"unresolved"` o `"ambiguous"`) e, quando disponibili, `candidates` come ID completi dei simboli. Omettono `to_short`.

Per C e C++, le direttive `#include` diventano simboli `meta/file` e `meta/include`, quindi i loro archi hanno ShortID. Questo registra la direttiva come scritta. Non significa che codeknit abbia trovato l'header.

Le metriche del grafo escludono gli archi incerti e riportano quanti sono stati esclusi. Il visualizzatore HTML del grafo mostra un conteggio delle relazioni incerte e non le disegna. Per ispezionare singoli archi incerti, usa `--edges` su `parse`. Consulta [Comandi del grafo](../graph-commands/) per questi report.

La grammatica degli archi è nel [riferimento del formato di output](../../reference/output-format/). Le regole e i limiti di risoluzione sono nel [riferimento del supporto linguistico](../../reference/language-support/#relationship-resolution).

## Ridurre la dimensione dell'output

Per comprimere i token ripetuti, aggiungi `--minify`:

```bash
codeknit parse ./src --minify
```

Questo scrive un dizionario in `dict.skt` (o una sezione `[dict]` inline). Sostituisci i codici come `d0` e `d1` con i loro valori del dizionario durante la lettura. Anche i tipi di relazione sono codificati, ma le etichette di risoluzione come `[unresolved]` rimangono leggibili.

Per controllare la dimensione dei chunk, usa `--max-lines` (predefinito `500`):

- In modalità `directory-flat`, è un obiettivo. Un blocco sorgente non viene mai suddiviso, quindi un chunk può superarlo di un blocco.
- In modalità `directory-tree`, è un limite rigido per ogni chunk sorgente. Il minimo è `3`. Ogni chunk ripete le sue intestazioni `[symbols]` e `## source-path`.

## Analizza header C++ denominati `.h`

Codeknit analizza i file `.h` come C per impostazione predefinita. Per gli header C++ con questa estensione, imposta `--header-language cpp`:

```bash
codeknit parse ./src --header-language cpp
```

Il flag influisce solo sui file `.h`. I file denominati `.hpp` e `.hxx` sono già analizzati come C++, e le altre estensioni mantengono i loro parser usuali.

## Controlla quali file vengono scansionati

Tutti e cinque i comandi di elaborazione del codice sorgente accettano questi flag. Sono elencati per intero nel [riferimento dei flag CLI](../../reference/cli-flags/).

```bash
codeknit parse ./src --include '**/*.go' --exclude '**/generated/' --strict
```

- `--include GLOB` e `--exclude GLOB` sono pattern ripetibili in stile gitignore relativi alla directory di input. Racchiusili tra virgolette in modo che la shell non li espanda. Le esclusioni hanno la precedenza. Le inclusioni non sovrascrivono `.gitignore`, il filtraggio dei file di test o le esclusioni integrate.
- `--max-file-size BYTES` salta i file più grandi. Il valore predefinito è `10485760` (10 MiB). `0` disabilita il limite.
- `--strict` fallisce prima di scrivere qualsiasi output se la scansione o l'analisi è incompleta. Il valore predefinito è `false`.

Codeknit salta e segnala i symlink e i file non regolari. Le regole `.gitignore` degli antenati si applicano indipendentemente dalla directory di lavoro corrente.

**Insidia: risultati parziali.** Percorsi illeggibili, file di dimensioni eccessive, symlink saltati e avvisi di sintassi rendono un'analisi incompleta. Senza `--strict`, codeknit mantiene i risultati utilizzabili e segnala i problemi:

- su stderr,
- nel campo `errors` del JSON, o
- nella sezione SKT `[errors]` (inline, o in `warnings.skt` nelle modalità directory).

Un input senza sorgenti supportate selezionate, o un'esecuzione in cui ogni sorgente letta fallisce, esce sempre senza successo. Per garanzie esatte, consulta [Affidabilità dell'analisi](../../reference/analysis-reliability/).

## Flag di parse

Questi flag sono specifici di `parse`. I flag condivisi sono nel [riferimento dei flag CLI](../../reference/cli-flags/).

| Flag               | Predefinito      | Descrizione                                                                                       |
| ------------------ | ---------------- | ------------------------------------------------------------------------------------------------- |
| `--output-mode`    | `directory-flat` | Modalità di output: `inline`, `directory-flat`, o `directory-tree`.                               |
| `--format`         | `skt`            | Formato di output: `skt` o `json`.                                                                |
| `--header-language`| `c`              | Parser per i file `.h`: `c` o `cpp`. Le altre estensioni rimangono invariate.                     |
| `--max-lines`      | `500`            | Obiettivo di righe per l'output flat; limite rigido (minimo 3) per ogni chunk sorgente tree.      |
| `--collect-test`   | `false`          | Includi i file di test.                                                                           |
| `--minify`         | `false`          | Usa la compressione basata su dizionario.                                                         |
| `--edges`          | `false`          | Includi la sezione `[edges]` (disattivata per impostazione predefinita per risparmiare token).    |
| `--clean`          | `false`          | Sostituisci l'output generato precedentemente, solo dopo che la generazione ha successo.          |
| `--workers`        | `NumCPU`         | Massimo numero di goroutine di parsing concorrenti.                                               |
| `--verbose`        | `false`          | Stampa informazioni di avanzamento.                                                               |

## Suggerimenti

- Usa `directory-flat` per la maggior parte dei progetti e leggi solo i chunk di cui hai bisogno.
- Usa `--minify` su codebase di grandi dimensioni per risparmiare token.
- Passa `--edges` solo quando hai bisogno dei dati sulle relazioni.
- Controlla la [matrice delle capacità testate e i limiti noti](../../reference/language-support/) prima di affidarti all'estrazione per uno specifico linguaggio.

## Vedi anche

- [Modalità di output](../output-modes/)
- [Riferimento del formato di output](../../reference/output-format/)
- [Affidabilità dell'analisi](../../reference/analysis-reliability/)
- [Comandi del grafo](../graph-commands/)