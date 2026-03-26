package com.catalogizer.androidtv.ui

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.tv.material3.ExperimentalTvMaterial3Api
import androidx.tv.material3.Surface
import com.catalogizer.androidtv.CatalogizerTVApplication
import com.catalogizer.androidtv.ui.navigation.TVNavigation
import com.catalogizer.androidtv.ui.splash.SplashContent
import com.catalogizer.androidtv.ui.theme.CatalogizerTVTheme
import com.catalogizer.androidtv.ui.screens.search.SearchViewModel
import com.catalogizer.androidtv.ui.viewmodel.AuthViewModel
import com.catalogizer.androidtv.ui.viewmodel.HomeViewModel
import com.catalogizer.androidtv.ui.viewmodel.MainViewModel
import kotlinx.coroutines.launch

class MainActivity : ComponentActivity() {

    private lateinit var authViewModel: AuthViewModel
    private lateinit var mainViewModel: MainViewModel
    private lateinit var homeViewModel: HomeViewModel
    private lateinit var searchViewModel: SearchViewModel

    @OptIn(ExperimentalTvMaterial3Api::class)
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        // Handle server URL from intent extra (for ADB testing)
        val dependencyContainer = (application as CatalogizerTVApplication).dependencyContainer
        intent.getStringExtra("server_url")?.let { url ->
            if (url.isNotBlank()) {
                dependencyContainer.switchServer(url)
                kotlinx.coroutines.CoroutineScope(kotlinx.coroutines.Dispatchers.IO).launch {
                    dependencyContainer.settingsRepository.updateServerUrl(url)
                }
            }
        }

        // Initialize ViewModels
        authViewModel = dependencyContainer.createAuthViewModel()
        mainViewModel = dependencyContainer.createMainViewModel()
        homeViewModel = dependencyContainer.createHomeViewModel()
        searchViewModel = dependencyContainer.createSearchViewModel()

        setContent {
            CatalogizerTVTheme {
                Surface(
                    modifier = Modifier.fillMaxSize()
                ) {
                    CatalogizerTVApp(
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
fun CatalogizerTVApp(
    authViewModel: AuthViewModel,
    mainViewModel: MainViewModel,
    homeViewModel: HomeViewModel,
    searchViewModel: SearchViewModel
) {
    val authState by authViewModel.authState.collectAsStateWithLifecycle()
    val isLoading by mainViewModel.isLoading.collectAsStateWithLifecycle()
    val context = LocalContext.current

    LaunchedEffect(Unit) {
        mainViewModel.initializeApp()
    }

    var splashComplete by remember { mutableStateOf(false) }

    if (!splashComplete) {
        SplashContent(
            isAppReady = !isLoading,
            onSplashComplete = { splashComplete = true }
        )
    } else {
        TVNavigation(
            isAuthenticated = authState.isAuthenticated,
            authViewModel = authViewModel,
            homeViewModel = homeViewModel,
            searchViewModel = searchViewModel
        )
    }
}