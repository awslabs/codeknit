---
title: Prise en charge des langages et fidélité de l'API
description: Capacités d'extraction testées, sélection des analyseurs et limitations connues des langages.
---

<!-- Maintenu dans docs/reference/language-support.md ; copié mot pour mot par docs/gen. -->

Codeknit extrait la syntaxe sans exécuter de compilateur, évaluer des macros ou charger les dépendances d'un projet. Le tableau ci-dessous décrit les constructions testées, et non la conformité complète du langage ou la résolution garantie des noms entre fichiers.

Chaque langage dispose d'une assertion exacte des paramètres de base et des signatures de retour dans `internal/plugin/api_fidelity_test.go`. La couverture des déclarations est également vérifiée dans `internal/plugin/coverage_test.go` et les packages de test individuels pour chaque langage.

| Langage    | Extensions                                  | Référence de signature                                                                                                            | Couverture des déclarations et des conteneurs                                                                           |
| ---------- | ------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------- |
| C          | `.c`, `.h`                                  | Paramètres nommés et types de retour dans les définitions et prototypes                                                           | Structs, unions, enums, typedefs, macros ; déclarations à l'intérieur des gardes d'en-tête                              |
| C++        | `.cpp`, `.cc`, `.cxx`, `.hpp`, `.hxx`, `.h` | Paramètres nommés et types de retour dans les fonctions, prototypes et déclarations de méthodes                                   | Classes, structs, méthodes, namespaces, templates ; gardes d'en-tête et wrappers `extern "C"`                           |
| C#         | `.cs`                                       | Paramètres typés nommés et types de retour explicites                                                                             | Classes, structs, interfaces, méthodes, champs, délégués, namespaces                                                    |
| Go         | `.go`                                       | Paramètres groupés, non nommés, variadiques, typés en fonction et génériques ; résultats nommés/multiples ; récepteurs génériques | Champs de struct avec types et tags, champs embarqués, signatures de méthodes d'interface, paramètres de type, packages |
| Java       | `.java`                                     | Paramètres typés nommés et types de retour explicites                                                                             | Classes, interfaces, constructeurs, méthodes, champs, classes imbriquées, packages                                      |
| JavaScript | `.js`, `.jsx`, `.mjs`, `.cjs`               | Paramètres nommés ; pas de type de retour inféré                                                                                  | Fonctions, fonctions fléchées, classes, champs, méthodes, déclarations contenant du JSX                                 |
| PHP        | `.php`                                      | Paramètres typés nommés et types de retour explicites                                                                             | Classes, interfaces, traits, enums, méthodes, propriétés, namespaces                                                    |
| Python     | `.py`, `.pyi`                               | Annotations, valeurs par défaut, `*args`, `**kwargs`, séparateurs positionnels/mot-clés, annotations de retour                    | Fonctions, classes, classes imbriquées, méthodes décorées ; récepteurs liés conventionnels omis                         |
| Ruby       | `.rb`                                       | Paramètres nommés ; pas de type de retour inféré                                                                                  | Classes, modules, méthodes d'instance/de singleton, constantes                                                          |
| Rust       | `.rs`                                       | Motifs typés, `self`/`&self`/`&mut self`, liaisons mutables, types de retour explicites                                           | Modules inline et imbriqués, champs de struct nommés, méthodes de traits et méthodes d'implémentation                   |
| Scala      | `.scala`, `.sc`                             | Paramètres typés nommés et types de retour explicites                                                                             | Classes, traits, objets, méthodes, enums, packages                                                                      |
| TypeScript | `.ts`, `.tsx`, `.mts`, `.cts`               | Paramètres typés nommés et annotations de retour explicites                                                                       | Fonctions, fonctions fléchées, interfaces, classes, champs, namespaces ; TSX utilise sa propre grammaire                |

## Sélection du langage pour les en-têtes

Toutes les commandes CLI de traitement de source acceptent `--header-language c|cpp`. Le choix ne s'applique qu'aux fichiers `.h`. Les autres extensions conservent leurs analyseurs habituels.

```bash
codeknit parse ./src --header-language cpp
codeknit fingerprint ./src --header-language c
codeknit graph analyze ./src --header-language cpp
```

La valeur par défaut est `c`, préservant l'extraction des macros, typedefs et unions en C. Utilisez `cpp` pour les en-têtes C++ nommés `.h` ; les fichiers nommés `.hpp` et `.hxx` utilisent déjà C++. Le choix est explicite car les deux grammaires peuvent accepter une partie de la syntaxe de l'autre langage sans extraire sa structure complète.

Les gardes d'en-tête et les branches conditionnelles sont parcourues sans évaluer leurs conditions. Les résultats peuvent donc inclure des déclarations provenant de branches mutuellement exclusives. L'expansion des macros, la résolution des includes et les flags de build ne sont pas appliqués. L'interface utilisateur utilise la sélection C par défaut pour les fichiers `.h`.

## Préservation de la forme de l'API

Par exemple, en Go :

```go
func Pair(a, b int, rest ...string) (string, error)
```

produit :

```text
Pair(a: int, b: int, rest: ...string) -> (string, error)
```

Les méthodes d'interface Go et les champs nommés en Go/Rust apparaissent comme des [symbols] dans les modes normaux et de _fingerprinting_. Les méthodes et champs conservent leur type propriétaire. Les modules Rust imbriqués conservent des noms qualifiés tels que `service.storage.read`, y compris lorsqu'un autre module déclare une fonction du même nom.

Rust permet aux champs et méthodes d'avoir le même nom. Les noms de champs scopés portent un suffixe `#field` en interne afin que leurs IDs et leurs arêtes de containment restent distincts ; les noms d'affichage et les signatures conservent le nom de champ d'origine.

Les signatures Python conservent les annotations, les valeurs par défaut, les splats et les séparateurs de convention d'appel. Un premier paramètre conventionnel `self` ou `cls` est omis d'une méthode liée, mais conservé dans une fonction libre ou une méthode statique.

Les signatures multilignes conservent les sauts de ligne source en JSON. SKT échappe les sauts de ligne en `\n` et `\r` afin que chaque symbole reste sur une seule ligne physique de sortie.

## Limites connues

- Les déclarateurs complexes de pointeurs, références, tableaux et pointeurs de fonction en C/C++ peuvent encore avoir des signatures incomplètes. Les déclarations générées par macro ne sont pas expansées.
- La gestion des constructeurs TypeScript/JavaScript et certaines formes de paramètres optionnels, rest ou destructurés sont incomplètes.
- Les contraintes génériques avancées et les formes de type en dehors des cas testés varient selon le langage. Les types de retour ne sont pas inférés.
- Les macros Rust ne sont pas expansées. Les déclarations de modules hors ligne sont enregistrées ; leurs fichiers doivent être inclus dans l'analyse. Les champs de tuple, les détails des variantes d'enum et les éléments associés ont une couverture partielle.
- L'extraction syntaxique ne résout pas la dispatch dynamique, les surcharges, les dépendances importées ou la sémantique de configuration de build avec la précision d'un compilateur.
- Les erreurs de syntaxe partielles peuvent encore omettre des déclarations. Inspectez les avertissements rapportés et vérifiez le code source avant d'appliquer une refactorisation.

Des modifications plus complètes de l'extraction changent le nombre de symboles et les ShortID. Régénérez la sortie du _skeleton_ existante lors d'une mise à niveau au lieu de la mélanger avec une sortie plus ancienne.
