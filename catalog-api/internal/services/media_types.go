package services

// MediaType represents the type of media content
type MediaType string

const (
	// Video content
	MediaTypeMovie       MediaType = "movie"
	MediaTypeTVSeries    MediaType = "tv_series"
	MediaTypeTVEpisode   MediaType = "tv_episode"
	MediaTypeConcert     MediaType = "concert"
	MediaTypeDocumentary MediaType = "documentary"
	MediaTypeCourse      MediaType = "course"
	MediaTypeTraining    MediaType = "training"
	MediaTypeVideo       MediaType = "video"

	// Audio content
	MediaTypeMusic     MediaType = "music"
	MediaTypeAlbum     MediaType = "album"
	MediaTypeAudiobook MediaType = "audiobook"
	MediaTypePodcast   MediaType = "podcast"

	// Games and Software
	MediaTypeGame       MediaType = "game"
	MediaTypeGameOS     MediaType = "game_os"
	MediaTypeSoftware   MediaType = "software"
	MediaTypeSoftwareOS MediaType = "software_os"

	// Books and Documents
	MediaTypeBook      MediaType = "book"
	MediaTypeEbook     MediaType = "ebook"
	MediaTypeComicBook MediaType = "comic_book"
	MediaTypeMagazine  MediaType = "magazine"
	MediaTypeNewspaper MediaType = "newspaper"
	MediaTypeJournal   MediaType = "journal"
	MediaTypeManual    MediaType = "manual"
	MediaTypeDocument  MediaType = "document"
)

// PlaybackState represents the current playback state
type PlaybackState string

const (
	PlaybackStatePlaying PlaybackState = "playing"
	PlaybackStatePaused  PlaybackState = "paused"
	PlaybackStateStopped PlaybackState = "stopped"
	PlaybackStateLoading PlaybackState = "loading"
	PlaybackStateError   PlaybackState = "error"
)

// RepeatMode represents repeat modes for media playback
type RepeatMode string

const (
	RepeatModeOff     RepeatMode = "off"
	RepeatModeTrack   RepeatMode = "track"
	RepeatModeAlbum   RepeatMode = "album"
	RepeatModeAll     RepeatMode = "all"
	RepeatModeShuffle RepeatMode = "shuffle"
)

// Quality represents media quality levels
type Quality string

const (
	QualityLow   Quality = "low"
	QualityUltra Quality = "ultra"
)
