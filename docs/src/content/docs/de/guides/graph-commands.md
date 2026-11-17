---
title: Graph-Befehle
description: Visualisieren und analysieren Sie die Codestruktur Ihres Projekts mit Graphalgorithmen.
---

Verwenden Sie die `graph`-Unterbefehle, um zu sehen, wie Ihr Code verbunden ist, strukturelle Probleme zu finden und zu entscheiden, welche Dateien zuerst refaktorisiert werden sollten. Dieser Leitfaden setzt voraus, dass Sie [codeknit installiert haben](../../installation/) und [`codeknit parse`](../parse-command/) ausführen können.

| Aufgabe | Befehl |
| --- | --- |
| Den Graphen visuell erkunden | `codeknit graph show` |
| Strukturelle Probleme finden | `codeknit graph analyze` |
| Volatile, wichtige Dateien finden | `codeknit graph hotspots` |

## Visualisieren des Codegraphen

`graph show` parst die Eingabe und schreibt eine eigenständige HTML-Datei. Symbole (Funktionen, Klassen, Typen) sind Knoten. Beziehungen (Aufrufe, enthält, implementiert) sind Kanten.

```bash
codeknit graph show ./myproject
```

Die Datei wird nach `./skeleton/codeknit-graph.html` geschrieben und öffnet sich in Ihrem Standardbrowser.

Um den Ausgabepfad zu wählen oder Testdateien einzuschließen:

```bash
codeknit graph show ./myproject -o graph.html
codeknit graph show ./src --collect-test
```

| Flag | Standardwert | Zweck |
| --- | --- | --- |
| `-o`, `--output` | `./skeleton/codeknit-graph.html` | Pfad zur HTML-Ausgabedatei |
| `--collect-test` | `false` | Testdateien einschließen |
| `--workers` | `0` (`NumCPU`) | Maximale Anzahl gleichzeitiger Parsing-Goroutinen |
| `--verbose` | `false` | Fortschrittsinformationen ausgeben |

### Unaufgelöste Beziehungen inspizieren

Der Viewer zeichnet nur aufgelöste Beziehungen. Er zeigt eine Anzahl unsicherer (`unresolved` oder `ambiguous`) Beziehungen an, zeichnet sie jedoch nicht oder deren Kandidatenlinks.

Um einzelne unsichere Kanten zu sehen, führen Sie `parse` mit `--edges` aus:

```bash
codeknit parse ./src --output-mode inline --edges
```

Fügen Sie `--format json` für die JSON-Ausgabe hinzu. Informationen dazu, wie Kanten und Kandidaten geschrieben werden, finden Sie im [Referenzhandbuch zum Ausgabeformat](../../reference/output-format/).

## Struktur analysieren und Probleme finden

`graph analyze` führt 22 Graphalgorithmen aus und schreibt einen LLM-lesbaren `.skt`-Bericht.

```bash
codeknit graph analyze ./myproject
```

Der Bericht wird nach `./skeleton/graph_analysis.skt` geschrieben. Lesen Sie ihn direkt oder geben Sie ihn zusammen mit dem relevanten Quellcode an ein LLM weiter.

### Den Bericht anpassen

Erhöhen Sie einen Schwellenwert, um weniger Ergebnisse zu sehen, oder senken Sie ihn, um mehr zu sehen:

```bash
codeknit graph analyze ./myproject -o analysis.skt --fan-threshold 15
codeknit graph graph analyze ./myproject --top-n 50
codeknit graph analyze ./src --collect-test
```

`--top-n 0` entfernt die Obergrenze für gerankte Abschnitte.

| Flag | Standardwert | Zweck |
| --- | --- | --- |
| `-o`, `--output` | `./skeleton/graph_analysis.skt` | Pfad zur `.skt`-Ausgabedatei |
| `--collect-test` | `false` | Testdateien einschließen |
| `--workers` | `0` (`NumCPU`) | Maximale Anzahl gleichzeitiger Parsing-Goroutinen |
| `--verbose` | `false` | Fortschrittsinformationen ausgeben |
| `--fan-threshold` | `10` | Mindest-Fan-in oder Fan-out, um ein Hub-Symbol zu kennzeichnen |
| `--god-threshold` | `15` | Mindestanzahl von Contains-Kanten, um eine god class oder Funktion zu kennzeichnen |
| `--max-inheritance-depth` | `5` | Kennzeichnet Vererbungsketten, die tiefer als dieser Wert sind |
| `--top-n` | `30` | Obergrenze für gerankte Abschnitte; `0` bedeutet keine Grenze |
| `--betweenness-threshold` | `0.001` | Mindest-Betweenness-Centrality für die Berichterstattung |
| `--propagation-cutoff` | `0.05` | Mindest-heuristischer Score für die Fortsetzung der Änderungsausbreitung |

### Wissen, welche Antwort welche Frage beantwortet

| Algorithmus | Dient zum Finden von |
| --- | --- |
| Zyklische Abhängigkeiten (Tarjans SCC) | Gruppen von Symbolen, die in einer Schleife voneinander abhängen |
| Hub-Erkennung | Symbole mit hoher Fan-in- oder Fan-out-Kopplung |
| Waisen-Erkennung | Kandidaten für dead code |
| Erkennung von god classes/Funktionen | Symbole, die eine übermäßige Anzahl von Kindern enthalten |
| Instabilitätsmetrik | Symbole und Pakete, die fragil oder stabil sind, unter Verwendung von Ce/(Ca+Ce) |
| Tiefe Vererbungsketten | Hierarchien, die tiefer als `--max-inheritance-depth` sind |
| Betweenness-Centrality | Engpass-Symbole, durch die viele Abhängigkeitspfade verlaufen |
| Artikulationspunkte | Einzelne Fehlerpunkte, deren Entfernung den Graphen aufteilt |
| PageRank | Symbole, die rekursiv wichtig sind |
| Transitives Fan-in | Auswirkungsradius: alles, was direkt oder indirekt von einem Symbol abhängt |
| Änderungsausbreitungssimulation | Wie weit sich eine Änderung an einem Symbol wahrscheinlich ausbreitet |
| Zirkuläre Paketabhängigkeiten | Paketabhängigkeitsschleifen auf Paketebene |
| Schichtverletzungserkennung | Abhängigkeiten, die Schichten in die falsche Richtung durchqueren |
| Erreichbarkeit von Einstiegspunkten | Code, den kein Einstiegspunkt erreicht |
| Schwach verbundene Komponenten | Isolierte Codecluster |
| Abhängigkeitsgewicht | Wie stark Pakete gekoppelt sind |
| Abstand von der Main Sequence | Pakete, deren Abstraktheit und Instabilität aus dem Gleichgewicht sind |
| Erkennung von Schrotflinten-Chirurgie | Änderungen, die viele Stellen in einem Paket betreffen würden |
| Erkennung von Feature Envy | Symbole, die ein anderes Paket mehr nutzen als ihr eigenes |
| Verstöße gegen stabile Abhängigkeiten | Stabile Pakete, die von weniger stabilen abhängen |
| Verstöße gegen Interface-Segregation | Schnittstellen, deren Mitglieder meist von ihren Nutzern ungenutzt bleiben |
| Containment-Tiefe | Tief verschachtelte Symbole |

### Ergebnisse richtig lesen

Diese Regeln beeinflussen, wie sehr Sie einem Ergebnis vertrauen können:

- **Unsichere Beziehungen werden ausgeschlossen.** Abhängigkeitsmetriken verwenden nur aufgelöste Kanten. Der Bericht zählt die `unresolved` und `ambiguous` Beziehungen, die ausgelassen wurden.
- **Datei- und Include-Metadaten werden ignoriert.** `meta/file`- und `meta/include`-Symbole beeinflussen keine Abhängigkeitsrankings oder Dead-Code-Ergebnisse.
- **Fan-in und Fan-out zählen unterschiedliche Nachbarn.** Containment zählt nicht als Abhängigkeit.
- **Vererbungstiefe umfasst alle Basen und Schnittstellen.** Zyklische Hierarchien wird keine Tiefe zugewiesen.
- **Erreichbarkeit beginnt bei `main`-, `Main`- oder `init`-Callable** und folgt Abhängigkeitskanten. Ein erreichbares Mitglied macht seinen Container relevant, aber nicht seine Geschwistermitglieder. Wenn der Scan keinen erkannten Einstiegspunkt findet, werden Erreichbarkeitsergebnisse ausgelassen.
- **Callbacks sind Referenzen, keine Aufrufe.** Ein Callback-Argument oder zurückgegebener Callable beweist nicht, dass der Callback ausgeführt wird.
- **Scores sind Heuristiken.** Die Änderungsausbreitung verwendet den stärksten Abhängigkeitspfad. Ihre Scores und die Schichtregeln sind keine gemessenen Wahrscheinlichkeiten oder Compiler-Garantien.

### Sicher auf Ergebnisse reagieren

Behandeln Sie jedes Ergebnis als Kandidaten für eine Überprüfung. Bevor Sie Code ändern, öffnen Sie die Symbole, die der Bericht nennt, und prüfen Sie auf:

- Externe Aufrufer außerhalb des gescannten Verzeichnisses.
- Dynamische Dispatch.
- Abhängigkeiten, die Sie nicht gescannt haben.

Die Analyse ist syntaxbasiert, daher kann jeder dieser Punkte die Schlussfolgerung ändern. Informationen zu den Auflösungsregeln und -grenzen finden Sie unter [Beziehungsauflösung im Referenzhandbuch zur Sprachunterstützung](../../reference/language-support/#relationship-resolution).

## Geschichtebewusste Hotspots finden

`graph hotspots` ordnet Dateien, die sich oft ändern und strukturell wichtige Positionen einnehmen. Es berichtet auch über temporale Kopplung: Dateien, die sich wiederholt in denselben Commits ändern. Führen Sie es in einem Verzeichnis innerhalb eines Git-Repositorys aus.

```bash
codeknit graph hotspots ./myproject
```

Der Bericht wird nach `./skeleton/hotspots.skt` geschrieben.

Er kombiniert zwei Gruppen von Metriken:

- **Geschichte:** Commit-Häufigkeit, hinzugefügte und gelöschte Zeilenänderungen (Churn) und Aktualität.
- **Struktur:** Datei-Ebene PageRank, transitives Fan-in und Betweenness.

Der Bericht enthält ein Konfidenzniveau basierend auf der Anzahl der analysierten relevanten Commits. Behandeln Sie Rankings aus einem Bericht mit geringer Konfidenz als vorläufig.

### Das Geschichtsfenster und Filter ändern

```bash
# Zwei Jahre analysieren und JSON schreiben
codeknit graph hotspots ./myproject --since 2y --format json -o hotspots.json

# Größere Commits zulassen und stärkere temporale Kopplung erfordern
codeknit graph hotspots . --max-files-per-commit 100 --min-cochanges 5
```

Merge-Commits werden ausgeschlossen, es sei denn, Sie übergeben `--include-merges`. Commits, die mehr Dateien ändern als `--max-files-per-commit`, werden als Massen-Commits übersprungen, sodass generierte oder mechanische Änderungen das Ranking nicht verzerren.

| Flag | Standardwert | Zweck |
| --- | --- | --- |
| `-o`, `--output` | `./skeleton/hotspots.skt` | Pfad zur Ausgabedatei |
| `--format` | `skt` | Ausgabeformat: `skt` oder `json` |
| `--since` | `12mo` | Geschichtsfenster, z. B. `180d`, `12mo` oder `2y` |
| `--max-commits` | `2000` | Maximale Anzahl zu inspizierender Commits |
| `--max-files-per-commit` | `50` | Commits ausschließen, die mehr Dateien ändern als dieser Wert |
| `--min-cochanges` | `3` | Mindestanzahl gemeinsamer Commits, um temporale Kopplung zu berichten |
| `--top-n` | `30` | Maximale Ergebnisse pro Abschnitt |
| `--include-merges` | `false` | Merge-Commits in Geschichtskennzahlen einschließen |
| `--collect-test` | `false` | Testdateien einschließen |
| `--workers` | `0` (`NumCPU`) | Maximale Anzahl gleichzeitiger Parsing-Goroutinen |
| `--verbose` | `false` | Fortschrittsinformationen ausgeben |

### Die JSON-Ausgabe lesen

Mit `--format json` hat jeder Hotspot `score`, `history_score`, `structure_score`, `recency_score`, `pagerank`, `betweenness`, `commits`, `churn` und `transitive_fan_in`. Jeder Eintrag zur temporalen Kopplung hat `left`, `right`, `strength` und `co_changes`. Die oberste Ebene zeichnet auch `commits_visited`, `commits_analyzed`, `skipped_merges` und `skipped_bulk_commits` auf. Verwenden Sie diese Zählungen, um zu prüfen, auf wie viel Geschichte das Ranking basiert.

## Steuern, welche Quellen gescannt werden

Alle fünf quellverarbeitenden Befehle (`parse`, `fingerprint`, `graph show`, `graph analyze` und `graph hotspots`) akzeptieren dieselben Auswahl-Flags:

| Flag | Standardwert | Zweck |
| --- | --- | --- |
| `--include GLOB` | Alle unterstützten Quellen | Pfade einschließen, die dem Muster entsprechen; wiederholbar |
| `--exclude GLOB` | Keine | Pfade oder Verzeichnisse ausschließen, die dem Muster entsprechen; wiederholbar. Ausschlüsse haben Vorrang vor Einschlüssen |
| `--max-file-size BYTES` | `10485760` (10 MiB) | Größere Dateien überspringen; `0` deaktiviert das Limit |
| `--strict` | `false` | Fehlschlagen, bevor die Ausgabe geschrieben wird, wenn der Scan oder das Parsen unvollständig ist |

Muster sind gitignore-artige Globs relativ zum Eingabeverzeichnis. Setzen Sie sie in Anführungszeichen, damit Ihre Shell sie nicht expandiert. Einschlüsse überschreiben nicht `.gitignore`-Regeln oder Testfilterung.

```bash
codeknit graph analyze ./src --include '**/*.go' --exclude '**/generated/' --strict
```

Symbolische Links und nicht reguläre Dateien werden übersprungen und gemeldet. Scan- und Parse-Probleme erscheinen auf stderr. Ohne `--strict` behält codeknit nutzbare Teilergebnisse bei, daher sollten Sie stderr prüfen, bevor Sie einem Bericht vertrauen. Mit `--strict` stoppt eine unvollständige Analyse, bevor eine Ausgabe geschrieben wird. Ein Durchlauf, bei dem jede Quelldatei fehlschlägt, schlägt immer fehl.

Für die genauen Garantien siehe die [Referenz zur Analysezuverlässigkeit](../../reference/analysis-reliability/).

## Verwandte Seiten

- [Referenz zu CLI-Flags](../../reference/cli-flags/): Jedes Flag für jeden Befehl.
- [Referenz zum Ausgabeformat](../../reference/output-format/): `.skt`- und JSON-Syntax.
- [Referenz zur Sprachunterstützung](../../reference/language-support/): Extraktions- und Auflösungsgrenzen.
- [Anleitung zum Parse-Befehl](../parse-command/): Symbole und Kanten für einzelne Dateien extrahieren.