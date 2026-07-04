package service

import (
	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/argSea/argsea-site-api/argHex/out_port"
)

type activityService struct {
	repo out_port.ActivityRepo
}

func NewActivityService(repo out_port.ActivityRepo) in_port.ActivityService {
	return activityService{
		repo: repo,
	}
}

func (a activityService) Record(message string, entityType string, entityID string) error {
	entry := domain.ActivityLog{
		Timestamp:  nowStamp(),
		Message:    message,
		EntityType: entityType,
		EntityId:   entityID,
	}

	_, err := a.repo.Add(entry)

	return err
}

func (a activityService) Recent(limit int64) (domain.ActivityLogs, error) {
	return a.repo.Recent(limit)
}
