---
title: CLI-Referenz
description: Vollständige Referenz für alle codeknit-Befehle und -Flags.
---

Auf dieser Seite sind alle `codeknit`-Befehle, ihre Argumente und Flags aufgelistet. Für anwendungsorientierte Anleitungen siehe die [Anleitung zum Parse-Befehl](../../guides/parse-command/), die [Anleitung zu Graph-Befehlen](../../guides/graph-commands/) und die [Anleitung zum Fingerprint-Befehl](../../guides/fingerprint-command/).

## Befehlsübersicht

| Befehl                                       | Zweck                                                        |
| --------------------------------------------- | -------------------------------------------------------------- |
| `codeknit`                                    | Startet die interaktive Terminal-Benutzeroberfläche.          |
| `codeknit parse <input-path> [output-dir]`    | Extrahiert Symbole und Beziehungen in `.skt`- oder JSON-Ausgabe. |
| `codeknit graph show <input-path>`            | Erzeugt einen interaktiven HTML-Graphen.                      |
| `codeknit graph analyze <input-path>`         | Führt strukturelle Graphalgorithmen aus und schreibt einen `.skt`-Bericht. |
| `codeknit graph hotspots <input-path>`        | Ordnet Änderungs-Hotspots nach Git-Verlauf und Graphstruktur. |
| `codeknit fingerprint <input-path>`           | Erkennt Duplikate und Beinahe-Duplikate im Code.              |
| `codeknit completion <shell>`                 | Generiert ein Shell-Completion-Skript.                        |

Fünf Befehle verarbeiten Quellcode: `parse`, `graph show`, `graph analyze`, `graph hotspots` und `fingerprint`. Sie teilen sich die in [Gemeinsame Flags zur Quellcodeverarbeitung](#gemeinsame-flags-zur-quellcodeverarbeitung) beschriebenen Flags.

## Globale Flags

| Flag              | Typ   | Standardwert | Beschreibung                  |
| ----------------- | ---- | ------- | ---------------------------- |
| `--version`       | bool | `false` | Gibt Versionsinformationen aus. |
| `-h`, `--help`    | bool | `false` | Zeigt die Hilfe für den Befehl an. |

## Gemeinsame Flags zur Quellcodeverarbeitung

Diese Flags erscheinen in der Tabelle jedes quellcodeverarbeitenden Befehls.

| Flag                    | Typ                | Standardwert               | Erlaubte Werte       | Beschreibung                                                                          |
| ----------------------- | ------------------- | --------------------- | -------------------- | ------------------------------------------------------------------------------------ |
| `--header-language`     | string              | `c`                   | `c`, `cpp`           | Parser für `.h`-Dateien.                                                          |
| `--include GLOB`        | string, wiederholbar | Alle unterstützten Quellen | Gitignore-Style-Glob | Enthält passende Quellpfade. Mehrere Muster bilden eine Vereinigung.               |
| `--exclude GLOB`        | string, wiederholbar | Keine                  | Gitignore-Style-Glob | Schließt passende Pfade oder Verzeichnisse aus.                                    |
| `--max-file-size BYTES` | int                 | `10485760` (10 MiB)   | `0` oder größer      | Überspringt Quelldateien, die größer als diese Größe sind. `0` deaktiviert das Limit. |
| `--strict`              | bool                | `false`                | —                    | Bricht ab, bevor die Ausgabe geschrieben wird, wenn Scan- oder Parse-Diagnosen auftreten. |

### `--header-language`

Gilt nur für `.h`-Dateien. Andere Erweiterungen behalten ihren üblichen Parser (`.hpp` und `.hxx` verwenden bereits C++).

- `c` (Standard) bewahrt die Extraktion von C-Makros, Typdefinitionen und Unions.
- `cpp` wählt den C++-Parser. Verwenden Sie dies für C++-Projekte, deren Header `.h` benannt sind.

Die TUI verwendet die Standardauswahl `c` für `.h`-Dateien.

```bash
codeknit parse ./src --header-language cpp
```

Für die getesteten Extraktionsfähigkeiten und bekannten Grenzen pro Sprache siehe [Sprachunterstützung und API-Treue](../language-support/). Diese Seite beschreibt getestete Konstrukte, nicht die vollständige Sprach- oder Compiler-Semantikabdeckung.

### Quellauswahl

- Muster verwenden schrägstrichgetrennte, gitignore-Style-Globs relativ zum Eingabeverzeichnis. Für eine einzelne Datei als Eingabe passen Muster auf den Basisnamen der Datei.
- Muster in Anführungszeichen setzen, damit die Shell sie nicht expandiert.
- Ausschlüsse haben Vorrang vor Einschlüssen.
- Einschlüsse überschreiben nicht `.gitignore`-Regeln, Testdatei-Filterung oder integrierte Ausschlüsse.
- Negationsmuster (`!pattern`) werden in `--include` oder `--exclude` nicht akzeptiert.
- Verzeichnisscans beachten `.gitignore`-Dateien in übergeordneten und untergeordneten Verzeichnissen unabhängig vom aktuellen Arbeitsverzeichnis.
- Die explizite Auswahl einer regulären Quelldatei umgeht Ignorieren und Testfilterung. Einschluss-, Ausschluss- und Größenlimits gelten weiterhin.
- Testdateien werden ausgeschlossen, es sei denn, `--collect-test` ist gesetzt.

```bash
codeknit parse ./src --include '**/*.go' --exclude '**/generated/' --strict
```

### Diagnosen und `--strict`

- Symlinks, Verzeichnissymlinks und nicht reguläre Dateien werden übersprungen und gemeldet.
- Nicht lesbare Pfade, fehlgeschlagene Ignore-Datei-Lesevorgänge, zu große Quellen und Syntaxwarnungen führen zu einer unvollständigen Analyse.
- Standardmäßig werden nutzbare Teilergebnisse beibehalten. Diagnosen erscheinen auf stderr. `parse` zeichnet sie auch in JSON `errors` oder in SKT `[errors]` auf.
- Mit `--strict` wird eine unvollständige Analyse abgelehnt, bevor die Ausgabe geschrieben wird.
- Ein Lauf ohne ausgewählte unterstützte Quellen oder bei dem jeder Quellcode-Lesevorgang fehlschlägt, beendet sich immer erfolglos.
- Die TUI wendet dieselbe Symlink-Richtlinie und das Standardlimit von 10 MiB an. Ihre Formulare bieten die Auswahl- und Strict-Flags nicht an.

Siehe [Analysezuverlässigkeit](../analysis-reliability/) für genaue Garantien.

## `codeknit`

```bash
codeknit
```

Das Ausführen von `codeknit` ohne Unterbefehl startet die interaktive Terminal-Benutzeroberfläche. Die UI bietet die Befehle `parse`, `graph show`, `graph analyze`, `graph hotspots` und `fingerprint` sowie deren Optionen an.

## `codeknit parse`

```bash
codeknit parse <input-path> [output-dir]
```

| Argument       | Erforderlich | Standardwert      | Beschreibung                                                              |
| -------------- | -------- | ------------ | ------------------------------------------------------------------------ |
| `input-path`   | Ja       | —            | Quellverzeichnis oder einzelne Quelldatei.                                  |
| `output-dir`   | Nein     | `./skeleton` | Ausgabeverzeichnis. Wird nicht verwendet, wenn `--output-mode` auf `inline` gesetzt ist. |

| Flag                    | Typ                | Standardwert            | Erlaubte Werte                                   | Beschreibung                                                                                         |
| ----------------------- | ------------------- | ------------------ | ------------------------------------------------ | --------------------------------------------------------------------------------------------------- |
| `--output-mode`         | string              | `directory-flat`   | `inline`, `directory-flat`, `directory-tree`     | Wo und wie die Ausgabe geschrieben wird.                                                            |
| `--format`              | string              | `skt`              | `skt`, `json`                                    | Ausgabeformat.                                                                                      |
| `--max-lines`           | int                 | `500`              | Für `directory-tree` 3 oder größer               | Zeilenziel für flache Ausgabe. Harte Quellcode-Chunk-Grenze für Baumausgabe.                        |
| `--collect-test`        | bool                | `false`            | —                                                | Enthält Testdateien.                                                                                 |
| `--minify`              | bool                | `false`            | —                                                | Aktiviert die komprimierte Ausgabe auf Basis eines Wörterbuchs.                                     |
| `--edges`               | bool                | `false`            | —                                                | Enthält den `[edges]`-Abschnitt. Standardmäßig deaktiviert, um Token zu sparen.                     |
| `--clean`               | bool                | `false`            | —                                                | Ersetzt zuvor generierte Ausgabe nur nach erfolgreicher Generierung.                                |
| `--workers`             | int                 | `0` (`NumCPU`)     | `0` oder größer                                  | Maximale Anzahl gleichzeitiger Parsing-Goroutinen.                                                  |
| `--verbose`             | bool                | `false`            | —                                                | Gibt Fortschrittsinformationen aus.                                                                  |
| `--header-language`     | string              | `c`                | `c`, `cpp`                                       | Parser für `.h`-Dateien.                                                                             |
| `--include GLOB`        | string, wiederholbar | Alle unterstützten Quellen | —                                             | Enthält passende Quellpfade.                                                                         |
| `--exclude GLOB`        | string, wiederholbar | Keine              | —                                                | Schließt passende Pfade oder Verzeichnisse aus.                                                      |
| `--max-file-size BYTES` | int                 | `10485760`         | `0` deaktiviert das Limit                        | Überspringt größere Quelldateien.                                                                    |
| `--strict`              | bool                | `false`            | —                                                | Bricht ab, bevor die Ausgabe geschrieben wird, wenn Scan- oder Parse-Diagnosen auftreten.           |

### Ausgabemodi

| Wert             | Verhalten                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------ |
| `inline`          | Schreibt in stdout. Es wird kein Ausgabeverzeichnis verwendet.                                   |
| `directory-flat`  | Schreibt chunkweise Dateien in das Ausgabeverzeichnis. Quellblöcke bleiben intakt, daher ist `--max-lines` ein Aufteilungsziel und keine harte Grenze. |
| `directory-tree`  | Spiegelt die Quellbaumstruktur. Namen behalten die Quelldateierweiterung (`foo.c.skt`, `foo.h.skt`). Aufgeteilte Dateien verwenden `foo.c_part1.skt`. |

### Ausgabeverhalten

- Ohne `--clean` führt vorhandene generierte Ausgabe für das ausgewählte Format im Ausgabeverzeichnis zu einem Fehler.
- Mit `--clean` werden veraltete `.skt`-Dateien (SKT-Format) oder `codeknit.json` (JSON-Format) nur nach erfolgreicher Generierung ersetzt. Unzusammenhängende Einträge bleiben erhalten.
- Im JSON-Format schreiben Verzeichnismodi `codeknit.json`. JSON ist ein einzelnes Dokument und wird nicht nach Zeilenanzahl aufgeteilt.
- `dict.skt` und `warnings.skt` sind reservierte Hilfsdateinamen.
- Verwenden Sie kein Quellverzeichnis oder eines seiner übergeordneten Verzeichnisse als Ausgabeverzeichnis.

Siehe [Ausgabemodi](../../guides/output-modes/) und die [Referenz zum Ausgabeformat](../output-format/) für Details.

```bash
codeknit parse ./src
codeknit parse ./src ./out --output-mode directory-tree
codeknit parse ./src --output-mode inline --format json --edges
```

## `codeknit graph show`

```bash
codeknit graph show <input-path>
```

| Argument       | Erforderlich | Beschreibung                              |
| -------------- | -------- | ---------------------------------------- |
| `input-path`   | Ja       | Quellverzeichnis oder einzelne Quelldatei. |

| Flag                    | Typ                | Standardwert                            | Erlaubte Werte          | Beschreibung                                                       |
| ----------------------- | ------------------- | ---------------------------------- | ----------------------- | ----------------------------------------------------------------- |
| `-o`, `--output`        | string              | `./skeleton/codeknit-graph.html`   | Dateipfad               | Ausgabedatei im HTML-Format.                                      |
| `--collect-test`        | bool                | `false`                            | —                       | Enthält Testdateien.                                               |
| `--workers`             | int                 | `0` (`NumCPU`)                     | `0` oder größer         | Maximale Anzahl gleichzeitiger Parsing-Goroutinen.                |
| `--verbose`             | bool                | `false`                            | —                       | Gibt Fortschrittsinformationen aus.                                |
| `--header-language`     | string              | `c`                                | `c`, `cpp`              | Parser für `.h`-Dateien.                                           |
| `--include GLOB`        | string, wiederholbar | Alle unterstützten Quellen         | —                       | Enthält passende Quellpfade.                                       |
| `--exclude GLOB`        | string, wiederholbar | Keine                              | —                       | Schließt passende Pfade oder Verzeichnisse aus.                    |
| `--max-file-size BYTES` | int                 | `10485760`                         | `0` deaktiviert das Limit | Überspringt größere Quelldateien.                                  |
| `--strict`              | bool                | `false`                            | —                       | Bricht ab, bevor die Ausgabe geschrieben wird, wenn Scan- oder Parse-Diagnosen auftreten. |

Der Befehl schreibt eine eigenständige HTML-Datei und öffnet sie im Standardbrowser. Der Viewer zeichnet aufgelöste Beziehungen und zeigt die Anzahl unsicherer Beziehungen an.

```bash
codeknit graph show ./src -o graph.html
```

## `codeknit graph analyze`

```bash
codeknit graph analyze <input-path>
```

| Argument       | Erforderlich | Beschreibung                              |
| -------------- | -------- | ---------------------------------------- |
| `input-path`   | Ja       | Quellverzeichnis oder einzelne Quelldatei. |

| Flag                       | Typ                | Standardwert                            | Erlaubte Werte          | Beschreibung                                                  |
| -------------------------- | ------------------- | ---------------------------------- | ----------------------- | ------------------------------------------------------------ |
| `-o`, `--output`           | string              | `./skeleton/graph_analysis.skt`    | Dateipfad               | Ausgabedatei im `.skt`-Format.                                |
| `--collect-test`           | bool                | `false`                            | —                       | Enthält Testdateien.                                          |
| `--workers`                | int                 | `0` (`NumCPU`)                     | `0` oder größer         | Maximale Anzahl gleichzeitiger Parsing-Goroutinen.           |
| `--verbose`                | bool                | `false`                            | —                       | Gibt Fortschrittsinformationen aus.                          |
| `--fan-threshold`          | int                 | `10`                               | —                       | Mindest-Fan-in oder Fan-out, um ein Hub-Symbol zu kennzeichnen. |
| `--god-threshold`          | int                 | `15`                               | —                       | Mindestanzahl von `contains`-Kanten, um eine God-Klasse oder -Funktion zu kennzeichnen. |
| `--max-inheritance-depth`  | int                 | `5`                                | —                       | Kennzeichnet Vererbungsketten, die tiefer als dieser Wert sind. |
| `--top-n`                  | int                 | `30`                               | `0` = kein Limit        | Begrenzt die Abschnitte der gerankten Ausgabe.                |
| `--betweenness-threshold`  | float               | `0.001`                            | —                       | Mindestwert der Betweenness-Centrality, um berichtet zu werden. |
| `--propagation-cutoff`     | float               | `0.05`                             | —                       | Mindestwahrscheinlichkeit, um die Änderungsfortpflanzung fortzusetzen. |
| `--header-language`        | string              | `c`                                | `c`, `cpp`              | Parser für `.h`-Dateien.                                      |
| `--include GLOB`           | string, wiederholbar | Alle unterstützten Quellen         | —                       | Enthält passende Quellpfade.                                  |
| `--exclude GLOB`           | string, wiederholbar | Keine                              | —                       | Schließt passende Pfade oder Verzeichnisse aus.               |
| `--max-file-size BYTES`    | int                 | `10485760`                         | `0` deaktiviert das Limit | Überspringt größere Quelldateien.                             |
| `--strict`                 | bool                | `false`                            | —                       | Bricht ab, bevor die Ausgabe geschrieben wird, wenn Scan- oder Parse-Diagnosen auftreten. |

Abhängigkeitsmetriken schließen unsichere Beziehungen und Datei-/Include-Metadaten aus. Der Bericht zeichnet die Anzahl der ausgeschlossenen Beziehungen auf.

```bash
codeknit graph analyze ./src -o analysis.skt --fan-threshold 15 --top-n 50
```

## `codeknit graph hotspots`

```bash
codeknit graph hotspots <input-path>
```

| Argument       | Erforderlich | Beschreibung                                                  |
| -------------- | -------- | ------------------------------------------------------------ |
| `input-path`   | Ja       | Quellverzeichnis innerhalb eines Git-Repositorys.            |

| Flag                       | Typ                | Standardwert                    | Erlaubte Werte                              | Beschreibung                                                  |
| -------------------------- | ------------------- | -------------------------- | ------------------------------------------- | ------------------------------------------------------------ |
| `-o`, `--output`           | string              | `./skeleton/hotspots.skt`  | Dateipfad                                   | Ausgabedatei.                                                 |
| `--format`                 | string              | `skt`                      | `skt`, `json`                               | Ausgabeformat.                                               |
| `--since`                  | string              | `12mo`                     | Fenster wie `180d`, `12mo` oder `2y`        | Zeitfenster für den Verlauf.                                  |
| `--max-commits`            | int                 | `2000`                     | —                                           | Maximale Anzahl zu prüfender Commits.                        |
| `--max-files-per-commit`   | int                 | `50`                       | —                                           | Schließt Commits aus, die mehr Dateien als dieser Wert berühren. |
| `--min-cochanges`          | int                 | `3`                        | —                                           | Mindestanzahl gemeinsamer Commits, um temporale Kopplung zu berichten. |
| `--top-n`                  | int                 | `30`                       | —                                           | Maximale Ergebnisse pro Abschnitt.                           |
| `--include-merges`         | bool                | `false`                    | —                                           | Enthält Merge-Commits in Verlaufsmetriken.                   |
| `--collect-test`           | bool                | `false`                    | —                                           | Enthält Testdateien.                                          |
| `--workers`                | int                 | `0` (`NumCPU`)             | `0` oder größer                             | Maximale Anzahl gleichzeitiger Parsing-Goroutinen.           |
| `--verbose`                | bool                | `false`                    | —                                           | Gibt Fortschrittsinformationen aus.                          |
| `--header-language`        | string              | `c`                        | `c`, `cpp`                                  | Parser für `.h`-Dateien.                                      |
| `--include GLOB`           | string, wiederholbar | Alle unterstützten Quellen | —                                           | Enthält passende Quellpfade.                                  |
| `--exclude GLOB`           | string, wiederholbar | Keine                      | —                                           | Schließt passende Pfade oder Verzeichnisse aus.               |
| `--max-file-size BYTES`    | int                 | `10485760`                 | `0` deaktiviert das Limit                   | Überspringt größere Quelldateien.                             |
| `--strict`                 | bool                | `false`                    | —                                           | Bricht ab, bevor die Ausgabe geschrieben wird, wenn Scan- oder Parse-Diagnosen auftreten. |

Der Befehl kombiniert Commit-Häufigkeit, Zeilenänderungen und Aktualität mit Datei-Level PageRank, transitivem Fan-in und Betweenness-Centrality. Er berichtet auch über temporale Kopplung zwischen Dateien, die gemeinsam geändert werden. Merge-Commits sind standardmäßig ausgeschlossen.

```bash
codeknit graph hotspots ./src --since 2y --format json -o hotspots.json
```

## `codeknit fingerprint`

```bash
codeknit fingerprint <input-path>
```

| Argument       | Erforderlich | Beschreibung                              |
| -------------- | -------- | ---------------------------------------- |
| `input-path`   | Ja       | Quellverzeichnis oder einzelne Quelldatei. |

| Flag                    | Typ                | Standardwert                          | Erlaubte Werte          | Beschreibung                                                                                                      |
| ----------------------- | ------------------- | -------------------------------- | ----------------------- | ---------------------------------------------------------------------------------------------------------------- |
| `-o`, `--output`        | string              | `./skeleton/fingerprints.skt`    | Dateipfad               | Ausgabedatei im `.skt`-Format.                                                                                     |
| `--min-similarity`      | int                 | `65`                             | `0`–`100`               | Mindest-Ähnlichkeitsprozentsatz, um berichtet zu werden.                                                          |
| `--max-similarity`      | int                 | `95`                             | `0`–`100`               | Maximaler Ähnlichkeitsprozentsatz, um berichtet zu werden.                                                        |
| `--show-all`            | bool                | `false`                          | —                       | Gibt auch den `[fingerprints]`-Abschnitt mit Roh-Tokendaten aus.                                                  |
| `--rerank`              | bool                | `false`                          | —                       | Findet semantische Nachbarn und ordnet Kandidaten mit Ollama-Embeddings neu. Erfordert `ollama serve` und `ollama pull qwen3-embedding:0.6b`. |
| `--model`               | string              | `qwen3-embedding:0.6b`           | Ollama-Modellname       | Ollama-Embedding-Modell, das mit `--rerank` verwendet wird.                                                       |
| `--collect-test`        | bool                | `false`                          | —                       | Enthält Testdateien.                                                                                              |
| `--workers`             | int                 | `0` (`NumCPU`)                   | `0` oder größer         | Maximale Anzahl gleichzeitiger Parsing-Goroutinen.                                                               |
| `--verbose`             | bool                | `false`                          | —                       | Gibt Fortschrittsinformationen aus.                                                                               |
| `--header-language`     | string              | `c`                              | `c`, `cpp`              | Parser für `.h`-Dateien.                                                                                          |
| `--include GLOB`        | string, wiederholbar | Alle unterstützten Quellen      | —                       | Enthält passende Quellpfade.                                                                                      |
| `--exclude GLOB`        | string, wiederholbar | Keine                           | —                       | Schließt passende Pfade oder Verzeichnisse aus.                                                                  |
| `--max-file-size BYTES` | int                 | `10485760`                       | `0` deaktiviert das Limit | Überspringt größere Quelldateien.                                                                                 |
| `--strict`              | bool                | `false`                          | —                       | Bricht ab, bevor die Ausgabe geschrieben wird, wenn Scan- oder Parse-Diagnosen auftreten.                        |

### Ähnlichkeitsbereich

Der Standardbereich von 65–95 % schließt 100 %-Treffer aus. Um exakte normalisierte Token-Treffer zu berichten, setzen Sie beide Grenzen auf `100`:

```bash
codeknit fingerprint ./src --min-similarity 100 --max-similarity 100
```

Exakte Kandidaten mit normalisierten Tokens sind vollständig. Bei 500 oder mehr berechtigten Symbolen verwenden andere Beinahe-Duplikat-Kandidaten eine approximative Top-50-Nachbarsuche. `--rerank` kann einen strukturell exakten Kandidaten entfernen. Siehe [Analysezuverlässigkeit](../analysis-reliability/#garantien-zur-duplikaterkennung) für Details.

## `codeknit completion`

```bash
codeknit completion <shell>
```

| Argument  | Erforderlich | Erlaubte Werte                         | Beschreibung                                   |
| --------- | -------- | -------------------------------------- | --------------------------------------------- |
| `shell`   | Ja       | `bash`, `zsh`, `fish`, `powershell`    | Shell, für die ein Completion-Skript generiert werden soll. |

Das Skript wird nach stdout geschrieben.

```bash
codeknit completion zsh >> ~/.zshrc
```

## Verwandte Seiten

- [Referenz zum Ausgabeformat](../output-format/)
- [Sprachunterstützung und API-Treue](../language-support/)
- [Analysezuverlässigkeit](../analysis-reliability/)