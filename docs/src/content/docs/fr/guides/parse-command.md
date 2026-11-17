---
title: Commande Parse
description: Extraire des informations structurelles du code source dans des fichiers .skt ou JSON.
---

`codeknit parse` lit les fichiers source et écrit leurs symboles (et, éventuellement, les relations entre eux) sous forme de fichiers `.skt` ou JSON. Ce guide couvre les situations que vous êtes le plus susceptible de rencontrer. Pour la liste complète des flags partagés par toutes les commandes, consultez la [référence des flags CLI](../../reference/cli-flags/). Pour la grammaire de sortie, consultez la [référence du format de sortie](../../reference/output-format/).

## Analyser un projet pour la première fois

```bash
codeknit parse ./src
```

codeknit écrit des fichiers `.skt` segmentés dans `./skeleton/` : `map_001.skt`, `map_002.skt`, et ainsi de suite. Ouvrez les segments sélectivement au lieu de tout charger en une fois.

Pour écrire ailleurs, passez le répertoire de sortie comme deuxième argument :

```bash
codeknit parse ./src ./output
```

N'utilisez pas le répertoire source, ou l'un de ses répertoires parents, comme répertoire de sortie.

## Choisir un mode de sortie

Sélectionnez un mode avec `--output-mode`.

| Mode             | Utilisez-le lorsque                                    | Résultat                                                                         |
| ---------------- | ---------------------------------------------- | ------------------------------------------------------------------------------ |
| `directory-flat` | Analyse de la plupart des projets (par défaut)                | `map_001.skt`, `map_002.skt`, ...                                              |
| `directory-tree` | Vous souhaitez consulter la structure d'un seul fichier       | Un fichier `.skt` par fichier source, reflétant l'arborescence source                          |
| `inline`         | Analyse d'un seul fichier, ou envoi vers un autre outil | Sortie sur stdout ; aucun répertoire de sortie n'est utilisé                                  |

En mode arborescence, les noms de fichiers conservent l'extension source : `src/foo.c` devient `src/foo.c.skt`, et `src/foo.h` devient `src/foo.h.skt`. Les gros fichiers sont divisés en `foo.c_part1.skt`, `foo.c_part2.skt`, et ainsi de suite.

```bash
codeknit parse ./src --output-mode directory-tree
```

Évitez d'utiliser `inline` sur un projet entier. Cela écrit tout sur stdout et peut saturer votre terminal ou la fenêtre de contexte d'un LLM.

Pour plus d'informations sur le choix d'un mode, consultez [Modes de sortie](../output-modes/).

## Réexécuter sur le même projet

codeknit refuse d'écrire dans un répertoire qui contient déjà une sortie générée pour le format sélectionné. Cela empêche le mélange de segments obsolètes et récents. Ajoutez `--clean` pour remplacer la sortie précédente :

```bash
codeknit parse ./src --clean
```

`--clean` ne supprime rien au préalable. codeknit génère la nouvelle sortie dans un répertoire de staging et remplace les anciens fichiers générés uniquement après que la génération a réussi. Si la génération échoue, ou si le processus est interrompu, la sortie précédente est conservée ou restaurée lors de l'exécution suivante. Les fichiers dans le répertoire de sortie que codeknit n'a pas générés sont préservés. Si l'un d'eux entre en collision avec un chemin généré, l'exécution échoue.

En mode JSON, les mêmes règles s'appliquent à `codeknit.json`.

**Piège : segments mélangés.** Les IDs de symboles (`S1`, `S2`, ...) et les résultats des relations peuvent changer entre les exécutions et les versions. Après une mise à niveau vers la version 0.5.0, régénérez l'ensemble de la sortie avec `--clean`. Ne combinez jamais des segments provenant de différentes exécutions.

## Analyser un seul fichier

Pointez `parse` sur le fichier et utilisez le mode `inline` :

```bash
codeknit parse ./src/main.go --output-mode inline
```

Lorsque vous sélectionnez explicitement un fichier source régulier, codeknit ignore les filtres de fichiers d'ignorance et de test pour celui-ci. Les motifs d'inclusion/exclusion et la limite de taille s'appliquent toujours.

## Obtenir du JSON pour les scripts

```bash
codeknit parse ./src --output-mode inline --format json --edges
```

Le document JSON possède les champs de niveau supérieur `files`, `symbols`, `edges` (omis lorsqu'il n'y en a pas), et `errors` (omis lorsqu'il n'y en a pas). Un petit exemple :

```json
{
  "files": ["app.go"],
  "symbols": [
    {
      "id": "app.go::User",
      "short_id": "S1",
      "name": "User",
      "file": "app.go",
      "category": "type",
      "kind": "struct",
      "signature": "type User struct",
      "span": [3, 3]
    },
    {
      "id": "app.go::Save",
      "short_id": "S2",
      "name": "Save",
      "file": "app.go",
      "category": "callable",
      "kind": "function",
      "signature": "Save(u: S1)",
      "span": [5, 5]
    }
  ],
  "edges": [
    {
      "from": "app.go::Save",
      "from_short": "S2",
      "to": "app.go::User",
      "to_short": "S1",
      "kind": "references"
    }
  ]
}
```

Pour écrire le JSON sur le disque à la place, retirez `--output-mode inline`. Dans les modes de répertoire, le JSON est écrit dans `codeknit.json`. Le JSON est un seul document et n'est pas divisé par nombre de lignes.

## Inclure les relations

Par défaut, la sortie ne contient que la section `[symbols]`, ce qui permet d'économiser des tokens. Ajoutez `--edges` lorsque vous avez besoin de savoir ce qui appelle, contient, hérite de, ou référence quoi :

```bash
codeknit parse ./src --edges
```

### Lire les relations incertaines

Codeknit analyse la syntaxe sans compilateur, donc certaines relations ne peuvent pas être établies. Chaque arête indique son degré de certitude :

```skt
S8 --references--> S5
S10 --references[unresolved]--> string, bool
S11 --calls[ambiguous]--> Helper [candidates=S12, S15]
```

- `--kind-->` est une relation résolue.
- `--kind[unresolved]-->` signifie que codeknit n'a pas pu établir la cible.
- `--kind[ambiguous]-->` signifie que plus d'une cible reste possible. Les IDs des candidats sont listés lorsque disponibles.

Les candidats sont des cibles possibles, pas des dépendances confirmées. Les points d'extrémité connus utilisent des IDs courts. Une cible non résolue sans symbole correspondant peut apparaître sous forme de nom, relatif au fichier source.

En JSON, les arêtes incertaines portent `resolution` (`"unresolved"` ou `"ambiguous"`) et, lorsque disponibles, `candidates` sous forme d'IDs complets de symboles. Elles omettent `to_short`.

Pour C et C++, les directives `#include` deviennent des symboles `meta/file` et `meta/include`, donc leurs arêtes ont des IDs courts. Cela enregistre la directive telle qu'écrite. Cela ne signifie pas que codeknit a trouvé le header.

Les métriques de graphe excluent les arêtes incertaines et indiquent combien ont été exclues. Le visualiseur de graphe HTML affiche un compte des relations incertaines et ne les dessine pas. Pour inspecter les arêtes incertaines individuelles, utilisez `--edges` sur `parse`. Consultez [Commandes de graphe](../graph-commands/) pour ces rapports.

La grammaire des arêtes se trouve dans la [référence du format de sortie](../../reference/output-format/). Les règles et limites de résolution sont dans la [référence de support des langages](../../reference/language-support/#relationship-resolution).

## Réduire la taille de la sortie

Pour compresser les tokens répétés, ajoutez `--minify` :

```bash
codeknit parse ./src --minify
```

Cela écrit un dictionnaire dans `dict.skt` (ou une section `[dict]` en ligne). Remplacez les codes tels que `d0` et `d1` par leurs valeurs de dictionnaire lors de la lecture. Les types de relations sont également encodés, mais les étiquettes de résolution telles que `[unresolved]` restent lisibles.

Pour contrôler la taille des segments, utilisez `--max-lines` (par défaut `500`) :

- En mode `directory-flat`, il s'agit d'une cible. Un bloc source n'est jamais divisé, donc un segment peut dépasser cette valeur d'un bloc.
- En mode `directory-tree`, il s'agit d'une limite stricte pour chaque segment source. Le minimum est `3`. Chaque segment répète ses en-têtes `[symbols]` et `## source-path`.

## Analyser les headers C++ nommés `.h`

Codeknit analyse les fichiers `.h` comme du C par défaut. Pour les headers C++ avec cette extension, définissez `--header-language cpp` :

```bash
codeknit parse ./src --header-language cpp
```

Le flag n'affecte que les fichiers `.h`. Les fichiers nommés `.hpp` et `.hxx` sont déjà analysés comme du C++, et les autres extensions conservent leurs analyseurs habituels.

## Contrôler quels fichiers sont analysés

Les cinq commandes de traitement de source acceptent ces flags. Ils sont listés en détail dans la [référence des flags CLI](../../reference/cli-flags/).

```bash
codeknit parse ./src --include '**/*.go' --exclude '**/generated/' --strict
```

- `--include GLOB` et `--exclude GLOB` sont répétables, des motifs de style gitignore relatifs au répertoire d'entrée. Mettez-les entre guillemets pour que votre shell ne les développe pas. Les exclusions l'emportent. Les inclusions ne remplacent pas `.gitignore`, le filtrage des fichiers de test, ou les exclusions intégrées.
- `--max-file-size BYTES` ignore les fichiers plus grands. La valeur par défaut est `10485760` (10 Mio). `0` désactive la limite.
- `--strict` échoue avant d'écrire une sortie si l'analyse ou l'analyse est incomplète. La valeur par défaut est `false`.

Codeknit ignore et signale les liens symboliques et les fichiers non réguliers. Les règles `.gitignore` des répertoires parents s'appliquent indépendamment de votre répertoire de travail.

**Piège : résultats partiels.** Les chemins illisibles, les fichiers trop grands, les liens symboliques ignorés et les avertissements de syntaxe rendent une analyse incomplète. Sans `--strict`, codeknit conserve les résultats utilisables et signale les problèmes :

- sur stderr,
- dans le champ JSON `errors`, ou
- dans la section SKT `[errors]` (en ligne, ou dans `warnings.skt` dans les modes de répertoire).

Une entrée sans sources prises en charge sélectionnées, ou une exécution dans laquelle chaque lecture de source échoue, se termine toujours sans succès. Pour des garanties exactes, consultez [Fiabilité de l'analyse](../../reference/analysis-reliability/).

## Flags de parse

Ces flags sont spécifiques à `parse`. Les flags partagés se trouvent dans la [référence des flags CLI](../../reference/cli-flags/).

| Flag               | Défaut          | Description                                                                                       |
| ------------------ | ---------------- | ------------------------------------------------------------------------------------------------- |
| `--output-mode`    | `directory-flat` | Mode de sortie : `inline`, `directory-flat`, ou `directory-tree`.                                     |
| `--format`         | `skt`            | Format de sortie : `skt` ou `json`.                                                                   |
| `--header-language`| `c`              | Analyseur pour les fichiers `.h` : `c` ou `cpp`. Les autres extensions restent inchangées.                              |
| `--max-lines`      | `500`            | Cible de lignes pour la sortie plate ; limite stricte (minimum 3) pour chaque segment source en arborescence.                   |
| `--collect-test`   | `false`          | Inclure les fichiers de test.                                                                               |
| `--minify`         | `false`          | Utiliser la compression basée sur un dictionnaire.                                                                 |
| `--edges`          | `false`          | Inclure la section `[edges]` (désactivée par défaut pour économiser des tokens).                                    |
| `--clean`          | `false`          | Remplacer la sortie générée précédente, uniquement après que la génération a réussi.                                   |
| `--workers`        | `NumCPU`         | Nombre maximal de goroutines d'analyse concurrentes.                                                            |
| `--verbose`        | `false`          | Afficher les informations de progression.                                                                       |

## Conseils

- Utilisez `directory-flat` pour la plupart des projets, et ne lisez que les segments dont vous avez besoin.
- Utilisez `--minify` sur les grandes bases de code pour économiser des tokens.
- Passez `--edges` uniquement lorsque vous avez besoin des données de relation.
- Vérifiez la [matrice de capacités testées et les limites connues](../../reference/language-support/) avant de vous fier à l'extraction pour un langage spécifique.

## Voir aussi

- [Modes de sortie](../output-modes/)
- [Référence du format de sortie](../../reference/output-format/)
- [Fiabilité de l'analyse](../../reference/analysis-reliability/)
- [Commandes de graphe](../graph-commands/)