package com.catalogizer.android.data.models

import android.os.Parcelable
import androidx.room.ColumnInfo
import androidx.room.Entity
import androidx.room.PrimaryKey
import androidx.room.TypeConverters
import com.catalogizer.android.data.local.Converters
import kotlinx.parcelize.Parcelize
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Parcelize
@Serializable
@Entity(tableName = "media_items")
@TypeConverters(Converters::class)
data class MediaItem(
    @PrimaryKey
    val id: Long,
    val title: String,
    @SerialName("media_type")
    @ColumnInfo(name = "media_type")
    val mediaType: String,
    val year: Int? = null,
    val description: String? = null,
    @SerialName("cover_image")
    @ColumnInfo(name = "cover_image")
    val coverImage: String? = null,
    val rating: Double? = null,
    val quality: String? = null,
    @SerialName("file_size")
    @ColumnInfo(name = "file_size")
    val fileSize: Long? = null,
    val duration: Int? = null,
    @SerialName("directory_path")
    @ColumnInfo(name = "directory_path")
    val directoryPath: String,
    @SerialName("smb_path")
    @ColumnInfo(name = "smb_path")
    val smbPath: String? = null,
    @SerialName("created_at")
    @ColumnInfo(name = "created_at")
    val createdAt: String,
    @SerialName("updated_at")
    @ColumnInfo(name = "updated_at")
    val updatedAt: String,
    @SerialName("external_metadata")
    @ColumnInfo(name = "external_metadata")
    val externalMetadata: List<ExternalMetadata>? = null,
    val versions: List<MediaVersion>? = null,
    @ColumnInfo(name = "is_favorite")
    val isFavorite: Boolean = false,
    @ColumnInfo(name = "watch_progress")
    val watchProgress: Double = 0.0,
    @ColumnInfo(name = "last_watched")
    val lastWatched: String? = null,
    @ColumnInfo(name = "is_downloaded")
    val isDownloaded: Boolean = false
) : Parcelable

@Parcelize
@Serializable
data class ExternalMetadata(
    val id: Long,
    @SerialName("media_id")
    val mediaId: Long,
    val provider: String,
    @SerialName("external_id")
    val externalId: String,
    val title: String,
    val description: String? = null,
    val year: Int? = null,
    val rating: Double? = null,
    @SerialName("poster_url")
    val posterUrl: String? = null,
    @SerialName("backdrop_url")
    val backdropUrl: String? = null,
    val genres: List<String>? = null,
    val cast: List<String>? = null,
    val crew: List<String>? = null,
    val metadata: Map<String, String>? = null,
    @SerialName("last_updated")
    val lastUpdated: String
) : Parcelable

@Parcelize
@Serializable
data class MediaVersion(
    val id: Long,
    @SerialName("media_id")
    val mediaId: Long,
    val version: String,
    val quality: String,
    @SerialName("file_path")
    val filePath: String,
    @SerialName("file_size")
    val fileSize: Long,
    val codec: String? = null,
    val resolution: String? = null,
    val bitrate: Long? = null,
    val language: String? = null
) : Parcelable

@Serializable
data class MediaSearchRequest(
    val query: String? = null,
    @SerialName("media_type")
    val mediaType: String? = null,
    @SerialName("year_min")
    val yearMin: Int? = null,
    @SerialName("year_max")
    val yearMax: Int? = null,
    @SerialName("rating_min")
    val ratingMin: Double? = null,
    val quality: String? = null,
    @SerialName("sort_by")
    val sortBy: String? = null,
    @SerialName("sort_order")
    val sortOrder: String? = null,
    val limit: Int = 20,
    val offset: Int = 0
)

@Serializable
data class MediaSearchResponse(
    val items: List<MediaItem>,
    val total: Int,
    val limit: Int,
    val offset: Int
)

@Serializable
data class MediaStats(
    @SerialName("total_items")
    val totalItems: Int,
    @SerialName("by_type")
    val byType: Map<String, Int>,
    @SerialName("by_quality")
    val byQuality: Map<String, Int>,
    @SerialName("total_size")
    val totalSize: Long,
    @SerialName("recent_additions")
    val recentAdditions: Int
)

enum class MediaType(val value: String, val displayName: String) {
    MOVIE("movie", "Movies"),
    TV_SHOW("tv_show", "TV Shows"),
    DOCUMENTARY("documentary", "Documentaries"),
    ANIME("anime", "Anime"),
    MUSIC("music", "Music"),
    AUDIOBOOK("audiobook", "Audiobooks"),
    PODCAST("podcast", "Podcasts"),
    GAME("game", "Games"),
    SOFTWARE("software", "Software"),
    EBOOK("ebook", "E-books"),
    TRAINING("training", "Training"),
    CONCERT("concert", "Concerts"),
    YOUTUBE_VIDEO("youtube_video", "YouTube"),
    SPORTS("sports", "Sports"),
    NEWS("news", "News"),
    OTHER("other", "Other");

    companion object {
        fun fromValue(value: String): MediaType {
            return values().find { it.value == value } ?: OTHER
        }

        fun getAllTypes(): List<MediaType> = values().toList()
    }
}

enum class QualityLevel(val value: String, val displayName: String) {
    CAM("cam", "CAM"),
    TS("ts", "TS"),
    DVDRIP("dvdrip", "DVD-Rip"),
    BRRIP("brrip", "BR-Rip"),
    HD_720P("720p", "720p HD"),
    HD_1080P("1080p", "1080p HD"),
    UHD_4K("4k", "4K UHD"),
    HDR("hdr", "HDR"),
    DOLBY_VISION("dolby_vision", "Dolby Vision");

    companion object {
        fun fromValue(value: String): QualityLevel? {
            return values().find { it.value == value }
        }

        fun getAllQualities(): List<QualityLevel> = values().toList()
    }
}

enum class SortOption(val value: String, val displayName: String) {
    TITLE("title", "Title"),
    YEAR("year", "Year"),
    RATING("rating", "Rating"),
    UPDATED_AT("updated_at", "Recently Updated"),
    CREATED_AT("created_at", "Recently Added"),
    FILE_SIZE("file_size", "File Size");

    companion object {
        fun fromValue(value: String): SortOption {
            return values().find { it.value == value } ?: UPDATED_AT
        }
    }
}

enum class SortOrder(val value: String, val displayName: String) {
    ASC("asc", "Ascending"),
    DESC("desc", "Descending");

    companion object {
        fun fromValue(value: String): SortOrder {
            return values().find { it.value == value } ?: DESC
        }
    }
}