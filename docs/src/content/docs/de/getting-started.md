---
title: Erste Schritte
description: Legen Sie mit codeknit in unter 5 Minuten los.
---

In diesem Tutorial werden Sie codeknit aus dem Quellcode erstellen, Ihren ersten Parse-Vorgang durchführen und die von ihm geschriebenen `.skt`-Ausgaben lesen. Sie werden ein kleines Paket aus dem codeknit-Repository parsen, sodass Sie kein eigenes Projekt benötigen.

## Voraussetzungen

- **Go 1.26 oder höher.** Führen Sie `go version` aus, um dies zu überprüfen.
- **Ein C-Compiler.** codeknit verwendet tree-sitter, das CGo benötigt.
- **Git** und **Make**.

## 1. codeknit erstellen

Klone Sie das Repository und erstellen Sie es:

```bash
git clone https://github.com/awslabs/codeknit.git
cd codeknit
make build
```

Sie sollten sehen, dass der Build ohne Fehler abgeschlossen wird und eine Binary unter `./bin/codeknit` erstellt wird.

## 2. codeknit zu Ihrem PATH hinzufügen

Fügen Sie vom `codeknit`-Verzeichnis aus `bin` zu Ihrem `PATH` hinzu, indem Sie die Zeile für Ihre Shell verwenden.

Für bash oder zsh:

```bash
export PATH="$PATH:$(pwd)/bin"
```

Für fish:

```fish
fish_add_path (pwd)/bin
```

Dies gilt für Ihre aktuelle bash- oder zsh-Sitzung. Um es für spätere Sitzungen beizubehalten, fügen Sie dieselbe Zeile zu `~/.bashrc` oder `~/.zshrc` hinzu, wobei der absolute Pfad zu `bin` anstelle von `$(pwd)/bin` verwendet wird.

## 3. Installation überprüfen

```bash
codeknit --version
```

Sie sollten die codeknit-Version angezeigt bekommen.

## 4. Ersten Parse-Vorgang durchführen

Bleiben Sie im `codeknit`-Verzeichnis und parsen Sie das Paket `internal/console`:

```bash
codeknit parse ./internal/console --edges
```

codeknit durchsucht das Verzeichnis, parst jede unterstützte Quelldatei und schreibt `.skt`-Dateien. Eine *`.skt`-Datei* ist eine kompakte Textdatei, die die Symbole in Ihrem Code und die Beziehungen zwischen ihnen auflistet. Das Flag `--edges` fügt die Beziehungen hinzu. Diese werden standardmäßig weggelassen, um Platz zu sparen.

Sie sollten eine kurze Zusammenfassung sehen, wenn der Befehl abgeschlossen ist, sowie ein neues Verzeichnis `./skeleton` in Ihrem aktuellen Verzeichnis.

Listen Sie die Ausgabe auf:

```bash
ls ./skeleton
```

Sie sollten eine Datei wie `map_001.skt` sehen.

## 5. Ausgabe lesen

Öffnen Sie die Datei:

```bash
cat ./skeleton/map_001.skt
```

Sie sollten zuerst einen Abschnitt `[symbols]` sehen. Ein *Symbol* ist ein benanntes Element in Ihrem Code, wie z. B. eine Funktion, ein Typ oder ein Feld. Symbole sind unter einer Überschrift `## <Quelldatei>` gruppiert. Ihre Ausgabe enthält Zeilen wie diese:

```skt
[symbols]
## <Pfad zu>/console.go
S1 module/package L5-L5 console
S2 type/struct L17-L25 Console {exported}
...
S10 callable/function L28-L41 New() -> *S2 {exported}
...
S12 callable/method L49-L49 Error(msg: string) {exported, receiver=*Console}
```

Jede Zeile hat dieselben Teile:

| Teil | Beispiel | Bedeutung |
| --- | --- | --- |
| ShortID | `S10` | Eine kurze Bezeichnung für das Symbol, die an anderer Stelle in der Ausgabe verwendet wird |
| Kategorie und Art | `callable/function` | Um welche Art von Symbol es sich handelt |
| Zeilenbereich | `L28-L41` | Die Zeilen, die das Symbol in der Quelldatei umfasst |
| Signatur | `New() -> *S2` | Der Name, die Parameter und der Rückgabetyp |
| Eigenschaften | `{exported}` | Zusätzliche Details in Klammern |

In `New() -> *S2` ist `S2` die ShortID des Typs `Console`. codeknit ersetzt Typnamen durch ShortIDs, wenn es das Symbol kennt. Zeilennummern können leicht abweichen, wenn Ihr Checkout neuer ist.

Scrollen Sie nach unten zum Abschnitt `[edges]`. Eine *Kante* ist eine Beziehung zwischen zwei Symbolen. Jede Kantenzeile nennt das Quellsymbol, die Art der Beziehung und das Zielsymbol oder die Zielsymbole:

```skt
[edges]
S2 --contains--> S4
S4 --calls--> S5
```

Dieses Beispiel zeigt die Struktur einer Kantenzeile. Es bedeutet "`S2` enthält `S4`" und "`S4` ruft `S5` auf". Sie können jede ShortID im Abschnitt `[symbols]` nachschlagen, um zu sehen, worauf sie sich bezieht.

## Was Sie erreicht haben

Sie haben codeknit erstellt, es zu Ihrem `PATH` hinzugefügt, `codeknit parse` ausgeführt und die Abschnitte `[symbols]` und `[edges]` der `.skt`-Ausgabe gelesen.

## Nächste Schritte

- Parsen Sie Ihr eigenes Projekt, wählen Sie ein anderes Ausgabe-Layout und führen Sie es sicher erneut aus: siehe den [Leitfaden zum Parse-Befehl](../guides/parse-command/).
- Visualisieren und analysieren Sie Ihre Codestruktur: siehe den [Leitfaden zu Graph-Befehlen](../guides/graph-commands/).
- Finden Sie Duplikate und Beinahe-Duplikate: siehe den [Leitfaden zum Fingerprint-Befehl](../guides/fingerprint-command/).
- Schlagen Sie jedes Detail des `.skt`-Formats nach: siehe die [Referenz zum Ausgabeformat](../reference/output-format/).
- Schlagen Sie jedes Befehlszeilen-Flag nach: siehe die [Referenz zu CLI-Flags](../reference/cli-flags/).