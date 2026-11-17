---
title: Parse-Befehl
description: Extrahieren Sie strukturelle Informationen aus Quellcode in .skt-Dateien oder JSON.
---

`codeknit parse` liest Quelldateien und schreibt deren Symbole (und optional Beziehungen zwischen ihnen) als `.skt`-Dateien oder JSON. Dieser Leitfaden behandelt die häufigsten Anwendungsfälle. Für die vollständige Liste der von allen Befehlen gemeinsam genutzten Flags siehe die [CLI-Flags-Referenz](../../reference/cli-flags/). Für die Ausgabesyntax siehe die [Referenz zum Ausgabeformat](../../reference/output-format/).

## Ein Projekt zum ersten Mal parsen

```bash
codeknit parse ./src
```

codeknit schreibt chunkweise `.skt`-Dateien nach `./skeleton/`: `map_001.skt`, `map_002.skt` und so weiter. Öffnen Sie die Chunks selektiv, anstatt alles auf einmal zu laden.

Um an einen anderen Ort zu schreiben, übergeben Sie das Ausgabeverzeichnis als zweites Argument:

```bash
codeknit parse ./src ./output
```

Verwenden Sie nicht das Quellverzeichnis oder eines seiner übergeordneten Verzeichnisse als Ausgabeverzeichnis.

## Einen Ausgabemodus wählen

Wählen Sie einen Modus mit `--output-mode`.

| Modus             | Verwenden Sie ihn, wenn                                    | Ergebnis                                                                         |
| ---------------- | ---------------------------------------------- | ------------------------------------------------------------------------------ |
| `directory-flat` | Die meisten Projekte parsen (Standard)                | `map_001.skt`, `map_002.skt`, ...                                              |
| `directory-tree` | Sie die Struktur einer einzelnen Datei nachschlagen möchten       | Eine `.skt`-Datei pro Quelldatei, die den Quellbaum widerspiegelt                          |
| `inline`         | Eine einzelne Datei parsen oder an ein anderes Tool weiterleiten | Ausgabe auf stdout; es wird kein Ausgabeverzeichnis verwendet                                  |

Im Baummodus behalten die Dateinamen die Quelldateierweiterung: `src/foo.c` wird zu `src/foo.c.skt`, und `src/foo.h` wird zu `src/foo.h.skt`. Große Dateien werden in `foo.c_part1.skt`, `foo.c_part2.skt` usw. aufgeteilt.

```bash
codeknit parse ./src --output-mode directory-tree
```

Vermeiden Sie `inline` für ein ganzes Projekt. Es schreibt alles auf stdout und kann Ihr Terminal oder ein LLM-Kontextfenster überfluten.

Weitere Informationen zur Auswahl eines Modus finden Sie unter [Ausgabemodi](../output-modes/).

## Erneutes Parsen desselben Projekts

codeknit weigert sich, in ein Verzeichnis zu schreiben, das bereits generierte Ausgabe für das ausgewählte Format enthält. Dies verhindert, dass veraltete und frische Chunks vermischt werden. Fügen Sie `--clean` hinzu, um die vorherige Ausgabe zu ersetzen:

```bash
codeknit parse ./src --clean
```

`--clean` löscht nichts im Voraus. codeknit generiert die neue Ausgabe in einem Staging-Verzeichnis und ersetzt die alten generierten Dateien erst, nachdem die Generierung erfolgreich war. Wenn die Generierung fehlschlägt oder der Prozess unterbrochen wird, bleibt die vorherige Ausgabe erhalten oder wird beim nächsten Lauf wiederhergestellt. Dateien im Ausgabeverzeichnis, die codeknit nicht generiert hat, bleiben erhalten. Wenn eine von ihnen mit einem generierten Pfad kollidiert, schlägt der Lauf fehl.

Im JSON-Modus gelten dieselben Regeln für `codeknit.json`.

**Fallstrick: gemischte Chunks.** Symbol-IDs (`S1`, `S2`, ...) und Beziehungsergebnisse können sich zwischen Läufen und Versionen ändern. Nach dem Upgrade auf 0.5.0 generieren Sie die gesamte Ausgabemenge mit `--clean` neu. Kombinieren Sie niemals Chunks aus verschiedenen Läufen.

## Eine einzelne Datei parsen

Richten Sie `parse` auf die Datei und verwenden Sie den `inline`-Modus:

```bash
codeknit parse ./src/main.go --output-mode inline
```

Wenn Sie eine reguläre Quelldatei explizit auswählen, überspringt codeknit die Ignorieren-Datei- und Testdatei-Filterung dafür. Include/Exclude-Muster und die Größenbeschränkung gelten weiterhin.

## JSON für Skripte abrufen

```bash
codeknit parse ./src --output-mode inline --format json --edges
```

Das JSON-Dokument enthält die Felder der obersten Ebene `files`, `symbols`, `edges` (weggelassen, wenn keine vorhanden sind) und `errors` (weggelassen, wenn keine vorhanden sind). Ein kleines Beispiel:

```json
{
  "files": ["app.go"],
  "symbols": [
    {
      "id": "app.go::User",
      "short_id": "S1",
      "name": "User",
      "file": "app.go",
      "category": "type",
      "kind": "struct",
      "signature": "type User struct",
      "span": [3, 3]
    },
    {
      "id": "app.go::Save",
      "short_id": "S2",
      "name": "Save",
      "file": "app.go",
      "category": "callable",
      "kind": "function",
      "signature": "Save(u: S1)",
      "span": [5, 5]
    }
  ],
  "edges": [
    {
      "from": "app.go::Save",
      "from_short": "S2",
      "to": "app.go::User",
      "to_short": "S1",
      "kind": "references"
    }
  ]
}
```

Um das JSON auf die Festplatte zu schreiben, lassen Sie `--output-mode inline` weg. In den Verzeichnismodi geht JSON nach `codeknit.json`. JSON ist ein einzelnes Dokument und wird nicht nach Zeilenanzahl aufgeteilt.

## Beziehungen einbeziehen

Standardmäßig enthält die Ausgabe nur den Abschnitt `[symbols]`, was Tokens spart. Fügen Sie `--edges` hinzu, wenn Sie wissen müssen, was was aufruft, enthält, erbt oder referenziert:

```bash
codeknit parse ./src --edges
```

### Unsichere Beziehungen lesen

Codeknit analysiert Syntax ohne Compiler, sodass einige Beziehungen nicht hergestellt werden können. Jede Kante gibt an, wie sicher sie ist:

```skt
S8 --references--> S5
S10 --references[unresolved]--> string, bool
S11 --calls[ambiguous]--> Helper [candidates=S12, S15]
```

- `--kind-->` ist eine aufgelöste Beziehung.
- `--kind[unresolved]-->` bedeutet, dass codeknit das Ziel nicht herstellen konnte.
- `--kind[ambiguous]-->` bedeutet, dass mehr als ein Ziel möglich bleibt. Kandidaten-IDs werden aufgelistet, wenn verfügbar.

Kandidaten sind mögliche Ziele, keine bestätigten Abhängigkeiten. Bekannte Endpunkte verwenden ShortIDs. Ein nicht aufgelöstes Ziel ohne passendes Symbol kann als Name relativ zur Quelldatei erscheinen.

In JSON tragen unsichere Kanten `resolution` (`"unresolved"` oder `"ambiguous"`) und, wenn verfügbar, `candidates` als vollständige Symbol-IDs. Sie lassen `to_short` weg.

Für C und C++ werden `#include`-Direktiven zu `meta/file`- und `meta/include`-Symbolen, sodass ihre Kanten ShortIDs haben. Dies zeichnet die Direktive so auf, wie sie geschrieben wurde. Es bedeutet nicht, dass codeknit die Header-Datei gefunden hat.

Graphmetriken schließen unsichere Kanten aus und melden, wie viele ausgeschlossen wurden. Der HTML-Graph-Viewer zeigt eine Anzahl unsicherer Beziehungen an und zeichnet sie nicht. Um einzelne unsichere Kanten zu inspizieren, verwenden Sie `--edges` bei `parse`. Siehe [Graph-Befehle](../graph-commands/) für diese Berichte.

Die Kantensyntax finden Sie in der [Referenz zum Ausgabeformat](../../reference/output-format/). Auflösungsregeln und -grenzen finden Sie in der [Referenz zur Sprachunterstützung](../../reference/language-support/#relationship-resolution).

## Ausgabegröße reduzieren

Um wiederholte Tokens zu komprimieren, fügen Sie `--minify` hinzu:

```bash
codeknit parse ./src --minify
```

Dies schreibt ein Wörterbuch nach `dict.skt` (oder einen `[dict]`-Abschnitt inline). Ersetzen Sie Codes wie `d0` und `d1` durch ihre Wörterbuchwerte beim Lesen. Beziehungstypen werden ebenfalls codiert, aber Auflösungslabels wie `[unresolved]` bleiben lesbar.

Um die Chunk-Größe zu steuern, verwenden Sie `--max-lines` (Standard `500`):

- Im Modus `directory-flat` ist es ein Zielwert. Ein Quellblock wird nie aufgeteilt, sodass ein Chunk ihn um einen Block überschreiten kann.
- Im Modus `directory-tree` ist es eine harte Grenze für jeden Quell-Chunk. Das Minimum ist `3`. Jeder Chunk wiederholt seine `[symbols]`- und `## source-path`-Header.

## C++-Header mit der Endung `.h` parsen

Codeknit parst `.h`-Dateien standardmäßig als C. Für C++-Header mit dieser Erweiterung setzen Sie `--header-language cpp`:

```bash
codeknit parse ./src --header-language cpp
```

Das Flag betrifft nur `.h`-Dateien. Dateien mit den Endungen `.hpp` und `.hxx` werden bereits als C++ geparst, und andere Erweiterungen behalten ihre üblichen Parser.

## Steuern, welche Dateien gescannt werden

Alle fünf Quellcode-verarbeitenden Befehle akzeptieren diese Flags. Sie sind vollständig in der [CLI-Flags-Referenz](../../reference/cli-flags/) aufgeführt.

```bash
codeknit parse ./src --include '**/*.go' --exclude '**/generated/' --strict
```

- `--include GLOB` und `--exclude GLOB` sind wiederholbare, gitignore-artige Muster relativ zum Eingabeverzeichnis. Setzen Sie sie in Anführungszeichen, damit Ihre Shell sie nicht expandiert. Ausschlüsse haben Vorrang. Einschlüsse überschreiben nicht `.gitignore`, Testdatei-Filterung oder integrierte Ausschlüsse.
- `--max-file-size BYTES` überspringt größere Dateien. Der Standardwert ist `10485760` (10 MiB). `0` deaktiviert die Grenze.
- `--strict` schlägt fehl, bevor eine Ausgabe geschrieben wird, wenn der Scan oder das Parsen unvollständig ist. Der Standardwert ist `false`.

Codeknit überspringt und meldet Symlinks und nicht reguläre Dateien. Übergeordnete `.gitignore`-Regeln gelten unabhängig von Ihrem Arbeitsverzeichnis.

**Fallstrick: Teilresultate.** Nicht lesbare Pfade, zu große Dateien, übersprungene Symlinks und Syntaxwarnungen machen eine Analyse unvollständig. Ohne `--strict` behält codeknit die nutzbaren Ergebnisse und meldet die Probleme:

- auf stderr,
- im JSON-Feld `errors`, oder
- im SKT-Abschnitt `[errors]` (inline oder in `warnings.skt` in Verzeichnismodi).

Eine Eingabe ohne ausgewählte unterstützte Quellen oder ein Lauf, bei dem jede gelesene Quelle fehlschlägt, beendet sich immer erfolglos. Für genaue Garantien siehe [Analysezuverlässigkeit](../../reference/analysis-reliability/).

## Parse-Flags

Diese Flags sind spezifisch für `parse`. Gemeinsame Flags finden Sie in der [CLI-Flags-Referenz](../../reference/cli-flags/).

| Flag               | Standard          | Beschreibung                                                                                       |
| ------------------ | ---------------- | ------------------------------------------------------------------------------------------------- |
| `--output-mode`    | `directory-flat` | Ausgabemodus: `inline`, `directory-flat` oder `directory-tree`.                                     |
| `--format`         | `skt`            | Ausgabeformat: `skt` oder `json`.                                                                   |
| `--header-language`| `c`              | Parser für `.h`-Dateien: `c` oder `cpp`. Andere Erweiterungen bleiben unverändert.                              |
| `--max-lines`      | `500`            | Zeilenziel für flache Ausgabe; harte Grenze (Minimum 3) für jeden Baum-Quell-Chunk.                   |
| `--collect-test`   | `false`          | Testdateien einbeziehen.                                                                               |
| `--minify`         | `false`          | Wörterbuchbasierte Komprimierung verwenden.                                                                 |
| `--edges`          | `false`          | Den Abschnitt `[edges]` einbeziehen (standardmäßig deaktiviert, um Tokens zu sparen).                                    |
| `--clean`          | `false`          | Vorherige generierte Ausgabe ersetzen, erst nach erfolgreicher Generierung.                                   |
| `--workers`        | `NumCPU`         | Maximale Anzahl gleichzeitiger Parsing-Goroutinen.                                                            |
| `--verbose`        | `false`          | Fortschrittsinformationen ausgeben.                                                                       |

## Tipps

- Verwenden Sie `directory-flat` für die meisten Projekte und lesen Sie nur die Chunks, die Sie benötigen.
- Verwenden Sie `--minify` bei großen Codebasen, um Tokens zu sparen.
- Übergeben Sie `--edges` nur, wenn Sie Beziehungsdaten benötigen.
- Überprüfen Sie die [getestete Fähigkeitsmatrix und bekannte Grenzen](../../reference/language-support/), bevor Sie sich auf die Extraktion für eine bestimmte Sprache verlassen.

## Siehe auch

- [Ausgabemodi](../output-modes/)
- [Referenz zum Ausgabeformat](../../reference/output-format/)
- [Analysezuverlässigkeit](../../reference/analysis-reliability/)
- [Graph-Befehle](../graph-commands/)