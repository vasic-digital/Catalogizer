package com.catalogizer.androidtv.data.tv

import org.junit.Assert.*
import org.junit.Test

class TvChannelSyncWorkerTest {

    @Test
    fun `SYNC_INTERVAL_HOURS is 6`() {
        assertEquals(6L, TvChannelSyncWorker.SYNC_INTERVAL_HOURS)
    }

    @Test
    fun `WORK_NAME is correct`() {
        assertEquals("tv_channel_sync", TvChannelSyncWorker.WORK_NAME)
    }
}
