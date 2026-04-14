package com.catalogizer.android.utils

import android.content.Context
import androidx.biometric.BiometricManager
import io.mockk.*
import org.junit.After
import org.junit.Assert.*
import org.junit.Before
import org.junit.Test

class BiometricHelperTest {

    private val mockContext = mockk<Context>(relaxed = true)
    private val mockBiometricManager = mockk<BiometricManager>()

    @Before
    fun setup() {
        mockkStatic(BiometricManager::class)
        every { BiometricManager.from(any()) } returns mockBiometricManager
    }

    @After
    fun tearDown() {
        clearAllMocks()
        unmockkStatic(BiometricManager::class)
    }

    // --- isAvailable ---

    @Test
    fun `isAvailable returns true when biometric hardware is available and enrolled`() {
        every {
            mockBiometricManager.canAuthenticate(BiometricManager.Authenticators.BIOMETRIC_STRONG)
        } returns BiometricManager.BIOMETRIC_SUCCESS

        val result = BiometricHelper.isAvailable(mockContext)

        assertTrue(result)
    }

    @Test
    fun `isAvailable returns false when no biometric hardware`() {
        every {
            mockBiometricManager.canAuthenticate(BiometricManager.Authenticators.BIOMETRIC_STRONG)
        } returns BiometricManager.BIOMETRIC_ERROR_NO_HARDWARE

        val result = BiometricHelper.isAvailable(mockContext)

        assertFalse(result)
    }

    @Test
    fun `isAvailable returns false when hardware unavailable`() {
        every {
            mockBiometricManager.canAuthenticate(BiometricManager.Authenticators.BIOMETRIC_STRONG)
        } returns BiometricManager.BIOMETRIC_ERROR_HW_UNAVAILABLE

        val result = BiometricHelper.isAvailable(mockContext)

        assertFalse(result)
    }

    @Test
    fun `isAvailable returns false when no biometrics enrolled`() {
        every {
            mockBiometricManager.canAuthenticate(BiometricManager.Authenticators.BIOMETRIC_STRONG)
        } returns BiometricManager.BIOMETRIC_ERROR_NONE_ENROLLED

        val result = BiometricHelper.isAvailable(mockContext)

        assertFalse(result)
    }

    @Test
    fun `isAvailable returns false for security update required`() {
        every {
            mockBiometricManager.canAuthenticate(BiometricManager.Authenticators.BIOMETRIC_STRONG)
        } returns BiometricManager.BIOMETRIC_ERROR_SECURITY_UPDATE_REQUIRED

        val result = BiometricHelper.isAvailable(mockContext)

        assertFalse(result)
    }

    @Test
    fun `isAvailable returns false for unsupported status code`() {
        every {
            mockBiometricManager.canAuthenticate(BiometricManager.Authenticators.BIOMETRIC_STRONG)
        } returns BiometricManager.BIOMETRIC_STATUS_UNKNOWN

        val result = BiometricHelper.isAvailable(mockContext)

        assertFalse(result)
    }

    @Test
    fun `isAvailable uses BIOMETRIC_STRONG authenticator`() {
        every {
            mockBiometricManager.canAuthenticate(any())
        } returns BiometricManager.BIOMETRIC_SUCCESS

        BiometricHelper.isAvailable(mockContext)

        verify {
            mockBiometricManager.canAuthenticate(BiometricManager.Authenticators.BIOMETRIC_STRONG)
        }
    }

    @Test
    fun `isAvailable calls BiometricManager from with provided context`() {
        every {
            mockBiometricManager.canAuthenticate(any())
        } returns BiometricManager.BIOMETRIC_SUCCESS

        BiometricHelper.isAvailable(mockContext)

        verify { BiometricManager.from(mockContext) }
    }

    // --- BiometricManager status codes coverage ---

    @Test
    fun `all BiometricManager error codes return false`() {
        val errorCodes = listOf(
            BiometricManager.BIOMETRIC_ERROR_NO_HARDWARE,
            BiometricManager.BIOMETRIC_ERROR_HW_UNAVAILABLE,
            BiometricManager.BIOMETRIC_ERROR_NONE_ENROLLED,
            BiometricManager.BIOMETRIC_ERROR_SECURITY_UPDATE_REQUIRED,
            BiometricManager.BIOMETRIC_STATUS_UNKNOWN
        )

        for (code in errorCodes) {
            every {
                mockBiometricManager.canAuthenticate(BiometricManager.Authenticators.BIOMETRIC_STRONG)
            } returns code

            assertFalse(
                "Expected false for error code $code",
                BiometricHelper.isAvailable(mockContext)
            )
        }
    }

    @Test
    fun `only BIOMETRIC_SUCCESS returns true`() {
        every {
            mockBiometricManager.canAuthenticate(BiometricManager.Authenticators.BIOMETRIC_STRONG)
        } returns BiometricManager.BIOMETRIC_SUCCESS

        assertTrue(BiometricHelper.isAvailable(mockContext))
    }
}
