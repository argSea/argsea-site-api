package service_test

import (
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/argSea/argsea-site-api/argHex/out_adapter"
	"github.com/argSea/argsea-site-api/argHex/service"
)

// newShelf wires a resume service over the REAL webstore adapter on a temp dir
// plus an in-memory record fake, so uploads exercise the actual disk half
// without mongo.
func newShelf(t *testing.T) (in_port.ResumeService, in_port.ActivityService, string) {
	t.Helper()

	dir := t.TempDir()
	activity := service.NewActivityService(out_adapter.NewActivityFakeOutAdapter())

	resumes := service.NewResumeService(
		out_adapter.NewResumeFakeOutAdapter(),
		out_adapter.NewMediaWebstoreAdapter(dir+string(filepath.Separator), "/media/images/"),
		activity,
	)

	return resumes, activity, dir
}

// publishedShelf wires a shelf holding one published cut, which is what the
// hoist guard asks for before it starts anything. Nothing here uploads, so the
// file half is never reached.
func publishedShelf() in_port.ResumeService {
	repo := out_adapter.NewResumeFakeOutAdapter()
	repo.Add(domain.Resume{Title: "senior software engineer", Published: true, Filename: "kept.pdf", URL: "/media/images/kept.pdf"})

	return service.NewResumeService(
		repo,
		out_adapter.NewMediaWebstoreAdapter("", "/media/images/"),
		service.NewActivityService(out_adapter.NewActivityFakeOutAdapter()),
	)
}

// emptyShelf wires a shelf with nothing on it, the state the hoist guard
// refuses on.
func emptyShelf() in_port.ResumeService {
	return service.NewResumeService(
		out_adapter.NewResumeFakeOutAdapter(),
		out_adapter.NewMediaWebstoreAdapter("", "/media/images/"),
		service.NewActivityService(out_adapter.NewActivityFakeOutAdapter()),
	)
}

// TestDeleteClearsThePdfWhateverTheWebPathSpelling pins that the stored
// filename is the name on disk, not something carved back out of the web path.
// A web_path without a trailing slash is a spelling the webstore adapter's own
// docblock declares supported, and under it the derived name matched no file,
// so a delete reported success and left the pdf behind.
func TestDeleteClearsThePdfWhateverTheWebPathSpelling(t *testing.T) {
	for _, webPath := range []string{"/media/images/", "/media/images"} {
		dir := t.TempDir()

		resumes := service.NewResumeService(
			out_adapter.NewResumeFakeOutAdapter(),
			out_adapter.NewMediaWebstoreAdapter(dir+string(filepath.Separator), webPath),
			service.NewActivityService(out_adapter.NewActivityFakeOutAdapter()),
		)

		stored, err := resumes.Create(domain.Resume{Title: "one"}, "application/pdf", []byte("%PDF-fake"))

		if nil != err {
			t.Fatalf("web_path %q: upload failed: %v", webPath, err)
		}

		if _, statErr := os.Stat(filepath.Join(dir, stored.Filename)); nil != statErr {
			t.Fatalf("web_path %q: the stored filename must name the file on disk: %v", webPath, statErr)
		}

		if err := resumes.Delete(stored.Id); nil != err {
			t.Fatalf("web_path %q: delete failed: %v", webPath, err)
		}

		entries, _ := os.ReadDir(dir)

		if 0 != len(entries) {
			t.Fatalf("web_path %q: delete reported success and left %d file(s) on disk", webPath, len(entries))
		}
	}
}

// TestDeleteRefusesARecordThatCannotNameItsFile pins the other way a delete
// could report success over an orphan: an empty name resolves to the media
// directory itself, and RemoveNamed would report whatever it did to that as a
// clean delete.
func TestDeleteRefusesARecordThatCannotNameItsFile(t *testing.T) {
	repo := out_adapter.NewResumeFakeOutAdapter()
	id, _ := repo.Add(domain.Resume{Title: "nameless"})
	dir := t.TempDir()

	resumes := service.NewResumeService(
		repo,
		out_adapter.NewMediaWebstoreAdapter(dir+string(filepath.Separator), "/media/images/"),
		service.NewActivityService(out_adapter.NewActivityFakeOutAdapter()),
	)

	if err := resumes.Delete(id); nil == err {
		t.Fatalf("expected a record with no filename refused rather than cleared")
	}

	if _, statErr := os.Stat(dir); nil != statErr {
		t.Fatalf("the media directory itself must survive: %v", statErr)
	}

	if "" == resumes.Read(id).Id {
		t.Fatalf("a refused delete must leave the record in place")
	}
}

// TestDeleteClearsThePdfBeforeTheRecord pins the order Delete's own docblock
// leans on, the way TestPublishClearsBeforeItSets does for the publish
// transition. Both orders end up somewhere defensible, so only a failure
// between the two writes tells them apart: with the file first, a delete that
// cannot reach the pdf leaves the record in place and stays retryable, where
// clearing the record first would strand the pdf with nothing pointing at it.
// The permission failure reaching the caller as itself is ruling 7 on this
// path: the admin renders what the API returns, and "could not delete" would
// hide a chmod the keeper has to go and undo.
func TestDeleteClearsThePdfBeforeTheRecord(t *testing.T) {
	if 0 == os.Geteuid() {
		t.Skip("root ignores the directory mode this test denies the removal with")
	}

	resumes, _, dir := newShelf(t)

	stored, err := resumes.Create(domain.Resume{Title: "one"}, "application/pdf", []byte("%PDF-fake"))

	if nil != err {
		t.Fatalf("upload failed: %v", err)
	}

	// a removal needs write on the directory, not on the file
	if chmodErr := os.Chmod(dir, 0500); nil != chmodErr {
		t.Fatalf("could not make the media directory unwritable: %v", chmodErr)
	}

	t.Cleanup(func() {
		os.Chmod(dir, 0700) // TempDir cannot clean up behind itself otherwise
	})

	deleteErr := resumes.Delete(stored.Id)

	if nil == deleteErr {
		t.Fatalf("expected the delete to fail while the pdf cannot be removed")
	}

	if !errors.Is(deleteErr, fs.ErrPermission) {
		t.Fatalf("the permission failure must reach the caller as itself, got %v", deleteErr)
	}

	if "" == resumes.Read(stored.Id).Id {
		t.Fatalf("the record must survive a delete that could not clear the pdf")
	}

	if _, statErr := os.Stat(filepath.Join(dir, stored.Filename)); nil != statErr {
		t.Fatalf("the pdf must still be on disk: %v", statErr)
	}
}

// TestPublishClearsBeforeItSets pins the write order the publish docblock leans
// on. The finished state reads the same either way round, so only the order of
// the writes shows it: clearing first means a failure in between leaves nothing
// published, which the hoist refuses loudly, instead of two cuts claiming the
// shelf in silence.
func TestPublishClearsBeforeItSets(t *testing.T) {
	repo := out_adapter.NewResumeFakeOutAdapter()
	first, _ := repo.Add(domain.Resume{Title: "one", Published: true})
	second, _ := repo.Add(domain.Resume{Title: "two"})

	resumes := service.NewResumeService(
		repo,
		out_adapter.NewMediaWebstoreAdapter("", "/media/images/"),
		service.NewActivityService(out_adapter.NewActivityFakeOutAdapter()),
	)

	if _, err := resumes.Publish(second); nil != err {
		t.Fatalf("publish failed: %v", err)
	}

	expected := []string{first + "=false", second + "=true"}

	if len(expected) != len(repo.Writes) {
		t.Fatalf("expected exactly the clear then the set, got %+v", repo.Writes)
	}

	for i, write := range expected {
		if write != repo.Writes[i] {
			t.Fatalf("expected writes %+v, got %+v", expected, repo.Writes)
		}
	}
}

func TestCreateResumeStoresThePdfUnderAGeneratedName(t *testing.T) {
	resumes, activity, dir := newShelf(t)

	saved, err := resumes.Create(domain.Resume{Title: "systems architect", Notes: "the infra-heavy cut"}, "application/pdf", []byte("%PDF-fake"))

	if nil != err {
		t.Fatalf("upload failed: %v", err)
	}

	if "" == saved.Id || "systems architect" != saved.Title || "the infra-heavy cut" != saved.Notes {
		t.Fatalf("unexpected resume entity: %+v", saved)
	}

	if saved.Published {
		t.Fatalf("a freshly stored cut must not be published: %+v", saved)
	}

	if !strings.HasSuffix(saved.Filename, ".pdf") || path.Base(saved.URL) != saved.Filename {
		t.Fatalf("expected a generated .pdf name matching the url, got %q / %q", saved.Filename, saved.URL)
	}

	bytes, readErr := os.ReadFile(filepath.Join(dir, saved.Filename))

	if nil != readErr || "%PDF-fake" != string(bytes) {
		t.Fatalf("expected the pdf on disk, got %q / %v", bytes, readErr)
	}

	entries, _ := activity.Recent(10)

	if 1 != len(entries) {
		t.Fatalf("expected a stored entry in the keeper's log, got %+v", entries)
	}
}

func TestCreateResumeRejectsAnythingButPdf(t *testing.T) {
	resumes, _, dir := newShelf(t)

	// the shelf's policy is its own; the darkroom's image set has no say here
	for _, contentType := range []string{"image/png", "image/jpeg", "image/svg+xml", "text/html", ""} {
		if _, err := resumes.Create(domain.Resume{Title: "cut"}, contentType, []byte("nope")); nil == err {
			t.Fatalf("expected content type %q rejected", contentType)
		}
	}

	entries, _ := os.ReadDir(dir)

	if 0 != len(entries) {
		t.Fatalf("a rejected upload must leave nothing on disk, found %d files", len(entries))
	}

	listed, _ := resumes.List()

	if 0 != len(listed) {
		t.Fatalf("a rejected upload must leave no record, found %d cuts", len(listed))
	}
}

func TestCreateResumeRequiresATitle(t *testing.T) {
	resumes, _, dir := newShelf(t)

	if _, err := resumes.Create(domain.Resume{Title: "   "}, "application/pdf", []byte("%PDF-fake")); nil == err {
		t.Fatalf("expected an untitled cut rejected")
	}

	entries, _ := os.ReadDir(dir)

	if 0 != len(entries) {
		t.Fatalf("a rejected upload must leave nothing on disk, found %d files", len(entries))
	}
}

func TestStoringTwoResumesNeverOverwritesTheFirst(t *testing.T) {
	resumes, _, dir := newShelf(t)

	first, _ := resumes.Create(domain.Resume{Title: "one"}, "application/pdf", []byte("first"))
	second, _ := resumes.Create(domain.Resume{Title: "two"}, "application/pdf", []byte("second"))

	if first.Filename == second.Filename {
		t.Fatalf("two uploads landed on the same name: %q", first.Filename)
	}

	entries, _ := os.ReadDir(dir)

	if 2 != len(entries) {
		t.Fatalf("expected both pdfs kept on disk, found %d files", len(entries))
	}

	bytes, _ := os.ReadFile(filepath.Join(dir, first.Filename))

	if "first" != string(bytes) {
		t.Fatalf("the first pdf was overwritten, reads %q", bytes)
	}
}

func TestPublishingClearsWhicheverHeldIt(t *testing.T) {
	resumes, _, _ := newShelf(t)

	first, _ := resumes.Create(domain.Resume{Title: "one"}, "application/pdf", []byte("first"))
	second, _ := resumes.Create(domain.Resume{Title: "two"}, "application/pdf", []byte("second"))

	if _, err := resumes.Publish(first.Id); nil != err {
		t.Fatalf("first publish failed: %v", err)
	}

	published, err := resumes.Publish(second.Id)

	if nil != err {
		t.Fatalf("second publish failed: %v", err)
	}

	if !published.Published || second.Id != published.Id {
		t.Fatalf("expected the second cut live, got %+v", published)
	}

	listed, _ := resumes.List()
	live := 0

	for _, resume := range listed {
		if resume.Published {
			live++
		}
	}

	if 1 != live {
		t.Fatalf("exactly one cut may be published, found %d", live)
	}

	if resumes.Read(first.Id).Published {
		t.Fatalf("the previously published cut must have been cleared")
	}
}

func TestPublishedIsEmptyUntilSomethingIsPublished(t *testing.T) {
	resumes, _, _ := newShelf(t)

	resumes.Create(domain.Resume{Title: "one"}, "application/pdf", []byte("first"))

	published, err := resumes.Published()

	if nil != err {
		t.Fatalf("published lookup failed: %v", err)
	}

	if "" != published.Id {
		t.Fatalf("a stored cut is a draft until it is published, got %+v", published)
	}
}

func TestPublishRejectsAnUnknownId(t *testing.T) {
	resumes, _, _ := newShelf(t)

	if _, err := resumes.Publish("nope"); nil == err {
		t.Fatalf("expected publish to reject an unknown resume id")
	}
}

func TestUpdateEditsTheTitleAndNotesAndLeavesThePdfAlone(t *testing.T) {
	resumes, _, _ := newShelf(t)

	stored, _ := resumes.Create(domain.Resume{Title: "one", Notes: "first pass"}, "application/pdf", []byte("first"))
	resumes.Publish(stored.Id)

	saved, err := resumes.Update(domain.Resume{Id: stored.Id, Title: "one, sharpened", Notes: "second pass"})

	if nil != err {
		t.Fatalf("update failed: %v", err)
	}

	if "one, sharpened" != saved.Title || "second pass" != saved.Notes {
		t.Fatalf("expected the title and notes edited, got %+v", saved)
	}

	if stored.Filename != saved.Filename || stored.URL != saved.URL {
		t.Fatalf("the pdf is immutable; expected %q/%q, got %q/%q", stored.Filename, stored.URL, saved.Filename, saved.URL)
	}

	if !saved.Published {
		t.Fatalf("an edit must not take the live cut off the shelf: %+v", saved)
	}
}

func TestUpdateRequiresATitle(t *testing.T) {
	resumes, _, _ := newShelf(t)

	stored, _ := resumes.Create(domain.Resume{Title: "one"}, "application/pdf", []byte("first"))

	if _, err := resumes.Update(domain.Resume{Id: stored.Id, Title: ""}); nil == err {
		t.Fatalf("expected an untitled edit rejected")
	}
}

func TestDeleteRemovesTheRecordAndThePdf(t *testing.T) {
	resumes, _, dir := newShelf(t)

	stored, _ := resumes.Create(domain.Resume{Title: "one"}, "application/pdf", []byte("first"))

	if err := resumes.Delete(stored.Id); nil != err {
		t.Fatalf("delete failed: %v", err)
	}

	listed, _ := resumes.List()

	if 0 != len(listed) {
		t.Fatalf("expected the record gone, got %+v", listed)
	}

	if _, statErr := os.Stat(filepath.Join(dir, stored.Filename)); !os.IsNotExist(statErr) {
		t.Fatalf("expected the pdf gone from disk, got %v", statErr)
	}
}

func TestDeleteRejectsAnUnknownId(t *testing.T) {
	resumes, _, _ := newShelf(t)

	if err := resumes.Delete("nope"); nil == err {
		t.Fatalf("expected delete to reject an unknown resume id")
	}
}

func TestListResumesIsNewestFirst(t *testing.T) {
	resumes, _, _ := newShelf(t)

	resumes.Create(domain.Resume{Title: "first"}, "application/pdf", []byte("1"))
	resumes.Create(domain.Resume{Title: "second"}, "application/pdf", []byte("2"))

	listed, err := resumes.List()

	if nil != err {
		t.Fatalf("list failed: %v", err)
	}

	if 2 != len(listed) || "second" != listed[0].Title || "first" != listed[1].Title {
		t.Fatalf("expected newest first, got %+v", listed)
	}
}

// TestDeleteRefusesThePublishedCut pins the refusal that keeps a published
// record from outliving its own pdf. Delete clears the file before the record,
// so without this a failure between the two writes would leave the shelf
// claiming a live resume whose file is gone; publishing another cut is what
// frees the old one.
func TestDeleteRefusesThePublishedCut(t *testing.T) {
	resumes, _, dir := newShelf(t)

	first, _ := resumes.Create(domain.Resume{Title: "one"}, "application/pdf", []byte("first"))
	second, _ := resumes.Create(domain.Resume{Title: "two"}, "application/pdf", []byte("second"))

	if _, err := resumes.Publish(first.Id); nil != err {
		t.Fatalf("publish failed: %v", err)
	}

	if err := resumes.Delete(first.Id); !errors.Is(err, in_port.ErrResumePublished) {
		t.Fatalf("expected the published cut refused, got %v", err)
	}

	if "" == resumes.Read(first.Id).Id {
		t.Fatalf("a refused delete must leave the record in place")
	}

	if _, statErr := os.Stat(filepath.Join(dir, first.Filename)); nil != statErr {
		t.Fatalf("a refused delete must leave the pdf on disk: %v", statErr)
	}

	if _, err := resumes.Publish(second.Id); nil != err {
		t.Fatalf("second publish failed: %v", err)
	}

	if err := resumes.Delete(first.Id); nil != err {
		t.Fatalf("expected the superseded cut deletable, got %v", err)
	}
}

// TestDeleteRefusesThePublishedCutBeforeItReachesTheDisk pins the order of the
// two guards: a published record that cannot name its file is refused as
// published, which is the thing the keeper can act on, not as a record to go
// and clear by hand.
func TestDeleteRefusesThePublishedCutBeforeItReachesTheDisk(t *testing.T) {
	repo := out_adapter.NewResumeFakeOutAdapter()
	id, _ := repo.Add(domain.Resume{Title: "nameless but live", Published: true})

	resumes := service.NewResumeService(
		repo,
		out_adapter.NewMediaWebstoreAdapter(t.TempDir()+string(filepath.Separator), "/media/images/"),
		service.NewActivityService(out_adapter.NewActivityFakeOutAdapter()),
	)

	if err := resumes.Delete(id); !errors.Is(err, in_port.ErrResumePublished) {
		t.Fatalf("expected the published refusal ahead of the filename guard, got %v", err)
	}
}

// TestUnpublishTakesTheLiveCutDownAndFreesIt pins the transition the delete
// refusal presupposes: down off the hoist with nothing put up in its place, and
// the cut then deletable. Publishing something else frees a cut too, but it is
// not the same act, and a shelf holding one cut has no something else.
func TestUnpublishTakesTheLiveCutDownAndFreesIt(t *testing.T) {
	resumes, activity, dir := newShelf(t)

	only, _ := resumes.Create(domain.Resume{Title: "the only cut"}, "application/pdf", []byte("only"))
	resumes.Publish(only.Id)

	saved, err := resumes.Unpublish(only.Id)

	if nil != err {
		t.Fatalf("unpublish failed: %v", err)
	}

	if saved.Published {
		t.Fatalf("expected the cut down off the hoist, got %+v", saved)
	}

	published, _ := resumes.Published()

	if "" != published.Id {
		t.Fatalf("expected nothing published, got %+v", published)
	}

	entries, _ := activity.Recent(10)

	if 3 != len(entries) {
		t.Fatalf("expected the store, the publish and the unpublish logged, got %+v", entries)
	}

	if err := resumes.Delete(only.Id); nil != err {
		t.Fatalf("expected the unpublished cut deletable, got %v", err)
	}

	if remaining, _ := os.ReadDir(dir); 0 != len(remaining) {
		t.Fatalf("expected the pdf gone with the record, found %d file(s)", len(remaining))
	}
}

// TestUnpublishLogsACutAlreadyDown pins that a cut already down is written and
// logged like any other. The cuts on the shelf are near-identical papers, so
// naming the wrong one is the likely slip rather than a typo'd id, and a silent
// success would leave the keeper nothing to notice it by.
func TestUnpublishLogsACutAlreadyDown(t *testing.T) {
	repo := out_adapter.NewResumeFakeOutAdapter()
	id, _ := repo.Add(domain.Resume{Title: "a draft", UpdatedAt: "2026-09-01T00:00:00Z"})
	activity := service.NewActivityService(out_adapter.NewActivityFakeOutAdapter())

	resumes := service.NewResumeService(
		repo,
		out_adapter.NewMediaWebstoreAdapter(t.TempDir()+string(filepath.Separator), "/media/images/"),
		activity,
	)

	saved, err := resumes.Unpublish(id)

	if nil != err || saved.Published || id != saved.Id {
		t.Fatalf("expected the named draft handed back down, got %+v / %v", saved, err)
	}

	if 1 != len(repo.Writes) || id+"=false" != repo.Writes[0] {
		t.Fatalf("expected the one clear %q, got %+v", id+"=false", repo.Writes)
	}

	if "2026-09-01T00:00:00Z" == resumes.Read(id).UpdatedAt {
		t.Fatalf("the write must move the stamp, still reads %q", resumes.Read(id).UpdatedAt)
	}

	entries, _ := activity.Recent(10)

	if 1 != len(entries) {
		t.Fatalf("expected the unpublish in the keeper's log, got %+v", entries)
	}
}

// TestUnpublishWritesOnlyTheCutItNames pins that taking one cut down is not a
// shelf-wide sweep. The finished state cannot tell the two apart, because only
// one cut is ever up, so this reads the writes: a sweep clearing every record
// leaves the same hoist empty and the same Published answer, and shows up only
// as the draft beside it losing its stamp.
func TestUnpublishWritesOnlyTheCutItNames(t *testing.T) {
	repo := out_adapter.NewResumeFakeOutAdapter()
	live, _ := repo.Add(domain.Resume{Title: "live", Published: true, Filename: "live.pdf"})
	draft, _ := repo.Add(domain.Resume{Title: "draft", UpdatedAt: "2026-09-01T00:00:00Z"})

	// a real save path, because the record names a file: an empty one resolves
	// the stored name against the working directory, and a delete is all it
	// would take to reach it
	resumes := service.NewResumeService(
		repo,
		out_adapter.NewMediaWebstoreAdapter(t.TempDir()+string(filepath.Separator), "/media/images/"),
		service.NewActivityService(out_adapter.NewActivityFakeOutAdapter()),
	)

	saved, err := resumes.Unpublish(live)

	if nil != err || live != saved.Id {
		t.Fatalf("expected the named cut handed back, got %+v / %v", saved, err)
	}

	if 1 != len(repo.Writes) || live+"=false" != repo.Writes[0] {
		t.Fatalf("expected exactly the one clear %q, got %+v", live+"=false", repo.Writes)
	}

	if "2026-09-01T00:00:00Z" != resumes.Read(draft).UpdatedAt {
		t.Fatalf("the draft beside it must be untouched, got %q", resumes.Read(draft).UpdatedAt)
	}
}

func TestUnpublishRejectsAnUnknownId(t *testing.T) {
	resumes, _, _ := newShelf(t)

	if _, err := resumes.Unpublish("nope"); nil == err {
		t.Fatalf("expected unpublish to reject an unknown resume id")
	}
}

// TestTheDeleteRefusalPointsAtUnpublishing pins the message, not just the
// refusal. It is the only instruction the keeper gets when a delete bounces,
// and it named publishing another cut back when that was the only way down.
func TestTheDeleteRefusalPointsAtUnpublishing(t *testing.T) {
	if !strings.Contains(in_port.ErrResumePublished.Error(), "unpublish") {
		t.Fatalf("the refusal must send the keeper to unpublish, reads %q", in_port.ErrResumePublished.Error())
	}
}
