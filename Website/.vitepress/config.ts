import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'Catalogizer',
  description: 'Multi-platform media collection manager',
  ignoreDeadLinks: true,
  themeConfig: {
    nav: [
      { text: 'Home', link: '/' },
      { text: 'Features', link: '/features' },
      { text: 'Download', link: '/download' },
      { text: 'Documentation', link: '/documentation' },
      { text: 'FAQ', link: '/faq' },
      { text: 'Support', link: '/support' },
    ],
    sidebar: [
      {
        text: 'Getting Started',
        items: [
          { text: 'Introduction', link: '/' },
          { text: 'Download', link: '/download' },
          { text: 'Quick Start', link: '/getting-started' },
        ]
      },
      {
        text: 'Learn More',
        items: [
          { text: 'Features', link: '/features' },
          { text: 'Documentation', link: '/documentation' },
          { text: 'FAQ', link: '/faq' },
          { text: 'Changelog', link: '/changelog' },
        ]
      },
      {
        text: 'Community',
        items: [
          { text: 'Support', link: '/support' },
        ]
      }
    ],
    socialLinks: [
      { icon: 'github', link: 'https://github.com/vasic-digital/Catalogizer' }
    ],
    footer: {
      message: 'Released under the MIT License.',
      copyright: 'Copyright 2024-2026 Vasic Digital'
    }
  }
})
