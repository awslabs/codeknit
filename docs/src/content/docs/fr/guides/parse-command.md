---
title: Commande Parse
description: Extraire des informations structurelles du code source dans des fichiers .skt ou JSON.
---

La commande `codeknit parse` extrait des informations structurelles de votre base de code — telles que les fonctions, classes, méthodes, variables et leurs relations — et les sortie dans un format compact `.skt` par défaut. Utilisez JSON lorsque vous avez besoin d'une sortie lisible par machine pour des scripts, des intégrations ou des outils en aval.

## Utilisation de base

```bash
codeknit parse <input-path> [output-dir]
```

- **`<input-path>`** : Chemin vers le répertoire ou le fichier que vous souhaitez analyser.
- **`[output-dir]`** : Répertoire de sortie optionnel. Si non fourni, la valeur par défaut est `./skeleton`.

### Exemples

```bash
# Analyser un projet, sortie dans le répertoire par défaut ./skeleton
codeknit parse ./src

# Analyser et écrire dans un répertoire de sortie personnalisé
codeknit parse ./src ./output

# Analyser un seul fichier et sortie vers stdout
codeknit parse ./src/main.go --output-mode inline

# Émettre du JSON lisible par machine vers stdout
codeknit parse ./src --output-mode inline --format json
```

## Modes de sortie

Utilisez `--output-mode` pour contrôler la structure de la sortie. Trois modes sont disponibles :

| Mode             | Description                                                                              | Meilleur pour                                            |
| ---------------- | ---------------------------------------------------------------------------------------- | -------------------------------------------------------- |
| `directory-flat` | Écrit des fichiers `.skt` segmentés (par exemple `map_001.skt`, `map_002.skt`) dans le répertoire de sortie. | ✅ **La plupart des projets** — mode par défaut et recommandé |
| `directory-tree` | Reproduit la structure du répertoire source, créant un fichier `.skt` par fichier source. | Naviguer dans la sortie aux côtés du code source         |
| `inline`         | Affiche toute la sortie sur stdout.                                                      | Fichiers uniques ou redirection vers d'autres outils     |

> **Astuce** : Utilisez `directory-flat` par défaut, sauf si vous travaillez avec un seul fichier. Évitez `inline` pour les entrées volumineuses car cela peut submerger les fenêtres de contexte.

## Flags

| Flag             | Défaut          | Description                                                                  |
| ---------------- | ---------------- | ---------------------------------------------------------------------------- |
| `--output-mode`  | `directory-flat` | Mode de sortie : `inline`, `directory-flat` ou `directory-tree`              |
| `--format`       | `skt`            | Format de sortie : `skt` ou `json`                                           |
| `--max-lines`    | `500`            | Nombre maximal de lignes par fichier de sortie en modes flat/tree            |
| `--collect-test` | `false`          | Inclure les fichiers de test dans l'analyse                                  |
| `--minify`       | `false`          | Activer la compression basée sur un dictionnaire pour réduire l'utilisation des tokens |
| `--edges`        | `false`          | Inclure la section `[edges]` avec les données de relation (appels, contient, etc.) |
| `--clean`        | `false`          | Supprimer les fichiers `.skt` existants dans le répertoire de sortie avant l'écriture |
| `--workers`      | `NumCPU`         | Nombre maximal de goroutines d'analyse concurrentes (0 = utiliser tous les cœurs CPU) |
| `--verbose`      | `false`          | Afficher les informations de progression et de timing pendant le traitement |

## Schémas courants

```bash
# Première exécution sur un projet
codeknit parse ./src
```

```bash
# Réexécution et nettoyage de la sortie précédente
codeknit parse ./src --clean
```

```bash
# Analyser un seul fichier vers stdout
codeknit parse ./src/main.go --output-mode inline
```

```bash
# Minifier la sortie pour les grandes bases de code
codeknit parse ./src --minify
```

```bash
# Inclure les arêtes de relation (par exemple, pour l'analyse de dépendances)
codeknit parse ./src --edges
```

```bash
# Émettre du JSON pour un autre outil
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
# Reproduire la structure de l'arborescence source dans la sortie
codeknit parse ./src --output-mode directory-tree
```

## Protection contre les sorties obsolètes

Si le répertoire de sortie contient déjà des fichiers `.skt` d'une exécution précédente, `codeknit` refusera d'écrire de nouvelles données pour éviter de mélanger des données obsolètes et fraîches.

Pour contourner ce comportement et nettoyer le répertoire de sortie avant l'écriture, utilisez le flag `--clean` :

```bash
codeknit parse ./src --clean
```

Cela garantit un ensemble de sortie frais et cohérent.

## Conseils

- ✅ **Utilisez `directory-flat` par défaut** pour la plupart des projets. Il offre un bon équilibre entre lisibilité et gestion.
- 🔍 Utilisez `--minify` sur les grandes bases de code pour réduire l'utilisation des tokens via un dictionnaire partagé (`dict.skt`).
- 🔗 La section `[edges]` est **exclue par défaut** pour économiser des tokens. Utilisez `--edges` lorsque vous avez besoin de données de relation comme `calls`, `contains` ou `inherits`.
- 🧾 Utilisez `--format json` lorsqu'un script ou une intégration a besoin de données structurées au lieu de `.skt`.
- 🧹 Utilisez toujours `--clean` lorsque vous réexécutez sur le même répertoire de sortie.
- 📁 Utilisez `directory-tree` si vous souhaitez corréler les fichiers `.skt` directement avec les fichiers source dans votre éditeur.