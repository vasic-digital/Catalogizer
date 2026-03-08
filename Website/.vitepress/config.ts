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
      { text: 'Video Course', link: '/course' },
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
        text: 'User Guide',
        items: [
          { text: 'Features', link: '/features' },
          { text: 'Web App Guide', link: '/guides/web-app' },
          { text: 'Desktop Guide', link: '/guides/desktop' },
          { text: 'Android Guide', link: '/guides/android' },
          { text: 'Android TV Guide', link: '/guides/android-tv' },
        ]
      },
      {
        text: 'Administration',
        items: [
          { text: 'Documentation', link: '/documentation' },
          { text: 'Configuration', link: '/guides/configuration' },
          { text: 'Security', link: '/guides/security' },
          { text: 'Monitoring', link: '/guides/monitoring' },
        ]
      },
      {
        text: 'Developer',
        items: [
          { text: 'Architecture', link: '/developer/architecture' },
          { text: 'API Reference', link: '/developer/api' },
          { text: 'Security', link: '/docs/developer-guide/security' },
          { text: 'Monitoring', link: '/docs/developer-guide/monitoring' },
          { text: 'API Reference (Full)', link: '/docs/developer-guide/api-reference' },
          { text: 'Testing', link: '/developer/testing' },
          { text: 'Contributing', link: '/developer/contributing' },
        ]
      },
      {
        text: 'Resources',
        items: [
          { text: 'Video Course', link: '/course' },
          { text: 'FAQ', link: '/faq' },
          { text: 'Changelog', link: '/changelog' },
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
