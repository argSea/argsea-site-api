package out_port

// LanternStateRepo persists the lantern's lastHoistedAt stamp so it survives a
// restart. Empty string means no hoist has ever succeeded.
type LanternStateRepo interface {
	LastHoistedAt() (string, error)
	SaveLastHoistedAt(stamp string) error
}
