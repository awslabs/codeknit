---
title: Commande Fingerprint
description: Détecter les doublons et quasi-doublons de code dans les fichiers et les langages à l'aide de hachage flou.
---

La commande `codeknit fingerprint` détecte les doublons et quasi-doublons de code dans votre base de code en utilisant le **Context-Triggered Piecewise Hashing (CTPH)**. Elle fonctionne entre les fichiers et même entre les langages de programmation en normalisant les noms de variables, les littéraux de chaîne et les annotations de type avant de calculer les empreintes structurelles normalisées.

## Ce qu'elle fait

`codeknit fingerprint` analyse chaque fonction, méthode, variable et type dans votre base de code et calcule une **empreinte structurelle normalisée** basée sur :

- Le flux de contrôle (`if`, `for`, `while`, `switch`)
- Les opérations (`=`, `+`, `==`, `&&`, `||`)
- Les appels, retours, affectations et création d'objets
- Les constructions du langage comme `try/catch`, `yield`, `await`, `defer`

Cette normalisation signifie que les **copier-coller renommés**, les **refactorisations triviales** et la **logique équivalente dans différents langages** peuvent encore être détectés comme des doublons.

L'algorithme utilise le **CTPH** (une variante de hachage roulant) pour trouver efficacement les quasi-doublons. Un code similaire produit des empreintes similaires, permettant une correspondance floue même lorsque le code a été légèrement modifié.

## Utilisation de base

```bash
codeknit fingerprint ./src
```

Cette commande :

- Analyse tous les fichiers source dans `./src`
- Calcule les empreintes structurelles
- Génère les résultats dans `./skeleton/fingerprints.skt`
- Signale les correspondances avec une similarité comprise entre **65 % et 95 %** (plage par défaut)

## Flags

| Flag               | Défaut                        | Description                                                                                                                                                |
| ------------------ | ----------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `-o`, `--output`   | `./skeleton/fingerprints.skt` | Chemin du fichier `.skt` de sortie                                                                                                                        |
| `--min-similarity` | `65`                          | Pourcentage de similarité minimum à signaler (0–100)                                                                                                      |
| `--max-similarity` | `95`                          | Pourcentage de similarité maximum à signaler (0–100)                                                                                                      |
| `--show-all`       | `false`                       | Inclure la section `[fingerprints]` avec les données brutes des jetons                                                                                     |
| `--rerank`         | `false`                       | Trouver les voisins sémantiques et réordonner les candidats à l'aide d'embeddings Ollama (nécessite : `ollama serve` et `ollama pull qwen3-embedding:0.6b`) |
| `--model`          | `qwen3-embedding:0.6b`        | Modèle d'embedding Ollama à utiliser avec `--rerank`                                                                                                       |
| `--collect-test`   | `false`                       | Inclure les fichiers de test dans l'analyse                                                                                                                |
| `--workers`        | `NumCPU`                      | Nombre maximal de goroutines de parsing concurrentes (0 = utiliser tous les cœurs CPU)                                                                     |
| `--verbose`        | `false`                       | Afficher les informations de progression pendant le traitement                                                                                             |

## Format de sortie

La sortie est un fichier `.skt` avec les sections suivantes :

### `[duplicates]` (toujours présente)

Liste les paires de symboles avec une similarité supérieure au seuil :

```skt
[duplicates]
similarity:96%  pkg/user.go::GetUser <-> pkg/admin.go::GetAdmin
similarity:88%  utils/str.go::TrimSpaces <-> lib/text.go::CleanString
```

Chaque ligne montre :

- Pourcentage de similarité
- Symbole de gauche (chemin du fichier, portée, nom)
- Symbole de droite (chemin du fichier, portée, nom)

### `[fingerprints]` (uniquement avec `--show-all`)

Contient les données brutes des empreintes pour chaque symbole :

```skt
[fingerprints]
validateToken  FP:3:a1b2c3...:d4e5f6...  tokens:8e0f1a2b...
```

Champs :

- Nom du symbole
- `FP:<version>:<hash1>:<hash2>` — empreinte CTPH
- `tokens:<hex>` — flux de jetons normalisé du corps

Cette section est utile pour le débogage ou la création d'outils en aval.

## Motifs courants

```bash
# Analyse par défaut
codeknit fingerprint ./codeknit/de/src
```

```bash
# Trouver uniquement les doublons exacts
codeknit fingerprint ./src --min-similarity 100
```

```bash
# Trouver du code modérément similaire (par exemple, même algorithme, noms différents)
codeknit fingerprint ./src --min-similarity 50 --max-similarity 80
```

```bash
# Utiliser la correspondance sémantique pour trouver des candidats supplémentaires et réduire les faux positifs
# Nécessite : ollama serve && ollama pull qwen3-embedding:0.6b
codeknit fingerprint ./src --rerank
```

```bash
# Utiliser un modèle d'embedding différent pour la correspondance sémantique
codeknit fingerprint ./src --rerank --model qwen3-embedding:4b
```

```bash
# Générer une liste complète des empreintes (pour les outils d'analyse)
codeknit fingerprint ./src --show-all
```

```bash
# Fichier de sortie personnalisé
codeknit fingerprint ./src -o duplicates.skt
```

## Choix d'une plage de similarité

| Plage    | Recommandation                                                                                     |
| -------- | -------------------------------------------------------------------------------------------------- |
| 96–100 % | Doublons structurels exacts ou quasi exacts. Très probablement du copier-coller.                   |
| 85–95 %  | Quasi-doublons. Généralement du copier-coller avec des modifications mineures (par exemple, variables renommées, ajout de logs). |
| 65–84 %  | Plage par défaut. Forte similarité structurelle. Bons candidats pour le refactoring.               |
| 50–64 %  | Similarité modérée. Même forme algorithmique mais détails différents. À examiner manuellement.     |
| < 50 %   | Généralement du bruit. Pas de duplication significative.                                           |

## Conseils

- **Les empreintes mesurent la structure, pas la signification** : Un score de similarité élevé signifie que le code _ressemble_ à un autre, pas qu'il _fait_ la même chose. Toujours examiner les deux symboles.
- **Utiliser `--rerank` pour la correspondance sémantique** : Les embeddings ajoutent des voisins sémantiques que la récupération structurelle peut manquer et filtrent les candidats qui ne concordent pas sémantiquement.
- **Les corps courts sont ignorés** : Les symboles avec moins de 4 jetons normalisés (par exemple, les accesseurs simples) sont ignorés pour éviter le bruit.
- **La correspondance inter-langages fonctionne** : Les constructions équivalentes (par exemple, une fonction Python et une fonction Go avec la même logique) peuvent correspondre, mais les motifs spécifiques à un langage peuvent produire des correspondances de faible similarité non pertinentes.
- **Une correspondance est un signal, pas un verdict** : Traitez chaque correspondance comme une invitation à enquêter — pas comme une preuve automatique de duplication.