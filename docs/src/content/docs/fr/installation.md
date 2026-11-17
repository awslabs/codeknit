---
title: Installation
description: Comment installer codeknit sur votre système.
---

Ce guide installe codeknit à partir des sources, l'ajoute à votre `PATH`, active les complétions de shell et vérifie qu'il fonctionne.

:::tip[Utilisez un assistant de codage IA ?]
Les utilisateurs de Claude Code, Codex et Kiro peuvent installer le plugin codeknit à la place. Le plugin télécharge le binaire la première fois qu'il s'exécute. Consultez [Utiliser codeknit avec des assistants IA](../guides/ai-assistants/).
:::

## Installer depuis les sources

**Prérequis**

- Go 1.26 ou ultérieur
- Un compilateur C. codeknit utilise CGo pour construire ses analyseurs tree-sitter.

**Étapes**

```bash
git clone https://github.com/awslabs/codeknit.git
cd codeknit
make build
```

Le binaire est écrit dans `./bin/codeknit`.

### Utiliser un binaire précompilé à la place

Des binaires précompilés pour Linux, macOS et Windows (amd64 et arm64) sont disponibles sur la [page des releases](https://github.com/awslabs/codeknit/releases). Téléchargez l'archive pour votre plateforme, extrayez-la et déplacez le binaire dans un répertoire de votre `PATH`. Chaque release inclut un fichier `checksums.txt` que vous pouvez utiliser pour vérifier le téléchargement.

## Ajouter codeknit à votre PATH

Ignorez cette section si vous avez déplacé le binaire dans un répertoire déjà présent dans votre `PATH`, comme `/usr/local/bin`.

Remplacez `/path/to/codeknit` par le répertoire contenant le binaire. Si vous avez compilé depuis les sources, il s'agit du répertoire `bin` à l'intérieur du dépôt cloné.

Ajoutez la ligne correspondant à votre shell dans son fichier de configuration :

```bash
# bash (~/.bashrc)
export PATH="$PATH:/path/to/codeknit"

# zsh (~/.zshrc)
export PATH="$PATH:/path/to/codeknit"

# fish (~/.config/fish/config.fish)
fish_add_path /path/to/codeknit
```

Rechargez votre shell ou sourcez le fichier que vous avez édité :

```bash
source ~/.bashrc   # bash
source ~/.zshrc    # zsh
```

Pour fish, ouvrez un nouveau terminal.

## Activer les complétions de shell

codeknit génère des scripts de complétion pour bash, zsh, fish et PowerShell. Exécutez la commande correspondant à votre shell une fois :

```bash
# bash
codeknit completion bash >> ~/.bashrc

# zsh
codeknit completion zsh >> ~/.zshrc

# fish
codeknit completion fish > ~/.config/fish/completions/codeknit.fish

# PowerShell
codeknit completion powershell >> $PROFILE
```

Les commandes pour bash, zsh et PowerShell ajoutent le script à votre profil. Les exécuter à nouveau ajouterait le script une seconde fois, donc exécutez chacune d'elles une seule fois.

Ouvrez un nouveau shell ou sourcez votre profil pour activer les complétions.

## Vérifier l'installation

```bash
codeknit --version
```

Si la commande affiche une version, codeknit est installé. Si votre shell rapporte `command not found`, vérifiez que le répertoire dans votre ligne `PATH` contient le binaire `codeknit`, puis rechargez votre shell.

## Configurer un environnement de développement

Ne faites cela que si vous prévoyez de contribuer à codeknit. Exécutez ces commandes depuis la racine du dépôt cloné :

```bash
make deps    # installer les dépendances
make setup   # configurer les hooks git
make test    # exécuter la suite de tests
```

`make lint` exécute le linter et le formateur.

## Étapes suivantes

- Exécutez votre première commande dans [Prise en main](../getting-started/).
- Parcourez tous les flags dans la [référence des flags CLI](../reference/cli-flags/).