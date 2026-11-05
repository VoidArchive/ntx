import prettier from 'eslint-config-prettier';
import { fileURLToPath } from 'node:url';
import { includeIgnoreFile } from '@eslint/compat';
import js from '@eslint/js';
import svelte from 'eslint-plugin-svelte';
import { defineConfig } from 'eslint/config';
import globals from 'globals';
import ts from 'typescript-eslint';
import { loadConfig } from '@sveltejs/load-config';

const svelteConfig = (await loadConfig('./', { traverse: false }))?.config;

const gitignorePath = fileURLToPath(new URL('./.gitignore', import.meta.url));

export default defineConfig(
	includeIgnoreFile(gitignorePath),
	{ ignores: ['src/lib/gen/**'] },
	js.configs.recommended,
	...ts.configs.recommended,
	...svelte.configs.recommended,
	prettier,
	...svelte.configs.prettier,
	{
		languageOptions: { globals: { ...globals.browser, ...globals.node } },

		rules: {
			// typescript-eslint strongly recommend that you do not use the no-undef lint rule on TypeScript projects.
			// see: https://typescript-eslint.io/troubleshooting/faqs/eslint/#i-get-errors-from-the-no-undef-rule-about-global-variables-not-being-defined-even-though-there-are-no-typescript-errors
			'no-undef': 'off',
			// Mechanical half of CODING_STANDARDS.md; svelte/prefer-const leaves $props and $derived alone.
			'svelte/prefer-const': 'error',
			'no-else-return': ['error', { allowElseIf: false }],
			'svelte/no-navigation-without-resolve': 'off',
			'@typescript-eslint/no-unused-vars': [
				'error',
				{
					argsIgnorePattern: '^_',
					varsIgnorePattern: '^_',
					caughtErrorsIgnorePattern: '^_'
				}
			]
		}
	},
	{
		files: ['**/*.svelte', '**/*.svelte.ts', '**/*.svelte.js'],

		rules: {
			// Claims read as facts, never as advice: no recommending words in rendered text.
			'no-restricted-syntax': [
				'error',
				{
					selector:
						':not(SvelteElement[name.name="script"]) > SvelteText[value=/\\b(buy|sell|bullish|bearish|target|signal|score)\\b/i]',
					message: 'Recommending word in user-facing copy (CODING_STANDARDS.md).'
				},
				{
					selector:
						'SvelteAttribute[key.name=/^(title|alt|placeholder|aria-label|aria-description)$/] > SvelteLiteral[value=/\\b(buy|sell|bullish|bearish|target|signal|score)\\b/i]',
					message: 'Recommending word in user-facing copy (CODING_STANDARDS.md).'
				}
			]
		},

		languageOptions: {
			parserOptions: {
				projectService: true,
				extraFileExtensions: ['.svelte'],
				parser: ts.parser,
				svelteConfig
			}
		}
	}
);
