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
          { text: 'Quick Start (5-min)', link: '/docs/getting-started/quick-start' },
          { text: 'Configuration', link: '/docs/getting-started/configuration' },
          { text: 'First Scan', link: '/docs/getting-started/first-scan' },
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
          { text: 'Platforms Overview', link: '/platforms' },
        ]
      },
      {
        text: 'Administration',
        items: [
          { text: 'Documentation', link: '/documentation' },
          { text: 'Configuration', link: '/guides/configuration' },
          { text: 'Security', link: '/guides/security' },
          { text: 'Monitoring', link: '/guides/monitoring' },
          { text: 'Security Overview', link: '/security' },
        ]
      },
      {
        text: 'Developer',
        items: [
          { text: 'Architecture', link: '/developer/architecture' },
          { text: 'System Architecture', link: '/architecture' },
          { text: 'API Reference', link: '/developer/api' },
          { text: 'Database', link: '/docs/developer-guide/database' },
          { text: 'Filesystem Protocols', link: '/docs/developer-guide/protocols' },
          { text: 'Media Pipeline', link: '/docs/developer-guide/media-pipeline' },
          { text: 'WebSocket Events', link: '/docs/developer-guide/websockets' },
          { text: 'Testing', link: '/developer/testing' },
          { text: 'Contributing', link: '/developer/contributing' },
          { text: 'Contributing Guide', link: '/contributing' },
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
