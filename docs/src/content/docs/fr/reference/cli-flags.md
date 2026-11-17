---
title: Référence de l'interface en ligne de commande
description: Référence complète pour toutes les commandes et options de codeknit.
---

Cette page répertorie chaque commande de codeknit, ses arguments et ses options. Pour des instructions orientées tâches, consultez le [guide de la commande parse](../../guides/parse-command/), le [guide des commandes graph](../../guides/graph-commands/), et le [guide de la commande fingerprint](../../guides/fingerprint-command/).

## Résumé des commandes

| Commande                                      | Objectif                                                        |
| --------------------------------------------- | -------------------------------------------------------------- |
| `codeknit`                                    | Lancer l'interface utilisateur interactive en terminal.         |
| `codeknit parse <input-path> [output-dir]`    | Extraire les symboles et relations dans une sortie `.skt` ou JSON. |
| `codeknit graph show <input-path>`            | Générer un graphe HTML interactif.                              |
| `codeknit graph analyze <input-path>`         | Exécuter des algorithmes de graphe structurel et écrire un rapport `.skt`. |
| `codeknit graph hotspots <input-path>`        | Classer les points chauds de modification en utilisant l'historique Git et la structure du graphe. |
| `codeknit fingerprint <input-path>`           | Détecter le code en doublon et quasi-doublon.                   |
| `codeknit completion <shell>`                 | Générer un script de complétion pour le shell.                  |

Cinq commandes traitent le code source : `parse`, `graph show`, `graph analyze`, `graph hotspots`, et `fingerprint`. Elles partagent les options décrites dans [Options partagées de traitement du code source](#options-partagées-de-traitement-du-code-source).

## Options globales

| Option            | Type | Valeur par défaut | Description                  |
| ----------------- | ---- | ----------------- | ---------------------------- |
| `--version`       | bool | `false`           | Afficher les informations de version. |
| `-h`, `--help`    | bool | `false`           | Afficher l'aide pour la commande. |

## Options partagées de traitement du code source

Ces options apparaissent dans le tableau de chaque commande de traitement du code source.

| Option                    | Type                | Valeur par défaut       | Valeurs autorisées       | Description                                                                          |
| ----------------------- | ------------------- | ----------------------- | ------------------------ | ------------------------------------------------------------------------------------ |
| `--header-language`     | string              | `c`                     | `c`, `cpp`               | Analyseur utilisé pour les fichiers `.h`.                                            |
| `--include GLOB`        | string, répétable   | Tous les sources supportés | Glob de style gitignore | Inclure les chemins de source correspondants. Plusieurs motifs forment une union.    |
| `--exclude GLOB`        | string, répétable   | Aucun                   | Glob de style gitignore | Exclure les chemins ou répertoires correspondants.                                   |
| `--max-file-size BYTES` | int                 | `10485760` (10 Mio)     | `0` ou plus              | Ignorer les fichiers source plus grands que cette taille. `0` désactive la limite.   |
| `--strict`              | bool                | `false`                 | —                        | Échouer avant l'écriture de la sortie si des diagnostics de scan ou d'analyse surviennent. |

### `--header-language`

S'applique uniquement aux fichiers `.h`. Les autres extensions conservent leur analyseur habituel (`.hpp` et `.hxx` utilisent déjà C++).

- `c` (par défaut) préserve l'extraction des macros C, des typedefs et des unions.
- `cpp` sélectionne l'analyseur C++. À utiliser pour les projets C++ dont les en-têtes sont nommés `.h`.

L'interface utilisateur utilise la sélection par défaut `c` pour les fichiers `.h`.

```bash
codeknit parse ./src --header-language cpp
```

Pour les capacités d'extraction testées et les limites connues par langage, consultez [Support des langages et fidélité de l'API](../language-support/). Cette page décrit les constructions testées, et non la couverture complète du langage ou la sémantique du compilateur.

### Sélection des sources

- Les motifs utilisent des globs de style gitignore séparés par des barres obliques, relatifs au répertoire d'entrée. Pour une entrée d'un seul fichier, les motifs correspondent au nom de base du fichier.
- Placez les motifs entre guillemets pour éviter leur expansion par le shell.
- Les exclusions prennent le pas sur les inclusions.
- Les inclusions ne remplacent pas les règles `.gitignore`, le filtrage des fichiers de test, ou les exclusions intégrées.
- Les motifs de négation (`!motif`) ne sont pas acceptés dans `--include` ou `--exclude`.
- Les analyses de répertoire respectent les fichiers `.gitignore` ancestraux et imbriqués indépendamment du répertoire de travail actuel.
- La sélection explicite d'un fichier source régulier contourne les règles d'ignorance et de filtrage des tests. Les inclusions, exclusions et limites de taille s'appliquent toujours.
- Les fichiers de test sont exclus sauf si `--collect-test` est défini.

```bash
codeknit parse ./src --include '**/*.go' --exclude '**/generated/' --strict
```

### Diagnostics et `--strict`

- Les liens symboliques, les liens symboliques de répertoire et les fichiers non réguliers sont ignorés et signalés.
- Les chemins illisibles, les échecs de lecture des fichiers d'ignorance, les sources trop volumineuses et les avertissements de syntaxe rendent une analyse incomplète.
- Par défaut, les résultats partiels utilisables sont conservés. Les diagnostics apparaissent sur stderr. `parse` les enregistre également dans les `errors` JSON ou dans les `[errors]` SKT.
- Avec `--strict`, une analyse incomplète est rejetée avant l'écriture de la sortie.
- Une exécution sans source supportée sélectionnée, ou dans laquelle chaque lecture de source échoue, se termine toujours sans succès.
- L'interface utilisateur applique la même politique de liens symboliques et la limite par défaut de 10 Mio. Ses formulaires n'exposent pas les options de sélection et strictes.

Consultez [Fiabilité de l'analyse](../analysis-reliability/) pour des garanties exactes.

## `codeknit`

```bash
codeknit
```

L'exécution de `codeknit` sans sous-commande lance l'interface utilisateur interactive en terminal. L'interface propose les commandes `parse`, `graph show`, `graph analyze`, `graph hotspots` et `fingerprint` ainsi que leurs options.

## `codeknit parse`

```bash
codeknit parse <input-path> [output-dir]
```

| Argument       | Requis | Valeur par défaut | Description                                                              |
| -------------- | ------ | ----------------- | ------------------------------------------------------------------------ |
| `input-path`   | Oui    | —                 | Répertoire source ou fichier source unique.                              |
| `output-dir`   | Non    | `./skeleton`      | Répertoire de sortie. Non utilisé lorsque `--output-mode` est `inline`. |

| Option                    | Type                | Valeur par défaut      | Valeurs autorisées                                   | Description                                                                                         |
| ----------------------- | ------------------- | ---------------------- | ---------------------------------------------------- | --------------------------------------------------------------------------------------------------- |
| `--output-mode`         | string              | `directory-flat`       | `inline`, `directory-flat`, `directory-tree`         | Où et comment la sortie est écrite.                                                                |
| `--format`              | string              | `skt`                  | `skt`, `json`                                        | Format de sortie.                                                                                   |
| `--max-lines`           | int                 | `500`                  | Pour `directory-tree`, 3 ou plus                     | Cible de lignes pour la sortie plate. Limite stricte de chunks de source pour la sortie en arbre.  |
| `--collect-test`        | bool                | `false`                | —                                                  | Inclure les fichiers de test.                                                                       |
| `--minify`              | bool                | `false`                | —                                                  | Activer la compression de sortie basée sur un dictionnaire.                                        |
| `--edges`               | bool                | `false`                | —                                                  | Inclure la section `[edges]`. Désactivé par défaut pour économiser des tokens.                      |
| `--clean`               | bool                | `false`                | —                                                  | Remplacer la sortie générée précédemment, uniquement après une génération réussie.                  |
| `--workers`             | int                 | `0` (`NumCPU`)         | `0` ou plus                                         | Nombre maximal de goroutines d'analyse concurrentes.                                                |
| `--verbose`             | bool                | `false`                | —                                                  | Afficher des informations de progression.                                                           |
| `--header-language`     | string              | `c`                    | `c`, `cpp`                                         | Analyseur utilisé pour les fichiers `.h`.                                                          |
| `--include GLOB`        | string, répétable   | Tous les sources supportés | —                                               | Inclure les chemins de source correspondants.                                                       |
| `--exclude GLOB`        | string, répétable   | Aucun                  | —                                                  | Exclure les chemins ou répertoires correspondants.                                                  |
| `--max-file-size BYTES` | int                 | `10485760`             | `0` désactive la limite                             | Ignorer les fichiers source plus grands.                                                            |
| `--strict`              | bool                | `false`                | —                                                  | Échouer avant la sortie si des diagnostics de scan ou d'analyse surviennent.                        |

### Modes de sortie

| Valeur             | Comportement                                                                                         |
| ----------------- | ---------------------------------------------------------------------------------------------------- |
| `inline`          | Écrire sur stdout. Aucun répertoire de sortie n'est utilisé.                                         |
| `directory-flat`  | Écrire des fichiers chunkés dans le répertoire de sortie. Les blocs de source restent intacts, donc `--max-lines` est une cible de division plutôt qu'une limite stricte. |
| `directory-tree`  | Reproduire l'arborescence des sources. Les noms conservent l'extension de la source (`foo.c.skt`, `foo.h.skt`). Les fichiers divisés utilisent `foo.c_part1.skt`. |

### Comportement de la sortie

- Sans `--clean`, la présence de sortie générée existante pour le format sélectionné dans le répertoire de sortie provoque une erreur.
- Avec `--clean`, les fichiers `.skt` obsolètes (format SKT) ou `codeknit.json` (format JSON) sont remplacés uniquement après une génération réussie. Les entrées non liées sont préservées.
- En format JSON, les modes de répertoire écrivent `codeknit.json`. JSON est un document unique et n'est pas divisé par nombre de lignes.
- `dict.skt` et `warnings.skt` sont des noms de fichiers auxiliaires réservés.
- Ne pas utiliser un répertoire source ou l'un de ses ancêtres comme répertoire de sortie.

Consultez [Modes de sortie](../../guides/output-modes/) et la [référence du format de sortie](../output-format/) pour plus de détails.

```bash
codeknit parse ./src
codeknit parse ./src ./out --output-mode directory-tree
codeknit parse ./src --output-mode inline --format json --edges
```

## `codeknit graph show`

```bash
codeknit graph show <input-path>
```

| Argument       | Requis | Description                              |
| -------------- | ------ | ---------------------------------------- |
| `input-path`   | Oui    | Répertoire source ou fichier source unique. |

| Option                    | Type                | Valeur par défaut                      | Valeurs autorisées          | Description                                                       |
| ----------------------- | ------------------- | -------------------------------------- | --------------------------- | ----------------------------------------------------------------- |
| `-o`, `--output`        | string              | `./skeleton/codeknit-graph.html`       | Chemin de fichier           | Fichier HTML de sortie.                                           |
| `--collect-test`        | bool                | `false`                                | —                           | Inclure les fichiers de test.                                     |
| `--workers`             | int                 | `0` (`NumCPU`)                         | `0` ou plus                 | Nombre maximal de goroutines d'analyse concurrentes.              |
| `--verbose`             | bool                | `false`                                | —                           | Afficher des informations de progression.                         |
| `--header-language`     | string              | `c`                                    | `c`, `cpp`                  | Analyseur utilisé pour les fichiers `.h`.                         |
| `--include GLOB`        | string, répétable   | Tous les sources supportés             | —                           | Inclure les chemins de source correspondants.                     |
| `--exclude GLOB`        | string, répétable   | Aucun                                  | —                           | Exclure les chemins ou répertoires correspondants.                |
| `--max-file-size BYTES` | int                 | `10485760`                             | `0` désactive la limite     | Ignorer les fichiers source plus grands.                          |
| `--strict`              | bool                | `false`                                | —                           | Échouer avant la sortie si des diagnostics de scan ou d'analyse surviennent. |

La commande écrit un fichier HTML autonome et l'ouvre dans le navigateur par défaut. Le visualiseur dessine les relations résolues et affiche un compte des relations incertaines.

```bash
codeknit graph show ./src -o graph.html
```

## `codeknit graph analyze`

```bash
codeknit graph analyze <input-path>
```

| Argument       | Requis | Description                              |
| -------------- | ------ | ---------------------------------------- |
| `input-path`   | Oui    | Répertoire source ou fichier source unique. |

| Option                       | Type                | Valeur par défaut                      | Valeurs autorisées          | Description                                                  |
| -------------------------- | ------------------- | -------------------------------------- | --------------------------- | ------------------------------------------------------------ |
| `-o`, `--output`           | string              | `./skeleton/graph_analysis.skt`        | Chemin de fichier           | Fichier `.skt` de sortie.                                    |
| `--collect-test`           | bool                | `false`                                | —                           | Inclure les fichiers de test.                                |
| `--workers`                | int                 | `0` (`NumCPU`)                         | `0` ou plus                 | Nombre maximal de goroutines d'analyse concurrentes.         |
| `--verbose`                | bool                | `false`                                | —                           | Afficher des informations de progression.                    |
| `--fan-threshold`          | int                 | `10`                                   | —                           | Seuil minimal de fan-in ou fan-out pour signaler un symbole hub. |
| `--god-threshold`          | int                 | `15`                                   | —                           | Nombre minimal d'arêtes `contains` pour signaler une god class ou une fonction. |
| `--max-inheritance-depth`  | int                 | `5`                                    | —                           | Signaler les chaînes d'héritage plus profondes que cette valeur. |
| `--top-n`                  | int                 | `30`                                   | `0` = pas de limite         | Limiter les sections de sortie classées.                     |
| `--betweenness-threshold`  | float               | `0.001`                                | —                           | Valeur minimale de centralité d'intermédiarité à rapporter.  |
| `--propagation-cutoff`     | float               | `0.05`                                 | —                           | Probabilité minimale pour poursuivre la propagation des changements. |
| `--header-language`        | string              | `c`                                    | `c`, `cpp`                  | Analyseur utilisé pour les fichiers `.h`.                    |
| `--include GLOB`           | string, répétable   | Tous les sources supportés             | —                           | Inclure les chemins de source correspondants.                |
| `--exclude GLOB`           | string, répétable   | Aucun                                  | —                           | Exclure les chemins ou répertoires correspondants.            |
| `--max-file-size BYTES`    | int                 | `10485760`                             | `0` désactive la limite     | Ignorer les fichiers source plus grands.                     |
| `--strict`                 | bool                | `false`                                | —                           | Échouer avant la sortie si des diagnostics de scan ou d'analyse surviennent. |

Les métriques de dépendance excluent les relations incertaines et les métadonnées de fichier/inclusion. Le rapport enregistre les comptes de relations exclues.

```bash
codeknit graph analyze ./src -o analysis.skt --fan-threshold 15 --top-n 50
```

## `codeknit graph hotspots`

```bash
codeknit graph hotspots <input-path>
```

| Argument       | Requis | Description                                                  |
| -------------- | ------ | ------------------------------------------------------------ |
| `input-path`   | Oui    | Répertoire source à l'intérieur d'un dépôt Git.              |

| Option                       | Type                | Valeur par défaut                    | Valeurs autorisées                              | Description                                                  |
| -------------------------- | ------------------- | ------------------------------------ | ----------------------------------------------- | ------------------------------------------------------------ |
| `-o`, `--output`           | string              | `./skeleton/hotspots.skt`            | Chemin de fichier                               | Fichier de sortie.                                           |
| `--format`                 | string              | `skt`                                | `skt`, `json`                                   | Format de sortie.                                            |
| `--since`                  | string              | `12mo`                               | Fenêtre telle que `180d`, `12mo`, ou `2y`       | Fenêtre d'historique.                                        |
| `--max-commits`            | int                 | `2000`                               | —                                               | Nombre maximal de commits à inspecter.                       |
| `--max-files-per-commit`   | int                 | `50`                                 | —                                               | Exclure les commits qui touchent plus de fichiers que cette valeur. |
| `--min-cochanges`          | int                 | `3`                                  | —                                               | Nombre minimal de commits partagés pour rapporter un couplage temporel. |
| `--top-n`                  | int                 | `30`                                 | —                                               | Nombre maximal de résultats par section.                     |
| `--include-merges`         | bool                | `false`                              | —                                               | Inclure les commits de fusion dans les métriques d'historique. |
| `--collect-test`           | bool                | `false`                              | —                                               | Inclure les fichiers de test.                                |
| `--workers`                | int                 | `0` (`NumCPU`)                       | `0` ou plus                                     | Nombre maximal de goroutines d'analyse concurrentes.         |
| `--verbose`                | bool                | `false`                              | —                                               | Afficher des informations de progression.                    |
| `--header-language`        | string              | `c`                                  | `c`, `cpp`                                      | Analyseur utilisé pour les fichiers `.h`.                    |
| `--include GLOB`           | string, répétable   | Tous les sources supportés           | —                                               | Inclure les chemins de source correspondants.                |
| `--exclude GLOB`           | string, répétable   | Aucun                                | —                                               | Exclure les chemins ou répertoires correspondants.            |
| `--max-file-size BYTES`    | int                 | `10485760`                           | `0` désactive la limite                         | Ignorer les fichiers source plus grands.                     |
| `--strict`                 | bool                | `false`                              | —                                               | Échouer avant la sortie si des diagnostics de scan ou d'analyse surviennent. |

La commande combine la fréquence des commits, le churn des lignes et la récence avec le PageRank au niveau des fichiers, le fan-in transitif et la centralité d'intermédiarité. Elle rapporte également le couplage temporel entre les fichiers qui changent ensemble. Les commits de fusion sont exclus par défaut.

```bash
codeknit graph hotspots ./src --since 2y --format json -o hotspots.json
```

## `codeknit fingerprint`

```bash
codeknit fingerprint <input-path>
```

| Argument       | Requis | Description                              |
| -------------- | ------ | ---------------------------------------- |
| `input-path`   | Oui    | Répertoire source ou fichier source unique. |

| Option                    | Type                | Valeur par défaut                      | Valeurs autorisées          | Description                                                                                                      |
| ----------------------- | ------------------- | -------------------------------------- | --------------------------- | ---------------------------------------------------------------------------------------------------------------- |
| `-o`, `--output`        | string              | `./skeleton/fingerprints.skt`          | Chemin de fichier           | Fichier `.skt` de sortie.                                                                                        |
| `--min-similarity`      | int                 | `65`                                   | `0`–`100`                   | Pourcentage minimal de similarité à rapporter.                                                                   |
| `--max-similarity`      | int                 | `95`                                   | `0`–`100`                   | Pourcentage maximal de similarité à rapporter.                                                                   |
| `--show-all`            | bool                | `false`                                | —                           | Également émettre la section `[fingerprints]` avec les données brutes de tokens.                                |
| `--rerank`              | bool                | `false`                                | —                           | Trouver les voisins sémantiques et reclasser les candidats avec des embeddings Ollama. Nécessite `ollama serve` et `ollama pull qwen3-embedding:0.6b`. |
| `--model`               | string              | `qwen3-embedding:0.6b`                 | Nom de modèle Ollama        | Modèle d'embedding Ollama à utiliser avec `--rerank`.                                                            |
| `--collect-test`        | bool                | `false`                                | —                           | Inclure les fichiers de test.                                                                                    |
| `--workers`             | int                 | `0` (`NumCPU`)                         | `0` ou plus                 | Nombre maximal de goroutines d'analyse concurrentes.                                                             |
| `--verbose`             | bool                | `false`                                | —                           | Afficher des informations de progression.                                                                        |
| `--header-language`     | string              | `c`                                    | `c`, `cpp`                  | Analyseur utilisé pour les fichiers `.h`.                                                                        |
| `--include GLOB`        | string, répétable   | Tous les sources supportés             | —                           | Inclure les chemins de source correspondants.                                                                     |
| `--exclude GLOB`        | string, répétable   | Aucun                                  | —                           | Exclure les chemins ou répertoires correspondants.                                                                |
| `--max-file-size BYTES` | int                 | `10485760`                             | `0` désactive la limite     | Ignorer les fichiers source plus grands.                                                                          |
| `--strict`              | bool                | `false`                                | —                           | Échouer avant la sortie si des diagnostics de scan ou d'analyse surviennent.                                     |

### Plage de similarité

La plage par défaut, 65–95 %, exclut les correspondances à 100 %. Pour rapporter les correspondances exactes de tokens normalisés, définissez les deux bornes à `100` :

```bash
codeknit fingerprint ./src --min-similarity 100 --max-similarity 100
```

Les candidats exacts de tokens normalisés sont complets. Avec 500 symboles éligibles ou plus, les autres candidats de quasi-doublons utilisent une récupération approximative des 50 voisins les plus proches. `--rerank` peut supprimer un candidat structurellement exact. Consultez [Fiabilité de l'analyse](../analysis-reliability/#garanties-de-détection-des-doublons) pour plus de détails.

## `codeknit completion`

```bash
codeknit completion <shell>
```

| Argument  | Requis | Valeurs autorisées                         | Description                                   |
| --------- | ------ | ------------------------------------------ | --------------------------------------------- |
| `shell`   | Oui    | `bash`, `zsh`, `fish`, `powershell`        | Shell pour lequel générer un script de complétion. |

Le script est écrit sur stdout.

```bash
codeknit completion zsh >> ~/.zshrc
```

## Pages connexes

- [Référence du format de sortie](../output-format/)
- [Support des langages et fidélité de l'API](../language-support/)
- [Fiabilité de l'analyse](../analysis-reliability/)