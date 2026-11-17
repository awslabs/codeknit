---
title: Fingerprint 명령어
description: 퍼지 해싱을 사용하여 파일 및 언어 간 중복 및 근사 중복 코드를 탐지합니다.
---

`codeknit fingerprint`는 모든 함수, 메서드, 변수 및 타입의 정규화된 본문 토큰에 대해 Context-Triggered Piecewise Hash(CTPH)를 계산합니다. 변수 이름, 문자열 리터럴 및 타입 주석은 해싱 전에 정규화됩니다. 따라서 이름 변경된 복사-붙여넣기 및 다른 언어로의 이식도 여전히 일치할 수 있습니다.

이 가이드에서는 중복을 찾는 방법, 유사도 범위를 조정하는 방법, 거짓 양성을 줄이는 방법 및 리팩터링할 대상을 결정하는 방법을 설명합니다.

## 시작하기 전에

- `codeknit`이 설치되어 있습니다. [설치](../../installation/)를 참조하세요.
- `./src`와 같은 소스 디렉터리가 있습니다.

## 프로젝트에서 중복 찾기

확인하려는 디렉터리에서 명령을 실행합니다:

```bash
codeknit fingerprint ./src
```

이 명령은 `./skeleton/fingerprints.skt` 파일을 작성하고 65%에서 95% 사이의 유사도를 보고합니다.

파일을 엽니다. `[duplicates]` 섹션은 항상 존재하며, 가장 강력한 일치가 먼저 표시됩니다:

```skt
[duplicates]
# similarity range: 65%-95%
similarity:93%  pkg/user.go::GetUser <-> pkg/admin.go::GetAdmin
similarity:88%  src/auth.ts::validateToken <-> src/session.ts::validateSession
similarity:81%  lib/math.py::compute_tax <-> lib/math.py::compute_vat
```

각 줄은 `similarity:<N>%  <file>::<symbol> <-> <file>::<symbol>` 형식입니다. 범위 내에 아무것도 없으면 섹션에는 `# no duplicates found`만 포함됩니다.

전체 출력 문법은 [출력 형식 참조](../../reference/output-format/)를 참조하세요.

## 정확한 중복만 찾기

기본 범위는 95%로 상한이 설정되어 있으므로 100% 일치는 제외됩니다. 정확한 일치를 나열하려면 두 경계를 모두 100으로 설정합니다:

```bash
codeknit fingerprint ./src --min-similarity 100 --max-similarity 100
```

"정확한"은 동일한 심볼 범주 내에서 완전한 정규화된 토큰 스트림이 동일하다는 의미입니다. 소스 텍스트가 동일하다는 의미나 두 심볼이 동등하다는 것이 증명된 것은 아닙니다.

## 느슨한 일치를 찾기 위해 범위 확장

최소값을 낮추어 세부 사항은 다르지만 알고리즘 형태를 공유하는 코드를 포함합니다:

```bash
codeknit fingerprint ./src --min-similarity 50 --max-similarity 80
```

50% 미만의 결과는 일반적으로 노이즈입니다.

## 의미론적 재순위화를 통해 거짓 양성 줄이기

구조적 결과가 노이즈가 많으면 로컬 Ollama 서버의 임베딩을 사용하여 후보를 재순위화합니다.

1. Ollama를 시작하고 기본 임베딩 모델을 가져옵니다:

   ```bash
   ollama serve
   ollama pull qwen3-embedding:0.6b
   ```

2. `--rerank` 플래그와 함께 명령을 실행합니다:

   ```bash
   codeknit fingerprint ./src --rerank
   ```

다른 임베딩 모델을 사용하려면 `--model`을 전달합니다:

```bash
codeknit fingerprint ./src --rerank --model qwen3-embedding:4b
```

재순위화는 의미론적 후보를 추가하고 결합된 점수를 기준으로 필터링합니다. 거짓 양성을 줄이지만 모든 거짓 양성을 제거하지는 않습니다. 또한 구조적으로 정확한 후보를 제거할 수도 있으며, 유사도 범위는 여전히 적용됩니다. 임베딩은 추출에 사용된 동일한 파싱된 소스 스냅샷에서 계산됩니다.

## 원시 지문 검사

모든 심볼의 지문을 보려면 `--show-all`을 추가합니다:

```bash
codeknit fingerprint ./src --show-all
```

이렇게 하면 소스 파일별로 그룹화된 `[fingerprints]` 섹션이 추가되며, 심볼당 한 줄씩 표시됩니다:

```skt
[fingerprints]
## src/auth.ts
validateToken  FP:3:a1b2c3...:d4e5f6...  tokens:8e0f1a2b...
```

줄 형식은 `<symbol-name>  FP:<blocksize>:<hash1>:<hash2>  tokens:<hex>`입니다. `tokens`는 정규화된 토큰 스트림의 16진수 인코딩입니다.

## 보고서를 다른 위치에 작성

```bash
codeknit fingerprint ./src -o duplicates.skt
```

## 스캔할 파일 제어

테스트 파일은 기본적으로 건너뜁니다. `--collect-test`로 포함합니다:

```bash
codeknit fingerprint ./src --collect-test
```

스캔을 좁히려면 소스 선택 플래그를 사용합니다. 패턴이 셸에 의해 확장되지 않도록 따옴표로 묶습니다:

```bash
codeknit fingerprint ./src --include '**/*.ts' --exclude '**/generated/'
```

기타 선택 제어:

- `--max-file-size BYTES`는 더 큰 파일을 건너뜁니다. 기본값은 `10485760`(10 MiB)입니다. `0`은 제한을 해제합니다.
- `--strict`는 스캔 또는 파싱이 불완전한 경우 출력 작성 전에 실패합니다.
- `--header-language c|cpp`는 모호한 `.h` 파일에 대한 파서를 선택합니다. [언어 지원](../../reference/language-support/#header-language-selection)을 참조하세요.

건너뛴 심볼릭 링크, 너무 큰 파일 및 구문 경고는 stderr에 표시됩니다. 포함은 `.gitignore` 또는 테스트 필터링을 재정의하지 않으며, 제외가 포함보다 우선합니다. 정확한 보장은 [분석 신뢰성 참조](../../reference/analysis-reliability/)를 참조하세요.

## 유사도 범위 선택

| 범위      | 일반적으로 발견되는 내용                                      |
| --------- | ------------------------------------------------------------- |
| `96-100%` | 정확하거나 거의 정확한 구조적 중복; 명확한 복사-붙여넣기     |
| `85-95%`  | 근사 중복; 일반적으로 사소한 편집이 있는 복사-붙여넣기        |
| `65-84%`  | 기본 범위; 강한 구조적 유사성, 종종 리팩터링 후보             |
| `50-64%`  | 세부 사항은 다르지만 동일한 알고리즘 형태                    |
| 50% 미만  | 일반적으로 노이즈                                            |

상위 두 행을 보려면 `--max-similarity`를 `100`으로 올립니다.

## 리팩터링할 대상 결정

1. 기본 범위에서 시작하여 상위 쌍을 읽습니다.
2. 각 쌍의 두 심볼을 열고 비교합니다.
3. 일치를 신호로 취급하고 증명으로 취급하지 않습니다. 지문은 식별자 이름이나 의도가 아닌 구조적 형태를 측정합니다. 동일한 논리가 반복적으로 작성된 것과 재귀적으로 작성된 것은 점수가 낮으며, 동일한 형태를 가진 관련 없는 코드는 점수가 높을 수 있습니다.
4. 85~95% 범위의 쌍은 일반적으로 통합하기에 가장 안전한 후보입니다. 65~84% 범위의 쌍은 종종 더 자세히 살펴볼 필요가 있습니다.
5. 목록이 노이즈가 많으면 범위를 좁히거나 `--rerank`를 추가합니다.

언어 간 일치는 동등한 구조에 대해 작동합니다. 언어별 기능은 허위 저유사도 일치를 생성할 수 있습니다.

유사도 백분율은 휴리스틱 순위이며 편집 거리와는 다릅니다. 프로젝트의 레이블이 지정된 교정 픽스처는 유사도 점수와 별도로 정밀도와 재현율을 보고합니다. 이는 회귀 검사이며 사용자의 코드에서 정확도를 추정하는 것이 아닙니다.

## 대규모 프로젝트

- 500개 미만의 적격 심볼이 있는 경우, codeknit은 동일한 범주의 모든 쌍을 점수화합니다.
- 500개 이상인 경우, 완전한 정규화된 토큰 그룹은 항상 유지됩니다. 기타 쌍은 심볼당 근사 상위 50개 이웃을 사용하므로 근사 중복 결과가 불완전할 수 있습니다.
- 정확한 중복을 찾으려면 위의 100/100 명령을 사용합니다. 이는 근사에 영향을 받지 않습니다.
- 입력과 점수가 안정적이면 순서가 결정적입니다.

## 문제 해결

| 증상                     | 원인과 해결 방법                                                                 |
| ------------------------ | -------------------------------------------------------------------------------- |
| `# no duplicates found`  | 범위 내에 쌍이 없습니다. `--min-similarity`를 낮추거나 `--max-similarity`를 올립니다. |
| 알려진 짧은 함수가 누락됨 | 정규화된 토큰이 4개 미만인 심볼은 건너뜁니다.                                   |
| 정확한 복사본이 누락됨    | 기본 최대값은 95%입니다. `--min-similarity 100 --max-similarity 100`을 사용합니다. `--rerank`를 사용하면 의미론적 필터가 이를 제거할 수도 있습니다. |
| `--rerank` 실패          | `ollama serve`가 실행 중이고 모델이 가져와졌는지 확인합니다.                    |
| stderr에 경고와 함께 부분 결과 | 보고된 파일을 수정하거나 불완전한 출력을 게시하지 않도록 `--strict`를 추가합니다. |

## 이 가이드에서 사용된 플래그

| 플래그                  | 기본값                       | 설명                                                                         |
| ----------------------- | ---------------------------- | ---------------------------------------------------------------------------- |
| `-o`, `--output`        | `./skeleton/fingerprints.skt` | 출력 `.skt` 파일 경로.                                                      |
| `--min-similarity`      | `65`                         | 보고할 최소 유사도 백분율(0~100).                                           |
| `--max-similarity`      | `95`                         | 보고할 최대 유사도 백분율(0~100).                                           |
| `--show-all`            | `false`                      | 원시 토큰 데이터가 포함된 `[fingerprints]` 섹션도 출력합니다.                |
| `--rerank`              | `false`                      | Ollama를 통해 의미론적 임베딩으로 후보를 재순위화합니다. `ollama serve` 및 `ollama pull qwen3-embedding:0.6b`가 필요합니다. |
| `--model`               | `qwen3-embedding:0.6b`       | `--rerank`와 함께 사용할 Ollama 임베딩 모델.                                |
| `--collect-test`        | `false`                      | 분석에 테스트 파일을 포함합니다.                                            |
| `--workers`             | `NumCPU`                     | 최대 동시 파싱 고루틴 수.                                                   |
| `--verbose`             | `false`                      | 진행 정보를 출력합니다.                                                     |

`--include`, `--exclude`, `--max-file-size`, `--strict`, `--header-language`를 포함한 전체 목록은 [CLI 플래그 참조](../../reference/cli-flags/)를 참조하세요.

## 관련 항목

- [Parse 명령어](../parse-command/)
- [Graph 명령어](../graph-commands/)