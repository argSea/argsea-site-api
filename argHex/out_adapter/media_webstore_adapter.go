package out_adapter

import (
	"math/rand"
	"os"

	"github.com/argSea/portfolio_blog_api/argHex/domain"
	"github.com/argSea/portfolio_blog_api/argHex/out_port"
	"github.com/argSea/portfolio_blog_api/argHex/utility"
)

type mediaWebstoreAdapter struct {
	save_path string
	web_path  string
}

func NewMediaWebstoreAdapter(save_path string, web_path string) out_port.MediaRepo {
	return mediaWebstoreAdapter{
		save_path: save_path,
		web_path:  web_path,
	}
}

func (m mediaWebstoreAdapter) UploadMedia(mime_type string, bytes []byte) (string, error) {
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

	// return file path
	return web_path + file_name, nil
}

func (m mediaWebstoreAdapter) GetMedia(media_id string) (domain.Media, error) {
	return domain.Media{}, nil
}

func (m mediaWebstoreAdapter) DeleteMedia(media_id string) error {
	return nil
}
