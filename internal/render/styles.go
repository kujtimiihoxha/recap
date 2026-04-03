package render

import (
	"charm.land/glamour/v2/ansi"
	"charm.land/lipgloss/v2"
)

const (
	hexCumin    = "#BF976F"
	hexBengal   = "#FF6E63"
	hexSriracha = "#EB4268"
	hexCoral    = "#FF577D"
	hexSalmon   = "#FF7F90"
	hexPony     = "#FF4FBF"
	hexCheeky   = "#FF79D0"
	hexMauve    = "#D46EFF"
	hexCharple  = "#6B50FF"
	hexHazy     = "#8B75FF"
	hexGuppy    = "#7272FF"
	hexMalibu   = "#00A4FF"
	hexZinc     = "#10B1AE"
	hexGuac     = "#12C78F"
	hexJulep    = "#00FFB2"
	hexBok      = "#68FFD6"
	hexCitron   = "#E8FF27"
	hexZest     = "#E8FE96"
	hexPepper   = "#201F26"
	hexCharcoal = "#3A3943"
	hexOyster   = "#605F6B"
	hexSquid    = "#858392"
	hexSmoke    = "#BFBCC8"
	hexSalt     = "#F1EFEF"
	hexButter   = "#FFFAF1"
)

var (
	colorPrimary  = lipgloss.Color(hexCharple)
	colorGreenDk  = lipgloss.Color(hexGuac)
	colorBlue     = lipgloss.Color(hexMalibu)
	colorRed      = lipgloss.Color(hexSriracha)
	colorYellow   = lipgloss.Color(hexZest)
	colorCoral    = lipgloss.Color(hexCoral)
	colorBgSubtle = lipgloss.Color(hexCharcoal)
	colorFgMuted  = lipgloss.Color(hexSquid)
	colorFgSubtle = lipgloss.Color(hexOyster)
	colorSmoke    = lipgloss.Color(hexSmoke)
	colorButter   = lipgloss.Color(hexButter)
)

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorBlue)

	dimStyle = lipgloss.NewStyle().Foreground(colorFgMuted)

	toolHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorButter).
			Background(colorPrimary).
			Padding(0, 1)

	resultHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorButter).
				Background(colorGreenDk).
				Padding(0, 1)

	statusPending = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(hexPepper)).
			Background(colorYellow).
			Padding(0, 1)

	statusReversed = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorButter).
			Background(colorRed).
			Padding(0, 1)

	warningStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorCoral)

	citationStyle = lipgloss.NewStyle().
			Foreground(colorFgSubtle).
			Italic(true)

	labelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorSmoke)

	summaryHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorButter).
				Background(lipgloss.Color(hexZinc)).
				Padding(0, 1)

	dividerStyle = lipgloss.NewStyle().
			Foreground(colorBgSubtle)
)

func sp(s string) *string { return &s }
func bp(b bool) *bool     { return &b }
func up(u uint) *uint     { return &u }

func customStyle() ansi.StyleConfig {
	return ansi.StyleConfig{
		Document: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				BlockPrefix: "\n",
				BlockSuffix: "\n",
				Color:       sp(hexSmoke),
			},
			Margin: up(0),
		},
		BlockQuote: ansi.StyleBlock{
			Indent:      up(1),
			IndentToken: sp("│ "),
		},
		List: ansi.StyleList{
			LevelIndent: 2,
		},
		Heading: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				BlockSuffix: "\n",
				Color:       sp(hexMalibu),
				Bold:        bp(true),
			},
		},
		H1: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Prefix:          " ",
				Suffix:          " ",
				Color:           sp(hexZest),
				BackgroundColor: sp(hexCharple),
				Bold:            bp(true),
			},
		},
		H2: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Prefix: "## ",
			},
		},
		H3: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Prefix: "### ",
			},
		},
		H4: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Prefix: "#### ",
			},
		},
		H5: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Prefix: "##### ",
			},
		},
		H6: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Prefix: "###### ",
				Color:  sp(hexGuac),
				Bold:   bp(false),
			},
		},
		Strikethrough: ansi.StylePrimitive{
			CrossedOut: bp(true),
		},
		Emph: ansi.StylePrimitive{
			Italic: bp(true),
		},
		Strong: ansi.StylePrimitive{
			Bold: bp(true),
		},
		HorizontalRule: ansi.StylePrimitive{
			Color:  sp(hexCharcoal),
			Format: "\n--------\n",
		},
		Item: ansi.StylePrimitive{
			BlockPrefix: "• ",
		},
		Enumeration: ansi.StylePrimitive{
			BlockPrefix: ". ",
		},
		Task: ansi.StyleTask{
			Ticked:   "[✓] ",
			Unticked: "[ ] ",
		},
		Link: ansi.StylePrimitive{
			Color:     sp(hexZinc),
			Underline: bp(true),
		},
		LinkText: ansi.StylePrimitive{
			Color: sp(hexGuac),
			Bold:  bp(true),
		},
		Image: ansi.StylePrimitive{
			Color:     sp(hexCheeky),
			Underline: bp(true),
		},
		ImageText: ansi.StylePrimitive{
			Color:  sp(hexSquid),
			Format: "Image: {{.text}} →",
		},
		Code: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Prefix:          " ",
				Suffix:          " ",
				Color:           sp(hexCoral),
				BackgroundColor: sp(hexCharcoal),
			},
		},
		CodeBlock: ansi.StyleCodeBlock{
			StyleBlock: ansi.StyleBlock{
				StylePrimitive: ansi.StylePrimitive{
					Color: sp(hexCharcoal),
				},
				Margin: up(0),
			},
			Chroma: &ansi.Chroma{
				Text: ansi.StylePrimitive{
					Color: sp(hexSmoke),
				},
				Error: ansi.StylePrimitive{
					Color:           sp(hexButter),
					BackgroundColor: sp(hexSriracha),
				},
				Comment: ansi.StylePrimitive{
					Color: sp(hexOyster),
				},
				CommentPreproc: ansi.StylePrimitive{
					Color: sp(hexBengal),
				},
				Keyword: ansi.StylePrimitive{
					Color: sp(hexMalibu),
				},
				KeywordReserved: ansi.StylePrimitive{
					Color: sp(hexPony),
				},
				KeywordNamespace: ansi.StylePrimitive{
					Color: sp(hexPony),
				},
				KeywordType: ansi.StylePrimitive{
					Color: sp(hexGuppy),
				},
				Operator: ansi.StylePrimitive{
					Color: sp(hexSalmon),
				},
				Punctuation: ansi.StylePrimitive{
					Color: sp(hexZest),
				},
				Name: ansi.StylePrimitive{
					Color: sp(hexSmoke),
				},
				NameBuiltin: ansi.StylePrimitive{
					Color: sp(hexCheeky),
				},
				NameTag: ansi.StylePrimitive{
					Color: sp(hexMauve),
				},
				NameAttribute: ansi.StylePrimitive{
					Color: sp(hexHazy),
				},
				NameClass: ansi.StylePrimitive{
					Color:     sp(hexSalt),
					Underline: bp(true),
					Bold:      bp(true),
				},
				NameDecorator: ansi.StylePrimitive{
					Color: sp(hexCitron),
				},
				NameFunction: ansi.StylePrimitive{
					Color: sp(hexGuac),
				},
				LiteralNumber: ansi.StylePrimitive{
					Color: sp(hexJulep),
				},
				LiteralString: ansi.StylePrimitive{
					Color: sp(hexCumin),
				},
				LiteralStringEscape: ansi.StylePrimitive{
					Color: sp(hexBok),
				},
				GenericDeleted: ansi.StylePrimitive{
					Color: sp(hexCoral),
				},
				GenericEmph: ansi.StylePrimitive{
					Italic: bp(true),
				},
				GenericInserted: ansi.StylePrimitive{
					Color: sp(hexGuac),
				},
				GenericStrong: ansi.StylePrimitive{
					Bold: bp(true),
				},
				GenericSubheading: ansi.StylePrimitive{
					Color: sp(hexSquid),
				},
				Background: ansi.StylePrimitive{
					BackgroundColor: sp(hexCharcoal),
				},
			},
		},
		Table: ansi.StyleTable{
			StyleBlock: ansi.StyleBlock{
				StylePrimitive: ansi.StylePrimitive{},
			},
		},
		DefinitionDescription: ansi.StylePrimitive{
			BlockPrefix: "\n ",
		},
	}
}
