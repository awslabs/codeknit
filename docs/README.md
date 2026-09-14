# Codeknit documentation

The site uses Astro Starlight. Run these commands from the repository root:

| Command               | Action                                                                        |
| --------------------- | ----------------------------------------------------------------------------- |
| `make docs-generate`  | Copy maintained references, then generate English pages with Bedrock          |
| `make docs-translate` | Copy maintained references, then translate changed English pages with Bedrock |
| `make docs-build`     | Build the site and `llms*.txt` bundles locally                                |
| `make docs`           | Generate, translate, and build in sequence                                    |
| `make docs-preview`   | Start the local preview server                                                |

## Sources and generated output

- `prompts/*.txt` controls model-generated pages registered in `gen/main.go`. Edit the prompt or its source context when changing generated documentation. `DOCS_FORCE=1 make docs` regenerates these pages even when they already exist.
- `reference/language-support.md` is the maintained source for the tested capability matrix. `reference/analysis-reliability.md` is the maintained scan, output-publication, and duplicate-retrieval contract. Both use the same verbatim-copy mechanism. The generator copies it verbatim to `src/content/docs/reference/language-support.md` on every run, before generation or translation. It is never rewritten by the English model, including with `--force`; a missing or stale output copy is restored. Edit the maintained source, not the output copy. To refresh only the maintained copies without AWS credentials or model calls, run `go run ./docs/gen --skip-generate --skip-translate`.
- All registered pages, including maintained references, participate in translation and content-hash invalidation. Generated translations live under `src/content/docs/<locale>/`.
- `astro.config.mjs` defines site navigation. `scripts/generate-llms.mjs` defines the page order in the agent documentation bundles. Add new pages to both. Until translations exist, the site and bundles use English fallback content.
- `dist/` is disposable build output. The build reads `src/content/docs/` and does not rewrite the Markdown sources.

The deployed GitHub Pages workflow builds checked-in content; it does not call Bedrock. Include the maintained source and its synchronized output copy in a change. Generator regression tests verify they match and that forced generation cannot enqueue the maintained page for model generation:

```bash
go test ./docs/gen
```
