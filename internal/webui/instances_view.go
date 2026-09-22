package webui

import "html/template"

// InstanceCardView is one repository rendered as an instance card: its id and
// type, the controls it shows, its mount point and the directories it covers.
type InstanceCardView struct {
	ID         string
	Type       string
	MountPoint string
	Controls   []Control
	Roots      []string
}

// RenderInstanceCards renders one card per repository.
func RenderInstanceCards(cards []InstanceCardView) (template.HTML, error) {
	return "", nil
}
