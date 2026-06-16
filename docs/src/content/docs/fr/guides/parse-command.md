---
title: Commande d'analyse
description: Extraire des informations structurelles du code source dans des fichiers .skt.
---

La commande `codeknit parse` extrait des informations structurelles de votre base de code — telles que les fonctions, classes, méthodes, variables et leurs relations — et les exporte dans un format `.skt` compact conçu pour une consommation efficace par les LLMs et les outils d'analyse.

## Utilisation de base

```bash
codeknit parse <input-path> [output-dir]
```

- **`<input-path>`** : Chemin vers le répertoire ou le fichier que vous souhaitez analyser.
- **`[output-dir]`** : Répertoire de sortie optionnel. Par défaut, `./skeleton` si non spécifié.

### Exemples

```bash
# Analyser un projet, sortie dans le répertoire par défaut ./skeleton
codeknit parse ./src

# Analyser et écrire dans un répertoire de sortie personnalisé
codeknit parse ./src ./output

# Analyser un seul fichier et afficher la sortie sur stdout
codeknit parse ./src/main.go --output-mode inline
```

## Modes de sortie

Utilisez `--output-mode` pour contrôler la structure de la sortie. Trois modes sont disponibles :

| Mode             | Description                                                                                              | Meilleur pour                                                 |
| ---------------- | -------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------- |
| `directory-flat` | Écrit des fichiers `.skt` segmentés (par ex. `map_001.skt`, `map_002.skt`) dans le répertoire de sortie. | ✅ **La plupart des projets** — mode par défaut et recommandé |
| `directory-tree` | Reproduit la structure du répertoire source, créant un fichier `.skt` par fichier source.                | Naviguer dans la sortie en parallèle du code source           |
| `inline`         | Affiche toute la sortie sur stdout.                                                                      | Fichiers uniques ou redirection vers d'autres outils          |

> **Astuce** : Privilégiez `directory-flat` sauf si vous travaillez avec un seul fichier. Évitez `inline` pour les gros volumes de données, car cela peut saturer les fenêtres de contexte.

## Options

| Option           | Valeur par défaut | Description                                                                           |
| ---------------- | ----------------- | ------------------------------------------------------------------------------------- |
| `--output-mode`  | `directory-flat`  | Format de sortie : `inline`, `directory-flat` ou `directory-tree`                     |
| `--max-lines`    | `500`             | Nombre maximal de lignes par fichier de sortie en modes flat/tree                     |
| `--collect-test` | `false`           | Inclure les fichiers de test dans l'analyse                                           |
| `--minify`       | `false`           | Activer la compression basée sur un dictionnaire pour réduire l'utilisation de tokens |
| `--edges`        | `false`           | Inclure la section `[edges]` avec les données de relations (appels, contenus, etc.)   |
| `--clean`        | `false`           | Supprimer les fichiers `.skt` existants dans le répertoire de sortie avant l'écriture |
| `--workers`      | `NumCPU`          | Nombre maximal de goroutines d'analyse concurrentes (0 = utiliser tous les cœurs CPU) |
| `--verbose`      | `false`           | Afficher les informations de progression et de timing pendant le traitement           |

## Motifs courants

```bash
# Première exécution sur un projet
codeknit parse ./src
```

```bash
# Réexécution avec nettoyage de la sortie précédente
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
# Inclure les arêtes de relation (par ex. pour l'analyse de dépendances)
codeknit parse ./src --edges
```

```bash
# Reproduire la structure de l'arborescence source dans la sortie
codeknit parse ./src --output-mode directory-tree
```

## Protection contre les sorties obsolètes

Si le répertoire de sortie contient déjà des fichiers `.skt` d'une exécution précédente, `codeknit` refusera d'écrire de nouvelles données pour éviter de mélanger des données obsolètes et fraîches.

Pour contourner ce comportement et nettoyer le répertoire de sortie avant l'écriture, utilisez l'option `--clean` :

```bash
codeknit parse ./src --clean
```

Cela garantit un ensemble de sortie frais et cohérent.

## Conseils

- ✅ **Privilégiez `directory-flat`** pour la plupart des projets. Il offre un bon équilibre entre lisibilité et gestion.
- 🔍 Utilisez `--minify` sur les grandes bases de code pour réduire l'utilisation de tokens via un dictionnaire partagé (`dict.skt`).
- 🔗 La section `[edges]` est **exclue par défaut** pour économiser des tokens. Utilisez `--edges` lorsque vous avez besoin de données de relations comme `calls`, `contains` ou `inherits`.
- 🧹 Utilisez toujours `--clean` lorsque vous réexécutez la commande sur le même répertoire de sortie.
- 📁 Utilisez `directory-tree` si vous souhaitez corréler les fichiers `.skt` directement avec les fichiers source dans votre éditeur.
