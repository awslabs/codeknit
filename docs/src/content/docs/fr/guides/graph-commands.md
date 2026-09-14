---
title: Commandes de graphe
description: Visualisez et analysez la structure de votre base de code avec des algorithmes de graphe.
---

codeknit propose des commandes de graphe pour visualiser la structure, exécuter des analyses automatisées et combiner le graphe de dépendances actuel avec l'historique des modifications Git.

## graph show

Génère une visualisation interactive de graphe en HTML de votre base de code.

```bash
codeknit graph show <input-path>
```

Cette commande analyse votre base de code et produit un fichier HTML autonome avec une visualisation interactive de graphe. Les symboles (fonctions, classes, types) apparaissent comme des nœuds, et leurs relations (appels, contient, implémente) comme des arêtes. La visualisation s'ouvre automatiquement dans votre navigateur par défaut.

Le visualiseur affiche les relations résolues. Il indique un compte d'incertitudes mais ne dessine pas les relations non résolues ou ambiguës individuelles ni leurs liens candidats. Inspectez ces relations avec la sortie de parse :

```bash
codeknit parse ./src --output-mode inline --edges
codeknit parse ./src --output-mode inline --format json --edges
```

### Flags

| Flag             | Défaut                          | Description                                  |
| ---------------- | -------------------------------- | -------------------------------------------- |
| `-o`, `--output` | `./skeleton/codeknit-graph.html` | Chemin du fichier HTML de sortie            |
| `--collect-test` | `false`                          | Inclure les fichiers de test dans l'analyse |
| `--workers`      | `NumCPU`                         | Nombre maximal de goroutines de parsing concurrentes |
| `--verbose`      | `false`                          | Afficher des informations de progression pendant le traitement |

### Exemples

```bash
# Generate default visualization
codeknit graph show ./myproject

# Custom output file
codeknit graph show ./myproject -o graph.html

# Include test files
codeknit graph show ./src --collect-test
```

## graph analyze

Exécute des algorithmes de graphe structurels sur votre base de code et génère un rapport `.skt` lisible par LLM contenant des insights sur la qualité du code.

```bash
codeknit graph analyze <input-path>
```

Cette commande signale des candidats pour la revue, y compris les dépendances cycliques, les symboles hub, le code mort, les god classes et les goulots d'étranglement architecturaux.

### Algorithmes

L'analyse inclut 22 algorithmes de graphe structurels :

- Dépendances cycliques (SCC de Tarjan)
- Détection de hubs (couplage élevé fan-in/fan-out)
- Détection d'orphelins (candidats de code mort)
- Détection de god class/function (enfants excessifs)
- Métrique d'instabilité (Ce/(Ca+Ce) de Robert C. Martin)
- Chaînes d'héritage profondes
- Centralité de betweenness (détection de goulots d'étranglement)
- Points d'articulation (points uniques de défaillance)
- PageRank (importance récursive)
- Fan-in transitif (rayon d'impact)
- Simulation de propagation de changements
- Dépendances cycliques de paquets
- Détection de violations de couche
- Accessibilité depuis les points d'entrée
- Composantes faiblement connectées
- Poids des dépendances (force de couplage des paquets)
- Distance de la Main Sequence (équilibre A+I)
- Détection de shotgun surgery
- Détection de feature envy
- Violations de dépendances stables
- Violations de ségrégation d'interfaces
- Profondeur de containment

### Interprétation du rapport

Les métriques de dépendance utilisent des relations résolues. Le rapport compte les relations non résolues et ambiguës exclues. Les métadonnées de fichier/inclusion sont exclues des classements de dépendances et des découvertes de code mort.

- Le fan-in et le fan-out comptent les voisins distincts. Les relations répétées n'augmentent pas ces comptes, et le containment n'est pas une dépendance.
- La profondeur d'héritage considère toutes les bases et interfaces. Les hiérarchies cycliques ne se voient pas attribuer de profondeur.
- L'accessibilité commence à partir des callables nommés `main`, `Main` ou `init` et suit les arêtes de dépendance. Atteindre un membre rend son conteneur pertinent sans marquer chaque frère accessible. Sans point d'entrée reconnu, les découvertes d'accessibilité sont omises.
- Les arguments de callback et les callables retournés sont des références, pas la preuve que le callback s'exécute.
- La propagation de changements utilise le chemin de dépendance le plus fort. Ses scores et ses règles de couche basées sur les répertoires sont des heuristiques, pas des probabilités mesurées.

Les appelants externes, le dispatch dynamique et les dépendances en dehors de l'analyse peuvent changer les conclusions. Voir [Prise en charge des langages et fidélité des API](../../reference/language-support/) pour les règles et limites de résolution.

### Flags

| Flag                      | Défaut                         | Description                                              |
| ------------------------- | ------------------------------- | -------------------------------------------------------- |
| `-o`, `--output`          | `./skeleton/graph_analysis.skt` | Chemin du fichier `.skt` de sortie                       |
| `--collect-test`          | `false`                         | Inclure les fichiers de test dans l'analyse              |
| `--workers`               | `NumCPU`                        | Nombre maximal de goroutines de parsing concurrentes     |
| `--verbose`               | `false`                         | Afficher des informations de progression pendant le traitement |
| `--fan-threshold`         | `10`                            | Seuil minimal de fan-in ou fan-out pour signaler un symbole hub |
| `--god-threshold`         | `15`                            | Nombre minimal d'arêtes de containment pour signaler une god class/function |
| `--max-inheritance-depth` | `5`                             | Signaler les chaînes d'héritage plus profondes que cette valeur |
| `--top-n`                 | `30`                            | Limiter les sections de sortie classées ; 0 = pas de limite |
| `--betweenness-threshold` | `0.001`                         | Valeur minimale de centralité de betweenness à rapporter |
| `--propagation-cutoff`    | `0.05`                          | Score heuristique minimal pour poursuivre la propagation de changements |

### Exemples

```bash
# Run structural analysis with defaults
codeknit graph analyze ./myproject

# Custom output and thresholds
codeknit graph analyze ./myproject -o analysis.skt --fan-threshold 15

# Show more results per section
codeknit graph analyze ./myproject --top-n 50

# Remove the cap on ranked sections
codeknit graph analyze ./myproject --top-n 0

# Include test files
codeknit graph analyze ./src --collect-test
```

## graph hotspots

Classe les fichiers qui sont à la fois fréquemment modifiés et structurellement importants :

```bash
codeknit graph hotspots <input-path>
```

Le score combine la fréquence des commits, le churn des lignes et la récence avec le PageRank au niveau des fichiers, le fan-in transitif et la centralité de betweenness. Le rapport identifie également le couplage temporel entre les fichiers qui changent répétitivement dans les mêmes commits.

Les commits de merge sont exclus par défaut. Les commits modifiant plus de 50 fichiers sont également exclus afin que les changements générés, vendored ou mécaniques en masse ne faussent pas les résultats.

### Flags

| Flag                     | Défaut                   | Description                                     |
| ------------------------ | ------------------------- | ----------------------------------------------- |
| `-o`, `--output`         | `./skeleton/hotspots.skt` | Chemin du fichier de sortie                     |
| `--format`               | `skt`                     | Format de sortie : `skt` ou `json`              |
| `--since`                | `12mo`                    | Fenêtre d'historique, telle que `180d`, `12mo`, ou `2y` |
| `--max-commits`          | `2000`                    | Nombre maximal de commits à inspecter          |
| `--max-files-per-commit` | `50`                      | Exclure les commits modifiant plus de fichiers |
| `--min-cochanges`        | `3`                       | Nombre minimal de commits partagés pour le couplage temporel |
| `--top-n`                | `30`                      | Nombre maximal de résultats par section de rapport |
| `--include-merges`       | `false`                   | Inclure les commits de merge                    |
| `--collect-test`         | `false`                   | Inclure les fichiers de test                    |
| `--workers`              | `NumCPU`                  | Nombre maximal de goroutines de parsing concurrentes |
| `--verbose`              | `false`                   | Afficher des informations de progression        |

### Exemples

```bash
# Analyze the last 12 months
codeknit graph hotspots ./myproject

# Analyze two years and emit JSON
codeknit graph hotspots ./myproject --since 2y --format json -o hotspots.json

# Include larger commits and require stronger coupling
codeknit graph hotspots . --max-files-per-commit 100 --min-cochanges 5
```