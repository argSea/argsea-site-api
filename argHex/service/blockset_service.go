package service

import (
	"errors"
	"log"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/argSea/argsea-site-api/argHex/out_port"
)

type blockSetService struct {
	repo     out_port.BlockSetRepo
	activity in_port.ActivityService
}

func NewBlockSetService(repo out_port.BlockSetRepo, activity in_port.ActivityService) in_port.BlockSetService {
	return blockSetService{
		repo:     repo,
		activity: activity,
	}
}

func (b blockSetService) List() (domain.BlockSets, error) {
	return b.repo.List()
}

// Create stores a new set verbatim. Blocks are stored as handed in, the same
// trust model as a caselog's blocks; the API does not interpret them.
func (b blockSetService) Create(set domain.BlockSet) (domain.BlockSet, error) {
	set.Id = ""

	id, err := b.repo.Add(set)

	if nil != err {
		return domain.BlockSet{}, err
	}

	saved := b.repo.Get(id)

	b.record("block set \""+saved.Name+"\" created", saved.Id)

	return saved, nil
}

func (b blockSetService) Delete(id string) error {
	existing := b.repo.Get(id)

	if "" == existing.Id {
		return errors.New("block set not found")
	}

	if err := b.repo.Remove(id); nil != err {
		return err
	}

	b.record("block set \""+existing.Name+"\" deleted", id)

	return nil
}

// Seed plants the one header set into an empty collection at boot. Anything
// already in the collection means a keeper has been here; the seed never runs
// twice and never touches existing sets.
func (b blockSetService) Seed() error {
	existing, err := b.repo.List()

	if nil != err {
		return err
	}

	if 0 != len(existing) {
		return nil
	}

	set := seedHeaderBlockSet()

	id, err := b.repo.Add(set)

	if nil != err {
		return err
	}

	b.record("block set \""+set.Name+"\" seeded", id)

	return nil
}

func (b blockSetService) record(message string, id string) {
	if err := b.activity.Record(message, domain.EntityBlockSet, id); nil != err {
		log.Printf("activity record failed for block set %v: %v\n", id, err)
	}
}
