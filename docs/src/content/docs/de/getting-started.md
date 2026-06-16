---
title: Erste Schritte
description: codeknit in weniger als 5 Minuten einrichten und loslegen.
---

# Erste Schritte

codeknit in weniger als 5 Minuten einrichten und loslegen.

## 1. Voraussetzungen

Sie benötigen:

- Go 1.26+
- Einen C-Compiler (CGo wird für tree-sitter benötigt)

## 2. Installation aus dem Quellcode

```bash
git clone https://github.com/awslabs/codeknit.git
cd codeknit
make build
# Binary befindet sich unter ./bin/codeknit
```

## 3. Zum PATH hinzufügen

Fügen Sie das Binary zu Ihrem Shell-PATH hinzu:

```bash
# bash (~/.bashrc)
export PATH="$PATH:$(pwd)/bin"

# zsh (~/.zshrc)
export PATH="$PATH:$(pwd)/bin"

# fish (~/.config/fish/config.fish)
fish_add_path $(pwd)/bin
```

Laden Sie Ihre Shell neu oder führen Sie `source ~/.bashrc` (oder `~/.zshrc`) aus, damit die Änderung wirksam wird.

## 4. Installation überprüfen

Prüfen Sie, ob codeknit funktioniert:

```bash
codeknit --version
```

## 5. Erster Parse-Vorgang

Führen Sie Ihren ersten Parse-Vorgang für ein Code-Repository durch:

```bash
codeknit parse ./myproject
```

Dieser Befehl:

- Parst alle Quelldateien in `./myproject`
- Extrahiert strukturelle Informationen (Funktionen, Klassen, Beziehungen)
- Schreibt chunked `.skt`-Dateien in `./skeleton/` (Standard-Ausgabeverzeichnis)

Wenn Sie diesen Befehl erneut ausführen, verwenden Sie `--clean`, um vorherige Ausgaben zu entfernen:

```bash
codeknit parse ./myproject --clean
```

## 6. Die Ausgabe lesen

Die `.skt`-Dateien enthalten strukturierte Code-Informationen. Hier ein kleines Beispiel:

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

Wichtige Abschnitte:

- `[symbols]`: Definitionen, gruppiert nach Datei, mit Name, Zeilenbereich und Metadaten
- `[edges]`: Beziehungen wie `contains`, `calls`, `inherits` oder `returns`

## 7. Nächste Schritte

Nachdem Sie Ihren ersten Parse-Vorgang durchgeführt haben:

- Erfahren Sie mehr über den Parse-Befehl: [Parse-Befehlsleitfaden](/codeknit/de/guides/parse-command/)
- Entdecken Sie strukturelle Analysen: [Graph-Befehlsleitfaden](/codeknit/de/guides/graph-commands/)
- Verstehen Sie die Duplikat-Erkennung: [Fingerprint-Befehlsleitfaden](/codeknit/de/guides/fingerprint-command/)
- Lesen Sie das vollständige Ausgabeformat: [Ausgabeformat-Referenz](/codeknit/de/reference/output-format/)
- Sehen Sie sich alle verfügbaren Flags an: [CLI-Flags-Referenz](/codeknit/de/reference/cli-flags/)
