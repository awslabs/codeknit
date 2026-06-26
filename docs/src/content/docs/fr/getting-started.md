---
title: Démarrer
description: Prenez en main codeknit en moins de 5 minutes.
---

# Démarrer

Prenez en main codeknit en moins de 5 minutes.

## 1. Prérequis

Vous aurez besoin de :

- Go 1.26+
- Un compilateur C (CGo est requis pour tree-sitter)

## 2. Installation depuis les sources

```bash
git clone https://github.com/awslabs/codeknit.git
cd codeknit
make build
# Le binaire se trouve dans ./bin/codeknit
```

## 3. Ajouter au PATH

Ajoutez le binaire au PATH de votre shell :

```bash
# bash (~/.bashrc)
export PATH="$PATH:$(pwd)/bin"

# zsh (~/.zshrc)
export PATH="$PATH:$(pwd)/bin"

# fish (~/.config/fish/config.fish)
fish_add_path $(pwd)/bin
```

Rechargez votre shell ou exécutez `source ~/.bashrc` (ou `~/.zshrc`) pour que la modification soit prise en compte.

## 4. Vérifier l'installation

Vérifiez que codeknit fonctionne :

```bash
codeknit --version
```

## 5. Première analyse

Exécutez votre première analyse sur une base de code :

```bash
codeknit parse ./myproject
```

Cette commande :

- Analyse tous les fichiers source dans `./myproject`
- Extrait les informations structurelles (fonctions, classes, relations)
- Écrit les fichiers `.skt` découpés dans `./skeleton/` (répertoire de sortie par défaut)

Si vous relancez cette commande, utilisez `--clean` pour supprimer les résultats précédents :

```bash
codeknit parse ./myproject --clean
```

## 6. Lecture de la sortie

Les fichiers `.skt` contiennent des informations structurées sur le code. Voici un petit exemple :

```skt
[symbols]
## src/main.go
S1 module/package L1-L1 main {}
S2 type/struct L5-L8 Server {exported}
S3 callable/function L10-L12 NewServer(addr: string) -> *S2 {exported}
S4 callable/method L14-L19 Start() {receiver=*Server}

[edges]
S2 --contains--> S4
S3 --returns--> S2
```

Sections clés :

- `[symbols]` : Définitions regroupées par fichier, montrant le nom, la plage de lignes et les métadonnées
- `[edges]` : Relations comme `contains`, `calls`, `inherits` ou `returns`

## 7. Étapes suivantes

Maintenant que vous avez exécuté votre première analyse :

- En savoir plus sur la commande parse : [Guide de la commande parse](/codeknit/fr/guides/parse-command/)
- Explorer l'analyse structurelle : [Guide des commandes graph](/codeknit/fr/guides/graph-commands/)
- Comprendre la détection des doublons : [Guide de la commande fingerprint](/codeknit/fr/guides/fingerprint-command/)
- Lire le format de sortie complet : [Référence du format de sortie](/codeknit/fr/reference/output-format/)
- Voir tous les indicateurs disponibles : [Référence des indicateurs CLI](/codeknit/fr/reference/cli-flags/)