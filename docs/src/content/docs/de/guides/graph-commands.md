---
title: Graph-Befehle
description: Visualisieren und analysieren Sie die Struktur Ihres Codebase mit Graph-Algorithmen.
---

codeknit bietet Graph-Befehle zur Visualisierung der Struktur, zur Durchführung automatisierter Analysen und zur Kombination des aktuellen Abhängigkeitsgraphen mit der Git-Änderungshistorie.

## graph show

Erzeugt eine interaktive HTML-Graph-Visualisierung Ihres Codebase.

```bash
codeknit graph show <input-path>
```

Dieser Befehl parst Ihr Codebase und erzeugt eine eigenständige HTML-Datei mit einer interaktiven Graph-Visualisierung. Symbole (Funktionen, Klassen, Typen) erscheinen als Knoten, und ihre Beziehungen (Aufrufe, enthält, implementiert) als Kanten. Die Visualisierung öffnet sich automatisch in Ihrem Standardbrowser.

Der Viewer zeichnet aufgelöste Beziehungen. Er zeigt eine Unsicherheitsanzahl an, zeichnet jedoch keine einzelnen unaufgelösten oder mehrdeutigen Beziehungen oder deren Kandidatenlinks. Überprüfen Sie diese Beziehungen mit der Parse-Ausgabe:

```bash
codeknit parse ./src --output-mode inline --edges
codeknit parse ./src --output-mode inline --format json --edges
```

### Flags

| Flag             | Standardwert                     | Beschreibung                                  |
| ---------------- | -------------------------------- | -------------------------------------------- |
| `-o`, `--output` | `./skeleton/codeknit-graph.html` | Pfad zur Ausgabe-HTML-Datei                  |
| `--collect-test` | `false`                          | Testdateien in die Analyse einbeziehen       |
| `--workers`      | `NumCPU`                         | Maximale Anzahl paralleler Parsing-Goroutinen|
| `--verbose`      | `false`                          | Fortschrittsinformationen während der Verarbeitung anzeigen |

### Beispiele

```bash
# Generate default visualization
codeknit graph show ./myproject

# Custom output file
codeknit graph show ./myproject -o graph.html

# Include test files
codeknit graph show ./src --collect-test
```

## graph analyze

Führt strukturelle Graph-Algorithmen auf Ihrem Codebase aus und erzeugt einen LLM-lesbaren `.skt`-Bericht mit Codequalitäts-Einsichten.

```bash
codeknit graph analyze <input-path>
```

Dieser Befehl meldet Kandidaten zur Überprüfung, einschließlich zyklischer Abhängigkeiten, Hub-Symbolen, totem Code, God-Klassen und architektonischen Engpässen.

### Algorithmen

Die Analyse umfasst 22 strukturelle Graph-Algorithmen:

- Zyklische Abhängigkeiten (Tarjans SCC)
- Hub-Erkennung (hohe Fan-in/Fan-out-Kopplung)
- Waisen-Erkennung (Kandidaten für toten Code)
- God-Klasse/Funktion-Erkennung (übermäßige Kinder)
- Instabilitätsmetrik (Robert C. Martins Ce/(Ca+Ce))
- Tiefe Vererbungsketten
- Betweenness-Zentralität (Engpass-Erkennung)
- Artikulationspunkte (einzelne Fehlerpunkte)
- PageRank (rekursive Wichtigkeit)
- Transitives Fan-in (Auswirkungsradius)
- Änderungssimulation
- Zirkuläre Paketabhängigkeiten
- Schichtverletzungserkennung
- Erreichbarkeit von Einstiegspunkten
- Schwach verbundene Komponenten
- Abhängigkeitsgewicht (Paketkopplungsstärke)
- Abstand von der Main Sequence (A+I-Balance)
- Schrotflinten-Chirurgie-Erkennung
- Feature-Neid-Erkennung
- Stabile Abhängigkeitsverletzungen
- Interface-Segregationsverletzungen
- Containment-Tiefe

### Interpretation des Berichts

Abhängigkeitsmetriken verwenden aufgelöste Beziehungen. Der Bericht zählt ausgeschlossene unaufgelöste und mehrdeutige Beziehungen. Datei-/Include-Metadaten sind von Abhängigkeitsrankings und Dead-Code-Findings ausgeschlossen.

- Fan-in und Fan-out zählen unterschiedliche Nachbarn. Wiederholte Beziehungen erhöhen diese Zählungen nicht, und Containment gilt nicht als Abhängigkeit.
- Vererbungstiefe berücksichtigt alle Basen und Interfaces. Zyklische Hierarchien wird keine Tiefe zugewiesen.
- Erreichbarkeit beginnt bei Callables mit den Namen `main`, `Main` oder `init` und folgt Abhängigkeitskanten. Das Erreichen eines Mitglieds macht seinen Container relevant, ohne jedes Geschwisterkind als erreichbar zu markieren. Ohne einen erkannten Einstiegspunkt werden Erreichbarkeits-Findings ausgelassen.
- Callback-Argumente und zurückgegebene Callables sind Referenzen, kein Beweis dafür, dass der Callback ausgeführt wird.
- Änderungssimulation verwendet den stärksten Abhängigkeitspfad. Ihre Scores und verzeichnisbasierten Schichtregeln sind Heuristiken, keine gemessenen Wahrscheinlichkeiten.

Externe Aufrufer, dynamische Dispatches und Abhängigkeiten außerhalb des Scans können die Schlussfolgerungen ändern. Weitere Informationen finden Sie unter [Sprachunterstützung und API-Treue](../../reference/language-support/).

### Flags

| Flag                      | Standardwert                     | Beschreibung                                              |
| ------------------------- | ------------------------------- | -------------------------------------------------------- |
| `-o`, `--output`          | `./skeleton/graph_analysis.skt` | Pfad zur `.skt`-Ausgabedatei                             |
| `--collect-test`          | `false`                         | Testdateien in die Analyse einbeziehen                   |
| `--workers`               | `NumCPU`                        | Maximale Anzahl paralleler Parsing-Goroutinen            |
| `--verbose`               | `false`                         | Fortschrittsinformationen während der Verarbeitung anzeigen |
| `--fan-threshold`         | `10`                            | Mindest-Fan-in oder Fan-out, um ein Hub-Symbol zu kennzeichnen |
| `--god-threshold`         | `15`                            | Mindestanzahl von Contains-Kanten, um eine God-Klasse/Funktion zu kennzeichnen |
| `--max-inheritance-depth` | `5`                             | Vererbungsketten kennzeichnen, die tiefer sind als dieser Wert |
| `--top-n`                 | `30`                            | Begrenzung der gerankten Ausgabesektionen; 0 = keine Begrenzung |
| `--betweenness-threshold` | `0.001`                         | Mindestwert der Betweenness-Zentralität, der gemeldet wird |
| `--propagation-cutoff`    | `0.05`                          | Mindest-Heuristik-Score, um die Änderungssimulation fortzusetzen |

### Beispiele

```bash
# Run structural analysis with defaults
codeknit graph analyze ./myproject

# Custom output and thresholds
codeknit graph analyze ./myproject -o analysis.skt --fan-threshold 15

# Show more results per section
codeknit graph analyze ./myproject --top-n 50

# Remove the cap on ranked sections
codeknit graph analyze ./myproject --top-n 0

# Include test files
codeknit graph analyze ./src --collect-test
```

## graph hotspots

Rankt Dateien, die sowohl häufig geändert als auch strukturell wichtig sind:

```bash
codeknit graph hotspots <input-path>
```

Der Score kombiniert Commit-Häufigkeit, Line Churn und Aktualität mit Datei-Level PageRank, transitivem Fan-in und Betweenness-Zentralität. Der Bericht identifiziert auch temporale Kopplung zwischen Dateien, die wiederholt in denselben Commits geändert werden.

Merge-Commits sind standardmäßig ausgeschlossen. Commits, die mehr als 50 Dateien ändern, werden ebenfalls ausgeschlossen, damit generierte, vendorte oder mechanische Massenänderungen die Ergebnisse nicht verzerren.

### Flags

| Flag                     | Standardwert               | Beschreibung                                     |
| ------------------------ | ------------------------- | ----------------------------------------------- |
| `-o`, `--output`         | `./skeleton/hotspots.skt` | Pfad zur Ausgabedatei                           |
| `--format`               | `skt`                     | Ausgabeformat: `skt` oder `json`                |
| `--since`                | `12mo`                    | Zeitfenster der Historie, z. B. `180d`, `12mo` oder `2y` |
| `--max-commits`          | `2000`                    | Maximale Anzahl zu prüfender Commits            |
| `--max-files-per-commit` | `50`                      | Commits ausschließen, die mehr Dateien ändern   |
| `--min-cochanges`        | `3`                       | Mindestanzahl gemeinsamer Commits für temporale Kopplung |
| `--top-n`                | `30`                      | Maximale Ergebnisse pro Berichtsabschnitt       |
| `--include-merges`       | `false`                   | Merge-Commits einbeziehen                       |
| `--collect-test`         | `false`                   | Testdateien einbeziehen                         |
| `--workers`              | `NumCPU`                  | Maximale Anzahl paralleler Parsing-Goroutinen   |
| `--verbose`              | `false`                   | Fortschrittsinformationen anzeigen              |

### Beispiele

```bash
# Analyze the last 12 months
codeknit graph hotspots ./myproject

# Analyze two years and emit JSON
codeknit graph hotspots ./myproject --since 2y --format json -o hotspots.json

# Include larger commits and require stronger coupling
codeknit graph hotspots . --max-files-per-commit 100 --min-cochanges 5
```