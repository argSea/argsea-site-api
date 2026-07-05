package out_adapter

import (
	"errors"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	"github.com/argSea/argsea-site-api/argHex/out_port"
	"github.com/argSea/argsea-site-api/argHex/utility"
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

// SaveNamed writes bytes under exactly file_name and returns the web path the
// site serves it from. Joined paths (unlike the legacy concatenation above)
// tolerate a config with or without trailing slashes.
func (m mediaWebstoreAdapter) SaveNamed(file_name string, bytes []byte) (string, error) {
	file, err := os.Create(filepath.Join(m.save_path, file_name))

	if nil != err {
		return "", err
	}

	defer file.Close()

	if _, err = file.Write(bytes); nil != err {
		return "", err
	}

	return strings.TrimRight(m.web_path, "/") + "/" + file_name, nil
}

// RemoveNamed deletes the named file; a file already gone is not an error.
func (m mediaWebstoreAdapter) RemoveNamed(file_name string) error {
	err := os.Remove(filepath.Join(m.save_path, file_name))

	if nil != err && !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	return nil
}
