---
title: Iniziare
description: Inizia a usare codeknit in meno di 5 minuti.
---

# Iniziare

Inizia a usare codeknit in meno di 5 minuti.

## 1. Prerequisiti

Avrai bisogno di:

- Go 1.26+
- Un compilatore C (CGo è richiesto per tree-sitter)

## 2. Installazione da sorgente

```bash
git clone https://github.com/awslabs/codeknit.git
cd codeknit
make build
# Il binario si trova in ./bin/codeknit
```

## 3. Aggiungere al PATH

Aggiungi il binario al PATH della tua shell:

```bash
# bash (~/.bashrc)
export PATH="$PATH:$(pwd)/bin"

# zsh (~/.zshrc)
export PATH="$PATH:$(pwd)/bin"

# fish (~/.config/fish/config.fish)
fish_add_path $(pwd)/bin
```

Ricarica la tua shell o esegui `source ~/.bashrc` (o `~/.zshrc`) affinché la modifica abbia effetto.

## 4. Verifica dell'installazione

Controlla che codeknit funzioni:

```bash
codeknit --version
```

## 5. Primo parse

Esegui il tuo primo parse su una codebase:

```bash
codeknit parse ./myproject
```

Questo comando:

- Effettua il parsing di tutti i file sorgente in `./myproject`
- Estrae informazioni strutturali (funzioni, classi, relazioni)
- Scrive file `.skt` suddivisi in `./skeleton/` (directory di output predefinita)

Se esegui nuovamente questo comando, usa `--clean` per rimuovere l'output precedente:

```bash
codeknit parse ./myproject --clean
```

## 6. Lettura dell'output

I file `.skt` contengono informazioni strutturate sul codice. Ecco un piccolo esempio:

```skt
[symbols]
## src/main.go
S1 module/package L1-L1 main {}
S2 type/struct L5-L8 Server {exported}
S3 callable/function L10-L12 NewServer(addr: string) -> *S2 {exported}
S4 callable/method L14-L19 Start() {receiver=*Server}

[edges]
S2 --contains--> S4
S3 --returns--> S2
```

Sezioni chiave:

- `[symbols]`: Definizioni raggruppate per file, che mostrano nome, **intervallo di righe** e metadati
- `[edges]`: Relazioni come `contains`, `calls`, `inherits` o `returns`

## 7. Passaggi successivi

Ora che hai eseguito il tuo primo parse:

- Approfondisci il comando di parsing: [Guida al comando parse](/codeknit/it/guides/parse-command/)
- Esplora l'**analisi del grafo**: [Guida ai comandi del grafo](/codeknit/it/guides/graph-commands/)
- Comprendi il rilevamento dei **duplicati**: [Guida al comando fingerprint](/codeknit/it/guides/fingerprint-command/)
- Leggi il formato di output completo: [Riferimento al formato di output](/codeknit/it/reference/output-format/)
- Visualizza tutte le flag disponibili: [Riferimento alle flag CLI](/codeknit/it/reference/cli-flags/)
