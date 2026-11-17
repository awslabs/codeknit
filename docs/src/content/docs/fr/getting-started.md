---
title: Démarrer
description: Prenez en main codeknit en moins de 5 minutes.
---

Dans ce tutoriel, vous allez compiler codeknit à partir des sources, exécuter votre première analyse et lire le fichier `.skt` généré. Vous allez analyser un petit package issu du dépôt codeknit, vous n'avez donc pas besoin de votre propre projet.

## Prérequis

- **Go 1.26 ou ultérieur.** Exécutez `go version` pour vérifier.
- **Un compilateur C.** codeknit utilise tree-sitter, qui nécessite CGo.
- **Git** et **Make**.

## 1. Compiler codeknit

Clonez le dépôt et compilez :

```bash
git clone https://github.com/awslabs/codeknit.git
cd codeknit
make build
```

Vous devriez voir la compilation se terminer sans erreurs, et un binaire à l'emplacement `./bin/codeknit`.

## 2. Ajouter codeknit à votre PATH

Depuis le répertoire `codeknit`, ajoutez `bin` à votre `PATH` en utilisant la ligne correspondant à votre shell.

Pour bash ou zsh :

```bash
export PATH="$PATH:$(pwd)/bin"
```

Pour fish :

```fish
fish_add_path (pwd)/bin
```

Cela s'applique à votre session bash ou zsh actuelle. Pour le conserver pour les sessions ultérieures, ajoutez la même ligne à `~/.bashrc` ou `~/.zshrc`, en remplaçant `$(pwd)/bin` par le chemin absolu vers `bin`.

## 3. Vérifier l'installation

```bash
codeknit --version
```

Vous devriez voir la version de codeknit s'afficher.

## 4. Exécuter votre première analyse

Restez dans le répertoire `codeknit` et analysez le package `internal/console` :

```bash
codeknit parse ./internal/console --edges
```

codeknit scanne le répertoire, analyse chaque fichier source pris en charge et génère des fichiers `.skt`. Un *fichier `.skt`* est un fichier texte compact qui répertorie les symboles de votre code et les relations entre eux. Le drapeau `--edges` ajoute les relations. Elles sont omises par défaut pour économiser de l'espace.

Vous devriez voir un bref résumé à la fin de la commande, ainsi qu'un nouveau répertoire `./skeleton` dans votre répertoire courant.

Listez le contenu du répertoire de sortie :

```bash
ls ./skeleton
```

Vous devriez voir un fichier tel que `map_001.skt`.

## 5. Lire la sortie

Ouvrez le fichier :

```bash
cat ./skeleton/map_001.skt
```

Vous devriez voir en premier une section `[symbols]`. Un *symbole* est un élément nommé dans votre code, comme une fonction, un type ou un champ. Les symboles sont regroupés sous un en-tête `## <chemin vers le fichier>`. Votre sortie inclut des lignes comme celles-ci :

```skt
[symbols]
## <chemin vers>/console.go
S1 module/package L5-L5 console
S2 type/struct L17-L25 Console {exporté}
...
S10 callable/function L28-L41 New() -> *S2 {exporté}
...
S12 callable/method L49-L49 Error(msg: string) {exporté, receiver=*Console}
```

Chaque ligne comprend les mêmes parties :

| Partie | Exemple | Signification |
| --- | --- | --- |
| ShortID | `S10` | Un identifiant court pour le symbole, utilisé ailleurs dans la sortie |
| Catégorie et type | `callable/function` | Le type de symbole |
| Plage de lignes | `L28-L41` | Les lignes couvertes par le symbole dans le fichier source |
| Signature | `New() -> *S2` | Le nom, les paramètres et le type de retour |
| Propriétés | `{exporté}` | Détails supplémentaires entre accolades |

Dans `New() -> *S2`, `S2` est le ShortID du type `Console`. codeknit remplace les noms de types par des ShortID lorsqu'il connaît le symbole. Les numéros de ligne peuvent différer légèrement si votre version du dépôt est plus récente.

Faites défiler vers la section `[edges]`. Une *arête* est une relation entre deux symboles. Chaque ligne d'arête nomme le symbole source, le type de relation et le ou les symboles cibles :

```skt
[edges]
S2 --contains--> S4
S4 --calls--> S5
```

Cet exemple montre la structure d'une ligne d'arête. Cela signifie que « `S2` contient `S4` » et que « `S4` appelle `S5` ». Vous pouvez rechercher chaque ShortID dans la section `[symbols]` pour voir à quoi il fait référence.

## Ce que vous avez accompli

Vous avez compilé codeknit, l'avez ajouté à votre `PATH`, exécuté `codeknit parse` et lu les sections `[symbols]` et `[edges]` de la sortie `.skt`.

## Étapes suivantes

- Analysez votre propre projet, choisissez une autre disposition de sortie et relancez en toute sécurité : consultez le [guide de la commande parse](../guides/parse-command/).
- Visualisez et analysez la structure du code : consultez le [guide des commandes graph](../guides/graph-commands/).
- Trouvez les doublons et quasi-doublons de code : consultez le [guide de la commande fingerprint](../guides/fingerprint-command/).
- Consultez chaque partie du format `.skt` : voir la [référence du format de sortie](../reference/output-format/).
- Consultez chaque drapeau de ligne de commande : voir la [référence des drapeaux CLI](../reference/cli-flags/).