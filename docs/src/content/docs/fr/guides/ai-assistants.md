---
title: Utilisation avec les assistants IA
description: Installez le plugin codeknit pour Claude Code, Codex et Kiro.
---

Installez le plugin codeknit dans Claude Code, Codex ou Kiro, maintenez-le à jour et posez à l'assistant des questions sur la structure et les doublons de code. Le plugin contient une compétence, `codeknit`. La compétence indique à l'assistant quand exécuter `codeknit parse`, `codeknit graph analyze` et `codeknit fingerprint`, ainsi que la manière de lire la sortie `.skt`.

Vous n'avez pas besoin d'installer vous-même le binaire `codeknit`. La première fois que la compétence s'exécute, elle lance son script intégré `scripts/install-codeknit.sh`, qui effectue les actions suivantes :

1. Quitte immédiatement si `codeknit` est déjà dans votre `PATH` ou dans `~/.local/bin`.
2. Sinon, télécharge la dernière version GitHub pour votre système d'exploitation et votre architecture.
3. Vérifie le téléchargement par rapport au fichier `checksums.txt` de la version.
4. Installe le binaire dans `~/.local/bin` et affiche son chemin absolu sur la dernière ligne de la sortie.

Si `~/.local/bin` n'est pas dans votre `PATH`, l'assistant appelle codeknit via ce chemin complet pour le reste de la session.

Pour contrôler l'installation, définissez ces variables d'environnement :

| Variable | Effet |
| --- | --- |
| `CODEKNIT_VERSION` | Fixe une version, par exemple `0.6.0`. |
| `CODEKNIT_INSTALL_DIR` | Modifie l'emplacement d'installation. |

Pour installer le binaire vous-même, consultez la section [Installation](../../installation/).

## Installer le plugin

Choisissez la section correspondant à votre assistant.

### Claude Code

Exécutez ces commandes dans votre shell. Dans une session Claude Code, exécutez les mêmes commandes sous la forme `/plugin ...`.

```bash
claude plugin marketplace add awslabs/codeknit --sparse .claude-plugin skills
claude plugin install codeknit@codeknit
```

### Codex

```bash
codex plugin marketplace add awslabs/codeknit --sparse .agents --sparse .codex-plugin --sparse skills
codex plugin add codeknit@codeknit
```

### Kiro IDE

1. Ouvrez le panneau Powers.
2. Choisissez **Add Custom Power**, puis **Import power from GitHub**.
3. Saisissez `https://github.com/awslabs/codeknit`.

### Kiro CLI

La CLI Kiro s'installe depuis un chemin local, commencez donc par cloner le dépôt :

```bash
git clone --depth 1 --filter=blob:none --sparse https://github.com/awslabs/codeknit ~/.codeknit-power
git -C ~/.codeknit-power sparse-checkout set skills
kiro-cli powers install ~/.codeknit-power
```

### À propos de `--sparse`

L'option `--sparse` est facultative. Elle limite le clonage aux manifestes du plugin et à `skills/` au lieu de l'arborescence source complète.

### Installer la compétence manuellement

Si votre assistant lit les Agent Skills mais ne prend pas en charge les plugins, copiez `skills/codeknit/` dans le répertoire des compétences de votre assistant :

- `~/.claude/skills/`
- `~/.codex/skills/`
- `~/.kiro/skills/`

## Mettre à jour le plugin

| Assistant | Commande ou action |
| --- | --- |
| Claude Code | `claude plugin marketplace update codeknit` |
| Codex | `codex plugin marketplace upgrade` |
| Kiro IDE | Choisissez **Check for updates** dans le panneau Powers. |
| Kiro CLI | Exécutez `git -C ~/.codeknit-power pull`, puis relancez `kiro-cli powers install ~/.codeknit-power`. |

La mise à jour du plugin ne met pas à jour un binaire déjà installé. Pour changer la version du binaire, définissez `CODEKNIT_VERSION` comme décrit ci-dessus.

Si vous mettez à niveau codeknit et que l'extraction change, les IDs des symboles peuvent également changer. Demandez à l'assistant de régénérer toute la sortie plutôt que de mélanger des morceaux de différentes exécutions.

## Ce que la compétence apprend à l'assistant

La compétence donne à l'assistant ces règles, afin que vous n'ayez pas à les répéter dans chaque demande :

- **Choisir la commande en fonction de la question.**
  - `parse` répond à ce qui existe, où cela se trouve, et ce qui appelle, hérite ou contient quoi.
  - `graph analyze` trouve les cycles, les hubs, le code mort, les *god classes* et les goulots d'étranglement.
  - `fingerprint` trouve les copies-colles et les candidats à la refactorisation.
- **Utiliser par défaut le mode de sortie `directory-flat`.** L'assistant écrit les fichiers chunkés dans `./skeleton/` et ne lit que les chunks dont il a besoin. Il utilise le mode de sortie `inline` uniquement pour un seul fichier ou une poignée de fichiers, car `inline` déverse tout dans la fenêtre de contexte.
- **Lire la sortie `.skt`, pas le code source brut, pour les questions structurelles.** L'assistant n'ouvre les fichiers source que pour inspecter ou modifier les détails d'implémentation, en utilisant les plages de lignes de la sortie. Le fichier `references/output-format.md` de la compétence contient le format complet. Voir aussi la [référence du format de sortie](../../reference/output-format/).
- **Choisir des plages de similarité et traiter les correspondances comme des signaux.** Les empreintes comparent la forme structurelle, pas l'intention. L'assistant lit les deux symboles d'une paire avant de proposer une refactorisation.

## Demander une analyse structurelle

1. Demandez à l'assistant d'analyser la structure de votre base de code. Par exemple : "Cartographie la structure de `./src` et dis-moi quels modules dépendent du package auth."
2. L'assistant s'assure que le binaire est installé, puis exécute `codeknit parse` avec les arêtes activées et lit la sortie `.skt`.
3. Posez des questions de suivi sur les dépendances, les chaînes d'appels ou le code mort. L'assistant répond à partir de la sortie.
4. Pour une analyse plus approfondie, demandez une revue structurelle. Par exemple : "Trouve les dépendances cycliques et les *god classes* dans `./src`." L'assistant exécute `codeknit graph analyze` et lit le rapport.

Les relations marquées `[unresolved]` ou `[ambiguous]` dans la sortie sont incertaines, pas des dépendances confirmées. Traitez les réponses qui en dépendent comme des candidats à vérifier.

Pour les commandes derrière ces demandes, consultez le [guide de la commande parse](../parse-command/) et le [guide des commandes graph](../graph-commands/).

## Demander une détection de doublons

1. Demandez à l'assistant de trouver du code dupliqué. Par exemple : "Trouve les fonctions copiées-collées dans `./src`."
2. L'assistant exécute `codeknit fingerprint` et lit la section `[duplicates]` de `./skeleton/fingerprints.skt`.
3. L'assistant ouvre les deux symboles de chaque paire signalée, puis propose une refactorisation.
4. Si les résultats sont bruyants, demandez-lui d'ajuster la plage de similarité ou d'ajouter un réordonnancement sémantique. Le réordonnancement nécessite un serveur Ollama local avec un modèle d'embedding téléchargé.

Pour les options derrière ces demandes, consultez le [guide de la commande fingerprint](../fingerprint-command/).

## Conseils

- Rappelez à l'assistant de lire les fichiers `.skt`, pas le code source brut, pour les questions structurelles. Cela permet de garder un contexte réduit, car le format est compact et contient les symboles, les relations et les plages de lignes.
- Utilisez `graph analyze` pour obtenir des insights sur la qualité du code. Les résultats sont des candidats à la revue, car les appelants externes et la dispatch dynamique peuvent modifier le résultat.
- Exécutez `fingerprint` avant une refactorisation importante pour repérer les copies-colles qui devraient être consolidées.
- Si l'assistant relance `parse` dans un répertoire contenant déjà une sortie `.skt`, codeknit refuse d'écrire. Demandez-lui de passer `--clean` pour remplacer la sortie précédente.
- Si le téléchargement du binaire échoue (hors ligne ou plateforme non supportée), obtenez une version manuellement depuis la [page des versions](https://github.com/awslabs/codeknit/releases), ou consultez la section [Installation](../../installation/).
- Si une exécution rapporte des résultats partiels, vérifiez stderr et la section `[errors]`. Consultez la section [fiabilité de l'analyse](../../reference/analysis-reliability/) pour plus de détails.