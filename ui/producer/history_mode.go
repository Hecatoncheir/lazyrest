package producer

type HistoryMode uint8

const (
	HistoryMetadata HistoryMode = iota
	HistoryFull
)

func (widget *Producer) historyModeValue() HistoryMode {
	if widget == nil {
		return HistoryMetadata
	}
	if HistoryMode(widget.historyMode.Load()) == HistoryFull {
		return HistoryFull
	}
	return HistoryMetadata
}

// HistoryMode reports how future history snapshots are persisted.
func (widget *Producer) HistoryMode() HistoryMode {
	return widget.historyModeValue()
}

// SetHistoryMode applies the mode to the next persisted snapshot. Switching to
// metadata mode also rewrites existing in-memory history without stored details.
func (widget *Producer) SetHistoryMode(mode HistoryMode) {
	if widget == nil {
		return
	}
	if mode != HistoryFull {
		mode = HistoryMetadata
	}
	if previous := HistoryMode(widget.historyMode.Swap(uint32(mode))); previous != mode {
		widget.persistHistory()
	}
}
