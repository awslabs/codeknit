---
title: Comandi Grafici
description: Visualizza e analizza la struttura della tua codebase con algoritmi di grafi.
---

codeknit fornisce comandi grafici per visualizzare la struttura, eseguire analisi automatizzate e combinare il grafo delle dipendenze corrente con la cronologia delle modifiche Git.

## graph show

Genera una visualizzazione interattiva del grafo in HTML della tua codebase.

```bash
codeknit graph show <input-path>
```

Questo comando analizza la tua codebase e produce un file HTML autonomo con una visualizzazione interattiva del grafo. I simboli (funzioni, classi, tipi) appaiono come nodi, e le loro relazioni (chiamate, contiene, implementa) come archi. La visualizzazione si apre automaticamente nel browser predefinito.

### Flag

| Flag             | Default                          | Descrizione                                  |
| ---------------- | -------------------------------- | -------------------------------------------- |
| `-o`, `--output` | `./skeleton/codeknit-graph.html` | Percorso del file HTML di output             |
| `--collect-test` | `false`                          | Includi i file di test nell'analisi          |
| `--workers`      | `NumCPU`                         | Numero massimo di goroutine di parsing concorrenti |
| `--verbose`      | `false`                          | Mostra informazioni di avanzamento durante l'elaborazione |

### Esempi

```skt
# Genera la visualizzazione predefinita
codeknit graph show ./myproject

# File di output personalizzato
codeknit graph show ./myproject -o graph.html

# Includi i file di test
codeknit graph show ./src --collect-test
```

## graph analyze

Esegue algoritmi strutturali su grafi sulla tua codebase ed emette un report `.skt` leggibile da LLM contenente insight sulla qualità del codice.

```bash
codeknit graph analyze <input-path>
```

Questo comando rileva problemi comuni di qualità del codice come dipendenze cicliche, simboli hub, codice morto, god class e colli di bottiglia architetturali.

### Algoritmi

L'analisi include 22 algoritmi strutturali su grafi:

- Dipendenze cicliche (Tarjan's SCC)
- Rilevamento di hub (alto accoppiamento fan-in/fan-out)
- Rilevamento di orfani (candidati di codice morto)
- Rilevamento di god class/function (figli eccessivi)
- Metrica di instabilità (Robert C. Martin's Ce/(Ca+Ce))
- Catene di ereditarietà profonde
- Centralità di betweenness (rilevamento di colli di bottiglia)
- Punti di articolazione (singoli punti di fallimento)
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

### Flag

| Flag                      | Default                         | Descrizione                                              |
| ------------------------- | ------------------------------- | -------------------------------------------------------- |
| `-o`, `--output`          | `./skeleton/graph_analysis.skt` | Percorso del file `.skt` di output                       |
| `--collect-test`          | `false`                          | Includi i file di test nell'analisi                      |
| `--workers`               | `NumCPU`                        | Numero massimo di goroutine di parsing concorrenti       |
| `--verbose`               | `false`                          | Mostra informazioni di avanzamento durante l'elaborazione |
| `--fan-threshold`         | `10`                            | Fan-in o fan-out minimo per segnalare un simbolo hub     |
| `--god-threshold`         | `15`                            | Conteggio minimo di archi "contiene" per segnalare una god class/function |
| `--max-inheritance-depth` | `5`                             | Segnala catene di ereditarietà più profonde di questo valore |
| `--top-n`                 | `30`                            | Limita le sezioni di output classificate; 0 = nessun limite |
| `--betweenness-threshold` | `0.001`                         | Valore minimo di centralità di betweenness da riportare  |
| `--propagation-cutoff`    | `0.05`                          | Probabilità minima per continuare la propagazione delle modifiche |

### Esempi

```skt
# Esegui l'analisi strutturale con valori predefiniti
codeknit graph analyze ./myproject

# Output e soglie personalizzati
codeknit graph analyze ./myproject -o analysis.skt --fan-threshold 15

# Mostra più risultati per sezione
codeknit graph analyze ./myproject --top-n 50

# Includi i file di test
codeknit graph analyze ./src --collect-test
```

## graph hotspots

Classifica i file che sono sia frequentemente modificati che strutturalmente importanti:

```bash
codeknit graph hotspots <input-path>
```

Il punteggio combina frequenza dei commit, churn delle righe e recenza con PageRank a livello di file, fan-in transitivo e centralità di betweenness. Il report identifica anche l'accoppiamento temporale tra file che vengono modificati ripetutamente negli stessi commit.

I commit di merge sono esclusi per impostazione predefinita. Sono esclusi anche i commit che modificano più di 50 file, in modo che modifiche bulk generate, vendored o meccaniche non distorcano i risultati.

### Flag

| Flag                     | Default                   | Descrizione                                      |
| ------------------------ | ------------------------- | ------------------------------------------------ |
| `-o`, `--output`         | `./skeleton/hotspots.skt` | Percorso del file di output                      |
| `--format`               | `skt`                     | Formato di output: `skt` o `json`                |
| `--since`                | `12mo`                    | Finestra temporale, come `180d`, `12mo`, o `2y`  |
| `--max-commits`          | `2000`                    | Numero massimo di commit da ispezionare          |
| `--max-files-per-commit` | `50`                      | Escludi i commit che modificano più file         |
| `--min-cochanges`        | `3`                       | Numero minimo di commit condivisi per l'accoppiamento temporale |
| `--top-n`                | `30`                      | Numero massimo di risultati per sezione del report |
| `--include-merges`       | `false`                   | Includi i commit di merge                        |
| `--collect-test`         | `false`                   | Includi i file di test                           |
| `--workers`              | `NumCPU`                  | Numero massimo di goroutine di parsing concorrenti |
| `--verbose`              | `false`                   | Mostra informazioni di avanzamento               |

### Esempi

```bash
# Analizza gli ultimi 12 mesi
codeknit graph hotspots /codeknit/it/myproject

# Analizza due anni ed emetti JSON
codeknit graph hotspots /codeknit/it/myproject --since 2y --format json -o hotspots.json

# Includi commit più grandi e richiedi un accoppiamento più forte
codeknit graph hotspots . --max-files-per-commit 100 --min-cochanges 5
```