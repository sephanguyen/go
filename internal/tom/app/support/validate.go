package support

import (
	"fmt"

	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"
)

func validateParentRemovedFromStudent(req *upb.EvtUser_ParentRemovedFromStudent) error {
	if req.GetParentId() == "" {
		return errEmptyParent
	}
	if req.GetStudentId() == "" {
		return errEmptyStudent
	}
	return nil
}

var (
	errEmptyParent  = fmt.Errorf("empty parent id")
	errEmptyStudent = fmt.Errorf("empty student id")
)
