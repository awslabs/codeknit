---
title: Fingerprint-Befehl
description: Erkennen von Duplikaten und Beinahe-Duplikaten im Code über Dateien und Sprachen hinweg mittels unscharfer Hashing-Verfahren.
---

Der Befehl `codeknit fingerprint` erkennt Duplikate und Beinahe-Duplikate im Code in Ihrer Codebasis unter Verwendung von **Context-Triggered Piecewise Hashing (CTPH)**. Er funktioniert über Dateien und sogar über Programmiersprachen hinweg, indem Variablennamen, String-Literale und Typannotationen normalisiert werden, bevor strukturelle Fingerabdrücke berechnet werden.

## Was er tut

`codeknit fingerprint` analysiert jede Funktion, Methode, Variable und jeden Typ in Ihrer Codebasis und berechnet einen **normalisierten strukturellen Fingerabdruck** basierend auf:

- Kontrollfluss (`if`, `for`, `while`, `switch`)
- Operationen (`=`, `+`, `==`, `&&`, `||`)
- Aufrufe, Rückgaben, Zuweisungen und Objekterstellung
- Sprachkonstrukte wie `try/catch`, `yield`, `await`, `defer`

Diese Normalisierung bedeutet, dass **umbenannte Kopien**, **triviale Refactorings** und **äquivalente Logik in verschiedenen Sprachen** dennoch als Duplikate erkannt werden können.

Der Algorithmus verwendet **CTPH** (eine Variante des Rolling-Hash-Verfahrens), um effizient Beinahe-Duplikate zu finden. Ähnlicher Code erzeugt ähnliche Fingerabdrücke, was unscharfes Matching ermöglicht, selbst wenn der Code leicht verändert wurde.

## Grundlegende Verwendung

```bash
codeknit fingerprint ./src
```

Dieser Befehl:

- Parst alle Quelldateien in `./src`
- Berechnet strukturelle Fingerabdrücke
- Gibt Ergebnisse in `./skeleton/fingerprints.skt` aus
- Meldet Übereinstimmungen mit einer Ähnlichkeit zwischen **65% und 95%** (Standardbereich)

## Flags

| Flag               | Standardwert                  | Beschreibung                                                                                                                                                |
| ------------------ | ----------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `-o`, `--output`   | `./skeleton/fingerprints.skt` | Pfad der Ausgabedatei `.skt`                                                                                                                                |
| `--min-similarity` | `65`                          | Mindestprozentsatz der Ähnlichkeit für die Meldung (0–100)                                                                                                  |
| `--max-similarity` | `95`                          | Höchstprozentsatz der Ähnlichkeit für die Meldung (0–100)                                                                                                   |
| `--show-all`       | `false`                       | Fügt den Abschnitt `[fingerprints]` mit Roh-Tokendaten hinzu                                                                                                |
| `--rerank`         | `false`                       | Findet semantische Nachbarn und ordnet Kandidaten mithilfe von Ollama-Embeddings neu (erfordert: `ollama serve` und `ollama pull qwen3-embedding:0.6b`) |
| `--model`          | `qwen3-embedding:0.6b`        | Ollama-Embedding-Modell, das mit `--rerank` verwendet werden soll                                                                                           |
| `--collect-test`   | `false`                       | Bezieht Testdateien in die Analyse ein                                                                                                                      |
| `--workers`        | `NumCPU`                      | Maximale Anzahl gleichzeitiger Parsing-Goroutines (0 = alle CPU-Kerne verwenden)                                                                           |
| `--verbose`        | `false`                       | Gibt Fortschrittsinformationen während der Verarbeitung aus                                                                                                 |

## Ausgabeformat

Die Ausgabe ist eine `.skt`-Datei mit den folgenden Abschnitten:

### `[duplicates]` (immer vorhanden)

Listet Paare von Symbolen mit einer Ähnlichkeit über dem Schwellenwert auf:

```skt
[duplicates]
similarity:96%  pkg/user.go::GetUser <-> pkg/admin.go::GetAdmin
similarity:88%  utils/str.go::TrimSpaces <-> lib/text.go::CleanString
```

Jede Zeile zeigt:

- Prozentsatz der Ähnlichkeit
- Linkes Symbol (Dateipfad, Gültigkeitsbereich, Name)
- Rechtes Symbol (Dateipfad, Gültigkeitsbereich, Name)

### `[fingerprints]` (nur mit `--show-all`)

Enthält Roh-Fingerabdruckdaten für jedes Symbol:

```skt
[fingerprints]
validateToken  FP:3:a1b2c3...:d4e5f6...  tokens:8e0f1a2b...
```

Felder:

- Symbolname
- `FP:<version>:<hash1>:<hash2>` — CTPH-Fingerabdruck
- `tokens:<hex>` — normalisierter Token-Stream des Körpers

Dieser Abschnitt ist nützlich für Debugging oder den Aufbau nachgelagerter Tools.

## Häufige Muster

```bash
# Standard-Scan
codeknit fingerprint ./codeknit/de/src
```

```bash
# Nur exakte Duplikate finden
codeknit fingerprint ./src --min-similarity 100
```

```bash
# Mäßig ähnlichen Code finden (z. B. gleicher Algorithmus, unterschiedliche Namen)
codeknit fingerprint ./src --min-similarity 50 --max-similarity 80
```

```bash
# Semantisches Matching verwenden, um zusätzliche Kandidaten zu finden und False Positives zu reduzieren
# Erfordert: ollama serve && ollama pull qwen3-embedding:0.6b
codeknit fingerprint ./src --rerank
```

```bash
# Ein anderes Embedding-Modell für semantisches Matching verwenden
codeknit fingerprint ./src --rerank --model qwen3-embedding:4b
```

```bash
# Vollständige Fingerabdruckliste ausgeben (für Analysetools)
codeknit fingerprint ./src --show-all
```

```bash
# Benutzerdefinierte Ausgabedatei
codeknit fingerprint ./src -o duplicates.skt
```

## Auswahl eines Ähnlichkeitsbereichs

| Bereich   | Richtlinie                                                                                 |
| ------- | ---------------------------------------------------------------------------------------- |
| 96–100% | Exakte oder fast exakte strukturelle Duplikate. Fast sicher Kopie-Einfügen.                  |
| 85–95%  | Beinahe-Duplikate. Meist Kopie-Einfügen mit kleinen Änderungen (z. B. umbenannte Variablen, hinzugefügte Logs). |
| 65–84%  | Standardbereich. Starke strukturelle Ähnlichkeit. Gute Kandidaten für Refactoring.            |
| 50–64%  | Mäßige Ähnlichkeit. Gleiche algorithmische Struktur, aber unterschiedliche Details. Manuell prüfen.      |
| < 50%   | Meist Rauschen. Keine bedeutende Duplizierung.                                               |

## Tipps

- **Fingerabdrücke messen Struktur, nicht Bedeutung**: Ein hoher Ähnlichkeitswert bedeutet, dass der Code _ähnlich aussieht_, nicht dass er _dasselbe tut_. Überprüfen Sie immer beide Symbole.
- **Verwenden Sie `--rerank` für semantisches Matching**: Embeddings fügen semantische Nachbarn hinzu, die strukturelle Abfragen übersehen können, und filtern Kandidaten, die semantisch nicht übereinstimmen.
- **Kurze Körper werden übersprungen**: Symbole mit weniger als 4 normalisierten Tokens (z. B. einfache Getter) werden ignoriert, um Rauschen zu vermeiden.
- **Sprachenübergreifendes Matching funktioniert**: Äquivalente Konstrukte (z. B. eine Python-Funktion und eine Go-Funktion mit derselben Logik) können übereinstimmen, aber sprachspezifische Muster können zu falschen Übereinstimmungen mit geringer Ähnlichkeit führen.
- **Eine Übereinstimmung ist ein Signal, kein Urteil**: Behandeln Sie jede Übereinstimmung als Aufforderung zur Untersuchung — nicht als automatischen Beweis für Duplizierung.