---
title: Référence du format de sortie
description: Référence complète du format de sortie `.skt` utilisé par codeknit.
---

Cette page décrit la grammaire de la sortie `.skt` et la sortie JSON équivalente. Pour les règles qui décident comment les relations sont résolues, consultez [Prise en charge des langages et fidélité de l'API](../language-support/).

## Vue d'ensemble

Un fichier `.skt` est un fichier texte compact composé de sections. Chaque section commence par une ligne d'en-tête entre crochets.

| Section        | Présente lorsque                                      | Contenu                                          |
| -------------- | ----------------------------------------------------- | ------------------------------------------------ |
| `[symbols]`    | Toujours                                              | Définitions des symboles regroupées par fichier source |
| `[edges]`      | `--edges` est défini                                  | Relations entre les symboles                     |
| `[errors]`     | Une erreur de scan ou d'analyse s'est produite        | Fichiers qui n'ont pas pu être entièrement traités |
| `[dict]`       | `--minify` est défini                                 | Dictionnaire qui mappe des codes courts à des jetons répétés |

Dans les modes de sortie de répertoire, `[errors]` et `[dict]` peuvent être écrits dans les fichiers auxiliaires `warnings.skt` et `dict.skt`. En mode `inline`, ils apparaissent dans la même sortie que les autres sections.

`--format json` produit un document JSON avec les mêmes informations. Voir [Sortie JSON](#json-output).

## [symbols]

Les symboles sont regroupés par fichier source. Chaque groupe commence par une ligne d'en-tête `## <path>`, suivie d'une ligne par symbole.

```skt
[symbols]
## pkg/services/auth.go
S1 module/package L1-L1 services {}
S2 type/struct L5-L8 AuthService {exported}
```

### Ligne de symbole

```text
ShortID catégorie/type Ldébut-Lfin signature {propriétés}
```

| Champ          | Description                                                                                               |
| -------------- | --------------------------------------------------------------------------------------------------------- |
| `ShortID`      | Identifiant séquentiel : `S1`, `S2`, `S3`, etc. Les arêtes font référence aux symboles par cette valeur. |
| `catégorie/type` | Classification du symbole. Voir [Catégories et types](#catégories-et-types).                              |
| `Ldébut-Lfin`  | Première et dernière ligne du symbole dans le fichier source.                                            |
| `signature`    | Nom du symbole, avec paramètres et type de retour pour les callables. Voir [Formats de signature](#formats-de-signature). |
| `{propriétés}` | Métadonnées entre accolades. Voir [Propriétés](#propriétés).                                              |

Une ligne de symbole reste sur une seule ligne physique. Les sauts de ligne à l'intérieur d'une signature sont échappés en `\n` et `\r`.

Les identifiants de symboles peuvent changer lorsque l'extraction change. Régénérez tous les morceaux de sortie ensemble au lieu de mélanger des morceaux de différentes exécutions. Cela s'applique en particulier lors de la mise à niveau vers la version 0.5.0, car les identifiants de symboles et les résultats de résolution des relations peuvent changer.

### Catégories et types

| Catégorie   | Types                                      | Exemples                                           |
| ---------- | ------------------------------------------ | -------------------------------------------------- |
| `callable` | function, method, constructor              | `callable/function`, `callable/method`             |
| `type`     | class, interface, struct, enum             | `type/class`, `type/interface`                     |
| `value`    | variable, constant, field                  | `value/variable`, `value/constant`                 |
| `module`   | package, namespace                         | `module/package`                                   |
| `meta`     | type parameters, files, includes, metadata | `meta/type_parameter`, `meta/file`, `meta/include` |

### Formats de signature

| Forme                         | Utilisé pour                                  |
| ---------------------------- | ----------------------------------------- |
| `nom`                       | Types, valeurs et modules                |
| `nom(params)`               | Callables sans type de retour           |
| `nom(params) -> typeRetour` | Callables avec un type de retour        |

Les paramètres sont écrits comme suit :

| Forme          | Utilisé pour                  |
| ------------- | ------------------------- |
| `nom`        | Langages sans types   |
| `nom: type`  | Langages avec types      |

Lorsqu'une référence de type correspond à un symbole connu, le nom du type est remplacé par l'identifiant court de ce symbole. Par exemple, `-> *S2` fait référence à un pointeur vers le symbole `S2`.

### Propriétés

Les propriétés apparaissent entre accolades après la signature, séparées par des virgules. Des accolades vides (`{}`) signifient que le symbole n'a pas de propriétés.

| Forme de propriété        | Exemple                 |
| -------------------- | ----------------------- |
| Drapeau                 | `exported`, `async`, `static` |
| `clé=valeur`          | `visibility=public`, `receiver=*AuthService`, `exported=false`, `resolution=unresolved` |

### Exemple

```skt
[symbols]
## pkg/services/auth.go
S1 module/package L1-L1 services {}
S2 type/struct L5-L8 AuthService {exported}
S3 callable/function L10-L12 NewAuthService(secret: string, ttl: int) -> *S2 {exported}
S4 callable/method L14-L19 Authenticate(token: string) {exported, receiver=*AuthService}
S5 callable/function L29-L31 verifyToken(token: string) -> bool {exported=false}
```

## [edges]

La section `[edges]` répertorie les relations entre les symboles. Elle n'est écrite que lorsque les arêtes sont demandées avec `--edges`.

### Ligne d'arête

```text
IDSource --type--> IDCible1, IDCible2
IDSource --type[non résolu]--> Cible1, Cible2
IDSource --type[ambigu]--> Cible [candidats=IDCandidat1, IDCandidat2]
```

| Partie                     | Description                                                                           |
| ------------------------ | ------------------------------------------------------------------------------------- |
| `IDSource`                 | Extrémité source.                                                                      |
| `type`                   | Type de relation. Voir [Types d'arêtes](#types-darêtes).                                     |
| `[non résolu]`, `[ambigu]` | Marqueur de résolution optionnel. Voir [Marqueurs de résolution](#marqueurs-de-résolution).        |
| `IDCible1, IDCible2`     | Une ou plusieurs cibles, séparées par des virgules.                                             |
| `[candidats=...]`       | Cibles possibles pour une arête ambiguë, sous forme de liste séparée par des virgules entre crochets.    |

Les cibles partagent une ligne uniquement lorsque la source, le type, le marqueur de résolution et l'ensemble des candidats sont identiques. Une définition référencée par une arête peut apparaître dans un fichier ultérieur ou un morceau de sortie. Chargez tous les morceaux avant d'interpréter les arêtes.

### Marqueurs de résolution

| Marqueur         | Signification                                                                              |
| -------------- | ------------------------------------------------------------------------------------ |
| Aucun           | L'analyseur a résolu la relation.                                              |
| `[non résolu]` | L'analyseur n'a pas pu établir la dépendance.                                     |
| `[ambigu]`  | Plus d'une interprétation reste possible. Les identifiants de candidats listent les cibles possibles. |

Un marqueur manquant n'est pas une garantie du compilateur. Les identifiants de candidats identifient des cibles possibles, pas des connexions confirmées. L'incertitude est toujours écrite dans cette grammaire SKT, jamais sous forme de JSON intégré.

### Formes d'extrémité

| Forme                      | Utilisé lorsque                                                                                         | Exemple                     |
| ------------------------- | ------------------------------------------------------------------------------------------------- | --------------------------- |
| Identifiant court                  | L'extrémité ou le candidat est un symbole connu. Cela inclut les extrémités des arêtes non résolues et les symboles d'inclusion. | `S4`, `S8`       |
| Nom nu                 | La cible n'a pas de symbole. Le nom est relatif au fichier source.                                | `Refresh`, `string`         |
| Identifiant complet entre guillemets            | L'identité est complexe, ou un nom littéral ressemble à un identifiant court.                                  | `"app.go::S2"`              |

Les identifiants complets entre guillemets échappent les guillemets, les barres obliques inverses et les sauts de ligne. Un nom littéral tel que `S2` doit être écrit comme un identifiant complet entre guillemets pour qu'il ne soit pas lu comme un identifiant court.

```skt
S4 --references[non résolu]--> string, bool
S5 --calls[ambigu]--> Helper [candidats=S8, S9]
S6 --references[non résolu]--> "app.go::S2"
```

### Types d'arêtes

| Type         | Signification                                         |
| ------------ | ----------------------------------------------- |
| `calls`      | invocation de fonction/méthode                      |
| `contains`   | classe contient méthode, module contient fonction |
| `inherits`   | classe étend une autre                           |
| `implements` | classe implémente interface                      |
| `overrides`  | méthode surcharge parent                         |
| `references` | symbole référence un autre                       |
| `imports`    | module importe un autre                          |
| `decorates`  | décorateur appliqué à un symbole                 |

Les callables retournés et les arguments de callback sont enregistrés comme `references`, pas comme un type séparé.

### Symboles de fichier et d'inclusion pour C et C++

Les sources C et C++ produisent des symboles `meta/file` et `meta/include`, de sorte que les relations d'inclusion ont des identifiants courts. Chaque orthographe d'en-tête apparaît une fois dans la liste des symboles, et les arêtes utilisent les identifiants courts. Les inclusions répétées au sein d'un fichier partagent une extrémité.

```skt
[symbols]
## src/main.c
S1 meta/file L1-L4 main.c
S2 meta/include L1-L1 "config.h" {resolution=unresolved}
S3 meta/include L2-L2 "utils.h" {resolution=unresolved}
S4 meta/include L3-L3 <stdio.h> {resolution=unresolved}

[edges]
S1 --references[non résolu]--> S2, S3, S4
```

Un symbole d'inclusion enregistre la directive telle qu'elle est écrite. Il ne représente pas une déclaration d'en-tête résolue.

### Effet sur l'analyse et la visualisation

- Les arêtes incertaines et les métadonnées de fichier/inclusion ne contribuent pas aux métriques de dépendance. Les rapports montrent le nombre de relations exclues.
- Le graphe HTML dessine les liens résolus et affiche un compte des relations incertaines. Il ne dessine pas les arêtes incertaines individuelles.

### Exemple

```skt
[edges]
S2 --contains--> S4
S4 --calls--> S5
S10 --inherits--> S2
S24 --implements--> S19
```

## [errors]

La section `[errors]` répertorie les fichiers qui n'ont pas été entièrement traités. Elle n'est présente que lorsque des diagnostics existent.

```text
- <chemin>: <message>
```

```skt
[errors]
- path/to/broken.go: erreur de syntaxe à la ligne 42
```

## [dict]

La section `[dict]` apparaît uniquement dans la sortie minifiée (`--minify`). Elle mappe des codes courts à des jetons répétés.

| Partie       | Description                                                                 |
| ---------- | --------------------------------------------------------------------------- |
| `d0`, `d1`, … | Codes du dictionnaire.                                                        |
| Ligne d'entrée | `- <code>: <jeton>`                                                         |

Pour lire la sortie minifiée, remplacez chaque code par sa valeur de dictionnaire. Les codes peuvent remplacer les paires catégorie/type et les propriétés dans `[symbols]`. Ils peuvent également remplacer les types de relations dans `[edges]`. Les marqueurs d'incertitude restent lisibles. Par exemple, `S1 --d2[non résolu]--> S3` est une arête non résolue dont le type est la valeur de `d2`.

### Exemple minifié

```skt
[dict]
- d0: async=false
- d1: callable/method
- d2: exported

[symbols]
## src/handler.py
S1 type/class L1-L6 Handler {}
S2 d1 L2-L3 __init__(name) {d0}
S3 d1 L5-L6 handle(request) {d0}

[edges]
S1 --contains--> S2, S3
```

Ici, `S2` est un `callable/method` avec la propriété `async=false`.

## Exemple complet

La sortie suivante contient `[symbols]`, `[edges]` et `[errors]`. Elle n'est pas minifiée, donc elle n'a pas de section `[dict]` ; voir [l'exemple minifié](#exemple-minifié).

```skt
[symbols]
## pkg/services/auth.go
S1 module/package L1-L1 services {}
S2 type/struct L5-L8 AuthService {exported}
S3 callable/function L10-L12 NewAuthService(secret: string, ttl: int) -> *S2 {exported}
S4 callable/method L14-L19 Authenticate(token: string) {exported, receiver=*AuthService}
S5 callable/function L29-L31 verifyToken(token: string) -> bool {exported=false}
## pkg/util/hash.go
S6 callable/function L3-L5 Hash(data: string) -> string {exported}
## pkg/util/hash_alt.go
S7 callable/function L3-L5 Hash(data: string) -> string {exported}

[edges]
S1 --contains--> S2, S3, S4, S5
S2 --contains--> S4
S3 --references--> S2
S4 --calls--> S5
S4 --calls[non résolu]--> Refresh
S5 --calls[ambigu]--> Hash [candidats=S6, S7]

[errors]
- pkg/legacy/broken.go: erreur de syntaxe à la ligne 42
```

## Sortie JSON

`--format json`