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

  it('exports CatalogizerClient from catalogizer-api-client', async () => {
    const mod = await import('../module-registry')
    expect(mod.CatalogizerClient).toBeDefined()
    expect(typeof mod.CatalogizerClient).toBe('function')
  })

  it('exports HttpClient from catalogizer-api-client', async () => {
    const mod = await import('../module-registry')
    expect(mod.HttpClient).toBeDefined()
    expect(typeof mod.HttpClient).toBe('function')
  })

  it('exports error classes from catalogizer-api-client', async () => {
    const mod = await import('../module-registry')
    expect(mod.CatalogizerError).toBeDefined()
    expect(mod.AuthenticationError).toBeDefined()
    expect(mod.NetworkError).toBeDefined()
    expect(mod.ValidationError).toBeDefined()
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
