package com.catalogizer.androidtv.data.tv

import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.PreferenceDataStoreFactory
import androidx.datastore.preferences.core.Preferences
import com.catalogizer.androidtv.data.repository.SettingsRepository
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.runTest
import org.junit.Assert.*
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.TemporaryFolder
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner
import org.robolectric.annotation.Config

@OptIn(ExperimentalCoroutinesApi::class)
@RunWith(RobolectricTestRunner::class)
@Config(manifest = Config.NONE)
class ChannelSettingsTest {

    @get:Rule
    val tmpFolder = TemporaryFolder()

    private lateinit var dataStore: DataStore<Preferences>
    private lateinit var repo: SettingsRepository

    @Before
    fun setup() {
        dataStore = PreferenceDataStoreFactory.create {
            tmpFolder.newFile("test_prefs.preferences_pb")
        }
        repo = SettingsRepository(dataStore)
    }

    @Test
    fun `saveChannelId persists and retrieves channel ID`() = runTest {
        repo.saveChannelId("default", 42L)
        assertEquals(42L, repo.getChannelId("default"))
    }

    @Test
    fun `getChannelId returns null for unknown key`() = runTest {
        assertNull(repo.getChannelId("nonexistent"))
    }

    @Test
    fun `saveChannelId overwrites existing value`() = runTest {
        repo.saveChannelId("movie", 10L)
        repo.saveChannelId("movie", 20L)
        assertEquals(20L, repo.getChannelId("movie"))
    }

    @Test
    fun `removeChannelId clears stored ID`() = runTest {
        repo.saveChannelId("movie", 10L)
        repo.removeChannelId("movie")
        assertNull(repo.getChannelId("movie"))
    }

    @Test
    fun `clearAllChannelIds removes all channel IDs`() = runTest {
        repo.saveChannelId("default", 1L)
        repo.saveChannelId("movie", 2L)
        repo.saveChannelId("tv_show", 3L)
        repo.clearAllChannelIds()
        assertNull(repo.getChannelId("default"))
        assertNull(repo.getChannelId("movie"))
        assertNull(repo.getChannelId("tv_show"))
    }

    @Test
    fun `saveLaunchAction persists and retrieves action`() = runTest {
        repo.saveLaunchAction("movie", LaunchAction.IMMEDIATE_PLAY)
        assertEquals(LaunchAction.IMMEDIATE_PLAY, repo.getLaunchAction("movie"))
    }

    @Test
    fun `getLaunchAction returns DETAIL by default`() = runTest {
        assertEquals(LaunchAction.DETAIL, repo.getLaunchAction("movie"))
    }

    @Test
    fun `clearAllLaunchActions resets all to default`() = runTest {
        repo.saveLaunchAction("movie", LaunchAction.IMMEDIATE_PLAY)
        repo.saveLaunchAction("music", LaunchAction.IMMEDIATE_PLAY)
        repo.clearAllLaunchActions()
        assertEquals(LaunchAction.DETAIL, repo.getLaunchAction("movie"))
        assertEquals(LaunchAction.DETAIL, repo.getLaunchAction("music"))
    }

    @Test
    fun `getAllLaunchActions returns all saved actions`() = runTest {
        repo.saveLaunchAction("movie", LaunchAction.IMMEDIATE_PLAY)
        repo.saveLaunchAction("tv_show", LaunchAction.DETAIL)
        val actions = repo.getAllLaunchActions()
        assertEquals(LaunchAction.IMMEDIATE_PLAY, actions["movie"])
        assertEquals(LaunchAction.DETAIL, actions["tv_show"])
    }
}
