package enums

import (
	"fmt"
	"io"
	"strings"
)

// ApprovalStatus is a custom type for document status
type ApprovalStatus string

var (
	// ApprovalPending indicates that the document is pending approval
	ApprovalPending ApprovalStatus = "PENDING"
	// ApprovalApproved indicates that the document has been approved
	ApprovalApproved ApprovalStatus = "APPROVED"
	// ApprovalRejected indicates that the document has been rejected
	ApprovalRejected ApprovalStatus = "REJECTED"
	// ApprovalNeedsReview indicates that the document needs review
	ApprovalStatusInvalid ApprovalStatus = "DOCUMENT_STATUS_INVALID"
)

// Values returns a slice of strings that represents all the possible values of the ApprovalStatus enum.
// Possible default values are "PENDING", "APPROVED", and "REJECTED"
func (ApprovalStatus) Values() (kinds []string) {
	for _, s := range []ApprovalStatus{ApprovalApproved, ApprovalPending, ApprovalRejected} {
		kinds = append(kinds, string(s))
	}

	return
}

// String returns the document status as a string
func (r ApprovalStatus) String() string {
	return string(r)
}

// ToApprovalStatus returns the document status enum based on string input
func ToApprovalStatus(r string) *ApprovalStatus {
	switch r := strings.ToUpper(r); r {
	case ApprovalApproved.String():
		return &ApprovalApproved
	case ApprovalPending.String():
		return &ApprovalPending
	case ApprovalRejected.String():
		return &ApprovalRejected
	case ApprovalStatusInvalid.String():
		return &ApprovalStatusInvalid
	default:
		return nil
	}
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r ApprovalStatus) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *ApprovalStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for ApprovalStatus, got: %T", v) //nolint:err113
	}

	*r = ApprovalStatus(str)

	return nil
}
