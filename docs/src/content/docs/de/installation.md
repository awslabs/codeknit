---
title: Installation
description: So installieren Sie codeknit auf Ihrem System.
---

codeknit kann aus dem Quellcode installiert werden. Die folgenden Schritte führen Sie durch die Einrichtung von codeknit auf Ihrem System.

## Aus dem Quellcode

Die primäre Installationsmethode ist das Bauen aus dem Quellcode. Sie benötigen:

- Go 1.26+
- Einen C-Compiler (erforderlich für tree-sitter über CGo)

Klone das Repository und baue das Binary:

```bash
git clone https://github.com/awslabs/codeknit.git
cd codeknit
make build
```

Das kompilierte Binary ist unter `./bin/codeknit` verfügbar.

## Zum PATH hinzufügen

Um `codeknit` von jedem Verzeichnis aus ausführen zu können, fügen Sie den Speicherort des Binarys zum PATH Ihres Systems hinzu.

Für **bash** (`~/.bashrc`):

```bash
export PATH="$PATH:/pfad/zu/codeknit"
```

Für **zsh** (`~/.zshrc`):

```bash
export PATH="$PATH:/pfad/zu/codeknit"
```

Für **fish** (`~/.config/fish/config.fish`):

```fish
fish_add_path /pfad/zu/codeknit
```

Nach dem Aktualisieren Ihrer Shell-Konfiguration laden Sie diese neu, indem Sie `source ~/.bashrc` (oder `~/.zshrc`) ausführen oder Ihr Terminal neu starten.

## Shell-Completion

codeknit unterstützt die Autovervollständigung für gängige Shells. Installieren Sie die Completion mit diesen Befehlen:

Für **bash**:

```bash
codeknit completion bash >> ~/.bashrc
```

Für **zsh**:

```bash
codeknit completion zsh >> ~/.zshrc
```

Für **fish**:

```bash
codeknit completion fish > ~/.config/fish/completions/codeknit.fish
```

Für **PowerShell**:

```powershell
codeknit completion powershell >> $PROFILE
```

## Installation überprüfen

Überprüfen Sie nach der Installation, ob codeknit korrekt eingerichtet ist:

```bash
codeknit --version
```

## Entwicklungsumgebung einrichten

Wenn Sie zu codeknit beitragen möchten, führen Sie diese zusätzlichen Befehle aus:

Entwicklungsabhängigkeiten installieren:

```bash
make deps
```

Git-Hooks einrichten:

```bash
make setup
```

Testsuite ausführen:

```bash
make test
```