package entities

type FileType string

const (
	FileTypeNone            FileType = "fileTypeNone"
	FileTypeAnimation       FileType = "fileTypeAnimation"
	FileTypeAudio           FileType = "fileTypeAudio"
	FileTypeDocument        FileType = "fileTypeDocument"
	FileTypePhoto           FileType = "fileTypePhoto"
	FileTypeProfilePhoto    FileType = "fileTypeProfilePhoto"
	FileTypeSecret          FileType = "fileTypeSecret"
	FileTypeSecretThumbnail FileType = "fileTypeSecretThumbnail"
	FileTypeSecure          FileType = "fileTypeSecure"
	FileTypeSticker         FileType = "fileTypeSticker"
	FileTypeThumbnail       FileType = "fileTypeThumbnail"
	FileTypeUnknown         FileType = "fileTypeUnknown"
	FileTypeVideo           FileType = "fileTypeVideo"
	FileTypeVideoNote       FileType = "fileTypeVideoNote"
	FileTypeVoiceNote       FileType = "fileTypeVoiceNote"
	FileTypeWallpaper       FileType = "fileTypeWallpaper"
)
