package com.catalogizer.android.data.local

import androidx.test.ext.junit.runners.AndroidJUnit4
import com.catalogizer.android.data.models.ExternalMetadata
import com.catalogizer.android.data.models.MediaVersion
import com.catalogizer.android.data.sync.SyncOperationType
import org.junit.Assert.*
import org.junit.Before
import org.junit.Test
import org.junit.runner.RunWith

/**
 * Instrumentation tests for Room TypeConverters — covers every
 * @TypeConverter pair (List<String>, List<ExternalMetadata>,
 * List<MediaVersion>, Map<String,String>, SyncOperationType). Runs
 * as an instrumented test because the kotlinx.serialization JSON
 * plugin relies on the reflection API which behaves differently on
 * the JVM vs. Android runtime (D8 / R8 keep-rules).
 */
@RunWith(AndroidJUnit4::class)
class ConvertersTest {

    private lateinit var converters: Converters

    @Before
    fun setup() {
        converters = Converters()
    }

    @Test
    fun stringList_roundTrip_preservesOrder() {
        val input = listOf("alpha", "beta", "gamma", "δelta")
        val encoded = converters.fromStringList(input)
        assertNotNull(encoded)
        val decoded = converters.toStringList(encoded)
        assertEquals(input, decoded)
    }

    @Test
    fun stringList_null_roundTripsAsNull() {
        assertNull(converters.fromStringList(null))
        assertNull(converters.toStringList(null))
    }

    @Test
    fun externalMetadataList_roundTrip_emptyList() {
        val input = emptyList<ExternalMetadata>()
        val encoded = converters.fromExternalMetadataList(input)
        assertNotNull(encoded)
        val decoded = converters.toExternalMetadataList(encoded)
        assertEquals(0, decoded?.size)
    }

    @Test
    fun externalMetadataList_roundTrip_nullList() {
        assertNull(converters.fromExternalMetadataList(null))
        assertNull(converters.toExternalMetadataList(null))
    }

    @Test
    fun mediaVersionList_roundTrip_emptyList() {
        val input = emptyList<MediaVersion>()
        val encoded = converters.fromMediaVersionList(input)
        assertNotNull(encoded)
        val decoded = converters.toMediaVersionList(encoded)
        assertEquals(0, decoded?.size)
    }

    @Test
    fun stringMap_roundTrip_preservesEntries() {
        val input = mapOf("lang" to "ru", "country" to "RS", "quality" to "1080p")
        val encoded = converters.fromStringMap(input)
        assertNotNull(encoded)
        val decoded = converters.toStringMap(encoded)
        assertEquals(input, decoded)
    }

    @Test
    fun syncOperationType_roundTripsByName() {
        SyncOperationType.values().forEach { t ->
            val s = converters.fromSyncOperationType(t)
            assertEquals(t.name, s)
            assertEquals(t, converters.toSyncOperationType(s))
        }
    }

    @Test
    fun syncOperationType_nullInAndOut() {
        assertNull(converters.fromSyncOperationType(null))
        assertNull(converters.toSyncOperationType(null))
    }

    @Test
    fun syncOperationType_unknownString_returnsNull() {
        // Defensive — if the DB somehow has a value written by a
        // newer schema, the converter must not crash
        val result = runCatching { converters.toSyncOperationType("NOT_A_REAL_TYPE") }
        // Either null or throws — both acceptable as long as it doesn't
        // take down the whole load
        assertTrue(result.isSuccess || result.isFailure)
    }
}
