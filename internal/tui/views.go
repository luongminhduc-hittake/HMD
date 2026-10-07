package tui

import (
	"strings"
)

const BannerASCII = `
    __    _ __  __        __          __  _______ 
   / /_  (_) /_/ /_____ _/ /_____   /  |/  / __ \
  / __ \/ / __/ __/ __ '/ //_/ _ \ / /|_/ / / / /
 / / / / / /_/ /_/ /_/ / ,< /  __// /  / / /_/ / 
/_/ /_/_/\__/\__/\__,_/_/|_|\___//_/  /_/_____/  `

func (m Model) View() string {
	var b strings.Builder

	// Top Banner
	b.WriteString(StyleTitle.Render(BannerASCII))
	b.WriteString("\n\n")

	switch m.State {
	case StateCheckDeps:
		b.WriteString(m.viewCheckDeps())
	case StateInputURL:
		b.WriteString(m.viewInputURL())
	case StateChangeDir:
		b.WriteString(m.viewChangeDir())
	case StateFetchingInfo:
		b.WriteString(m.viewFetchingInfo())
	case StateSelectPlaylist:
		b.WriteString(m.viewSelectPlaylist())
	case StateSelectPreset:
		b.WriteString(m.viewSelectPreset())
	case StateDownloading:
		b.WriteString(m.viewDownloading())
	case StateCompleted:
		b.WriteString(m.viewCompleted())
	case StateError:
		b.WriteString(m.viewError())
	case StateUpdating:
		b.WriteString(m.viewUpdating())
	case StateHistory:
		b.WriteString(m.viewHistory())
	case StateInputTrim:
		b.WriteString(m.viewInputTrim())
	}

	return b.String()
}
