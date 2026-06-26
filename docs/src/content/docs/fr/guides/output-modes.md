---
title: Modes de sortie
description: Choisissez le mode de sortie adapté à la taille de votre projet et à votre flux de travail.
---

codeknit prend en charge trois modes de sortie, contrôlés par l'indicateur `--output-mode`. Chaque mode détermine comment la structure du code extraite est écrite sur le disque (ou stdout).

Le mode de sortie est distinct du format de sortie. Le format par défaut est `.skt` ; passez `--format json` pour émettre le même résultat d'analyse sous forme de JSON lisible par machine. Dans les modes de répertoire, le JSON est écrit dans `codeknit.json`. En mode `inline`, le JSON est écrit sur stdout.

### directory-flat (par défaut, recommandé)

- **Comportement** : Écrit des fichiers `.skt` segmentés tels que `map_001.skt`, `map_002.skt`, etc.
- **Répertoire de sortie** : `./skeleton/` par défaut
- **Segmentation** : Les fichiers sont segmentés lorsqu'ils dépassent la limite `--max-lines` (par défaut : 500 lignes)
- **Cas d'utilisation** : Idéal pour la plupart des projets. Maintient la sortie organisée et lisible en limitant la taille des fichiers. Vous pouvez lire uniquement les segments pertinents pour votre tâche.
- **Minification** : Lorsque `--minify` est activé, un fichier `dict.skt` est également généré dans le répertoire de sortie, contenant les mappages de tokens pour les valeurs compressées.

Exemple :

```bash
codeknit parse ./src
# Sortie : ./skeleton/map_001.skt, map_002.skt, ...
```

### directory-tree

- **Comportement** : Reproduit exactement la structure du répertoire source.
- **Répertoire de sortie** : `./skeleton/` par défaut
- **Mappage** : Un fichier `.skt` est créé par fichier source, à un chemin correspondant.
- **Cas d'utilisation** : Idéal lorsque vous souhaitez consulter rapidement la structure d'un fichier spécifique. Utile pour la navigation aux côtés du codebase d'origine.

Exemple :

```bash
codeknit parse ./src --output-mode directory-tree
# Sortie : ./skeleton/src/handler.skt, ./skeleton/pkg/db.skt, etc.
```

### inline

- **Comportement** : Affiche toute la sortie sur stdout.
- **Répertoire de sortie** : Aucun créé
- **Cas d'utilisation** : Recommandé uniquement pour les fichiers uniques ou les très petits projets (moins de 5 fichiers). Utile pour rediriger la sortie vers un autre outil ou inspecter un fichier unique de manière interactive.

Exemple :

```bash
codeknit parse ./src/main.go --output-mode inline
# Sortie : imprimée directement dans le terminal
```

### Format JSON

- **Comportement** : Émet un document JSON unique contenant `files`, `symbols`, `edges` optionnels et `errors` optionnels.
- **Emplacement de sortie** : `codeknit.json` dans les modes de répertoire, ou stdout en mode `inline`.
- **Cas d'utilisation** : Idéal pour les scripts, les intégrations d'éditeur, les vérifications CI et les outils nécessitant des données structurées.

Exemple :

```bash
codeknit parse ./src --output-mode inline --format json --edges
```

Exemple de sortie :

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

### Tableau de décision

| Mode             | Meilleur pour                                | Emplacement de sortie                                     |
| ---------------- | -------------------------------------------- | --------------------------------------------------------- |
| `directory-flat` | La plupart des projets (par défaut, recommandé) | `./skeleton/map_001.skt`, `map_002.skt`, ...              |
| `directory-tree` | Navigation de la sortie aux côtés du code source | `./skeleton/<chemin miroir>.skt`                          |
| `inline`         | Fichier unique, redirection vers un autre outil | stdout — à utiliser uniquement pour les fichiers uniques ou les très petits projets |

| Format | Meilleur pour                           | Sortie                                                   |
| ------ | --------------------------------------- | -------------------------------------------------------- |
| `skt`  | Contexte LLM et inspection humaine      | Fichiers `.skt` ou stdout                                |
| `json` | Scripts et intégration structurée       | `codeknit.json` dans les modes de répertoire, ou stdout en mode `inline` |

### Règles empiriques

- **En cas de doute** → utilisez `directory-flat` (par défaut)
- **Inspection d'un fichier unique** → `inline` est acceptable
- **Plus de quelques fichiers** → préférez `directory-flat` ou `directory-tree`
- **Grandes bases de code** → ajoutez `--minify` pour réduire l'utilisation de tokens
- **Réexécution sur la même sortie** → utilisez `--clean` pour supprimer les fichiers `.skt` obsolètes

### Minification

L'indicateur `--minify` active la compression basée sur un dictionnaire des tokens répétés (par exemple, les clés de propriété comme `exported`, `async`, ou les noms de types courants). Lorsqu'il est activé :

- Les valeurs répétées sont remplacées par des codes courts (`d0`, `d1`, `d2`, ...)
- Un fichier `dict.skt` est écrit dans le répertoire de sortie, mappant les codes aux valeurs d'origine
- Réduit considérablement la taille de la sortie pour les grandes bases de code
- Fonctionne dans les modes `directory-flat` et `directory-tree`

Exemple de sortie minifiée :

```skt
[dict]
- d0: exported
- d1: callable/function

[symbols]
## src/main.py
S1 d1 L1-L5 main() {d0}
```