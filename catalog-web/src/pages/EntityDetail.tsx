import { useParams, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { entityApi } from '@/lib/mediaApi'
import { Button } from '@/components/ui/Button'
import { ArrowLeft, ChevronRight, Globe, Star } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import { EntityHero, ChildrenList, FilesList, DuplicatesList } from '@/components/entity/EntityDetailView'
import toast from 'react-hot-toast'

export function EntityDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const entityId = parseInt(id || '0', 10)

  const { data: entity, isLoading } = useQuery({
    queryKey: ['entity', entityId],
    queryFn: () => entityApi.getEntity(entityId),
    enabled: entityId > 0,
  })

  const { data: childrenData } = useQuery({
    queryKey: ['entityChildren', entityId],
    queryFn: () => entityApi.getEntityChildren(entityId),
    enabled: entityId > 0 && (entity?.children_count || 0) > 0,
  })

  const { data: filesData } = useQuery({
    queryKey: ['entityFiles', entityId],
    queryFn: () => entityApi.getEntityFiles(entityId),
    enabled: entityId > 0,
  })

  const { data: duplicatesData } = useQuery({
    queryKey: ['entityDuplicates', entityId],
    queryFn: () => entityApi.getEntityDuplicates(entityId),
    enabled: entityId > 0,
  })

  const refreshMutation = useMutation({
    mutationFn: () => entityApi.refreshEntityMetadata(entityId),
    onSuccess: () => {
      toast.success('Metadata refresh queued')
      queryClient.invalidateQueries({ queryKey: ['entity', entityId] })
    },
    onError: () => toast.error('Failed to refresh metadata'),
  })

  const favoriteMutation = useMutation({
    mutationFn: (favorite: boolean) =>
      entityApi.updateUserMetadata(entityId, { favorite }),
    onSuccess: () => {
      toast.success('Updated')
      queryClient.invalidateQueries({ queryKey: ['entity', entityId] })
    },
  })

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600" />
      </div>
    )
  }

  if (!entity) {
    return (
      <div className="text-center py-12">
        <p className="text-gray-500 dark:text-gray-400">Entity not found</p>
        <Button variant="outline" className="mt-4" onClick={() => navigate('/browse')}>
          Back to Browse
        </Button>
      </div>
    )
  }

  const children = childrenData?.items || []
  const files = filesData?.files || []
  const duplicates = duplicatesData?.duplicates || []

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-2">
        <Button variant="ghost" size="icon" onClick={() => navigate(-1)}>
          <ArrowLeft className="h-5 w-5" />
        </Button>
        <span className="text-sm text-gray-500 dark:text-gray-400 capitalize">
          {entity.media_type.replace(/_/g, ' ')}
        </span>
        <ChevronRight className="h-4 w-4 text-gray-400" />
        <span className="text-sm font-medium text-gray-900 dark:text-white">
          {entity.title}
        </span>
      </div>

      <EntityHero
        entity={entity}
        files={files}
        duplicateCount={duplicates.length}
        onFavorite={() => favoriteMutation.mutate(true)}
        onRefresh={() => refreshMutation.mutate()}
        refreshPending={refreshMutation.isPending}
      />

      <ChildrenList
        children={children}
        mediaType={entity.media_type}
        onChildClick={(childId) => navigate(`/entity/${childId}`)}
      />

      <FilesList files={files} />

      {entity.external_metadata && entity.external_metadata.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Globe className="h-5 w-5" />
              External Metadata
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {entity.external_metadata.map((meta) => (
                <div key={meta.id} className="flex items-center gap-3 p-3 rounded-lg bg-gray-50 dark:bg-gray-800">
                  <div className="flex-1">
                    <span className="font-medium text-gray-900 dark:text-white capitalize">{meta.provider}</span>
                    <span className="text-sm text-gray-500 ml-2">ID: {meta.external_id}</span>
                  </div>
                  {meta.rating != null && (
                    <span className="flex items-center gap-1 text-sm">
                      <Star className="h-3 w-3 text-yellow-500" /> {meta.rating}
                    </span>
                  )}
                  {meta.review_url && (
                    <a href={meta.review_url} target="_blank" rel="noopener noreferrer" className="text-sm text-blue-600 hover:underline">
                      View
                    </a>
                  )}
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      <DuplicatesList
        duplicates={duplicates}
        onDuplicateClick={(dupId) => navigate(`/entity/${dupId}`)}
      />
    </div>
  )
}

export default EntityDetail
