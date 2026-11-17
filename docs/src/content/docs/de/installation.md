---
title: Installation
description: So installieren Sie codeknit auf Ihrem System.
---

Diese Anleitung installiert codeknit aus dem Quellcode, fügt es zu Ihrem `PATH` hinzu, aktiviert Shell-Vervollständigungen und überprüft, ob es läuft.

:::tip[Nutzen Sie einen KI-Coding-Assistenten?]
Claude Code-, Codex- und Kiro-Nutzer können stattdessen das codeknit-Plugin installieren. Das Plugin lädt das Binary beim ersten Start herunter. Siehe [codeknit mit KI-Assistenten verwenden](../guides/ai-assistants/).
:::

## Aus dem Quellcode installieren

**Voraussetzungen**

- Go 1.26 oder höher
- Ein C-Compiler. codeknit verwendet CGo, um seine tree-sitter-Parser zu bauen.

**Schritte**

```bash
git clone https://github.com/awslabs/codeknit.git
cd codeknit
make build
```

Das Binary wird nach `./bin/codeknit` geschrieben.

### Stattdessen ein vorgefertigtes Binary verwenden

Vorgefertigte Binaries für Linux, macOS und Windows (amd64 und arm64) finden Sie auf der [Releases-Seite](https://github.com/awslabs/codeknit/releases). Laden Sie das Archiv für Ihre Plattform herunter, entpacken Sie es und verschieben Sie das Binary in Ihren `PATH`. Jedes Release enthält eine `checksums.txt`-Datei, mit der Sie den Download überprüfen können.

## codeknit zu Ihrem PATH hinzufügen

Überspringen Sie diesen Abschnitt, wenn Sie das Binary in ein Verzeichnis verschoben haben, das bereits in Ihrem `PATH` enthalten ist, z. B. `/usr/local/bin`.

Ersetzen Sie `/path/to/codeknit` durch das Verzeichnis, das das Binary enthält. Wenn Sie aus dem Quellcode gebaut haben, ist dies das `bin`-Verzeichnis innerhalb des geklonten Repositorys.

Fügen Sie die Zeile für Ihre Shell zu ihrer Konfigurationsdatei hinzu:

```bash
# bash (~/.bashrc)
export PATH="$PATH:/path/to/codeknit"

# zsh (~/.zshrc)
export PATH="$PATH:/path/to/codeknit"

# fish (~/.config/fish/config.fish)
fish_add_path /path/to/codeknit
```

Laden Sie Ihre Shell neu oder führen Sie die Datei aus, die Sie bearbeitet haben:

```bash
source ~/.bashrc   # bash
source ~/.zshrc    # zsh
```

Für fish öffnen Sie ein neues Terminal.

## Shell-Vervollständigungen aktivieren

codeknit generiert Vervollständigungsskripte für bash, zsh, fish und PowerShell. Führen Sie den Befehl für Ihre Shell einmal aus:

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

Die Befehle für bash, zsh und PowerShell hängen an Ihr Profil an. Wenn Sie sie erneut ausführen, wird das Skript ein zweites Mal hinzugefügt, daher führen Sie jeden Befehl nur einmal aus.

Öffnen Sie eine neue Shell oder laden Sie Ihr Profil, um die Vervollständigungen zu aktivieren.

## Installation überprüfen

```bash
codeknit --version
```

Wenn der Befehl eine Version ausgibt, ist codeknit installiert. Wenn Ihre Shell `command not found` meldet, überprüfen Sie, ob das Verzeichnis in Ihrer `PATH`-Zeile das `codeknit`-Binary enthält, und laden Sie dann Ihre Shell neu.

## Entwicklungsumgebung einrichten

Führen Sie dies nur aus, wenn Sie zu codeknit beitragen möchten. Führen Sie diese Befehle vom Stammverzeichnis des geklonten Repositorys aus:

```bash
make deps    # Abhängigkeiten installieren
make setup   # Git-Hooks konfigurieren
make test    # Testsuite ausführen
```

`make lint` führt den Linter und Formatter aus.

## Nächste Schritte

- Führen Sie Ihren ersten Befehl in [Erste Schritte](../getting-started/) aus.
- Durchsuchen Sie alle Flags im [CLI-Flags-Referenz](../reference/cli-flags/).