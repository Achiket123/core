// Package events implements a generic compliance provider that reads JSON events
//
// Each event should be a JSON object using the following schema:
//
//	{
//	  "id": "unique finding ID",
//	  "name": "short title",
//	  "description": "optional details",
//	  "link": "optional provider URL",
//	  "passed": true,
//	  "controls": [
//	    {"standard": "SOC2", "version": "", "ids": ["CC1.1"]}
//	  ]
//	}
//
// Events may be newline delimited or contained in a JSON array. Systems can
// publish events in this format via Pub/Sub, HTTP, or any mechanism. The
// provider converts them to generic compliance reports so multiple sources can
// be aggregated
package events
