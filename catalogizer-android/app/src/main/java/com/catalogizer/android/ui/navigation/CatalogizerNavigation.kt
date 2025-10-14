package com.catalogizer.android.ui.navigation

import androidx.compose.runtime.Composable
import androidx.compose.material3.Text
import com.catalogizer.android.ui.viewmodel.AuthViewModel

@Composable
fun CatalogizerNavigation(
    isAuthenticated: Boolean,
    authViewModel: AuthViewModel
) {
    if (isAuthenticated) {
        Text("Authenticated - Main Screen")
    } else {
        Text("Login Screen")
    }
}