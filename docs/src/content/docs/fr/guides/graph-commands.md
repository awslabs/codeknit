---
title: Commandes de graphe
description: Visualisez et analysez la structure de votre base de code avec des algorithmes de graphe.
---

Utilisez les sous-commandes `graph` pour voir comment votre code est connecté, identifier les problèmes structurels et décider quels fichiers réusiner en priorité. Ce guide suppose que vous avez [installé codeknit](../../installation/) et que vous pouvez exécuter [`codeknit parse`](../parse-command/).

| Tâche | Commande |
| --- | --- |
| Explorer le graphe visuellement | `codeknit graph show` |
| Identifier les problèmes structurels | `codeknit graph analyze` |
| Trouver les fichiers importants et volatils | `codeknit graph hotspots` |

## Visualiser le graphe de code

`graph show` analyse l'entrée et écrit un fichier HTML autonome. Les symboles (fonctions, classes, types) sont des nœuds. Les relations (appels, contient, implémente) sont des arêtes.

```bash
codeknit graph show ./myproject
```

Le fichier est écrit dans `./skeleton/codeknit-graph.html` et s'ouvre dans votre navigateur par défaut.

Pour choisir le chemin de sortie ou inclure les fichiers de test :

```bash
codeknit graph show ./myproject -o graph.html
codeknit graph show ./src --collect-test
```

| Flag | Défaut | Objectif |
| --- | --- | --- |
| `-o`, `--output` | `./skeleton/codeknit-graph.html` | Chemin du fichier HTML de sortie |
| `--collect-test` | `false` | Inclure les fichiers de test |
| `--workers` | `0` (`NumCPU`) | Nombre maximal de goroutines de parsing concurrentes |
| `--verbose` | `false` | Afficher les informations de progression |

### Inspecter les relations non résolues

Le visualiseur ne dessine que les relations résolues. Il affiche un compte des relations incertaines (`unresolved` ou `ambiguous`) mais ne les dessine pas ni leurs liens candidats.

Pour voir les arêtes incertaines individuelles, exécutez `parse` avec `--edges` :

```bash
codeknit parse ./src --output-mode inline --edges
```

Ajoutez `--format json` pour une sortie JSON. Consultez la [référence du format de sortie](../../reference/output-format/) pour savoir comment les arêtes et les candidats sont écrits.

## Analyser la structure et identifier les problèmes

`graph analyze` exécute 22 algorithmes de graphe et écrit un rapport `.skt` lisible par un LLM.

```bash
codeknit graph analyze ./myproject
```

Le rapport est écrit dans `./skeleton/graph_analysis.skt`. Lisez-le directement ou donnez-le à un LLM avec le code source pertinent.

### Ajuster le rapport

Augmentez un seuil pour voir moins de résultats, ou diminuez-le pour en voir plus :

```bash
codeknit graph analyze ./myproject -o analysis.skt --fan-threshold 15
codeknit graph analyze ./myproject --top-n 50
codeknit graph analyze ./src --collect-test
```

`--top-n 0` supprime la limite sur les sections classées.

| Flag | Défaut | Objectif |
| --- | --- | --- |
| `-o`, `--output` | `./skeleton/graph_analysis.skt` | Chemin du fichier `.skt` de sortie |
| `--collect-test` | `false` | Inclure les fichiers de test |
| `--workers` | `0` (`NumCPU`) | Nombre maximal de goroutines de parsing concurrentes |
| `--verbose` | `false` | Afficher les informations de progression |
| `--fan-threshold` | `10` | Seuil minimal de fan-in ou fan-out pour signaler un symbole hub |
| `--god-threshold` | `15` | Nombre minimal d'arêtes de type contient pour signaler une god class ou une fonction |
| `--max-inheritance-depth` | `5` | Signaler les chaînes d'héritage plus profondes que cette valeur |
| `--top-n` | `30` | Limiter les sections classées ; `0` signifie aucune limite |
| `--betweenness-threshold` | `0.001` | Centralité d'intermédiarité minimale à rapporter |
| `--propagation-cutoff` | `0.05` | Score heuristique minimal pour poursuivre la propagation des changements |

### Savoir quelle réponse correspond à quelle question

| Algorithme | Utilisez-le pour trouver |
| --- | --- |
| Dépendances cycliques (SCC de Tarjan) | Groupes de symboles qui dépendent les uns des autres en boucle |
| Détection de hubs | Symboles avec un couplage fan-in ou fan-out élevé |
| Détection d'orphelins | Candidats de code mort |
| Détection de god class/fonction | Symboles qui contiennent un nombre excessif d'enfants |
| Métrique d'instabilité | Symboles et paquets fragiles ou stables, en utilisant Ce/(Ca+Ce) |
| Chaînes d'héritage profondes | Hiérarchies plus profondes que `--max-inheritance-depth` |
| Centralité d'intermédiarité | Symboles goulot d'étranglement par lesquels passent de nombreux chemins de dépendance |
| Points d'articulation | Points uniques de défaillance dont la suppression divise le graphe |
| PageRank | Symboles récursivement importants |
| Fan-in transitif | Rayon d'impact : tout ce qui dépend d'un symbole, directement ou indirectement |
| Simulation de propagation des changements | Jusqu'où un changement d'un symbole est susceptible de se propager |
| Dépendances circulaires de paquets | Boucles de dépendance au niveau des paquets |
| Détection de violation de couche | Dépendances qui traversent les couches dans le mauvais sens |
| Accessibilité depuis les points d'entrée | Code qu'aucun point d'entrée n'atteint |
| Composantes faiblement connectées | Clusters de code isolés |
| Poids des dépendances | Force du couplage entre les paquets |
| Distance de la séquence principale | Paquets dont l'abstraction et l'instabilité sont déséquilibrées |
| Détection de chirurgie à la mitraillette | Changements qui toucheraient de nombreux endroits dans un paquet |
| Détection d'envie de fonctionnalité | Symboles qui utilisent un autre paquet plus que le leur |
| Violations de dépendance stable | Paquets stables qui dépendent de paquets moins stables |
| Violations de ségrégation d'interface | Interfaces dont les membres sont majoritairement inutilisés par leurs utilisateurs |
| Profondeur de containment | Symboles profondément imbriqués |

### Lire les résultats correctement

Ces règles influencent la confiance que vous pouvez accorder à un résultat :

- **Les relations incertaines sont exclues.** Les métriques de dépendance n'utilisent que les arêtes résolues. Le rapport compte les relations `unresolved` et `ambiguous` qu'il a laissées de côté.
- **Les métadonnées de fichier et d'inclusion sont ignorées.** Les symboles `meta/file` et `meta/include` n'affectent pas les classements de dépendance ou les résultats de code mort.
- **Le fan-in et le fan-out comptent les voisins distincts.** Le containment n'est pas considéré comme une dépendance.
- **La profondeur d'héritage couvre toutes les bases et interfaces.** Les hiérarchies cycliques ne se voient pas attribuer de profondeur.
- **L'accessibilité commence à `main`, `Main` ou aux callables `init`** et suit les arêtes de dépendance. Un membre accessible rend son conteneur pertinent, mais pas ses membres frères. Si l'analyse ne trouve aucun point d'entrée reconnu, les résultats d'accessibilité sont omis.
- **Les callbacks sont des références, pas des appels.** Un argument de callback ou un callable retourné ne prouve pas que le callback s'exécute.
- **Les scores sont des heuristiques.** La propagation des changements utilise le chemin de dépendance le plus fort. Ses scores et les règles de couche ne sont pas des probabilités mesurées ou des garanties de compilateur.

### Agir en toute sécurité sur les résultats

Considérez chaque résultat comme un candidat à examiner. Avant de modifier le code, ouvrez les symboles nommés dans le rapport et vérifiez :

- Les appelants externes en dehors du répertoire analysé.
- Le dispatch dynamique.
- Les dépendances que vous n'avez pas analysées.

L'analyse est basée sur la syntaxe, donc chacun de ces éléments peut changer la conclusion. Consultez [la résolution des relations dans la référence de support des langages](../../reference/language-support/#relationship-resolution) pour les règles et limites de résolution.

## Trouver les points chauds sensibles à l'historique

`graph hotspots` classe les fichiers qui changent souvent et occupent des positions structurellement importantes. Il rapporte également le couplage temporel : les fichiers qui changent répétitivement dans les mêmes commits. Exécutez-le sur un répertoire à l'intérieur d'un dépôt Git.

```bash
codeknit graph hotspots ./myproject
```

Le rapport est écrit dans `./skeleton/hotspots.skt`.

Il combine deux groupes de métriques :

- **Historique :** fréquence des commits, churn des lignes ajoutées et supprimées, et récence.
- **Structure :** PageRank au niveau des fichiers, fan-in transitif et centralité d'intermédiarité.

Le rapport inclut un niveau de confiance basé sur le nombre de commits pertinents analysés. Considérez les classements d'un rapport à faible confiance comme provisoires.

### Modifier la fenêtre d'historique et les filtres

```bash
# Analyser deux ans et écrire en JSON
codeknit graph hotspots ./myproject --since 2y --format json -o hotspots.json

# Autoriser des commits plus larges et exiger un couplage temporel plus fort
codeknit graph hotspots . --max-files-per-commit 100 --min-cochanges 5
```

Les commits de fusion sont exclus sauf si vous passez `--include-merges`. Les commits qui touchent plus de fichiers que `--max-files-per-commit` sont ignorés en tant que commits en masse, donc les changements générés ou mécaniques ne faussent pas le classement.

| Flag | Défaut | Objectif |
| --- | --- | --- |
| `-o`, `--output` | `./skeleton/hotspots.skt` | Chemin du fichier de sortie |
| `--format` | `skt` | Format de sortie : `skt` ou `json` |
| `--since` | `12mo` | Fenêtre d'historique, comme `180d`, `12mo`, ou `2y` |
| `--max-commits` | `2000` | Nombre maximal de commits à inspecter |
| `--max-files-per-commit` | `50` | Exclure les commits qui modifient plus de fichiers que cette valeur |
| `--min-cochanges` | `3` | Nombre minimal de commits partagés pour rapporter le couplage temporel |
| `--top-n` | `30` | Nombre maximal de résultats par section |
| `--include-merges` | `false` | Inclure les commits de fusion dans les métriques d'historique |
| `--collect-test` | `false` | Inclure les fichiers de test |
| `--workers` | `0` (`NumCPU`) | Nombre maximal de goroutines de parsing concurrentes |
| `--verbose` | `false` | Afficher les informations de progression |

### Lire la sortie JSON

Avec `--format json`, chaque point chaud a `score`, `history_score`, `structure_score`, `recency_score`, `pagerank`, `betweenness`, `commits`, `churn`, et `transitive_fan_in`. Chaque entrée de couplage temporel a `left`, `right`, `strength`, et `co_changes`. Le niveau supérieur enregistre également `commits_visited`, `commits_analyzed`, `skipped_merges`, et `skipped_bulk_commits`. Utilisez ces comptes pour vérifier sur quelle quantité d'historique le classement est basé.

## Contrôler les sources analysées

Les cinq commandes de traitement des sources (`parse`, `fingerprint`, `graph show`, `graph analyze`, et `graph hotspots`) acceptent les mêmes flags de sélection :

| Flag | Défaut | Objectif |
| --- | --- | --- |
| `--include GLOB` | Toutes les sources prises en charge | Inclure les chemins correspondants ; répétable |
| `--exclude GLOB` | Aucun | Exclure les chemins ou répertoires correspondants ; répétable. Les exclusions l'emportent sur les inclusions |
| `--max-file-size BYTES` | `10485760` (10 Mio) | Ignorer les fichiers plus volumineux ; `0` désactive la limite |
| `--strict` | `false` | Échouer avant d'écrire la sortie si l'analyse ou le parsing est incomplet |

Les motifs sont des globs de style gitignore relatifs au répertoire d'entrée. Mettez-les entre guillemets pour que votre shell ne les développe pas. Les inclusions ne remplacent pas les règles `.gitignore` ou le filtrage des tests.

```bash
codeknit graph analyze ./src --include '**/*.go' --exclude '**/generated/' --strict
```

Les liens symboliques et les fichiers non réguliers sont ignorés et signalés. Les problèmes d'analyse et de parsing apparaissent sur stderr. Sans `--strict`, codeknit conserve des résultats partiels utilisables, donc vérifiez stderr avant de faire confiance à un rapport. Avec `--strict`, une analyse incomplète s'arrête avant que toute sortie ne soit écrite. Une exécution où chaque lecture de source échoue échoue toujours.

Pour les garanties exactes, consultez la [référence de fiabilité de l'analyse](../../reference/analysis-reliability/).

## Pages connexes

- [Référence des flags CLI](../../reference/cli-flags/) : chaque flag pour chaque commande.
- [Référence du format de sortie](../../reference/output-format/) : syntaxe `.skt` et JSON.
- [Référence de support des langages](../../reference/language-support/) : limites d'extraction et de résolution.
- [Guide de la commande parse](../parse-command/) : extraire les symboles et les arêtes pour des fichiers individuels.