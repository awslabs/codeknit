---
title: Commande Parse
description: Extraire des informations structurelles du code source dans des fichiers .skt ou JSON.
---

La commande `codeknit parse` extrait des informations structurelles de votre base de code — telles que les fonctions, classes, méthodes, variables et leurs relations — et les sortie au format compact `.skt` par défaut. Utilisez JSON lorsque vous avez besoin d'une sortie lisible par machine pour des scripts, des intégrations ou des outils en aval.

## Utilisation de base

```bash
codeknit parse <input-path> [output-dir]
```

- **`<input-path>`** : Chemin vers le répertoire ou le fichier que vous souhaitez analyser.
- **`[output-dir]`** : Répertoire de sortie optionnel. Si non fourni, par défaut `./skeleton`.

### Exemples

```bash
# Parse a project, output to default directory ./skeleton
codeknit parse ./src

# Parse and write to a custom output directory
codeknit parse ./src ./output

# Parse a single file and output to stdout
codeknit parse ./src/main.go --output-mode inline

# Emit machine-readable JSON to stdout
codeknit parse ./src --output-mode inline --format json
```

## Modes de sortie

Utilisez `--output-mode` pour contrôler la structure de la sortie. Trois modes sont disponibles :

| Mode             | Description                                                                              | Meilleur pour                                            |
| ---------------- | ---------------------------------------------------------------------------------------- | -------------------------------------------------------- |
| `directory-flat` | Écrit des fichiers `.skt` segmentés (par exemple `map_001.skt`, `map_002.skt`) dans le répertoire de sortie. | ✅ **La plupart des projets** — mode par défaut et recommandé |
| `directory-tree` | Reproduit la structure du répertoire source, créant un fichier `.skt` par fichier source. | Naviguer dans la sortie en parallèle du code source      |
| `inline`         | Affiche toute la sortie sur stdout.                                                      | Fichiers uniques ou redirection vers d'autres outils     |

> **Astuce** : Utilisez `directory-flat` par défaut, sauf si vous travaillez avec un seul fichier. Évitez `inline` pour les grandes entrées car cela peut submerger les fenêtres de contexte.

## Flags

| Flag             | Défaut           | Description                                                                  |
| ---------------- | ---------------- | ---------------------------------------------------------------------------- |
| `--output-mode`  | `directory-flat` | Mode de sortie : `inline`, `directory-flat`, ou `directory-tree`             |
| `--format`       | `skt`            | Format de sortie : `skt` ou `json`                                           |
| `--max-lines`    | `500`            | Nombre maximal de lignes par fichier de sortie en modes flat/tree           |
| `--collect-test` | `false`          | Inclure les fichiers de test dans l'analyse                                  |
| `--minify`       | `false`          | Activer la compression basée sur un dictionnaire pour réduire l'utilisation de tokens |
| `--edges`        | `false`          | Inclure la section `[edges]` avec les données de relation (appels, contient, etc.) |
| `--clean`        | `false`          | Supprimer les fichiers `.skt` existants dans le répertoire de sortie avant l'écriture |
| `--workers`      | `NumCPU`         | Nombre maximal de goroutines d'analyse concurrentes (0 = utiliser tous les cœurs CPU) |
| `--verbose`      | `false`          | Afficher les informations de progression et de timing pendant le traitement |

## Modèles courants

```bash
# First run on a project
codeknit parse ./src
```

```bash
# Re-run and clean previous output
codeknit parse ./src --clean
```

```bash
# Parse a single file to stdout
codeknit parse ./src/main.go --output-mode inline
```

```bash
# Minify output for large codebases
codeknit parse ./src --minify
```

```bash
# Include relationship edges (e.g., for dependency analysis)
codeknit parse ./src --edges
```

```bash
# Emit JSON for another tool
codeknit parse ./src --output-mode inline --format json --edges
```

Exemple de sortie JSON :

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

```bash
# Mirror source tree structure in output
codeknit parse ./src --output-mode directory-tree
```

## Incertitude des relations

Avec `--edges`, Codeknit préserve les relations même lorsque la syntaxe et la portée n'établissent pas une cible unique :

```skt
S1 --calls--> S2
S1 --references[unresolved]--> string, bool
S1 --calls[ambiguous]--> Helper [candidates=S3, S4]
```

Un statut omis signifie résolu par l'analyseur. `unresolved` signifie que la dépendance n'a pas pu être établie ; `ambiguous` signifie que des interprétations concurrentes restent possibles. Les candidats sont des cibles possibles, pas des connexions confirmées. Les points d'extrémité connus utilisent des ShortID ; les cibles non résolues sans symboles peuvent apparaître sous forme de noms.

JSON transporte les mêmes informations dans `resolution` et `candidates`, et omet `to_short` pour les cibles incertaines. C/C++ inclut des symboles `meta/file` et `meta/include` afin que les arêtes utilisent des ShortID ; cela enregistre les directives écrites sans prétendre à une résolution de niveau compilateur des en-têtes.

Les métriques de dépendance excluent les relations incertaines. Le graphe HTML affiche leur nombre sans dessiner de liens individuels incertains. Consultez la [référence du format de sortie](../../reference/output-format/) pour la grammaire et la [référence de support des langages](../../reference/language-support/) pour les limites de résolution.

## Protection contre les sorties obsolètes

Si le répertoire de sortie contient déjà des fichiers `.skt` d'une exécution précédente, `codeknit` refusera d'écrire de nouvelles données pour éviter de mélanger des données obsolètes et fraîches.

Pour contourner ce comportement et nettoyer le répertoire de sortie avant l'écriture, utilisez le flag `--clean` :

```bash
codeknit parse ./src --clean
```

Cela garantit un ensemble de sortie frais et cohérent.

Lors de la mise à niveau vers `0.5.0`, régénérez l'ensemble complet de la sortie. Les IDs de symboles et les résultats de relations peuvent changer ; ne mélangez pas les chunks de différentes exécutions.

## Conseils

- ✅ **Utilisez `directory-flat` par défaut** pour la plupart des projets. Il offre un bon équilibre entre lisibilité et gestion.
- 🔍 Utilisez `--minify` sur les grandes bases de code pour réduire l'utilisation de tokens via un dictionnaire partagé (`dict.skt`).
- 🔗 La section `[edges]` est **exclue par défaut** pour économiser des tokens. Utilisez `--edges` lorsque vous avez besoin de données de relation comme `calls`, `contains` ou `inherits`.
- 🧾 Utilisez `--format json` lorsqu'un script ou une intégration nécessite des données structurées au lieu de `.skt`.
- 🧹 Utilisez toujours `--clean` lorsque vous réexécutez sur le même répertoire de sortie.
- 📁 Utilisez `directory-tree` si vous souhaitez corréler les fichiers `.skt` directement avec les fichiers source dans votre éditeur.