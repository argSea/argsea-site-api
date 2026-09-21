package service

import (
	"errors"
	"log"
	"sort"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/argSea/argsea-site-api/argHex/out_port"
)

// The shelf takes papers, not photographs. This is the resume destination's own
// slice of the upload policy and it widens nothing in the darkroom: the two run
// through the same chokepoint holding different sets.
var resumeUploads = uploadPolicy{
	types:     map[string]bool{"application/pdf": true},
	rejection: "only pdf uploads are allowed",
	reject:    resumeRejection,
}

func resumeRejection(message string) error {
	return in_port.ResumeValidationError{Message: message}
}

type resumeService struct {
	repo     out_port.ResumeRepo
	files    out_port.MediaRepo
	activity in_port.ActivityService
}

// NewResumeService wires the resume shelf onto its two halves: the PDFs on disk
// behind the same file half the darkroom writes to, the titles and notes in
// mongo, plus the keeper's log.
func NewResumeService(repo out_port.ResumeRepo, files out_port.MediaRepo, activity in_port.ActivityService) in_port.ResumeService {
	return resumeService{
		repo:     repo,
		files:    files,
		activity: activity,
	}
}

// List returns every stored cut newest first. Fixed-width stamps make the
// reverse string sort chronological.
func (r resumeService) List() (domain.Resumes, error) {
	resumes, err := r.repo.List()

	if nil != err {
		return nil, err
	}

	sort.SliceStable(resumes, func(i, j int) bool {
		return resumes[i].CreatedAt > resumes[j].CreatedAt
	})

	return resumes, nil
}

// Read hands one stored cut back so the keeper can open it for comparison.
func (r resumeService) Read(id string) domain.Resume {
	return r.repo.Get(id)
}

// Create stores a new cut: the PDF lands under a generated name, so an upload
// can never collide with a file already on disk and nothing is ever overwritten.
// A new cut is never published; that is its own transition.
func (r resumeService) Create(resume domain.Resume, mime_type string, bytes []byte) (domain.Resume, error) {
	if err := resumeUploads.allow(mime_type); nil != err {
		return domain.Resume{}, err
	}

	if !domain.ValidResumeTitle(resume.Title) {
		return domain.Resume{}, in_port.ResumeValidationError{Message: "a title is required"}
	}

	// the name comes back from the file half rather than being carved out of the
	// web path: RemoveNamed takes the name, and how the adapter joins one onto
	// web_path is its business, not something this side may assume a shape for
	file_name, url, err := r.files.UploadMedia(mime_type, bytes)

	if nil != err {
		return domain.Resume{}, err
	}

	now := nowStamp()

	id, err := r.repo.Add(domain.Resume{
		Title:     resume.Title,
		Notes:     resume.Notes,
		Filename:  file_name,
		URL:       url,
		CreatedAt: now,
		UpdatedAt: now,
	})

	if nil != err {
		// the file half landed but the metadata half didn't; pull the file back
		// so the shelf never holds an orphan
		if removeErr := r.files.RemoveNamed(file_name); nil != removeErr {
			log.Printf("could not remove orphaned resume file %v: %v\n", file_name, removeErr)
		}

		return domain.Resume{}, err
	}

	r.record("resume \""+resume.Title+"\" stored", id)

	return r.repo.Get(id), nil
}

// Update edits the title and notes the keeper picks a cut by. The PDF itself is
// immutable, so the stored file, its URL and whether the cut is published all
// ride through untouched; publishing is its own transition.
func (r resumeService) Update(resume domain.Resume) (domain.Resume, error) {
	existing := r.repo.Get(resume.Id)

	if "" == existing.Id {
		return domain.Resume{}, in_port.ResumeValidationError{Message: "resume not found"}
	}

	if !domain.ValidResumeTitle(resume.Title) {
		return domain.Resume{}, in_port.ResumeValidationError{Message: "a title is required"}
	}

	existing.Title = resume.Title
	existing.Notes = resume.Notes
	existing.UpdatedAt = nowStamp()

	if err := r.repo.Set(existing); nil != err {
		return domain.Resume{}, err
	}

	saved := r.repo.Get(existing.Id)
	r.record("resume \""+saved.Title+"\" edited", saved.Id)

	return saved, nil
}

// Publish makes one cut the live one and clears whichever held it, so "which
// resume is live" is never a question worth going and checking. The clear and
// the set are two writes rather than one: the store has no transaction, and
// clearing first means a failure in between leaves nothing published, which the
// hoist guard refuses loudly, instead of two cuts claiming the shelf in silence.
func (r resumeService) Publish(id string) (domain.Resume, error) {
	target := r.repo.Get(id)

	if "" == target.Id {
		return domain.Resume{}, in_port.ResumeValidationError{Message: "resume not found"}
	}

	resumes, err := r.repo.List()

	if nil != err {
		return domain.Resume{}, err
	}

	for _, resume := range resumes {
		if !resume.Published || resume.Id == target.Id {
			continue
		}

		resume.Published = false
		resume.UpdatedAt = nowStamp()

		if setErr := r.repo.Set(resume); nil != setErr {
			return domain.Resume{}, setErr
		}
	}

	target.Published = true
	target.UpdatedAt = nowStamp()

	if err := r.repo.Set(target); nil != err {
		return domain.Resume{}, err
	}

	saved := r.repo.Get(target.Id)
	r.record("resume \""+saved.Title+"\" published", saved.Id)

	return saved, nil
}

// Published is the one live cut, or a zero Resume when the shelf holds none.
// The hoist asks this before it builds anything.
func (r resumeService) Published() (domain.Resume, error) {
	resumes, err := r.repo.List()

	if nil != err {
		return domain.Resume{}, err
	}

	for _, resume := range resumes {
		if resume.Published {
			return resume, nil
		}
	}

	return domain.Resume{}, nil
}

// Delete removes the PDF and the record behind it, and refuses outright while
// the cut is the published one. The invariant that refusal protects is that a
// published record always has its PDF: the file goes first below, so a delete
// that got past here and then failed clearing the record would leave the shelf
// claiming a live resume whose file is gone, and the hoist guard would pass on
// it. Publish another cut first; that transition is the only way off the hoist.
func (r resumeService) Delete(id string) error {
	resume := r.repo.Get(id)

	if "" == resume.Id {
		return in_port.ResumeValidationError{Message: "resume not found"}
	}

	if resume.Published {
		return in_port.ErrResumePublished
	}

	if "" == resume.Filename {
		// an empty name resolves to the media directory itself, and RemoveNamed
		// would report whatever it did to it as a clean delete; a record that
		// cannot name its file is one to stop on, not guess behind
		return errors.New("resume record carries no filename; its pdf has to be cleared by hand")
	}

	// the file goes first: a failure here leaves the record intact and the
	// delete retryable, where clearing the record first strands the pdf on disk
	// with nothing left pointing at it
	if err := r.files.RemoveNamed(resume.Filename); nil != err {
		return err
	}

	if err := r.repo.Remove(id); nil != err {
		return err
	}

	r.record("resume \""+resume.Title+"\" removed", id)

	return nil
}

func (r resumeService) record(message string, id string) {
	if err := r.activity.Record(message, domain.EntityResume, id); nil != err {
		log.Printf("activity record failed for resume %v: %v\n", id, err)
	}
}
