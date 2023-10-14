package utils

import (
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/jackc/pgtype"
)

func GetNextPaging(limit, offset pgtype.Int8) *cpb.Paging {
	return &cpb.Paging{
		Limit: uint32(limit.Int),
		Offset: &cpb.Paging_OffsetInteger{
			OffsetInteger: limit.Int + offset.Int,
		},
	}
}

func GetPrevPaging(limit, offset pgtype.Int8) *cpb.Paging {
	preOffset := offset.Int - limit.Int

	if preOffset < 0 {
		preOffset = 0
	}

	return &cpb.Paging{
		Limit: uint32(limit.Int),
		Offset: &cpb.Paging_OffsetInteger{
			OffsetInteger: preOffset,
		},
	}
}
