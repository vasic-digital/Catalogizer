package com.catalogizer.androidtv.ui.theme

import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.sp
import org.junit.Assert.*
import org.junit.Test

/**
 * Tests for TV typography definitions: validates font sizes are scaled
 * for 10-foot TV viewing, verifies font weight assignments, and confirms
 * readability constraints.
 */
class TypeTest {

    // --- Display styles ---

    @Test
    fun `displayLarge fontSize is 26sp`() {
        assertEquals(26.sp, TVTypography.displayLarge.fontSize)
    }

    @Test
    fun `displayLarge lineHeight is 32sp`() {
        assertEquals(32.sp, TVTypography.displayLarge.lineHeight)
    }

    @Test
    fun `displayLarge fontWeight is Normal`() {
        assertEquals(FontWeight.Normal, TVTypography.displayLarge.fontWeight)
    }

    @Test
    fun `displayLarge letterSpacing is negative 0_25sp`() {
        assertEquals((-0.25).sp, TVTypography.displayLarge.letterSpacing)
    }

    @Test
    fun `displayMedium fontSize is 22sp`() {
        assertEquals(22.sp, TVTypography.displayMedium.fontSize)
    }

    @Test
    fun `displayMedium lineHeight is 28sp`() {
        assertEquals(28.sp, TVTypography.displayMedium.lineHeight)
    }

    @Test
    fun `displaySmall fontSize is 18sp`() {
        assertEquals(18.sp, TVTypography.displaySmall.fontSize)
    }

    @Test
    fun `displaySmall lineHeight is 24sp`() {
        assertEquals(24.sp, TVTypography.displaySmall.lineHeight)
    }

    // --- Headline styles ---

    @Test
    fun `headlineLarge fontSize is 16sp`() {
        assertEquals(16.sp, TVTypography.headlineLarge.fontSize)
    }

    @Test
    fun `headlineLarge lineHeight is 22sp`() {
        assertEquals(22.sp, TVTypography.headlineLarge.lineHeight)
    }

    @Test
    fun `headlineMedium fontSize is 15sp`() {
        assertEquals(15.sp, TVTypography.headlineMedium.fontSize)
    }

    @Test
    fun `headlineMedium lineHeight is 20sp`() {
        assertEquals(20.sp, TVTypography.headlineMedium.lineHeight)
    }

    @Test
    fun `headlineSmall fontSize is 14sp`() {
        assertEquals(14.sp, TVTypography.headlineSmall.fontSize)
    }

    @Test
    fun `headlineSmall lineHeight is 18sp`() {
        assertEquals(18.sp, TVTypography.headlineSmall.lineHeight)
    }

    // --- Title styles ---

    @Test
    fun `titleLarge fontSize is 13sp`() {
        assertEquals(13.sp, TVTypography.titleLarge.fontSize)
    }

    @Test
    fun `titleLarge fontWeight is Medium`() {
        assertEquals(FontWeight.Medium, TVTypography.titleLarge.fontWeight)
    }

    @Test
    fun `titleMedium fontSize is 12sp`() {
        assertEquals(12.sp, TVTypography.titleMedium.fontSize)
    }

    @Test
    fun `titleMedium fontWeight is Medium`() {
        assertEquals(FontWeight.Medium, TVTypography.titleMedium.fontWeight)
    }

    @Test
    fun `titleSmall fontSize is 11sp`() {
        assertEquals(11.sp, TVTypography.titleSmall.fontSize)
    }

    @Test
    fun `titleSmall fontWeight is Medium`() {
        assertEquals(FontWeight.Medium, TVTypography.titleSmall.fontWeight)
    }

    // --- Body styles ---

    @Test
    fun `bodyLarge fontSize is 12sp`() {
        assertEquals(12.sp, TVTypography.bodyLarge.fontSize)
    }

    @Test
    fun `bodyLarge fontWeight is Normal`() {
        assertEquals(FontWeight.Normal, TVTypography.bodyLarge.fontWeight)
    }

    @Test
    fun `bodyMedium fontSize is 11sp`() {
        assertEquals(11.sp, TVTypography.bodyMedium.fontSize)
    }

    @Test
    fun `bodySmall fontSize is 10sp`() {
        assertEquals(10.sp, TVTypography.bodySmall.fontSize)
    }

    // --- Label styles ---

    @Test
    fun `labelLarge fontSize is 11sp`() {
        assertEquals(11.sp, TVTypography.labelLarge.fontSize)
    }

    @Test
    fun `labelLarge fontWeight is Medium`() {
        assertEquals(FontWeight.Medium, TVTypography.labelLarge.fontWeight)
    }

    @Test
    fun `labelMedium fontSize is 10sp`() {
        assertEquals(10.sp, TVTypography.labelMedium.fontSize)
    }

    @Test
    fun `labelSmall fontSize is 10sp`() {
        assertEquals(10.sp, TVTypography.labelSmall.fontSize)
    }

    @Test
    fun `labelSmall fontWeight is Medium`() {
        assertEquals(FontWeight.Medium, TVTypography.labelSmall.fontWeight)
    }

    // --- Font family ---

    @Test
    fun `all styles use default font family`() {
        assertEquals(FontFamily.Default, TVTypography.displayLarge.fontFamily)
        assertEquals(FontFamily.Default, TVTypography.displayMedium.fontFamily)
        assertEquals(FontFamily.Default, TVTypography.displaySmall.fontFamily)
        assertEquals(FontFamily.Default, TVTypography.headlineLarge.fontFamily)
        assertEquals(FontFamily.Default, TVTypography.headlineMedium.fontFamily)
        assertEquals(FontFamily.Default, TVTypography.headlineSmall.fontFamily)
        assertEquals(FontFamily.Default, TVTypography.titleLarge.fontFamily)
        assertEquals(FontFamily.Default, TVTypography.titleMedium.fontFamily)
        assertEquals(FontFamily.Default, TVTypography.titleSmall.fontFamily)
        assertEquals(FontFamily.Default, TVTypography.bodyLarge.fontFamily)
        assertEquals(FontFamily.Default, TVTypography.bodyMedium.fontFamily)
        assertEquals(FontFamily.Default, TVTypography.bodySmall.fontFamily)
        assertEquals(FontFamily.Default, TVTypography.labelLarge.fontFamily)
        assertEquals(FontFamily.Default, TVTypography.labelMedium.fontFamily)
        assertEquals(FontFamily.Default, TVTypography.labelSmall.fontFamily)
    }

    // --- TV readability requirements ---

    @Test
    fun `no font size is below 10sp for TV readability`() {
        val allFontSizes = listOf(
            TVTypography.displayLarge.fontSize,
            TVTypography.displayMedium.fontSize,
            TVTypography.displaySmall.fontSize,
            TVTypography.headlineLarge.fontSize,
            TVTypography.headlineMedium.fontSize,
            TVTypography.headlineSmall.fontSize,
            TVTypography.titleLarge.fontSize,
            TVTypography.titleMedium.fontSize,
            TVTypography.titleSmall.fontSize,
            TVTypography.bodyLarge.fontSize,
            TVTypography.bodyMedium.fontSize,
            TVTypography.bodySmall.fontSize,
            TVTypography.labelLarge.fontSize,
            TVTypography.labelMedium.fontSize,
            TVTypography.labelSmall.fontSize
        )
        for (fontSize in allFontSizes) {
            assertTrue(
                "Font size $fontSize is below minimum 10sp for TV",
                fontSize >= 10.sp
            )
        }
    }

    // --- Size hierarchy ---

    @Test
    fun `display sizes decrease from large to small`() {
        assertTrue(TVTypography.displayLarge.fontSize > TVTypography.displayMedium.fontSize)
        assertTrue(TVTypography.displayMedium.fontSize > TVTypography.displaySmall.fontSize)
    }

    @Test
    fun `headline sizes decrease from large to small`() {
        assertTrue(TVTypography.headlineLarge.fontSize > TVTypography.headlineMedium.fontSize)
        assertTrue(TVTypography.headlineMedium.fontSize > TVTypography.headlineSmall.fontSize)
    }

    @Test
    fun `title sizes decrease from large to small`() {
        assertTrue(TVTypography.titleLarge.fontSize > TVTypography.titleMedium.fontSize)
        assertTrue(TVTypography.titleMedium.fontSize > TVTypography.titleSmall.fontSize)
    }

    @Test
    fun `body sizes decrease from large to small`() {
        assertTrue(TVTypography.bodyLarge.fontSize > TVTypography.bodyMedium.fontSize)
        assertTrue(TVTypography.bodyMedium.fontSize > TVTypography.bodySmall.fontSize)
    }

    @Test
    fun `display is larger than headline`() {
        assertTrue(TVTypography.displaySmall.fontSize > TVTypography.headlineLarge.fontSize)
    }

    @Test
    fun `headline is larger than title`() {
        assertTrue(TVTypography.headlineSmall.fontSize > TVTypography.titleLarge.fontSize)
    }

    @Test
    fun `title is larger than body at same scale`() {
        assertTrue(TVTypography.titleSmall.fontSize >= TVTypography.bodyMedium.fontSize)
    }

    // --- Line height proportions ---

    @Test
    fun `line heights are larger than font sizes`() {
        assertTrue(TVTypography.displayLarge.lineHeight > TVTypography.displayLarge.fontSize)
        assertTrue(TVTypography.headlineLarge.lineHeight > TVTypography.headlineLarge.fontSize)
        assertTrue(TVTypography.titleLarge.lineHeight > TVTypography.titleLarge.fontSize)
        assertTrue(TVTypography.bodyLarge.lineHeight > TVTypography.bodyLarge.fontSize)
        assertTrue(TVTypography.labelLarge.lineHeight > TVTypography.labelLarge.fontSize)
    }

    // --- Letter spacing ---

    @Test
    fun `body and label styles have positive letter spacing`() {
        assertTrue(TVTypography.bodyLarge.letterSpacing >= 0.sp)
        assertTrue(TVTypography.bodyMedium.letterSpacing >= 0.sp)
        assertTrue(TVTypography.bodySmall.letterSpacing >= 0.sp)
        assertTrue(TVTypography.labelLarge.letterSpacing >= 0.sp)
        assertTrue(TVTypography.labelMedium.letterSpacing >= 0.sp)
        assertTrue(TVTypography.labelSmall.letterSpacing >= 0.sp)
    }
}
