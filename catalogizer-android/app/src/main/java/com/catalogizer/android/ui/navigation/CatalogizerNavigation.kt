package com.catalogizer.android.ui.navigation

import androidx.compose.runtime.Composable
import androidx.navigation.NavHostController
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import com.catalogizer.android.ui.screens.home.HomeScreen
import com.catalogizer.android.ui.screens.login.LoginScreen
import com.catalogizer.android.ui.screens.search.SearchScreen
import com.catalogizer.android.ui.screens.settings.SettingsScreen
import com.catalogizer.android.ui.viewmodel.AuthViewModel
import com.catalogizer.android.ui.viewmodel.HomeViewModel
import com.catalogizer.android.ui.viewmodel.SearchViewModel

sealed class Screen(val route: String) {
    object Login : Screen("login")
    object Home : Screen("home")
    object Search : Screen("search")
    object Settings : Screen("settings")
}

@Composable
fun CatalogizerNavigation(
    isAuthenticated: Boolean,
    authViewModel: AuthViewModel,
    homeViewModel: HomeViewModel,
    searchViewModel: SearchViewModel,
    navController: NavHostController = rememberNavController()
) {
    val startDestination = if (isAuthenticated) Screen.Home.route else Screen.Login.route

    NavHost(
        navController = navController,
        startDestination = startDestination
    ) {
        composable(Screen.Login.route) {
            LoginScreen(
                authViewModel = authViewModel,
                onLoginSuccess = {
                    navController.navigate(Screen.Home.route) {
                        popUpTo(Screen.Login.route) { inclusive = true }
                    }
                }
            )
        }

        composable(Screen.Home.route) {
            HomeScreen(
                viewModel = homeViewModel,
                onNavigateToSearch = {
                    navController.navigate(Screen.Search.route)
                },
                onNavigateToSettings = {
                    navController.navigate(Screen.Settings.route)
                },
                onNavigateToMediaDetail = { _ ->
                    // Media detail screen will be added in a future update
                }
            )
        }

        composable(Screen.Search.route) {
            SearchScreen(
                viewModel = searchViewModel,
                onNavigateBack = {
                    navController.popBackStack()
                },
                onNavigateToMediaDetail = { _ ->
                    // Media detail screen will be added in a future update
                }
            )
        }

        composable(Screen.Settings.route) {
            SettingsScreen(
                onNavigateBack = {
                    navController.popBackStack()
                },
                onLogout = {
                    authViewModel.logout()
                    navController.navigate(Screen.Login.route) {
                        popUpTo(0) { inclusive = true }
                    }
                }
            )
        }
    }
}
