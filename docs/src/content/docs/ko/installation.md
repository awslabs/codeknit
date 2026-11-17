---
title: 설치
description: 시스템에 codeknit을 설치하는 방법입니다.
---

이 가이드에서는 codeknit을 소스로부터 설치하고, `PATH`에 추가하며, 셸 완성 기능을 활성화하고, 실행되는지 확인하는 방법을 설명합니다.

:::tip[AI 코딩 어시스턴트를 사용하나요?]
Claude Code, Codex, Kiro 사용자는 codeknit 플러그인을 대신 설치할 수 있습니다. 플러그인은 처음 실행할 때 바이너리를 다운로드합니다. [AI 어시스턴트와 함께 codeknit 사용하기](../guides/ai-assistants/)를 참조하세요.
:::

## 소스로부터 설치

**필수 조건**

- Go 1.26 이상
- C 컴파일러. codeknit은 tree-sitter 파서를 빌드하기 위해 CGo를 사용합니다.

**단계**

```bash
git clone https://github.com/awslabs/codeknit.git
cd codeknit
make build
```

바이너리는 `./bin/codeknit`에 생성됩니다.

### 사전 빌드된 바이너리 사용하기

Linux, macOS, Windows(amd64 및 arm64)용 사전 빌드된 바이너리는 [릴리스 페이지](https://github.com/awslabs/codeknit/releases)에서 확인할 수 있습니다. 플랫폼에 맞는 아카이브를 다운로드하고 압축을 해제한 다음, 바이너리를 `PATH`에 있는 디렉터리로 이동하세요. 각 릴리스에는 다운로드를 확인할 수 있는 `checksums.txt` 파일이 포함되어 있습니다.

## codeknit을 PATH에 추가하기

이 섹션은 바이너리를 이미 `PATH`에 있는 디렉터리(예: `/usr/local/bin`)로 이동한 경우 건너뛰어도 됩니다.

`/path/to/codeknit`을 바이너리가 포함된 디렉터리로 바꿉니다. 소스로부터 빌드한 경우, 이는 클론된 저장소 내의 `bin` 디렉터리입니다.

셸에 맞는 줄을 해당 셸의 설정 파일에 추가하세요:

```bash
# bash (~/.bashrc)
export PATH="$PATH:/path/to/codeknit"

# zsh (~/.zshrc)
export PATH="$PATH:/path/to/codeknit"

# fish (~/.config/fish/config.fish)
fish_add_path /path/to/codeknit
```

셸을 다시 로드하거나, 수정한 파일을 소싱하세요:

```bash
source ~/.bashrc   # bash
source ~/.zshrc    # zsh
```

fish의 경우, 새 터미널을 여세요.

## 셸 완성 기능 활성화하기

codeknit은 bash, zsh, fish, PowerShell용 완성 스크립트를 생성합니다. 각 셸에 맞는 명령을 한 번 실행하세요:

```bash
# bash
codeknit completion bash >> ~/.bashrc

# zsh
codeknit completion zsh >> ~/.zshrc

# fish
codeknit completion fish > ~/.config/fish/completions/codeknit.fish

# PowerShell
codeknit completion powershell >> $PROFILE
```

bash, zsh, PowerShell 명령은 프로필에 내용을 추가합니다. 명령을 다시 실행하면 스크립트가 두 번 추가되므로, 각 명령을 한 번만 실행하세요.

새 셸을 열거나 프로필을 소싱하여 완성 기능을 활성화하세요.

## 설치 확인하기

```bash
codeknit --version
```

명령어가 버전을 출력하면 codeknit이 설치된 것입니다. 셸에서 `command not found`가 표시되면, `PATH`에 지정한 디렉터리에 `codeknit` 바이너리가 있는지 확인한 다음 셸을 다시 로드하세요.

## 개발 환경 설정하기

codeknit에 기여할 계획이 있는 경우에만 이 단계를 수행하세요. 클론된 저장소의 루트에서 다음 명령을 실행하세요:

```bash
make deps    # 의존성 설치
make setup   # git 훅 구성
make test    # 테스트 스위트 실행
```

`make lint`는 린터와 포맷터를 실행합니다.

## 다음 단계

- [시작하기](../getting-started/)에서 첫 번째 명령어를 실행해 보세요.
- [CLI 플래그 참조](../reference/cli-flags/)에서 모든 플래그를 살펴보세요.