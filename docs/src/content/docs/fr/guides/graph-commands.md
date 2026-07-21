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

Cette commande analyse votre base de code et produit un fichier HTML autonome avec une visualisation interactive de graphe. Les symboles (fonctions, classes, types) apparaissent sous forme de nœuds, et leurs relations (appels, contient, implémente) sous forme d'arêtes. La visualisation s'ouvre automatiquement dans votre navigateur par défaut.

### Flags

| Flag             | Default                          | Description                                  |
| ---------------- | -------------------------------- | -------------------------------------------- |
| `-o`, `--output` | `./skeleton/codeknit-graph.html` | Chemin du fichier HTML de sortie             |
| `--collect-test` | `false`                          | Inclure les fichiers de test dans l'analyse  |
| `--workers`      | `NumCPU`                         | Nombre maximal de goroutines d'analyse concurrentes |
| `--verbose`      | `false`                          | Afficher les informations de progression pendant le traitement |

### Exemples

```skt
# Générer une visualisation par défaut
codeknit graph show ./myproject

# Fichier de sortie personnalisé
codeknit graph show ./myproject -o graph.html

# Inclure les fichiers de test
codeknit graph show ./src --collect-test
```

## graph analyze

Exécute des algorithmes de graphe structurels sur votre base de code et génère un rapport `.skt` lisible par LLM contenant des insights sur la qualité du code.

```bash
codeknit graph analyze <input-path>
```

Cette commande détecte des problèmes courants de qualité du code tels que les dépendances cycliques, les symboles hub, le code mort, les god classes et les goulots d'étranglement architecturaux.

### Algorithmes

L'analyse inclut 22 algorithmes de graphe structurels :

- Dépendances cycliques (Tarjan's SCC)
- Détection de hubs (couplage élevé fan-in/fan-out)
- Détection d'orphelins (candidats au code mort)
- Détection de god class/function (nombre excessif d'enfants)
- Métrique d'instabilité (Robert C. Martin's Ce/(Ca+Ce))
- Chaînes d'héritage profondes
- Centralité de betweenness (détection de goulots d'étranglement)
- Points d'articulation (points uniques de défaillance)
- PageRank (importance récursive)
- Fan-in transitif (rayon d'impact)
- Simulation de propagation des changements
- Dépendances circulaires entre packages
- Détection de violations de couches
- Accessibilité depuis les points d'entrée
- Composants faiblement connectés
- Poids des dépendances (force du couplage entre packages)
- Distance par rapport à la Main Sequence (équilibre A+I)
- Détection de shotgun surgery
- Détection de feature envy
- Violations de dépendances stables
- Violations de ségrégation d'interfaces
- Profondeur de containment

### Flags

| Flag                      | Default                         | Description                                              |
| ------------------------- | ------------------------------- | -------------------------------------------------------- |
| `-o`, `--output`          | `./skeleton/graph_analysis.skt` | Chemin du fichier `.skt` de sortie                       |
| `--collect-test`          | `false`                         | Inclure les fichiers de test dans l'analyse              |
| `--workers`               | `NumCPU`                        | Nombre maximal de goroutines d'analyse concurrentes      |
| `--verbose`               | `false`                         | Afficher les informations de progression pendant le traitement |
| `--fan-threshold`         | `10`                            | Fan-in ou fan-out minimum pour signaler un symbole hub   |
| `--god-threshold`         | `15`                            | Nombre minimal d'arêtes contains pour signaler une god class/function |
| `--max-inheritance-depth` | `5`                             | Signaler les chaînes d'héritage plus profondes que cette valeur |
| `--top-n`                 | `30`                            | Limiter les sections de sortie classées ; 0 = pas de limite |
| `--betweenness-threshold` | `0.001`                         | Valeur minimale de centralité de betweenness à rapporter |
| `--propagation-cutoff`    | `0.05`                          | Probabilité minimale pour poursuivre la propagation des changements |

### Exemples

```skt
# Exécuter une analyse structurelle avec les valeurs par défaut
codeknit graph analyze ./myproject

# Sortie et seuils personnalisés
codeknit graph analyze ./myproject -o analysis.skt --fan-threshold 15

# Afficher plus de résultats par section
codeknit graph analyze ./myproject --top-n 50

# Inclure les fichiers de test
codeknit graph analyze ./src --collect-test
```

## graph hotspots

Classe les fichiers qui sont à la fois fréquemment modifiés et structurellement importants :

```bash
codeknit graph hotspots <input-path>
```

Le score combine la fréquence des commits, le churn des lignes et la récence avec le PageRank au niveau des fichiers, le fan-in transitif et la centralité de betweenness. Le rapport identifie également le couplage temporel entre les fichiers qui changent répétitivement dans les mêmes commits.

Les commits de merge sont exclus par défaut. Les commits modifiant plus de 50 fichiers sont également exclus afin que les modifications générées, vendues ou mécaniques en masse ne faussent pas les résultats.

### Flags

| Flag                     | Default                   | Description                                      |
| ------------------------ | ------------------------- | ------------------------------------------------ |
| `-o`, `--output`         | `./skeleton/hotspots.skt` | Chemin du fichier de sortie                      |
| `--format`               | `skt`                     | Format de sortie : `skt` ou `json`               |
| `--since`                | `12mo`                    | Fenêtre d'historique, par exemple `180d`, `12mo`, ou `2y` |
| `--max-commits`          | `2000`                    | Nombre maximal de commits à inspecter           |
| `--max-files-per-commit` | `50`                      | Exclure les commits modifiant plus de fichiers  |
| `--min-cochanges`        | `3`                       | Nombre minimal de commits partagés pour le couplage temporel |
| `--top-n`                | `30`                      | Nombre maximal de résultats par section de rapport |
| `--include-merges`       | `false`                   | Inclure les commits de merge                     |
| `--collect-test`         | `false`                   | Inclure les fichiers de test                     |
| `--workers`              | `NumCPU`                  | Nombre maximal de goroutines d'analyse concurrentes |
| `--verbose`              | `false`                   | Afficher les informations de progression         |

### Exemples

```bash
# Analyser les 12 derniers mois
codeknit graph hotspots ./myproject

# Analyser deux ans et émettre du JSON
codeknit graph hotspots ./myproject --since 2y --format json -o hotspots.json

# Inclure des commits plus larges et exiger un couplage plus fort
codeknit graph hotspots . --max-files-per-commit 100 --min-cochanges 5
```