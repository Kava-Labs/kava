package util

import (
	abci "github.com/tendermint/tendermint/abci/types"
)

// FilterEventsByType returns a slice of events that match the given type.
func FilterEventsByType(events []abci.Event, eventType string) []abci.Event {
	filteredEvents := []abci.Event{}

	for _, event := range events {
		if event.Type == eventType {
			filteredEvents = append(filteredEvents, event)
		}
	}

	return filteredEvents
}

// FilterTxEventsByType returns a slice of events that match the given type
// from any and all txs in a slice of ResponseDeliverTx.
func FilterTxEventsByType(txs []*abci.ResponseDeliverTx, eventType string) []abci.Event {
	filteredEvents := []abci.Event{}

	for _, tx := range txs {
		events := FilterEventsByType(tx.Events, eventType)
		filteredEvents = append(filteredEvents, events...)
	}

	return filteredEvents
}
