package enum

type Tribe int

const (
	TribeAny       Tribe = 0
	TribeRomans    Tribe = 1
	TribeTeutons   Tribe = 2
	TribeGauls     Tribe = 3
	TribeNature    Tribe = 4
	TribeNatars    Tribe = 5
	TribeEgyptians Tribe = 6
	TribeHuns      Tribe = 7
)

func (t Tribe) String() string {
	switch t {
	case TribeAny:
		return "Any"
	case TribeRomans:
		return "Romans"
	case TribeTeutons:
		return "Teutons"
	case TribeGauls:
		return "Gauls"
	case TribeNature:
		return "Nature"
	case TribeNatars:
		return "Natars"
	case TribeEgyptians:
		return "Egyptians"
	case TribeHuns:
		return "Huns"
	default:
		return "Unknown"
	}
}

var AllTribes = []Tribe{
	TribeAny, TribeRomans, TribeTeutons, TribeGauls,
	TribeNature, TribeNatars, TribeEgyptians, TribeHuns,
}
