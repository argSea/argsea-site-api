package in_port

import "github.com/argSea/argsea-site-api/argHex/domain"

type MediaService interface {
	UploadMedia(mime_type string, bytes []byte) (string, error)
	GetMedia(media_id string) (domain.Media, error)
	DeleteMedia(media_id string) error
}
