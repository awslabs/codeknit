---
title: Guida introduttiva
description: Inizia a utilizzare codeknit in meno di 5 minuti.
---

In questa guida introduttiva compilerai codeknit dal sorgente, eseguirai il tuo primo parsing e leggerai l'output `.skt` generato. Analizzerai un piccolo pacchetto dalla repository di codeknit, quindi non avrai bisogno di un tuo progetto.

## Prerequisiti

- **Go 1.26 o successivo.** Esegui `go version` per verificare.
- **Un compilatore C.** codeknit utilizza tree-sitter, che necessita di CGo.
- **Git** e **Make**.

## 1. Compilare codeknit

Clona la repository e compila:

```bash
git clone https://github.com/awslabs/codeknit.git
cd codeknit
make build
```

Dovresti vedere la compilazione terminare senza errori e trovare un binario in `./bin/codeknit`.

## 2. Aggiungere codeknit al PATH

Dalla directory `codeknit`, aggiungi `bin` al tuo `PATH` utilizzando la riga appropriata per la tua shell.

Per bash o zsh:

```bash
export PATH="$PATH:$(pwd)/bin"
```

Per fish:

```fish
fish_add_path (pwd)/bin
```

Questa modifica si applica solo alla sessione corrente di bash o zsh. Per mantenerla anche nelle sessioni future, aggiungi la stessa riga a `~/.bashrc` o `~/.zshrc`, sostituendo `$(pwd)/bin` con il percorso assoluto alla directory `bin`.

## 3. Verificare l'installazione

```bash
codeknit --version
```

Dovresti vedere la versione di codeknit stampata a schermo.

## 4. Eseguire il primo parsing

Rimani nella directory `codeknit` e analizza il pacchetto `internal/console`:

```bash
codeknit parse ./internal/console --edges
```

codeknit scansiona la directory, analizza ogni file sorgente supportato e scrive i file `.skt`. Un *file `.skt`* è un file di testo compatto che elenca i simboli nel tuo codice e le relazioni tra di essi. Il flag `--edges` aggiunge le relazioni. Queste sono omesse di default per risparmiare spazio.

Dovresti vedere un breve riassunto al termine del comando e una nuova directory `./skeleton` nella directory corrente.

Elenca l'output:

```bash
ls ./skeleton
```

Dovresti vedere un file come `map_001.skt`.

## 5. Leggere l'output

Apri il file:

```bash
cat ./skeleton/map_001.skt
```

Dovresti vedere prima una sezione `[symbols]`. Un *simbolo* è un elemento nominato nel tuo codice, come una funzione, un tipo o un campo. I simboli sono raggruppati sotto un'intestazione `## <percorso del file>`. Il tuo output includerà righe come queste:

```skt
[symbols]
## <path to>/console.go
S1 module/package L5-L5 console
S2 type/struct L17-L25 Console {exported}
...
S10 callable/function L28-L41 New() -> *S2 {exported}
...
S12 callable/method L49-L49 Error(msg: string) {exported, receiver=*Console}
```

Ogni riga ha le stesse parti:

| Parte | Esempio | Significato |
| --- | --- | --- |
| Short ID | `S10` | Un'etichetta breve per il simbolo, utilizzata altrove nell'output |
| Categoria e tipo | `callable/function` | Che tipo di simbolo è |
| Intervallo di righe | `L28-L41` | Le righe coperte dal simbolo nel file sorgente |
| Firma | `New() -> *S2` | Il nome, i parametri e il tipo di ritorno |
| Proprietà | `{exported}` | Dettagli aggiuntivi tra parentesi graffe |

In `New() -> *S2`, `S2` è lo Short ID del tipo `Console`. codeknit sostituisce i nomi dei tipi con gli Short ID quando conosce il simbolo. I numeri di riga possono differire leggermente se la tua copia è più recente.

Scorri verso il basso fino alla sezione `[edges]`. Un *arco* è una relazione tra due simboli. Ogni riga di arco indica il simbolo di origine, il tipo di relazione e il simbolo o i simboli di destinazione:

```skt
[edges]
S2 --contains--> S4
S4 --calls--> S5
```

Questo esempio mostra la struttura di una riga di arco. Significa che "`S2` contiene `S4`" e "`S4` chiama `S5`". Puoi cercare ogni Short ID nella sezione `[symbols]` per vedere a cosa si riferisce.

## Cosa hai ottenuto

Hai compilato codeknit, lo hai aggiunto al tuo `PATH`, eseguito `codeknit parse` e letto le sezioni `[symbols]` e `[edges]` dell'output `.skt`.

## Passaggi successivi

- Analizza il tuo progetto, scegli un layout di output diverso e riesegui in sicurezza: consulta la [guida al comando parse](../guides/parse-command/).
- Visualizza e analizza la struttura del codice: consulta la [guida ai comandi graph](../guides/graph-commands/).
- Trova codice duplicato e quasi duplicato: consulta la [guida al comando fingerprint](../guides/fingerprint-command/).
- Consulta ogni parte del formato `.skt`: vedi il [riferimento al formato di output](../reference/output-format/).
- Consulta ogni flag della riga di comando: vedi il [riferimento ai flag CLI](../reference/cli-flags/).