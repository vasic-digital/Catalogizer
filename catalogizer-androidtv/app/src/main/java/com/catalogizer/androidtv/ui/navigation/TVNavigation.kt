package com.catalogizer.androidtv.ui.navigation

import androidx.compose.runtime.Composable
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
import com.catalogizer.androidtv.ui.viewmodel.AuthViewModel
import com.catalogizer.androidtv.ui.viewmodel.HomeViewModel

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

@Composable
fun TVNavigation(
    isAuthenticated: Boolean,
    authViewModel: AuthViewModel,
    homeViewModel: HomeViewModel,
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
            SettingsScreen(
                authViewModel = authViewModel,
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
}