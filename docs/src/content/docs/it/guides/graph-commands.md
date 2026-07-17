---
title: Comandi Grafici
description: Visualizza e analizza la struttura del codice con algoritmi di grafo.
---

codeknit offre due potenti comandi grafici per aiutarti a comprendere e migliorare la struttura del codice: `graph show` per la visualizzazione interattiva e `graph analyze` per l'analisi strutturale automatizzata.

## graph show

Genera una visualizzazione interattiva del grafo in HTML della codebase.

```bash
codeknit graph show <input-path>
```

Questo comando analizza la codebase e produce un file HTML autonomo con una visualizzazione interattiva del grafo. I simboli (funzioni, classi, tipi) appaiono come nodi, e le loro relazioni (chiamate, contiene, implementa) come archi. La visualizzazione si apre automaticamente nel browser predefinito.

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

Esegue algoritmi di grafo strutturali sulla codebase ed emette un report `.skt` leggibile da LLM contenente insight sulla qualità del codice.

```bash
codeknit graph analyze <input-path>
```

Questo comando rileva problemi comuni di qualità del codice come dipendenze cicliche, simboli hub, codice morto, god class e colli di bottiglia architetturali.

### Algoritmi

L'analisi include 22 algoritmi di grafo strutturali:

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
- Raggiungibilità dagli entry point
- Componenti debolmente connessi
- Peso delle dipendenze (forza di accoppiamento dei package)
- Distanza dalla Main Sequence (bilanciamento A+I)
- Rilevamento di shotgun surgery
- Rilevamento di feature envy
- Violazioni del principio delle dipendenze stabili
- Violazioni del principio di segregazione delle interfacce
- Profondità di contenimento

### Flag

| Flag                      | Default                         | Descrizione                                              |
| ------------------------- | ------------------------------- | -------------------------------------------------------- |
| `-o`, `--output`          | `./skeleton/graph_analysis.skt` | Percorso del file `.skt` di output                       |
| `--collect-test`          | `false`                         | Includi i file di test nell'analisi                      |
| `--workers`               | `NumCPU`                        | Numero massimo di goroutine di parsing concorrenti       |
| `--verbose`               | `false`                         | Mostra informazioni di avanzamento durante l'elaborazione |
| `--fan-threshold`         | `10`                            | Fan-in o fan-out minimo per segnalare un simbolo hub     |
| `--god-threshold`         | `15`                            | Conteggio minimo di archi contains per segnalare una god class/function |
| `--max-inheritance-depth` | `5`                             | Segnala catene di ereditarietà più profonde di questo valore |
| `--top-n`                 | `30`                            | Limita le sezioni di output classificate; 0 = nessun limite |
| `--betweenness-threshold` | `0.001`                         | Valore minimo di betweenness centrality da riportare     |
| `--propagation-cutoff`    | `0.05`                          | Probabilità minima per continuare la propagazione delle modifiche |

### Esempi

```skt
# Esegui l'analisi strutturale con i valori predefiniti
codeknit graph analyze ./myproject

# Output e soglie personalizzati
codeknit graph analyze ./myproject -o analysis.skt --fan-threshold 15

# Mostra più risultati per sezione
codeknit graph analyze ./myproject --top-n 50

# Includi i file di test
codeknit graph analyze ./src --collect-test
```
