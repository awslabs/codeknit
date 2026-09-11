---
title: Sprachunterstützung und API-Treue
description: Getestete Extraktionsfähigkeiten, Parser-Auswahl und bekannte Sprachbeschränkungen.
---

<!-- Maintained in docs/reference/language-support.md; copied verbatim by docs/gen. -->

Codeknit extrahiert Syntax, ohne einen Compiler auszuführen, Makros zu evaluieren oder die Abhängigkeiten eines Projekts zu laden. Die folgende Matrix beschreibt getestete Konstrukte, nicht die vollständige Sprachkonformität oder garantierte Namensauflösung über Dateien hinweg.

Jede Sprache verfügt über eine exakte grundlegende Parameter-/Rückgabesignatur-Prüfung in `internal/plugin/api_fidelity_test.go`. Die Deklarationsabdeckung wird auch in `internal/plugin/coverage_test.go` und den einzelnen Sprachtestpaketen überprüft.

| Sprache    | Erweiterungen                               | Signatur-Baseline                                                                                                                | Deklarations- und Container-Abdeckung                                                                           |
| ---------- | ------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------- |
| C          | `.c`, `.h`                                  | Benannte Parameter und Rückgabetypen in Definitionen und Prototypen                                                              | Structs, Unions, Enums, Typedefs, Makros; Deklarationen innerhalb von Header-Guards                             |
| C++        | `.cpp`, `.cc`, `.cxx`, `.hpp`, `.hxx`, `.h` | Benannte Parameter und Rückgabetypen in Funktionen, Prototypen und Methodendeklarationen                                         | Klassen, Structs, Methoden, Namespaces, Templates; Header-Guards und `extern "C"`-Wrapper                       |
| C#         | `.cs`                                       | Benannte typisierte Parameter und explizite Rückgabetypen                                                                        | Klassen, Structs, Interfaces, Methoden, Felder, Delegates, Namespaces                                           |
| Go         | `.go`                                       | Gruppierte, unbenannte, variadische, funktionsgetypte und generische Parameter; benannte/mehrere Ergebnisse; generische Receiver | Struct-Felder mit Typen und Tags, eingebettete Felder, Interface-Methodensignaturen, Typparameter, Pakete       |
| Java       | `.java`                                     | Benannte typisierte Parameter und explizite Rückgabetypen                                                                        | Klassen, Interfaces, Konstruktoren, Methoden, Felder, verschachtelte Klassen, Pakete                            |
| JavaScript | `.js`, `.jsx`, `.mjs`, `.cjs`               | Benannte Parameter; kein inferierter Rückgabetyp                                                                                 | Funktionen, Arrow-Funktionen, Klassen, Felder, Methoden, JSX-enthaltende Deklarationen                          |
| PHP        | `.php`                                      | Benannte typisierte Parameter und explizite Rückgabetypen                                                                        | Klassen, Interfaces, Traits, Enums, Methoden, Eigenschaften, Namespaces                                         |
| Python     | `.py`, `.pyi`                               | Annotationen, Defaults, `*args`, `**kwargs`, Positions-/Schlüsselwort-Trenner, Rückgabeannotationen                              | Funktionen, Klassen, verschachtelte Klassen, dekorierte Methoden; konventionelle gebundene Receiver weggelassen |
| Ruby       | `.rb`                                       | Benannte Parameter; kein inferierter Rückgabetyp                                                                                 | Klassen, Module, Instanz-/Singleton-Methoden, Konstanten                                                        |
| Rust       | `.rs`                                       | Typisierte Muster, `self`/`&self`/`&mut self`, mutable Bindings, explizite Rückgabetypen                                         | Inline- und verschachtelte Module, benannte Struct-Felder, Trait-Methoden und Implementierungsmethoden          |
| Scala      | `.scala`, `.sc`                             | Benannte typisierte Parameter und explizite Rückgabetypen                                                                        | Klassen, Traits, Objekte, Methoden, Enums, Pakete                                                               |
| TypeScript | `.ts`, `.tsx`, `.mts`, `.cts`               | Benannte typisierte Parameter und explizite Rückgabeannotationen                                                                 | Funktionen, Arrow-Funktionen, Interfaces, Klassen, Felder, Namespaces; TSX verwendet eine eigene Grammatik      |

## Header-Sprachauswahl

Alle CLI-Befehle zur Quellcodeverarbeitung akzeptieren `--header-language c|cpp`. Die Auswahl gilt nur für `.h`-Dateien. Andere Erweiterungen behalten ihre üblichen Parser.

```bash
codeknit parse ./src --header-language cpp
codeknit fingerprint ./src --header-language c
codeknit graph analyze ./src --header-language cpp
```

Der Standard ist `c`, um die Extraktion von C-Makros, Typedefs und Unions zu erhalten. Verwenden Sie `cpp` für C++-Header mit der Erweiterung `.h`; Dateien mit den Erweiterungen `.hpp` und `.hxx` verwenden bereits C++. Die Auswahl ist explizit, da beide Grammatiken einen Teil der Syntax der jeweils anderen Sprache akzeptieren können, ohne deren vollständige Struktur zu extrahieren.

Header-Guards und bedingte Zweige werden traversiert, ohne ihre Bedingungen zu evaluieren. Die Ergebnisse können daher Deklarationen aus sich gegenseitig ausschließenden Zweigen enthalten. Makroexpansion, Include-Auflösung und Build-Flags werden nicht angewendet. Die TUI verwendet die Standardauswahl C für `.h`-Dateien.

## Beibehaltung der API-Form

Zum Beispiel Go:

```go
func Pair(a, b int, rest ...string) (string, error)
```

erzeugt:

```text
Pair(a: int, b: int, rest: ...string) -> (string, error)
```

Go-Interface-Methoden und benannte Felder in Go/Rust erscheinen als **Symbole** in beiden normalen und Fingerprint-Modi. Methoden und Felder behalten ihren zugehörigen Typ. Verschachtelte Rust-Module behalten qualifizierte Namen wie `service.storage.read`, auch wenn ein anderes Modul eine gleichnamige Funktion deklariert.

Rust erlaubt Felder und Methoden mit demselben Namen. Feldbezogene Namen tragen intern ein `#field`-Suffix, sodass ihre IDs und Containment-**Kanten** unterscheidbar bleiben; Anzeigenamen und Signaturen behalten den ursprünglichen Feldnamen bei.

Python-Signaturen behalten Annotationen, Defaults, Splat-Operatoren und Aufrufkonvention-Trenner. Ein konventioneller erster `self`- oder `cls`-Parameter wird bei einer gebundenen Methode weggelassen, aber in einer freien Funktion oder statischen Methode beibehalten.

Mehrzeilige Signaturen behalten Zeilenumbrüche im JSON-Format bei. SKT maskiert Zeilenumbrüche als `\n` und `\r`, sodass jedes **Symbol** auf einer physischen Ausgabezeile bleibt.

## Bekannte Einschränkungen

- Komplexe C/C++-Zeiger-, Referenz-, Array- und Funktionszeiger-Deklaratoren können weiterhin unvollständige Signaturen aufweisen. Makrogenerierte Deklarationen werden nicht expandiert.
- Die Handhabung von TypeScript/JavaScript-Konstruktoren und einige optionale, Rest- oder destrukturierte Parameterformen sind unvollständig.
- Fortgeschrittene generische Constraints und Typformen außerhalb der getesteten Fälle variieren je nach Sprache. Rückgabetypen werden nicht inferiert.
- Rust-Makros werden nicht expandiert. Out-of-Line-Moduldeklarationen werden erfasst; ihre Dateien müssen im Scan enthalten sein. Tuple-Felder, Enum-Variantendetails und assoziierte Elemente haben teilweise Abdeckung.
- Die Syntaxextraktion löst dynamische Dispatches, Überladungen, importierte Abhängigkeiten oder Build-Konfigurationssemantik nicht mit Compiler-Präzision auf.
- Teilweise Syntaxfehler können weiterhin Deklarationen auslassen. Überprüfen Sie gemeldete Warnungen und verifizieren Sie den Quellcode, bevor Sie ein Refactoring anwenden.

Vollständigere Extraktionen ändern die Symbolanzahl und ShortIDs. Generieren Sie vorhandene Skelett-Ausgaben bei einem Upgrade neu, anstatt sie mit älterer Ausgabe zu mischen.
