---
title: Fiabilité de l'analyse
description: Sélection des sources, analyses incomplètes, publication des résultats et garanties de détection des doublons.
---

<!-- Maintenu dans docs/reference/analysis-reliability.md ; copié textuellement par docs/gen. -->

Ces comportements décrivent les modifications non publiées de Go/CLI suivant la version 0.5.0.

## Sélection des sources et diagnostics

Les cinq commandes de traitement des sources (`parse`, `fingerprint`, `graph show`, `graph analyze` et `graph hotspots`) acceptent :

| Flag                    | Défaut                | Signification                                                            |
| ----------------------- | --------------------- | ------------------------------------------------------------------------ |
| `--include GLOB`        | Toutes les sources prises en charge | Inclure les chemins de sources correspondants ; répéter pour plusieurs motifs. |
| `--exclude GLOB`        | Aucun                 | Exclure les chemins ou répertoires correspondants ; répéter pour plusieurs motifs. |
| `--max-file-size BYTES` | `10485760` (10 Mio)   | Ignorer les fichiers sources plus volumineux ; `0` désactive cette limite. |
| `--strict`              | `false`               | Échouer avant d'émettre des résultats si des diagnostics de scan ou d'analyse surviennent. |

Les motifs utilisent des globs de style gitignore séparés par des barres obliques, relatifs au répertoire d'entrée sélectionné. Pour une entrée mono-fichier, les motifs correspondent à son nom de base. Mettez les motifs entre guillemets pour éviter leur expansion par le shell. Les inclusions forment une union ; les exclusions prennent le pas. Les inclusions ne remplacent pas `.gitignore`, le filtrage des tests ou les exclusions intégrées. Les motifs de négation (`!motif`) ne sont pas acceptés dans ces flags ; utilisez les contrôles d'inclusion et d'exclusion séparés.

```bash
codeknit parse ./src --include '**/*.go' --exclude '**/generated/' --strict
codeknit fingerprint ./src --max-file-size 2097152 --include '**/*.ts'
codeknit parse ./src/app.go --output-mode inline --format json
```

Les analyses de répertoires respectent les fichiers `.gitignore` ancestraux et imbriqués par rapport à leur emplacement de projet, indépendamment du répertoire de travail du processus. Le répertoire `.git` ou le marqueur de worktree le plus proche établit la limite du dépôt. En dehors d'un dépôt, les fichiers ignore ancestraux s'appliquent toujours. Un sous-répertoire sélectionné sous un répertoire ignoré reste ignoré. La sélection explicite d'un fichier source régulier contourne ignore et le filtrage des tests, mais applique toujours les limites d'inclusion/exclusion et de taille.

Les liens symboliques de sources, les liens symboliques de répertoires et les sources non régulières sont ignorés et signalés. Les lectures sont confinées à la racine du scan, et la limite de taille est appliquée à nouveau lors de la lecture afin qu'un fichier qui grossit après sa découverte ne puisse pas la contourner. Les exclusions de répertoires intégrées incluent `.git`, `.hg`, `.svn`, `node_modules`, les caches Python et les environnements virtuels Python courants. Les fichiers de test restent exclus sauf si `--collect-test` est défini.

Les répertoires/fichiers illisibles, les échecs de lecture des fichiers ignore, les liens symboliques ignorés, les sources trop volumineuses et les avertissements de syntaxe rendent une analyse incomplète. Par défaut, les résultats partiels utilisables sont conservés et les diagnostics apparaissent sur stderr. Le JSON d'analyse inclut des enregistrements de chemin/raison dans `errors` ; le SKT inclut `[errors]` en ligne ou dans `warnings.skt`. Les commandes graph/fingerprint signalent également les diagnostics sur stderr. Les fichiers délibérément ignorés ou filtrés ne sont pas des avertissements. `--strict` rejette les résultats incomplets avant la publication des résultats. Une entrée sans sources prises en charge sélectionnées, ou une exécution dans laquelle chaque lecture de source échoue, se termine toujours sans succès.

L'interface utilisateur textuelle utilise la même politique de liens symboliques et la limite par défaut de 10 Mio. Ses formulaires n'exposent pas encore les nouveaux flags de sélection et strict.

## Noms de sortie et publication

La sortie en arbre ajoute `.skt` au nom complet de la source : `src/foo.c` devient `src/foo.c.skt`, tandis que `src/foo.h` devient `src/foo.h.skt`. La sortie divisée utilise `src/foo.c_part1.skt`, `src/foo.c_part2.skt`, et ainsi de suite. Les fichiers de dictionnaire et de diagnostic conservent les noms réservés `dict.skt` et `warnings.skt`. Les chemins de sortie en conflit, y compris les collisions de casse uniquement, échouent visiblement.

Le SKT en arbre nécessite `--max-lines` d'au moins 3. Chaque bloc de symboles répète ses en-têtes `[symbols]` et `## source-path`. Les blocs contenant uniquement des arêtes portent `[edges]`. Chaque bloc de source respecte la limite ; les dictionnaires et les fichiers de diagnostic sont des sorties auxiliaires séparées. Le SKT plat conserve les blocs de source intacts, donc sa limite de lignes est une cible de division plutôt qu'une limite stricte pour un grand bloc de source. Le JSON est un document unique et n'est pas divisé par nombre de lignes.

La génération de répertoire s'achève dans un répertoire de staging frère. Sans `--clean`, la sortie générée existante pour le format sélectionné provoque une erreur. Avec `--clean`, les fichiers `.skt` obsolètes (mode SKT) ou `codeknit.json` (mode JSON) sont remplacés uniquement après une génération réussie. Les entrées non liées sont préservées ; une collision avec l'une d'elles provoque une erreur. N'utilisez pas un répertoire source ou son ancêtre comme répertoire de sortie.

La publication maintient un verrou spécifique au répertoire, sauvegarde la sortie précédente et la restaure en cas d'échec du remplacement. Si un processus se termine entre les renommages de répertoires, la prochaine invocation restaure la sauvegarde avant de générer quoi que ce soit. Le fichier `.NAME.codeknit.lock` frère reste intentionnellement pour coordonner les futurs éditeurs ; ne supprimez pas un verrou actif ou une sauvegarde de récupération. Il existe un bref intervalle entre les deux renommages de répertoires, donc les lecteurs nécessitant un accès ininterrompu doivent se coordonner avec la génération. Cela fournit une récupération en cas de défaillance du processus, mais pas une garantie contre une défaillance de stockage ou une coupure de courant.

Les rapports individuels sont écrits et synchronisés vers un frère temporaire avant de remplacer la destination. Les fichiers existants ne sont pas tronqués pendant la génération. La sortie standard en ligne ne peut pas être annulée si son consommateur échoue.

## Garanties de détection des doublons

La correspondance exacte signifie des flux de jetons normalisés complets identiques dans la même catégorie de symbole parmi les déclarations éligibles pour le *fingerprinting*. Cela ne signifie pas un texte source identique ou une équivalence sémantique prouvée. Les groupes de candidats exacts utilisent les flux de jetons eux-mêmes, plutôt que l'égalité de *fuzzy hashing*.

En dessous de 500 symboles éligibles, la détection structurelle compare toutes les paires de même catégorie. À partir de 500 ou plus, toutes les paires de jetons exacts sont conservées indépendamment de la récupération des candidats, et les autres paires utilisent les 50 voisins structurels approximatifs les plus proches par symbole. La récupération des quasi-doublons n'est donc pas exhaustive sur les grandes entrées. La récupération analyse toujours tous les vecteurs ; elle ne rend pas l'algorithme global sous-quadratique.

Le filtrage de la plage de similarité s'applique toujours. La plage par défaut est de 65–95 %, ce qui exclut les correspondances à 100 %. Pour inclure les correspondances exactes, utilisez :

```bash
codeknit fingerprint ./src --min-similarity 100 --max-similarity 100
```

L'option `--rerank` ajoute des candidats sémantiques, applique l'accord cosinus et filtre sur le score combiné ; elle peut supprimer un candidat structurellement exact. Le texte d'intégration provient du même instantané de source borné utilisé pour l'extraction. Les entrées stables et les scores fixes produisent un ordonnancement déterministe, y compris les égalités ; les réponses du modèle elles-mêmes sont des entrées externes.

`internal/emitter/calibration_test.go` évalue un petit fixture source étiqueté à un seuil de 70 %, rapportant les vrais/faux positifs et négatifs, la précision et le rappel séparément des scores de similarité. Le fixture contient une paire exacte, deux paires de quasi-clones et trois paires sans rapport, exercées ci-dessous et au-delà de la limite de 500 symboles. Il s'agit d'un contrôle de régression, et non d'une estimation représentative de la précision dans le monde réel. Des budgets de calibration et de performance multilingues plus importants restent un travail en cours.