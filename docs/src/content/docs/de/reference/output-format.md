---
title: Referenz zum Ausgabefomat
description: Vollständige Referenz für das von codeknit verwendete .skt-Ausgabeformat.
---

Das `.skt`-Format (Skeleton) ist ein kompaktes, menschenlesbares Textformat, das von `codeknit` verwendet wird, um extrahierte **Codestruktur** darzustellen. Es enthält Symbole, Beziehungen und Metadaten in einer minimalen Form, die für die Nutzung durch LLMs und strukturelle Analysen geeignet ist.

Eine `.skt`-Datei ist in Abschnitte unterteilt. Jeder Abschnitt beginnt mit einer Überschrift in eckigen Klammern. Die Abschnitte können in beliebiger Reihenfolge erscheinen, obwohl `[symbols]` typischerweise zuerst kommt.

## [symbols]

Der Abschnitt `[symbols]` listet alle extrahierten Symbole, gruppiert nach ihrer Quelldatei. Jede Datei wird mit einer `##`-Überschrift gefolgt vom Dateipfad eingeführt.

### Zeilenformat

Jedes Symbol wird in einer einzelnen Zeile mit folgender Struktur dargestellt:

```
ShortID category/kind Lstart-Lend signature {properties}
```

### Felder

- **ShortID**: Eine sequentielle Kennung, die jedem Symbol zugewiesen wird (z. B. `S1`, `S2`, `S3`). Wird als Referenz in Kanten und anderen Abschnitten verwendet.
- **Kategorie/Art**: Ein durch Schrägstrich getrenntes Paar, das die Kategorie und die spezifische Art des Symbols angibt.
- **Lstart-Lend**: Der **Zeilenbereich** im Quellcode, in dem das Symbol definiert ist (z. B. `L10-L15`).
- **Signatur**: Der Name und die Typinformationen des Symbols. Das Format hängt vom Symbol ab:
  - `name` — für Typen, Werte, Module
  - `name(params)` — für Callables ohne Rückgabetyp
  - `name(params) -> returnType` — für Callables mit Rückgabetyp
- **{Eigenschaften}**: Optionale Metadaten, die in geschweiften Klammern eingeschlossen sind. Mehrere Eigenschaften werden durch Kommas getrennt.

### Parameter

- In untypisierten Sprachen: `paramName`
- In typisierten Sprachen: `paramName: type`
- Typreferenzen, die bekannten Symbolen entsprechen, werden durch ihre ShortIDs ersetzt (z. B. `config: S5` anstelle von `config: Config`).

### Eigenschaften

Häufige Eigenschaften sind:

- `async`: `true` oder `false`
- `exported`: `true` oder `false`
- `static`: vorhanden, wenn das Symbol statisch ist
- `visibility=public|private|protected`
- `receiver=*TypeName`: für Methoden, gibt den Empfängertyp an

### Symbolkategorien und -arten

| Kategorie   | Arten                                      | Beispiele                                           |
| ----------- | ------------------------------------------ | --------------------------------------------------- |
| `callable`  | function, method, constructor              | `callable/function`, `callable/method`              |
| `type`      | class, interface, struct, enum             | `type/class`, `type/interface`                      |
| `value`     | variable, constant, field                  | `value/variable`, `value/constant`                  |
| `module`    | package, namespace                         | `module/package`                                    |
| `meta`      | type parameters, files, includes, metadata | `meta/type_parameter`, `meta/file`, `meta/include`  |

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

Der Abschnitt `[edges]` zeichnet Beziehungen auf, wenn `--edges` aktiviert ist. Bekannte Endpunkte verwenden ihre ShortIDs; Ziele, die nicht zu einem Symbol aufgelöst werden können, können als Namen erscheinen.

### Zeilenformat

```
FromID --kind--> ToID1, ToID2
FromID --kind[unresolved]--> Target1, Target2
FromID --kind[ambiguous]--> Target [candidates=CandidateID1, CandidateID2]
```

Ein fehlender Status bedeutet, dass der Analysator die Beziehung aufgelöst hat. `unresolved` bedeutet, dass die Abhängigkeit nicht hergestellt werden konnte; `ambiguous` bedeutet, dass konkurrierende Interpretationen möglich bleiben. Kandidaten-IDs identifizieren mögliche Ziele, keine bestätigten Verbindungen. Die Auflösung basiert auf Syntax und Gültigkeitsbereich, nicht auf einer Compiler-Garantie.

Ziele werden nur gruppiert, wenn ihre Quelle, Art, Auflösung und Kandidatenmenge übereinstimmen. Namen ohne Symbole sind relativ zur Quelldatei. Komplexe Identitäten verwenden zitierte vollständige IDs mit escaped Anführungszeichen, Backslashes und Zeilenumbrüchen. Ein wörtlicher Name, der einer ShortID ähnelt, muss als vollständige ID zitiert werden, z. B. `"app.go::S2"`. Symboldefinitionen können in späteren Dateien oder Ausgabe-Chunks vorkommen.

```skt
S4 --references[unresolved]--> string, bool
S5 --calls[ambiguous]--> Helper [candidates=S8, S9]
S6 --references[unresolved]--> "app.go::S2"
```

### Datei- und Include-Endpunkte

C/C++-Includes haben explizite Metadaten-Symbole. Header-Schreibweisen erscheinen einmal in der Symbolliste, und Kanten verwenden ShortIDs, obwohl die Header-Auflösung ungelöst bleibt:

```skt
[symbols]
## src/main.c
S1 meta/file L1-L3 main.c
S2 meta/include L1-L1 "config.h" {resolution=unresolved}
S3 meta/include L2-L2 <stdio.h> {resolution=unresolved}

[edges]
S1 --references[unresolved]--> S2, S3
```

Ein Include-Symbol zeichnet die geschriebene Direktive auf, nicht eine aufgelöste Header-Deklaration. Wiederholte Includes innerhalb einer Datei teilen sich einen Endpunkt. Datei-/Include-Metadaten werden aus Abhängigkeitsrankings und **Dead-Code**-Ergebnissen ausgeschlossen.

Unsichere Beziehungen werden im Parse-Output beibehalten, aber aus Abhängigkeitsmetriken ausgeschlossen. Graphberichte zeigen die Anzahl ausgeschlossener Beziehungen an; der HTML-Viewer zeigt deren Anzahl an, ohne einzelne unsichere Kanten zu zeichnen.

Beim Upgrade auf `0.5.0` sollten alle Ausgabe-Chunks gemeinsam neu generiert werden. Symbol-IDs und Beziehungsergebnisse können sich ändern, daher dürfen Chunks aus verschiedenen Läufen nicht gemischt werden.

### Kantenarten

| Art           | Bedeutung                                      |
| ------------- | ---------------------------------------------- |
| `calls`       | Funktions-/Methodenaufruf                      |
| `contains`    | Klasse enthält Methode, Modul enthält Funktion |
| `inherits`    | Klasse erweitert eine andere Klasse            |
| `implements`  | Klasse implementiert Interface                 |
| `overrides`   | Methode überschreibt Elternmethode             |
| `references`  | Symbol referenziert ein anderes Symbol         |
| `imports`     | Modul importiert ein anderes Modul             |
| `decorates`   | Dekorator auf ein Symbol angewendet            |

### Beispiel

```skt
[edges]
S2 --contains--> S4
S4 --calls--> S5
S10 --inherits--> S2
S24 --implements--> S19
```

## [errors]

Der Abschnitt `[errors]` listet Dateien auf, die nicht vollständig geparst werden konnten.

### Format

Jede Zeile beginnt mit `-`, gefolgt vom Dateipfad und der Fehlermeldung:

```
- path/to/file.go: syntax error at line 42
```

### Beispiel

```skt
[errors]
- src/broken.go: unexpected token at line 10
- tests/corner_case.py: unterminated string literal
```

## [dict]

Der Abschnitt `[dict]` erscheint nur, wenn das Flag `--minify` verwendet wird. Er bildet kurze Dictionary-Codes auf wiederholte String-Tokens ab, um die Ausgabegröße zu reduzieren.

### Format

Jede Zeile bildet einen Dictionary-Code (`d0`, `d1` usw.) auf seinen erweiterten Wert ab:

```
- d0: async=false
- d1: callable/method
- d2: exported
```

Im Rest der Datei ersetzen diese Codes ihre vollständigen Werte.

Beziehungstypen können ebenfalls Dictionary-Codes verwenden, während Statusbezeichnungen lesbar bleiben: `S1 --d2[unresolved]--> S3`. Erweitern Sie `d2` mithilfe des Dictionarys und behalten Sie den `unresolved`-Status bei.

### Beispiel

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

## Vollständiges Beispiel

```skt
[dict]
- d0: exported
- d1: callable/function

[symbols]
## main.go
S1 module/package L1-L1 main {}
S2 type/struct L5-L8 Server {d0}
S3 d1 L10-L12 NewServer(addr: string) -> *S2 {d0}
S4 callable/method L14-L20 Serve() {d0, receiver=*Server}
S5 callable/function L22-L25 handleError(err: error) -> bool {}

[edges]
S2 --contains--> S4
S4 --calls--> S5
S3 --references--> S2

[errors]
- utils/broken.go: syntax error at line 5
```

## JSON-Beziehungen

Verwenden Sie JSON, wenn ein Skript oder eine Integration strukturierte Ausgabe benötigt:

```bash
codeknit parse ./src --output-mode inline --format json --edges
```

Jede Kante enthält `from`, `to` und `kind`. Bekannte Quell-IDs erscheinen in `from_short`; aufgelöste Ziel-IDs erscheinen in `to_short`. Unsichere Kanten fügen `resolution` und, falls verfügbar, `candidates` mit vollständigen Symbol-IDs hinzu. Sie lassen `to_short` weg, selbst wenn das Ziel ein Include-Metadaten-Symbol hat.

Zum Beispiel kann ein mehrdeutiger Aufruf wie folgt dargestellt werden:

```json
{
  "from": "app.py::run",
  "from_short": "S1",
  "to": "app.py::Helper",
  "kind": "calls",
  "resolution": "ambiguous",
  "candidates": ["helpers/a.py::Helper", "helpers/b.py::Helper"]
}
```

JSON und SKT bewahren dieselbe Unsicherheit mit unterschiedlichen Kodierungen. Weitere Informationen zu Auflösungsregeln und bekannten Grenzen finden Sie unter [Sprachunterstützung und API-Treue](../language-support/).