package compmgr

import "context"

// Provider defines the methods a cloud compliance provider must implement.
type Provider interface {
	// ListReports returns the available compliance reports from the provider.
	ListReports(ctx context.Context) ([]Report, error)
}

// Report represents a generic compliance report returned from a provider.
// Control represents a mapping between a compliance framework and the specific
// control identifiers referenced by a provider report.
type Control struct {
	// Standard is the name of the compliance framework, e.g. "SOC2" or
	// "ISO27001".
	Standard string
	// Version of the standard if available.
	Version string
	// IDs contains the control identifiers within the framework that the
	// report is associated with.
	IDs []string
}

// Report represents a generic compliance report returned from a provider.
type Report struct {
	ID          string
	Name        string
	Description string
	Link        string
	Passed      bool
	// Controls lists the specific compliance controls that this report
	// relates to, grouped by framework.
	Controls []Control
	// ControlIDs is a flat list of control identifiers derived from the
	// Controls slice. It is kept for convenience when sending data to
	// systems that only expect IDs.
	ControlIDs []string
}
