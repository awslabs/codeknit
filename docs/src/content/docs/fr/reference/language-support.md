---
title: Prise en charge des langages et fidélité de l'API
description: Capacités d'extraction testées, sélection du parseur et limitations connues des langages.
---

<!-- Maintenu dans docs/reference/language-support.md ; copié textuellement par docs/gen. -->

Codeknit extrait la syntaxe sans exécuter de compilateur, évaluer des macros ou charger les dépendances d'un projet. Le tableau ci-dessous décrit les constructions testées, et non la conformité complète du langage ou la résolution garantie des noms entre fichiers.

Chaque langage dispose d'une assertion exacte de base pour les paramètres/retours de signature dans `internal/plugin/api_fidelity_test.go`. La couverture des déclarations est également vérifiée dans `internal/plugin/coverage_test.go` et les packages de test individuels pour chaque langage.

| Langage   | Extensions                                  | Référence de signature                                                                                       | Couverture des déclarations et conteneurs                                                                   |
| ---------- | ------------------------------------------- | ------------------------------------------------------------------------------------------------------------ | ---------------------------------------------------------------------------------------------------------- |
| C          | `.c`, `.h`                                  | Paramètres nommés et types de retour dans les définitions et prototypes                                      | Structures, unions, énumérations, typedefs, macros ; déclarations à l'intérieur des gardes d'en-tête      |
| C++        | `.cpp`, `.cc`, `.cxx`, `.hpp`, `.hxx`, `.h` | Paramètres nommés et types de retour dans les fonctions, prototypes et déclarations de méthodes              | Classes, structures, méthodes, espaces de noms, templates ; gardes d'en-tête et wrappers `extern "C"`     |
| C#         | `.cs`                                       | Paramètres typés nommés et types de retour explicites                                                        | Classes, structures, interfaces, méthodes, champs, délégués, espaces de noms                               |
| Go         | `.go`                                       | Paramètres groupés, non nommés, variadiques, typés par fonction et génériques ; résultats nommés/multiples ; récepteurs génériques | Champs de structure avec types et tags, champs embarqués, signatures de méthodes d'interface, paramètres de type, packages |
| Java       | `.java`                                     | Paramètres typés nommés et types de retour explicites                                                        | Classes, interfaces, constructeurs, méthodes, champs, classes imbriquées, packages                         |
| JavaScript | `.js`, `.jsx`, `.mjs`, `.cjs`               | Paramètres nommés ; pas de type de retour inféré                                                             | Fonctions, fonctions fléchées, classes, champs, méthodes, déclarations contenant du JSX                    |
| PHP        | `.php`                                      | Paramètres typés nommés et types de retour explicites                                                        | Classes, interfaces, traits, énumérations, méthodes, propriétés, espaces de noms                           |
| Python     | `.py`, `.pyi`                               | Annotations, valeurs par défaut, `*args`, `**kwargs`, séparateurs positionnels/mot-clés, annotations de retour | Fonctions, classes, classes imbriquées, méthodes décorées ; récepteurs liés conventionnels omis            |
| Ruby       | `.rb`                                       | Paramètres nommés ; pas de type de retour inféré                                                             | Classes, modules, méthodes d'instance/de singleton, constantes                                             |
| Rust       | `.rs`                                       | Motifs typés, `self`/`&self`/`&mut self`, liaisons mutables, types de retour explicites                      | Modules en ligne et imbriqués, champs de structure nommés, méthodes de traits et méthodes d'implémentation |
| Scala      | `.scala`, `.sc`                             | Paramètres typés nommés et types de retour explicites                                                        | Classes, traits, objets, méthodes, énumérations, packages                                                  |
| TypeScript | `.ts`, `.tsx`, `.mts`, `.cts`               | Paramètres typés nommés et annotations de retour explicites                                                 | Fonctions, fonctions fléchées, interfaces, classes, champs, espaces de noms ; TSX utilise sa propre grammaire |

## Sélection du langage d'en-tête

Toutes les commandes CLI de traitement de source acceptent `--header-language c|cpp`. Le choix s'applique uniquement aux fichiers `.h`. Les autres extensions conservent leurs parseurs habituels.

```bash
codeknit parse ./src --header-language cpp
codeknit fingerprint ./src --header-language c
codeknit graph analyze ./src --header-language cpp
```

La valeur par défaut est `c`, préservant l'extraction des macros, typedefs et unions C. Utilisez `cpp` pour les en-têtes C++ nommés `.h` ; les fichiers nommés `.hpp` et `.hxx` utilisent déjà C++. Le choix est explicite car les deux grammaires peuvent accepter une partie de la syntaxe de l'autre langage sans extraire sa structure complète.

Les gardes d'en-tête et les branches conditionnelles sont parcourues sans évaluer leurs conditions. Les résultats peuvent donc inclure des déclarations provenant de branches mutuellement exclusives. L'expansion des macros, la résolution des inclusions et les flags de build ne sont pas appliqués. L'interface utilisateur utilise la sélection C par défaut pour les fichiers `.h`.

## Préservation de la forme de l'API

Par exemple, en Go :

```go
func Pair(a, b int, rest ...string) (string, error)
```

produit :

```text
Pair(a: int, b: int, rest: ...string) -> (string, error)
```

Les méthodes d'interface Go et les champs nommés Go/Rust apparaissent comme des symboles dans les modes normal et d'empreinte. Les méthodes et champs conservent leur type propriétaire. Les modules Rust imbriqués conservent des noms qualifiés tels que `service.storage.read`, y compris lorsqu'un autre module déclare une fonction de même nom.

Rust permet aux champs et méthodes d'avoir le même nom. Les noms de champs scopés portent en interne un suffixe `#field` afin que leurs IDs et arêtes de containment restent distincts ; les noms d'affichage et les signatures conservent le nom de champ d'origine.

Les signatures Python conservent les annotations, les valeurs par défaut, les splats et les séparateurs de convention d'appel. Un premier paramètre conventionnel `self` ou `cls` est omis d'une méthode liée, mais conservé dans une fonction libre ou une méthode statique.

Les signatures multilignes conservent les sauts de ligne source en JSON. SKT échappe les sauts de ligne en `\n` et `\r` afin que chaque symbole reste sur une seule ligne physique de sortie.

## Résolution des relations

Les relations conservent la qualification du récepteur et la propriété lexicale. Les imports relatifs et les imports renommés JavaScript/TypeScript et Python fournissent un contexte cible. Les déclarations de packages Go et Java peuvent être résolues entre fichiers dans le même répertoire. Un import non apparié ne retombe jamais sur une déclaration non liée, et les candidats concurrents restent ambigus.

Les liaisons de récepteur locales prennent precedence sur les imports d'espaces de noms, y compris lorsqu'un paramètre masque un espace de noms importé. Les imports de packages Python considèrent les déclarations dans le module `__init__` du package, et non les modules descendants arbitraires. Les ré-exports qui ne peuvent pas être établis à partir de ce module restent non résolus.

L'extraction des appelables en Java, C++, C#, Scala et TypeScript conserve les plages d'octets des déclarations afin que les corps surchargés puissent avoir des appelants et des portées d'alias distincts, y compris les déclarations sur la même ligne. La sélection d'une cible surchargée reste conservative. Les relations de substitution explicites en Java, C++, C#, Scala et TypeScript recherchent la déclaration héritée la plus proche ; les bases manquantes, les cycles ou les candidats concurrents conservent l'incertitude au lieu de produire des auto-liens.

Il s'agit d'une analyse basée sur la syntaxe, et non d'un compilateur ou d'un graphe d'appels runtime. Les spécificateurs de packages JavaScript nus, les ré-exports de packages, les configurations de build, les types de récepteurs complexes et le dispatch dynamique peuvent rester non résolus. Un nom unique ailleurs dans le dépôt n'est pas une preuve suffisante d'une dépendance.

JSON préserve les arêtes incertaines avec `resolution: "unresolved"` ou `"ambiguous"` et inclut les IDs des candidats lorsque disponibles. Ces arêtes n'ont pas d'ID court cible.

SKT utilise la même notation d'arête pour toutes les relations, avec un marqueur de résolution optionnel :

```text
S8 --references--> S5
S10 --references[unresolved]--> string, bool
S11 --calls[ambiguous]--> Helper [candidates=S12, S15]
```

Les points d'extrémité connus et les candidats utilisent des IDs courts, y compris les points d'extrémité de relations non résolues. Les cibles incertaines sans symbole sont des noms relatifs au fichier source. Les noms littéraux qui ressemblent à des IDs courts, et autres identités complexes, utilisent des IDs complets entre guillemets pour éviter l'ambiguïté. Les arêtes avec la même source, le même type, la même résolution et le même ensemble de candidats partagent une liste de cibles. Le lecteur SKT préserve la résolution et les candidats, y compris les références à des symboles dans des chunks de sortie ultérieurs. La sortie minifiée encode le type de relation tout en gardant le marqueur de résolution lisible.

Les inclusions C/C++ ont des symboles explicites `meta/file` et `meta/include`. Les orthographes d'en-tête apparaissent une fois dans la liste des symboles ; les arêtes utilisent leurs IDs courts :

```text
[symbols]
## src/main.c
S1 meta/file L1-L4 main.c
S2 meta/include L1-L1 "config.h" {resolution=unresolved}
S3 meta/include L2-L2 "utils.h" {resolution=unresolved}
S4 meta/include L3-L3 <stdio.h> {resolution=unresolved}

[edges]
S1 --references[unresolved]--> S2, S3, S4
```

Un symbole d'inclusion décrit la directive écrite, et non une déclaration d'en-tête résolue. Les inclusions répétées au sein d'un fichier partagent un point d'extrémité. Les métadonnées de fichier/inclusion sont exclues des classements de dépendances et des résultats de code mort ; l'ajout d'IDs ne prétend pas qu'un en-tête a été trouvé via la configuration de recherche d'inclusion du compilateur.

Les métriques de dépendance et les liens du graphe HTML n'utilisent que les arêtes résolues. Les rapports affichent les comptes de relations exclues, et le graphe HTML affiche un compte d'incertitude. Le fan-in et le fan-out comptent les voisins distincts. L'analyse d'héritage considère toutes les bases et interfaces ; les dépendances cycliques n'ont pas de profondeur rapportée. La reachabilité commence à partir des appelables nommés `main`, `Main` ou `init` et suit les arêtes de dépendance. Atteindre un membre rend son conteneur pertinent sans rendre chaque frère atteignable. Sans points d'entrée reconnus, les résultats de reachability sont omis.

Les arguments de rappel et les appelables retournés sont des références, et non une preuve d'invocation. La résolution d'alias conserve l'identité du fichier, de la fonction et de l'objet ; les affectations conflictuelles restent incertaines. Les résultats du graphe sont des candidats à la revue : les appelants externes et le comportement dynamique peuvent modifier le résultat. Les règles de couche basées sur le répertoire et les poids de propagation des changements sont des heuristiques, et non des probabilités mesurées.

## Limites connues

- Les déclarateurs complexes de pointeurs, références, tableaux et pointeurs de fonction en C/C++ peuvent encore avoir des signatures incomplètes. Les déclarations générées par macro ne sont pas expansées.
- La gestion des constructeurs TypeScript/JavaScript et certaines formes de paramètres optionnels, rest ou destructurés sont incomplètes.
- Les contraintes génériques avancées et les formes de type en dehors des cas testés varient selon le langage. Les types de retour ne sont pas inférés.
- Les macros Rust ne sont pas expansées. Les déclarations de modules hors ligne sont enregistrées ; leurs fichiers doivent être inclus dans l'analyse. Les champs de tuples, les détails des variantes d'énumération et les éléments associés ont une couverture partielle.
- L'extraction syntaxique ne résout pas le dispatch dynamique, les surcharges, les dépendances importées ou la sémantique de configuration de build avec la précision d'un compilateur.
- Les erreurs de syntaxe partielles peuvent encore omettre des déclarations. Inspectez les avertissements rapportés et vérifiez le code source avant d'appliquer une refactorisation.

Une extraction plus complète modifie les comptes de symboles et les IDs courts. Régénérez la sortie squelette existante lors de la mise à niveau au lieu de la mélanger avec une sortie plus ancienne.