---
title: Commandes de graphe
description: Visualisez et analysez la structure de votre base de code avec des algorithmes de graphe.
---

codeknit propose deux commandes de graphe puissantes pour vous aider à comprendre et améliorer la structure de votre base de code : `graph show` pour la visualisation interactive et `graph analyze` pour l'analyse structurelle automatisée.

## graph show

Génère une visualisation interactive de graphe au format HTML de votre base de code.

```bash
codeknit graph show <input-path>
```

Cette commande analyse votre base de code et produit un fichier HTML autonome avec une visualisation interactive de graphe. Les symboles (fonctions, classes, types) apparaissent sous forme de nœuds, et leurs relations (appels, contient, implémente) sous forme d'arêtes. La visualisation s'ouvre automatiquement dans votre navigateur par défaut.

### Flags

| Flag             | Défaut                          | Description                                  |
| ---------------- | -------------------------------- | -------------------------------------------- |
| `-o`, `--output` | `./skeleton/codeknit-graph.html` | Chemin du fichier HTML de sortie             |
| `--collect-test` | `false`                          | Inclure les fichiers de test dans l'analyse  |
| `--workers`      | `NumCPU`                         | Nombre maximal de goroutines de parsing concurrentes |
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

Exécute des algorithmes de graphe structurels sur votre base de code et génère un rapport `.skt` lisible par LLM contenant des informations sur la qualité du code.

```bash
codeknit graph analyze <input-path>
```

Cette commande détecte des problèmes courants de qualité du code tels que les dépendances cycliques, les symboles hub, le code mort, les god classes et les goulots d'étranglement architecturaux.

### Algorithmes

L'analyse inclut 17 algorithmes de graphe structurels :

- Dépendances cycliques (Tarjan's SCC)
- Détection de hubs (couplage élevé fan-in/fan-out)
- Détection d'orphelins (candidats au code mort)
- Détection de god class/function (nombre excessif d'enfants)
- Métrique d'instabilité (Ce/(Ca+Ce) de Robert C. Martin)
- Chaînes d'héritage profondes
- Centralité de betweenness (détection de goulots d'étranglement)
- Points d'articulation (points uniques de défaillance)
- PageRank (importance récursive)
- Fan-in transitif (rayon d'impact)
- Simulation de propagation de changements
- Dépendances circulaires de packages
- Détection de violations de couches
- Accessibilité depuis les points d'entrée
- Composants faiblement connectés
- Poids des dépendances (force de couplage des packages)
- Distance par rapport à la Main Sequence (équilibre A+I)

### Flags

| Flag                      | Défaut                         | Description                                              |
| ------------------------- | ------------------------------- | -------------------------------------------------------- |
| `-o`, `--output`          | `./skeleton/graph_analysis.skt` | Chemin du fichier `.skt` de sortie                       |
| `--collect-test`          | `false`                         | Inclure les fichiers de test dans l'analyse              |
| `--workers`               | `NumCPU`                        | Nombre maximal de goroutines de parsing concurrentes     |
| `--verbose`               | `false`                         | Afficher les informations de progression pendant le traitement |
| `--fan-threshold`         | `10`                            | Nombre minimal de fan-in ou fan-out pour signaler un symbole hub |
| `--god-threshold`         | `15`                            | Nombre minimal d'arêtes de type "contient" pour signaler une god class/function |
| `--max-inheritance-depth` | `5`                             | Signaler les chaînes d'héritage plus profondes que cette valeur |
| `--top-n`                 | `30`                            | Limiter les sections de sortie classées ; 0 = pas de limite |
| `--betweenness-threshold` | `0.001`                         | Valeur minimale de centralité de betweenness à rapporter |
| `--propagation-cutoff`    | `0.05`                          | Probabilité minimale pour continuer la propagation des changements |

### Exemples

```skt
# Exécuter une analyse structurelle avec les valeurs par défaut
codeknit graph analyze ./myproject

# Sortie personnalisée et seuils
codeknit graph analyze ./myproject -o analysis.skt --fan-threshold 15

# Afficher plus de résultats par section
codeknit graph analyze ./myproject --top-n 50

# Inclure les fichiers de test
codeknit graph analyze ./src --collect-test
```