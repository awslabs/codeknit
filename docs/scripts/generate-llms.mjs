/**
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { promises as fs } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const docsRoot = path.resolve(__dirname, '..');
const contentRoot = path.join(docsRoot, 'src', 'content', 'docs');
const distRoot = path.join(docsRoot, 'dist');

const siteBase = 'https://awslabs.github.io/codeknit';
const siteTitle = 'codeknit';
const siteDescription = 'Parse source code into compact structural maps that LLMs can actually use.';

const localeDirs = ['fr', 'it', 'es', 'de', 'vi', 'zh-cn', 'ja', 'ko'];
const pageOrder = [
	'index.mdx',
	'getting-started.md',
	'installation.md',
	'guides/parse-command.md',
	'guides/graph-commands.md',
	'guides/fingerprint-command.md',
	'guides/output-modes.md',
	'guides/ai-assistants.md',
	'reference/output-format.md',
	'reference/cli-flags.md',
];

async function main() {
	await removeStalePluginOutput();
	const rootPages = await readPages(contentRoot);

	await writeText('llms.txt', renderIndex());
	await writeText('llms-full.txt', renderDocumentSet(rootPages, 'full', 'en'));
	await writeText('llms-small.txt', renderDocumentSet(rootPages, 'small', 'en'));

	for (const locale of localeDirs) {
		const localeRoot = path.join(contentRoot, locale);
		const pages = await readPages(localeRoot);
		if (pages.length === 0) continue;
		await writeText(path.join(locale, 'llms-full.txt'), renderDocumentSet(pages, 'full', locale));
		await writeText(path.join(locale, 'llms-small.txt'), renderDocumentSet(pages, 'small', locale));
	}
}

async function removeStalePluginOutput() {
	await fs.rm(path.join(distRoot, '_llms-txt'), { recursive: true, force: true });
}

async function readPages(root) {
	const pages = [];
	for (const relativePath of pageOrder) {
		const filePath = path.join(root, relativePath);
		try {
			const source = await fs.readFile(filePath, 'utf8');
			pages.push(parsePage(source, relativePath));
		} catch (error) {
			if (error.code !== 'ENOENT') throw error;
		}
	}
	return pages;
}

function parsePage(source, relativePath) {
	const frontmatter = source.match(/^---\n([\s\S]*?)\n---\n?/);
	const data = frontmatter ? parseFrontmatter(frontmatter[1]) : {};
	const rawBody = frontmatter ? source.slice(frontmatter[0].length) : source;
	const title = data.title || titleFromPath(relativePath);
	const description = data.description || (relativePath === 'index.mdx' ? siteDescription : '');
	const body = cleanMarkdown(rawBody, title, description);
	return { title, description, body };
}

function parseFrontmatter(source) {
	const data = {};
	for (const line of source.split('\n')) {
		const match = line.match(/^\s*([A-Za-z0-9_-]+):\s*(.*)$/);
		if (!match) continue;
		const [, key, value] = match;
		if (value === '') continue;
		data[key] = stripQuotes(value);
	}
	return data;
}

function cleanMarkdown(source, title, description) {
	let body = source
		.replace(/^import\s+.*$/gm, '')
		.replace(/<!--[\s\S]*?-->/g, '')
		.replace(/<Card\s+title="([^"]+)"[^>]*>/g, '\n\n## $1\n')
		.replace(/<\/?CardGrid>/g, '')
		.replace(/<\/Card>/g, '')
		.replace(/<[^>]+>/g, '')
		.trim();

	if (body.length === 0 && description) body = description;
	if (!body.startsWith('# ')) body = `# ${title}\n\n${description ? `> ${description}\n\n` : ''}${body}`;
	return body.replace(/\n{3,}/g, '\n\n').trim();
}

function renderIndex() {
	const fullUrl = `${siteBase}/llms-full.txt`;
	const smallUrl = `${siteBase}/llms-small.txt`;
	return [
		`# ${siteTitle}`,
		'',
		`> ${siteDescription}`,
		'',
		'## Documentation Sets',
		'',
		`- [Abridged documentation](${smallUrl}): a compact version of the documentation for ${siteTitle}`,
		`- [Complete documentation](${fullUrl}): the full documentation for ${siteTitle}`,
		'',
		'## Notes',
		'',
		'- The complete documentation includes all English documentation content.',
		'- Locale-specific complete and abridged files are available under each locale path.',
		'',
	].join('\n');
}

function renderDocumentSet(pages, mode, locale) {
	const kind = mode === 'small' ? 'abridged' : 'full';
	const system = `<SYSTEM>This is the ${kind} developer documentation for ${siteTitle}${locale === 'en' ? '' : ` (${locale})`}</SYSTEM>`;
	const content = pages.map((page) => renderPage(page, mode)).join('\n\n---\n\n');
	return `${system}\n\n${content}\n`;
}

function renderPage(page, mode) {
	if (mode === 'full') return page.body;
	const compactBody = page.body
		.replace(/\n{2,}/g, '\n')
		.replace(/[ \t]+/g, ' ')
		.trim();
	return compactBody;
}

async function writeText(relativePath, content) {
	const outputPath = path.join(distRoot, relativePath);
	await fs.mkdir(path.dirname(outputPath), { recursive: true });
	await fs.writeFile(outputPath, content, 'utf8');
}

function stripQuotes(value) {
	return value.replace(/^['"]|['"]$/g, '').trim();
}

function titleFromPath(relativePath) {
	const name = path.basename(relativePath, path.extname(relativePath));
	return name
		.split('-')
		.map((word) => word.charAt(0).toUpperCase() + word.slice(1))
		.join(' ');
}

await main();
