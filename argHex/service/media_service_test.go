package service_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/argSea/argsea-site-api/argHex/out_adapter"
	"github.com/argSea/argsea-site-api/argHex/service"
)

// newDarkroom wires a media service over the REAL webstore adapter on a temp
// dir plus an in-memory metadata fake, so uploads exercise the actual disk
// half without mongo.
func newDarkroom(t *testing.T) (in_port.MediaService, in_port.ActivityService, string) {
	t.Helper()

	dir := t.TempDir()
	activity := service.NewActivityService(out_adapter.NewActivityFakeOutAdapter())

	media := service.NewMediaService(
		out_adapter.NewMediaWebstoreAdapter(dir+string(filepath.Separator), "/media/images"),
		out_adapter.NewMediaMetaFakeOutAdapter(),
		activity,
	)

	return media, activity, dir
}

func TestCreateMediaWritesFileAndMetadata(t *testing.T) {
	media, activity, dir := newDarkroom(t)

	saved, err := media.CreateMedia("logo.png", "image/png", []byte("png-bytes"))

	if nil != err {
		t.Fatalf("upload failed: %v", err)
	}

	if "" == saved.Id || "logo.png" != saved.Filename || "/media/images/logo.png" != saved.URL || "" == saved.CreatedAt {
		t.Fatalf("unexpected media entity: %+v", saved)
	}

	bytes, readErr := os.ReadFile(filepath.Join(dir, "logo.png"))

	if nil != readErr || "png-bytes" != string(bytes) {
		t.Fatalf("expected the file developed on disk, got %q / %v", bytes, readErr)
	}

	entries, _ := activity.Recent(10)

	if 1 != len(entries) {
		t.Fatalf("expected an upload entry in the keeper's log, got %+v", entries)
	}
}

func TestCreateMediaRejectsNonImages(t *testing.T) {
	media, _, dir := newDarkroom(t)

	for _, contentType := range []string{"application/pdf", "text/html", "video/mp4", "image/svg+xml", ""} {
		if _, err := media.CreateMedia("payload.png", contentType, []byte("nope")); nil == err {
			t.Fatalf("expected content type %q rejected", contentType)
		}
	}

	entries, _ := os.ReadDir(dir)

	if 0 != len(entries) {
		t.Fatalf("a rejected upload must leave nothing on disk, found %d files", len(entries))
	}

	listed, _ := media.ListMedia()

	if 0 != len(listed) {
		t.Fatalf("a rejected upload must leave no metadata, found %d items", len(listed))
	}
}

func TestUploadMediaRejectsSvg(t *testing.T) {
	media, _, dir := newDarkroom(t)

	// the legacy base64 path (profile pictures, contact icons) runs the same
	// image-type gate as CreateMedia; an svg must never land on disk
	if _, err := media.UploadMedia("image/svg+xml", []byte("<svg onload=alert(1)>")); nil == err {
		t.Fatalf("expected image/svg+xml rejected on the base64 path")
	}

	entries, _ := os.ReadDir(dir)

	if 0 != len(entries) {
		t.Fatalf("a rejected upload must leave nothing on disk, found %d files", len(entries))
	}
}

func TestUploadMediaAcceptsImage(t *testing.T) {
	media, _, _ := newDarkroom(t)

	// a legit profile picture still goes through the base64 path
	if _, err := media.UploadMedia("image/png", []byte("png-bytes")); nil != err {
		t.Fatalf("expected image/png accepted on the base64 path, got %v", err)
	}
}

func TestCreateMediaSanitizesFilenameToItsBase(t *testing.T) {
	media, _, dir := newDarkroom(t)

	saved, err := media.CreateMedia("../../escape.png", "image/png", []byte("x"))

	if nil != err {
		t.Fatalf("upload failed: %v", err)
	}

	if "escape.png" != saved.Filename {
		t.Fatalf("expected the filename reduced to its base, got %q", saved.Filename)
	}

	if _, statErr := os.Stat(filepath.Join(dir, "escape.png")); nil != statErr {
		t.Fatalf("expected the file inside the media dir: %v", statErr)
	}
}

func TestCreateMediaRejectsDuplicateFilename(t *testing.T) {
	media, _, _ := newDarkroom(t)

	if _, err := media.CreateMedia("logo.png", "image/png", []byte("one")); nil != err {
		t.Fatalf("first upload failed: %v", err)
	}

	if _, err := media.CreateMedia("logo.png", "image/png", []byte("two")); nil == err {
		t.Fatalf("expected a duplicate filename rejected")
	}
}

func TestListMediaIsNewestFirst(t *testing.T) {
	media, _, _ := newDarkroom(t)

	media.CreateMedia("first.png", "image/png", []byte("1"))
	media.CreateMedia("second.png", "image/png", []byte("2"))

	listed, err := media.ListMedia()

	if nil != err {
		t.Fatalf("list failed: %v", err)
	}

	if 2 != len(listed) || "second.png" != listed[0].Filename || "first.png" != listed[1].Filename {
		t.Fatalf("expected newest first, got %+v", listed)
	}
}

func TestDeleteMediaRemovesFileAndMetadata(t *testing.T) {
	media, _, dir := newDarkroom(t)

	saved, _ := media.CreateMedia("gone.png", "image/png", []byte("x"))

	if err := media.DeleteMedia(saved.Id); nil != err {
		t.Fatalf("delete failed: %v", err)
	}

	listed, _ := media.ListMedia()

	if 0 != len(listed) {
		t.Fatalf("expected the metadata gone, got %+v", listed)
	}

	if _, statErr := os.Stat(filepath.Join(dir, "gone.png")); !os.IsNotExist(statErr) {
		t.Fatalf("expected the file gone from disk, got %v", statErr)
	}
}

func TestDeleteMediaRejectsUnknownId(t *testing.T) {
	media, _, _ := newDarkroom(t)

	if err := media.DeleteMedia("nope"); nil == err {
		t.Fatalf("expected delete to reject an unknown media id")
	}
}
