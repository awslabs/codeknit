---
title: Affidabilità dell'analisi
description: Selezione delle sorgenti, analisi incomplete, pubblicazione dell'output e garanzie di rilevamento dei duplicati.
---

Questi comportamenti descrivono le modifiche non rilasciate a Go/CLI successive alla versione 0.5.0.

## Selezione delle sorgenti e diagnostica

Tutti e cinque i comandi di elaborazione delle sorgenti (`parse`, `fingerprint`, `graph show`, `graph analyze` e `graph hotspots`) accettano:

| Flag                    | Default               | Significato                                                             |
| ----------------------- | --------------------- | ----------------------------------------------------------------------- |
| `--include GLOB`        | Tutte le sorgenti supportate | Includi i percorsi delle sorgenti corrispondenti; ripeti per selezionare diversi pattern.   |
| `--exclude GLOB`        | Nessuno               | Escludi percorsi o directory corrispondenti; ripeti per diversi pattern. |
| `--max-file-size BYTES` | `10485760` (10 MiB)   | Salta i file sorgente più grandi; `0` disabilita questo limite.          |
| `--strict`              | `false`               | Termina con errore prima di emettere risultati se si verificano diagnostiche di scansione o parsing.    |

I pattern utilizzano glob in stile gitignore separati da slash, relativi alla directory di input selezionata. Per un input a file singolo, i pattern corrispondono al suo basename. Racchiudi i pattern tra virgolette in modo che la shell non li espanda. Gli include formano un'unione; gli exclude hanno la precedenza. Gli include non sovrascrivono `.gitignore`, il filtro dei test o le esclusioni integrate. I pattern di negazione (`!pattern`) non sono accettati in questi flag; utilizza i controlli separati di include ed exclude.

```bash
codeknit parse ./src --include '**/*.go' --exclude '**/generated/' --strict
codeknit fingerprint ./src --max-file-size 2097152 --include '**/*.ts'
codeknit parse ./src/app.go --output-mode inline --format json
```

Le scansioni delle directory rispettano i file `.gitignore` antenati e nidificati relativi alla loro posizione nel progetto, indipendentemente dalla directory di lavoro del processo. La directory `.git` più vicina o il marker del worktree stabilisce il confine del repository. Al di fuori di un repository, si applicano comunque i file ignore antenati. Una sottodirectory selezionata sotto una directory ignorata rimane ignorata. Selezionare esplicitamente un file sorgente regolare bypassa ignore e il filtro dei test, ma applica comunque i limiti di include/exclude e dimensione.

I symlink delle sorgenti, i symlink delle directory e le sorgenti non regolari vengono saltati e segnalati. Le letture sono confinate alla radice della scansione e il limite di dimensione viene applicato nuovamente durante la lettura, in modo che un file che cresce dopo la scoperta non possa bypassarlo. Le esclusioni integrate delle directory includono `.git`, `.hg`, `.svn`, `node_modules`, le cache Python e gli ambienti virtuali Python comuni. I file di test rimangono esclusi a meno che non venga impostato `--collect-test`.

Directory/file illeggibili, letture dei file ignore fallite, symlink saltati, sorgenti di dimensioni eccessive e avvisi di sintassi rendono un'analisi incompleta. Per impostazione predefinita, i risultati parziali utilizzabili vengono mantenuti e le diagnostiche appaiono su stderr. Il JSON di parsing include record di percorso/motivo in `errors`; SKT include `[errors]` inline o in `warnings.skt`. Anche i comandi graph/fingerprint riportano diagnostiche su stderr. I file deliberatamente ignorati o filtrati non sono avvisi. `--strict` rifiuta i risultati incompleti prima della pubblicazione dell'output. Un input senza sorgenti supportate selezionate, o un'esecuzione in cui ogni lettura della sorgente fallisce, termina sempre senza successo.

La TUI utilizza la stessa politica per i symlink e il limite predefinito di 10 MiB. Le sue forme non espongono ancora i nuovi flag di selezione e strict.

## Nomi dell'output e pubblicazione

L'output ad albero aggiunge `.skt` al nome completo della sorgente: `src/foo.c` diventa `src/foo.c.skt`, mentre `src/foo.h` diventa `src/foo.h.skt`. L'output suddiviso utilizza `src/foo.c_part1.skt`, `src/foo.c_part2.skt` e così via. I file di dizionario e diagnostica mantengono i nomi riservati `dict.skt` e `warnings.skt`. I percorsi di output in conflitto, comprese le collisioni solo per maiuscole/minuscole, falliscono visibilmente.

L'SKT ad albero richiede `--max-lines` di almeno 3. Ogni chunk di simboli ripete le intestazioni `[symbols]` e `## source-path`. I chunk solo archi riportano `[edges]`. Ogni chunk di sorgente rispetta il limite; i dizionari e i file di diagnostica sono output ausiliari separati. L'SKT flat mantiene intatti i blocchi di sorgente, quindi il suo limite di righe è un obiettivo di suddivisione piuttosto che un limite rigido per un blocco di sorgente di grandi dimensioni. Il JSON è un singolo documento e non viene suddiviso per conteggio delle righe.

La generazione della directory viene completata in una directory di staging sorella. Senza `--clean`, l'output generato esistente per il formato selezionato causa un errore. Con `--clean`, i file `.skt` obsoleti (modalità SKT) o `codeknit.json` (modalità JSON) vengono sostituiti solo dopo una generazione riuscita. Le voci non correlate vengono preservate; una collisione con una di esse causa un errore. Non utilizzare una directory sorgente o un suo antenato come directory di output.

La pubblicazione mantiene un lock specifico per la directory, esegue il backup dell'output precedente e lo ripristina se la sostituzione fallisce. Se un processo termina tra i rinomini delle directory, la successiva invocazione ripristina il backup prima di generare qualsiasi cosa. Il file `.NAME.codeknit.lock` fratello rimane intenzionalmente per coordinare le future pubblicazioni; non rimuovere un lock attivo o un backup di ripristino. Esiste un breve intervallo tra i due rinomini delle directory, quindi i lettori che richiedono accesso ininterrotto dovrebbero coordinarsi con la generazione. Questo fornisce il ripristino in caso di guasto del processo, non una garanzia contro il guasto dello storage o la perdita di alimentazione.

I report individuali vengono scritti e sincronizzati su un file temporaneo fratello prima di sostituire la destinazione. I file esistenti non vengono troncati durante la generazione. Lo stdout inline non può essere ripristinato se il suo consumatore fallisce.

## Garanzie di rilevamento dei duplicati

Il matching esatto significa flussi di token normalizzati completi identici nella stessa categoria di simboli tra le dichiarazioni idonee per il fingerprinting. Non significa testo sorgente identico o equivalenza semantica provata. I gruppi di candidati esatti utilizzano i flussi di token stessi, piuttosto che l'uguaglianza dell'hash fuzzy.

Sotto i 500 simboli idonei, il rilevamento strutturale confronta tutte le coppie della stessa categoria. A 500 o più, tutte le coppie di token esatti vengono mantenute indipendentemente dal recupero dei candidati, e altre coppie utilizzano i primi 50 vicini strutturali approssimativi per simbolo. Il recupero dei quasi duplicati non è quindi esaustivo su input di grandi dimensioni. Il recupero scansiona comunque tutti i vettori; non rende l'algoritmo complessivo subquadratico.

Si applica comunque il filtraggio dell'intervallo di similarità. L'intervallo predefinito è 65–95%, che esclude le corrispondenze al 100%. Per includere le corrispondenze esatte, usa:

```bash
codeknit fingerprint ./src --min-similarity 100 --max-similarity 100
```

L'opzione `--rerank` aggiunge candidati semantici, applica l'accordo coseno e filtra in base al punteggio combinato; può rimuovere un candidato strutturalmente esatto. Il testo dell'embedding proviene dallo stesso snapshot di sorgente limitato utilizzato per l'estrazione. Input stabili e punteggi fissi producono un ordinamento deterministico, inclusi i pareggi; le risposte del modello sono input esterni.

`internal/emitter/calibration_test.go` valuta un piccolo fixture di sorgente etichettato a una soglia del 70%, riportando veri/falsi positivi e negativi, precisione e richiamo separatamente dai punteggi di similarità. Il fixture ha una coppia esatta, due coppie di quasi cloni e tre coppie non correlate, esercitate sotto e oltre il limite di 500 simboli. Questo è un controllo di regressione, non una stima rappresentativa dell'accuratezza nel mondo reale. Budget più ampi per la calibrazione multilingue e le prestazioni rimangono lavoro in roadmap.