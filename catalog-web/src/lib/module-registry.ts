/**
 * Module Registry - Integrates all @vasic-digital local packages.
 *
 * Each package is imported here to ensure it is compiled and available
 * in the catalog-web bundle. Components and types are re-exported for
 * use throughout the app.
 */

// ---------------------------------------------------------------------------
// 1. Auth context from shared package
//    NOTE: catalog-web has its own AuthProvider in contexts/AuthContext.tsx.
//    The shared package (@vasic-digital/auth-context) could replace it in the
//    future once @tanstack/react-query versions are aligned (shared package
//    targets v5, catalog-web currently uses v4).
// ---------------------------------------------------------------------------
export { AuthProvider as SharedAuthProvider, useAuth as useSharedAuth } from '@vasic-digital/auth-context'
export type { AuthContextType, AuthProviderProps } from '@vasic-digital/auth-context'

// ---------------------------------------------------------------------------
// 2. API client from shared package
//    NOTE: catalog-web has its own axios instance in lib/api.ts and separate
//    per-domain API modules (mediaApi, collectionsApi, etc.). The shared
//    CatalogizerClient unifies all endpoints behind a single facade and could
//    replace the individual API modules in a future refactor.
// ---------------------------------------------------------------------------
export { CatalogizerClient, HttpClient, CatalogizerError, AuthenticationError, NetworkError, ValidationError } from '@vasic-digital/catalogizer-api-client'
export type { ClientConfig, ApiResponse } from '@vasic-digital/catalogizer-api-client'

// ---------------------------------------------------------------------------
// 3. Collection management components from shared package
//    NOTE: catalog-web has an internal Collections page (pages/Collections.tsx).
//    The shared CollectionList, CollectionCard, CollectionForm, and
//    SmartRuleBuilder components could be adopted there.
// ---------------------------------------------------------------------------
export { CollectionList, CollectionCard, CollectionForm, SmartRuleBuilder } from '@vasic-digital/collection-manager'
export type { CollectionListProps, CollectionCardProps, CollectionFormProps, SmartRuleBuilderProps } from '@vasic-digital/collection-manager'

// ---------------------------------------------------------------------------
// 4. Dashboard analytics components from shared package
//    NOTE: catalog-web has internal dashboard components in
//    components/dashboard/ (e.g. ActivityFeed). The shared StatsCard,
//    EntityStatsGrid, MediaDistributionBar, and ActivityFeed could replace
//    or supplement them.
// ---------------------------------------------------------------------------
export { StatsCard, EntityStatsGrid, MediaDistributionBar, ActivityFeed as SharedActivityFeed } from '@vasic-digital/dashboard-analytics'
export type { StatsCardProps, EntityStatsGridProps, MediaDistributionBarProps, ActivityFeedProps, ActivityItem } from '@vasic-digital/dashboard-analytics'

// ---------------------------------------------------------------------------
// 5. Media browser components from shared package
//    NOTE: catalog-web has internal entity browser components in
//    components/entity/ (EntityCard, EntityGrid, TypeSelector) and
//    pages/EntityBrowser.tsx. The shared package provides equivalent
//    components that could be swapped in.
// ---------------------------------------------------------------------------
export {
  EntityBrowser as SharedEntityBrowser,
  EntityGrid as SharedEntityGrid,
  EntityCard as SharedEntityCard,
  TypeSelector as SharedTypeSelector,
  Pagination,
} from '@vasic-digital/media-browser'
export type { EntityBrowserProps, EntityGridProps, EntityCardProps, TypeSelectorProps, PaginationProps } from '@vasic-digital/media-browser'

// ---------------------------------------------------------------------------
// 6. Media player components from shared package
//    NOTE: catalog-web has an internal MediaPlayer in components/media/
//    MediaPlayer.tsx. The shared package provides a headless useMediaPlayer
//    hook and composable MediaPlayer + PlayerControls that could replace it.
// ---------------------------------------------------------------------------
export {
  MediaPlayer as SharedMediaPlayer,
  PlayerControls,
  useMediaPlayer,
} from '@vasic-digital/media-player'
export type { MediaPlayerProps, PlayerControlsProps, MediaPlayerState, MediaPlayerControls } from '@vasic-digital/media-player'
