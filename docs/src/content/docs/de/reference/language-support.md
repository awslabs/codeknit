---
title: Sprachunterstützung und API-Treue
description: Getestete Extraktionsfähigkeiten, Parser-Auswahl und bekannte Sprachbeschränkungen.
---

<!-- Maintained in docs/reference/language-support.md; copied verbatim by docs/gen. -->

Codeknit extrahiert Syntax, ohne einen Compiler auszuführen, Makros zu evaluieren oder die Abhängigkeiten eines Projekts zu laden. Die folgende Matrix beschreibt getestete Konstrukte, nicht die vollständige Sprachkonformität oder garantierte Namensauflösung über Dateien hinweg.

Jede Sprache verfügt über eine exakte grundlegende Parameter-/Rückgabesignatur-Behauptung in `internal/plugin/api_fidelity_test.go`. Die Deklarationsabdeckung wird auch in `internal/plugin/coverage_test.go` und den einzelnen Sprachtestpaketen überprüft.

| Sprache     | Erweiterungen                                  | Signatur-Baseline                                                                                           | Deklarations- und Containerabdeckung                                                                         |
| ----------- | ---------------------------------------------- | ----------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------- |
| C           | `.c`, `.h`                                     | Benannte Parameter und Rückgabetypen in Definitionen und Prototypen                                        | Structs, Unions, Enums, Typedefs, Makros; Deklarationen innerhalb von Header-Guards                        |
| C++         | `.cpp`, `.cc`, `.cxx`, `.hpp`, `.hxx`, `.h`   | Benannte Parameter und Rückgabetypen in Funktionen, Prototypen und Methodendeklarationen                   | Klassen, Structs, Methoden, Namespaces, Templates; Header-Guards und `extern "C"`-Wrapper                  |
| C#          | `.cs`                                          | Benannte typisierte Parameter und explizite Rückgabetypen                                                   | Klassen, Structs, Interfaces, Methoden, Felder, Delegates, Namespaces                                      |
| Go          | `.go`                                          | Gruppierte, unbenannte, variadische, funktionsgetypte und generische Parameter; benannte/mehrere Ergebnisse; generische Receiver | Struct-Felder mit Typen und Tags, eingebettete Felder, Interface-Methodensignaturen, Typparameter, Pakete |
| Java        | `.java`                                        | Benannte typisierte Parameter und explizite Rückgabetypen                                                   | Klassen, Interfaces, Konstruktoren, Methoden, Felder, verschachtelte Klassen, Pakete                       |
| JavaScript  | `.js`, `.jsx`, `.mjs`, `.cjs`                  | Benannte Parameter; kein inferierter Rückgabetyp                                                           | Funktionen, Arrow-Funktionen, Klassen, Felder, Methoden, JSX-enthaltende Deklarationen                     |
| PHP         | `.php`                                         | Benannte typisierte Parameter und explizite Rückgabetypen                                                   | Klassen, Interfaces, Traits, Enums, Methoden, Eigenschaften, Namespaces                                    |
| Python      | `.py`, `.pyi`                                  | Annotationen, Defaults, `*args`, `**kwargs`, Positions-/Schlüsselwort-Trenner, Rückgabeannotationen        | Funktionen, Klassen, verschachtelte Klassen, dekorierte Methoden; konventionelle gebundene Receiver weggelassen |
| Ruby        | `.rb`                                          | Benannte Parameter; kein inferierter Rückgabetyp                                                           | Klassen, Module, Instanz-/Singleton-Methoden, Konstanten                                                   |
| Rust        | `.rs`                                          | Getypte Muster, `self`/`&self`/`&mut self`, mutable Bindings, explizite Rückgabetypen                       | Inline- und verschachtelte Module, benannte Struct-Felder, Trait-Methoden und Implementierungsmethoden     |
| Scala       | `.scala`, `.sc`                                | Benannte typisierte Parameter und explizite Rückgabetypen                                                   | Klassen, Traits, Objekte, Methoden, Enums, Pakete                                                          |
| TypeScript  | `.ts`, `.tsx`, `.mts`, `.cts`                  | Benannte typisierte Parameter und explizite Rückgabeannotationen                                            | Funktionen, Arrow-Funktionen, Interfaces, Klassen, Felder, Namespaces; TSX verwendet eigene Grammatik      |

## Header-Sprachauswahl

Alle CLI-Befehle zur Quellcodeverarbeitung akzeptieren `--header-language c|cpp`. Die Auswahl gilt nur für `.h`-Dateien. Andere Erweiterungen behalten ihre üblichen Parser bei.

```bash
codeknit parse ./src --header-language cpp
codeknit fingerprint ./src --header-language c
codeknit graph analyze ./src --header-language cpp
```

Die Standardeinstellung ist `c`, um die Extraktion von C-Makros, Typedefs und Unions zu erhalten. Verwenden Sie `cpp` für C++-Header mit der Erweiterung `.h`; Dateien mit den Erweiterungen `.hpp` und `.hxx` verwenden bereits C++. Die Auswahl ist explizit, da beide Grammatiken einen Teil der Syntax der jeweils anderen Sprache akzeptieren können, ohne deren vollständige Struktur zu extrahieren.

Header-Guards und bedingte Verzweigungen werden traversiert, ohne ihre Bedingungen zu evaluieren. Die Ergebnisse können daher Deklarationen aus sich gegenseitig ausschließenden Zweigen enthalten. Makroexpansion, Include-Auflösung und Build-Flags werden nicht angewendet. Die TUI verwendet die Standardauswahl C für `.h`-Dateien.

## Beibehaltung der API-Form

Zum Beispiel Go:

```go
func Pair(a, b int, rest ...string) (string, error)
```

erzeugt:

```text
Pair(a: int, b: int, rest: ...string) -> (string, error)
```

Go-Interface-Methoden und benannte Felder in Go/Rust erscheinen als Symbole sowohl im normalen als auch im Fingerprint-Modus. Methoden und Felder behalten ihren zugehörigen Typ bei. Verschachtelte Rust-Module behalten qualifizierte Namen wie `service.storage.read`, auch wenn ein anderes Modul eine gleichnamige Funktion deklariert.

Rust erlaubt Felder und Methoden mit demselben Namen. Feldbezogene Namen tragen intern ein `#field`-Suffix, sodass ihre IDs und Containment-Kanten unterscheidbar bleiben; Anzeigenamen und Signaturen behalten den ursprünglichen Feldnamen bei.

Python-Signaturen behalten Annotationen, Defaults, Splat-Operatoren und Aufrufkonvention-Trenner bei. Ein konventioneller erster `self`- oder `cls`-Parameter wird bei einer gebundenen Methode weggelassen, bleibt jedoch in einer freien Funktion oder statischen Methode erhalten.

Mehrzeilige Signaturen behalten Zeilenumbrüche im JSON-Format bei. SKT maskiert Zeilenumbrüche als `\n` und `\r`, sodass jedes Symbol auf einer physischen Ausgabezeile bleibt.

## Auflösen von Beziehungen

Beziehungen behalten Empfängerqualifizierung und lexikalischen Besitz bei. Relative Importe und umbenannte JavaScript/TypeScript- und Python-Importe liefern Zielkontext. Go- und Java-Paketdeklarationen können über Dateien im selben Verzeichnis aufgelöst werden. Ein nicht aufgelöster Import fällt niemals auf eine nicht verwandte Deklaration zurück, und konkurrierende Kandidaten bleiben mehrdeutig.

Lokale Empfängerbindungen haben Vorrang vor Namespace-Importen, auch wenn ein Parameter einen importierten Namespace überschattet. Python-Paketimporte berücksichtigen Deklarationen im `__init__`-Modul des Pakets, nicht in beliebigen untergeordneten Modulen. Re-Exporte, die nicht aus diesem Modul festgestellt werden können, bleiben unaufgelöst.

Die Extrahierung von Java-, C++-, C#-, Scala- und TypeScript-Aufrufbaren behält Byte-Bereiche von Deklarationen bei, sodass überladene Körper unterschiedliche Aufrufer und Alias-Bereiche haben können, einschließlich Deklarationen in derselben Zeile. Die Auswahl eines überladenen Ziels bleibt konservativ. Explizite Überschreibungsbeziehungen in Java, C++, C#, Scala und TypeScript suchen nach der nächstgelegenen geerbten Deklaration; fehlende Basen, Zyklen oder konkurrierende Kandidaten behalten Unsicherheit bei, anstatt Selbstverknüpfungen zu erzeugen.

Dies ist eine syntaxbasierte Analyse, kein Compiler oder Laufzeit-Aufrufgraph. Bare JavaScript-Paketspezifikationen, Paket-Re-Exporte, Build-Konfigurationen, komplexe Empfängertypen und dynamische Dispatch können unaufgelöst bleiben. Ein eindeutiger Name an anderer Stelle im Repository ist kein ausreichender Beweis für eine Abhängigkeit.

JSON bewahrt unsichere Kanten mit `resolution: "unresolved"` oder `"ambiguous"` und enthält Kandidaten-IDs, wenn verfügbar. Diese Kanten haben keine Ziel-ShortID.

SKT verwendet dieselbe Kantennotation für alle Beziehungen, mit einem optionalen Auflösungsmarker:

```text
S8 --references--> S5
S10 --references[unresolved]--> string, bool
S11 --calls[ambiguous]--> Helper [candidates=S12, S15]
```

Bekannte Endpunkte und Kandidaten verwenden ShortIDs, einschließlich der Endpunkte unaufgelöster Beziehungen. Bare unsichere Ziele ohne Symbol sind Namen relativ zur Quelldatei. Literale Namen, die wie ShortIDs aussehen, und andere komplexe Identitäten verwenden zitierte vollständige IDs, um Mehrdeutigkeiten zu vermeiden. Kanten mit derselben Quelle, Art, Auflösung und Kandidatenmenge teilen sich eine Zielliste. Der SKT-Leser bewahrt Auflösung und Kandidaten, einschließlich Referenzen auf Symbole in späteren Ausgabechunks. Minimierte Ausgabe kodiert die Beziehungsart, während der Auflösungsmarker lesbar bleibt.

C/C++-Includes haben explizite `meta/file`- und `meta/include`-Symbole. Header-Schreibweisen erscheinen einmal in der Symbolliste; Kanten verwenden ihre ShortIDs:

```text
[symbols]
## src/main.c
S1 meta/file L1-L4 main.c
S2 meta/include L1-L1 "config.h" {resolution=unresolved}
S3 meta/include L2-L2 "utils.h" {resolution=unresolved}
S4 meta/include L3-L3 <stdio.h> {resolution=unresolved}

[edges]
S1 --references[unresolved]--> S2, S3, S4
```

Ein Include-Symbol beschreibt die geschriebene Direktive, nicht eine aufgelöste Header-Deklaration. Wiederholte Includes innerhalb einer Datei teilen sich einen Endpunkt. Datei-/Include-Metadaten sind von Abhängigkeitsrankings und Dead-Code-Findings ausgeschlossen; das Hinzufügen von IDs bedeutet nicht, dass ein Header durch die Include-Suchkonfiguration des Compilers gefunden wurde.

Abhängigkeitsmetriken und HTML-Graph-Links verwenden nur aufgelöste Kanten. Berichte zeigen ausgeschlossene Beziehunganzahlen, und der HTML-Graph zeigt eine Unsicherheitsanzahl an. Fan-in und Fan-out zählen unterschiedliche Nachbarn. Vererbungsanalyse berücksichtigt alle Basen und Interfaces; zyklische Vererbung hat keine gemeldete Tiefe. Erreichbarkeit beginnt bei benannten `main`-, `Main`- oder `init`-Aufrufbaren und folgt Abhängigkeitskanten. Das Erreichen eines Mitglieds macht seinen Container relevant, ohne jedes Geschwisterkind erreichbar zu machen. Ohne erkannte Einstiegspunkte werden Erreichbarkeitsfindings weggelassen.

Callback-Argumente und zurückgegebene Aufrufbare sind Referenzen, kein Beweis für eine Invokation. Alias-Auflösung behält Datei-, Funktions- und Objektidentität bei; widersprüchliche Zuweisungen bleiben unsicher. Graphfindings sind Überprüfungskandidaten: Externe Aufrufer und dynamisches Verhalten können das Ergebnis ändern. Verzeichnisbasierte Schichtregeln und Change-Propagation-Gewichte sind Heuristiken, keine gemessenen Wahrscheinlichkeiten.

## Bekannte Einschränkungen

- Komplexe C/C++-Zeiger-, Referenz-, Array- und Funktionszeiger-Deklaratoren können weiterhin unvollständige Signaturen aufweisen. Makro-generierte Deklarationen werden nicht expandiert.
- Die Handhabung von TypeScript/JavaScript-Konstruktoren und einige optionale, Rest- oder destrukturierte Parameterformen sind unvollständig.
- Fortgeschrittene generische Constraints und Typpformen außerhalb der getesteten Fälle variieren je nach Sprache. Rückgabetypen werden nicht inferiert.
- Rust-Makros werden nicht expandiert. Out-of-Line-Moduldeklarationen werden erfasst; ihre Dateien müssen im Scan enthalten sein. Tuple-Felder, Enum-Variantendetails und assoziierte Elemente haben teilweise Abdeckung.
- Syntaxextraktion löst dynamisches Dispatch, Überladungen, importierte Abhängigkeiten oder Build-Konfigurationssemantik nicht mit Compiler-Präzision auf.
- Teilweise Syntaxfehler können weiterhin Deklarationen auslassen. Überprüfen Sie gemeldete Warnungen und verifizieren Sie den Quellcode, bevor Sie ein Refactoring anwenden.

Vollständigere Extraktion ändert Symbolanzahlen und ShortIDs. Generieren Sie vorhandene Skelettausgabe bei einem Upgrade neu, anstatt sie mit älterer Ausgabe zu mischen.