package service

import (
	"log"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/argSea/argsea-site-api/argHex/out_port"
)

// maxBearings caps the TL;DR strip: the homepage renders at most three lines,
// so anything past that would be stored and never shown.
const maxBearings = 3

type watchService struct {
	repo     out_port.WatchRepo
	activity in_port.ActivityService
}

func NewWatchService(repo out_port.WatchRepo, activity in_port.ActivityService) in_port.WatchService {
	return watchService{
		repo:     repo,
		activity: activity,
	}
}

func (s watchService) Get() domain.Watch {
	return s.repo.Get()
}

// Save upserts the singleton. KeptAt is stamped here so a client can never
// backdate or forge the watch; whatever value came over the wire is discarded.
// An empty Letter is a valid write (the cleared watch), not an error.
func (s watchService) Save(watch domain.Watch) (domain.Watch, error) {
	watch.KeptAt = nowStamp()

	if maxBearings < len(watch.Bearings) {
		watch.Bearings = watch.Bearings[:maxBearings]
	}

	saved, err := s.repo.Save(watch)

	if nil != err {
		return domain.Watch{}, err
	}

	if err := s.activity.Record("current watch updated", domain.EntityWatch, saved.Id); nil != err {
		log.Printf("activity record failed for watch %v: %v\n", saved.Id, err)
	}

	return saved, nil
}
