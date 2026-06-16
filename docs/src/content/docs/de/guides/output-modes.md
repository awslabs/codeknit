---
title: Ausgabemodi
description: Wählen Sie den richtigen Ausgabemodus für die Größe Ihres Projekts und Ihren Workflow.
---

codeknit unterstützt drei Ausgabemodi, die durch das Flag `--output-mode` gesteuert werden. Jeder Modus bestimmt, wie die extrahierte Codestruktur auf die Festplatte (oder stdout) geschrieben wird.

### directory-flat (Standard, empfohlen)

- **Verhalten**: Schreibt chunkweise `.skt`-Dateien wie `map_001.skt`, `map_002.skt` usw.
- **Ausgabeverzeichnis**: `./skeleton/` standardmäßig
- **Aufteilung**: Dateien werden aufgeteilt, wenn sie das Limit `--max-lines` (Standard: 500 Zeilen) überschreiten
- **Anwendungsfall**: Am besten für die meisten Projekte geeignet. Hält die Ausgabe organisiert und lesbar, indem die Dateigröße begrenzt wird. Sie können nur die für Ihre Aufgabe relevanten Chunks lesen.
- **Minimierung**: Wenn `--minify` aktiviert ist, wird auch eine `dict.skt`-Datei im Ausgabeverzeichnis generiert, die Token-Zuordnungen für komprimierte Werte enthält.

Beispiel:

```bash
codeknit parse ./src
# Ausgabe: ./skeleton/map_001.skt, map_002.skt, ...
```

### directory-tree

- **Verhalten**: Spiegelt die Quellverzeichnisstruktur exakt wider.
- **Ausgabeverzeichnis**: `./skeleton/` standardmäßig
- **Zuordnung**: Pro Quelldatei wird eine `.skt`-Datei am entsprechenden Pfad erstellt.
- **Anwendungsfall**: Ideal, wenn Sie die Struktur einer bestimmten Datei schnell nachschlagen möchten. Nützlich für die Navigation neben dem ursprünglichen Code-Repository.

Beispiel:

```bash
codeknit parse ./src --output-mode directory-tree
# Ausgabe: ./skeleton/src/handler.skt, ./skeleton/pkg/db.skt, usw.
```

### inline

- **Verhalten**: Gibt die gesamte Ausgabe an stdout aus.
- **Ausgabeverzeichnis**: Keines erstellt
- **Anwendungsfall**: Nur für einzelne Dateien oder sehr kleine Projekte (weniger als 5 Dateien) empfohlen. Nützlich, wenn die Ausgabe an ein anderes Tool weitergeleitet oder eine einzelne Datei interaktiv inspiziert wird.

Beispiel:

```bash
codeknit parse ./src/main.go --output-mode inline
# Ausgabe: direkt im Terminal ausgegeben
```

### Entscheidungstabelle

| Modus            | Am besten geeignet für                     | Ausgabepfad                                                       |
| ---------------- | ------------------------------------------ | ----------------------------------------------------------------- |
| `directory-flat` | Die meisten Projekte (Standard, empfohlen) | `./skeleton/map_001.skt`, `map_002.skt`, ...                      |
| `directory-tree` | Navigation der Ausgabe neben Quellcode     | `./skeleton/<gespiegelter Pfad>.skt`                              |
| `inline`         | Einzelne Datei, Weiterleitung an ein Tool  | stdout — nur für einzelne Dateien oder winzige Projekte verwenden |

### Faustregeln

- **Im Zweifel** → `directory-flat` (Standard) verwenden
- **Inspektion einer einzelnen Datei** → `inline` ist akzeptabel
- **Mehr als ein paar Dateien** → `directory-flat` oder `directory-tree` bevorzugen
- **Große Codebasen** → `--minify` hinzufügen, um den Token-Verbrauch zu reduzieren
- **Erneutes Ausführen auf derselben Ausgabe** → `--clean` verwenden, um veraltete `.skt`-Dateien zu entfernen

### Minimierung

Das Flag `--minify` aktiviert die wörterbuchbasierte Komprimierung wiederholter Tokens (z. B. Eigenschaftsschlüssel wie `exported`, `async` oder gängige Typnamen). Wenn aktiviert:

- Wiederholte Werte werden durch kurze Codes (`d0`, `d1`, `d2`, ...) ersetzt
- Eine `dict.skt`-Datei wird im Ausgabeverzeichnis geschrieben, die Codes den ursprünglichen Werten zuordnet
- Reduziert die Ausgabegöße für große Codebasen erheblich
- Funktioniert sowohl im `directory-flat`- als auch im `directory-tree`-Modus

Beispiel für minimierte Ausgabe:

```skt
[dict]
- d0: exported
- d1: callable/function

[symbols]
## src/main.py
S1 d1 L1-L5 main() {d0}
```

Dieses Format bewahrt alle Informationen, während der Token-Fußabdruck minimiert wird, was es ideal für LLM-basierte Analysen macht.
