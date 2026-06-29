#!/usr/bin/env bash
set -euo pipefail

usage() {
	cat <<'EOF'
Install codeknit assistant skills.

Usage:
  scripts/install-skills.sh [--assistant codex|kiro|claude|all] [--ref REF] [--force] [--dry-run]

Examples:
  curl -fsSL https://raw.githubusercontent.com/awslabs/codeknit/main/scripts/install-skills.sh | bash
  curl -fsSL https://raw.githubusercontent.com/awslabs/codeknit/main/scripts/install-skills.sh | bash -s -- --assistant codex

Options:
  --assistant  Target assistant. Defaults to all.
  --ref        Git ref to download skills from when not running in a checkout. Defaults to main.
  --force      Replace existing installed codeknit skills.
  --dry-run    Print the actions without copying files.
  -h, --help   Show this help.

Environment:
  CODEX_HOME   Codex home directory. Defaults to ~/.codex.
  CODEKNIT_REF Git ref to download skills from. Defaults to main.
EOF
}

assistant="all"
ref="${CODEKNIT_REF:-main}"
force=0
dry_run=0

while [[ $# -gt 0 ]]; do
	case "$1" in
		--assistant)
			if [[ $# -lt 2 ]]; then
				echo "error: --assistant requires a value" >&2
				exit 2
			fi
			assistant="$2"
			shift 2
			;;
		--ref)
			if [[ $# -lt 2 ]]; then
				echo "error: --ref requires a value" >&2
				exit 2
			fi
			ref="$2"
			shift 2
			;;
		--force)
			force=1
			shift
			;;
		--dry-run)
			dry_run=1
			shift
			;;
		-h|--help)
			usage
			exit 0
			;;
		*)
			echo "error: unknown argument: $1" >&2
			usage >&2
			exit 2
			;;
	esac
done

case "$assistant" in
	codex|kiro|claude|all) ;;
	*)
		echo "error: invalid assistant: $assistant" >&2
		exit 2
		;;
esac

script_path="${BASH_SOURCE[0]:-}"
repo_root=""
skills_dir=""
if [[ -n "$script_path" && -f "$script_path" ]]; then
	repo_root="$(cd "$(dirname "$script_path")/.." && pwd)"
	if [[ -d "$repo_root/skills" ]]; then
		skills_dir="$repo_root/skills"
	fi
fi

raw_base="https://raw.githubusercontent.com/awslabs/codeknit/$ref/skills"
skills=(
	"codeknit-parse"
	"codeknit-fingerprint"
)
skill_files=(
	"codeknit-parse/SKILL.md"
	"codeknit-parse/OUTPUT-FORMAT.md"
	"codeknit-fingerprint/SKILL.md"
)

run() {
	if [[ "$dry_run" -eq 1 ]]; then
		printf 'dry-run:'
		printf ' %q' "$@"
		printf '\n'
		return
	fi
	"$@"
}

target_dir() {
	case "$1" in
		codex)
			printf '%s/skills' "${CODEX_HOME:-$HOME/.codex}"
			;;
		kiro)
			printf '%s/.kiro/skills' "$HOME"
			;;
		claude)
			printf '%s/.claude/skills' "$HOME"
			;;
	esac
}

copy_skill() {
	local skill="$1"
	local dest="$2"

	if [[ -n "$skills_dir" ]]; then
		local src="$skills_dir/$skill"
		if [[ ! -d "$src" ]]; then
			echo "error: missing skill source: $src" >&2
			exit 1
		fi
		run cp -R "$src" "$dest"
		return
	fi

	run mkdir -p "$dest"
	for file in "${skill_files[@]}"; do
		[[ "$file" == "$skill/"* ]] || continue
		local name="${file#"$skill/"}"
		local url="$raw_base/$file"
		if [[ "$dry_run" -eq 1 ]]; then
			run curl -fsSL "$url" -o "$dest/$name"
		else
			curl -fsSL "$url" -o "$dest/$name"
		fi
	done
}

install_for() {
	local target="$1"
	local dest_root
	dest_root="$(target_dir "$target")"

	echo "==> Installing skills for $target into $dest_root"
	if [[ -z "$skills_dir" ]]; then
		echo "    source: $raw_base"
	fi
	run mkdir -p "$dest_root"

	for skill in "${skills[@]}"; do
		local dest="$dest_root/$skill"

		if [[ -e "$dest" ]]; then
			if [[ "$force" -ne 1 ]]; then
				echo "skip: $dest already exists (use --force to replace)"
				continue
			fi
			run rm -rf "$dest"
		fi

		copy_skill "$skill" "$dest"
		if [[ "$dry_run" -eq 1 ]]; then
			echo "would install: $dest"
		else
			echo "installed: $dest"
		fi
	done
}

if [[ "$assistant" == "all" ]]; then
	for target in codex kiro claude; do
		install_for "$target"
	done
else
	install_for "$assistant"
fi
