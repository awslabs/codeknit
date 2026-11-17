---
title: Referenz zum Ausgabefomat
description: Vollständige Referenz für das von codeknit verwendete .skt-Ausgabeformat.
---

Diese Seite beschreibt die Grammatik der `.skt`-Ausgabe und die äquivalente JSON-Ausgabe. Für die Regeln, die entscheiden, wie Beziehungen aufgelöst werden, siehe [Sprachunterstützung und API-Treue](../language-support/).

## Übersicht

Eine `.skt`-Datei ist eine kompakte Textdatei, die aus Abschnitten besteht. Jeder Abschnitt beginnt mit einer Kopfzeile in eckigen Klammern.

| Abschnitt     | Vorhanden, wenn                                      | Inhalt                                           |
| ------------- | ---------------------------------------------------- | ------------------------------------------------ |
| `[symbols]`   | Immer                                                | Symboldefinitionen, gruppiert nach Quelldatei    |
| `[edges]`     | `--edges` ist gesetzt                                | Beziehungen zwischen Symbolen                    |
| `[errors]`    | Ein Scan- oder Parse-Diagnose ist aufgetreten        | Dateien, die nicht vollständig verarbeitet werden konnten |
| `[dict]`      | `--minify` ist gesetzt                               | Wörterbuch, das Kurzkodierungen zu wiederholten Tokens abbildet |

In Verzeichnis-Ausgabemodi können `[errors]` und `[dict]` in die Hilfsdateien `warnings.skt` und `dict.skt` geschrieben werden. Im `inline`-Modus erscheinen sie in derselben Ausgabe wie die anderen Abschnitte.

`--format json` erzeugt ein JSON-Dokument mit denselben Informationen. Siehe [JSON-Ausgabe](#json-ausgabe).

## [symbols]

Symbole sind nach Quelldatei gruppiert. Jede Gruppe beginnt mit einer Kopfzeile `## <Pfad>`, gefolgt von einer Zeile pro Symbol.

```skt
[symbols]
## pkg/services/auth.go
S1 module/package L1-L1 services {}
S2 type/struct L5-L8 AuthService {exported}
```

### Symbolzeile

```text
ShortID Kategorie/Art Lstart-Lend Signatur {Eigenschaften}
```

| Feld            | Beschreibung                                                                                               |
| --------------- | --------------------------------------------------------------------------------------------------------- |
| `ShortID`       | Sequenzielle Kennung: `S1`, `S2`, `S3` usw. Kanten verweisen auf Symbole mit diesem Wert.                 |
| `Kategorie/Art` | Klassifizierung des Symbols. Siehe [Kategorien und Arten](#kategorien-und-arten).                         |
| `Lstart-Lend`   | Erste und letzte Zeile des Symbols in der Quelldatei.                                                     |
| `Signatur`      | Symbolname, mit Parametern und Rückgabetyp für Callables. Siehe [Signaturformate](#signaturformate).     |
| `{Eigenschaften}` | Metadaten in Klammern. Siehe [Eigenschaften](#eigenschaften).                                             |

Eine Symbolzeile bleibt auf einer physischen Zeile. Zeilenumbrüche innerhalb einer Signatur werden als `\n` und `\r` escaped.

Symbol-IDs können sich ändern, wenn sich die Extraktion ändert. Erzeugen Sie alle Ausgabechunks gemeinsam, anstatt Chunks aus verschiedenen Läufen zu mischen. Dies gilt insbesondere beim Upgrade auf 0.5.0, da sich Symbol-IDs und Ergebnisse der Beziehungsauflösung ändern können.

### Kategorien und Arten

| Kategorie | Arten                                      | Beispiele                                           |
| --------- | ------------------------------------------ | --------------------------------------------------- |
| `callable` | function, method, constructor              | `callable/function`, `callable/method`              |
| `type`     | class, interface, struct, enum             | `type/class`, `type/interface`                      |
| `value`    | variable, constant, field                  | `value/variable`, `value/constant`                  |
| `module`   | package, namespace                         | `module/package`                                    |
| `meta`     | type parameters, files, includes, metadata | `meta/type_parameter`, `meta/file`, `meta/include`  |

### Signaturformate

| Form                         | Verwendet für                                  |
| ---------------------------- | ---------------------------------------------- |
| `name`                       | Typen, Werte und Module                        |
| `name(params)`               | Callables ohne Rückgabetyp                     |
| `name(params) -> returnType` | Callables mit Rückgabetyp                      |

Parameter werden wie folgt geschrieben:

| Form          | Verwendet für                  |
| ------------- | ------------------------------ |
| `name`        | Sprachen ohne Typen            |
| `name: type`  | Sprachen mit Typen             |

Wenn ein Typverweis mit einem bekannten Symbol übereinstimmt, wird der Typname durch die ShortID dieses Symbols ersetzt. Zum Beispiel verweist `-> *S2` auf einen Zeiger auf das Symbol `S2`.

### Eigenschaften

Eigenschaften erscheinen in Klammern nach der Signatur, getrennt durch Kommas. Leere Klammern (`{}`) bedeuten, dass das Symbol keine Eigenschaften hat.

| Eigenschaftsform   | Beispiel                 |
| ------------------ | ------------------------ |
| Flag               | `exported`, `async`, `static` |
| `Schlüssel=Wert`   | `visibility=public`, `receiver=*AuthService`, `exported=false`, `resolution=unresolved` |

### Beispiel

```skt
[symbols]
## pkg/services/auth.go
S1 module/package L1-L1 services {}
S2 type/struct L5-L8 AuthService {exported}
S3 callable/function L10-L12 NewAuthService(secret: string, ttl: int) -> *S2 {exported}
S4 callable/method L14-L19 Authenticate(token: string) {exported, receiver=*AuthService}
S5 callable/function L29-L31 verifyToken(token: string) -> bool {exported=false}
```

## [edges]

Der Abschnitt `[edges]` listet Beziehungen zwischen Symbolen auf. Er wird nur geschrieben, wenn Kanten mit `--edges` angefordert werden.

### Kantenzeile

```text
FromID --Art--> ToID1, ToID2
FromID --Art[unresolved]--> Ziel1, Ziel2
FromID --Art[ambiguous]--> Ziel [candidates=KandidatID1, KandidatID2]
```

| Teil                     | Beschreibung                                                                           |
| ------------------------ | -------------------------------------------------------------------------------------- |
| `FromID`                 | Quellendpunkt.                                                                         |
| `Art`                    | Art der Beziehung. Siehe [Kantenarten](#kantenarten).                                  |
| `[unresolved]`, `[ambiguous]` | Optionaler Auflösungsmarker. Siehe [Auflösungsmarker](#auflösungsmarker).          |
| `ToID1, ToID2`           | Ein oder mehrere Ziele, durch Kommas getrennt.                                         |
| `[candidates=...]`       | Mögliche Ziele für eine mehrdeutige Kante, als kommagetrennte Liste in Klammern.       |

Ziele teilen sich eine Zeile nur, wenn Quelle, Art, Auflösungsmarker und Kandidatensatz identisch sind. Eine Definition, auf die eine Kante verweist, kann in einer späteren Datei oder einem Ausgabechunk vorkommen. Laden Sie alle Chunks, bevor Sie Kanten interpretieren.

### Auflösungsmarker

| Marker         | Bedeutung                                                                              |
| -------------- | -------------------------------------------------------------------------------------- |
| Keiner         | Der Analysator hat die Beziehung aufgelöst.                                            |
| `[unresolved]` | Der Analysator konnte die Abhängigkeit nicht herstellen.                               |
| `[ambiguous]`  | Mehr als eine Interpretation bleibt möglich. Die Kandidaten-IDs listen mögliche Ziele auf. |

Ein fehlender Marker ist keine Compiler-Garantie. Kandidaten-IDs identifizieren mögliche Ziele, keine bestätigten Verbindungen. Unsicherheit wird immer in dieser SKT-Grammatik geschrieben, niemals als eingebettetes JSON.

### Endpunktformen

| Form                      | Wann verwendet                                                                                         | Beispiel                     |
| ------------------------- | ----------------------------------------------------------------------------------------------------- | --------------------------- |
| ShortID                   | Der Endpunkt oder Kandidat ist ein bekanntes Symbol. Dies umfasst Endpunkte von unaufgelösten Kanten und Include-Symbolen. | `S4`, `S8`       |
| Bare Name                 | Das Ziel hat kein Symbol. Der Name ist relativ zur Quelldatei.                                        | `Refresh`, `string`         |
| Angeführter vollständiger ID | Die Identität ist komplex oder ein Literalname ähnelt einer ShortID.                                  | `"app.go::S2"`              |

Angeführte vollständige IDs escapen Anführungszeichen, Backslashes und Zeilenumbrüche. Ein Literalname wie `S2` muss als angeführter vollständiger ID geschrieben werden, damit er nicht als ShortID gelesen wird.

```skt
S4 --references[unresolved]--> string, bool
S5 --calls[ambiguous]--> Helper [candidates=S8, S9]
S6 --references[unresolved]--> "app.go::S2"
```

### Kantenarten

| Art          | Bedeutung                                         |
| ------------ | ------------------------------------------------- |
| `calls`      | Funktions-/Methodenaufruf                         |
| `contains`   | Klasse enthält Methode, Modul enthält Funktion    |
| `inherits`   | Klasse erweitert eine andere                      |
| `implements` | Klasse implementiert Interface                    |
| `overrides`  | Methode überschreibt Elternmethode                |
| `references` | Symbol verweist auf ein anderes                   |
| `imports`    | Modul importiert ein anderes                      |
| `decorates`  | Dekorator wird auf Symbol angewendet              |

Zurückgegebene Callables und Callback-Argumente werden als `references` aufgezeichnet, nicht als separate Art.

### C- und C++-Datei- und Include-Symbole

C- und C++-Quellen erzeugen `meta/file`- und `meta/include`-Symbole, sodass Include-Beziehungen ShortIDs haben. Jede Header-Schreibweise erscheint einmal in der Symbolliste, und Kanten verwenden die ShortIDs. Wiederholte Includes innerhalb einer Datei teilen sich einen Endpunkt.

```skt
[symbols]
## src/main.c
S1 meta/file L1-L4 main.c
S2 meta/include L1-L1 "config.h" {resolution=unresolved}
S3 meta/include L2-L2 "utils.h" {resolution=unresolved}
S4 meta/include L3-L3 <stdio.h> {resolution=unresolved}

[edges]
S1 --references[unresolved]--> S2, S3, S4
```

Ein Include-Symbol zeichnet die Direktive so auf, wie sie geschrieben wurde. Es stellt keine aufgelöste Header-Deklaration dar.

### Auswirkungen auf Analyse und Visualisierung

- Unsichere Kanten und Datei-/Include-Metadaten tragen nicht zu Abhängigkeitsmetriken bei. Berichte zeigen die Anzahl der ausgeschlossenen Beziehungen an.
- Der HTML-Graph zeichnet aufgelöste Links und zeigt eine Anzahl unsicherer Beziehungen an. Er zeichnet keine einzelnen unsicheren Kanten.

### Beispiel

```skt
[edges]
S2 --contains--> S4
S4 --calls--> S5
S10 --inherits--> S2
S24 --implements--> S19
```

## [errors]

Der Abschnitt `[errors]` listet Dateien auf, die nicht vollständig verarbeitet wurden. Er ist nur vorhanden, wenn Diagnosen existieren.

```text
- <Pfad>: <Nachricht>
```

```skt
[errors]
- path/to/broken.go: syntax error at line 42
```

## [dict]

Der Abschnitt `[dict]` erscheint nur in minimierter Ausgabe (`--minify`). Er bildet Kurzkodierungen auf wiederholte Tokens ab.

| Teil       | Beschreibung                                                                 |
| ---------- | --------------------------------------------------------------------------- |
| `d0`, `d1`, … | Wörterbuchcodes.                                                        |
| Eintragszeile | `- <Code>: <Token>`                                                         |

Um minimierte Ausgabe zu lesen, ersetzen Sie jeden Code durch seinen Wörterbuchwert. Codes können Kategorie/Art-Paare und Eigenschaften in `[symbols]` ersetzen. Sie können auch Beziehungsarten in `[edges]` ersetzen. Unsicherheitsmarker bleiben lesbar. Zum Beispiel ist `S1 --d2[unresolved]--> S3` eine unaufgelöste Kante, deren Art der Wert von `d2` ist.

### Minimiertes Beispiel

```skt
[dict]
- d0: async=false
- d1: callable/method
- d2: exported

[symbols]
## src/handler.py
S1 type/class L1-L6 Handler {}
S2 d1 L2-L3 __init__(name) {d0}
S3 d1 L5-L6 handle(request) {d0}

[edges]
S1 --contains--> S2, S3
```

Hier ist `S2` ein `callable/method` mit der Eigenschaft `async=false`.

## Vollständiges Beispiel

Die folgende Ausgabe enthält `[symbols]`, `[edges]` und `[errors]`. Sie ist nicht minimiert, hat also keinen `[dict]`-Abschnitt; siehe [das minimierte Beispiel](#minimiertes-beispiel).

```skt
[symbols]
## pkg/services/auth.go
S1 module/package L1-L1 services {}
S2 type/struct L5-L8 AuthService {exported}
S3 callable/function L10-L12 NewAuthService(secret: string, ttl: int) -> *S2 {exported}
S4 callable/method L14-L19 Authenticate(token: string) {exported, receiver=*AuthService}
S5 callable/function L29-L31 verifyToken(token: string) -> bool {exported=false}
## pkg/util/hash.go
S6 callable/function L3-L5 Hash(data: string) -> string {exported}
## pkg/util/hash_alt.go
S7 callable/function L3-L5 Hash(data: string) -> string {exported}

[edges]
S1 --contains--> S2, S3, S4, S5
S2 --contains--> S4
S3 --references--> S2
S4 --calls--> S5
S4 --calls[unresolved]--> Refresh
S5 --calls[ambiguous]--> Hash [candidates=S6, S7]

[errors]
- pkg/legacy/broken.go: syntax error at line 42
```

## JSON-Ausgabe

`--format json`