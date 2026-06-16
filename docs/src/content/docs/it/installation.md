---
title: Installazione
description: Come installare codeknit sul proprio sistema.
---

codeknit può essere installato dal sorgente. I seguenti passaggi guideranno nell'impostazione di codeknit sul proprio sistema.

## Dal sorgente

Il metodo di installazione principale è la compilazione dal sorgente. Sono necessari:

- Go 1.26+
- Un compilatore C (richiesto per tree-sitter tramite CGo)

Clonare il repository e compilare il binario:

```bash
git clone https://github.com/awslabs/codeknit.git
cd codeknit
make build
```

Il binario compilato sarà disponibile in `./bin/codeknit`.

## Aggiungere al PATH

Per eseguire `codeknit` da qualsiasi directory, aggiungere la posizione del binario al PATH del sistema.

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

Dopo aver aggiornato la configurazione della shell, ricaricarla eseguendo `source ~/.bashrc` (o `~/.zshrc`) o riavviare il terminale.

## Completamento della shell

codeknit supporta il completamento automatico per le shell più diffuse. Installare i completamenti utilizzando questi comandi:

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

## Verifica dell'installazione

Dopo l'installazione, verificare che codeknit sia configurato correttamente:

```bash
codeknit --version
```

## Configurazione per lo sviluppo

Se si contribuisce a codeknit, eseguire questi comandi aggiuntivi:

Installare le dipendenze di sviluppo:

```bash
make deps
```

Configurare i git hooks:

```bash
make setup
```

Eseguire la suite di test:

```bash
make test
```
