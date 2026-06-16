---
title: Utilisation avec les assistants IA
description: Configurez codeknit en tant que compétence pour Kiro, Claude Code et d'autres assistants de codage IA.
---

codeknit est livré avec des compétences prêtes à l'emploi qui enseignent aux assistants de codage IA comment l'utiliser efficacement. Ces compétences permettent aux assistants d'extraire la structure du code, de détecter les doublons et d'effectuer des analyses structurelles sans invitation manuelle.

## Aperçu des compétences

codeknit fournit deux compétences :

- **`codeknit-parse`** : Apprend aux assistants à extraire la structure du code (fonctions, classes, méthodes, variables) et les relations (appels, héritage, inclusion) dans des fichiers `.skt`.
- **`codeknit-fingerprint`** : Apprend aux assistants à détecter les doublons et les quasi-doublons de code en utilisant le fuzzy hashing.

Chaque compétence inclut une documentation que l'assistant lit à la demande pour comprendre l'utilisation, les flags, les modes de sortie et les workflows.

## Installation

Copiez les répertoires de compétences dans le dossier des compétences de votre assistant.

Pour **Kiro** :

```bash
cp -r skills/codeknit-parse ~/.kiro/skills/codeknit-parse
cp -r skills/codeknit-fingerprint ~/.kiro/skills/codeknit-fingerprint
```

Pour **Claude Code** :

```bash
cp -r skills/codeknit-parse ~/.claude/skills/codeknit-parse
cp -r skills/codeknit-fingerprint ~/.claude/skills/codeknit-fingerprint
```

Après l'installation, l'assistant sait automatiquement comment invoquer les commandes codeknit, sélectionner les flags appropriés et interpréter la sortie `.skt`.

## Ce que chaque compétence enseigne

### codeknit-parse

La compétence `codeknit-parse` apprend aux assistants à :

- Exécuter `codeknit parse` avec les flags appropriés pour différents scénarios
- Choisir le bon mode de sortie :
  - `directory-flat` (par défaut) pour la plupart des projets
  - `inline` pour les fichiers uniques ou les petites entrées
  - `directory-tree` pour refléter la structure source
- Lire et interpréter les fichiers de sortie `.skt`, y compris les sections `[symbols]`, `[edges]` et la section optionnelle `[dict]`
- Utiliser les données structurelles pour le refactoring, la cartographie des dépendances et la revue de code
- Exécuter `codeknit graph analyze` pour des insights plus profonds sur la qualité du code (dépendances cycliques, symboles hub, god classes, etc.)

### codeknit-fingerprint

La compétence `codeknit-fingerprint` apprend aux assistants à :

- Utiliser `codeknit fingerprint` pour la détection de doublons, les audits DRY et l'identification de refactoring
- Sélectionner des plages de similarité appropriées (`--min-similarity`, `--max-similarity`)
- Lire la section `[duplicates]` pour identifier les quasi-doublons de code
- Comprendre que les fingerprints mesurent la forme structurelle, pas l'intention sémantique
- Utiliser `--rerank` avec les embeddings Ollama pour réduire les faux positifs lorsque nécessaire

## Exemples de workflow

### Analyse structurelle

1. Demandez à l'assistant d'analyser la structure de votre base de code
2. Il exécute `codeknit parse ./src` et lit les fichiers `.skt` résultants
3. Il répond aux questions structurelles : dépendances, chaînes d'appels, code mort
4. Pour des insights plus profonds, il exécute `codeknit graph analyze ./src` et interprète le rapport

```skt
[symbols]
## src/service.go
S1 type/struct L5-L8 AuthService {}
S2 callable/method L10-L15 Authenticate(token: string) {receiver=*AuthService}

[edges]
S1 --contains--> S2
```

### Détection de doublons

1. Demandez à l'assistant de trouver du code dupliqué
2. Il exécute `codeknit fingerprint ./src`
3. Il lit la section `[duplicates]` dans la sortie
4. Il examine les paires signalées et propose une consolidation

```skt
[duplicates]
S1, S2: 87% similarité
S3, S4: 76% similarité
```

## Conseils

- **Lisez toujours les fichiers `.skt`, pas le code source brut, pour les questions structurelles** — ils contiennent la structure extraite dans un format compact et fiable
- Utilisez `codeknit graph analyze` pour découvrir des problèmes de qualité de code comme les dépendances cycliques, les symboles hub et les chaînes d'héritage profondes
- Exécutez `codeknit fingerprint` avant les grands refactoring pour identifier le code copié-collé qui devrait être consolidé
- Le format `.skt` est conçu pour être économe en tokens, ce qui le rend idéal pour les fenêtres de contexte des LLM
- Utilisez `--minify` pour réduire davantage l'utilisation des tokens lors du traitement de grandes bases de code
