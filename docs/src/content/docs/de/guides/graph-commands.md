---
title: Graph-Befehle
description: Visualisieren und analysieren Sie die Struktur Ihres Codebase mit Graphalgorithmen.
---

codeknit bietet Graph-Befehle zur Visualisierung der Struktur, zur automatisierten Analyse und zur Kombination des aktuellen Abhängigkeitsgraphen mit der Git-Änderungshistorie.

## graph show

Erzeugt eine interaktive HTML-Graphvisualisierung Ihres Codebase.

```bash
codeknit graph show <input-path>
```

Dieser Befehl parst Ihren Codebase und erzeugt eine eigenständige HTML-Datei mit einer interaktiven Graphvisualisierung. Symbole (Funktionen, Klassen, Typen) erscheinen als Knoten, und ihre Beziehungen (Aufrufe, enthält, implementiert) als Kanten. Die Visualisierung öffnet sich automatisch in Ihrem Standardbrowser.

### Flags

| Flag             | Standardwert                     | Beschreibung                                  |
| ---------------- | -------------------------------- | -------------------------------------------- |
| `-o`, `--output` | `./skeleton/codeknit-graph.html` | Pfad zur Ausgabedatei (HTML)                 |
| `--collect-test` | `false`                          | Testdateien in die Analyse einbeziehen       |
| `--workers`      | `NumCPU`                         | Maximale Anzahl paralleler Parsing-Goroutinen |
| `--verbose`      | `false`                          | Fortschrittsinformationen während der Verarbeitung anzeigen |

### Beispiele

```skt
# Standardvisualisierung erzeugen
codeknit graph show ./myproject

# Benutzerdefinierte Ausgabedatei
codeknit graph show ./myproject -o graph.html

# Testdateien einbeziehen
codeknit graph show ./src --collect-test
```

## graph analyze

Führt strukturelle Graphalgorithmen auf Ihrem Codebase aus und gibt einen LLM-lesbaren `.skt`-Bericht mit Codequalitätserkenntnissen aus.

```bash
codeknit graph analyze <input-path>
```

Dieser Befehl erkennt häufige Codequalitätsprobleme wie zyklische Abhängigkeiten, Hub-Symbole, toten Code, God Classes und architektonische Engpässe.

### Algorithmen

Die Analyse umfasst 22 strukturelle Graphalgorithmen:

- Zyklische Abhängigkeiten (Tarjans SCC)
- Hub-Erkennung (hohe Fan-in/Fan-out-Kopplung)
- Waisen-Erkennung (Kandidaten für toten Code)
- God Class/Function-Erkennung (übermäßige Kinder)
- Instabilitätsmetrik (Robert C. Martins Ce/(Ca+Ce))
- Tiefe Vererbungsketten
- Betweenness Centrality (Engpasserkennung)
- Artikulationspunkte (einzelne Fehlerquellen)
- PageRank (rekursive Wichtigkeit)
- Transitives Fan-in (Auswirkungsradius)
- Change-Propagation-Simulation
- Zirkuläre Paketabhängigkeiten
- Layer-Verletzungserkennung
- Erreichbarkeit von Einstiegspunkten
- Schwach verbundene Komponenten
- Abhängigkeitsgewicht (Paketkopplungsstärke)
- Distanz zur Main Sequence (A+I-Balance)
- Shotgun-Surgery-Erkennung
- Feature-Envy-Erkennung
- Stable-Dependency-Verletzungen
- Interface-Segregation-Verletzungen
- Containment-Tiefe

### Flags

| Flag                      | Standardwert                     | Beschreibung                                              |
| ------------------------- | ------------------------------- | -------------------------------------------------------- |
| `-o`, `--output`          | `./skeleton/graph_analysis.skt` | Pfad zur `.skt`-Ausgabedatei                             |
| `--collect-test`          | `false`                          | Testdateien in die Analyse einbeziehen                   |
| `--workers`               | `NumCPU`                         | Maximale Anzahl paralleler Parsing-Goroutinen            |
| `--verbose`               | `false`                          | Fortschrittsinformationen während der Verarbeitung anzeigen |
| `--fan-threshold`         | `10`                             | Mindest-Fan-in oder -Fan-out, um ein Hub-Symbol zu kennzeichnen |
| `--god-threshold`         | `15`                             | Mindestanzahl von Contains-Kanten, um eine God Class/Function zu kennzeichnen |
| `--max-inheritance-depth` | `5`                              | Vererbungsketten kennzeichnen, die tiefer als dieser Wert sind |
| `--top-n`                 | `30`                             | Begrenzung der ausgegebenen Ranglistenabschnitte; 0 = keine Begrenzung |
| `--betweenness-threshold` | `0.001`                          | Mindestwert der Betweenness Centrality für die Berichterstattung |
| `--propagation-cutoff`    | `0.05`                           | Mindestwahrscheinlichkeit für die Fortsetzung der Change Propagation |

### Beispiele

```skt
# Strukturelle Analyse mit Standardwerten ausführen
codeknit graph analyze ./myproject

# Benutzerdefinierte Ausgabe und Schwellenwerte
codeknit graph analyze ./myproject -o analysis.skt --fan-threshold 15

# Mehr Ergebnisse pro Abschnitt anzeigen
codeknit graph analyze ./myproject --top-n 50

# Testdateien einbeziehen
codeknit graph analyze ./src --collect-test
```

## graph hotspots

Rankt Dateien, die sowohl häufig geändert werden als auch strukturell wichtig sind:

```bash
codeknit graph hotspots <input-path>
```

Der Score kombiniert Commit-Häufigkeit, Zeilenänderungen und Aktualität mit Datei-Level PageRank, transitivem Fan-in und Betweenness Centrality. Der Bericht identifiziert auch temporale Kopplung zwischen Dateien, die wiederholt in denselben Commits geändert werden.

Merge-Commits sind standardmäßig ausgeschlossen. Commits, die mehr als 50 Dateien ändern, werden ebenfalls ausgeschlossen, damit generierte, vendorte oder mechanische Massenänderungen die Ergebnisse nicht verzerren.

### Flags

| Flag                     | Standardwert                 | Beschreibung                                      |
| ------------------------ | ---------------------------- | ------------------------------------------------ |
| `-o`, `--output`         | `./skeleton/hotspots.skt`    | Pfad zur Ausgabedatei                            |
| `--format`               | `skt`                        | Ausgabemodus: `skt` oder `json`                  |
| `--since`                | `12mo`                       | Zeitfenster der Historie, z. B. `180d`, `12mo` oder `2y` |
| `--max-commits`          | `2000`                       | Maximale Anzahl zu prüfender Commits             |
| `--max-files-per-commit` | `50`                         | Commits mit mehr geänderten Dateien ausschließen |
| `--min-cochanges`        | `3`                          | Mindestanzahl gemeinsamer Commits für temporale Kopplung |
| `--top-n`                | `30`                         | Maximale Ergebnisse pro Berichtsabschnitt        |
| `--include-merges`       | `false`                      | Merge-Commits einbeziehen                        |
| `--collect-test`         | `false`                      | Testdateien einbeziehen                          |
| `--workers`              | `NumCPU`                     | Maximale Anzahl paralleler Parsing-Goroutinen    |
| `--verbose`              | `false`                      | Fortschrittsinformationen anzeigen               |

### Beispiele

```bash
# Die letzten 12 Monate analysieren
codeknit graph hotspots ./myproject

# Zwei Jahre analysieren und JSON ausgeben
codeknit graph hotspots ./myproject --since 2y --format json -o hotspots.json

# Größere Commits einbeziehen und stärkere Kopplung erfordern
codeknit graph hotspots . --max-files-per-commit 100 --min-cochanges 5
```