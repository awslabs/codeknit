/**
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';
import starlightLlmsTxt from 'starlight-llms-txt';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const sktGrammar = require('./src/skt.tmLanguage.json');

// https://astro.build/config
export default defineConfig({
	site: 'https://awslabs.github.io/codeknit',
	base: '/codeknit',
	integrations: [
		starlight({
			title: 'codeknit',
			plugins: [starlightLlmsTxt()],
			social: [{ icon: 'github', label: 'GitHub', href: 'https://github.com/awslabs/codeknit' }],
			expressiveCode: {
				shiki: {
					langs: [sktGrammar],
				},
			},
			defaultLocale: 'root',
			locales: {
				root: { label: 'English', lang: 'en' },
				fr: { label: 'Français', lang: 'fr' },
				it: { label: 'Italiano', lang: 'it' },
				es: { label: 'Español', lang: 'es' },
				de: { label: 'Deutsch', lang: 'de' },
				vi: { label: 'Tiếng Việt', lang: 'vi' },
				'zh-cn': { label: '简体中文', lang: 'zh-CN' },
				ja: { label: '日本語', lang: 'ja' },
				ko: { label: '한국어', lang: 'ko' },
			},
			sidebar: [
				{
					label: 'Start Here',
					items: [
						{ slug: 'getting-started' },
						{ slug: 'installation' },
					],
				},
				{
					label: 'Usage',
					items: [
						{ slug: 'guides/parse-command' },
						{ slug: 'guides/graph-commands' },
						{ slug: 'guides/fingerprint-command' },
						{ slug: 'guides/output-modes' },
						{ slug: 'guides/ai-assistants' },
					],
				},
				{
					label: 'Reference',
					items: [
						{ slug: 'reference/output-format' },
						{ slug: 'reference/cli-flags' },
					],
				},
			],
		}),
	],
});
