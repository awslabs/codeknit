---
title: Comandi Graph
description: Visualizza e analizza la struttura della tua codebase con algoritmi di grafo.
---

Usa i sottocomandi `graph` per vedere come il tuo codice è connesso, trovare problemi strutturali e decidere quali file rifattorizzare per primi. Questa guida presuppone che tu abbia [installato codeknit](../../installation/) e possa eseguire [`codeknit parse`](../parse-command/).

| Attività | Comando |
| --- | --- |
| Esplora il grafo visivamente | `codeknit graph show` |
| Trova problemi strutturali | `codeknit graph analyze` |
| Trova file volatili e importanti | `codeknit graph hotspots` |

## Visualizza il grafo del codice

`graph show` analizza l'input e scrive un file HTML autonomo. I simboli (funzioni, classi, tipi) sono nodi. Le relazioni (chiamate, contiene, implementa) sono archi.

```bash
codeknit graph show ./myproject
```

Il file viene scritto in `./skeleton/codeknit-graph.html` e si apre nel browser predefinito.

Per scegliere il percorso di output o includere file di test:

```bash
codeknit graph show ./myproject -o graph.html
codeknit graph show ./src --collect-test
```

| Flag | Default | Scopo |
| --- | --- | --- |
| `-o`, `--output` | `./skeleton/codeknit-graph.html` | Percorso del file HTML di output |
| `--collect-test` | `false` | Includi file di test |
| `--workers` | `0` (`NumCPU`) | Massimo numero di goroutine di parsing concorrenti |
| `--verbose` | `false` | Stampa informazioni di avanzamento |

### Ispeziona relazioni non risolte

Il visualizzatore disegna solo le relazioni risolte. Mostra un conteggio delle relazioni incerte (`unresolved` o `ambiguous`) ma non le disegna né i loro collegamenti candidati.

Per vedere i singoli archi incerti, esegui `parse` con `--edges`:

```bash
codeknit parse ./src --output-mode inline --edges
```

Aggiungi `--format json` per l'output JSON. Consulta il [riferimento del formato di output](../../reference/output-format/) per vedere come vengono scritti gli archi e i candidati.

## Analizza la struttura e trova problemi

`graph analyze` esegue 22 algoritmi di grafo e scrive un report `.skt` leggibile da LLM.

```bash
codeknit graph analyze ./myproject
```

Il report viene scritto in `./skeleton/graph_analysis.skt`. Leggilo direttamente o forniscilo a un LLM insieme al sorgente rilevante.

### Personalizza il report

Aumenta una soglia per vedere meno risultati, o abbassala per vederne di più:

```bash
codeknit graph analyze ./myproject -o analysis.skt --fan-threshold 15
codeknit graph analyze ./myproject --top-n 50
codeknit graph analyze ./src --collect-test
```

`--top-n 0` rimuove il limite sulle sezioni classificate.

| Flag | Default | Scopo |
| --- | --- | --- |
| `-o`, `--output` | `./skeleton/graph_analysis.skt` | Percorso del file `.skt` di output |
| `--collect-test` | `false` | Includi file di test |
| `--workers` | `0` (`NumCPU`) | Massimo numero di goroutine di parsing concorrenti |
| `--verbose` | `false` | Stampa informazioni di avanzamento |
| `--fan-threshold` | `10` | Minimo fan-in o fan-out per segnalare un simbolo hub |
| `--god-threshold` | `15` | Minimo conteggio di archi di contenimento per segnalare una god class o funzione |
| `--max-inheritance-depth` | `5` | Segnala catene di ereditarietà più profonde di questo valore |
| `--top-n` | `30` | Limita le sezioni classificate; `0` significa nessun limite |
| `--betweenness-threshold` | `0.001` | Centralità di betweenness minima da riportare |
| `--propagation-cutoff` | `0.05` | Punteggio euristico minimo per continuare la propagazione delle modifiche |

### Sappi quale risultato risponde a quale domanda

| Algoritmo | Usalo per trovare |
| --- | --- |
| Dipendenze cicliche (Tarjan's SCC) | Gruppi di simboli che dipendono l'uno dall'altro in un ciclo |
| Rilevamento hub | Simboli con accoppiamento fan-in o fan-out elevato |
| Rilevamento orfani | Candidati per dead code |
| Rilevamento god class/funzione | Simboli che contengono un numero eccessivo di figli |
| Metrica di instabilità | Simboli e pacchetti fragili o stabili, usando Ce/(Ca+Ce) |
| Catene di ereditarietà profonde | Gerarchie più profonde di `--max-inheritance-depth` |
| Centralità di betweenness | Simboli collo di bottiglia attraverso cui passano molti percorsi di dipendenza |
| Punti di articolazione | Singoli punti di fallimento la cui rimozione divide il grafo |
| PageRank | Simboli che sono ricorsivamente importanti |
| Fan-in transitivo | Raggio d'impatto: tutto ciò che dipende da un simbolo, direttamente o indirettamente |
| Simulazione di propagazione delle modifiche | Quanto lontano è probabile che si diffonda una modifica a un simbolo |
| Dipendenze cicliche tra pacchetti | Cicli di dipendenza a livello di pacchetto |
| Rilevamento violazioni di layer | Dipendenze che attraversano i layer nella direzione sbagliata |
| Raggiungibilità da punti di ingresso | Codice che nessun punto di ingresso raggiunge |
| Componenti debolmente connessi | Cluster isolati di codice |
| Peso delle dipendenze | Quanto fortemente i pacchetti sono accoppiati |
| Distanza dalla Main Sequence | Pacchetti la cui astrattezza e instabilità sono sbilanciate |
| Rilevamento shotgun surgery | Modifiche che toccherebbero molti punti in un pacchetto |
| Rilevamento feature envy | Simboli che usano un altro pacchetto più del proprio |
| Violazioni di dipendenza stabile | Pacchetti stabili che dipendono da pacchetti meno stabili |
| Violazioni di segregazione delle interfacce | Interfacce i cui membri sono per lo più inutilizzati dai loro utenti |
| Profondità di contenimento | Simboli profondamente nidificati |

### Leggi i risultati correttamente

Queste regole influenzano quanto puoi fidarti di un risultato:

- **Le relazioni incerte sono escluse.** Le metriche di dipendenza usano solo archi risolti. Il report conta le relazioni `unresolved` e `ambiguous` che ha escluso.
- **I metadati di file e include sono ignorati.** I simboli `meta/file` e `meta/include` non influenzano le classifiche di dipendenza o i risultati di dead code.
- **Fan-in e fan-out contano vicini distinti.** Il contenimento non conta come dipendenza.
- **La profondità di ereditarietà copre tutte le basi e le interfacce.** Le gerarchie cicliche non vengono assegnate a una profondità.
- **La raggiungibilità parte da callable `main`, `Main` o `init`** e segue gli archi di dipendenza. Un membro raggiungibile rende rilevante il suo contenitore, ma non i suoi membri fratelli. Se la scansione non trova alcun punto di ingresso riconosciuto, i risultati di raggiungibilità vengono omessi.
- **I callback sono riferimenti, non chiamate.** Un argomento o un callable restituito non dimostra che il callback venga eseguito.
- **I punteggi sono euristici.** La propagazione delle modifiche usa il percorso di dipendenza più forte. I suoi punteggi e le regole di layer non sono probabilità misurate o garanzie del compilatore.

### Agisci sui risultati in sicurezza

Tratta ogni risultato come un candidato per la revisione. Prima di modificare il codice, apri i simboli che il report nomina e verifica:

- Chiamanti esterni al di fuori della directory scansionata.
- Dispatch dinamico.
- Dipendenze che non hai scansionato.

L'analisi è basata sulla sintassi, quindi uno qualsiasi di questi può cambiare la conclusione. Consulta [la risoluzione delle relazioni nel riferimento del supporto linguistico](../../reference/language-support/#relationship-resolution) per le regole e i limiti di risoluzione.

## Trova hotspot consapevoli della storia

`graph hotspots` classifica i file che cambiano spesso e si trovano in posizioni strutturalmente importanti. Riporta anche l'accoppiamento temporale: file che cambiano ripetutamente negli stessi commit. Eseguilo su una directory all'interno di un repository Git.

```bash
codeknit graph hotspots ./myproject
```

Il report viene scritto in `./skeleton/hotspots.skt`.

Combina due gruppi di metriche:

- **Storia:** frequenza dei commit, churn delle righe aggiunte e cancellate, e recenza.
- **Struttura:** PageRank a livello di file, fan-in transitivo e betweenness.

Il report include un livello di confidenza basato sul numero di commit rilevanti analizzati. Tratta le classifiche di un report a bassa confidenza come provvisorie.

### Modifica la finestra temporale e i filtri

```bash
# Analizza due anni e scrivi JSON
codeknit graph hotspots ./myproject --since 2y --format json -o hotspots.json

# Consenti commit più grandi e richiedi un accoppiamento temporale più forte
codeknit graph hotspots . --max-files-per-commit 100 --min-cochanges 5
```

I commit di merge sono esclusi a meno che non si passi `--include-merges`. I commit che toccano più file di `--max-files-per-commit` vengono saltati come commit di massa, quindi le modifiche generate o meccaniche non distorcono la classifica.

| Flag | Default | Scopo |
| --- | --- | --- |
| `-o`, `--output` | `./skeleton/hotspots.skt` | Percorso del file di output |
| `--format` | `skt` | Formato di output: `skt` o `json` |
| `--since` | `12mo` | Finestra temporale, come `180d`, `12mo` o `2y` |
| `--max-commits` | `2000` | Massimo numero di commit da ispezionare |
| `--max-files-per-commit` | `50` | Escludi i commit che modificano più file di questo numero |
| `--min-cochanges` | `3` | Minimo numero di commit condivisi per riportare l'accoppiamento temporale |
| `--top-n` | `30` | Massimo numero di risultati per sezione |
| `--include-merges` | `false` | Includi i commit di merge nelle metriche storiche |
| `--collect-test` | `false` | Includi file di test |
| `--workers` | `0` (`NumCPU`) | Massimo numero di goroutine di parsing concorrenti |
| `--verbose` | `false` | Stampa informazioni di avanzamento |

### Leggi l'output JSON

Con `--format json`, ogni hotspot ha `score`, `history_score`, `structure_score`, `recency_score`, `pagerank`, `betweenness`, `commits`, `churn` e `transitive_fan_in`. Ogni voce di accoppiamento temporale ha `left`, `right`, `strength` e `co_changes`. Il livello superiore registra anche `commits_visited`, `commits_analyzed`, `skipped_merges` e `skipped_bulk_commits`. Usa questi conteggi per verificare su quanta storia si basa la classifica.

## Controlla quali sorgenti vengono scansionati

Tutti e cinque i comandi di elaborazione dei sorgenti (`parse`, `fingerprint`, `graph show`, `graph analyze` e `graph hotspots`) accettano gli stessi flag di selezione:

| Flag | Default | Scopo |
| --- | --- | --- |
| `--include GLOB` | Tutti i sorgenti supportati | Includi i percorsi corrispondenti; ripetibile |
| `--exclude GLOB` | Nessuno | Escludi i percorsi o le directory corrispondenti; ripetibile. Le esclusioni prevalgono sulle inclusioni |
| `--max-file-size BYTES` | `10485760` (10 MiB) | Salta i file più grandi; `0` disabilita il limite |
| `--strict` | `false` | Fallisci prima di scrivere l'output se la scansione o il parsing è incompleto |

I pattern sono glob gitignore-style relativi alla directory di input. Racchiusili tra virgolette in modo che la shell non li espanda. Le inclusioni non sovrascrivono le regole `.gitignore` o il filtraggio dei test.

```bash
codeknit graph analyze ./src --include '**/*.go' --exclude '**/generated/' --strict
```

I symlink e i file non regolari vengono saltati e riportati. I problemi di scansione e parsing appaiono su stderr. Senza `--strict`, codeknit mantiene risultati parziali utilizzabili, quindi controlla stderr prima di fidarti di un report. Con `--strict`, un'analisi incompleta si interrompe prima che venga scritto qualsiasi output. Un'esecuzione in cui ogni sorgente letto fallisce fallisce sempre.

Per le garanzie esatte, consulta il [riferimento sull'affidabilità dell'analisi](../../reference/analysis-reliability/).

## Pagine correlate

- [Riferimento dei flag CLI](../../reference/cli-flags/): ogni flag per ogni comando.
- [Riferimento del formato di output](../../reference/output-format/): sintassi `.skt` e JSON.
- [Riferimento del supporto linguistico](../../reference/language-support/): limiti di estrazione e risoluzione.
- [Guida del comando parse](../parse-command/): estrai simboli e archi per singoli file.