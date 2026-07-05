package in_port

import "github.com/argSea/argsea-site-api/argHex/domain"

// MediaService is the darkroom seam: named uploads carry mongo metadata plus a
// file on disk, while UploadMedia stays the legacy base64 path (disk-only,
// random name) the user adapter still calls.
type MediaService interface {
	UploadMedia(mime_type string, bytes []byte) (string, error)
	ListMedia() (domain.MediaList, error)
	CreateMedia(file_name string, mime_type string, bytes []byte) (domain.Media, error)
	DeleteMedia(media_id string) error
}
