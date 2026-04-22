package com.catalogizer.android.testutils

import androidx.compose.ui.test.junit4.createComposeRule
import androidx.compose.ui.test.onRoot
import androidx.compose.ui.test.printToLog
import androidx.test.ext.junit.runners.AndroidJUnit4
import org.junit.Rule
import org.junit.runner.RunWith

/**
 * Base class for Compose UI tests.
 */
@RunWith(AndroidJUnit4::class)
abstract class ComposeTestBase {
    
    @get:Rule
    val composeTestRule = createComposeRule()
    
    // Common test utilities for Compose UI tests
    protected fun waitForIdle() {
        composeTestRule.waitForIdle()
    }
    
    protected fun printComposeTree() {
        composeTestRule.onRoot().printToLog("COMPOSE_TREE")
    }
}
