---
title: CLI-Referenz
description: Vollständige Referenz für alle codeknit-Befehle und -Flags.
---

## codeknit

Startet die interaktive Terminal-Benutzeroberfläche (TUI), die Sie durch verfügbare Befehle und Optionen führt.

```bash
codeknit
```

## codeknit parse

Extrahiert strukturelle Informationen aus Quellcode in `.skt`-Dateien oder JSON.

```bash
codeknit parse <input-path> [output-dir]
```

| Flag             | Typ    | Standardwert      | Beschreibung                                                                                     |
| ---------------- | ------ | ----------------- | ------------------------------------------------------------------------------------------------ |
| `--output-mode`  | string | `directory-flat`  | Ausgabemodus: `inline`, `directory-flat` oder `directory-tree`                                  |
| `--format`       | string | `skt`             | Ausgabeformat: `skt` oder `json`                                                                |
| `--max-lines`    | int    | `500`             | Maximale Zeilen pro Ausgabedatei (gilt für `directory-flat`- und `directory-tree`-Modi)         |
| `--collect-test` | bool   | `false`           | Testdateien in die Analyse einbeziehen                                                          |
| `--minify`       | bool   | `false`           | Aktiviert die wörterbuchbasierte Minimierung der Ausgabe                                        |
| `--edges`        | bool   | `false`           | Fügt den `[edges]`-Abschnitt in die Ausgabe ein (standardmäßig deaktiviert, um Tokens zu sparen) |
| `--clean`        | bool   | `false`           | Entfernt veraltete `.skt`-Dateien aus dem Ausgabeverzeichnis vor dem Schreiben                  |
| `--workers`      | int    | `0` (NumCPU)      | Maximale Anzahl gleichzeitiger Parsing-Goroutinen                                              |
| `--verbose`      | bool   | `false`           | Gibt Fortschrittsinformationen während der Verarbeitung aus                                     |

Das Ausgabeverzeichnis ist standardmäßig `./skeleton`, wenn es nicht angegeben wird. Im `inline`-Modus wird die Ausgabe nach stdout geschrieben und kein Verzeichnis verwendet. Mit `--format json` wird die Verzeichnisausgabe als `codeknit.json` geschrieben.

## codeknit graph show

Erzeugt eine interaktive HTML-Graphvisualisierung der Codestruktur.

```bash
codeknit graph show <input-path>
```

| Flag             | Typ    | Standardwert                     | Beschreibung                                      |
| ---------------- | ------ | -------------------------------- | ------------------------------------------------- |
| `-o`, `--output` | string | `./skeleton/codeknit-graph.html` | Pfad zur HTML-Ausgabedatei                        |
| `--collect-test` | bool   | `false`                          | Testdateien in die Analyse einbeziehen            |
| `--workers`      | int    | `0` (NumCPU)                     | Maximale Anzahl gleichzeitiger Parsing-Goroutinen|
| `--verbose`      | bool   | `false`                          | Gibt Fortschrittsinformationen während der Verarbeitung aus |

Die erzeugte HTML-Datei ist in sich abgeschlossen und öffnet sich automatisch in Ihrem Standardbrowser.

## codeknit graph analyze

Führt strukturelle Analysealgorithmen aus und gibt einen LLM-lesbaren `.skt`-Report aus.

```bash
codeknit graph analyze <input-path>
```

| Flag                      | Typ     | Standardwert                    | Beschreibung                                                                                     |
| ------------------------- | ------- | ------------------------------- | ------------------------------------------------------------------------------------------------ |
| `-o`, `--output`          | string  | `./skeleton/graph_analysis.skt` | Pfad zur `.skt`-Ausgabedatei                                                                     |
| `--collect-test`          | bool    | `false`                         | Testdateien in die Analyse einbeziehen                                                          |
| `--workers`               | int     | `0` (NumCPU)                    | Maximale Anzahl gleichzeitiger Parsing-Goroutinen                                              |
| `--verbose`               | bool    | `false`                         | Gibt Fortschrittsinformationen während der Verarbeitung aus                                     |
| `--fan-threshold`         | int     | `10`                            | Mindest-Fan-in oder Fan-out, um ein Hub-Symbol zu kennzeichnen                                  |
| `--god-threshold`         | int     | `15`                            | Mindestanzahl von Contains-Kanten, um eine god class/function zu kennzeichnen                   |
| `--max-inheritance-depth` | int     | `5`                             | Kennzeichnet Vererbungsketten, die tiefer als dieser Wert sind                                  |
| `--top-n`                 | int     | `30`                            | Begrenzt die Ranglisten-Ausgabeabschnitte; `0` bedeutet keine Begrenzung                        |
| `--betweenness-threshold` | float64 | `0.001`                         | Mindestwert der Betweenness-Centrality, um berichtet zu werden                                  |
| `--propagation-cutoff`    | float64 | `0.05`                          | Mindestwahrscheinlichkeit, um die Change-Propagation-Simulation fortzusetzen                    |

## codeknit graph hotspots

Ordnet Dateien nach Git-Verlauf und struktureller Bedeutung und berichtet über temporale Kopplung zwischen Dateien, die wiederholt gemeinsam geändert werden.

```bash
codeknit graph hotspots <input-path>
```

| Flag                     | Typ    | Standardwert              | Beschreibung                                                                                     |
| ------------------------ | ------ | ------------------------- | ------------------------------------------------------------------------------------------------ |
| `-o`, `--output`         | string | `./skeleton/hotspots.skt` | Pfad zur Ausgabedatei                                                                            |
| `--format`               | string | `skt`                     | Ausgabeformat: `skt` oder `json`                                                                |
| `--since`                | string | `12mo`                    | Zeitfenster des Verlaufs, z. B. `180d`, `12mo` oder `2y`                                        |
| `--max-commits`          | int    | `2000`                    | Maximale Anzahl zu prüfender Commits                                                            |
| `--max-files-per-commit` | int    | `50`                      | Ausschluss von Commits, die mehr Dateien ändern                                                 |
| `--min-cochanges`        | int    | `3`                       | Mindestanzahl gemeinsamer Commits für temporale Kopplung                                         |
| `--top-n`                | int    | `30`                      | Maximale Ergebnisse pro Berichtsabschnitt                                                       |
| `--include-merges`       | bool   | `false`                   | Merge-Commits einbeziehen                                                                        |
| `--collect-test`         | bool   | `false`                   | Testdateien einbeziehen                                                                          |
| `--workers`              | int    | `0` (NumCPU)              | Maximale Anzahl gleichzeitiger Parsing-Goroutinen                                              |
| `--verbose`              | bool   | `false`                   | Gibt Fortschrittsinformationen aus                                                              |

## codeknit fingerprint

Erkennt Duplikate und Beinahe-Duplikate im Code mithilfe von fuzzy hashing.

```bash
codeknit fingerprint <input-path>
```

| Flag               | Typ    | Standardwert                   | Beschreibung                                                                                     |
| ------------------ | ------ | ------------------------------ | ------------------------------------------------------------------------------------------------ |
| `-o`, `--output`   | string | `./skeleton/fingerprints.skt`  | Pfad zur `.skt`-Ausgabedatei                                                                     |
| `--min-similarity` | int    | `65`                           | Mindest-Ähnlichkeit in Prozent, um berichtet zu werden (0–100)                                  |
| `--max-similarity` | int    | `95`                           | Höchst-Ähnlichkeit in Prozent, um berichtet zu werden (0–100)                                   |
| `--show-all`       | bool   | `false`                        | Fügt den `[fingerprints]`-Abschnitt mit Roh-Tokendaten ein                                       |
| `--rerank`         | bool   | `false`                        | Ordnet CTPH-Kandidaten mithilfe semantischer Embeddings via Ollama neu (erfordert `ollama serve` und `ollama pull qwen3-embedding:0.6b`) |
| `--model`          | string | `qwen3-embedding:0.6b`         | Ollama-Embedding-Modell, das mit `--rerank` verwendet wird                                       |
| `--collect-test`   | bool   | `false`                        | Testdateien in die Analyse einbeziehen                                                          |
| `--workers`        | int    | `0` (NumCPU)                   | Maximale Anzahl gleichzeitiger Parsing-Goroutinen                                              |
| `--verbose`        | bool   | `false`                        | Gibt Fortschrittsinformationen während der Verarbeitung aus                                     |

## codeknit completion

Generiert Shell-Vervollständigungsskripte für unterstützte Shells.

```bash
codeknit completion <shell>
```

Unterstützte Shells: `bash`, `zsh`, `fish`, `powershell`.

## Globale Flags

| Flag           | Beschreibung                       |
| -------------- | ---------------------------------- |
| `--version`    | Gibt Versionsinformationen aus     |
| `--help`, `-h` | Zeigt Hilfe für den aktuellen Befehl an |