package components

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// CreateCardListFooter creates a styled footer with controls text
func CreateCardListFooter(footerText string) tview.Primitive {
	footerBox := tview.NewBox().
		SetBorder(true).
		SetBorderColor(tcell.ColorYellow).
		SetTitle(" Controls ").
		SetTitleColor(tcell.ColorYellow)

	// Store footerText in closure for the draw function
	footerTextForDraw := footerText
	footerBox.SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
		// Draw border manually to ensure it's visible
		defStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorBlack)

		// Draw horizontal lines
		for i := x + 1; i < x+width-1; i++ {
			screen.SetContent(i, y, '─', nil, defStyle)
			screen.SetContent(i, y+height-1, '─', nil, defStyle)
		}

		// Draw vertical lines
		for i := y + 1; i < y+height-1; i++ {
			screen.SetContent(x, i, '│', nil, defStyle)
			screen.SetContent(x+width-1, i, '│', nil, defStyle)
		}

		// Draw corners
		screen.SetContent(x, y, '┌', nil, defStyle)
		screen.SetContent(x+width-1, y, '┐', nil, defStyle)
		screen.SetContent(x, y+height-1, '└', nil, defStyle)
		screen.SetContent(x+width-1, y+height-1, '┘', nil, defStyle)

		// Draw title on top border
		title := " Controls "
		titleX := x + 2
		if titleX+len(title) < x+width-2 {
			for i, r := range title {
				screen.SetContent(titleX+i, y, r, nil, tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorBlack))
			}
		}

		// Get inner rectangle (area inside border)
		innerX := x + 1
		innerY := y + 1
		innerWidth := width - 2
		innerHeight := height - 2

		// Draw text centered in the inner area
		textY := innerY + (innerHeight / 2)
		if textY >= innerY && textY < innerY+innerHeight && innerHeight > 0 && innerWidth > 0 {
			// Use Print to draw the text with white color
			tview.Print(screen, footerTextForDraw, innerX, textY, innerWidth, tview.AlignCenter, tcell.ColorWhite)
		}

		return innerX, innerY, innerWidth, innerHeight
	})

	return footerBox
}

