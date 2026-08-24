package footer

import (
	"github.com/Hecatoncheir/lazyrest/finder"
)

func (widget *Footer) SelectFile(file finder.File) {
	widget.selectedFile = &file
	widget.suiteName = ""
	widget.render()
}
