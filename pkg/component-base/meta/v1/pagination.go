package v1

// DefaultLimit define the default number of records to be retrieved.
const DefaultLimit = 1000

// LimitAndOffset contains offset and limit fields.
type LimitAndOffset struct {
	Offset int
	Limit  int
}

// Unpointer fill LimitAndOffset with default values if offset/limit is nil
// or it will be filled with the passed value.
func Unpointer(opt ListOptions) *LimitAndOffset {
	var o, l int = 0, DefaultLimit

	if opt.PageNo != nil || opt.PageSize != nil {
		if opt.PageSize != nil {
			l = int(*opt.PageSize)
		}

		if opt.PageNo != nil {
			o = (int(*opt.PageNo) - 1) * l
		}
	} else {
		if opt.Offset != nil {
			o = int(*opt.Offset)
		}

		if opt.Limit != nil {
			l = int(*opt.Limit)
		}
	}

	return &LimitAndOffset{
		Offset: o,
		Limit:  l,
	}
}
