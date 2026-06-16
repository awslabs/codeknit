---
title: Référence du format de sortie
description: Référence complète du format de sortie .skt utilisé par codeknit.
---

Le format `.skt` (skeleton) est un format texte compact et lisible par l'homme utilisé par `codeknit` pour représenter la structure du code extraite. Il contient des symboles, des relations et des métadonnées sous une forme minimale adaptée à la consommation par les LLM et à l'analyse structurelle.

Un fichier `.skt` est divisé en sections. Chaque section commence par un en-tête entre crochets. Les sections peuvent apparaître dans n'importe quel ordre, bien que `[symbols]` apparaisse généralement en premier.

## [symbols]

La section `[symbols]` liste tous les symboles extraits, regroupés par fichier source. Chaque fichier est introduit par un en-tête `##` suivi du chemin du fichier.

### Format de ligne

Chaque symbole est représenté sur une seule ligne avec la structure suivante :

```
ShortID catégorie/type Ldébut-Lfin signature {propriétés}
```

### Champs

- **ShortID** : Un identifiant séquentiel attribué à chaque symbole (par exemple, `S1`, `S2`, `S3`). Utilisé comme référence dans les arêtes et autres sections.
- **catégorie/type** : Une paire séparée par une barre oblique indiquant la catégorie et le type spécifique du symbole.
- **Ldébut-Lfin** : La plage de lignes dans le fichier source où le symbole est défini (par exemple, `L10-L15`).
- **signature** : Le nom et les informations de type du symbole. Le format dépend du symbole :
  - `nom` — pour les types, valeurs, modules
  - `nom(paramètres)` — pour les callables sans type de retour
  - `nom(paramètres) -> typeRetour` — pour les callables avec type de retour
- **{propriétés}** : Métadonnées optionnelles entre accolades. Plusieurs propriétés sont séparées par des virgules.

### Paramètres

- Dans les langages non typés : `nomParamètre`
- Dans les langages typés : `nomParamètre: type`
- Les références de type qui correspondent à des symboles connus sont remplacées par leurs ShortIDs (par exemple, `config: S5` au lieu de `config: Config`).

### Propriétés

Propriétés courantes :

- `async` : `true` ou `false`
- `exported` : `true` ou `false`
- `static` : présent si le symbole est statique
- `visibility=public|private|protected`
- `receiver=*NomType` : pour les méthodes, indique le type du récepteur

### Catégories et types de symboles

| Catégorie  | Types                          | Exemples                               |
| ---------- | ------------------------------ | -------------------------------------- |
| `callable` | function, method, constructor  | `callable/function`, `callable/method` |
| `type`     | class, interface, struct, enum | `type/class`, `type/interface`         |
| `value`    | variable, constant, field      | `value/variable`, `value/constant`     |
| `module`   | package, namespace             | `module/package`                       |
| `meta`     | type parameters, metadata      | `meta/type_parameter`                  |

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

La section `[edges]` définit les relations entre les symboles en utilisant leurs ShortIDs.

### Format de ligne

```
IDSource --type--> IDCible1, IDCible2
```

Plusieurs IDs cibles sont séparés par des virgules. Chaque ligne représente une relation directionnelle.

### Types d'arêtes

| Type         | Signification                                             |
| ------------ | --------------------------------------------------------- |
| `calls`      | invocation de fonction/méthode                            |
| `contains`   | classe contient une méthode, module contient une fonction |
| `inherits`   | classe étend une autre classe                             |
| `implements` | classe implémente une interface                           |
| `overrides`  | méthode redéfinit une méthode parente                     |
| `references` | symbole référence un autre symbole                        |
| `imports`    | module importe un autre module                            |
| `decorates`  | décorateur appliqué à un symbole                          |

### Exemple

```skt
[edges]
S2 --contains--> S4
S4 --calls--> S5
S10 --inherits--> S2
S24 --implements--> S19
```

## [errors]

La section `[errors]` liste les fichiers qui n'ont pas pu être analysés complètement.

### Format

Chaque ligne commence par `-` suivi du chemin du fichier et du message d'erreur :

```
- chemin/vers/fichier.go: erreur de syntaxe à la ligne 42
```

### Exemple

```skt
[errors]
- src/broken.go: unexpected token at line 10
- tests/corner_case.py: unterminated string literal
```

## [dict]

La section `[dict]` apparaît uniquement lorsque le drapeau `--minify` est utilisé. Elle mappe des codes de dictionnaire courts à des jetons de chaîne répétés pour réduire la taille de la sortie.

### Format

Chaque ligne mappe un code de dictionnaire (`d0`, `d1`, etc.) à sa valeur développée :

```
- d0: async=false
- d1: callable/method
- d2: exported
```

Dans le reste du fichier, ces codes remplacent leurs valeurs complètes.

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
S3 --returns--> S2

[errors]
- utils/broken.go: syntax error at line 5
```
