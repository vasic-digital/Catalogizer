package com.catalogizer.androidtv.data.tv

/**
 * Controls what happens when a user clicks a program card on the Android TV home screen.
 * Configurable per media type in Settings.
 */
enum class LaunchAction {
    /** Open the MediaDetailScreen where the user can choose to play. */
    DETAIL,
    /** Start playback immediately. */
    IMMEDIATE_PLAY;

    companion object {
        fun fromString(value: String): LaunchAction {
            return values().find { it.name == value } ?: DETAIL
        }
    }
}
