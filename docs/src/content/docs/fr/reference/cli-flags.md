---
title: Référence de l'interface en ligne de commande
description: Référence complète pour toutes les commandes et options de codeknit.
---

## codeknit

Lance l'interface utilisateur terminal interactive (TUI), qui vous guide à travers les commandes et options disponibles.

```bash
codeknit
```

## codeknit parse

Extrait les informations structurelles du code source dans des fichiers `.skt`.

```bash
codeknit parse <input-path> [output-dir]
```

| Flag             | Type   | Défaut           | Description                                                                                                |
| ---------------- | ------ | ---------------- | ---------------------------------------------------------------------------------------------------------- |
| `--output-mode`  | string | `directory-flat` | Mode de sortie : `inline`, `directory-flat` ou `directory-tree`                                            |
| `--max-lines`    | int    | `500`            | Nombre maximal de lignes par fichier de sortie (s'applique aux modes `directory-flat` et `directory-tree`) |
| `--collect-test` | bool   | `false`          | Inclure les fichiers de test dans l'analyse                                                                |
| `--minify`       | bool   | `false`          | Activer la minification de la sortie basée sur un dictionnaire                                             |
| `--edges`        | bool   | `false`          | Inclure la section `[edges]` dans la sortie (désactivé par défaut pour économiser des tokens)              |
| `--clean`        | bool   | `false`          | Supprimer les fichiers `.skt` obsolètes du répertoire de sortie avant l'écriture                           |
| `--workers`      | int    | `0` (NumCPU)     | Nombre maximal de goroutines de parsing concurrentes                                                       |
| `--verbose`      | bool   | `false`          | Afficher les informations de progression pendant le traitement                                             |

Le répertoire de sortie par défaut est `./skeleton` lorsqu'il n'est pas spécifié. En mode `inline`, la sortie est écrite sur stdout et aucun répertoire n'est utilisé.

## codeknit graph show

Génère une visualisation interactive en HTML de la structure du codebase sous forme de graphe.

```bash
codeknit graph show <input-path>
```

| Flag             | Type   | Défaut                           | Description                                                    |
| ---------------- | ------ | -------------------------------- | -------------------------------------------------------------- |
| `-o`, `--output` | string | `./skeleton/codeknit-graph.html` | Chemin du fichier HTML de sortie                               |
| `--collect-test` | bool   | `false`                          | Inclure les fichiers de test dans l'analyse                    |
| `--workers`      | int    | `0` (NumCPU)                     | Nombre maximal de goroutines de parsing concurrentes           |
| `--verbose`      | bool   | `false`                          | Afficher les informations de progression pendant le traitement |

Le fichier HTML généré est autonome et s'ouvre automatiquement dans votre navigateur par défaut.

## codeknit graph analyze

Exécute des algorithmes d'analyse de graphe et émet un rapport `.skt` lisible par un LLM.

```bash
codeknit graph analyze <input-path>
```

| Flag                      | Type    | Défaut                          | Description                                                                      |
| ------------------------- | ------- | ------------------------------- | -------------------------------------------------------------------------------- |
| `-o`, `--output`          | string  | `./skeleton/graph_analysis.skt` | Chemin du fichier `.skt` de sortie                                               |
| `--collect-test`          | bool    | `false`                         | Inclure les fichiers de test dans l'analyse                                      |
| `--workers`               | int     | `0` (NumCPU)                    | Nombre maximal de goroutines de parsing concurrentes                             |
| `--verbose`               | bool    | `false`                         | Afficher les informations de progression pendant le traitement                   |
| `--fan-threshold`         | int     | `10`                            | Nombre minimal de fan-in ou fan-out pour signaler un symbole hub                 |
| `--god-threshold`         | int     | `15`                            | Nombre minimal d'arêtes de type "contient" pour signaler une god class/fonction  |
| `--max-inheritance-depth` | int     | `5`                             | Signale les chaînes d'héritage plus profondes que cette valeur                   |
| `--top-n`                 | int     | `30`                            | Limite les sections de sortie classées ; `0` signifie aucune limite              |
| `--betweenness-threshold` | float64 | `0.001`                         | Valeur minimale de centralité d'intermédiarité à rapporter                       |
| `--propagation-cutoff`    | float64 | `0.05`                          | Probabilité minimale pour continuer la simulation de propagation des changements |

## codeknit fingerprint

Détecte les doublons et quasi-doublons de code en utilisant le fuzzy hashing.

```bash
codeknit fingerprint <input-path>
```

| Flag               | Type   | Défaut                        | Description                                                                                                                                      |
| ------------------ | ------ | ----------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| `-o`, `--output`   | string | `./skeleton/fingerprints.skt` | Chemin du fichier `.skt` de sortie                                                                                                               |
| `--min-similarity` | int    | `65`                          | Pourcentage minimal de similarité à rapporter (0–100)                                                                                            |
| `--max-similarity` | int    | `95`                          | Pourcentage maximal de similarité à rapporter (0–100)                                                                                            |
| `--show-all`       | bool   | `false`                       | Inclure la section `[fingerprints]` avec les données brutes de tokens                                                                            |
| `--rerank`         | bool   | `false`                       | Reclasser les candidats CTPH en utilisant des embeddings sémantiques via Ollama (nécessite `ollama serve` et `ollama pull qwen3-embedding:0.6b`) |
| `--model`          | string | `qwen3-embedding:0.6b`        | Modèle d'embedding Ollama à utiliser avec `--rerank`                                                                                             |
| `--collect-test`   | bool   | `false`                       | Inclure les fichiers de test dans l'analyse                                                                                                      |
| `--workers`        | int    | `0` (NumCPU)                  | Nombre maximal de goroutines de parsing concurrentes                                                                                             |
| `--verbose`        | bool   | `false`                       | Afficher les informations de progression pendant le traitement                                                                                   |

## codeknit completion

Génère des scripts de complétion pour les shells pris en charge.

```bash
codeknit completion <shell>
```

Shells pris en charge : `bash`, `zsh`, `fish`, `powershell`.

## Flags globaux

| Flag           | Description                              |
| -------------- | ---------------------------------------- |
| `--version`    | Affiche les informations de version      |
| `--help`, `-h` | Affiche l'aide pour la commande en cours |
