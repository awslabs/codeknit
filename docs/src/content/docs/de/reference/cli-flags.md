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

Extrahiert strukturelle Informationen aus Quellcode in `.skt`-Dateien.

```bash
codeknit parse <input-path> [output-dir]
```

| Flag             | Typ    | Standardwert     | Beschreibung                                                                                     |
| ---------------- | ------ | ---------------- | ------------------------------------------------------------------------------------------------ |
| `--output-mode`  | string | `directory-flat` | Ausgabemodus: `inline`, `directory-flat` oder `directory-tree`                                   |
| `--max-lines`    | int    | `500`            | Maximale Zeilen pro Ausgabedatei (gilt für `directory-flat` und `directory-tree`-Modi)           |
| `--collect-test` | bool   | `false`          | Testdateien in die Analyse einbeziehen                                                           |
| `--minify`       | bool   | `false`          | Aktiviert die wörterbuchbasierte Ausgabeminimierung                                              |
| `--edges`        | bool   | `false`          | Fügt den `[edges]`-Abschnitt in die Ausgabe ein (standardmäßig deaktiviert, um Tokens zu sparen) |
| `--clean`        | bool   | `false`          | Entfernt veraltete `.skt`-Dateien aus dem Ausgabeverzeichnis vor dem Schreiben                   |
| `--workers`      | int    | `0` (NumCPU)     | Maximale Anzahl gleichzeitiger Parsing-Goroutinen                                                |
| `--verbose`      | bool   | `false`          | Gibt Fortschrittsinformationen während der Verarbeitung aus                                      |

Das Ausgabeverzeichnis ist standardmäßig `./skeleton`, wenn es nicht angegeben wird. Im `inline`-Modus wird die Ausgabe nach stdout geschrieben und kein Verzeichnis verwendet.

## codeknit graph show

Erzeugt eine interaktive HTML-Graphvisualisierung der Codestruktur.

```bash
codeknit graph show <input-path>
```

| Flag             | Typ    | Standardwert                     | Beschreibung                                                |
| ---------------- | ------ | -------------------------------- | ----------------------------------------------------------- |
| `-o`, `--output` | string | `./skeleton/codeknit-graph.html` | Pfad zur HTML-Ausgabedatei                                  |
| `--collect-test` | bool   | `false`                          | Testdateien in die Analyse einbeziehen                      |
| `--workers`      | int    | `0` (NumCPU)                     | Maximale Anzahl gleichzeitiger Parsing-Goroutinen           |
| `--verbose`      | bool   | `false`                          | Gibt Fortschrittsinformationen während der Verarbeitung aus |

Die generierte HTML-Datei ist in sich abgeschlossen und öffnet sich automatisch in Ihrem Standardbrowser.

## codeknit graph analyze

Führt strukturelle Analysealgorithmen aus und gibt einen LLM-lesbaren `.skt`-Bericht aus.

```bash
codeknit graph analyze <input-path>
```

| Flag                      | Typ     | Standardwert                    | Beschreibung                                                                  |
| ------------------------- | ------- | ------------------------------- | ----------------------------------------------------------------------------- |
| `-o`, `--output`          | string  | `./skeleton/graph_analysis.skt` | Pfad zur `.skt`-Ausgabedatei                                                  |
| `--collect-test`          | bool    | `false`                         | Testdateien in die Analyse einbeziehen                                        |
| `--workers`               | int     | `0` (NumCPU)                    | Maximale Anzahl gleichzeitiger Parsing-Goroutinen                             |
| `--verbose`               | bool    | `false`                         | Gibt Fortschrittsinformationen während der Verarbeitung aus                   |
| `--fan-threshold`         | int     | `10`                            | Mindest-Fan-in oder Fan-out, um ein Hub-Symbol zu kennzeichnen                |
| `--god-threshold`         | int     | `15`                            | Mindestanzahl von Contains-Kanten, um eine god class/Funktion zu kennzeichnen |
| `--max-inheritance-depth` | int     | `5`                             | Kennzeichnet Vererbungsketten, die tiefer als dieser Wert sind                |
| `--top-n`                 | int     | `30`                            | Begrenzt die Ranglisten-Ausgabeabschnitte; `0` bedeutet keine Begrenzung      |
| `--betweenness-threshold` | float64 | `0.001`                         | Mindestwert der Betweenness-Zentralität, der berichtet wird                   |
| `--propagation-cutoff`    | float64 | `0.05`                          | Mindestwahrscheinlichkeit, um die Change-Propagation-Simulation fortzusetzen  |

## codeknit fingerprint

Erkennt Duplikate und Beinahe-Duplikate im Code mithilfe von fuzzy hashing.

```bash
codeknit fingerprint <input-path>
```

| Flag               | Typ    | Standardwert                  | Beschreibung                                                                                                                             |
| ------------------ | ------ | ----------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
| `-o`, `--output`   | string | `./skeleton/fingerprints.skt` | Pfad zur `.skt`-Ausgabedatei                                                                                                             |
| `--min-similarity` | int    | `65`                          | Mindest-Ähnlichkeit in Prozent, die berichtet wird (0–100)                                                                               |
| `--max-similarity` | int    | `95`                          | Maximal-Ähnlichkeit in Prozent, die berichtet wird (0–100)                                                                               |
| `--show-all`       | bool   | `false`                       | Fügt den `[fingerprints]`-Abschnitt mit Roh-Tokendaten ein                                                                               |
| `--rerank`         | bool   | `false`                       | Ordnet CTPH-Kandidaten mithilfe semantischer Embeddings via Ollama neu (erfordert `ollama serve` und `ollama pull qwen3-embedding:0.6b`) |
| `--model`          | string | `qwen3-embedding:0.6b`        | Ollama-Embedding-Modell, das mit `--rerank` verwendet wird                                                                               |
| `--collect-test`   | bool   | `false`                       | Testdateien in die Analyse einbeziehen                                                                                                   |
| `--workers`        | int    | `0` (NumCPU)                  | Maximale Anzahl gleichzeitiger Parsing-Goroutinen                                                                                        |
| `--verbose`        | bool   | `false`                       | Gibt Fortschrittsinformationen während der Verarbeitung aus                                                                              |

## codeknit completion

Erzeugt Shell-Vervollständigungsskripte für unterstützte Shells.

```bash
codeknit completion <shell>
```

Unterstützte Shells: `bash`, `zsh`, `fish`, `powershell`.

## Globale Flags

| Flag           | Beschreibung                            |
| -------------- | --------------------------------------- |
| `--version`    | Gibt Versionsinformationen aus          |
| `--help`, `-h` | Zeigt Hilfe für den aktuellen Befehl an |
