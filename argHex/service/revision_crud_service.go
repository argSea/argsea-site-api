package service

import (
	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/argSea/argsea-site-api/argHex/out_port"
)

type revisionService struct {
	repo out_port.RevisionRepo
}

func NewRevisionService(repo out_port.RevisionRepo) in_port.RevisionService {
	return revisionService{
		repo: repo,
	}
}

// Snapshot appends a new current revision, then clears the current flag on the
// entity's other revisions. Insert-first ordering means a failure between the
// two steps leaves two current revisions — which self-heals on the next
// snapshot — never zero.
func (r revisionService) Snapshot(entityType string, entityID string, snapshot string, summary string) (string, error) {
	rev := domain.Revision{
		EntityType: entityType,
		EntityId:   entityID,
		Snapshot:   snapshot,
		Summary:    summary,
		IsCurrent:  true,
		CreatedAt:  nowStamp(),
	}

	id, err := r.repo.Add(rev)

	if nil != err {
		return "", err
	}

	if err := r.repo.ClearCurrentExcept(entityID, id); nil != err {
		return "", err
	}

	return id, nil
}

func (r revisionService) List(entityType string, entityID string, limit int64) (domain.Revisions, error) {
	return r.repo.List(entityID, limit)
}

func (r revisionService) Get(id string) domain.Revision {
	return r.repo.Get(id)
}
