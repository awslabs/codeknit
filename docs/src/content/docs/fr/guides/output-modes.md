---
title: Modes de sortie
description: Choisissez le bon mode de sortie pour la taille de votre projet et votre flux de travail.
---

codeknit sépare deux décisions : où la sortie est dirigée (le **mode de sortie**) et comment elle est encodée (le **format de sortie**). Cette page explique pourquoi chaque option existe, son coût et quand elle est adaptée. Pour la grammaire `.skt`, consultez la [référence du format de sortie](../../reference/output-format/). Pour les commandes, consultez le [guide de la commande parse](../parse-command/).

## Deux choix indépendants

Le drapeau `--output-mode` sélectionne l'un des trois modes : `directory-flat`, `directory-tree` ou `inline`. Le drapeau `--format` sélectionne `skt` (par défaut) ou `json`. N'importe quel mode peut être combiné avec l'un ou l'autre format.

Les deux choix sont séparés car ils répondent à des questions différentes. Le mode décide comment la sortie est divisée et située sur le disque, ce qui importe pour une personne ou un LLM qui la lit. Le format décide qui la consomme : un LLM ou un lecteur humain, ou un script.

## directory-flat

`directory-flat` est le mode par défaut. Il écrit des fichiers de morceaux numérotés, `map_001.skt`, `map_002.skt`, et ainsi de suite, dans `./skeleton/` sauf si vous spécifiez un autre répertoire de sortie.

**Pourquoi il existe.** Une grande base de code produit plus de structure que ce qui peut tenir dans une seule fenêtre de contexte. Le découpage en morceaux permet à un lecteur d'ouvrir uniquement les fichiers dont il a besoin au lieu de tout charger.

**Comment fonctionne le découpage.** codeknit conserve intact chaque bloc d'un fichier source et commence un nouveau morceau entre les blocs une fois qu'un morceau atteint environ `--max-lines` (par défaut 500). La limite est une cible, pas une borne stricte : un fichier source avec un très grand bloc peut produire un morceau plus long que 500 lignes. Conserver les blocs intacts signifie qu'un symbole n'est jamais séparé de ses frères dans le même fichier.

**Compromis.** Les morceaux ne sont pas organisés par chemin source. Pour trouver la structure d'un fichier, vous recherchez son en-tête `## source-path` dans les morceaux. Si vous consultez souvent des fichiers spécifiques, `directory-tree` convient mieux.

**Minification.** Avec `--minify`, le dictionnaire est écrit dans `dict.skt` à côté des morceaux. Voir [Minification](#minification).

## directory-tree

`directory-tree` reflète votre arborescence source sous le répertoire de sortie. Chaque fichier source obtient un fichier `.skt` au chemin correspondant.

**Nommage.** codeknit ajoute `.skt` au nom complet du fichier source plutôt que de remplacer l'extension. `src/foo.c` devient `src/foo.c.skt`, et `src/foo.h` devient `src/foo.h.skt`. Sans cette règle, les deux seraient mappés à `foo.skt` et entreraient en collision. Lorsque la sortie d'un fichier est divisée, les parties sont nommées `foo.c_part1.skt`, `foo.c_part2.skt`, et ainsi de suite. Les chemins de sortie en conflit, y compris les collisions de casse, échouent visiblement au lieu de s'écraser mutuellement.

**Le découpage diffère du mode plat.** En mode arborescence, `--max-lines` est une limite stricte sur chaque morceau source, avec un minimum de 3. Chaque morceau de symbole répète les en-têtes `[symbols]` et `## source-path` afin qu'il puisse être lu seul. Les morceaux contenant uniquement des arêtes portent `[edges]`. La limite s'applique aux morceaux source, pas aux fichiers auxiliaires tels que `dict.skt` et `warnings.skt`.

**Compromis.** La sortie en arborescence est facile à naviguer aux côtés du code source, mais elle produit de nombreux petits fichiers. Un grand projet peut en générer des milliers, et un lecteur qui souhaite une vue d'ensemble doit ouvrir de nombreux fichiers.

## inline

`inline` écrit tout sur stdout. Il ne crée aucun répertoire de sortie ni aucune zone de staging.

**Pourquoi il existe.** Il convient à deux cas : inspecter rapidement un seul fichier, et diriger la sortie vers un autre outil. C'est le seul mode où rien n'est écrit sur le disque.

**Compromis.** Rien n'est divisé. Le résultat complet atterrit dans le terminal ou dans le contexte du consommateur en une seule fois, ce qui submerge un lecteur sur tout ce qui dépasse une poignée de fichiers. Considérez-le comme adapté à un seul fichier ou à un très petit projet, soit environ moins de cinq fichiers. Comme stdout ne peut pas être annulé, un consommateur défaillant peut voir une sortie partielle. Les modes de répertoire n'ont pas ce problème.

## Formats de sortie

| Format | Meilleur pour | Sortie |
| --- | --- | --- |
| `skt` | Contexte LLM et inspection humaine | Fichiers `.skt` ou stdout |
| `json` | Scripts et intégration structurée | `codeknit.json` en modes répertoire, ou stdout en mode inline |

**skt** est compact et orienté ligne. Chaque symbole tient sur une ligne physique, avec les sauts de ligne dans les signatures échappés, de sorte qu'un lecteur peut charger une partie de la sortie sans en briser la structure. C'est pourquoi le découpage fonctionne pour `.skt`.

**json** existe pour les consommateurs qui ont besoin de champs typés plutôt que de texte à lire. Il s'agit d'un seul document avec `files`, `symbols` et, lorsque les arêtes sont incluses, `edges`. Comme il s'agit d'un seul document, il n'est pas divisé par nombre de lignes. Dans les modes répertoire, `--max-lines` n'a donc aucun effet sur JSON, et un grand projet produit un seul gros `codeknit.json`.

Un exemple compact :

```json
{
  "files": ["app.go"],
  "symbols": [
    {"id": "app.go::User", "short_id": "S1", "name": "User", "file": "app.go",
     "category": "type", "kind": "struct", "signature": "type User struct", "span": [3, 3]},
    {"id": "app.go::Save", "short_id": "S2", "name": "Save", "file": "app.go",
     "category": "callable", "kind": "function", "signature": "Save(u: S1)", "span": [5, 5]}
  ],
  "edges": [
    {"from": "app.go::Save", "from_short": "S2", "to": "app.go::User",
     "to_short": "S1", "kind": "references"}
  ]
}
```

## Choisir un mode

| Mode | Meilleur pour | Sortie |
| --- | --- | --- |
| `directory-flat` | La plupart des projets (par défaut) | `./skeleton/map_001.skt`, `map_002.skt`, ... |
| `directory-tree` | Naviguer aux côtés du code source | `./skeleton/<chemin reflété>.skt` |
| `inline` | Fichier unique, redirection | stdout |

Règles empiriques :

- **En cas de doute**, utilisez `directory-flat`, le mode par défaut.
- **Un seul fichier** : `inline` convient.
- **Au-delà de quelques fichiers** : utilisez `directory-flat` ou `directory-tree`.
- **Grandes bases de code** : ajoutez `--minify`.
- **Réexécution dans le même répertoire** : ajoutez `--clean`.

## Staging, réexécutions et `--clean`

Les modes répertoire n'écrivent pas dans le répertoire de sortie au fur et à mesure. codeknit génère la sortie dans un répertoire de staging frère et ne la publie qu'après une génération réussie. Une exécution échouée laisse donc votre sortie précédente en place.

Par défaut, codeknit refuse d'écrire lorsque la sortie générée pour le format sélectionné existe déjà. La raison en est que les identifiants de symboles (`S1`, `S2`, ...) peuvent changer à chaque modification de l'extraction. Mélanger des morceaux de différentes exécutions donnerait des arêtes pointant vers les mauvais symboles. `--clean` autorise le remplacement : après une génération réussie, les fichiers `.skt` obsolètes (en mode SKT) ou `codeknit.json` (en mode JSON) sont remplacés. Les entrées non liées dans le répertoire sont préservées. Si l'une d'elles entre en collision avec un chemin généré, l'exécution échoue au lieu de l'écraser.

La publication prend un verrou spécifique au répertoire, sauvegarde la sortie précédente et la restaure en cas d'échec du remplacement. Si le processus est interrompu entre les deux renommages de répertoire, la prochaine exécution restaure la sauvegarde avant de générer quoi que ce soit. Cela permet de récupérer une défaillance du processus. Cela ne rend pas la mise à jour du répertoire instantanée : il y a un bref intervalle entre les renommages, donc un lecteur qui a besoin d'un accès ininterrompu doit se coordonner avec la génération. Cela ne protège pas non plus contre une défaillance de stockage ou une coupure de courant.

Ne choisissez pas un répertoire source, ou un ancêtre de celui-ci, comme répertoire de sortie. Pour les garanties complètes de publication, consultez la [référence sur la fiabilité de l'analyse](../../reference/analysis-reliability/).

## Minification

`--minify` applique une compression basée sur un dictionnaire. codeknit remplace les jetons répétés, tels que `callable/method` ou `exported`, par des codes courts (`d0`, `d1`, ...). La correspondance est stockée dans une section `[dict]`, qui dans les modes répertoire est le fichier `dict.skt`.

L'économie provient de la répétition. La plupart des lignes de symboles partagent les mêmes catégories, types et propriétés, donc une grande base de code répète ces chaînes des milliers de fois. Le coût est l'indirection : un lecteur doit développer les codes à l'aide du dictionnaire, et un seul morceau n'est plus autonome. Le dictionnaire est une sortie auxiliaire distincte, donc il ne compte pas dans `--max-lines`.

La sortie minifiée encode le type de relation mais conserve les marqueurs de résolution lisibles, donc les arêtes `unresolved` et `ambiguous` restent distinguables. Pour les petits projets, les économies sont modestes et l'étape supplémentaire en vaut rarement la peine. Pour les grands projets, les économies en jetons peuvent être significatives.

## Pages connexes

- [Guide de la commande parse](../parse-command/) : les commandes pour chaque mode.
- [Référence du format de sortie](../../reference/output-format/) : la grammaire `.skt` et JSON.
- [Référence des drapeaux CLI](../../reference/cli-flags/) : chaque drapeau et valeur par défaut.
- [Référence sur la fiabilité de l'analyse](../../reference/analysis-reliability/) : nommage de la sortie et garanties de publication.