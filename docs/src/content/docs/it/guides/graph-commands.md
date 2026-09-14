---
title: Comandi Grafici
description: Visualizza e analizza la struttura del tuo codebase con algoritmi grafici.
---

codeknit fornisce comandi grafici per visualizzare la struttura, eseguire analisi automatizzate e combinare il grafo delle dipendenze corrente con la cronologia delle modifiche Git.

## graph show

Genera una visualizzazione grafica interattiva in HTML del tuo codebase.

```bash
codeknit graph show <input-path>
```

Questo comando analizza il tuo codebase e produce un file HTML autonomo con una visualizzazione grafica interattiva. I simboli (funzioni, classi, tipi) appaiono come nodi, e le loro relazioni (chiamate, contiene, implementa) come archi. La visualizzazione si apre automaticamente nel browser predefinito.

Il visualizzatore disegna le relazioni risolte. Mostra un conteggio delle incertezze ma non disegna le singole relazioni non risolte o ambigue o i loro collegamenti candidati. Ispeziona queste relazioni con l'output del comando parse:

```bash
codeknit parse ./src --output-mode inline --edges
codeknit parse ./src --output-mode inline --format json --edges
```

### Flag

| Flag             | Default                          | Description                                  |
| ---------------- | -------------------------------- | -------------------------------------------- |
| `-o`, `--output` | `./skeleton/codeknit-graph.html` | Percorso del file HTML di output             |
| `--collect-test` | `false`                          | Includi i file di test nell'analisi          |
| `--workers`      | `NumCPU`                         | Massimo numero di goroutine di parsing concorrenti |
| `--verbose`      | `false`                          | Stampa informazioni di avanzamento durante l'elaborazione |

### Esempi

```bash
# Generate default visualization
codeknit graph show ./myproject

# Custom output file
codeknit graph show ./myproject -o graph.html

# Include test files
codeknit graph show ./src --collect-test
```

## graph analyze

Esegue algoritmi strutturali sul grafo del tuo codebase ed emette un report `.skt` leggibile da LLM contenente insight sulla qualità del codice.

```bash
codeknit graph analyze <input-path>
```

Questo comando segnala candidati per la revisione, incluse dipendenze cicliche, simboli hub, codice morto, god class e colli di bottiglia architetturali.

### Algoritmi

L'analisi include 22 algoritmi strutturali del grafo:

- Dipendenze cicliche (SCC di Tarjan)
- Rilevamento di hub (alto accoppiamento fan-in/fan-out)
- Rilevamento di orfani (candidati di codice morto)
- Rilevamento di god class/function (figli eccessivi)
- Metrica di instabilità (Ce/(Ca+Ce) di Robert C. Martin)
- Catene di ereditarietà profonde
- Centralità di betweenness (rilevamento di colli di bottiglia)
- Punti di articolazione (punti singoli di guasto)
- PageRank (importanza ricorsiva)
- Fan-in transitivo (raggio d'impatto)
- Simulazione di propagazione delle modifiche
- Dipendenze cicliche tra package
- Rilevamento di violazioni di layer
- Raggiungibilità dai punti di ingresso
- Componenti debolmente connessi
- Peso delle dipendenze (forza di accoppiamento tra package)
- Distanza dalla Main Sequence (bilanciamento A+I)
- Rilevamento di shotgun surgery
- Rilevamento di feature envy
- Violazioni di dipendenze stabili
- Violazioni di segregazione delle interfacce
- Profondità di contenimento

### Interpretazione del report

Le metriche di dipendenza utilizzano relazioni risolte. Il report conta le relazioni non risolte e ambigue escluse. I metadati di file/inclusione sono esclusi dalle classifiche delle dipendenze e dai risultati di codice morto.

- Fan-in e fan-out contano vicini distinti. Le relazioni ripetute non aumentano questi conteggi e il contenimento non è una dipendenza.
- La profondità di ereditarietà considera tutte le basi e le interfacce. Le gerarchie cicliche non hanno una profondità assegnata.
- La raggiungibilità parte da callable denominati `main`, `Main` o `init` e segue gli archi delle dipendenze. Raggiungere un membro rende il suo contenitore rilevante senza marcare ogni fratello come raggiungibile. Senza un punto di ingresso riconosciuto, i risultati di raggiungibilità sono omessi.
- Gli argomenti di callback e i callable restituiti sono riferimenti, non prova che il callback venga eseguito.
- La propagazione delle modifiche utilizza il percorso di dipendenza più forte. I suoi punteggi e le regole di layer basate su directory sono euristiche, non probabilità misurate.

I chiamanti esterni, il dispatch dinamico e le dipendenze al di fuori della scansione possono cambiare le conclusioni. Consulta [Supporto Linguistico e Fedeltà API](../../reference/language-support/) per le regole e i limiti di risoluzione.

### Flag

| Flag                      | Default                         | Description                                              |
| ------------------------- | ------------------------------- | -------------------------------------------------------- |
| `-o`, `--output`          | `./skeleton/graph_analysis.skt` | Percorso del file `.skt` di output                       |
| `--collect-test`          | `false`                         | Includi i file di test nell'analisi                      |
| `--workers`               | `NumCPU`                        | Massimo numero di goroutine di parsing concorrenti       |
| `--verbose`               | `false`                         | Stampa informazioni di avanzamento durante l'elaborazione |
| `--fan-threshold`         | `10`                            | Minimo fan-in o fan-out per segnalare un simbolo hub     |
| `--god-threshold`         | `15`                            | Minimo conteggio di archi contiene per segnalare una god class/function |
| `--max-inheritance-depth` | `5`                             | Segnala catene di ereditarietà più profonde di questo    |
| `--top-n`                 | `30`                            | Limita le sezioni di output classificate; 0 = nessun limite |
| `--betweenness-threshold` | `0.001`                         | Valore minimo di centralità di betweenness da riportare  |
| `--propagation-cutoff`    | `0.05`                          | Punteggio euristico minimo per continuare la propagazione delle modifiche |

### Esempi

```bash
# Run structural analysis with defaults
codeknit graph analyze ./myproject

# Custom output and thresholds
codeknit graph analyze ./myproject -o analysis.skt --fan-threshold 15

# Show more results per section
codeknit graph analyze ./myproject --top-n 50

# Remove the cap on ranked sections
codeknit graph analyze ./myproject --top-n 0

# Include test files
codeknit graph analyze ./src --collect-test
```

## graph hotspots

Classifica i file che sono sia frequentemente modificati che strutturalmente importanti:

```bash
codeknit graph hotspots <input-path>
```

Il punteggio combina frequenza dei commit, churn delle righe e recenza con PageRank a livello di file, fan-in transitivo e centralità di betweenness. Il report identifica anche l'accoppiamento temporale tra file che cambiano ripetutamente negli stessi commit.

I commit di merge sono esclusi per impostazione predefinita. Anche i commit che modificano più di 50 file sono esclusi, in modo che le modifiche bulk generate, vendute o meccaniche non distorcano i risultati.

### Flag

| Flag                     | Default                   | Description                                     |
| ------------------------ | ------------------------- | ----------------------------------------------- |
| `-o`, `--output`         | `./skeleton/hotspots.skt` | Percorso del file di output                     |
| `--format`               | `skt`                     | Formato di output: `skt` o `json`               |
| `--since`                | `12mo`                    | Finestra temporale, come `180d`, `12mo` o `2y`  |
| `--max-commits`          | `2000`                    | Massimo numero di commit da ispezionare         |
| `--max-files-per-commit` | `50`                      | Escludi i commit che modificano più file        |
| `--min-cochanges`        | `3`                       | Minimo numero di commit condivisi per l'accoppiamento temporale |
| `--top-n`                | `30`                      | Massimo numero di risultati per sezione del report |
| `--include-merges`       | `false`                   | Includi i commit di merge                       |
| `--collect-test`         | `false`                   | Includi i file di test                          |
| `--workers`              | `NumCPU`                  | Massimo numero di goroutine di parsing concorrenti |
| `--verbose`              | `false`                   | Stampa informazioni di avanzamento              |

### Esempi

```bash
# Analyze the last 12 months
codeknit graph hotspots ./myproject

# Analyze two years and emit JSON
codeknit graph hotspots ./myproject --since 2y --format json -o hotspots.json

# Include larger commits and require stronger coupling
codeknit graph hotspots . --max-files-per-commit 100 --min-cochanges 5
```