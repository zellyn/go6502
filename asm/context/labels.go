package context

/* Labeler is an interface that the flavors implement to handle label correction. */
type Labeler interface {
	LastLabel() string
	SetLastLabel(label string)
	FixLabel(label string, macroCall int, locals map[string]bool) (string, error)
	IsNewParentLabel(label string) bool
}

type LabelerBase struct {
	lastLabel string
}

func (lb *LabelerBase) LastLabel() string {
	return lb.lastLabel
}

func (lb *LabelerBase) SetLastLabel(l string) {
	lb.lastLabel = l
}
