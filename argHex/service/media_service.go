package service

import (
	"github.com/argSea/portfolio_blog_api/argHex/domain"
	"github.com/argSea/portfolio_blog_api/argHex/in_port"
	"github.com/argSea/portfolio_blog_api/argHex/out_port"
)

type mediaService struct {
	mediaRepo out_port.MediaRepo
}

//NewMediaService creates a new media service
func NewMediaService(mediaRepo out_port.MediaRepo) in_port.MediaService {
	return mediaService{
		mediaRepo: mediaRepo,
	}
}

func (m mediaService) UploadMedia(mime_type string, bytes []byte) (string, error) {
	return m.mediaRepo.UploadMedia(mime_type, bytes)
}

func (m mediaService) GetMedia(media_id string) (domain.Media, error) {
	return m.mediaRepo.GetMedia(media_id)
}

func (m mediaService) DeleteMedia(media_id string) error {
	return m.mediaRepo.DeleteMedia(media_id)
}
