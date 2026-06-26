---
title: Parse-Befehl
description: Extrahieren von strukturellen Informationen aus Quellcode in .skt-Dateien oder JSON.
---

Der Befehl `codeknit parse` extrahiert strukturelle Informationen aus Ihrem Codebase — wie Funktionen, Klassen, Methoden, Variablen und deren Beziehungen — und gibt sie standardmäßig im kompakten `.skt`-Format aus. Verwenden Sie JSON, wenn maschinenlesbare Ausgaben für Skripte, Integrationen oder nachgelagerte Tools benötigt werden.

## Grundlegende Verwendung

```bash
codeknit parse <input-path> [output-dir]
```

- **`<input-path>`**: Pfad zum Verzeichnis oder zur Datei, die analysiert werden soll.
- **`[output-dir]`**: Optionales Ausgabeverzeichnis. Falls nicht angegeben, wird standardmäßig `./skeleton` verwendet.

### Beispiele

```bash
# Projekt analysieren, Ausgabe im Standardverzeichnis ./skeleton
codeknit parse ./src

# Analysieren und in ein benutzerdefiniertes Ausgabeverzeichnis schreiben
codeknit parse ./src ./output

# Einzelne Datei analysieren und Ausgabe auf stdout
codeknit parse ./src/main.go --output-mode inline

# Maschinenlesbares JSON auf stdout ausgeben
codeknit parse ./src --output-mode inline --format json
```

## Ausgabemodi

Verwenden Sie `--output-mode`, um die Struktur der Ausgabe zu steuern. Drei Modi stehen zur Verfügung:

| Modus             | Beschreibung                                                                              | Am besten geeignet für                                  |
| ---------------- | ---------------------------------------------------------------------------------------- | --------------------------------------------------- |
| `directory-flat` | Schreibt aufgeteilte `.skt`-Dateien (z. B. `map_001.skt`, `map_002.skt`) in das Ausgabeverzeichnis. | ✅ **Die meisten Projekte** — Standard- und empfohlener Modus |
| `directory-tree` | Spiegelt die Verzeichnisstruktur der Quelle wider und erstellt eine `.skt`-Datei pro Quelldatei.        | Navigation der Ausgabe neben dem Quellcode             |
| `inline`         | Gibt die gesamte Ausgabe auf stdout aus.                                                              | Einzelne Dateien oder Weiterleitung an andere Tools               |

> **Tipp**: Verwenden Sie standardmäßig `directory-flat`, es sei denn, Sie arbeiten mit einer einzelnen Datei. Vermeiden Sie `inline` bei großen Eingaben, da dies Kontextfenster überlasten kann.

## Flags

| Flag             | Standardwert          | Beschreibung                                                                  |
| ---------------- | ---------------- | ---------------------------------------------------------------------------- |
| `--output-mode`  | `directory-flat` | Ausgabemodus: `inline`, `directory-flat` oder `directory-tree`                 |
| `--format`       | `skt`            | Ausgabeformat: `skt` oder `json`                                               |
| `--max-lines`    | `500`            | Maximale Zeilen pro Ausgabedatei in den Modi `flat`/`tree`                             |
| `--collect-test` | `false`          | Testdateien in die Analyse einbeziehen                                               |
| `--minify`       | `false`          | Aktiviert die wörterbuchbasierte Komprimierung, um den Token-Verbrauch zu reduzieren                    |
| `--edges`        | `false`          | Fügt den Abschnitt `[edges]` mit Beziehungsdaten hinzu (Aufrufe, Enthält-Beziehungen, etc.) |
| `--clean`        | `false`          | Entfernt vorhandene `.skt`-Dateien im Ausgabeverzeichnis vor dem Schreiben          |
| `--workers`      | `NumCPU`         | Maximale Anzahl paralleler Parsing-Goroutinen (0 = alle CPU-Kerne verwenden)      |
| `--verbose`      | `false`          | Gibt Fortschritts- und Zeitinformationen während der Verarbeitung aus                      |

## Häufige Muster

```bash
# Erste Analyse eines Projekts
codeknit parse ./src
```

```bash
# Erneute Analyse und Bereinigung vorheriger Ausgabe
codeknit parse ./src --clean
```

```bash
# Einzelne Datei auf stdout analysieren
codeknit parse ./src/main.go --output-mode inline
```

```bash
# Ausgabe für große Codebasen komprimieren
codeknit parse ./src --minify
```

```bash
# Beziehungs-Kanten einbeziehen (z. B. für Abhängigkeitsanalysen)
codeknit parse ./src --edges
```

```bash
# JSON für ein anderes Tool ausgeben
codeknit parse ./src --output-mode inline --format json --edges
```

Beispiel-JSON-Ausgabe:

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
# Verzeichnisstruktur der Quelle in der Ausgabe spiegeln
codeknit parse ./src --output-mode directory-tree
```

## Schutz vor veralteter Ausgabe

Falls das Ausgabeverzeichnis bereits `.skt`-Dateien aus einem vorherigen Durchlauf enthält, weigert sich `codeknit`, neue Ausgaben zu schreiben, um das Vermischen von veralteten und aktuellen Daten zu verhindern.

Um dieses Verhalten zu überschreiben und das Ausgabeverzeichnis vor dem Schreiben zu bereinigen, verwenden Sie das Flag `--clean`:

```bash
codeknit parse ./src --clean
```

Dies stellt eine frische, konsistente Ausgabe sicher.

## Tipps

- ✅ **Verwenden Sie standardmäßig `directory-flat`** für die meisten Projekte. Es bietet eine gute Balance zwischen Lesbarkeit und Handhabbarkeit.
- 🔍 Verwenden Sie `--minify` bei großen Codebasen, um den Token-Verbrauch durch ein gemeinsames Wörterbuch (`dict.skt`) zu reduzieren.
- 🔗 Der Abschnitt `[edges]` ist **standardmäßig ausgeschlossen**, um Tokens zu sparen. Verwenden Sie `--edges`, wenn Sie Beziehungsdaten wie `calls`, `contains` oder `inherits` benötigen.
- 🧾 Verwenden Sie `--format json`, wenn ein Skript oder eine Integration strukturierte Daten anstelle von `.skt` benötigt.
- 🧹 Verwenden Sie immer `--clean`, wenn Sie die Analyse im selben Ausgabeverzeichnis erneut durchführen.
- 📁 Verwenden Sie `directory-tree`, wenn Sie `.skt`-Dateien direkt mit den Quelldateien in Ihrem Editor verknüpfen möchten.