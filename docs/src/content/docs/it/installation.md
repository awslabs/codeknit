---
title: Installazione
description: Come installare codeknit sul proprio sistema.
---

Questa guida installa codeknit dal sorgente, lo aggiunge al `PATH`, abilita i completamenti della shell e verifica che funzioni.

:::tip[Usi un assistente di codifica AI?]
Gli utenti di Claude Code, Codex e Kiro possono installare il plugin codeknit. Il plugin scarica il binario la prima volta che viene eseguito. Vedi [Usa codeknit con assistenti AI](../guides/ai-assistants/).
:::

## Installazione dal sorgente

**Prerequisiti**

- Go 1.26 o successivo
- Un compilatore C. codeknit utilizza CGo per costruire i suoi parser tree-sitter.

**Passaggi**

```bash
git clone https://github.com/awslabs/codeknit.git
cd codeknit
make build
```

Il binario viene scritto in `./bin/codeknit`.

### Usa un binario precompilato

I binari precompilati per Linux, macOS e Windows (amd64 e arm64) sono disponibili nella [pagina delle release](https://github.com/awslabs/codeknit/releases). Scarica l'archivio per la tua piattaforma, estrailo e sposta il binario nel tuo `PATH`. Ogni release include un file `checksums.txt` con cui puoi verificare il download.

## Aggiungi codeknit al PATH

Salta questa sezione se hai spostato il binario in una directory già presente nel `PATH`, come `/usr/local/bin`.

Sostituisci `/path/to/codeknit` con la directory che contiene il binario. Se hai compilato dal sorgente, questa è la directory `bin` all'interno del repository clonato.

Aggiungi la riga per la tua shell al suo file di configurazione:

```bash
# bash (~/.bashrc)
export PATH="$PATH:/path/to/codeknit"

# zsh (~/.zshrc)
export PATH="$PATH:/path/to/codeknit"

# fish (~/.config/fish/config.fish)
fish_add_path /path/to/codeknit
```

Ricarica la shell o esegui il source del file che hai modificato:

```bash
source ~/.bashrc   # bash
source ~/.zshrc    # zsh
```

Per fish, apri un nuovo terminale.

## Abilita i completamenti della shell

codeknit genera script di completamento per bash, zsh, fish e PowerShell. Esegui il comando per la tua shell una volta:

```bash
# bash
codeknit completion bash >> ~/.bashrc

# zsh
codeknit completion zsh >> ~/.zshrc

# fish
codeknit completion fish > ~/.config/fish/completions/codeknit.fish

# PowerShell
codeknit completion powershell >> $PROFILE
```

I comandi per bash, zsh e PowerShell aggiungono il contenuto al tuo profilo. Eseguendoli nuovamente, lo script viene aggiunto una seconda volta, quindi esegui ciascun comando solo una volta.

Apri una nuova shell o esegui il source del tuo profilo per attivare i completamenti.

## Verifica l'installazione

```bash
codeknit --version
```

Se il comando stampa una versione, codeknit è installato. Se la shell riporta `command not found`, verifica che la directory nella riga del `PATH` contenga il binario `codeknit`, quindi ricarica la shell.

## Configura un ambiente di sviluppo

Fai questo solo se hai intenzione di contribuire a codeknit. Esegui questi comandi dalla radice del repository clonato:

```bash
make deps    # installa le dipendenze
make setup   # configura i git hooks
make test    # esegui la test suite
```

`make lint` esegue il linter e il formatter.

## Passaggi successivi

- Esegui il tuo primo comando in [Guida introduttiva](../getting-started/).
- Esplora ogni flag nella [riferimento dei flag della CLI](../reference/cli-flags/).