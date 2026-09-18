package out_port

// MediaRepo is the file half of the darkroom: bytes on disk in, web paths out.
type MediaRepo interface {
	// UploadMedia stores bytes under a random generated name (the legacy base64
	// path) and returns that name and the file's web path, in that order. The
	// name comes back because it is the only handle RemoveNamed takes and the
	// web path is not one: how a name joins onto web_path is the adapter's
	// business, so a caller carving the name back out of the path is guessing at
	// a spelling the config is free to change.
	UploadMedia(mime_type string, bytes []byte) (string, string, error)
	// SaveNamed stores bytes under exactly file_name and returns the file's web
	// path.
	SaveNamed(file_name string, bytes []byte) (string, error)
	// RemoveNamed deletes the named file from disk; a file already gone is not
	// an error.
	RemoveNamed(file_name string) error
}
