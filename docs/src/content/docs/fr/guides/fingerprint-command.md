---
title: Commande Fingerprint
description: Détecter les doublons et quasi-doublons de code dans les fichiers et les langages à l'aide du hachage flou.
---

`codeknit fingerprint` calcule un hachage contextuel par morceaux déclenché (CTPH) sur les jetons normalisés du corps de chaque fonction, méthode, variable et type. Les noms de variables, les littéraux de chaîne et les annotations de type sont normalisés avant le hachage. Les copier-coller renommés et les ports vers d'autres langages peuvent donc toujours correspondre.

Ce guide montre comment trouver des doublons, ajuster la plage de similarité, réduire les faux positifs et décider des éléments à refactoriser.

## Avant de commencer

- `codeknit` est installé. Voir [Installation](../../installation/).
- Vous disposez d'un répertoire source à analyser, tel que `./src`.

## Trouver des doublons dans un projet

Exécutez la commande sur le répertoire que vous souhaitez vérifier :

```bash
codeknit fingerprint ./src
```

La commande écrit `./skeleton/fingerprints.skt` et signale les paires avec une similarité comprise entre 65 % et 95 %.

Ouvrez le fichier. La section `[duplicates]` est toujours présente, avec les correspondances les plus fortes en premier :

```skt
[duplicates]
# similarity range: 65%-95%
similarity:93%  pkg/user.go::GetUser <-> pkg/admin.go::GetAdmin
similarity:88%  src/auth.ts::validateToken <-> src/session.ts::validateSession
similarity:81%  lib/math.py::compute_tax <-> lib/math.py::compute_vat
```

Chaque ligne a la forme `similarity:<N>%  <fichier>::<symbole> <-> <fichier>::<symbole>`. Si rien ne correspond à la plage, la section contient uniquement `# no duplicates found`.

Pour la grammaire complète de la sortie, consultez la [référence du format de sortie](../../reference/output-format/).

## Trouver uniquement les doublons exacts

La plage par défaut s'arrête à 95 %, donc les correspondances à 100 % sont exclues. Pour lister les correspondances exactes, définissez les deux bornes à 100 :

```bash
codeknit fingerprint ./src --min-similarity 100 --max-similarity 100
```

"Exact" signifie que les flux de jetons normalisés complets sont identiques au sein de la même catégorie de symbole. Cela ne signifie pas que le texte source est identique, ni que les deux symboles sont prouvés équivalents.

## Élargir la plage pour trouver des correspondances plus lâches

Réduisez le minimum pour inclure du code qui partage une forme algorithmique mais diffère dans les détails :

```bash
codeknit fingerprint ./src --min-similarity 50 --max-similarity 80
```

Les résultats inférieurs à 50 % sont généralement du bruit.

## Réduire les faux positifs avec un réordonnancement sémantique

Si les résultats structurels sont bruyants, réordonnez les candidats avec des embeddings provenant d'un serveur Ollama local.

1. Démarrez Ollama et téléchargez le modèle d'embedding par défaut :

   ```bash
   ollama serve
   ollama pull qwen3-embedding:0.6b
   ```

2. Exécutez la commande avec `--rerank` :

   ```bash
   codeknit fingerprint ./src --rerank
   ```

Pour utiliser un autre modèle d'embedding, passez `--model` :

```bash
codeknit fingerprint ./src --rerank --model qwen3-embedding:4b
```

Le réordonnancement ajoute des candidats sémantiques et filtre sur un score combiné. Il réduit les faux positifs mais ne les élimine pas tous. Il peut également supprimer un candidat qui est structurellement exact, et la plage de similarité s'applique toujours. Les embeddings sont calculés à partir du même instantané de source analysé utilisé pour l'extraction.

## Inspecter les empreintes brutes

Pour voir l'empreinte de chaque symbole, ajoutez `--show-all` :

```bash
codeknit fingerprint ./src --show-all
```

Cela ajoute une section `[fingerprints]`, regroupée par fichier source, une ligne par symbole :

```skt
[fingerprints]
## src/auth.ts
validateToken  FP:3:a1b2c3...:d4e5f6...  tokens:8e0f1a2b...
```

Le format de la ligne est `<nom-du-symbole>  FP:<taille-bloc>:<hash1>:<hash2>  tokens:<hex>`. `tokens` est le flux de jetons normalisé encodé en hexadécimal.

## Écrire le rapport ailleurs

```bash
codeknit fingerprint ./src -o duplicates.skt
```

## Contrôler les fichiers analysés

Les fichiers de test sont ignorés par défaut. Incluez-les avec `--collect-test` :

```bash
codeknit fingerprint ./src --collect-test
```

Pour affiner l'analyse, utilisez les indicateurs de sélection de source. Placez les motifs entre guillemets pour que votre shell ne les développe pas :

```bash
codeknit fingerprint ./src --include '**/*.ts' --exclude '**/generated/'
```

Autres contrôles de sélection :

- `--max-file-size BYTES` ignore les fichiers plus volumineux. La valeur par défaut est `10485760` (10 Mio). `0` désactive la limite.
- `--strict` échoue avant d'écrire la sortie si l'analyse ou l'analyse syntaxique est incomplète.
- `--header-language c|cpp` sélectionne l'analyseur pour les fichiers `.h` ambigus. Voir [Prise en charge des langages](../../reference/language-support/#selection-de-la-langue-dentete).

Les liens symboliques ignorés, les fichiers trop volumineux et les avertissements de syntaxe apparaissent sur stderr. Les inclusions ne remplacent pas `.gitignore` ou le filtrage des tests, et les exclusions ont priorité sur les inclusions. Consultez la [référence de fiabilité de l'analyse](../../reference/analysis-reliability/) pour les garanties exactes.

## Choisir une plage de similarité

| Plage      | Ce qu'elle révèle généralement                                      |
| ---------- | -------------------------------------------------------------------- |
| `96-100%`  | Doublons structurels exacts ou quasi-exacts ; copier-coller évident |
| `85-95%`   | Quasi-doublons ; généralement copier-coller avec des modifications mineures |
| `65-84%`   | Plage par défaut ; forte similarité structurelle, souvent candidats à la refactorisation |
| `50-64%`   | Même forme algorithmique avec des détails différents                |
| En dessous de 50% | Généralement du bruit                                               |

Pour voir les deux premières lignes, augmentez `--max-similarity` à `100`.

## Décider quoi refactoriser

1. Commencez par la plage par défaut et lisez les paires les plus importantes.
2. Ouvrez les deux symboles de chaque paire et comparez-les.
3. Considérez une correspondance comme un signal, pas comme une preuve. Les empreintes mesurent la forme structurelle, pas les noms d'identificateurs ou l'intention. La même logique écrite de manière itérative et récursive obtient un score faible, et du code sans rapport avec la même forme peut obtenir un score élevé.
4. Les paires dans la plage 85–95 % sont généralement les candidats les plus sûrs à consolider. Les paires dans la plage 65–84 % nécessitent souvent un examen plus approfondi.
5. Si la liste est bruyante, réduisez la plage ou ajoutez `--rerank`.

Les correspondances inter-langages fonctionnent pour les constructions équivalentes. Les fonctionnalités spécifiques à un langage peuvent produire des correspondances de faible similarité erronées.

Les pourcentages de similarité sont une heuristique de classement, pas une distance d'édition. Le projet de calibration étiqueté rapporte la précision et le rappel séparément des scores de similarité. Il s'agit d'un contrôle de régression, pas d'une estimation de la précision sur votre code.

## Grands projets

- En dessous de 500 symboles éligibles, codeknit évalue chaque paire de même catégorie.
- À partir de 500 symboles ou plus, les groupes complets de jetons normalisés exacts sont toujours conservés. Les autres paires utilisent environ 50 voisins les plus proches par symbole, donc les résultats de quasi-doublons peuvent être incomplets.
- Pour les doublons exacts, utilisez la commande 100/100 ci-dessus. Elle n'est pas affectée par l'approximation.
- L'ordre est déterministe pour des entrées et des scores stables.

## Dépannage

| Symptôme | Cause et solution |
| -------- | ----------------- |
| `# no duplicates found` | Aucune paire ne correspond à la plage. Réduisez `--min-similarity` ou augmentez `--max-similarity`. |
| Une fonction courte connue est manquante | Les symboles avec moins de 4 jetons normalisés sont ignorés. |
| Des copies exactes sont manquantes | Le maximum par défaut est 95 %. Utilisez `--min-similarity 100 --max-similarity 100`. Avec `--rerank`, le filtre sémantique peut également les supprimer. |
| `--rerank` échoue | Vérifiez que `ollama serve` est en cours d'exécution et que le modèle est téléchargé. |
| Résultats partiels avec des avertissements sur stderr | Corrigez les fichiers signalés, ou ajoutez `--strict` pour échouer au lieu de publier une sortie incomplète. |

## Indicateurs utilisés dans ce guide

| Indicateur          | Valeur par défaut            | Description                                                                 |
| ------------------- | ---------------------------- | --------------------------------------------------------------------------- |
| `-o`, `--output`    | `./skeleton/fingerprints.skt` | Chemin du fichier `.skt` de sortie.                                        |
| `--min-similarity`  | `65`                         | Pourcentage de similarité minimum à signaler (0–100).                      |
| `--max-similarity`  | `95`                         | Pourcentage de similarité maximum à signaler (0–100).                      |
| `--show-all`        | `false`                      | Émet également la section `[fingerprints]` avec les données de jetons brutes. |
| `--rerank`          | `false`                      | Réordonne les candidats avec des embeddings sémantiques via Ollama. Nécessite `ollama serve` et `ollama pull qwen3-embedding:0.6b`. |
| `--model`           | `qwen3-embedding:0.6b`       | Modèle d'embedding Ollama à utiliser avec `--rerank`.                      |
| `--collect-test`    | `false`                      | Inclut les fichiers de test dans l'analyse.                                 |
| `--workers`         | `NumCPU`                     | Nombre maximal de goroutines d'analyse concurrentes.                        |
| `--verbose`         | `false`                      | Affiche les informations de progression.                                    |

Pour la liste complète, y compris `--include`, `--exclude`, `--max-file-size`, `--strict` et `--header-language`, consultez la [référence des indicateurs CLI](../../reference/cli-flags/).

## Voir aussi

- [Commande Parse](../parse-command/)
- [Commandes Graph](../graph-commands/)