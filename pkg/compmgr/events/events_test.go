package events

import (
	"bytes"
	"context"
	"testing"
)

func TestProvider(t *testing.T) {
	data := `[
        {"id":"f1","name":"issue","passed":false},
        {"id":"p1","name":"pass","passed":true}
    ]`
	p := New(WithReader(bytes.NewBufferString(data)))
	reports, err := p.ListReports(context.Background())
	if err != nil {
		t.Fatalf("ListReports failed: %v", err)
	}
	if len(reports) != 2 {
		t.Fatalf("expected 2 reports, got %d", len(reports))
	}
	if reports[0].ID != "f1" || reports[1].Passed != true {
		t.Fatalf("unexpected reports: %+v", reports)
	}
}
