package com.catalogizer.androidtv.data.local

import androidx.room.TypeConverter
import kotlinx.serialization.encodeToString
import kotlinx.serialization.json.Json

object Converters {
    @TypeConverter
    @JvmStatic
    fun fromStringList(value: List<String>?): String? {
        return value?.let { Json.encodeToString(it) }
    }

    @TypeConverter
    @JvmStatic
    fun toStringList(value: String?): List<String>? {
        return value?.let { Json.decodeFromString<List<String>>(it) }
    }

    @TypeConverter
    @JvmStatic
    fun fromExternalMetadataList(value: List<com.catalogizer.androidtv.data.models.ExternalMetadata>?): String? {
        return value?.let { Json.encodeToString(it) }
    }

    @TypeConverter
    @JvmStatic
    fun toExternalMetadataList(value: String?): List<com.catalogizer.androidtv.data.models.ExternalMetadata>? {
        return value?.let { Json.decodeFromString<List<com.catalogizer.androidtv.data.models.ExternalMetadata>>(it) }
    }

    @TypeConverter
    @JvmStatic
    fun fromMediaVersionList(value: List<com.catalogizer.androidtv.data.models.MediaVersion>?): String? {
        return value?.let { Json.encodeToString(it) }
    }

    @TypeConverter
    @JvmStatic
    fun toMediaVersionList(value: String?): List<com.catalogizer.androidtv.data.models.MediaVersion>? {
        return value?.let { Json.decodeFromString<List<com.catalogizer.androidtv.data.models.MediaVersion>>(it) }
    }

    @TypeConverter
    @JvmStatic
    fun fromStringMap(value: Map<String, String>?): String? {
        return value?.let { Json.encodeToString(it) }
    }

    @TypeConverter
    @JvmStatic
    fun toStringMap(value: String?): Map<String, String>? {
        return value?.let { Json.decodeFromString<Map<String, String>>(it) }
    }
}