package com.catalogizer.androidtv.ui.navigation

import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.navigation.NavHostController
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import com.catalogizer.androidtv.ui.screens.home.HomeScreen
import com.catalogizer.androidtv.ui.screens.login.LoginScreen
import com.catalogizer.androidtv.ui.screens.media.MediaDetailScreen
import com.catalogizer.androidtv.ui.screens.player.MediaPlayerScreen
import com.catalogizer.androidtv.ui.screens.search.SearchScreen
import com.catalogizer.androidtv.ui.screens.settings.SettingsScreen
import com.catalogizer.androidtv.ui.screens.search.SearchViewModel
import com.catalogizer.androidtv.ui.viewmodel.AuthViewModel
import com.catalogizer.androidtv.ui.viewmodel.HomeViewModel

/**
 * Sealed class defining the navigation route constants for all TV app screens.
 */
sealed class TVScreen(val route: String) {
    object Login : TVScreen("login")
    object Home : TVScreen("home")
    object Search : TVScreen("search")
    object MediaDetail : TVScreen("media_detail/{mediaId}") {
        fun createRoute(mediaId: Long) = "media_detail/$mediaId"
    }
    object Player : TVScreen("player/{mediaId}") {
        fun createRoute(mediaId: Long) = "player/$mediaId"
    }
    object Settings : TVScreen("settings")
}

/**
 * Top-level navigation host for the TV app wiring screen routes to their composables.
 * Selects the start destination based on [isAuthenticated] state and provides
 * navigation to login, home, search, media detail, player, and settings screens.
 */
@Composable
fun TVNavigation(
    isAuthenticated: Boolean,
    authViewModel: AuthViewModel,
    homeViewModel: HomeViewModel,
    searchViewModel: SearchViewModel,
    deepLinkMediaId: Long = -1L,
    deepLinkAction: String? = null,
    navController: NavHostController = rememberNavController()
) {
    val startDestination = if (isAuthenticated) TVScreen.Home.route else TVScreen.Login.route

    NavHost(
        navController = navController,
        startDestination = startDestination
    ) {
        composable(TVScreen.Login.route) {
            LoginScreen(
                authViewModel = authViewModel,
                onLoginSuccess = {
                    navController.navigate(TVScreen.Home.route) {
                        popUpTo(TVScreen.Login.route) { inclusive = true }
                    }
                }
            )
        }

        composable(TVScreen.Home.route) {
            HomeScreen(
                onNavigateToSearch = {
                    navController.navigate(TVScreen.Search.route)
                },
                onNavigateToSettings = {
                    navController.navigate(TVScreen.Settings.route)
                },
                onNavigateToMediaDetail = { mediaId ->
                    navController.navigate(TVScreen.MediaDetail.createRoute(mediaId))
                },
                onNavigateToPlayer = { mediaId ->
                    navController.navigate(TVScreen.Player.createRoute(mediaId))
                },
                viewModel = homeViewModel
            )
        }

        composable(TVScreen.Search.route) {
            SearchScreen(
                viewModel = searchViewModel,
                onNavigateBack = {
                    navController.popBackStack()
                },
                onNavigateToMediaDetail = { mediaId ->
                    navController.navigate(TVScreen.MediaDetail.createRoute(mediaId))
                }
            )
        }

        composable(TVScreen.MediaDetail.route) { backStackEntry ->
            val mediaId = backStackEntry.arguments?.getString("mediaId")?.toLongOrNull() ?: 0L
            MediaDetailScreen(
                mediaId = mediaId,
                onNavigateBack = {
                    navController.popBackStack()
                },
                onNavigateToPlayer = { id ->
                    navController.navigate(TVScreen.Player.createRoute(id))
                }
            )
        }

        composable(TVScreen.Player.route) { backStackEntry ->
            val mediaId = backStackEntry.arguments?.getString("mediaId")?.toLongOrNull() ?: 0L
            MediaPlayerScreen(
                mediaId = mediaId,
                onNavigateBack = {
                    navController.popBackStack()
                }
            )
        }

        composable(TVScreen.Settings.route) {
            val container = com.catalogizer.androidtv.DependencyContainer.getInstance(
                androidx.compose.ui.platform.LocalContext.current
            )
            SettingsScreen(
                settingsViewModel = container.createSettingsViewModel(),
                onNavigateBack = {
                    navController.popBackStack()
                },
                onLogout = {
                    navController.navigate(TVScreen.Login.route) {
                        popUpTo(0) { inclusive = true }
                    }
                }
            )
        }
    }

    // Handle deep link navigation from home screen channels
    LaunchedEffect(deepLinkMediaId) {
        if (deepLinkMediaId > 0 && isAuthenticated) {
            when (deepLinkAction) {
                "play" -> navController.navigate(TVScreen.Player.createRoute(deepLinkMediaId))
                else -> navController.navigate(TVScreen.MediaDetail.createRoute(deepLinkMediaId))
            }
        }
    }
}