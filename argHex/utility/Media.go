package utility

// global map of mime types to file extensions
var mimeToFileExt = map[string]string{
	// images
	"image/jpeg":    ".jpg",
	"image/png":     ".png",
	"image/gif":     ".gif",
	"image/bmp":     ".bmp",
	"image/webp":    ".webp",
	"image/svg+xml": ".svg",
	"image/tiff":    ".tiff",

	// video
	"video/mp4":  ".mp4",
	"video/mpeg": ".mpeg",
	"video/ogg":  ".ogg",
	"video/webm": ".webm",
	"video/3gpp": ".3gp",
}

func MimeToFileExt(mime_type string) string {
	return mimeToFileExt[mime_type]
}
