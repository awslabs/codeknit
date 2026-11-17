---
title: Verwendung mit KI-Assistenten
description: Installieren Sie das codeknit-Plugin für Claude Code, Codex und Kiro.
---

Installieren Sie das codeknit-Plugin in Claude Code, Codex oder Kiro, halten Sie es aktuell und stellen Sie dem Assistenten strukturelle Fragen sowie Fragen zu Duplikat-Code. Das Plugin enthält eine Fähigkeit, `codeknit`. Die Fähigkeit teilt dem Assistenten mit, wann `codeknit parse`, `codeknit graph analyze` und `codeknit fingerprint` ausgeführt werden sollen und wie die `.skt`-Ausgabe zu lesen ist.

Sie müssen das `codeknit`-Binary nicht selbst installieren. Beim ersten Ausführen der Fähigkeit führt es das gebündelte `scripts/install-codeknit.sh` aus, das Folgendes tut:

1. Beendet sich sofort, wenn `codeknit` bereits auf Ihrem `PATH` oder in `~/.local/bin` vorhanden ist.
2. Lädt andernfalls das neueste GitHub-Release für Ihr Betriebssystem und Ihre Architektur herunter.
3. Überprüft den Download anhand der Release-Datei `checksums.txt`.
4. Installiert das Binary nach `~/.local/bin` und gibt den absoluten Pfad in der letzten Zeile der Ausgabe aus.

Falls `~/.local/bin` nicht in Ihrem `PATH` enthalten ist, ruft der Assistent codeknit für den Rest der Sitzung über den vollständigen Pfad auf.

Um die Installation zu steuern, setzen Sie diese Umgebungsvariablen:

| Variable | Wirkung |
| --- | --- |
| `CODEKNIT_VERSION` | Legt eine Release-Version fest, z. B. `0.6.0`. |
| `CODEKNIT_INSTALL_DIR` | Ändert das Installationsverzeichnis. |

Um das Binary stattdessen selbst zu installieren, siehe [Installation](../../installation/).

## Plugin installieren

Wählen Sie den Abschnitt für Ihren Assistenten aus.

### Claude Code

Führen Sie diese Befehle in Ihrer Shell aus. Innerhalb einer Claude-Code-Sitzung führen Sie dieselben Befehle als `/plugin ...` aus.

```bash
claude plugin marketplace add awslabs/codeknit --sparse .claude-plugin skills
claude plugin install codeknit@codeknit
```

### Codex

```bash
codex plugin marketplace add awslabs/codeknit --sparse .agents --sparse .codex-plugin --sparse skills
codex plugin add codeknit@codeknit
```

### Kiro IDE

1. Öffnen Sie das Powers-Panel.
2. Wählen Sie **Add Custom Power** und dann **Import power from GitHub**.
3. Geben Sie `https://github.com/awslabs/codeknit` ein.

### Kiro CLI

Die Kiro-CLI installiert von einem lokalen Pfad, daher klonen Sie zunächst das Repository:

```bash
git clone --depth 1 --filter=blob:none --sparse https://github.com/awslabs/codeknit ~/.codeknit-power
git -C ~/.codeknit-power sparse-checkout set skills
kiro-cli powers install ~/.codeknit-power
```

### Zu `--sparse`

Die Option `--sparse` ist optional. Sie beschränkt das Klonen auf die Plugin-Manifeste und `skills/` anstelle des gesamten Quellbaums.

### Fähigkeit manuell installieren

Falls Ihr Assistent Agent Skills liest, aber keine Plugin-Unterstützung bietet, kopieren Sie `skills/codeknit/` in das Fähigkeiten-Verzeichnis Ihres Assistenten:

- `~/.claude/skills/`
- `~/.codex/skills/`
- `~/.kiro/skills/`

## Plugin aktualisieren

| Assistent | Befehl oder Aktion |
| --- | --- |
| Claude Code | `claude plugin marketplace update codeknit` |
| Codex | `codex plugin marketplace upgrade` |
| Kiro IDE | Wählen Sie **Check for updates** im Powers-Panel. |
| Kiro CLI | Führen Sie `git -C ~/.codeknit-power pull` aus und dann erneut `kiro-cli powers install ~/.codeknit-power`. |

Das Aktualisieren des Plugins aktualisiert kein bereits installiertes Binary. Um die Binary-Version zu ändern, setzen Sie `CODEKNIT_VERSION` wie oben beschrieben.

Falls Sie codeknit aktualisieren und sich die Extraktion ändert, können sich auch Symbol-IDs ändern. Bitten Sie den Assistenten, die gesamte Ausgabe neu zu generieren, anstatt Teile aus verschiedenen Läufen zu mischen.

## Was die Fähigkeit dem Assistenten beibringt

Die Fähigkeit gibt dem Assistenten diese Regeln vor, sodass Sie sie nicht in jeder Anfrage wiederholen müssen:

- **Wählen Sie den Befehl anhand der Frage.**
  - `parse` beantwortet, was existiert, wo es sich befindet und was was aufruft, erbt oder enthält.
  - `graph analyze` findet Zyklen, Hubs, toten Code, God Classes und Engpässe.
  - `fingerprint` findet Kopier-Einfüge- und Refactoring-Kandidaten.
- **Standardmäßig `directory-flat`-Ausgabe verwenden.** Der Assistent schreibt Dateiteile nach `./skeleton/` und liest nur die benötigten Teile. Er verwendet die `inline`-Ausgabe nur für eine einzelne Datei oder eine Handvoll Dateien, da `inline` alles in das Kontextfenster schreibt.
- **`.skt`-Ausgabe lesen, nicht den Roh-Quellcode, für strukturelle Fragen.** Der Assistent öffnet Quelldateien nur, um Implementierungsdetails zu prüfen oder zu ändern, wobei er die Zeilenbereiche aus der Ausgabe verwendet. Die Datei `references/output-format.md` der Fähigkeit enthält das vollständige Format. Siehe auch die [Referenz zum Ausgabeformat](../../reference/output-format/).
- **Ähnlichkeitsbereiche auswählen und Treffer als Signale behandeln.** Fingerprints vergleichen die strukturelle Form, nicht die Absicht. Der Assistent liest beide Symbole eines Paares, bevor er ein Refactoring vorschlägt.

## Strukturanalyse anfordern

1. Bitten Sie den Assistenten, die Struktur Ihres Codebase zu analysieren. Beispiel: "Kartografiere die Struktur von `./src` und teile mir mit, welche Module vom Auth-Paket abhängen."
2. Der Assistent stellt sicher, dass das Binary installiert ist, führt dann `codeknit parse` mit aktivierten Kanten aus und liest die `.skt`-Ausgabe.
3. Stellen Sie Folgefragen zu Abhängigkeiten, Aufrufketten oder totem Code. Der Assistent antwortet anhand der Ausgabe.
4. Für eine tiefere Analyse bitten Sie um eine strukturelle Überprüfung. Beispiel: "Finde zyklische Abhängigkeiten und God Classes in `./src`." Der Assistent führt `codeknit graph analyze` aus und liest den Bericht.

Beziehungen, die in der Ausgabe als `[unresolved]` oder `[ambiguous]` markiert sind, sind unsicher und keine bestätigten Abhängigkeiten. Behandeln Sie Antworten, die darauf basieren, als Kandidaten zur Überprüfung.

Für die Befehle hinter diesen Anfragen siehe [Anleitung zum Parse-Befehl](../parse-command/) und [Anleitung zu den Graph-Befehlen](../graph-commands/).

## Duplikaterkennung anfordern

1. Bitten Sie den Assistenten, duplizierten Code zu finden. Beispiel: "Finde kopierte Funktionen in `./src`."
2. Der Assistent führt `codeknit fingerprint` aus und liest den Abschnitt `[duplicates]` von `./skeleton/fingerprints.skt`.
3. Der Assistent öffnet beide Symbole jedes markierten Paares und schlägt dann ein Refactoring vor.
4. Falls die Ergebnisse zu viele falsche Treffer enthalten, bitten Sie ihn, den Ähnlichkeitsbereich anzupassen oder eine semantische Neubewertung hinzuzufügen. Die Neubewertung erfordert einen lokalen Ollama-Server mit einem eingebetteten Modell.

Für die Optionen hinter diesen Anfragen siehe [Anleitung zum Fingerprint-Befehl](../fingerprint-command/).

## Tipps

- Erinnern Sie den Assistenten daran, `.skt`-Dateien und nicht den Roh-Quellcode für strukturelle Fragen zu lesen. Dies hält den Kontext klein, da das Format kompakt ist und Symbole, Beziehungen und Zeilenbereiche enthält.
- Verwenden Sie `graph analyze` für Erkenntnisse zur Codequalität. Ergebnisse sind Überprüfungskandidaten, da externe Aufrufer und dynamische Dispatches das Ergebnis beeinflussen können.
- Führen Sie vor einem großen Refactoring einen Fingerprint durch, um Kopier-Einfüge-Code zu erkennen, der konsolidiert werden sollte.
- Falls der Assistent `parse` in ein Verzeichnis ausführt, das bereits `.skt`-Ausgabe enthält, verweigert codeknit das Schreiben. Bitten Sie ihn, `--clean` zu übergeben, um die frühere Ausgabe zu ersetzen.
- Falls der Binary-Download fehlschlägt (offline oder nicht unterstützte Plattform), holen Sie sich manuell ein Release von der [Releases-Seite](https://github.com/awslabs/codeknit/releases) oder siehe [Installation](../../installation/).
- Falls ein Lauf Teilergebnisse meldet, überprüfen Sie stderr und den Abschnitt `[errors]`. Weitere Details finden Sie unter [Analysezuverlässigkeit](../../reference/analysis-reliability/).