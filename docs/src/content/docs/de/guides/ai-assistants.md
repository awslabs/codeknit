---
title: Verwendung mit KI-Assistenten
description: Richten Sie codeknit als Skill für Kiro, Claude Code und andere KI-Coding-Assistenten ein.
---

codeknit wird mit vorgefertigten Skills ausgeliefert, die KI-Coding-Assistenten beibringen, wie sie es effektiv nutzen können. Diese Skills ermöglichen es Assistenten, Codestrukturen zu extrahieren, Duplikate zu erkennen und strukturelle Analysen ohne manuelle Aufforderungen durchzuführen.

## Skills-Übersicht

codeknit bietet zwei Skills:

- **`codeknit-parse`**: Lehrt Assistenten, Codestrukturen (Funktionen, Klassen, Methoden, Variablen) und Beziehungen (Aufrufe, Vererbung, Enthaltensein) in `.skt`-Dateien zu extrahieren.
- **`codeknit-fingerprint`**: Lehrt Assistenten, duplizierten und beinahe-duplizierten Code mithilfe von fuzzy hashing zu erkennen.

Jeder Skill enthält Dokumentation, die der Assistent bei Bedarf liest, um Nutzung, Flags, Ausgabeformate und Workflows zu verstehen.

## Installation

Kopieren Sie die Skill-Verzeichnisse in den Skills-Ordner Ihres Assistenten.

Für **Kiro**:

```bash
cp -r skills/codeknit-parse ~/.kiro/skills/codeknit-parse
cp -r skills/codeknit-fingerprint ~/.kiro/skills/codeknit-fingerprint
```

Für **Claude Code**:

```bash
cp -r skills/codeknit-parse ~/.claude/skills/codeknit-parse
cp -r skills/codeknit-fingerprint ~/.claude/skills/codeknit-fingerprint
```

Nach der Installation weiß der Assistent automatisch, wie er codeknit-Befehle aufrufen, geeignete Flags auswählen und `.skt`-Ausgaben interpretieren kann.

## Was jeder Skill lehrt

### codeknit-parse

Der `codeknit-parse`-Skill lehrt Assistenten:

- `codeknit parse` mit geeigneten Flags für verschiedene Szenarien auszuführen
- Den richtigen Ausgabemodus zu wählen:
  - `directory-flat` (Standard) für die meisten Projekte
  - `inline` für einzelne Dateien oder kleine Eingaben
  - `directory-tree`, um die Quellstruktur widerzuspiegeln
- `.skt`-Ausgabedateien zu lesen und zu interpretieren, einschließlich der Abschnitte `[symbols]`, `[edges]` und optional `[dict]`
- Strukturelle Daten für Refactoring, Abhängigkeitskartierung und Code-Reviews zu nutzen
- `codeknit graph analyze` für tiefere Einblicke in die Codequalität auszuführen (zyklische Abhängigkeiten, Hub-Symbole, god classes usw.)

### codeknit-fingerprint

Der `codeknit-fingerprint`-Skill lehrt Assistenten:

- `codeknit fingerprint` für Duplikaterkennung, DRY-Audits und Refactoring-Identifikation zu verwenden
- Geeignete Ähnlichkeitsbereiche (`--min-similarity`, `--max-similarity`) auszuwählen
- Den Abschnitt `[duplicates]` zu lesen, um beinahe-duplizierten Code zu identifizieren
- Zu verstehen, dass Fingerprints die strukturelle Form und nicht die semantische Absicht messen
- `--rerank` mit Ollama-Embeddings zu verwenden, um falsch-positive Ergebnisse bei Bedarf zu reduzieren

## Workflow-Beispiele

### Strukturelle Analyse

1. Bitten Sie den Assistenten, die Struktur Ihres Codebase zu analysieren
2. Er führt `codeknit parse ./src` aus und liest die resultierenden `.skt`-Dateien
3. Er beantwortet strukturelle Fragen: Abhängigkeiten, Aufrufketten, dead code
4. Für tiefere Einblicke führt er `codeknit graph analyze ./src` aus und interpretiert den Bericht

```skt
[symbols]
## src/service.go
S1 type/struct L5-L8 AuthService {}
S2 callable/method L10-L15 Authenticate(token: string) {receiver=*AuthService}

[edges]
S1 --contains--> S2
```

### Duplikaterkennung

1. Bitten Sie den Assistenten, duplizierten Code zu finden
2. Er führt `codeknit fingerprint ./src` aus
3. Er liest den Abschnitt `[duplicates]` in der Ausgabe
4. Er untersucht die markierten Paare und schlägt Konsolidierungen vor

```skt
[duplicates]
S1, S2: 87% Ähnlichkeit
S3, S4: 76% Ähnlichkeit
```

## Tipps

- **Lesen Sie für strukturelle Fragen immer `.skt`-Dateien, nicht den Rohquellcode** – sie enthalten die extrahierte Struktur in einem kompakten, zuverlässigen Format
- Verwenden Sie `codeknit graph analyze`, um Codequalitätsprobleme wie zyklische Abhängigkeiten, Hub-Symbole und tiefe Vererbungsketten aufzudecken
- Führen Sie `codeknit fingerprint` vor großen Refactorings aus, um kopierten Code zu identifizieren, der konsolidiert werden sollte
- Das `.skt`-Format ist darauf ausgelegt, token-effizient zu sein, was es ideal für LLM-Kontextfenster macht
- Verwenden Sie `--minify`, um den Token-Verbrauch bei der Verarbeitung großer Codebasen weiter zu reduzieren
