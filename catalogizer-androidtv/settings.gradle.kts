pluginManagement {
    repositories {
        google()
        mavenCentral()
        gradlePluginPortal()
    }
}
dependencyResolutionManagement {
    repositoriesMode.set(RepositoriesMode.FAIL_ON_PROJECT_REPOS)
    repositories {
        google()
        mavenCentral()
    }
}

// Enable toolchain download repositories for Java toolchain auto-provisioning
plugins {
    id("org.gradle.toolchains.foojay-resolver-convention") version "0.8.0"
}

rootProject.name = "catalogizer-androidtv"
include(":app")

// Android-Toolkit submodule is available at ../Android-Toolkit/
// To include specific modules (when build configs are aligned):
// includeBuild("../Android-Toolkit")
