import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'Warp',
  description: 'Go TUI Layout Engine — tabs, splits, flex, floats, modals',
  lang: 'en-US',
  base: '/',

  themeConfig: {
    nav: [
      { text: 'Guide', link: '/guide/getting-started' },
      { text: 'API', link: '/api/' },
      { text: 'Specs', link: '/specs/api' },
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
            { text: 'Word Wrap', link: '/api/wrap' },
            { text: 'Styles', link: '/api/styles' },
          ],
        },
      ],
      '/specs/': [
        {
          text: 'Specifications',
          items: [
            { text: 'API', link: '/specs/api' },
            { text: 'Architecture', link: '/specs/architecture' },
            { text: 'Integration', link: '/specs/integration' },
            { text: 'Patterns', link: '/specs/patterns' },
            { text: 'State', link: '/specs/state' },
          ],
        },
      ],
    },

    socialLinks: [
      { icon: 'github', link: 'https://github.com/starframe-dev/warp' },
    ],

    footer: {
      message: 'Released under the MIT License.',
    },
  },
})
