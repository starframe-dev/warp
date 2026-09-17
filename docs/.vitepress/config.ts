import { readdir, readFile } from 'node:fs/promises'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import type { Plugin } from 'vite'
import { defineConfig } from 'vitepress'

const docsRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')

async function htmlFiles(root: string): Promise<string[]> {
  const entries = await readdir(root, { withFileTypes: true }).catch(() => [])
  const files: string[] = []

  for (const entry of entries) {
    const absolutePath = path.join(root, entry.name)
    if (entry.isDirectory()) {
      files.push(...await htmlFiles(absolutePath))
    } else if (entry.isFile() && entry.name.endsWith('.html')) {
      files.push(absolutePath)
    }
  }

  return files
}

function apiHtmlPlugin(): Plugin {
  return {
    name: 'copy-api-html',
    async generateBundle() {
      const roots = [
        { source: path.join(docsRoot, 'api'), target: 'api' },
        { source: path.join(docsRoot, 'ru', 'api'), target: 'ru/api' },
      ]

      for (const root of roots) {
        for (const absolutePath of await htmlFiles(root.source)) {
          const relativePath = path.relative(root.source, absolutePath).split(path.sep).join('/')
          this.emitFile({
            type: 'asset',
            fileName: path.posix.join(root.target, relativePath),
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
              ],
            },
          ],
          '/api/': [
            {
              text: 'API Reference',
              items: [
                { text: 'Overview', link: '/api/' },
                { text: 'Warp', link: '/api/warp.html' },
                { text: 'TabGroup', link: '/api/tabgroup.html' },
                { text: 'Tab', link: '/api/tab.html' },
                { text: 'Panel', link: '/api/panel.html' },
                { text: 'Split & Flex', link: '/api/split.html' },
                { text: 'Float', link: '/api/float.html' },
                { text: 'Collapsible', link: '/api/collapsible.html' },
                { text: 'Scrollable', link: '/api/scrollable.html' },
                { text: 'Dropdown', link: '/api/dropdown.html' },
                { text: 'Selectable', link: '/api/selectable.html' },
                { text: 'Input', link: '/api/input.html' },
                { text: 'Modal', link: '/api/modal.html' },
                { text: 'Popover', link: '/api/popover.html' },
                { text: 'Focus', link: '/api/focus.html' },
                { text: 'Element tree', link: '/api/element.html' },
                { text: 'Word Wrap', link: '/api/wrap.html' },
                { text: 'Styles', link: '/api/styles.html' },
                { text: 'Theme', link: '/api/theme.html' },
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
              ],
            },
          ],
          '/ru/api/': [
            {
              text: 'API',
              items: [
                { text: 'Обзор', link: '/ru/api/' },
                { text: 'Warp', link: '/ru/api/warp.html' },
                { text: 'TabGroup', link: '/ru/api/tabgroup.html' },
                { text: 'Tab', link: '/ru/api/tab.html' },
                { text: 'Panel', link: '/ru/api/panel.html' },
                { text: 'Split & Flex', link: '/ru/api/split.html' },
                { text: 'Float', link: '/ru/api/float.html' },
                { text: 'Collapsible', link: '/ru/api/collapsible.html' },
                { text: 'Scrollable', link: '/ru/api/scrollable.html' },
                { text: 'Dropdown', link: '/ru/api/dropdown.html' },
                { text: 'Selectable', link: '/ru/api/selectable.html' },
                { text: 'Input', link: '/ru/api/input.html' },
                { text: 'Modal', link: '/ru/api/modal.html' },
                { text: 'Popover', link: '/ru/api/popover.html' },
                { text: 'Focus', link: '/ru/api/focus.html' },
                { text: 'Дерево элементов', link: '/ru/api/element.html' },
                { text: 'Word Wrap', link: '/ru/api/wrap.html' },
                { text: 'Стили', link: '/ru/api/styles.html' },
                { text: 'Тема', link: '/ru/api/theme.html' },
              ],
            },
          ],
        },
      },
    },
  },

  vite: {
    plugins: [apiHtmlPlugin()],
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
