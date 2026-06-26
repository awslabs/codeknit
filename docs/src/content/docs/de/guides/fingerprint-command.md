---
title: Fingerprint-Befehl
description: Erkennen von Duplikaten und Beinahe-Duplikaten in Code über Dateien und Sprachen hinweg mithilfe von **fuzzy hashing**.
---

Der `codeknit fingerprint`-Befehl erkennt Duplikate und Beinahe-Duplikate in Ihrem Codebase mithilfe von **Context-Triggered Piecewise Hashing (CTPH)**. Er funktioniert über Dateien und sogar über Programmiersprachen hinweg, indem Variablennamen, String-Literale und Typannotationen normalisiert werden, bevor strukturelle **Fingerprints** berechnet werden.

## Was er macht

`codeknit fingerprint` analysiert jede Funktion, Methode, Variable und jeden Typ in Ihrem Codebase und berechnet einen **normalisierten strukturellen Fingerprint** basierend auf:

- Kontrollfluss (`if`, `for`, `while`, `switch`)
- Operationen (`=`, `+`, `==`, `&&`, `||`)
- Aufrufe, Rückgaben, Zuweisungen und Objekterstellung
- Sprachkonstrukte wie `try/catch`, `yield`, `await`, `defer`

Diese Normalisierung bedeutet, dass **umbenanntes Copy-Paste**, **triviale Refactorings** und **äquivalente Logik in verschiedenen Sprachen** dennoch als Duplikate erkannt werden können.

Der Algorithmus verwendet **CTPH** (eine Variante des Rolling-Hash), um effizient Beinahe-Duplikate zu finden. Ähnlicher Code erzeugt ähnliche Fingerprints, was ein **fuzzy matching** selbst bei leicht modifiziertem Code ermöglicht.

## Grundlegende Verwendung

```bash
codeknit fingerprint ./src
```

Dieser Befehl:

- Parst alle Quelldateien in `./src`
- Berechnet strukturelle Fingerprints
- Gibt Ergebnisse in `./skeleton/fingerprints.skt` aus
- Meldet Übereinstimmungen mit einer **Ähnlichkeit** zwischen **65% und 95%** (Standardbereich)

## Flags

| Flag               | Standardwert                  | Beschreibung                                                                                                                                                |
| ------------------ | ----------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `-o`, `--output`   | `./skeleton/fingerprints.skt` | Pfad der Ausgabedatei `.skt`                                                                                                                                |
| `--min-similarity` | `65`                          | Mindest-**Ähnlichkeit** in Prozent zur Meldung (0–100)                                                                                                      |
| `--max-similarity` | `95`                          | Höchst-**Ähnlichkeit** in Prozent zur Meldung (0–100)                                                                                                       |
| `--show-all`       | `false`                       | Fügt den Abschnitt `[fingerprints]` mit Roh-Tokendaten hinzu                                                                                               |
| `--rerank`         | `false`                       | Ordnet CTPH-Kandidaten mithilfe semantischer Embeddings via Ollama neu, um False Positives zu eliminieren (erfordert: `ollama serve` und `ollama pull qwen3-embedding:0.6b`) |
| `--model`          | `qwen3-embedding:0.6b`        | Ollama-Embedding-Modell für `--rerank`                                                                                                                      |
| `--collect-test`   | `false`                       | Bezieht Testdateien in die Analyse ein                                                                                                                     |
| `--workers`        | `NumCPU`                      | Maximale Anzahl paralleler Parsing-Goroutinen (0 = alle CPU-Kerne nutzen)                                                                                  |
| `--verbose`        | `false`                       | Gibt Fortschrittsinformationen während der Verarbeitung aus                                                                                                |

## Ausgabeformat

Die Ausgabe ist eine `.skt`-Datei mit den folgenden Abschnitten:

### `[duplicates]` (immer vorhanden)

Listet Paare von Symbolen mit einer **Ähnlichkeit** über dem Schwellenwert auf:

```skt
[duplicates]
similarity:96%  pkg/user.go::GetUser <-> pkg/admin.go::GetAdmin
similarity:88%  utils/str.go::TrimSpaces <-> lib/text.go::CleanString
```

Jede Zeile zeigt:

- **Ähnlichkeit** in Prozent
- Linkes Symbol (Dateipfad, Scope, Name)
- Rechtes Symbol (Dateipfad, Scope, Name)

### `[fingerprints]` (nur mit `--show-all`)

Enthält Roh-Fingerprint-Daten für jedes Symbol:

```skt
[fingerprints]
validateToken  FP:3:a1b2c3...:d4e5f6...  tokens:8e0f1a2b...
```

Felder:

- Symbolname
- `FP:<version>:<hash1>:<hash2>` — CTPH-Fingerprint
- `tokens:<hex>` — normalisierter Token-Stream des Körpers

Dieser Abschnitt ist nützlich für Debugging oder den Aufbau nachgelagerter Tools.

## Häufige Muster

```bash
# Standard-Scan
codeknit fingerprint ./src
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
# Semantisches Neuranking verwenden, um False Positives zu reduzieren
# Erfordert: ollama serve && ollama pull qwen3-embedding:0.6b
codeknit fingerprint ./src --rerank
```

```bash
# Ein anderes Embedding-Modell für Neuranking verwenden
codeknit fingerprint ./src --rerank --model qwen3-embedding:4b
```

```bash
# Vollständige Fingerprint-Liste ausgeben (für Analysetools)
codeknit fingerprint ./src --show-all
```

```bash
# Benutzerdefinierte Ausgabedatei
codeknit fingerprint ./src -o duplicates.skt
```

## Auswahl eines Ähnlichkeitsbereichs

| Bereich   | Richtlinie                                                                                 |
| --------- | ------------------------------------------------------------------------------------------ |
| 96–100%   | Exakte oder fast exakte strukturelle Duplikate. Fast sicher Copy-Paste.                   |
| 85–95%    | Beinahe-Duplikate. Meist Copy-Paste mit kleinen Änderungen (z. B. umbenannte Variablen, hinzugefügtes Logging). |
| 65–84%    | Standardbereich. Starke strukturelle **Ähnlichkeit**. Gute Kandidaten für Refactoring.    |
| 50–64%    | Mäßige **Ähnlichkeit**. Gleiche algorithmische Struktur, aber unterschiedliche Details. Manuell prüfen. |
| < 50%     | Meist Rauschen. Keine bedeutende Duplizierung.                                            |

## Tipps

- **Fingerprints messen Struktur, nicht Bedeutung**: Ein hoher **Ähnlichkeit**swert bedeutet, dass der Code _ähnlich aussieht_, nicht dass er _dasselbe tut_. Überprüfen Sie immer beide Symbole.
- **Verwenden Sie `--rerank` bei verrauschten Ergebnissen**: Wenn Sie viele False Positives erhalten, aktivieren Sie semantisches Neuranking, um Übereinstimmungen mithilfe von Embeddings zu filtern.
- **Kurze Körper werden übersprungen**: Symbole mit weniger als 4 normalisierten Tokens (z. B. einfache Getter) werden ignoriert, um Rauschen zu vermeiden.
- **Sprachenübergreifendes Matching funktioniert**: Äquivalente Konstrukte (z. B. eine Python-Funktion und eine Go-Funktion mit derselben Logik) können übereinstimmen, aber sprachspezifische Muster können zu falschen Übereinstimmungen mit geringer **Ähnlichkeit** führen.
- **Eine Übereinstimmung ist ein Signal, kein Urteil**: Behandeln Sie jede Übereinstimmung als Aufforderung zur Untersuchung — nicht als automatischen Beweis für Duplizierung.