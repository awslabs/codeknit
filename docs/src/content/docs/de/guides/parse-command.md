---
title: Parse-Befehl
description: Extrahiert strukturelle Informationen aus Quellcode in .skt-Dateien.
---

Der Befehl `codeknit parse` extrahiert strukturelle Informationen aus Ihrem Codebase — wie Funktionen, Klassen, Methoden, Variablen und deren Beziehungen — und gibt sie in einem kompakten `.skt`-Format aus, das für die effiziente Nutzung durch LLMs und Analyse-Tools entwickelt wurde.

## Grundlegende Verwendung

```bash
codeknit parse <input-path> [output-dir]
```

- **`<input-path>`**: Pfad zum Verzeichnis oder zur Datei, die Sie parsen möchten.
- **`[output-dir]`**: Optionales Ausgabeverzeichnis. Wenn nicht angegeben, wird standardmäßig `./skeleton` verwendet.

### Beispiele

```bash
# Projekt parsen, Ausgabe im Standardverzeichnis ./skeleton
codeknit parse ./src

# Parsen und in ein benutzerdefiniertes Ausgabeverzeichnis schreiben
codeknit parse ./src ./output

# Einzelne Datei parsen und nach stdout ausgeben
codeknit parse ./src/main.go --output-mode inline
```

## Ausgabemodi

Verwenden Sie `--output-mode`, um zu steuern, wie die Ausgabe strukturiert wird. Drei Modi sind verfügbar:

| Modus            | Beschreibung                                                                                        | Am besten geeignet für                                        |
| ---------------- | --------------------------------------------------------------------------------------------------- | ------------------------------------------------------------- |
| `directory-flat` | Schreibt aufgeteilte `.skt`-Dateien (z. B. `map_001.skt`, `map_002.skt`) in das Ausgabeverzeichnis. | ✅ **Die meisten Projekte** — Standard- und empfohlener Modus |
| `directory-tree` | Spiegelt die Quellverzeichnisstruktur wider und erstellt eine `.skt`-Datei pro Quelldatei.          | Navigation der Ausgabe neben dem Quellcode                    |
| `inline`         | Gibt alle Daten nach stdout aus.                                                                    | Einzelne Dateien oder Weiterleitung an andere Tools           |

> **Tipp**: Verwenden Sie standardmäßig `directory-flat`, es sei denn, Sie arbeiten mit einer einzelnen Datei. Vermeiden Sie `inline` bei großen Eingaben, da dies Kontextfenster überlasten kann.

## Flags

| Flag             | Standardwert     | Beschreibung                                                                    |
| ---------------- | ---------------- | ------------------------------------------------------------------------------- |
| `--output-mode`  | `directory-flat` | Ausgabemodus: `inline`, `directory-flat` oder `directory-tree`                  |
| `--max-lines`    | `500`            | Maximale Zeilen pro Ausgabedatei in Flat/Tree-Modi                              |
| `--collect-test` | `false`          | Testdateien in die Analyse einbeziehen                                          |
| `--minify`       | `false`          | Aktiviert dictionary-basierte Komprimierung, um die Token-Nutzung zu reduzieren |
| `--edges`        | `false`          | Fügt den Abschnitt `[edges]` mit Beziehungsdaten hinzu (Aufrufe, Enthält, usw.) |
| `--clean`        | `false`          | Entfernt vorhandene `.skt`-Dateien im Ausgabeverzeichnis vor dem Schreiben      |
| `--workers`      | `NumCPU`         | Maximale Anzahl gleichzeitiger Parsing-Goroutinen (0 = alle CPU-Kerne nutzen)   |
| `--verbose`      | `false`          | Gibt Fortschritts- und Zeitinformationen während der Verarbeitung aus           |

## Häufige Muster

```bash
# Erstmaliger Durchlauf für ein Projekt
codeknit parse ./src
```

```bash
# Erneuter Durchlauf und Bereinigung vorheriger Ausgabe
codeknit parse ./src --clean
```

```bash
# Einzelne Datei nach stdout parsen
codeknit parse ./src/main.go --output-mode inline
```

```bash
# Ausgabe für große Codebases minimieren
codeknit parse ./src --minify
```

```bash
# Beziehungs-Kanten einbeziehen (z. B. für Abhängigkeitsanalysen)
codeknit parse ./src --edges
```

```bash
# Quellbaumstruktur in der Ausgabe spiegeln
codeknit parse ./src --output-mode directory-tree
```

## Schutz vor veralteter Ausgabe

Wenn das Ausgabeverzeichnis bereits `.skt`-Dateien aus einem vorherigen Durchlauf enthält, weigert sich `codeknit`, neue Ausgaben zu schreiben, um das Mischen von veralteten und frischen Daten zu verhindern.

Um dieses Verhalten zu überschreiben und das Ausgabeverzeichnis vor dem Schreiben zu bereinigen, verwenden Sie das Flag `--clean`:

```bash
codeknit parse ./src --clean
```

Dies stellt sicher, dass ein frischer, konsistenter Ausgabesatz erstellt wird.

## Tipps

- ✅ **Verwenden Sie standardmäßig `directory-flat`** für die meisten Projekte. Es bietet eine gute Balance zwischen Lesbarkeit und Handhabbarkeit.
- 🔍 Verwenden Sie `--minify` bei großen Codebases, um die Token-Nutzung durch ein gemeinsames Wörterbuch (`dict.skt`) zu reduzieren.
- 🔗 Der Abschnitt `[edges]` ist **standardmäßig ausgeschlossen**, um Token zu sparen. Verwenden Sie `--edges`, wenn Sie Beziehungsdaten wie `calls`, `contains` oder `inherits` benötigen.
- 🧹 Verwenden Sie immer `--clean`, wenn Sie denselben Ausgabepfad erneut verwenden.
- 📁 Verwenden Sie `directory-tree`, wenn Sie `.skt`-Dateien direkt mit Quelldateien in Ihrem Editor korrelieren möchten.
