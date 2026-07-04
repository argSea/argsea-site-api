package service

import (
	"log"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/argSea/argsea-site-api/argHex/out_port"
)

type siteCopyService struct {
	repo     out_port.SiteCopyRepo
	activity in_port.ActivityService
}

func NewSiteCopyService(repo out_port.SiteCopyRepo, activity in_port.ActivityService) in_port.SiteCopyService {
	return siteCopyService{
		repo:     repo,
		activity: activity,
	}
}

func (s siteCopyService) Get() domain.SiteCopy {
	return s.repo.Get()
}

// Save upserts the singleton. SiteCopy is plain text throughout, so no HTML
// sanitizing applies.
func (s siteCopyService) Save(copy domain.SiteCopy) (domain.SiteCopy, error) {
	copy.UpdatedAt = nowStamp()

	saved, err := s.repo.Save(copy)

	if nil != err {
		return domain.SiteCopy{}, err
	}

	if err := s.activity.Record("signal flags updated", domain.EntityCopy, saved.Id); nil != err {
		log.Printf("activity record failed for site copy %v: %v\n", saved.Id, err)
	}

	return saved, nil
}
