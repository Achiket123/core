package events

import (
	"bytes"
	"context"
	"encoding/json"
	"io"

	"github.com/theopenlane/core/pkg/compmgr"
)

// Event describes the JSON structure that external systems can publish.
type Event struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Link        string            `json:"link,omitempty"`
	Passed      bool              `json:"passed,omitempty"`
	Controls    []compmgr.Control `json:"controls,omitempty"`
}

// Filter allows excluding events from processing. Returning false skips the event.
type Filter func(*Event) bool

// Provider converts external events into compliance reports.
type Provider struct {
	r      io.Reader
	filter Filter
}

// Option configures the Provider.
type Option func(*Provider)

// WithReader sets the source reader for event JSON.
func WithReader(r io.Reader) Option { return func(p *Provider) { p.r = r } }

// WithFilter sets an event filter.
func WithFilter(f Filter) Option { return func(p *Provider) { p.filter = f } }

// New returns a Provider that reads events from the supplied reader.
func New(opts ...Option) *Provider {
	p := &Provider{}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// ListReports parses available events and converts them to Reports.
func (p *Provider) ListReports(_ context.Context) ([]compmgr.Report, error) {
	if p.r == nil {
		return nil, nil
	}
	data, err := io.ReadAll(p.r)
	if err != nil {
		return nil, err
	}

	var events []Event
	if err := json.Unmarshal(data, &events); err != nil {
		// fall back to newline-delimited JSON objects
		dec := json.NewDecoder(bytes.NewReader(data))
		for {
			var e Event
			if err := dec.Decode(&e); err == io.EOF {
				break
			} else if err != nil {
				return nil, err
			}
			events = append(events, e)
		}
	}

	var reports []compmgr.Report
	for _, e := range events {
		if p.filter != nil && !p.filter(&e) {
			continue
		}
		ids := flatten(e.Controls)
		reports = append(reports, compmgr.Report{
			ID:          e.ID,
			Name:        e.Name,
			Description: e.Description,
			Link:        e.Link,
			Passed:      e.Passed,
			Controls:    e.Controls,
			ControlIDs:  ids,
		})
	}
	return reports, nil
}

func flatten(cs []compmgr.Control) []string {
	var ids []string
	for _, c := range cs {
		ids = append(ids, c.IDs...)
	}
	return ids
}

var _ compmgr.Provider = (*Provider)(nil)
