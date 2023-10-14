package errcode

import (
	"fmt"
)

const (
	InternalError    = 50000
	BadRequest       = 40000
	MissingMandatory = 40001
	InvalidData      = 40004
	PermissionDenied = 40300
	InvalidSignature = 40301
	InvalidPublicKey = 40302
	NotFound         = 40400
)

type HasErrCode interface {
	ErrCode() int
}

type Error struct {
	Code      int
	Err       error
	UserID    string
	Resource  string
	FieldName string
	Message   string
}

func (err Error) Error() string {
	switch err.Code {
	case MissingMandatory:
		return fmt.Sprintf(`%s is mandatory`, err.FieldName)
	case InvalidData:
		return fmt.Sprintf(`%s is invalid`, err.FieldName)
	case PermissionDenied:
		return "permission denied"
	case InvalidSignature:
		return "invalid signature"
	case InvalidPublicKey:
		return "invalid public key"
	case NotFound:
		return fmt.Sprintf(`%s not found`, err.Resource)
	case InternalError:
		return "internal error"
	default:
		if err.Err != nil {
			return err.Err.Error()
		}
	}

	return ""
}
