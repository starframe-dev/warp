import { readdir, readFile } from 'node:fs/promises'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import type { Plugin } from 'vite'
import { defineConfig } from 'vitepress'

const docsRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')

async function generatedFiles(root: string): Promise<string[]> {
  const entries = await readdir(root, { withFileTypes: true }).catch(() => [])
  const files: string[] = []

  for (const entry of entries) {
    const absolutePath = path.join(root, entry.name)
    if (entry.isDirectory()) {
      files.push(...await generatedFiles(absolutePath))
      continue
    }
    if (entry.isFile() && entry.name.endsWith('.html')) {
      files.push(absolutePath)
    }
  }

  return files
}

function codeCheckDocsPlugin(): Plugin {
  return {
    name: 'copy-code-check-docs',
    async generateBundle() {
      for (const language of ['en', 'ru']) {
        const languageRoot = path.join(docsRoot, language)
        const files = await generatedFiles(languageRoot)

        for (const absolutePath of files) {
          const relativePath = path.relative(languageRoot, absolutePath).split(path.sep).join('/')
          this.emitFile({
            type: 'asset',
            fileName: path.posix.join(language, relativePath),
            source: await readFile(absolutePath),
          })
        }
      }
    },
  }
}

export default defineConfig({
  title: 'Warp',
  description: 'Go TUI Layout Engine — tabs, splits, flex, floats, modals, popover',
  lang: 'en-US',
  base: '/warp/docs/',

  locales: {
    root: {
      label: 'English',
      lang: 'en-US',
      link: '/',
      themeConfig: {
        nav: [
          { text: 'Guide', link: '/guide/getting-started' },
          { text: 'API', link: '/api/' },
          { text: 'GitHub', link: 'https://github.com/starframe-dev/warp' },
        ],
        sidebar: {
          '/guide/': [
            {
              text: 'Guide',
              items: [
                { text: 'Getting Started', link: '/guide/getting-started' },
                { text: 'Architecture', link: '/guide/architecture' },
                { text: 'Layouts', link: '/guide/layouts' },
                { text: 'Components', link: '/guide/components' },
                { text: 'Focus & Input', link: '/guide/focus-input' },
                { text: 'Generated source reference', link: '/guide/generated-reference' },
              ],
            },
          ],
          '/api/': [
            {
              text: 'API Reference',
              items: [
                { text: 'Overview', link: '/api/' },
                { text: 'Warp', link: '/api/warp' },
                { text: 'TabGroup', link: '/api/tabgroup' },
                { text: 'Tab', link: '/api/tab' },
                { text: 'Panel', link: '/api/panel' },
                { text: 'Split & Flex', link: '/api/split' },
                { text: 'Float', link: '/api/float' },
                { text: 'Collapsible', link: '/api/collapsible' },
                { text: 'Scrollable', link: '/api/scrollable' },
                { text: 'Dropdown', link: '/api/dropdown' },
                { text: 'Selectable', link: '/api/selectable' },
                { text: 'Input', link: '/api/input' },
                { text: 'Modal', link: '/api/modal' },
                { text: 'Popover', link: '/api/popover' },
                { text: 'Focus', link: '/api/focus' },
                { text: 'Element tree', link: '/api/element' },
                { text: 'Word Wrap', link: '/api/wrap' },
                { text: 'Styles', link: '/api/styles' },
                { text: 'Theme', link: '/api/theme' },
              ],
            },
          ],
        },
      },
    },
    ru: {
      label: 'Русский',
      lang: 'ru-RU',
      link: '/ru/',
      themeConfig: {
        nav: [
          { text: 'Руководство', link: '/ru/guide/getting-started' },
          { text: 'API', link: '/ru/api/' },
          { text: 'GitHub', link: 'https://github.com/starframe-dev/warp' },
        ],
        sidebar: {
          '/ru/guide/': [
            {
              text: 'Руководство',
              items: [
                { text: 'Начало работы', link: '/ru/guide/getting-started' },
                { text: 'Архитектура', link: '/ru/guide/architecture' },
                { text: 'Компоновка', link: '/ru/guide/layouts' },
                { text: 'Компоненты', link: '/ru/guide/components' },
                { text: 'Фокус и ввод', link: '/ru/guide/focus-input' },
                { text: 'Сгенерированный справочник', link: '/ru/guide/generated-reference' },
              ],
            },
          ],
          '/ru/api/': [
            {
              text: 'API',
              items: [
                { text: 'Обзор', link: '/ru/api/' },
                { text: 'Warp', link: '/ru/api/warp' },
                { text: 'TabGroup', link: '/ru/api/tabgroup' },
                { text: 'Tab', link: '/ru/api/tab' },
                { text: 'Panel', link: '/ru/api/panel' },
                { text: 'Split & Flex', link: '/ru/api/split' },
                { text: 'Float', link: '/ru/api/float' },
                { text: 'Collapsible', link: '/ru/api/collapsible' },
                { text: 'Scrollable', link: '/ru/api/scrollable' },
                { text: 'Dropdown', link: '/ru/api/dropdown' },
                { text: 'Selectable', link: '/ru/api/selectable' },
                { text: 'Input', link: '/ru/api/input' },
                { text: 'Modal', link: '/ru/api/modal' },
                { text: 'Popover', link: '/ru/api/popover' },
                { text: 'Focus', link: '/ru/api/focus' },
                { text: 'Дерево элементов', link: '/ru/api/element' },
                { text: 'Word Wrap', link: '/ru/api/wrap' },
                { text: 'Стили', link: '/ru/api/styles' },
                { text: 'Тема', link: '/ru/api/theme' },
              ],
            },
          ],
        },
      },
    },
  },

  vite: {
    plugins: [codeCheckDocsPlugin()],
  },

  themeConfig: {
    socialLinks: [
      { icon: 'github', link: 'https://github.com/starframe-dev/warp' },
    ],
    footer: {
      message: 'Released under the MIT License.',
    },
  },
})
