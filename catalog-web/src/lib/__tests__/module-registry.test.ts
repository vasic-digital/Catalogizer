describe('module-registry', () => {
  it('exports SharedAuthProvider from auth-context', async () => {
    const mod = await import('../module-registry')
    expect(mod.SharedAuthProvider).toBeDefined()
    expect(typeof mod.SharedAuthProvider).toBe('function')
  })

  it('exports useSharedAuth from auth-context', async () => {
    const mod = await import('../module-registry')
    expect(mod.useSharedAuth).toBeDefined()
    expect(typeof mod.useSharedAuth).toBe('function')
  })

  it('exports type-only API client types from catalogizer-api-client', async () => {
    // Runtime exports (CatalogizerClient, HttpClient, error classes) were
    // removed because the shared package uses Node.js EventEmitter which is
    // not available in the browser bundle. Only type exports remain
    // (ClientConfig, ApiResponse). Verify the module still loads cleanly.
    const mod = await import('../module-registry')
    expect(mod).toBeDefined()
  })

  it('exports collection-manager components', async () => {
    const mod = await import('../module-registry')
    expect(mod.CollectionList).toBeDefined()
    expect(mod.CollectionCard).toBeDefined()
    expect(mod.CollectionForm).toBeDefined()
    expect(mod.SmartRuleBuilder).toBeDefined()
  })

  it('exports dashboard-analytics components', async () => {
    const mod = await import('../module-registry')
    expect(mod.StatsCard).toBeDefined()
    expect(mod.EntityStatsGrid).toBeDefined()
    expect(mod.MediaDistributionBar).toBeDefined()
    expect(mod.SharedActivityFeed).toBeDefined()
  })

  it('exports media-browser components', async () => {
    const mod = await import('../module-registry')
    expect(mod.SharedEntityBrowser).toBeDefined()
    expect(mod.SharedEntityGrid).toBeDefined()
    expect(mod.SharedEntityCard).toBeDefined()
    expect(mod.SharedTypeSelector).toBeDefined()
    expect(mod.Pagination).toBeDefined()
  })

  it('exports media-player components', async () => {
    const mod = await import('../module-registry')
    expect(mod.SharedMediaPlayer).toBeDefined()
    expect(mod.PlayerControls).toBeDefined()
    expect(mod.useMediaPlayer).toBeDefined()
  })
})
