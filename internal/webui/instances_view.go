package webui

import (
	"bytes"
	"fmt"
	"html/template"
)

// InstanceCardView is one repository rendered as an instance card: its id and
// type, the controls it shows, its mount point and the directories it covers.
type InstanceCardView struct {
	ID         string
	Type       string
	MountPoint string
	Controls   []Control
	Roots      []string
}

// instanceCard is one card as the partial renders it: the card's controls with
// the positional mount point appended, and its roots as a row control.
type instanceCard struct {
	ID       string
	Type     string
	Controls []Control
	Roots    Control
}

// RenderInstanceCards renders one card per repository.
func RenderInstanceCards(cards []InstanceCardView) (template.HTML, error) {
	t, err := template.ParseFS(embedded, "templates/instances.html", "templates/controls.html")
	if err != nil {
		return "", fmt.Errorf("webui: parse instances: %w", err)
	}
	views := make([]instanceCard, len(cards))
	for i, c := range cards {
		controls := append([]Control(nil), c.Controls...)
		if c.MountPoint != "" {
			controls = append(controls, Control{
				Kind:  KindText,
				Path:  fmt.Sprintf("repositories[%d].mount_point", i),
				Label: "Mount point",
				Value: c.MountPoint,
			})
		}
		roots := make([]Option, len(c.Roots))
		for j, r := range c.Roots {
			roots[j] = Option{Value: r, Label: r}
		}
		views[i] = instanceCard{
			ID:       c.ID,
			Type:     c.Type,
			Controls: controls,
			Roots: Control{
				Kind:    KindRow,
				Path:    fmt.Sprintf("repositories[%d].roots", i),
				Label:   "Directories",
				Options: roots,
			},
		}
	}
	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "instances", views); err != nil {
		return "", fmt.Errorf("webui: render instances: %w", err)
	}
	return template.HTML(buf.String()), nil // #nosec G203 -- built from the embedded template, which escapes every card value
}

// InstancesView is the Instances page model: the rendered cards and the
// page-level error banner.
type InstancesView struct {
	Chrome
	Cards  template.HTML
	Errors []string
}
