---
title: Ausgabemodus-Referenz
description: Vollständige Referenz für das von codeknit verwendete .skt-Ausgabeformat.
---

Das `.skt`-Format (Skeleton) ist ein kompaktes, menschenlesbares Textformat, das von `codeknit` verwendet wird, um extrahierte **Codestruktur** darzustellen. Es enthält **Symbole**, Beziehungen und Metadaten in einer minimalen Form, die für den LLM-Verbrauch und die strukturelle Analyse geeignet ist.

Eine `.skt`-Datei ist in Abschnitte unterteilt. Jeder Abschnitt beginnt mit einer Kopfzeile in eckigen Klammern. Die Abschnitte können in beliebiger Reihenfolge erscheinen, wobei `[symbols]` typischerweise zuerst kommt.

## [symbols]

Der Abschnitt `[symbols]` listet alle extrahierten **Symbole** auf, gruppiert nach ihrer Quelldatei. Jede Datei wird mit einer `##`-Überschrift eingeleitet, gefolgt vom Dateipfad.

### Zeilenformat

Jedes **Symbol** wird in einer einzelnen Zeile mit folgender Struktur dargestellt:

```
ShortID category/kind Lstart-Lend signature {properties}
```

### Felder

- **ShortID**: Ein sequenzieller Bezeichner, der jedem **Symbol** zugewiesen wird (z. B. `S1`, `S2`, `S3`). Wird als Referenz in **Kanten** und anderen Abschnitten verwendet.
- **category/kind**: Ein durch Schrägstrich getrenntes Paar, das die Kategorie und die spezifische Art des **Symbols** angibt.
- **Lstart-Lend**: Der **Zeilenbereich** in der Quelldatei, in dem das **Symbol** definiert ist (z. B. `L10-L15`).
- **signature**: Der Name und die Typinformationen des **Symbols**. Das Format hängt vom **Symbol** ab:
  - `name` — für Typen, Werte, Module
  - `name(params)` — für aufrufbare Elemente ohne Rückgabetyp
  - `name(params) -> returnType` — für aufrufbare Elemente mit Rückgabetyp
- **{properties}**: Optionale Metadaten, die in geschweiften Klammern eingeschlossen sind. Mehrere Eigenschaften werden durch Kommas getrennt.

### Parameter

- In untypisierten Sprachen: `paramName`
- In typisierten Sprachen: `paramName: type`
- Typreferenzen, die bekannten **Symbolen** entsprechen, werden durch ihre ShortIDs ersetzt (z. B. `config: S5` anstelle von `config: Config`).

### Eigenschaften

Häufige Eigenschaften sind:

- `async`: `true` oder `false`
- `exported`: `true` oder `false`
- `static`: vorhanden, wenn das **Symbol** statisch ist
- `visibility=public|private|protected`
- `receiver=*TypeName`: für Methoden, gibt den Empfängertyp an

### Symbolkategorien und -arten

| Kategorie  | Arten                          | Beispiele                              |
| ---------- | ------------------------------ | -------------------------------------- |
| `callable` | function, method, constructor  | `callable/function`, `callable/method` |
| `type`     | class, interface, struct, enum | `type/class`, `type/interface`         |
| `value`    | variable, constant, field      | `value/variable`, `value/constant`     |
| `module`   | package, namespace             | `module/package`                       |
| `meta`     | type parameters, metadata      | `meta/type_parameter`                  |

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

Der Abschnitt `[edges]` definiert Beziehungen zwischen **Symbolen** mithilfe ihrer ShortIDs.

### Zeilenformat

```
FromID --kind--> ToID1, ToID2
```

Mehrere Ziel-IDs werden durch Kommas getrennt. Jede Zeile stellt eine gerichtete Beziehung dar.

### Kantenarten

| Art          | Bedeutung                                      |
| ------------ | ---------------------------------------------- |
| `calls`      | Aufruf einer Funktion/Methode                  |
| `contains`   | Klasse enthält Methode, Modul enthält Funktion |
| `inherits`   | Klasse erbt von einer anderen Klasse           |
| `implements` | Klasse implementiert Interface                 |
| `overrides`  | Methode überschreibt Elternmethode             |
| `references` | **Symbol** referenziert ein anderes **Symbol** |
| `imports`    | Modul importiert ein anderes Modul             |
| `decorates`  | Dekorator wird auf ein **Symbol** angewendet   |

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

Der Abschnitt `[dict]` erscheint nur, wenn das Flag `--minify` verwendet wird. Er bildet kurze Dictionary-Codes auf wiederholte String-Token ab, um die Ausgabegöße zu reduzieren.

### Format

Jede Zeile bildet einen Dictionary-Code (`d0`, `d1` usw.) auf seinen erweiterten Wert ab:

```
- d0: async=false
- d1: callable/method
- d2: exported
```

Im Rest der Datei ersetzen diese Codes ihre vollständigen Werte.

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
S3 --returns--> S2

[errors]
- utils/broken.go: syntax error at line 5
```
