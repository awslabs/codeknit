---
title: Utilizzo con assistenti AI
description: Configura codeknit come skill per Kiro, Claude Code e altri assistenti AI per la programmazione.
---

codeknit include skill pronte all'uso che insegnano agli assistenti AI per la programmazione come utilizzarlo efficacemente. Queste skill permettono agli assistenti di estrarre la struttura del codice, rilevare duplicati e eseguire analisi strutturali senza prompt manuali.

## Panoramica delle skill

codeknit fornisce due skill:

- **`codeknit-parse`**: Insegna agli assistenti a estrarre la struttura del codice (funzioni, classi, metodi, variabili) e le relazioni (chiamate, ereditarietà, contenimento) in file `.skt`.
- **`codeknit-fingerprint`**: Insegna agli assistenti a rilevare codice duplicato e quasi duplicato utilizzando fuzzy hashing.

Ogni skill include documentazione che l'assistente legge su richiesta per comprendere l'utilizzo, i flag, le modalità di output e i flussi di lavoro.

## Installazione

Copia le directory delle skill nella cartella delle skill del tuo assistente.

Per **Kiro**:

```bash
cp -r skills/codeknit-parse ~/.kiro/skills/codeknit-parse
cp -r skills/codeknit-fingerprint ~/.kiro/skills/codeknit-fingerprint
```

Per **Claude Code**:

```bash
cp -r skills/codeknit-parse ~/.claude/skills/codeknit-parse
cp -r skills/codeknit-fingerprint ~/.claude/skills/codeknit-fingerprint
```

Dopo l'installazione, l'assistente saprà automaticamente come invocare i comandi di codeknit, selezionare i flag appropriati e interpretare l'output `.skt`.

## Cosa insegna ogni skill

### codeknit-parse

La skill `codeknit-parse` insegna agli assistenti a:

- Eseguire `codeknit parse` con i flag appropriati per diversi scenari
- Scegliere la giusta modalità di output:
  - `directory-flat` (predefinita) per la maggior parte dei progetti
  - `inline` per singoli file o input di piccole dimensioni
  - `directory-tree` per rispecchiare la struttura del sorgente
- Leggere e interpretare i file di output `.skt`, incluse le sezioni `[symbols]`, `[edges]` e le sezioni opzionali `[dict]`
- Utilizzare i dati strutturali per refactoring, mappatura delle dipendenze e revisione del codice
- Eseguire `codeknit graph analyze` per approfondimenti sulla qualità del codice (dipendenze cicliche, simboli hub, god classes, ecc.)

### codeknit-fingerprint

La skill `codeknit-fingerprint` insegna agli assistenti a:

- Utilizzare `codeknit fingerprint` per il rilevamento di duplicati, audit DRY e identificazione di refactoring
- Selezionare intervalli di similarità appropriati (`--min-similarity`, `--max-similarity`)
- Leggere la sezione `[duplicates]` per identificare codice quasi duplicato
- Comprendere che i fingerprint misurano la forma strutturale, non l'intento semantico
- Utilizzare `--rerank` con gli embedding di Ollama per ridurre i falsi positivi quando necessario

## Esempi di flusso di lavoro

### Analisi strutturale

1. Chiedi all'assistente di analizzare la struttura del tuo codebase
2. Esegue `codeknit parse ./src` e legge i file `.skt` risultanti
3. Risponde a domande strutturali: dipendenze, catene di chiamate, dead code
4. Per approfondimenti, esegue `codeknit graph analyze ./src` e interpreta il report

```skt
[symbols]
## src/service.go
S1 type/struct L5-L8 AuthService {}
S2 callable/method L10-L15 Authenticate(token: string) {receiver=*AuthService}

[edges]
S1 --contains--> S2
```

### Rilevamento di duplicati

1. Chiedi all'assistente di trovare codice duplicato
2. Esegue `codeknit fingerprint ./src`
3. Legge la sezione `[duplicates]` nell'output
4. Indaga sulle coppie segnalate e propone la consolidazione

```skt
[duplicates]
S1, S2: 87% similarità
S3, S4: 76% similarità
```

## Consigli

- **Leggi sempre i file `.skt`, non il sorgente grezzo, per domande strutturali** — contengono la struttura estratta in un formato compatto e affidabile
- Utilizza `codeknit graph analyze` per scoprire problemi di qualità del codice come dipendenze cicliche, simboli hub e catene di ereditarietà profonde
- Esegui `codeknit fingerprint` prima di grandi refactoring per identificare il codice copiato e incollato che dovrebbe essere consolidato
- Il formato `.skt` è progettato per essere efficiente in termini di token, rendendolo ideale per le finestre di contesto degli LLM
- Utilizza `--minify` per ridurre ulteriormente l'uso di token quando si elaborano codebase di grandi dimensioni