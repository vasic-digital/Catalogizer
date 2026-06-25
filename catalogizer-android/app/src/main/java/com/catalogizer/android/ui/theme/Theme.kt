package com.catalogizer.android.ui.theme

import android.app.Activity
import android.os.Build
import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.graphics.toArgb
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.platform.LocalView
import androidx.core.view.WindowCompat

// Color values live in Color.kt (OpenDesign Catalogizer-Blue token palette,
// §11.4.162) and are mapped onto the Material 3 ColorScheme roles below.

private val LightColorScheme = lightColorScheme(
    primary = CatalogizerLightPrimary,
    onPrimary = CatalogizerLightOnPrimary,
    primaryContainer = CatalogizerLightPrimaryContainer,
    onPrimaryContainer = CatalogizerLightOnPrimaryContainer,
    secondary = CatalogizerLightSecondary,
    onSecondary = CatalogizerLightOnSecondary,
    secondaryContainer = CatalogizerLightSecondaryContainer,
    onSecondaryContainer = CatalogizerLightOnSecondaryContainer,
    tertiary = CatalogizerLightTertiary,
    onTertiary = CatalogizerLightOnTertiary,
    tertiaryContainer = CatalogizerLightTertiaryContainer,
    onTertiaryContainer = CatalogizerLightOnTertiaryContainer,
    error = CatalogizerLightError,
    errorContainer = CatalogizerLightErrorContainer,
    onError = CatalogizerLightOnError,
    onErrorContainer = CatalogizerLightOnErrorContainer,
    background = CatalogizerLightBackground,
    onBackground = CatalogizerLightOnBackground,
    surface = CatalogizerLightSurface,
    onSurface = CatalogizerLightOnSurface,
    surfaceVariant = CatalogizerLightSurfaceVariant,
    onSurfaceVariant = CatalogizerLightOnSurfaceVariant,
    outline = CatalogizerLightOutline
)

private val DarkColorScheme = darkColorScheme(
    primary = CatalogizerDarkPrimary,
    onPrimary = CatalogizerDarkOnPrimary,
    primaryContainer = CatalogizerDarkPrimaryContainer,
    onPrimaryContainer = CatalogizerDarkOnPrimaryContainer,
    secondary = CatalogizerDarkSecondary,
    onSecondary = CatalogizerDarkOnSecondary,
    secondaryContainer = CatalogizerDarkSecondaryContainer,
    onSecondaryContainer = CatalogizerDarkOnSecondaryContainer,
    tertiary = CatalogizerDarkTertiary,
    onTertiary = CatalogizerDarkOnTertiary,
    tertiaryContainer = CatalogizerDarkTertiaryContainer,
    onTertiaryContainer = CatalogizerDarkOnTertiaryContainer,
    error = CatalogizerDarkError,
    errorContainer = CatalogizerDarkErrorContainer,
    onError = CatalogizerDarkOnError,
    onErrorContainer = CatalogizerDarkOnErrorContainer,
    background = CatalogizerDarkBackground,
    onBackground = CatalogizerDarkOnBackground,
    surface = CatalogizerDarkSurface,
    onSurface = CatalogizerDarkOnSurface,
    surfaceVariant = CatalogizerDarkSurfaceVariant,
    onSurfaceVariant = CatalogizerDarkOnSurfaceVariant,
    outline = CatalogizerDarkOutline
)

/**
 * Catalogizer Material 3 theme composable that applies light/dark color schemes,
 * dynamic color on Android 12+, and configures the status bar appearance.
 */
@Composable
fun CatalogizerTheme(
    darkTheme: Boolean = isSystemInDarkTheme(),
    dynamicColor: Boolean = true,
    content: @Composable () -> Unit
) {
    val colorScheme = when {
        dynamicColor && Build.VERSION.SDK_INT >= Build.VERSION_CODES.S -> {
            val context = LocalContext.current
            if (darkTheme) dynamicDarkColorScheme(context) else dynamicLightColorScheme(context)
        }
        darkTheme -> DarkColorScheme
        else -> LightColorScheme
    }

    val view = LocalView.current
    if (!view.isInEditMode) {
        SideEffect {
            val window = (view.context as Activity).window
            window.statusBarColor = colorScheme.primary.toArgb()
            WindowCompat.getInsetsController(window, view).isAppearanceLightStatusBars = darkTheme
        }
    }

    MaterialTheme(
        colorScheme = colorScheme,
        typography = androidx.compose.material3.Typography(),
        content = content
    )
}