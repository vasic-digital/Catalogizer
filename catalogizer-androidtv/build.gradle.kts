// Top-level build file where you can add configuration options common to all sub-projects/modules.
plugins {
    id("com.android.application") version "8.2.2" apply false
    id("org.jetbrains.kotlin.android") version "1.9.22" apply false
    id("com.google.dagger.hilt.android") version "2.48" apply false
    id("org.jetbrains.kotlin.plugin.serialization") version "1.9.22" apply false
    // §11.4.107/.158/.159 host-side visual-proof harness — renders real Compose-for-TV
    // screens to PNG on the JVM (no device). Pinned to the Robolectric-4.11.1 /
    // Kotlin-1.9.22 / Compose-1.6.8 era per cited research (do NOT bump without re-pinning).
    id("io.github.takahirom.roborazzi") version "1.13.0" apply false
    // Firebase — Analytics, Crashlytics, Performance, App Distribution
    id("com.google.gms.google-services") version "4.4.2" apply false
    id("com.google.firebase.crashlytics") version "3.0.3" apply false
    id("com.google.firebase.firebase-perf") version "1.4.2" apply false
}

buildscript {
    dependencies {
        classpath("com.google.dagger:hilt-android-gradle-plugin:2.48")
    }
}