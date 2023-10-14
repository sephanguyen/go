package errcode

import (
	"fmt"
)

const (
	InternalError      = 50000
	BadRequest         = 40000
	MissingMandatory   = 40001
	DataExist          = 40002
	DuplicatedData     = 40003
	InvalidData        = 40004
	InvalidPayloadSize = 40005
	InvalidMaximumRows = 40006
	MissingField       = 40007
	UpdateFieldFail    = 40008
	RemoveLocationFail = 40009
	PermissionDenied   = 40300
	InvalidSignature   = 40301
	InvalidPublicKey   = 40302
	NotFound           = 40400
)

// DomainError represent a domain error that contains specific error information
// relates to business logic and it error handling standard.
// Please follow to the new standardize error handling: https://manabie.atlassian.net/browse/LT-40632

type DomainError interface {
	error
	DomainError() string
	DomainCode() int
}

type HasErrCode interface {
	ErrCode() int
}

type Error struct {
	Code int
	Err  error

	UserID    string
	Resource  string
	FieldName string
	Message   string

	// to identify the student who upsert failed
	Index int
}

func (err Error) Error() string {
	switch err.Code {
	case MissingMandatory:
		return fmt.Sprintf(`%s is mandatory`, err.FieldName)
	case DataExist:
		return fmt.Sprintf(`%s is existed in system`, err.FieldName)
	case DuplicatedData:
		return fmt.Sprintf(`%s is duplicated in payload`, err.FieldName)
	case InvalidData:
		return fmt.Sprintf(`%s is invalid`, err.FieldName)
	case InvalidPayloadSize:
		return fmt.Sprintf(`payload size is bigger than allowed`)
	case PermissionDenied:
		return fmt.Sprintf(`permission denied`)
	case InvalidSignature:
		return fmt.Sprintf(`invalid signature`)
	case InvalidPublicKey:
		return fmt.Sprintf(`invalid public key`)
	case NotFound:
		return fmt.Sprintf(`%s not found`, err.Resource)
	case InternalError:
		return fmt.Sprintf(`internal error`)
	case InvalidMaximumRows:
		return `Invalid number of row. The maximum number of rows is 1000`
	case MissingField:
		return fmt.Sprintf(`%s does not exist`, err.FieldName)
	case UpdateFieldFail:
		return fmt.Sprintf(`%s cannot be updated`, err.FieldName)
	case RemoveLocationFail:
		return `Unable to remove location of the active student course`
	default:
		// if we forget assign somewhere, app will be crash easily
		// for safety, we must check here
		if err.Err != nil {
			return err.Err.Error()
		}
	}
	return ""
}
