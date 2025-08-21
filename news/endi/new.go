package endi

type Endi struct {
	oldArticleTitles []string
}

func NewEndi(oldArticleTitles []string) *Endi {
	return &Endi{
		oldArticleTitles: oldArticleTitles,
	}
}
