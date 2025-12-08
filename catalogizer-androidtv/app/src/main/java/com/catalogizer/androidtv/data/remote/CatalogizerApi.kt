package com.catalogizer.androidtv.data.remote

import com.catalogizer.androidtv.data.models.MediaItem
import com.catalogizer.androidtv.data.models.MediaSearchRequest
import retrofit2.Response
import retrofit2.http.*

interface CatalogizerApi {

    // Authentication endpoints
    @POST("api/v1/auth/login")
    suspend fun login(@Body credentials: Map<String, String>): Response<LoginResponse>

    @POST("api/v1/auth/refresh")
    suspend fun refreshToken(@Body token: Map<String, String>): Response<LoginResponse>

    // Catalog endpoints
    @GET("api/v1/catalog")
    suspend fun getCatalog(): Response<List<String>>

    @GET("api/v1/search")
    suspend fun searchMedia(@QueryMap params: Map<String, String>): Response<List<MediaItem>>

    @GET("api/v1/media/{id}")
    suspend fun getMediaById(@Path("id") id: Long): Response<MediaItem>

    @GET("api/v1/catalog-info/{path}")
    suspend fun getMediaInfo(@Path("path") path: String): Response<MediaItem>

    @POST("api/v1/media/recognize")
    suspend fun recognizeMedia(@Body request: Map<String, Any>): Response<MediaItem>

    // Media management endpoints
    @PUT("api/v1/media/{id}/progress")
    suspend fun updateWatchProgress(@Path("id") id: Long, @Body progress: Map<String, Double>): Response<Unit>

    @PUT("api/v1/media/{id}/favorite")
    suspend fun updateFavoriteStatus(@Path("id") id: Long, @Body favorite: Map<String, Boolean>): Response<Unit>
}

// Auth response models
@kotlinx.serialization.Serializable
data class LoginResponse(
    val token: String,
    @kotlinx.serialization.SerialName("user_id")
    val userId: Long,
    val username: String,
    @kotlinx.serialization.SerialName("expires_at")
    val expiresAt: String? = null
)