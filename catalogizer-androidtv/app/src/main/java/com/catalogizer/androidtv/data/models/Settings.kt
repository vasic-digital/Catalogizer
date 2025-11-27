package com.catalogizer.androidtv.data.models

data class Settings(
    val enableNotifications: Boolean,
    val enableAutoPlay: Boolean,
    val streamingQuality: String,
    val enableSubtitles: Boolean,
    val subtitleLanguage: String
)