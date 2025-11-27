package com.catalogizer.androidtv.data.local

import androidx.room.TypeConverter
import kotlinx.serialization.encodeToString
import kotlinx.serialization.json.Json

object Converters {
    @TypeConverter
    @JvmStatic
    fun fromStringList(value: List<String>): String {
        return Json.encodeToString(value)
    }

    @TypeConverter
    @JvmStatic
    fun toStringList(value: String): List<String> {
        return Json.decodeFromString<List<String>>(value)
    }

    @TypeConverter
    @JvmStatic
    fun fromExternalMetadataList(value: List<com.catalogizer.androidtv.data.models.ExternalMetadata>): String {
        return Json.encodeToString(value)
    }

    @TypeConverter
    @JvmStatic
    fun toExternalMetadataList(value: String): List<com.catalogizer.androidtv.data.models.ExternalMetadata> {
        return Json.decodeFromString<List<com.catalogizer.androidtv.data.models.ExternalMetadata>>(value)
    }

    @TypeConverter
    @JvmStatic
    fun fromMediaVersionList(value: List<com.catalogizer.androidtv.data.models.MediaVersion>): String {
        return Json.encodeToString(value)
    }

    @TypeConverter
    @JvmStatic
    fun toMediaVersionList(value: String): List<com.catalogizer.androidtv.data.models.MediaVersion> {
        return Json.decodeFromString<List<com.catalogizer.androidtv.data.models.MediaVersion>>(value)
    }
}