---
title: Graph-Befehle
description: Visualisieren und analysieren Sie die Codestruktur Ihrer Codebasis mit Graphalgorithmen.
---

codeknit bietet zwei leistungsstarke Graph-Befehle, um Ihnen zu helfen, die Struktur Ihrer Codebasis zu verstehen und zu verbessern: `graph show` für interaktive Visualisierung und `graph analyze` für automatisierte strukturelle Analyse.

## graph show

Erzeugt eine interaktive HTML-Graphvisualisierung Ihrer Codebasis.

```bash
codeknit graph show <input-path>
```

Dieser Befehl parst Ihre Codebasis und erzeugt eine eigenständige HTML-Datei mit einer interaktiven Graphvisualisierung. Symbole (Funktionen, Klassen, Typen) erscheinen als Knoten, und ihre Beziehungen (Aufrufe, enthält, implementiert) als Kanten. Die Visualisierung öffnet sich automatisch in Ihrem Standardbrowser.

### Flags

| Flag             | Standardwert                     | Beschreibung                                  |
| ---------------- | -------------------------------- | -------------------------------------------- |
| `-o`, `--output` | `./skeleton/codeknit-graph.html` | Pfad der Ausgabedatei (HTML)                 |
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

Führt strukturelle Graphalgorithmen auf Ihrer Codebasis aus und gibt einen LLM-lesbaren `.skt`-Bericht mit Codequalitätserkenntnissen aus.

```bash
codeknit graph analyze <input-path>
```

Dieser Befehl erkennt gängige Codequalitätsprobleme wie zyklische Abhängigkeiten, Hub-Symbole, toten Code, God-Klassen und architektonische Engpässe.

### Algorithmen

Die Analyse umfasst 17 strukturelle Graphalgorithmen:

- Zyklische Abhängigkeiten (Tarjans SCC)
- Hub-Erkennung (hohe Fan-in/Fan-out-Kopplung)
- Waisen-Erkennung (Kandidaten für toten Code)
- God-Klasse/Funktion-Erkennung (übermäßige Kinder)
- Instabilitätsmetrik (Robert C. Martins Ce/(Ca+Ce))
- Tiefe Vererbungsketten
- Betweenness-Zentralität (Engpass-Erkennung)
- Artikulationspunkte (einzelne Fehlerquellen)
- PageRank (rekursive Wichtigkeit)
- Transitiver Fan-in (Auswirkungsradius)
- Änderungsausbreitungssimulation
- Zirkuläre Paketabhängigkeiten
- Schichtverletzungserkennung
- Erreichbarkeit von Einstiegspunkten
- Schwach verbundene Komponenten
- Abhängigkeitsgewicht (Paketkopplungsstärke)
- Abstand von der Main Sequence (A+I-Balance)

### Flags

| Flag                      | Standardwert                     | Beschreibung                                              |
| ------------------------- | ------------------------------- | -------------------------------------------------------- |
| `-o`, `--output`          | `./skeleton/graph_analysis.skt` | Pfad der `.skt`-Ausgabedatei                             |
| `--collect-test`          | `false`                         | Testdateien in die Analyse einbeziehen                   |
| `--workers`               | `NumCPU`                        | Maximale Anzahl paralleler Parsing-Goroutinen            |
| `--verbose`               | `false`                         | Fortschrittsinformationen während der Verarbeitung anzeigen |
| `--fan-threshold`         | `10`                            | Mindest-Fan-in oder -Fan-out, um ein Hub-Symbol zu kennzeichnen |
| `--god-threshold`         | `15`                            | Mindestanzahl von Contains-Kanten, um eine God-Klasse/Funktion zu kennzeichnen |
| `--max-inheritance-depth` | `5`                             | Vererbungsketten kennzeichnen, die tiefer als dieser Wert sind |
| `--top-n`                 | `30`                            | Begrenzung der gerankten Ausgabesektionen; 0 = keine Begrenzung |
| `--betweenness-threshold` | `0.001`                         | Mindestwert der Betweenness-Zentralität, der gemeldet wird |
| `--propagation-cutoff`    | `0.05`                          | Mindestwahrscheinlichkeit, um die Änderungsausbreitung fortzusetzen |

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