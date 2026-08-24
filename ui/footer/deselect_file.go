package footer

func (widget *Footer) DeselectFile() {
	widget.selectedFile = nil
	widget.suiteName = ""
	widget.render()
}
