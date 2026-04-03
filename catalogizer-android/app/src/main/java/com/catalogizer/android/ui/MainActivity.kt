package com.catalogizer.android.ui

import android.os.Bundle
import android.widget.Toast
import androidx.activity.ComponentActivity
import androidx.activity.addCallback
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
import com.catalogizer.android.ui.splash.SplashContent
import com.catalogizer.android.ui.theme.CatalogizerTheme
import com.catalogizer.android.ui.viewmodel.AuthViewModel
import com.catalogizer.android.ui.viewmodel.HomeViewModel
import com.catalogizer.android.ui.viewmodel.MainViewModel
import com.catalogizer.android.ui.viewmodel.SearchViewModel

/**
 * Main entry point activity for the Catalogizer Android app.
 * Initializes ViewModels from [DependencyContainer], configures edge-to-edge display,
 * installs the splash screen, and implements double-back-press to exit.
 */
class MainActivity : ComponentActivity() {

    private lateinit var authViewModel: AuthViewModel
    private lateinit var mainViewModel: MainViewModel
    private lateinit var homeViewModel: HomeViewModel
    private lateinit var searchViewModel: SearchViewModel
    private var backPressedOnce = false
    private val resetBackPressRunnable = Runnable { backPressedOnce = false }

    override fun onCreate(savedInstanceState: Bundle?) {
        // Install splash screen
        val splashScreen = installSplashScreen()

        super.onCreate(savedInstanceState)

        // Double-back-press to exit — prevents accidental exits
        onBackPressedDispatcher.addCallback(this) {
            if (backPressedOnce) {
                isEnabled = false
                onBackPressedDispatcher.onBackPressed()
                return@addCallback
            }
            backPressedOnce = true
            Toast.makeText(this@MainActivity, "Press back again to exit", Toast.LENGTH_SHORT).show()
            window.decorView.postDelayed(resetBackPressRunnable, 2000)
        }

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

    override fun onDestroy() {
        window.decorView.removeCallbacks(resetBackPressRunnable)
        super.onDestroy()
    }
}

/**
 * Root composable that manages the splash-to-navigation transition based on
 * authentication state from [AuthViewModel] and loading state from [MainViewModel].
 */
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

    var splashComplete by remember { mutableStateOf(false) }

    if (!splashComplete) {
        SplashContent(
            isAppReady = !isLoading,
            onSplashComplete = { splashComplete = true }
        )
    } else {
        CatalogizerNavigation(
            isAuthenticated = authState.isAuthenticated,
            authViewModel = authViewModel,
            homeViewModel = homeViewModel,
            searchViewModel = searchViewModel
        )
    }
}
