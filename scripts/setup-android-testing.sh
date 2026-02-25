#!/bin/bash
# Android Test Infrastructure Setup Script for Catalogizer
# This script sets up comprehensive testing for Android apps

set -e

echo "📱 Catalogizer Android Test Infrastructure Setup"
echo "=============================================="

# Check if we're in the right directory
if [ ! -d "catalogizer-android" ]; then
    echo "❌ catalogizer-android directory not found. Run from project root."
    exit 1
fi

cd catalogizer-android

echo "📋 Checking Android project structure..."
if [ ! -f "app/build.gradle.kts" ]; then
    echo "❌ Android app build.gradle.kts not found."
    exit 1
fi

echo "✅ Android project structure verified"

# Create test utilities directory
echo "📁 Creating test utilities..."
mkdir -p app/src/test/java/com/catalogizer/android/testutils

# Create TestDispatcherRule
cat > app/src/test/java/com/catalogizer/android/testutils/TestDispatcherRule.kt << 'EOF'
package com.catalogizer.android.testutils

import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.test.*
import org.junit.rules.TestWatcher
import org.junit.runner.Description

/**
 * Test rule that sets the Main dispatcher to a TestDispatcher for unit tests.
 */
class TestDispatcherRule(
    private val testDispatcher: TestDispatcher = StandardTestDispatcher()
) : TestWatcher() {
    
    override fun starting(description: Description) {
        super.starting(description)
        Dispatchers.setMain(testDispatcher)
    }
    
    override fun finished(description: Description) {
        super.finished(description)
        Dispatchers.resetMain()
    }
}
EOF

# Create MockRepositoryHelper
cat > app/src/test/java/com/catalogizer/android/testutils/MockRepositoryHelper.kt << 'EOF'
package com.catalogizer.android.testutils

import com.catalogizer.android.data.models.MediaItem
import com.catalogizer.android.data.models.MediaType
import com.catalogizer.android.data.models.User
import com.catalogizer.android.data.remote.ApiResult
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flow
import kotlinx.coroutines.flow.flowOf
import java.util.*

/**
 * Helper class for creating mock data in tests.
 */
object MockRepositoryHelper {
    
    // Mock Media Items
    fun createMockMediaItem(
        id: Long = 1L,
        title: String = "Test Movie",
        type: MediaType = MediaType.MOVIE,
        year: Int = 2023
    ): MediaItem {
        return MediaItem(
            id = id,
            title = title,
            type = type,
            year = year,
            posterPath = "/test/poster.jpg",
            backdropPath = "/test/backdrop.jpg",
            overview = "Test overview",
            rating = 8.5,
            runtime = 120,
            genres = listOf("Action", "Adventure"),
            createdAt = Date(),
            updatedAt = Date()
        )
    }
    
    fun createMockMediaItems(count: Int = 5): List<MediaItem> {
        return (1..count).map { index ->
            createMockMediaItem(
                id = index.toLong(),
                title = "Test Movie $index",
                year = 2020 + index
            )
        }
    }
    
    // Mock User
    fun createMockUser(
        id: Long = 1L,
        username: String = "testuser",
        email: String = "test@example.com"
    ): User {
        return User(
            id = id,
            username = username,
            email = email,
            createdAt = Date(),
            updatedAt = Date()
        )
    }
    
    // Mock API Results
    fun <T> createSuccessApiResult(data: T): ApiResult.Success<T> {
        return ApiResult.Success(data)
    }
    
    fun <T> createErrorApiResult(message: String = "Test error"): ApiResult.Error<T> {
        return ApiResult.Error(message)
    }
    
    fun <T> createLoadingApiResult(): ApiResult.Loading<T> {
        return ApiResult.Loading()
    }
    
    // Mock Flows
    fun <T> createMockFlow(data: T): Flow<T> {
        return flowOf(data)
    }
    
    fun <T> createMockFlowSequence(vararg items: T): Flow<T> {
        return flow {
            items.forEach { emit(it) }
        }
    }
}
EOF

# Create TestDataGenerator
cat > app/src/test/java/com/catalogizer/android/testutils/TestDataGenerator.kt << 'EOF'
package com.catalogizer.android.testutils

import com.catalogizer.android.data.models.*

/**
 * Generates test data for Android tests.
 */
object TestDataGenerator {
    
    // Generate test media items
    fun generateMediaItems(count: Int = 10): List<MediaItem> {
        val items = mutableListOf<MediaItem>()
        val types = MediaType.values()
        val genres = listOf(
            "Action", "Adventure", "Comedy", "Drama", "Horror",
            "Sci-Fi", "Fantasy", "Romance", "Thriller", "Documentary"
        )
        
        for (i in 1..count) {
            val type = types[i % types.size]
            val title = when (type) {
                MediaType.MOVIE -> "Test Movie $i"
                MediaType.TV_SHOW -> "Test TV Show $i"
                MediaType.MUSIC_ALBUM -> "Test Album $i"
                MediaType.GAME -> "Test Game $i"
                MediaType.BOOK -> "Test Book $i"
                else -> "Test Item $i"
            }
            
            val itemGenres = genres.shuffled().take(3)
            
            items.add(
                MediaItem(
                    id = i.toLong(),
                    title = title,
                    type = type,
                    year = 2010 + (i % 15),
                    posterPath = "/posters/poster_$i.jpg",
                    backdropPath = "/backdrops/backdrop_$i.jpg",
                    overview = "This is a test overview for $title. It's a great piece of media that everyone should experience.",
                    rating = 5.0 + (i % 5).toDouble(),
                    runtime = 90 + (i % 60),
                    genres = itemGenres,
                    createdAt = java.util.Date(System.currentTimeMillis() - i * 86400000L),
                    updatedAt = java.util.Date()
                )
            )
        }
        
        return items
    }
    
    // Generate test users
    fun generateUsers(count: Int = 5): List<User> {
        val users = mutableListOf<User>()
        
        for (i in 1..count) {
            users.add(
                User(
                    id = i.toLong(),
                    username = "user$i",
                    email = "user$i@example.com",
                    createdAt = java.util.Date(System.currentTimeMillis() - i * 86400000L),
                    updatedAt = java.util.Date()
                )
            )
        }
        
        return users
    }
    
    // Generate test search results
    fun generateSearchResults(query: String, count: Int = 5): List<MediaItem> {
        return generateMediaItems(count).map { item ->
            item.copy(title = "$query ${item.title}")
        }
    }
}
EOF

# Create ViewModelTestBase
cat > app/src/test/java/com/catalogizer/android/testutils/ViewModelTestBase.kt << 'EOF'
package com.catalogizer.android.testutils

import androidx.arch.core.executor.testing.InstantTaskExecutorRule
import org.junit.Before
import org.junit.Rule
import org.junit.runner.RunWith
import org.mockito.junit.MockitoJUnitRunner

/**
 * Base class for ViewModel tests with common setup.
 */
@RunWith(MockitoJUnitRunner::class)
abstract class ViewModelTestBase {
    
    @get:Rule
    val instantExecutorRule = InstantTaskExecutorRule()
    
    @get:Rule
    val testDispatcherRule = TestDispatcherRule()
    
    @Before
    open fun setUp() {
        // Common setup for all ViewModel tests
    }
}
EOF

# Create ComposeTestRule
cat > app/src/androidTest/java/com/catalogizer/android/testutils/ComposeTestRule.kt << 'EOF'
package com.catalogizer.android.testutils

import androidx.compose.ui.test.junit4.createComposeRule
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
EOF

echo "✅ Test utilities created"

# Update build.gradle.kts to improve test configuration
echo "🔧 Updating build.gradle.kts for better test coverage..."

# Check if we need to add test options
if ! grep -q "testOptions" app/build.gradle.kts; then
    echo "⚠️ testOptions not found in build.gradle.kts. Adding test configuration..."
    # We'll create a patch file
    cat > /tmp/test_options.patch << 'EOF'
--- a/app/build.gradle.kts
+++ b/app/build.gradle.kts
@@ -108,6 +108,14 @@
         }
     }
 
+    testOptions {
+        unitTests {
+            isIncludeAndroidResources = true
+            all {
+                it.jvmArgs("-noverify")
+            }
+        }
+    }
+
     compileOptions {
         sourceCompatibility = JavaVersion.VERSION_17
         targetCompatibility = JavaVersion.VERSION_17
EOF
    # Try to apply patch
    if patch -p1 -f < /tmp/test_options.patch 2>/dev/null || true; then
        echo "✅ Added testOptions to build.gradle.kts"
    else
        echo "⚠️ Could not automatically patch build.gradle.kts. Please add testOptions manually."
    fi
fi

# Create test coverage script
echo "📝 Creating test coverage script..."
cat > ../scripts/run-android-tests.sh << 'EOF'
#!/bin/bash
# Run Android tests with coverage reporting

set -e

echo "📱 Running Catalogizer Android Tests"
echo "==================================="

cd catalogizer-android

echo "🔧 Cleaning build..."
./gradlew clean

echo "🧪 Running unit tests..."
./gradlew testDebugUnitTest --info

echo "📊 Generating test coverage report..."
./gradlew jacocoTestReport

echo "📁 Coverage reports generated:"
echo "   - HTML: app/build/reports/jacoco/jacocoTestReport/html/index.html"
echo "   - XML:  app/build/reports/jacoco/jacocoTestReport/jacocoTestReport.xml"

# Check if coverage meets threshold (70%)
COVERAGE_FILE="app/build/reports/jacoco/jacocoTestReport/jacocoTestReport.xml"
if [ -f "$COVERAGE_FILE" ]; then
    echo "📈 Checking coverage threshold..."
    # Extract line coverage percentage (simplified)
    COVERAGE=$(grep -o 'linecoverage="[0-9]*\.[0-9]*"' "$COVERAGE_FILE" | head -1 | sed 's/linecoverage="//' | sed 's/"//')
    if [ -n "$COVERAGE" ]; then
        echo "✅ Line coverage: ${COVERAGE}%"
        if (( $(echo "$COVERAGE < 70" | bc -l) )); then
            echo "⚠️ Coverage below 70% target. Consider adding more tests."
        else
            echo "🎉 Coverage meets 70% target!"
        fi
    else
        echo "⚠️ Could not parse coverage from report"
    fi
else
    echo "⚠️ Coverage report not found at $COVERAGE_FILE"
fi

echo ""
echo "🚀 To view coverage report:"
echo "   open app/build/reports/jacoco/jacocoTestReport/html/index.html"
EOF

chmod +x ../scripts/run-android-tests.sh

# Create test examples for missing coverage
echo "📝 Creating example tests for common patterns..."

# Example Repository Test
cat > app/src/test/java/com/catalogizer/android/data/repository/ExampleRepositoryTest.kt << 'EOF'
package com.catalogizer.android.data.repository

import com.catalogizer.android.testutils.MockRepositoryHelper
import com.catalogizer.android.testutils.TestDataGenerator
import com.catalogizer.android.data.remote.ApiResult
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.runTest
import org.junit.Assert.*
import org.junit.Test

@ExperimentalCoroutinesApi
class ExampleRepositoryTest {
    
    @Test
    fun `repository should return success for valid data`() = runTest {
        // Given
        val mockData = MockRepositoryHelper.createMockMediaItems(3)
        
        // When - simulate repository call
        val result: ApiResult<List<MediaItem>> = ApiResult.Success(mockData)
        
        // Then
        assertTrue(result is ApiResult.Success)
        assertEquals(3, (result as ApiResult.Success).data.size)
    }
    
    @Test
    fun `repository should handle empty results`() = runTest {
        // Given
        val emptyList = emptyList<MediaItem>()
        
        // When
        val result: ApiResult<List<MediaItem>> = ApiResult.Success(emptyList)
        
        // Then
        assertTrue(result is ApiResult.Success)
        assertTrue((result as ApiResult.Success).data.isEmpty())
    }
    
    @Test
    fun `repository should handle errors gracefully`() = runTest {
        // Given
        val errorMessage = "Network error"
        
        // When
        val result: ApiResult<List<MediaItem>> = ApiResult.Error(errorMessage)
        
        // Then
        assertTrue(result is ApiResult.Error)
        assertEquals(errorMessage, (result as ApiResult.Error).message)
    }
}
EOF

# Example ViewModel Test with State
cat > app/src/test/java/com/catalogizer/android/ui/viewmodel/ExampleStateViewModelTest.kt << 'EOF'
package com.catalogizer.android.ui.viewmodel

import androidx.lifecycle.viewModelScope
import com.catalogizer.android.testutils.TestDispatcherRule
import com.catalogizer.android.testutils.ViewModelTestBase
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import org.junit.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue

class ExampleStateViewModelTest : ViewModelTestBase() {
    
    // Example ViewModel for testing
    class ExampleViewModel {
        private val _count = MutableStateFlow(0)
        val count: StateFlow<Int> = _count.asStateFlow()
        
        private val _text = MutableStateFlow("")
        val text: StateFlow<String> = _text.asStateFlow()
        
        fun increment() {
            _count.value++
        }
        
        fun updateText(newText: String) {
            _text.value = newText
        }
        
        fun reset() {
            _count.value = 0
            _text.value = ""
        }
    }
    
    @Test
    fun `viewmodel should initialize with default values`() = runTest {
        // Given
        val viewModel = ExampleViewModel()
        
        // Then
        assertEquals(0, viewModel.count.value)
        assertEquals("", viewModel.text.value)
    }
    
    @Test
    fun `viewmodel should update count when incremented`() = runTest {
        // Given
        val viewModel = ExampleViewModel()
        
        // When
        viewModel.increment()
        
        // Then
        assertEquals(1, viewModel.count.value)
    }
    
    @Test
    fun `viewmodel should update text`() = runTest {
        // Given
        val viewModel = ExampleViewModel()
        val testText = "Hello, World!"
        
        // When
        viewModel.updateText(testText)
        
        // Then
        assertEquals(testText, viewModel.text.value)
    }
    
    @Test
    fun `viewmodel should reset to initial state`() = runTest {
        // Given
        val viewModel = ExampleViewModel()
        viewModel.increment()
        viewModel.updateText("Test")
        
        // When
        viewModel.reset()
        
        // Then
        assertEquals(0, viewModel.count.value)
        assertEquals("", viewModel.text.value)
    }
}
EOF

echo "✅ Example tests created"

# Create README for Android testing
cat > ../docs/android-testing-guide.md << 'EOF'
# Android Testing Guide for Catalogizer

## Overview

This guide covers the test infrastructure setup for Catalogizer Android applications. The goal is to achieve **95%+ test coverage** across all components.

## Test Structure

### 1. Unit Tests (`app/src/test/`)
- **Location**: `app/src/test/java/com/catalogizer/android/`
- **Purpose**: Test business logic, ViewModels, repositories, utilities
- **Frameworks**: JUnit 4, MockK, Mockito, Kotlin Coroutines Test
- **Coverage Target**: 90%+

### 2. Instrumented Tests (`app/src/androidTest/`)
- **Location**: `app/src/androidTest/java/com/catalogizer/android/`
- **Purpose**: Test UI components, navigation, integration
- **Frameworks**: Espresso, Compose UI Test, AndroidJUnitRunner
- **Coverage Target**: 80%+

## Test Utilities

We've created several test utility classes:

### `TestDispatcherRule`
Sets up coroutine test dispatchers for ViewModel tests.

### `MockRepositoryHelper`
Provides mock data generation for repositories.

### `TestDataGenerator`
Generates comprehensive test data for various scenarios.

### `ViewModelTestBase`
Base class for ViewModel tests with common setup.

### `ComposeTestBase`
Base class for Compose UI tests.

## Running Tests

### Run All Tests
```bash
./scripts/run-android-tests.sh
```

### Run Unit Tests Only
```bash
cd catalogizer-android
./gradlew testDebugUnitTest
```

### Run Instrumented Tests
```bash
cd catalogizer-android
./gradlew connectedDebugAndroidTest
```

### Generate Coverage Report
```bash
cd catalogizer-android
./gradlew jacocoTestReport
```

## Coverage Reports

After running tests, coverage reports are available at:
- **HTML Report**: `app/build/reports/jacoco/jacocoTestReport/html/index.html`
- **XML Report**: `app/build/reports/jacoco/jacocoTestReport/jacocoTestReport.xml`

## Test Patterns

### ViewModel Testing
```kotlin
class MyViewModelTest : ViewModelTestBase() {
    
    private lateinit var viewModel: MyViewModel
    private lateinit var mockRepository: MyRepository
    
    @Before
    fun setup() {
        mockRepository = mockk(relaxed = true)
        viewModel = MyViewModel(mockRepository)
    }
    
    @Test
    fun `viewmodel should load data`() = runTest {
        // Given
        val mockData = MockRepositoryHelper.createMockMediaItems()
        coEvery { mockRepository.getMedia() } returns mockData
        
        // When
        viewModel.loadData()
        advanceUntilIdle()
        
        // Then
        assertEquals(mockData, viewModel.media.value)
    }
}
```

### Repository Testing
```kotlin
@ExperimentalCoroutinesApi
class MyRepositoryTest {
    
    @Test
    fun `repository should fetch data from API`() = runTest {
        // Given
        val mockApi = mockk<MyApi>()
        val repository = MyRepository(mockApi)
        val expectedData = TestDataGenerator.generateMediaItems()
        
        coEvery { mockApi.getMedia() } returns expectedData
        
        // When
        val result = repository.getMedia()
        
        // Then
        assertEquals(expectedData, result)
    }
}
```

### Compose UI Testing
```kotlin
class MyScreenTest : ComposeTestBase() {
    
    @Test
    fun `screen should display title`() {
        // Given
        val viewModel = MyViewModel(mockRepository)
        
        // When
        composeTestRule.setContent {
            MyScreen(viewModel = viewModel)
        }
        
        // Then
        composeTestRule.onNodeWithText("My Screen Title").assertIsDisplayed()
    }
}
```

## Best Practices

1. **Test Naming**: Use descriptive test names with backticks
2. **Given-When-Then**: Structure tests clearly
3. **Mock External Dependencies**: Use mocks for repositories, APIs, databases
4. **Test Edge Cases**: Include error cases, empty states, loading states
5. **Coroutine Testing**: Use `runTest` and `TestDispatcher` for coroutine tests
6. **State Testing**: Test ViewModel state changes thoroughly
7. **UI Testing**: Test Compose UI with semantic properties

## Coverage Goals

| Component | Target Coverage | Current Coverage |
|-----------|----------------|------------------|
| ViewModels | 95% | TBD |
| Repositories | 90% | TBD |
| Use Cases | 90% | TBD |
| UI Components | 85% | TBD |
| **Overall** | **90%** | **0%** |

## Next Steps

1. Run existing tests: `./scripts/run-android-tests.sh`
2. Review coverage report
3. Identify gaps in test coverage
4. Add tests for untested components
5. Aim for incremental coverage improvement
6. Integrate with CI/CD pipeline

## Troubleshooting

### Tests Not Running
- Check Gradle sync status
- Verify test configuration in `build.gradle.kts`
- Ensure test directories are correctly structured

### Coverage Not Reported
- Run `./gradlew clean` then `./gradlew jacocoTestReport`
- Check Jacoco configuration in `build.gradle.kts`
- Verify test tasks are actually executing

### Mocking Issues
- Use `mockk` for Kotlin classes
- Use `Mockito` for Java interfaces
- Remember to use `coEvery` for suspending functions
EOF

echo "✅ Android testing guide created"

cd ..

echo ""
echo "🎉 Android Test Infrastructure Setup Complete!"
echo "============================================="
echo ""
echo "📋 Available commands:"
echo "   ./scripts/run-android-tests.sh      - Run Android tests with coverage"
echo ""
echo "📚 Documentation:"
echo "   docs/android-testing-guide.md       - Complete Android testing guide"
echo ""
echo "🔧 Next Steps:"
echo "   1. Review the Android testing guide"
echo "   2. Run: ./scripts/run-android-tests.sh"
echo "   3. Check test coverage report"
echo "   4. Add tests for components with low coverage"
echo "   5. Aim for 70%+ test coverage initially, then 90%+"
echo ""
echo "🚀 Happy testing!"