---
title: Parse-Befehl
description: Extrahiert strukturelle Informationen aus Quellcode in .skt-Dateien oder JSON.
---

Der `codeknit parse`-Befehl extrahiert strukturelle Informationen aus Ihrem Codebase — wie Funktionen, Klassen, Methoden, Variablen und deren Beziehungen — und gibt sie standardmäßig im kompakten `.skt`-Format aus. Verwenden Sie JSON, wenn Sie maschinenlesbare Ausgaben für Skripte, Integrationen oder nachgelagerte Tools benötigen.

## Grundlegende Verwendung

```bash
codeknit parse <input-path> [output-dir]
```

- **`<input-path>`**: Pfad zum Verzeichnis oder zur Datei, die Sie parsen möchten.
- **`[output-dir]`**: Optionales Ausgabeverzeichnis. Wenn nicht angegeben, wird standardmäßig `./skeleton` verwendet.

### Beispiele

```bash
# Parse a project, output to default directory ./skeleton
codeknit parse ./src

# Parse and write to a custom output directory
codeknit parse ./src ./output

# Parse a single file and output to stdout
codeknit parse ./src/main.go --output-mode inline

# Emit machine-readable JSON to stdout
codeknit parse ./src --output-mode inline --format json
```

## Ausgabemodi

Verwenden Sie `--output-mode`, um zu steuern, wie die Ausgabe strukturiert wird. Drei Modi sind verfügbar:

| Modus             | Beschreibung                                                                              | Am besten geeignet für                                  |
| ----------------- | ----------------------------------------------------------------------------------------- | ------------------------------------------------------- |
| `directory-flat`  | Schreibt aufgeteilte `.skt`-Dateien (z. B. `map_001.skt`, `map_002.skt`) in das Ausgabeverzeichnis. | ✅ **Die meisten Projekte** — Standard- und empfohlener Modus |
| `directory-tree`  | Spiegelt die Quellverzeichnisstruktur wider und erstellt eine `.skt`-Datei pro Quelldatei. | Navigation der Ausgabe neben dem Quellcode             |
| `inline`          | Gibt die gesamte Ausgabe nach stdout aus.                                                 | Einzelne Dateien oder Weiterleitung an andere Tools    |

> **Tipp**: Verwenden Sie standardmäßig `directory-flat`, es sei denn, Sie arbeiten mit einer einzelnen Datei. Vermeiden Sie `inline` bei großen Eingaben, da dies Kontextfenster überlasten kann.

## Flags

| Flag              | Standardwert      | Beschreibung                                                                  |
| ----------------- | ----------------- | ----------------------------------------------------------------------------- |
| `--output-mode`   | `directory-flat`  | Ausgabemodus: `inline`, `directory-flat` oder `directory-tree`                |
| `--format`        | `skt`             | Ausgabeformat: `skt` oder `json`                                              |
| `--max-lines`     | `500`             | Maximale Zeilen pro Ausgabedatei in den Modi flat/tree                        |
| `--collect-test`  | `false`           | Testdateien in die Analyse einbeziehen                                        |
| `--minify`        | `false`           | Wörterbuchbasierte Komprimierung aktivieren, um die Token-Nutzung zu reduzieren |
| `--edges`         | `false`           | Den Abschnitt `[edges]` mit Beziehungsdaten einbeziehen (Aufrufe, enthält, etc.) |
| `--clean`         | `false`           | Vorhandene `.skt`-Dateien im Ausgabeverzeichnis vor dem Schreiben entfernen   |
| `--workers`       | `NumCPU`          | Maximale Anzahl paralleler Parsing-Goroutinen (0 = alle CPU-Kerne verwenden)  |
| `--verbose`       | `false`           | Fortschritts- und Zeitinformationen während der Verarbeitung ausgeben        |

## Häufige Muster

```bash
# First run on a project
codeknit parse ./src
```

```bash
# Re-run and clean previous output
codeknit parse ./src --clean
```

```bash
# Parse a single file to stdout
codeknit parse ./src/main.go --output-mode inline
```

```bash
# Minify output for large codebases
codeknit parse ./src --minify
```

```bash
# Include relationship edges (e.g., for dependency analysis)
codeknit parse ./src --edges
```

```bash
# Emit JSON for another tool
codeknit parse ./src --output-mode inline --format json --edges
```

Beispielausgabe im JSON-Format:

```json
{
  "files": ["app.go"],
  "symbols": [
    {
      "id": "app.go::User",
      "short_id": "S1",
      "name": "User",
      "file": "app.go",
      "category": "type",
      "kind": "struct",
      "signature": "type User struct",
      "span": [3, 3]
    },
    {
      "id": "app.go::Save",
      "short_id": "S2",
      "name": "Save",
      "file": "app.go",
      "category": "callable",
      "kind": "function",
      "signature": "Save(u: S1)",
      "span": [5, 5]
    }
  ],
  "edges": [
    {
      "from": "app.go::Save",
      "from_short": "S2",
      "to": "app.go::User",
      "to_short": "S1",
      "kind": "references"
    }
  ]
}
```

```bash
# Mirror source tree structure in output
codeknit parse ./src --output-mode directory-tree
```

## Unsicherheit bei Beziehungen

Mit `--edges` bewahrt Codeknit Beziehungen, selbst wenn Syntax und Scope kein eindeutiges Ziel festlegen:

```skt
S1 --calls--> S2
S1 --references[unresolved]--> string, bool
S1 --calls[ambiguous]--> Helper [candidates=S3, S4]
```

Ein fehlender Status bedeutet, dass der Analysator die Beziehung aufgelöst hat. `unresolved` bedeutet, dass die Abhängigkeit nicht hergestellt werden konnte; `ambiguous` bedeutet, dass konkurrierende Interpretationen möglich bleiben. Kandidaten sind mögliche Ziele, keine bestätigten Verbindungen. Bekannte Endpunkte verwenden ShortIDs; ungelöste Ziele ohne Symbole können als Namen erscheinen.

JSON enthält dieselben Informationen in `resolution` und `candidates` und lässt `to_short` für unsichere Ziele weg. C/C++-Includes enthalten `meta/file`- und `meta/include`-Symbole, sodass Kanten ShortIDs verwenden; dies zeichnet geschriebene Direktiven auf, ohne eine Auflösung auf Compiler-Ebene zu behaupten.

Abhängigkeitsmetriken schließen unsichere Beziehungen aus. Der HTML-Graph zeigt deren Anzahl an, ohne einzelne unsichere Links zu zeichnen. Weitere Informationen finden Sie im [Referenzhandbuch zum Ausgabeformat](../../reference/output-format/) und in der [Referenz zur Sprachunterstützung](../../reference/language-support/) für Auflösungsgrenzen.

## Schutz vor veralteter Ausgabe

Wenn das Ausgabeverzeichnis bereits `.skt`-Dateien aus einem vorherigen Durchlauf enthält, weigert sich `codeknit`, neue Ausgaben zu schreiben, um das Mischen von veralteten und aktuellen Daten zu verhindern.

Um dieses Verhalten zu überschreiben und das Ausgabeverzeichnis vor dem Schreiben zu bereinigen, verwenden Sie das `--clean`-Flag:

```bash
codeknit parse ./src --clean
```

Dies stellt sicher, dass ein frischer, konsistenter Ausgabesatz erstellt wird.

Beim Upgrade auf `0.5.0` sollte der gesamte Ausgabesatz neu generiert werden. Symbol-IDs und Beziehungsergebnisse können sich ändern; mischen Sie keine Chunks aus verschiedenen Durchläufen.

## Tipps

- ✅ **Verwenden Sie standardmäßig `directory-flat`** für die meisten Projekte. Es bietet eine gute Balance zwischen Lesbarkeit und Handhabbarkeit.
- 🔍 Verwenden Sie `--minify` bei großen Codebasen, um die Token-Nutzung durch ein gemeinsames Wörterbuch (`dict.skt`) zu reduzieren.
- 🔗 Der Abschnitt `[edges]` ist **standardmäßig ausgeschlossen**, um Tokens zu sparen. Verwenden Sie `--edges`, wenn Sie Beziehungsdaten wie `calls`, `contains` oder `inherits` benötigen.
- 🧾 Verwenden Sie `--format json`, wenn ein Skript oder eine Integration strukturierte Daten anstelle von `.skt` benötigt.
- 🧹 Verwenden Sie immer `--clean`, wenn Sie denselben Ausgabepfad erneut ausführen.
- 📁 Verwenden Sie `directory-tree`, wenn Sie `.skt`-Dateien direkt mit Quelldateien in Ihrem Editor korrelieren möchten.