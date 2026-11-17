---
title: Fingerprint-Befehl
description: Erkennen von Duplikaten und Beinahe-Duplikaten in Code über Dateien und Sprachen hinweg mithilfe von fuzzy Hashing.
---

`codeknit fingerprint` berechnet einen Context-Triggered Piecewise Hash (CTPH) über die normalisierten Body-Tokens jeder Funktion, Methode, Variable und jedes Typs. Variablennamen, String-Literale und Typannotationen werden vor dem Hashing normalisiert. Umbenannte Kopien und Portierungen in andere Sprachen können daher trotzdem übereinstimmen.

Diese Anleitung zeigt, wie Duplikate gefunden, der Ähnlichkeitsbereich angepasst, False Positives reduziert und entschieden wird, was refaktorisiert werden soll.

## Bevor Sie beginnen

- `codeknit` ist installiert. Siehe [Installation](../../installation/).
- Sie haben ein Quellverzeichnis zum Scannen, z. B. `./src`.

## Duplikate in einem Projekt finden

Führen Sie den Befehl auf dem Verzeichnis aus, das Sie überprüfen möchten:

```bash
codeknit fingerprint ./src
```

Der Befehl schreibt `./skeleton/fingerprints.skt` und meldet Paare mit einer Ähnlichkeit von 65 % bis 95 %.

Öffnen Sie die Datei. Der Abschnitt `[duplicates]` ist immer vorhanden, mit den stärksten Übereinstimmungen zuerst:

```skt
[duplicates]
# similarity range: 65%-95%
similarity:93%  pkg/user.go::GetUser <-> pkg/admin.go::GetAdmin
similarity:88%  src/auth.ts::validateToken <-> src/session.ts::validateSession
similarity:81%  lib/math.py::compute_tax <-> lib/math.py::compute_vat
```

Jede Zeile hat das Format `similarity:<N>%  <file>::<symbol> <-> <file>::<symbol>`. Wenn nichts in den Bereich fällt, enthält der Abschnitt nur `# no duplicates found`.

Für die vollständige Ausgabesyntax siehe die [Referenz zum Ausgabemodus](../../reference/output-format/).

## Nur exakte Duplikate finden

Der Standardbereich endet bei 95 %, sodass 100 %-Übereinstimmungen ausgeschlossen sind. Um exakte Übereinstimmungen aufzulisten, setzen Sie beide Grenzen auf 100:

```bash
codeknit fingerprint ./src --min-similarity 100 --max-similarity 100
```

„Exakt“ bedeutet, dass die vollständigen normalisierten Token-Streams innerhalb derselben Symbolkategorie identisch sind. Es bedeutet nicht, dass der Quelltext identisch ist oder dass die beiden Symbole nachweislich äquivalent sind.

## Den Bereich erweitern, um lockerere Übereinstimmungen zu finden

Senken Sie das Minimum, um Code einzuschließen, der eine algorithmische Struktur teilt, sich aber in Details unterscheidet:

```bash
codeknit fingerprint ./src --min-similarity 50 --max-similarity 80
```

Ergebnisse unter 50 % sind normalerweise Rauschen.

## False Positives mit semantischem Reranking reduzieren

Wenn die strukturellen Ergebnisse verrauscht sind, ordnen Sie Kandidaten mit Embeddings von einem lokalen Ollama-Server neu.

1. Starten Sie Ollama und laden Sie das Standard-Embedding-Modell herunter:

   ```bash
   ollama serve
   ollama pull qwen3-embedding:0.6b
   ```

2. Führen Sie den Befehl mit `--rerank` aus:

   ```bash
   codeknit fingerprint ./src --rerank
   ```

Um ein anderes Embedding-Modell zu verwenden, übergeben Sie `--model`:

```bash
codeknit fingerprint ./src --rerank --model qwen3-embedding:4b
```

Reranking fügt semantische Kandidaten hinzu und filtert basierend auf einem kombinierten Score. Es reduziert False Positives, entfernt aber nicht alle. Es kann auch einen Kandidaten entfernen, der strukturell exakt ist, und der Ähnlichkeitsbereich gilt weiterhin. Embeddings werden aus demselben geparsten Quell-Snapshot berechnet, der für die Extraktion verwendet wird.

## Die rohen Fingerprints inspizieren

Um den Fingerprint jedes Symbols zu sehen, fügen Sie `--show-all` hinzu:

```bash
codeknit fingerprint ./src --show-all
```

Dies fügt einen Abschnitt `[fingerprints]` hinzu, gruppiert nach Quelldatei, eine Zeile pro Symbol:

```skt
[fingerprints]
## src/auth.ts
validateToken  FP:3:a1b2c3...:d4e5f6...  tokens:8e0f1a2b...
```

Das Zeilenformat ist `<symbol-name>  FP:<blocksize>:<hash1>:<hash2>  tokens:<hex>`. `tokens` ist der hex-kodierte normalisierte Token-Stream.

## Den Bericht an einem anderen Ort speichern

```bash
codeknit fingerprint ./src -o duplicates.skt
```

## Steuern, welche Dateien gescannt werden

Testdateien werden standardmäßig übersprungen. Fügen Sie sie mit `--collect-test` hinzu:

```bash
codeknit fingerprint ./src --collect-test
```

Um den Scan einzugrenzen, verwenden Sie die Flags zur Quellenauswahl. Setzen Sie die Muster in Anführungszeichen, damit Ihre Shell sie nicht expandiert:

```bash
codeknit fingerprint ./src --include '**/*.ts' --exclude '**/generated/'
```

Weitere Steuerungsmöglichkeiten:

- `--max-file-size BYTES` überspringt größere Dateien. Der Standardwert ist `10485760` (10 MiB). `0` deaktiviert das Limit.
- `--strict` bricht vor dem Schreiben der Ausgabe ab, wenn der Scan oder das Parsen unvollständig ist.
- `--header-language c|cpp` wählt den Parser für mehrdeutige `.h`-Dateien aus. Siehe [Sprachunterstützung](../../reference/language-support/#header-language-selection).

Übersprungene Symlinks, zu große Dateien und Syntaxwarnungen erscheinen auf stderr. Includes überschreiben `.gitignore` oder Testfilter nicht, und Excludes haben Vorrang vor Includes. Siehe die [Referenz zur Analysezuverlässigkeit](../../reference/analysis-reliability/) für genaue Garantien.

## Einen Ähnlichkeitsbereich wählen

| Bereich    | Was normalerweise gefunden wird                                      |
| ---------- | -------------------------------------------------------------------- |
| `96-100%`  | Exakte oder fast exakte strukturelle Duplikate; klare Kopien        |
| `85-95%`   | Beinahe-Duplikate; normalerweise Kopien mit trivialen Änderungen     |
| `65-84%`   | Standardbereich; starke strukturelle Ähnlichkeit, oft Refaktor-Kandidaten |
| `50-64%`   | Gleiche algorithmische Struktur mit unterschiedlichen Details       |
| Unter 50%  | Normalerweise Rauschen                                              |

Um die ersten beiden Zeilen zu sehen, erhöhen Sie `--max-similarity` auf `100`.

## Entscheiden, was refaktorisiert werden soll

1. Beginnen Sie mit dem Standardbereich und lesen Sie die obersten Paare.
2. Öffnen Sie beide Symbole jedes Paares und vergleichen Sie sie.
3. Behandeln Sie eine Übereinstimmung als Signal, nicht als Beweis. Fingerprints messen die strukturelle Form, nicht die Bezeichnernamen oder die Absicht. Dieselbe Logik, iterativ und rekursiv geschrieben, erzielt eine niedrige Punktzahl, und nicht verwandter Code mit derselben Form kann eine hohe Punktzahl erzielen.
4. Paare im Bereich von 85–95 % sind normalerweise die sichersten Kandidaten zur Konsolidierung. Paare im Bereich von 65–84 % erfordern oft einen genaueren Blick.
5. Wenn die Liste verrauscht ist, verengen Sie den Bereich oder fügen Sie `--rerank` hinzu.

Sprachübergreifende Übereinstimmungen funktionieren für äquivalente Konstrukte. Sprachspezifische Merkmale können zu falschen Übereinstimmungen mit niedriger Ähnlichkeit führen.

Ähnlichkeitsprozente sind eine heuristische Rangfolge, keine Editierdistanz. Das beschriftete Kalibrierungs-Fixture des Projekts meldet Präzision und Recall separat von den Ähnlichkeitswerten. Es handelt sich um eine Regressionsprüfung, nicht um eine Schätzung der Genauigkeit in Ihrem Code.

## Große Projekte

- Bei weniger als 500 berechtigten Symbolen bewertet codeknit jedes Paar derselben Kategorie.
- Ab 500 oder mehr werden vollständige exakte normalisierte-Token-Gruppen immer beibehalten. Andere Paare verwenden ungefähre Top-50-Nachbarn pro Symbol, sodass Ergebnisse für Beinahe-Duplikate unvollständig sein können.
- Für exakte Duplikate verwenden Sie den Befehl 100/100 oben. Er ist nicht von der Approximation betroffen.
- Die Sortierung ist deterministisch für stabile Eingaben und Bewertungen.

## Fehlerbehebung

| Symptom | Ursache und Lösung |
| ------- | ------------------ |
| `# no duplicates found` | Kein Paar fällt in den Bereich. Senken Sie `--min-similarity` oder erhöhen Sie `--max-similarity`. |
| Eine bekannte kurze Funktion fehlt | Symbole mit weniger als 4 normalisierten Tokens werden übersprungen. |
| Exakte Kopien fehlen | Das Standardmaximum beträgt 95 %. Verwenden Sie `--min-similarity 100 --max-similarity 100`. Mit `--rerank` kann der semantische Filter sie ebenfalls entfernen. |
| `--rerank` schlägt fehl | Überprüfen Sie, ob `ollama serve` läuft und das Modell heruntergeladen wurde. |
| Teilweise Ergebnisse mit Warnungen auf stderr | Beheben Sie die gemeldeten Dateien oder fügen Sie `--strict` hinzu, um stattdessen einen Fehler zu verursachen. |

## In dieser Anleitung verwendete Flags

| Flag               | Standardwert                  | Beschreibung                                                                 |
| ------------------ | ----------------------------- | --------------------------------------------------------------------------- |
| `-o`, `--output`   | `./skeleton/fingerprints.skt` | Pfad zur `.skt`-Ausgabedatei.                                               |
| `--min-similarity` | `65`                          | Mindestprozentsatz der Ähnlichkeit zur Meldung (0–100).                     |
| `--max-similarity` | `95`                          | Höchstprozentsatz der Ähnlichkeit zur Meldung (0–100).                      |
| `--show-all`       | `false`                       | Gibt auch den Abschnitt `[fingerprints]` mit rohen Tokendaten aus.         |
| `--rerank`         | `false`                       | Ordnet Kandidaten mit semantischen Embeddings via Ollama neu. Erfordert `ollama serve` und `ollama pull qwen3-embedding:0.6b`. |
| `--model`          | `qwen3-embedding:0.6b`        | Ollama-Embedding-Modell zur Verwendung mit `--rerank`.                     |
| `--collect-test`   | `false`                       | Fügt Testdateien in die Analyse ein.                                        |
| `--workers`        | `NumCPU`                      | Maximale Anzahl gleichzeitiger Parsing-Goroutinen.                         |
| `--verbose`        | `false`                       | Gibt Fortschrittsinformationen aus.                                         |

Für die vollständige Liste, einschließlich `--include`, `--exclude`, `--max-file-size`, `--strict` und `--header-language`, siehe die [Referenz zu CLI-Flags](../../reference/cli-flags/).

## Verwandte Themen

- [Parse-Befehl](../parse-command/)
- [Graph-Befehle](../graph-commands/)