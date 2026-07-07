package out_port

import "github.com/argSea/argsea-site-api/argHex/domain"

type CatDesignRepo interface {
	List() (domain.CatDesigns, error)
	Get(id string) domain.CatDesign
	Add(design domain.CatDesign) (string, error)
	Set(design domain.CatDesign) error
	Remove(id string) error
}
