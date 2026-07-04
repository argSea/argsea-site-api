package out_adapter

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/argSea/portfolio_blog_api/argHex/domain"
	"github.com/argSea/portfolio_blog_api/argHex/out_port"
	"github.com/argSea/portfolio_blog_api/argHex/stores"
	"github.com/argSea/portfolio_blog_api/argHex/utility"
)

type mediaMongoWebstoreAdapter struct {
	store     *stores.Mordor
	save_path string
	web_path  string
}

func NewMediaMongoWebstoreAdapter(store *stores.Mordor, save_path string, web_path string) out_port.MediaRepo {
	return mediaMongoWebstoreAdapter{
		store:     store,
		save_path: save_path,
		web_path:  web_path,
	}
}

func (m mediaMongoWebstoreAdapter) UploadMedia(mime_type string, bytes []byte) (string, error) {
	file_type := utility.MimeToFileExt(mime_type)

	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, 16)

	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}

	save_path := m.save_path
	web_path := m.web_path

	file_name := string(b) + file_type

	// open file handle
	file, err := os.Create(save_path + file_name)

	if err != nil {
		return "", err
	}

	defer file.Close()

	// write bytes to file
	_, err = file.Write(bytes)

	if err != nil {
		return "", err
	}

	// create media object
	media := domain.Media{
		Title:     file_name,
		Filename:  web_path + file_name,
		URL:       web_path + file_name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// save media object
	new_id, write_err := m.store.Write(media)

	if write_err != nil {
		return "", write_err
	}

	// return file path
	return new_id, nil
}

func (m mediaMongoWebstoreAdapter) GetMedia(media_id string) (domain.Media, error) {
	var media domain.Media
	err := m.store.Get("_id", media_id, &media)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v", err)
		return domain.Media{}, err
	}

	return media, nil

}

func (m mediaMongoWebstoreAdapter) DeleteMedia(media_id string) error {
	return nil
}
