package service

import (
	"errors"
	"log"
	"path/filepath"
	"sort"
	"strings"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/argSea/argsea-site-api/argHex/out_port"
)

// The darkroom takes photographs only — these content types are the whole
// vocabulary, and anything else is rejected before a byte lands on disk.
var mediaImageTypes = map[string]bool{
	"image/png":     true,
	"image/jpeg":    true,
	"image/jpg":     true,
	"image/gif":     true,
	"image/svg+xml": true,
	"image/webp":    true,
}

type mediaService struct {
	mediaRepo out_port.MediaRepo
	meta      out_port.MediaMetaRepo
	activity  in_port.ActivityService
}

// NewMediaService wires the darkroom onto its two halves — files on disk,
// metadata in mongo — plus the ship's log.
func NewMediaService(mediaRepo out_port.MediaRepo, meta out_port.MediaMetaRepo, activity in_port.ActivityService) in_port.MediaService {
	return mediaService{
		mediaRepo: mediaRepo,
		meta:      meta,
		activity:  activity,
	}
}

// UploadMedia is the legacy base64 path the user adapter still calls: bytes to
// disk under a random name, no metadata document.
func (m mediaService) UploadMedia(mime_type string, bytes []byte) (string, error) {
	return m.mediaRepo.UploadMedia(mime_type, bytes)
}

// ListMedia returns every darkroom item newest first. Fixed-width stamps make
// the reverse string sort chronological.
func (m mediaService) ListMedia() (domain.MediaList, error) {
	media, err := m.meta.List()

	if nil != err {
		return nil, err
	}

	sort.SliceStable(media, func(i, j int) bool {
		return media[i].CreatedAt > media[j].CreatedAt
	})

	return media, nil
}

// CreateMedia develops a named upload: image content types only, the file on
// disk and a metadata document in mongo. The filename is reduced to its base
// so an upload can never escape the media directory.
func (m mediaService) CreateMedia(file_name string, mime_type string, bytes []byte) (domain.Media, error) {
	if !mediaImageTypes[mime_type] {
		return domain.Media{}, errors.New("only image uploads are allowed (png, jpeg, gif, svg, webp)")
	}

	file_name = filepath.Base(strings.TrimSpace(file_name))

	if "" == file_name || "." == file_name || ".." == file_name || string(filepath.Separator) == file_name {
		return domain.Media{}, errors.New("a filename is required")
	}

	// one name, one file — a duplicate would silently overwrite the first
	// print while both metadata documents kept pointing at it
	existing, err := m.meta.List()

	if nil != err {
		return domain.Media{}, err
	}

	for _, item := range existing {
		if item.Filename == file_name {
			return domain.Media{}, errors.New("a media item named \"" + file_name + "\" already exists")
		}
	}

	url, err := m.mediaRepo.SaveNamed(file_name, bytes)

	if nil != err {
		return domain.Media{}, err
	}

	id, err := m.meta.Add(domain.Media{
		Filename:  file_name,
		URL:       url,
		CreatedAt: nowStamp(),
	})

	if nil != err {
		// the file half landed but the metadata half didn't — pull the file back
		// so the darkroom never holds an orphan print
		if removeErr := m.mediaRepo.RemoveNamed(file_name); nil != removeErr {
			log.Printf("could not remove orphaned media file %v: %v\n", file_name, removeErr)
		}

		return domain.Media{}, err
	}

	m.record("media \""+file_name+"\" uploaded", id)

	return m.meta.Get(id), nil
}

// DeleteMedia removes the metadata document and the file behind it. Documents
// that reference the filename are deliberately untouched — detaching is the
// admin's job client-side.
func (m mediaService) DeleteMedia(media_id string) error {
	media := m.meta.Get(media_id)

	if "" == media.Id {
		return errors.New("media not found")
	}

	if err := m.meta.Remove(media_id); nil != err {
		return err
	}

	if err := m.mediaRepo.RemoveNamed(media.Filename); nil != err {
		return err
	}

	m.record("media \""+media.Filename+"\" deleted", media_id)

	return nil
}

func (m mediaService) record(message string, id string) {
	if err := m.activity.Record(message, domain.EntityMedia, id); nil != err {
		log.Printf("activity record failed for media %v: %v\n", id, err)
	}
}
