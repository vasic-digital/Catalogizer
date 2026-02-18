const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080'

export interface AssetInfo {
  id: string
  type: string
  status: 'pending' | 'resolving' | 'ready' | 'failed' | 'expired'
  content_type: string
  size: number
  entity_type: string
  entity_id: string
  created_at: string
  updated_at: string
  resolved_at?: string
}

export interface AssetRequest {
  type: string
  source_hint?: string
  entity_type: string
  entity_id: string
}

export function assetUrl(assetId: string): string {
  return `${API_BASE}/api/v1/assets/${assetId}`
}

export async function getAssetsByEntity(
  entityType: string,
  entityId: string | number,
  token?: string
): Promise<AssetInfo[]> {
  const headers: Record<string, string> = {}
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  const response = await fetch(
    `${API_BASE}/api/v1/assets/by-entity/${entityType}/${entityId}`,
    { headers }
  )

  if (!response.ok) {
    return []
  }

  return response.json()
}

export async function requestAsset(
  req: AssetRequest,
  token?: string
): Promise<{ asset_id: string }> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  }
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  const response = await fetch(`${API_BASE}/api/v1/assets/request`, {
    method: 'POST',
    headers,
    body: JSON.stringify(req),
  })

  if (!response.ok) {
    throw new Error(`Asset request failed: ${response.status}`)
  }

  return response.json()
}
