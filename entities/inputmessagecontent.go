package entities

type InputMessageContent interface {
	Type() string
}

// NewInputMessageFormattedText creates a new InputMessageText
func NewInputMessageFormattedText(
	text string,
	disableWebPagePreview bool,
	clearDraft bool,
	entities []TextEntity,
) *InputMessageText {
	return &InputMessageText{
		Text:                  &FormattedText{Text: text, Entities: entities},
		DisableWebPagePreview: disableWebPagePreview,
		ClearDraft:            clearDraft,
	}
}

// NewInputMessageVideoFileID creates a new InputMessageVideo
func NewInputMessageVideoFileID(
	fileID int64,
	thumb *InputThumbnail,
	caption *FormattedText,
	width int64,
	height int64,
	duration int64,
	supportsStreaming bool,
	ttl int64,
) *InputMessageVideo {
	return &InputMessageVideo{
		Video:             NewInputFileID(fileID),
		Thumbnail:         thumb,
		Caption:           caption,
		Width:             width,
		Height:            height,
		Duration:          duration,
		SupportsStreaming: supportsStreaming,
		Ttl:               ttl,
	}
}

// NewInputMessageVideoRemote creates a new InputMessageVideo
func NewInputMessageVideoRemote(
	remoteFileID string,
	thumb *InputThumbnail,
	caption *FormattedText,
	width int64,
	height int64,
	duration int64,
	supportsStreaming bool,
	ttl int64,
) *InputMessageVideo {
	return &InputMessageVideo{
		Video:             NewInputFileRemote(remoteFileID),
		Thumbnail:         thumb,
		Caption:           caption,
		Width:             width,
		Height:            height,
		Duration:          duration,
		SupportsStreaming: supportsStreaming,
		Ttl:               ttl,
	}
}

// NewInputMessageVideoLocal creates a new InputMessageVideo
func NewInputMessageVideoLocal(
	path string,
	thumb *InputThumbnail,
	caption *FormattedText,
	width int64,
	height int64,
	duration int64,
	supportsStreaming bool,
	ttl int64,
) *InputMessageVideo {
	return &InputMessageVideo{
		Video:             NewInputFileLocal(path),
		Thumbnail:         thumb,
		Caption:           caption,
		Width:             width,
		Height:            height,
		Duration:          duration,
		SupportsStreaming: supportsStreaming,
		Ttl:               ttl,
	}
}

// NewInputMessagePhotoFileID creates a new InputMessagePhoto
func NewInputMessagePhotoFileID(
	fileID int64,
	thumbnail *InputThumbnail,
	width int64,
	height int64,
	caption *FormattedText,
) *InputMessagePhoto {
	return &InputMessagePhoto{
		Photo:     NewInputFileID(fileID),
		Thumbnail: thumbnail,
		Width:     width,
		Height:    height,
		Caption:   caption,
	}
}

func NewInputMessagePhotoRemote(
	remoteFileID string,
	thumbnail *InputThumbnail,
	width int64,
	height int64,
	caption *FormattedText,
) *InputMessagePhoto {
	return &InputMessagePhoto{
		Photo:     NewInputFileRemote(remoteFileID),
		Thumbnail: thumbnail,
		Width:     width,
		Height:    height,
		Caption:   caption,
	}
}

func NewInputMessagePhotoLocal(
	path string,
	thumbnail *InputThumbnail,
	width int64,
	height int64,
	caption *FormattedText,
) *InputMessagePhoto {
	return &InputMessagePhoto{
		Photo:     NewInputFileLocal(path),
		Thumbnail: thumbnail,
		Width:     width,
		Height:    height,
		Caption:   caption,
	}
}
