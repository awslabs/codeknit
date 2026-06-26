---
title: Installazione
description: Come installare codeknit sul proprio sistema.
---

codeknit può essere installato dal sorgente. I seguenti passaggi ti guideranno nella configurazione di codeknit sul tuo sistema.

## Dal sorgente

Il metodo di installazione principale è la compilazione dal sorgente. Avrai bisogno di:

- Go 1.26+
- Un compilatore C (richiesto per tree-sitter tramite CGo)

Clona il repository e compila il binario:

```bash
git clone https://github.com/awslabs/codeknit.git
cd codeknit
make build
```

Il binario compilato sarà disponibile in `./bin/codeknit`.

## Aggiungi a PATH

Per eseguire `codeknit` da qualsiasi directory, aggiungi la posizione del binario al PATH del tuo sistema.

Per **bash** (`~/.bashrc`):

```bash
export PATH="$PATH:/percorso/verso/codeknit"
```

Per **zsh** (`~/.zshrc`):

```bash
export PATH="$PATH:/percorso/verso/codeknit"
```

Per **fish** (`~/.config/fish/config.fish`):

```fish
fish_add_path /percorso/verso/codeknit
```

Dopo aver aggiornato la configurazione della shell, ricaricala eseguendo `source ~/.bashrc` (o `~/.zshrc`) o riavvia il terminale.

## Completamento della shell

codeknit supporta il completamento automatico per le shell più diffuse. Installa i completamenti utilizzando questi comandi:

Per **bash**:

```bash
codeknit completion bash >> ~/.bashrc
```

Per **zsh**:

```bash
codeknit completion zsh >> ~/.zshrc
```

Per **fish**:

```bash
codeknit completion fish > ~/.config/fish/completions/codeknit.fish
```

Per **PowerShell**:

```powershell
codeknit completion powershell >> $PROFILE
```

## Verifica installazione

Dopo l'installazione, verifica che codeknit sia configurato correttamente:

```bash
codeknit --version
```

## Configurazione per lo sviluppo

Se stai contribuendo a codeknit, esegui questi comandi aggiuntivi:

Installa le dipendenze di sviluppo:

```bash
make deps
```

Configura i git hooks:

```bash
make setup
```

Esegui la suite di test:

```bash
make test
```