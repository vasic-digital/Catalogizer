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
