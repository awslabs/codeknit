---
title: CLI-Referenz
description: Vollständige Referenz für alle codeknit-Befehle und Flags.
---

Alle Befehle zur Quellcode-Verarbeitung akzeptieren `--header-language c|cpp` für `.h`-Dateien. Der Standardwert ist `c`; wählen Sie `cpp` für C++-Header mit der Endung `.h`. Andere Dateiendungen sind davon nicht betroffen. Weitere Informationen finden Sie unter [Sprachunterstützung](../language-support/).

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

| Flag             | Typ    | Standardwert     | Beschreibung                                                                                   |
| ---------------- | ------ | ---------------- | ---------------------------------------------------------------------------------------------- |
| `--output-mode`  | string | `directory-flat` | Ausgabemodus: `inline`, `directory-flat` oder `directory-tree`                                 |
| `--format`       | string | `skt`            | Ausgabeformat: `skt` oder `json`                                                               |
| `--max-lines`    | int    | `500`            | Maximale Zeilen pro Ausgabedatei (gilt für `directory-flat` und `directory-tree`-Modi)         |
| `--collect-test` | bool   | `false`          | Testdateien in die Analyse einbeziehen                                                         |
| `--minify`       | bool   | `false`          | Dictionary-basierte Ausgabeminimierung aktivieren                                              |
| `--edges`        | bool   | `false`          | `[edges]`-Abschnitt in die Ausgabe einbeziehen (standardmäßig deaktiviert, um Token zu sparen) |
| `--clean`        | bool   | `false`          | Veraltete `.skt`-Dateien aus dem Ausgabeverzeichnis vor dem Schreiben entfernen                |
| `--workers`      | int    | `0` (NumCPU)     | Maximale Anzahl gleichzeitiger Parsing-Goroutinen                                              |
| `--verbose`      | bool   | `false`          | Fortschrittsinformationen während der Verarbeitung ausgeben                                    |

Das Ausgabeverzeichnis ist standardmäßig `./skeleton`, wenn es nicht angegeben wird. Im `inline`-Modus wird die Ausgabe nach stdout geschrieben und kein Verzeichnis verwendet. Bei `--format json` wird die Verzeichnisausgabe als `codeknit.json` geschrieben.

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
| `--verbose`      | bool   | `false`                          | Fortschrittsinformationen während der Verarbeitung ausgeben |

Die erzeugte HTML-Datei ist in sich abgeschlossen und öffnet sich automatisch in Ihrem Standardbrowser.

## codeknit graph analyze

Führt strukturelle Analysealgorithmen aus und gibt einen LLM-lesbaren `.skt`-Bericht aus.

```bash
codeknit graph analyze <input-path>
```

| Flag                      | Typ     | Standardwert                    | Beschreibung                                                                   |
| ------------------------- | ------- | ------------------------------- | ------------------------------------------------------------------------------ |
| `-o`, `--output`          | string  | `./skeleton/graph_analysis.skt` | Pfad zur `.skt`-Ausgabedatei                                                   |
| `--collect-test`          | bool    | `false`                         | Testdateien in die Analyse einbeziehen                                         |
| `--workers`               | int     | `0` (NumCPU)                    | Maximale Anzahl gleichzeitiger Parsing-Goroutinen                              |
| `--verbose`               | bool    | `false`                         | Fortschrittsinformationen während der Verarbeitung ausgeben                    |
| `--fan-threshold`         | int     | `10`                            | Mindest-Fan-in oder Fan-out, um ein Hub-Symbol zu kennzeichnen                 |
| `--god-threshold`         | int     | `15`                            | Mindestanzahl von Contains-Kanten, um eine God-Klasse/Funktion zu kennzeichnen |
| `--max-inheritance-depth` | int     | `5`                             | Erbtiefe-Ketten kennzeichnen, die diesen Wert überschreiten                    |
| `--top-n`                 | int     | `30`                            | Begrenzung der gerankten Ausgabesektionen; `0` bedeutet keine Begrenzung       |
| `--betweenness-threshold` | float64 | `0.001`                         | Mindestwert der Betweenness-Zentralität, der berichtet wird                    |
| `--propagation-cutoff`    | float64 | `0.05`                          | Mindestwahrscheinlichkeit, um die Change-Propagation-Simulation fortzusetzen   |

## codeknit graph hotspots

Rankt Dateien anhand der Git-Historie und struktureller Bedeutung und berichtet über temporale Kopplung zwischen Dateien, die wiederholt gemeinsam geändert werden.

```bash
codeknit graph hotspots <input-path>
```

| Flag                     | Typ    | Standardwert              | Beschreibung                                             |
| ------------------------ | ------ | ------------------------- | -------------------------------------------------------- |
| `-o`, `--output`         | string | `./skeleton/hotspots.skt` | Pfad zur Ausgabedatei                                    |
| `--format`               | string | `skt`                     | Ausgabeformat: `skt` oder `json`                         |
| `--since`                | string | `12mo`                    | Zeitfenster der Historie, z. B. `180d`, `12mo` oder `2y` |
| `--max-commits`          | int    | `2000`                    | Maximale Anzahl zu prüfender Commits                     |
| `--max-files-per-commit` | int    | `50`                      | Commits mit mehr geänderten Dateien ausschließen         |
| `--min-cochanges`        | int    | `3`                       | Mindestanzahl gemeinsamer Commits für temporale Kopplung |
| `--top-n`                | int    | `30`                      | Maximale Ergebnisse pro Berichtsabschnitt                |
| `--include-merges`       | bool   | `false`                   | Merge-Commits einbeziehen                                |
| `--collect-test`         | bool   | `false`                   | Testdateien einbeziehen                                  |
| `--workers`              | int    | `0` (NumCPU)              | Maximale Anzahl gleichzeitiger Parsing-Goroutinen        |
| `--verbose`              | bool   | `false`                   | Fortschrittsinformationen ausgeben                       |

## codeknit fingerprint

Erkennt Duplikate und Beinahe-Duplikate im Code mithilfe von Fuzzy Hashing.

```bash
codeknit fingerprint <input-path>
```

| Flag               | Typ    | Standardwert                  | Beschreibung                                                                                                                                           |
| ------------------ | ------ | ----------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `-o`, `--output`   | string | `./skeleton/fingerprints.skt` | Pfad zur `.skt`-Ausgabedatei                                                                                                                           |
| `--min-similarity` | int    | `65`                          | Mindest-Ähnlichkeit in Prozent (0–100), die berichtet wird                                                                                             |
| `--max-similarity` | int    | `95`                          | Höchst-Ähnlichkeit in Prozent (0–100), die berichtet wird                                                                                              |
| `--show-all`       | bool   | `false`                       | `[fingerprints]`-Abschnitt mit Roh-Tokendaten einbeziehen                                                                                              |
| `--rerank`         | bool   | `false`                       | Semantische Nachbarn finden und Kandidaten mithilfe von Ollama-Embeddings neu ranken (erfordert `ollama serve` und `ollama pull qwen3-embedding:0.6b`) |
| `--model`          | string | `qwen3-embedding:0.6b`        | Ollama-Embedding-Modell, das mit `--rerank` verwendet wird                                                                                             |
| `--collect-test`   | bool   | `false`                       | Testdateien in die Analyse einbeziehen                                                                                                                 |
| `--workers`        | int    | `0` (NumCPU)                  | Maximale Anzahl gleichzeitiger Parsing-Goroutinen                                                                                                      |
| `--verbose`        | bool   | `false`                       | Fortschrittsinformationen während der Verarbeitung ausgeben                                                                                            |

## codeknit completion

Generiert Shell-Vervollständigungsskripte für unterstützte Shells.

```bash
codeknit completion <shell>
```

Unterstützte Shells: `bash`, `zsh`, `fish`, `powershell`.

## Globale Flags

| Flag           | Beschreibung                            |
| -------------- | --------------------------------------- |
| `--version`    | Versionsinformationen ausgeben          |
| `--help`, `-h` | Hilfe für den aktuellen Befehl anzeigen |
