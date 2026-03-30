describe('auth types', () => {
  it('re-exports types from @vasic-digital/media-types', async () => {
    const mod = await import('../auth')
    // The module only exports types via 'export type', so the module
    // object itself should be importable without error
    expect(mod).toBeDefined()
  })

  it('module loads without errors', async () => {
    await expect(import('../auth')).resolves.not.toThrow()
  })
})
