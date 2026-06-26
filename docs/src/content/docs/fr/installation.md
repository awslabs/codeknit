---
title: Installation
description: Comment installer codeknit sur votre système.
---

codeknit peut être installé à partir du code source. Les étapes suivantes vous guideront pour configurer codeknit sur votre système.

## À partir du code source

La méthode d'installation principale consiste à compiler à partir du code source. Vous aurez besoin de :

- Go 1.26+
- Un compilateur C (requis pour tree-sitter via CGo)

Clonez le dépôt et compilez le binaire :

```bash
git clone https://github.com/awslabs/codeknit.git
cd codeknit
make build
```

Le binaire compilé sera disponible à l'emplacement `./bin/codeknit`.

## Ajouter au PATH

Pour exécuter `codeknit` depuis n'importe quel répertoire, ajoutez l'emplacement du binaire à la variable PATH de votre système.

Pour **bash** (`~/.bashrc`) :

```bash
export PATH="$PATH:/chemin/vers/codeknit"
```

Pour **zsh** (`~/.zshrc`) :

```bash
export PATH="$PATH:/chemin/vers/codeknit"
```

Pour **fish** (`~/.config/fish/config.fish`) :

```fish
fish_add_path /chemin/vers/codeknit
```

Après avoir mis à jour la configuration de votre shell, rechargez-la en exécutant `source ~/.bashrc` (ou `~/.zshrc`) ou redémarrez votre terminal.

## Complétions de shell

codeknit prend en charge la complétion automatique pour les shells populaires. Installez les complétions en utilisant ces commandes :

Pour **bash** :

```bash
codeknit completion bash >> ~/.bashrc
```

Pour **zsh** :

```bash
codeknit completion zsh >> ~/.zshrc
```

Pour **fish** :

```bash
codeknit completion fish > ~/.config/fish/completions/codeknit.fish
```

Pour **PowerShell** :

```powershell
codeknit completion powershell >> $PROFILE
```

## Vérifier l'installation

Après l'installation, vérifiez que codeknit est correctement configuré :

```bash
codeknit --version
```

## Configuration pour le développement

Si vous contribuez à codeknit, exécutez ces commandes supplémentaires :

Installez les dépendances de développement :

```bash
make deps
```

Configurez les hooks git :

```bash
make setup
```

Exécutez la suite de tests :

```bash
make test
```