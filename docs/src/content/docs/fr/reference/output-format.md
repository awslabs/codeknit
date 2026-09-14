---
title: Référence du format de sortie
description: Référence complète du format de sortie .skt utilisé par codeknit.
---

Le format `.skt` (skeleton) est un format de texte compact et lisible par l'humain utilisé par `codeknit` pour représenter la structure du code extraite. Il contient des symboles, des relations et des métadonnées sous une forme minimale adaptée à la consommation par les LLM et à l'analyse structurelle.

Un fichier `.skt` est divisé en sections. Chaque section commence par un en-tête entre crochets. Les sections peuvent apparaître dans n'importe quel ordre, bien que `[symbols]` vienne généralement en premier.

## [symbols]

La section `[symbols]` répertorie tous les symboles extraits, regroupés par leur fichier source. Chaque fichier est introduit par un en-tête `##` suivi du chemin du fichier.

### Format de ligne

Chaque symbole est représenté sur une seule ligne avec la structure suivante :

```
ShortID category/kind Lstart-Lend signature {properties}
```

### Champs

- **ShortID** : Un identifiant séquentiel attribué à chaque symbole (par exemple, `S1`, `S2`, `S3`). Utilisé comme référence dans les arêtes et autres sections.
- **catégorie/type** : Une paire séparée par une barre oblique indiquant la catégorie et le type spécifique du symbole.
- **Ldébut-Lfin** : La plage de lignes dans le fichier source où le symbole est défini (par exemple, `L10-L15`).
- **signature** : Le nom du symbole et les informations de type. Le format dépend du symbole :
  - `name` — pour les types, valeurs, modules
  - `name(params)` — pour les callables sans type de retour
  - `name(params) -> returnType` — pour les callables avec type de retour
- **{propriétés}** : Métadonnées optionnelles entre accolades. Plusieurs propriétés sont séparées par des virgules.

### Paramètres

- Dans les langages non typés : `paramName`
- Dans les langages typés : `paramName: type`
- Les références de type qui correspondent à des symboles connus sont remplacées par leurs ShortID (par exemple, `config: S5` au lieu de `config: Config`).

### Propriétés

Les propriétés courantes incluent :

- `async` : `true` ou `false`
- `exported` : `true` ou `false`
- `static` : présent si le symbole est statique
- `visibility=public|private|protected`
- `receiver=*TypeName` : pour les méthodes, indique le type récepteur

### Catégories et types de symboles

| Catégorie   | Types                                      | Exemples                                           |
| ---------- | ------------------------------------------ | -------------------------------------------------- |
| `callable` | function, method, constructor              | `callable/function`, `callable/method`             |
| `type`     | class, interface, struct, enum             | `type/class`, `type/interface`                     |
| `value`    | variable, constant, field                  | `value/variable`, `value/constant`                 |
| `module`   | package, namespace                         | `module/package`                                   |
| `meta`     | type parameters, files, includes, metadata | `meta/type_parameter`, `meta/file`, `meta/include` |

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

La section `[edges]` enregistre les relations lorsque `--edges` est activé. Les points d'extrémité connus utilisent leurs ShortID ; les cibles qui ne peuvent pas être résolues en symbole peuvent apparaître sous forme de noms.

### Format de ligne

```
FromID --kind--> ToID1, ToID2
FromID --kind[unresolved]--> Target1, Target2
FromID --kind[ambiguous]--> Target [candidates=CandidateID1, CandidateID2]
```

Un statut omis signifie que l'analyseur a résolu la relation. `unresolved` signifie qu'il n'a pas pu établir la dépendance ; `ambiguous` signifie que des interprétations concurrentes restent possibles. Les ID de candidats identifient des cibles possibles, pas des connexions confirmées. La résolution est basée sur la syntaxe et la portée, pas sur une garantie du compilateur.

Les cibles sont regroupées uniquement lorsque leur source, leur type, leur résolution et leur ensemble de candidats correspondent. Les noms sans symboles sont relatifs au fichier source. Les identités complexes utilisent des ID complets entre guillemets avec des guillemets, des barres obliques inverses et des sauts de ligne échappés. Un nom littéral qui ressemble à un ShortID doit être cité comme un ID complet, tel que `"app.go::S2"`. Les définitions de symboles peuvent apparaître dans des fichiers ultérieurs ou des morceaux de sortie.

```skt
S4 --references[unresolved]--> string, bool
S5 --calls[ambiguous]--> Helper [candidates=S8, S9]
S6 --references[unresolved]--> "app.go::S2"
```

### Points d'extrémité de fichiers et d'inclusions

Les inclusions C/C++ ont des symboles de métadonnées explicites. Les orthographes des en-têtes apparaissent une fois dans la liste des symboles, et les arêtes utilisent des ShortID même si la résolution des en-têtes reste non résolue :

```skt
[symbols]
## src/main.c
S1 meta/file L1-L3 main.c
S2 meta/include L1-L1 "config.h" {resolution=unresolved}
S3 meta/include L2-L2 <stdio.h> {resolution=unresolved}

[edges]
S1 --references[unresolved]--> S2, S3
```

Un symbole d'inclusion enregistre la directive écrite, pas une déclaration d'en-tête résolue. Les inclusions répétées au sein d'un fichier partagent un point d'extrémité. Les métadonnées de fichier/inclusion sont exclues des classements de dépendances et des découvertes de code mort.

Les relations incertaines sont préservées dans la sortie d'analyse mais exclues des métriques de dépendance. Les rapports de graphe affichent les comptes de relations exclues ; le visualiseur HTML montre leur nombre sans dessiner les arêtes individuelles incertaines.

Lors de la mise à niveau vers `0.5.0`, régénérez tous les morceaux de sortie ensemble. Les ID de symboles et les résultats de relations peuvent changer, donc les morceaux de différentes exécutions ne doivent pas être mélangés.

### Types d'arêtes

| Type         | Signification                                         |
| ------------ | ----------------------------------------------------- |
| `calls`      | invocation de fonction/méthode                        |
| `contains`   | classe contient méthode, module contient fonction     |
| `inherits`   | classe étend une autre classe                         |
| `implements` | classe implémente une interface                       |
| `overrides`  | méthode remplace la méthode parente                   |
| `references` | symbole référence un autre symbole                    |
| `imports`    | module importe un autre module                        |
| `decorates`  | décorateur appliqué à un symbole                      |

### Exemple

```skt
[edges]
S2 --contains--> S4
S4 --calls--> S5
S10 --inherits--> S2
S24 --implements--> S19
```

## [errors]

La section `[errors]` répertorie les fichiers qui n'ont pas pu être analysés complètement.

### Format

Chaque ligne commence par `-` suivi du chemin du fichier et du message d'erreur :

```
- path/to/file.go: syntax error at line 42
```

### Exemple

```skt
[errors]
- src/broken.go: unexpected token at line 10
- tests/corner_case.py: unterminated string literal
```

## [dict]

La section `[dict]` apparaît uniquement lorsque l'indicateur `--minify` est utilisé. Elle mappe des codes de dictionnaire courts (`d0`, `d1`, etc.) à des jetons de chaîne répétés pour réduire la taille de la sortie.

### Format

Chaque ligne mappe un code de dictionnaire (`d0`, `d1`, etc.) à sa valeur développée :

```
- d0: async=false
- d1: callable/method
- d2: exported
```

Dans le reste du fichier, ces codes remplacent leurs valeurs complètes.

Les types de relations peuvent également utiliser des codes de dictionnaire tandis que les étiquettes de statut restent lisibles : `S1 --d2[unresolved]--> S3`. Développez `d2` en utilisant le dictionnaire et conservez le statut non résolu.

### Exemple

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

## Exemple complet

```skt
[dict]
- d0: exported
- d1: callable/function

[symbols]
## main.go
S1 module/package L1-L1 main {}
S2 type/struct L5-L8 Server {d0}
S3 d1 L10-L12 NewServer(addr: string) -> *S2 {d0}
S4 callable/method L14-L20 Serve() {d0, receiver=*Server}
S5 callable/function L22-L25 handleError(err: error) -> bool {}

[edges]
S2 --contains--> S4
S4 --calls--> S5
S3 --references--> S2

[errors]
- utils/broken.go: syntax error at line 5
```

## Relations JSON

Utilisez JSON lorsqu'un script ou une intégration a besoin d'une sortie structurée :

```bash
codeknit parse ./src --output-mode inline --format json --edges
```

Chaque arête inclut `from`, `to` et `kind`. Les ID de source connus apparaissent dans `from_short` ; les ID de cible résolus apparaissent dans `to_short`. Les arêtes incertaines ajoutent `resolution` et, lorsqu'ils sont disponibles, `candidates` contenant les ID complets des symboles. Elles omettent `to_short`, même lorsque la cible a un symbole de métadonnées d'inclusion.

Par exemple, un appel ambigu peut être représenté comme suit :

```json
{
  "from": "app.py::run",
  "from_short": "S1",
  "to": "app.py::Helper",
  "kind": "calls",
  "resolution": "ambiguous",
  "candidates": ["helpers/a.py::Helper", "helpers/b.py::Helper"]
}
```

JSON et SKT préservent la même incertitude avec des encodages différents. Consultez [Prise en charge des langages et fidélité de l'API](../language-support/) pour les règles de résolution et les limites connues.