package com.catalogizer.androidtv.data.remote

import com.catalogizer.androidtv.data.models.MediaItem
import com.catalogizer.androidtv.data.models.MediaSearchRequest
import retrofit2.Response
import retrofit2.http.*

interface CatalogizerApi {

    @GET("api/v1/catalog")
    suspend fun getCatalog(): Response<List<String>>

    @GET("api/v1/search")
    suspend fun searchMedia(@QueryMap params: Map<String, String>): Response<List<MediaItem>>

    @GET("api/v1/catalog-info/{path}")
    suspend fun getMediaInfo(@Path("path") path: String): Response<MediaItem>

    @POST("api/v1/media/recognize")
    suspend fun recognizeMedia(@Body request: Map<String, Any>): Response<MediaItem>

    @PUT("api/v1/media/{id}/progress")
    suspend fun updateWatchProgress(@Path("id") id: Long, @Body progress: Map<String, Double>): Response<Unit>

    @PUT("api/v1/media/{id}/favorite")
    suspend fun updateFavoriteStatus(@Path("id") id: Long, @Body favorite: Map<String, Boolean>): Response<Unit>
}