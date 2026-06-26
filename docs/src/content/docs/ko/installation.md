---
title: 설치
description: 시스템에 codeknit을 설치하는 방법.
---

codeknit은 소스에서 설치할 수 있습니다. 다음 단계는 시스템에 codeknit을 설정하는 방법을 안내합니다.

## 소스에서 설치

기본 설치 방법은 소스에서 빌드하는 것입니다. 다음이 필요합니다:

- Go 1.26+
- C 컴파일러 (tree-sitter를 위한 CGo 필요)

리포지토리를 클론하고 바이너리를 빌드합니다:

```bash
git clone https://github.com/awslabs/codeknit.git
cd codeknit
make build
```

컴파일된 바이너리는 `./bin/codeknit`에서 사용할 수 있습니다.

## PATH에 추가

`codeknit`을 어느 디렉토리에서나 실행하려면 바이너리 위치를 시스템의 PATH에 추가하세요.

**bash**용 (`~/.bashrc`):

```bash
export PATH="$PATH:/path/to/codeknit"
```

**zsh**용 (`~/.zshrc`):

```bash
export PATH="$PATH:/path/to/codeknit"
```

**fish**용 (`~/.config/fish/config.fish`):

```fish
fish_add_path /path/to/codeknit
```

셸 설정을 업데이트한 후 `source ~/.bashrc` (또는 `~/.zshrc`)를 실행하거나 터미널을 다시 시작하여 설정을 다시 로드하세요.

## 셸 자동 완성

codeknit은 인기 있는 셸에 대한 자동 완성을 지원합니다. 다음 명령어를 사용하여 자동 완성을 설치하세요:

**bash**용:

```bash
codeknit completion bash >> ~/.bashrc
```

**zsh**용:

```bash
codeknit completion zsh >> ~/.zshrc
```

**fish**용:

```bash
codeknit completion fish > ~/.config/fish/completions/codeknit.fish
```

**PowerShell**용:

```powershell
codeknit completion powershell >> $PROFILE
```

## 설치 확인

설치 후 codeknit이 올바르게 설정되었는지 확인하세요:

```bash
codeknit --version
```

## 개발 환경 설정

codeknit에 기여하는 경우 다음 추가 명령어를 실행하세요:

개발 의존성 설치:

```bash
make deps
```

git 훅 설정:

```bash
make setup
```

테스트 스위트 실행:

```bash
make test
```