package out_port

// MediaRepo is the file half of the darkroom: bytes on disk in, web paths out.
type MediaRepo interface {
	// UploadMedia stores bytes under a random generated name (the legacy base64
	// path) and returns the file's web path.
	UploadMedia(mime_type string, bytes []byte) (string, error)
	// SaveNamed stores bytes under exactly file_name and returns the file's web
	// path.
	SaveNamed(file_name string, bytes []byte) (string, error)
	// RemoveNamed deletes the named file from disk; a file already gone is not
	// an error.
	RemoveNamed(file_name string) error
}
