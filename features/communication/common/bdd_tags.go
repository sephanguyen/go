package common

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/manabie-com/backend/features/communication/common/entities"
	"github.com/manabie-com/backend/internal/golibs/idutil"

	"github.com/pkg/errors"
)

func (s *NotificationSuite) CreatesNumberOfTags(ctx context.Context, num string, tagKeyword string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.Organization == nil || len(stepState.Organization.Staffs) == 0 {
		return ctx, errors.New("missing created organization and staff with granted role step")
	}

	numTags := 0
	if num == "random" {
		numTags = RandRangeIn(5, 10)
	} else {
		var err error
		numTags, err = strconv.Atoi(num)
		if err != nil {
			return ctx, fmt.Errorf("s.CreatesNumberOfTags: %v", err)
		}
	}
	tags := []*entities.Tag{}
	for i := 0; i < numTags; i++ {
		tag := &entities.Tag{
			ID:         idutil.ULIDNow(),
			Name:       idutil.ULIDNow(),
			IsArchived: false,
		}
		if tagKeyword != "random" {
			tag.Name = "tag-" + tagKeyword + "-" + idutil.ULIDNow()
		}
		tags = append(tags, tag)
	}

	err := s.CreateTags(stepState.Organization.Staffs[0], tags)
	if err != nil {
		return ctx, fmt.Errorf("s.CreatesNumberOfTags: %v", err)
	}

	stepState.Tags = append(stepState.Tags, tags...)

	return StepStateToContext(ctx, stepState), nil
}

func (s *NotificationSuite) CreatesTagsWithNames(ctx context.Context, tagNames string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.Organization == nil || len(stepState.Organization.Staffs) == 0 {
		return ctx, errors.New("missing created organization and staff with granted role step")
	}
	names := strings.Split(tagNames, ",")
	tags := []*entities.Tag{}
	for i := 0; i < len(names); i++ {
		tag := &entities.Tag{
			ID:         idutil.ULIDNow(),
			Name:       names[i],
			IsArchived: false,
		}
		tags = append(tags, tag)
	}

	err := s.CreateTags(stepState.Organization.Staffs[0], tags)
	if err != nil {
		return ctx, fmt.Errorf("s.CreatesNumberOfTags: %v", err)
	}

	stepState.Tags = append(stepState.Tags, tags...)

	return StepStateToContext(ctx, stepState), nil
}
