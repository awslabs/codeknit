---
title: Comando Fingerprint
description: Rileva codice duplicato e quasi duplicato tra file e linguaggi utilizzando fuzzy hashing.
---

`codeknit fingerprint` calcola un Context-Triggered Piecewise Hash (CTPH) sui token normalizzati del corpo di ogni funzione, metodo, variabile e tipo. I nomi delle variabili, i letterali stringa e le annotazioni dei tipi vengono normalizzati prima dell'hashing. Pertanto, copie-incolla rinominate e porting su altri linguaggi possono comunque corrispondere.

Questa guida mostra come trovare duplicati, regolare l'intervallo di similarità, ridurre i falsi positivi e decidere cosa rifattorizzare.

## Prima di iniziare

- `codeknit` è installato. Vedi [Installazione](../../installation/).
- Hai una directory sorgente da scansionare, come `./src`.

## Trova duplicati in un progetto

Esegui il comando sulla directory che desideri controllare:

```bash
codeknit fingerprint ./src
```

Il comando scrive `./skeleton/fingerprints.skt` e riporta coppie con similarità dal 65% al 95%.

Apri il file. La sezione `[duplicates]` è sempre presente, con le corrispondenze più forti per prime:

```skt
[duplicates]
# similarity range: 65%-95%
similarity:93%  pkg/user.go::GetUser <-> pkg/admin.go::GetAdmin
similarity:88%  src/auth.ts::validateToken <-> src/session.ts::validateSession
similarity:81%  lib/math.py::compute_tax <-> lib/math.py::compute_vat
```

Ogni riga ha la forma `similarity:<N>%  <file>::<symbol> <-> <file>::<symbol>`. Se nulla rientra nell'intervallo, la sezione contiene solo `# no duplicates found`.

Per la grammatica completa dell'output, consulta il [riferimento del formato di output](../../reference/output-format/).

## Trova solo duplicati esatti

L'intervallo predefinito arriva fino al 95%, quindi le corrispondenze al 100% sono escluse. Per elencare le corrispondenze esatte, imposta entrambi i limiti a 100:

```bash
codeknit fingerprint ./src --min-similarity 100 --max-similarity 100
```

"Esatto" significa che i flussi di token normalizzati completi sono identici all'interno della stessa categoria di simboli. Non significa che il testo sorgente sia identico, o che i due simboli siano equivalenti provati.

## Amplia l'intervallo per trovare corrispondenze più lasche

Abbassa il minimo per includere codice che condivide una forma algoritmica ma differisce nei dettagli:

```bash
codeknit fingerprint ./src --min-similarity 50 --max-similarity 80
```

I risultati sotto il 50% sono solitamente rumore.

## Ridurre i falsi positivi con il reranking semantico

Se i risultati strutturali sono rumorosi, riordina i candidati con embeddings da un server Ollama locale.

1. Avvia Ollama e scarica il modello di embedding predefinito:

   ```bash
   ollama serve
   ollama pull qwen3-embedding:0.6b
   ```

2. Esegui il comando con `--rerank`:

   ```bash
   codeknit fingerprint ./src --rerank
   ```

Per utilizzare un altro modello di embedding, passa `--model`:

```bash
codeknit fingerprint ./src --rerank --model qwen3-embedding:4b
```

Il reranking aggiunge candidati semantici e filtra in base a un punteggio combinato. Riduce i falsi positivi ma non li elimina tutti. Può anche rimuovere un candidato che è strutturalmente esatto, e l'intervallo di similarità rimane valido. Gli embeddings sono calcolati dallo stesso snapshot del sorgente analizzato utilizzato per l'estrazione.

## Ispeziona le fingerprint grezze

Per vedere la fingerprint di ogni simbolo, aggiungi `--show-all`:

```bash
codeknit fingerprint ./src --show-all
```

Questo aggiunge una sezione `[fingerprints]`, raggruppata per file sorgente, una riga per simbolo:

```skt
[fingerprints]
## src/auth.ts
validateToken  FP:3:a1b2c3...:d4e5f6...  tokens:8e0f1a2b...
```

Il formato della riga è `<symbol-name>  FP:<blocksize>:<hash1>:<hash2>  tokens:<hex>`. `tokens` è il flusso di token normalizzato codificato in esadecimale.

## Scrivi il report altrove

```bash
codeknit fingerprint ./src -o duplicates.skt
```

## Controlla quali file vengono scansionati

I file di test vengono saltati per impostazione predefinita. Includili con `--collect-test`:

```bash
codeknit fingerprint ./src --collect-test
```

Per restringere la scansione, usa i flag di selezione del sorgente. Racchiudi i pattern tra virgolette in modo che la shell non li espanda:

```bash
codeknit fingerprint ./src --include '**/*.ts' --exclude '**/generated/'
```

Altri controlli di selezione:

- `--max-file-size BYTES` salta i file più grandi. Il valore predefinito è `10485760` (10 MiB). `0` disabilita il limite.
- `--strict` fallisce prima di scrivere l'output se la scansione o l'analisi è incompleta.
- `--header-language c|cpp` seleziona il parser per i file `.h` ambigui. Vedi [Supporto linguaggi](../../reference/language-support/#header-language-selection).

I symlink saltati, i file troppo grandi e gli avvisi di sintassi appaiono su stderr. Gli include non sovrascrivono `.gitignore` o il filtro dei test, e gli exclude hanno la precedenza sugli include. Per garanzie esatte, consulta il [riferimento sull'affidabilità dell'analisi](../../reference/analysis-reliability/).

## Scegli un intervallo di similarità

| Intervallo  | Cosa solitamente rivela                                      |
| ----------- | ------------------------------------------------------------ |
| `96-100%`   | Duplicati strutturali esatti o quasi esatti; chiaro copia-incolla |
| `85-95%`    | Quasi duplicati; solitamente copia-incolla con modifiche banali |
| `65-84%`    | Intervallo predefinito; forte similarità strutturale, spesso candidati per la rifattorizzazione |
| `50-64%`    | Stessa forma algoritmica con dettagli diversi                |
| Sotto 50%   | Solitamente rumore                                           |

Per vedere le prime due righe, alza `--max-similarity` a `100`.

## Decidi cosa rifattorizzare

1. Inizia con l'intervallo predefinito e leggi le coppie principali.
2. Apri entrambi i simboli di ogni coppia e confrontali.
3. Considera una corrispondenza come un segnale, non come una prova. Le fingerprint misurano la forma strutturale, non i nomi degli identificatori o l'intento. La stessa logica scritta in modo iterativo e ricorsivo ottiene un punteggio basso, e codice non correlato con la stessa forma può ottenere un punteggio alto.
4. Le coppie nell'intervallo 85–95% sono solitamente i candidati più sicuri da consolidare. Le coppie nell'intervallo 65–84% spesso richiedono un'analisi più approfondita.
5. Se l'elenco è rumoroso, restringi l'intervallo o aggiungi `--rerank`.

Le corrispondenze tra linguaggi funzionano per costrutti equivalenti. Le funzionalità specifiche del linguaggio possono produrre corrispondenze spurie a bassa similarità.

Le percentuali di similarità sono un ranking euristico, non una distanza di editing. Il progetto di calibrazione etichettato riporta precisione e recall separatamente dai punteggi di similarità. È un controllo di regressione, non una stima dell'accuratezza sul tuo codice.

## Progetti di grandi dimensioni

- Sotto i 500 simboli idonei, codeknit valuta ogni coppia della stessa categoria.
- A 500 o più, i gruppi completi di token normalizzati esatti vengono sempre mantenuti. Altre coppie utilizzano i 50 vicini approssimativi per simbolo, quindi i risultati di quasi duplicati possono essere incompleti.
- Per i duplicati esatti, usa il comando 100/100 sopra. Non è influenzato dall'approssimazione.
- L'ordinamento è deterministico per input e punteggi stabili.

## Risoluzione dei problemi

| Sintomo | Causa e soluzione |
| ------- | ----------------- |
| `# no duplicates found` | Nessuna coppia rientra nell'intervallo. Abbassa `--min-similarity` o alza `--max-similarity`. |
| Manca una funzione breve nota | I simboli con meno di 4 token normalizzati vengono saltati. |
| Mancano copie esatte | Il massimo predefinito è 95%. Usa `--min-similarity 100 --max-similarity 100`. Con `--rerank`, il filtro semantico può anche rimuoverle. |
| `--rerank` fallisce | Verifica che `ollama serve` sia in esecuzione e che il modello sia stato scaricato. |
| Risultati parziali con avvisi su stderr | Correggi i file segnalati, o aggiungi `--strict` per fallire invece di pubblicare output incompleto. |

## Flag utilizzati in questa guida

| Flag               | Predefinito                     | Descrizione                                                                 |
| ------------------ | ------------------------------- | --------------------------------------------------------------------------- |
| `-o`, `--output`   | `./skeleton/fingerprints.skt`   | Percorso del file `.skt` di output.                                        |
| `--min-similarity` | `65`                            | Percentuale minima di similarità da riportare (0–100).                     |
| `--max-similarity` | `95`                            | Percentuale massima di similarità da riportare (0–100).                    |
| `--show-all`       | `false`                         | Emette anche la sezione `[fingerprints]` con i dati grezzi dei token.      |
| `--rerank`         | `false`                         | Riordina i candidati con embeddings semantici tramite Ollama. Richiede `ollama serve` e `ollama pull qwen3-embedding:0.6b`. |
| `--model`          | `qwen3-embedding:0.6b`          | Modello di embedding Ollama da utilizzare con `--rerank`.                  |
| `--collect-test`   | `false`                         | Include i file di test nell'analisi.                                       |
| `--workers`        | `NumCPU`                        | Massimo numero di goroutine di parsing concorrenti.                        |
| `--verbose`        | `false`                         | Stampa informazioni di avanzamento.                                        |

Per l'elenco completo, inclusi `--include`, `--exclude`, `--max-file-size`, `--strict` e `--header-language`, consulta il [riferimento dei flag CLI](../../reference/cli-flags/).

## Correlati

- [Comando Parse](../parse-command/)
- [Comandi Graph](../graph-commands/)