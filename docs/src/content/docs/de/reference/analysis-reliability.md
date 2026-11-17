---
title: Analysezuverlässigkeit
description: Quellauswahl, unvollständige Analysen, Ausgabeveröffentlichung und Garantien zur Duplikaterkennung.
---

<!-- Maintained in docs/reference/analysis-reliability.md; copied verbatim by docs/gen. -->

Diese Verhaltensweisen beschreiben die nicht veröffentlichten Go/CLI-Änderungen nach Version 0.5.0.

## Quellauswahl und Diagnosen

Alle fünf Quellverarbeitungsbefehle (`parse`, `fingerprint`, `graph show`, `graph analyze` und `graph hotspots`) akzeptieren:

| Flag                    | Standardwert          | Bedeutung                                                             |
| ----------------------- | --------------------- | --------------------------------------------------------------------- |
| `--include GLOB`        | Alle unterstützten Quellen | Einschließen passender Quellpfade; wiederholen, um mehrere Muster auszuwählen. |
| `--exclude GLOB`        | Keine                 | Ausschließen passender Pfade oder Verzeichnisse; wiederholen für mehrere Muster. |
| `--max-file-size BYTES` | `10485760` (10 MiB)   | Überspringen größerer Quelldateien; `0` deaktiviert diese Grenze.    |
| `--strict`              | `false`               | Abbrechen vor der Ausgabe von Ergebnissen, wenn Scan- oder Parse-Diagnosen auftreten. |

Muster verwenden schrägstrichgetrennte, gitignore-artige Globs relativ zum ausgewählten Eingabeverzeichnis. Für eine einzelne Datei als Eingabe passen die Muster auf deren Basisnamen. Muster in Anführungszeichen setzen, damit die Shell sie nicht expandiert. Einschlüsse bilden eine Vereinigung; Ausschlüsse haben Vorrang. Einschlüsse überschreiben nicht `.gitignore`, Testfilterung oder integrierte Ausschlüsse. Negationsmuster (`!pattern`) werden in diesen Flags nicht akzeptiert; stattdessen separate Einschluss- und Ausschlusssteuerungen verwenden.

```bash
codeknit parse ./src --include '**/*.go' --exclude '**/generated/' --strict
codeknit fingerprint ./src --max-file-size 2097152 --include '**/*.ts'
codeknit parse ./src/app.go --output-mode inline --format json
```

Verzeichnisscans beachten übergeordnete und verschachtelte `.gitignore`-Dateien relativ zu ihrem Projektstandort, unabhängig vom Arbeitsverzeichnis des Prozesses. Das nächstgelegene `.git`-Verzeichnis oder der Worktree-Marker legt die Repository-Grenze fest. Außerhalb eines Repositorys gelten weiterhin übergeordnete Ignore-Dateien. Ein ausgewähltes Unterverzeichnis unterhalb eines ignorierten Verzeichnisses bleibt ignoriert. Die explizite Auswahl einer regulären Quelldatei umgeht Ignorieren und Testfilterung, wendet aber weiterhin Einschluss/Ausschluss- und Größengrenzen an.

Quellsymlinks, Verzeichnissymlinks und nicht-reguläre Quellen werden übersprungen und gemeldet. Lesevorgänge sind auf das Scan-Stammverzeichnis beschränkt, und die Größengrenze wird erneut beim Lesen durchgesetzt, sodass eine Datei, die nach der Entdeckung wächst, diese nicht umgehen kann. Integrierte Verzeichnisausschlüsse umfassen `.git`, `.hg`, `.svn`, `node_modules`, Python-Caches und gängige Python-Virtual-Umgebungen. Testdateien bleiben ausgeschlossen, es sei denn, `--collect-test` ist gesetzt.

Nicht lesbare Verzeichnisse/Dateien, fehlgeschlagene Ignore-Datei-Lesevorgänge, übersprungene Symlinks, zu große Quellen und Syntaxwarnungen machen eine Analyse unvollständig. Standardmäßig werden nutzbare Teilergebnisse beibehalten und Diagnosen erscheinen auf stderr. Parse-JSON enthält Pfad/Grund-Datensätze in `errors`; SKT enthält `[errors]` inline oder in `warnings.skt`. Graph/Fingerprint-Befehle melden Diagnosen ebenfalls auf stderr. Absichtlich ignorierte oder gefilterte Dateien sind keine Warnungen. `--strict` lehnt unvollständige Ergebnisse vor der Ausgabeveröffentlichung ab. Eine Eingabe ohne ausgewählte unterstützte Quellen oder ein Lauf, bei dem jede Quellenlesung fehlschlägt, beendet sich immer erfolglos.

Die TUI verwendet dieselbe Symlink-Richtlinie und den Standardwert von 10 MiB. Ihre Formulare bieten die neuen Auswahl- und Strict-Flags noch nicht an.

## Ausgabenamen und Veröffentlichung

Baumausgabe hängt `.skt` an den vollständigen Quellnamen an: `src/foo.c` wird zu `src/foo.c.skt`, während `src/foo.h` zu `src/foo.h.skt` wird. Aufgeteilte Ausgabe verwendet `src/foo.c_part1.skt`, `src/foo.c_part2.skt` und so weiter. Wörterbuch- und Diagnosedateien behalten die reservierten Namen `dict.skt` und `warnings.skt`. Kollidierende Ausgabepfade, einschließlich reiner Groß-/Kleinschreibungskollisionen, schlagen sichtbar fehl.

Baum-SKT erfordert `--max-lines` von mindestens 3. Jeder Symbolabschnitt wiederholt seine `[symbols]`- und `## source-path`-Header. Kanten-only-Abschnitte enthalten `[edges]`. Jeder Quellenabschnitt respektiert die Grenze; Wörterbücher und Diagnosedateien sind separate Hilfsausgaben. Flaches SKT behält Quellblöcke intakt, sodass seine Zeilengrenze ein Aufteilungsziel und keine harte Grenze für einen großen Quellblock ist. JSON ist ein einzelnes Dokument und wird nicht nach Zeilenanzahl aufgeteilt.

Verzeichniserstellung erfolgt in einem benachbarten Staging-Verzeichnis. Ohne `--clean` führt vorhandene generierte Ausgabe für das ausgewählte Format zu einem Fehler. Mit `--clean` werden veraltete `.skt`-Dateien (SKT-Modus) oder `codeknit.json` (JSON-Modus) erst nach erfolgreicher Erstellung ersetzt. Nicht verwandte Einträge bleiben erhalten; eine Kollision mit einem führt zu einem Fehler. Das Quellverzeichnis oder dessen übergeordnetes Verzeichnis nicht als Ausgabeverzeichnis verwenden.

Die Veröffentlichung hält eine verzeichnisspezifische Sperre, sichert die vorherige Ausgabe und stellt sie wieder her, falls der Ersatz fehlschlägt. Wenn ein Prozess zwischen Verzeichnisumbenennungen beendet wird, stellt der nächste Aufruf die Sicherung wieder her, bevor etwas generiert wird. Die benachbarte `.NAME.codeknit.lock`-Datei bleibt absichtlich erhalten, um zukünftige Veröffentlichungen zu koordinieren; eine aktive Sperre oder Wiederherstellungssicherung nicht entfernen. Es gibt ein kurzes Intervall zwischen den beiden Verzeichnisumbenennungen, sodass Leser, die ununterbrochenen Zugriff benötigen, sich mit der Erstellung koordinieren sollten. Dies bietet Wiederherstellung bei Prozessfehlern, keine Garantie gegen Speicherfehler oder Stromausfall.

Einzelne Berichte werden in ein temporäres benachbartes Verzeichnis geschrieben und synchronisiert, bevor das Ziel ersetzt wird. Bestehende Dateien werden während der Erstellung nicht trunkiert. Inline-Stdout kann nicht zurückgesetzt werden, wenn dessen Verbraucher fehlschlägt.

## Garantien zur Duplikaterkennung

Exakte Übereinstimmung bedeutet identische vollständige normalisierte Token-Ströme in derselben Symbolkategorie unter Deklarationen, die für die Fingerprinting-Berechnung infrage kommen. Es bedeutet nicht identischen Quelltext oder nachgewiesene semantische Äquivalenz. Exakte Kandidatengruppen verwenden die Token-Ströme selbst, anstatt auf fuzzy-hash-Gleichheit zu basieren.

Unter 500 infrage kommenden Symbolen vergleicht die strukturelle Erkennung alle Paare derselben Kategorie. Bei 500 oder mehr werden alle exakten Token-Paare unabhängig von der Kandidatenabfrage beibehalten, und andere Paare verwenden die ungefähren top-50 strukturellen Nachbarn pro Symbol. Die Abfrage von Beinahe-Duplikaten ist daher bei großen Eingaben nicht erschöpfend. Die Abfrage durchsucht weiterhin alle Vektoren; sie macht den Gesamtalgorithmus nicht subquadratisch.

Die Filterung des Ähnlichkeitsbereichs gilt weiterhin. Der Standardbereich liegt bei 65–95%, was 100%-Übereinstimmungen ausschließt. Um exakte Übereinstimmungen einzuschließen, verwenden Sie:

```bash
codeknit fingerprint ./src --min-similarity 100 --max-similarity 100
```

Optionales `--rerank` fügt semantische Kandidaten hinzu, wendet Cosinus-Übereinstimmung an und filtert nach dem kombinierten Score; es kann einen strukturell exakten Kandidaten entfernen. Einbettungstext stammt aus demselben begrenzten Quell-Snapshot, der für die Extraktion verwendet wurde. Stabile Eingaben und feste Scores erzeugen deterministische Reihenfolgen, einschließlich Gleichständen; Modellantworten selbst sind externe Eingaben.

`internal/emitter/calibration_test.go` evaluiert eine kleine gelabelte Quell-Fixture bei einem Schwellenwert von 70% und berichtet True/False-Positives und -Negatives, Präzision und Recall getrennt von Ähnlichkeitsscores. Die Fixture enthält ein exaktes Paar, zwei Beinahe-Klon-Paare und drei nicht verwandte Paare, die sowohl unterhalb als auch oberhalb der 500-Symbol-Grenze getestet werden. Dies ist eine Regressionprüfung, kein repräsentativer Schätzwert für die Genauigkeit in der realen Welt. Größere mehrsprachige Kalibrierungs- und Leistungsbudgets bleiben Roadmap-Arbeit.