package com.catalogizer.android.ui

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.core.splashscreen.SplashScreen.Companion.installSplashScreen
import androidx.core.view.WindowCompat
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.catalogizer.android.CatalogizerApplication
import com.catalogizer.android.ui.navigation.CatalogizerNavigation
import com.catalogizer.android.ui.theme.CatalogizerTheme
import com.catalogizer.android.ui.viewmodel.AuthViewModel
import com.catalogizer.android.ui.viewmodel.HomeViewModel
import com.catalogizer.android.ui.viewmodel.MainViewModel
import com.catalogizer.android.ui.viewmodel.SearchViewModel

class MainActivity : ComponentActivity() {

    private lateinit var authViewModel: AuthViewModel
    private lateinit var mainViewModel: MainViewModel
    private lateinit var homeViewModel: HomeViewModel
    private lateinit var searchViewModel: SearchViewModel

    override fun onCreate(savedInstanceState: Bundle?) {
        // Install splash screen
        val splashScreen = installSplashScreen()

        super.onCreate(savedInstanceState)

        // Initialize ViewModels
        val dependencyContainer = (application as CatalogizerApplication).dependencyContainer
        authViewModel = dependencyContainer.createAuthViewModel()
        mainViewModel = dependencyContainer.createMainViewModel()
        homeViewModel = dependencyContainer.createHomeViewModel()
        searchViewModel = dependencyContainer.createSearchViewModel()

        // Configure edge-to-edge display
        WindowCompat.setDecorFitsSystemWindows(window, false)

        // Keep splash screen until app is ready
        splashScreen.setKeepOnScreenCondition {
            mainViewModel.isLoading.value
        }

        setContent {
            CatalogizerTheme {
                Surface(
                    modifier = Modifier.fillMaxSize(),
                    color = MaterialTheme.colorScheme.background
                ) {
                    CatalogizerApp(
                        authViewModel = authViewModel,
                        mainViewModel = mainViewModel,
                        homeViewModel = homeViewModel,
                        searchViewModel = searchViewModel
                    )
                }
            }
        }
    }
}

@Composable
fun CatalogizerApp(
    authViewModel: AuthViewModel,
    mainViewModel: MainViewModel,
    homeViewModel: HomeViewModel,
    searchViewModel: SearchViewModel
) {
    val authState by authViewModel.authState.collectAsStateWithLifecycle()
    val isLoading by mainViewModel.isLoading.collectAsStateWithLifecycle()

    LaunchedEffect(Unit) {
        mainViewModel.initializeApp()
    }

    if (!isLoading) {
        CatalogizerNavigation(
            isAuthenticated = authState.isAuthenticated,
            authViewModel = authViewModel,
            homeViewModel = homeViewModel,
            searchViewModel = searchViewModel
        )
    }
}
