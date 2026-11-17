---
title: Utilizzo con assistenti AI
description: Installa il plugin codeknit per Claude Code, Codex e Kiro.
---

Installa il plugin codeknit in Claude Code, Codex o Kiro, mantienilo aggiornato e chiedi all'assistente domande sulla struttura del codice e sui duplicati. Il plugin contiene una skill, `codeknit`. La skill indica all'assistente quando eseguire `codeknit parse`, `codeknit graph analyze` e `codeknit fingerprint`, e come leggere l'output `.skt`.

Non è necessario installare il binario `codeknit` autonomamente. La prima volta che la skill viene eseguita, esegue il suo script incluso `scripts/install-codeknit.sh`, che svolge le seguenti azioni:

1. Esce immediatamente se `codeknit` è già presente nel tuo `PATH` o in `~/.local/bin`.
2. Altrimenti, scarica l'ultima release GitHub per il tuo sistema operativo e architettura.
3. Verifica il download rispetto al file `checksums.txt` della release.
4. Installa il binario in `~/.local/bin` e stampa il suo percorso assoluto nell'ultima riga dell'output.

Se `~/.local/bin` non è nel tuo `PATH`, l'assistente chiama codeknit utilizzando quel percorso completo per il resto della sessione.

Per controllare l'installazione, imposta queste variabili d'ambiente:

| Variabile | Effetto |
| --- | --- |
| `CODEKNIT_VERSION` | Blocca una release, ad esempio `0.6.0`. |
| `CODEKNIT_INSTALL_DIR` | Modifica la posizione di installazione. |

Per installare il binario autonomamente, consulta [Installazione](../../installation/).

## Installa il plugin

Scegli la sezione relativa al tuo assistente.

### Claude Code

Esegui questi comandi nel tuo shell. All'interno di una sessione di Claude Code, esegui gli stessi comandi come `/plugin ...`.

```bash
claude plugin marketplace add awslabs/codeknit --sparse .claude-plugin skills
claude plugin install codeknit@codeknit
```

### Codex

```bash
codex plugin marketplace add awslabs/codeknit --sparse .agents --sparse .codex-plugin --sparse skills
codex plugin add codeknit@codeknit
```

### Kiro IDE

1. Apri il pannello Powers.
2. Scegli **Add Custom Power**, poi **Import power from GitHub**.
3. Inserisci `https://github.com/awslabs/codeknit`.

### Kiro CLI

La Kiro CLI installa da un percorso locale, quindi clona prima il repository:

```bash
git clone --depth 1 --filter=blob:none --sparse https://github.com/awslabs/codeknit ~/.codeknit-power
git -C ~/.codeknit-power sparse-checkout set skills
kiro-cli powers install ~/.codeknit-power
```

### Informazioni su `--sparse`

L'opzione `--sparse` è facoltativa. Limita la clonazione ai manifest del plugin e a `skills/` invece dell'intero albero sorgente.

### Installa la skill manualmente

Se il tuo assistente legge le Agent Skills ma non supporta i plugin, copia `skills/codeknit/` nella directory delle skills per il tuo assistente:

- `~/.claude/skills/`
- `~/.codex/skills/`
- `~/.kiro/skills/`

## Aggiorna il plugin

| Assistente | Comando o azione |
| --- | --- |
| Claude Code | `claude plugin marketplace update codeknit` |
| Codex | `codex plugin marketplace upgrade` |
| Kiro IDE | Scegli **Check for updates** nel pannello Powers. |
| Kiro CLI | Esegui `git -C ~/.codeknit-power pull`, poi esegui nuovamente `kiro-cli powers install ~/.codeknit-power`. |

Aggiornare il plugin non aggiorna un binario già installato. Per cambiare la versione del binario, imposta `CODEKNIT_VERSION` come descritto sopra.

Se aggiorni codeknit e l'estrazione cambia, anche gli ID dei simboli possono cambiare. Chiedi all'assistente di rigenerare tutto l'output invece di mescolare chunk provenienti da esecuzioni diverse.

## Cosa insegna la skill all'assistente

La skill fornisce all'assistente queste regole, così non devi ripeterle in ogni richiesta:

- **Scegli il comando in base alla domanda.**
  - `parse` risponde a cosa esiste, dove si trova e cosa chiama, eredita o contiene cosa.
  - `graph analyze` trova cicli, hub, dead code, god class e colli di bottiglia.
  - `fingerprint` trova copie-incolla e candidati per il refactoring.
- **Usa come predefinita la modalità di output `directory-flat`.** L'assistente scrive i file chunkati in `./skeleton/` e legge solo i chunk di cui ha bisogno. Utilizza l'output `inline` solo per un singolo file o una manciata di file, perché `inline` scarica tutto nella finestra di contesto.
- **Leggi l'output `.skt`, non il sorgente grezzo, per domande strutturali.** L'assistente apre i file sorgente solo per ispezionare o modificare i dettagli di implementazione, utilizzando gli intervalli di righe dall'output. Il file `references/output-format.md` della skill contiene il formato completo. Vedi anche il [riferimento del formato di output](../../reference/output-format/).
- **Scegli intervalli di similarità e tratta le corrispondenze come segnali.** I fingerprint confrontano la forma strutturale, non l'intento. L'assistente legge entrambi i simboli di una coppia prima di proporre un refactoring.

## Richiedi un'analisi strutturale

1. Chiedi all'assistente di analizzare la struttura del tuo codebase. Ad esempio: "Mappa la struttura di `./src` e dimmi quali moduli dipendono dal pacchetto auth."
2. L'assistente verifica che il binario sia installato, poi esegue `codeknit parse` con gli archi abilitati e legge l'output `.skt`.
3. Fai domande di follow-up su dipendenze, catene di chiamate o dead code. L'assistente risponde dall'output.
4. Per un'analisi più approfondita, chiedi una revisione strutturale. Ad esempio: "Trova dipendenze cicliche e god class in `./src`." L'assistente esegue `codeknit graph analyze` e legge il report.

Le relazioni contrassegnate come `[unresolved]` o `[ambiguous]` nell'output sono incerte, non dipendenze confermate. Tratta le risposte che si basano su di esse come candidati da verificare.

Per i comandi dietro queste richieste, consulta [la guida al comando parse](../parse-command/) e [la guida ai comandi graph](../graph-commands/).

## Richiedi il rilevamento dei duplicati

1. Chiedi all'assistente di trovare codice duplicato. Ad esempio: "Trova funzioni copiate-incollate in `./src`."
2. L'assistente esegue `codeknit fingerprint` e legge la sezione `[duplicates]` di `./skeleton/fingerprints.skt`.
3. L'assistente apre entrambi i simboli di ogni coppia segnalata, poi propone un refactoring.
4. Se i risultati sono rumorosi, chiedigli di regolare l'intervallo di similarità o di aggiungere un reranking semantico. Il reranking richiede un server Ollama locale con un modello di embedding scaricato.

Per le opzioni dietro queste richieste, consulta [la guida al comando fingerprint](../fingerprint-command/).

## Suggerimenti

- Ricorda all'assistente di leggere i file `.skt`, non il sorgente grezzo, per domande strutturali. Questo mantiene il contesto piccolo, poiché il formato è compatto e contiene simboli, relazioni e intervalli di righe.
- Usa `graph analyze` per approfondimenti sulla qualità del codice. I risultati sono candidati per la revisione, perché chiamanti esterni e dispatch dinamico possono cambiare il risultato.
- Esegui fingerprint prima di un refactoring ampio per individuare copie-incolla che dovrebbero essere consolidate.
- Se l'assistente riesegue `parse` in una directory che contiene già output `.skt`, codeknit rifiuta di scrivere. Chiedigli di passare `--clean` per sostituire l'output precedente.
- Se il download del binario fallisce (offline o piattaforma non supportata), ottieni una release manualmente dalla [pagina delle release](https://github.com/awslabs/codeknit/releases), o consulta [Installazione](../../installation/).
- Se un'esecuzione riporta risultati parziali, controlla stderr e la sezione `[errors]`. Vedi [affidabilità dell'analisi](../../reference/analysis-reliability/) per i dettagli.