package com.catalogizer.androidtv.ui

import android.os.Bundle
import android.widget.Toast
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
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

/**
 * Main entry point activity for the Catalogizer Android TV app.
 * Initializes ViewModels from [DependencyContainer], handles server URL from
 * intent extras (for ADB testing), and implements double-back-press to exit.
 */
class MainActivity : ComponentActivity() {

    private lateinit var authViewModel: AuthViewModel
    private lateinit var mainViewModel: MainViewModel
    private lateinit var homeViewModel: HomeViewModel
    private lateinit var searchViewModel: SearchViewModel
    private var backPressedOnce = false

    // Prevent accidental exit on Android TV — intercept finish() call
    // Back presses flow normally through Compose navigation; only when navigation
    // exhausts its stack does the system call finish(). We guard that final exit.
    override fun finish() {
        if (backPressedOnce) {
            super.finish()
            return
        }
        backPressedOnce = true
        Toast.makeText(this, "Press back again to exit", Toast.LENGTH_SHORT).show()
        window.decorView.postDelayed({ backPressedOnce = false }, 2000)
    }

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

        // Extract deep link from ChannelDeepLinkActivity
        val deepLinkMediaId = intent?.getLongExtra("deep_link_media_id", -1L) ?: -1L
        val deepLinkAction = intent?.getStringExtra("deep_link_action")

        setContent {
            CatalogizerTVTheme {
                Surface(
                    modifier = Modifier.fillMaxSize()
                ) {
                    CatalogizerTVApp(
                        authViewModel = authViewModel,
                        mainViewModel = mainViewModel,
                        homeViewModel = homeViewModel,
                        searchViewModel = searchViewModel,
                        deepLinkMediaId = deepLinkMediaId,
                        deepLinkAction = deepLinkAction
                    )
                }
            }
        }
    }
}

/**
 * Root composable for the TV app managing the splash-to-navigation transition
 * based on authentication state and app initialization progress.
 */
@Composable
fun CatalogizerTVApp(
    authViewModel: AuthViewModel,
    mainViewModel: MainViewModel,
    homeViewModel: HomeViewModel,
    searchViewModel: SearchViewModel,
    deepLinkMediaId: Long = -1L,
    deepLinkAction: String? = null
) {
    val authState by authViewModel.authState.collectAsStateWithLifecycle()
    val isLoading by mainViewModel.isLoading.collectAsStateWithLifecycle()

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
            searchViewModel = searchViewModel,
            deepLinkMediaId = deepLinkMediaId,
            deepLinkAction = deepLinkAction
        )
    }
}