---
title: Ausgabemodi
description: Wählen Sie den passenden Ausgabemodus für die Größe Ihres Projekts und Ihren Workflow.
---

codeknit trennt zwei Entscheidungen: wohin die Ausgabe geht (**Ausgabemodus**) und wie sie kodiert wird (**Ausgabeformat**). Diese Seite erklärt, warum jede Option existiert, welche Kosten sie verursacht und wann sie passt. Für die `.skt`-Grammatik siehe die [Referenz zum Ausgabeformat](../../reference/output-format/). Für Befehle siehe den [Leitfaden zum Parse-Befehl](../parse-command/).

## Zwei unabhängige Entscheidungen

Das Flag `--output-mode` wählt einen von drei Modi aus: `directory-flat`, `directory-tree` oder `inline`. Das Flag `--format` wählt `skt` (Standard) oder `json`. Jeder Modus kann mit jedem Format kombiniert werden.

Die beiden Entscheidungen sind getrennt, weil sie unterschiedliche Fragen beantworten. Der Modus entscheidet, wie die Ausgabe aufgeteilt und auf der Festplatte platziert wird, was für eine Person oder ein LLM beim Lesen wichtig ist. Das Format entscheidet, wer sie konsumiert: ein LLM oder menschlicher Leser, oder ein Skript.

## directory-flat

`directory-flat` ist der Standard. Es schreibt nummerierte Chunk-Dateien, `map_001.skt`, `map_002.skt` und so weiter, in `./skeleton/`, es sei denn, Sie übergeben ein anderes Ausgabeverzeichnis.

**Warum es existiert.** Ein großes Codebase erzeugt mehr Struktur, als in ein Kontextfenster passt. Die Aufteilung in Chunks ermöglicht es einem Leser, nur die benötigten Dateien zu öffnen, anstatt alles zu laden.

**Wie die Aufteilung funktioniert.** codeknit hält den Block jeder Quelldatei intakt und beginnt einen neuen Chunk zwischen Blöcken, sobald ein Chunk etwa `--max-lines` (Standard 500) erreicht. Die Grenze ist ein Ziel, keine harte Vorgabe: Eine Quelldatei mit einem sehr großen Block kann einen Chunk erzeugen, der länger als 500 Zeilen ist. Die Blöcke intakt zu halten bedeutet, dass ein Symbol nie von seinen Geschwistern in derselben Datei getrennt wird.

**Kompromiss.** Chunks sind nicht nach Quellpfad organisiert. Um die Struktur einer Datei zu finden, suchen Sie in den Chunks nach ihrem `## source-path`-Header. Wenn Sie häufig bestimmte Dateien nachschlagen, passt `directory-tree` besser.

**Minifizierung.** Mit `--minify` wird das Wörterbuch in `dict.skt` neben den Chunks geschrieben. Siehe [Minifizierung](#minifizierung).

## directory-tree

`directory-tree` spiegelt Ihren Quellbaum unter dem Ausgabeverzeichnis wider. Jede Quelldatei erhält eine `.skt`-Datei am entsprechenden Pfad.

**Benennung.** codeknit hängt `.skt` an den vollständigen Quelldateinamen an, anstatt die Erweiterung zu ersetzen. `src/foo.c` wird zu `src/foo.c.skt`, und `src/foo.h` wird zu `src/foo.h.skt`. Ohne diese Regel würden beide auf `foo.skt` abgebildet und kollidieren. Wenn die Ausgabe einer Datei aufgeteilt wird, werden die Teile `foo.c_part1.skt`, `foo.c_part2.skt` und so weiter benannt. Konfligierende Ausgabepfade, einschließlich Kollisionsfällen bei Groß-/Kleinschreibung, schlagen sichtbar fehl, anstatt sich gegenseitig zu überschreiben.

**Aufteilung unterscheidet sich vom Flat-Modus.** Im Tree-Modus ist `--max-lines` eine harte Grenze für jeden Quell-Chunk, mit einem Minimum von 3. Jeder Symbol-Chunk wiederholt die `[symbols]`- und `## source-path`-Header, sodass er allein gelesen werden kann. Edge-only-Chunks enthalten `[edges]`. Die Grenze gilt für Quell-Chunks, nicht für Hilfsdateien wie `dict.skt` und `warnings.skt`.

**Kompromiss.** Tree-Ausgabe ist einfach neben dem Quellcode zu navigieren, erzeugt aber viele kleine Dateien. Ein großes Projekt kann Tausende davon hervorbringen, und ein Leser, der einen breiten Überblick möchte, muss viele Dateien öffnen.

## inline

`inline` schreibt alles nach stdout. Es erstellt kein Ausgabeverzeichnis und keinen Staging-Bereich.

**Warum es existiert.** Es eignet sich für zwei Fälle: schnelle Inspektion einer einzelnen Datei und Weiterleitung der Ausgabe an ein anderes Tool. Es ist der einzige Modus, in dem nichts auf die Festplatte geschrieben wird.

**Kompromiss.** Nichts wird aufgeteilt. Das gesamte Ergebnis landet auf einmal im Terminal oder im Kontext des Konsumenten, sodass es einen Leser bei mehr als einer Handvoll Dateien überfordert. Betrachten Sie es als geeignet für eine einzelne Datei oder ein sehr kleines Projekt, etwa weniger als fünf Dateien. Da stdout nicht zurückgerollt werden kann, kann ein fehlgeschlagener Konsument partielle Ausgabe sehen. Verzeichnismodi haben dieses Problem nicht.

## Ausgabeformate

| Format | Am besten für | Ausgabe |
| --- | --- | --- |
| `skt` | LLM-Kontext und menschliche Inspektion | `.skt`-Dateien oder stdout |
| `json` | Skripte und strukturierte Integration | `codeknit.json` in Verzeichnismodi oder stdout im Inline-Modus |

**skt** ist kompakt und zeilenorientiert. Jedes Symbol befindet sich auf einer physischen Zeile, wobei Zeilenumbrüche in Signaturen escaped sind, sodass ein Leser einen Teil der Ausgabe laden kann, ohne deren Struktur zu brechen. Deshalb funktioniert die Aufteilung in Chunks für `.skt`.

**json** existiert für Konsumenten, die typisierte Felder statt Text zum Lesen benötigen. Es ist ein einzelnes Dokument mit `files`, `symbols` und, wenn Kanten enthalten sind, `edges`. Da es ein Dokument ist, wird es nicht nach Zeilenanzahl aufgeteilt. In Verzeichnismodi hat `--max-lines` daher keine Wirkung auf JSON, und ein großes Projekt erzeugt eine große `codeknit.json`.

Ein kompaktes Beispiel:

```json
{
  "files": ["app.go"],
  "symbols": [
    {"id": "app.go::User", "short_id": "S1", "name": "User", "file": "app.go",
     "category": "type", "kind": "struct", "signature": "type User struct", "span": [3, 3]},
    {"id": "app.go::Save", "short_id": "S2", "name": "Save", "file": "app.go",
     "category": "callable", "kind": "function", "signature": "Save(u: S1)", "span": [5, 5]}
  ],
  "edges": [
    {"from": "app.go::Save", "from_short": "S2", "to": "app.go::User",
     "to_short": "S1", "kind": "references"}
  ]
}
```

## Auswahl eines Modus

| Modus | Am besten für | Ausgabe |
| --- | --- | --- |
| `directory-flat` | Die meisten Projekte (Standard) | `./skeleton/map_001.skt`, `map_002.skt`, ... |
| `directory-tree` | Navigation neben dem Quellcode | `./skeleton/<gespiegelter Pfad>.skt` |
| `inline` | Einzelne Datei, Weiterleitung | stdout |

Faustregeln:

- **Im Zweifel** verwenden Sie `directory-flat`, den Standard.
- **Eine einzelne Datei**: `inline` ist in Ordnung.
- **Mehr als ein paar Dateien**: Verwenden Sie `directory-flat` oder `directory-tree`.
- **Große Codebases**: Fügen Sie `--minify` hinzu.
- **Neustart im selben Verzeichnis**: Fügen Sie `--clean` hinzu.

## Staging, Neustarts und `--clean`

Verzeichnismodi schreiben nicht direkt in das Ausgabeverzeichnis. codeknit generiert die Ausgabe in einem benachbarten Staging-Verzeichnis und veröffentlicht sie erst nach erfolgreicher Generierung. Ein fehlgeschlagener Lauf hinterlässt daher Ihre vorherige Ausgabe an ihrem Platz.

Standardmäßig weigert sich codeknit zu schreiben, wenn generierte Ausgabe für das ausgewählte Format bereits existiert. Der Grund ist, dass Symbol-IDs (`S1`, `S2`, ...) sich ändern können, wenn sich die Extraktion ändert. Das Mischen von Chunks aus verschiedenen Läufen würde Kanten ergeben, die auf falsche Symbole verweisen. `--clean` stimmt dem Ersatz zu: Nach erfolgreicher Generierung werden veraltete `.skt`-Dateien (im SKT-Modus) oder `codeknit.json` (im JSON-Modus) ersetzt. Unzusammenhängende Einträge im Verzeichnis bleiben erhalten. Wenn einer von ihnen mit einem generierten Pfad kollidiert, schlägt der Lauf fehl, anstatt ihn zu überschreiben.

Die Veröffentlichung nimmt eine verzeichnisspezifische Sperre, sichert die vorherige Ausgabe und stellt sie wieder her, wenn der Ersatz fehlschlägt. Wenn der Prozess zwischen den beiden Verzeichnisumbenennungen unterbrochen wird, stellt der nächste Lauf die Sicherung wieder her, bevor etwas generiert wird. Dies stellt die Wiederherstellung nach Prozessfehlern sicher. Es macht die Verzeichnisaktualisierung nicht augenblicklich: Es gibt ein kurzes Intervall zwischen den Umbenennungen, sodass ein Leser, der ununterbrochenen Zugriff benötigt, sich mit der Generierung koordinieren sollte. Es schützt auch nicht vor Speicherfehlern oder Stromausfällen.

Wählen Sie kein Quellverzeichnis oder ein übergeordnetes Verzeichnis davon als Ausgabeverzeichnis. Für die vollständigen Veröffentlichunggarantien siehe die [Referenz zur Analysezuverlässigkeit](../../reference/analysis-reliability/).

## Minifizierung

`--minify` wendet eine wörterbuchbasierte Komprimierung an. codeknit ersetzt wiederholte Tokens wie `callable/method` oder `exported` durch kurze Codes (`d0`, `d1`, ...). Die Zuordnung wird in einem `[dict]`-Abschnitt gespeichert, der in Verzeichnismodi die Datei `dict.skt` ist.

Die Einsparung ergibt sich aus Wiederholungen. Die meisten Symbolzeilen teilen dieselben Kategorien, Arten und Eigenschaften, sodass ein großes Codebase diese Zeichenfolgen tausende Male wiederholt. Der Preis ist eine Indirektion: Ein Leser muss Codes mit dem Wörterbuch expandieren, und ein einzelner Chunk ist nicht mehr eigenständig. Das Wörterbuch ist eine separate Hilfsausgabe, sodass es nicht gegen `--max-lines` zählt.

Minifizierte Ausgabe kodiert die Beziehungskategorie, behält aber Auflösungsmarker lesbar, sodass `unresolved` und `ambiguous` Kanten unterscheidbar bleiben. Bei kleinen Projekten sind die Einsparungen bescheiden und der zusätzliche Schritt lohnt sich selten. Bei großen Projekten können die Einsparungen an Tokens erheblich sein.

## Verwandte Seiten

- [Leitfaden zum Parse-Befehl](../parse-command/): die Befehle für jeden Modus.
- [Referenz zum Ausgabeformat](../../reference/output-format/): die `.skt`- und JSON-Grammatik.
- [Referenz zu CLI-Flags](../../reference/cli-flags/): jedes Flag und Standardwert.
- [Referenz zur Analysezuverlässigkeit](../../reference/analysis-reliability/): Benennung der Ausgabe und Veröffentlichunggarantien.