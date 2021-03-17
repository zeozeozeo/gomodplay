package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/benwiggins/gomodplay/pkg/mod"
	"github.com/benwiggins/gomodplay/pkg/speaker"
	"github.com/gdamore/tcell/v2"
)

var backgroundColour = tcell.GetColor("#282a36")
var effectColour = tcell.GetColor("#88DEEB")
var songColour = tcell.GetColor("#F879C0")
var patternNoteFgColour = tcell.GetColor("#F879C0")
var patternSampleFgColour = tcell.GetColor("#ffb86c")

var sampleBgColour = tcell.GetColor("#282a36")
var sampleFgColour = tcell.GetColor("#626A86")
var sampleHighlightBgColour = tcell.GetColor("#526A9E")
var sampleHighlightFgColour = tcell.GetColor("#bc91f3")

var patternHighlightBgColor = tcell.GetColor("#526A9E")
var patternHighlightFgColor = tcell.GetColor("#bc91f3")

var boxBgColour = tcell.GetColor("#282a36")
var boxFgColour = tcell.GetColor("#526A9E")

var meterColour1 = tcell.GetColor("#E1FA8C")
var meterColour2 = tcell.GetColor("#50FA7B")

var songStyle = tcell.StyleDefault.Background(backgroundColour).Bold(true).Foreground(songColour)
var sampleStyle = tcell.StyleDefault.Background(sampleBgColour).Foreground(sampleFgColour)
var sampleHighlightStyle = tcell.StyleDefault.Background(sampleHighlightBgColour).Foreground(sampleHighlightFgColour).Bold(true)

func drawSamples(s tcell.Screen, player *mod.Player) {
	xPos, yPos := 1, 1
	width, height := 27, 33

	drawBox(s, xPos, yPos, xPos+width, yPos+height)
	xPos++
	yPos++

	drawText(s, xPos, yPos, width-2, 1, songStyle, player.Song.Name)
	yPos++

	currentlyPlaying := make(map[int]bool, len(player.State.Channels))
	for _, channel := range player.State.Channels {
		if channel.SampleNum > 0 {
			currentlyPlaying[int(channel.SampleNum)-1] = true
		}
	}

	for idx, sample := range player.Song.Samples {
		if val := currentlyPlaying[idx]; val {
			drawText(s, xPos, yPos, width-2, 1, sampleHighlightStyle, fmt.Sprintf("%02d %-20s", idx+1, sample.Name))
		} else {
			drawText(s, xPos, yPos, width-2, 1, sampleStyle, fmt.Sprintf("%02d %-20s", idx+1, sample.Name))
		}
		yPos++
	}
}

var sampleSize = 8
var leftValues = make([]float32, sampleSize)
var rightValues = make([]float32, sampleSize)

func drawMeters(s tcell.Screen, player *mod.Player) {
	x, y := 1, 35
	width, height := 126, 3
	drawBox(s, x, y, x+width, y+height)
	xPos := x + 1
	yPos := y + 1

	left, right := player.State.SampleValues()
	if left < 0 {
		left = -left
	}
	if right < 0 {
		right = -right
	}
	if left > 1 {
		left = 1
	}
	if right > 1 {
		right = 1
	}

	leftValues = append(leftValues[1:], left)
	rightValues = append(rightValues[1:], right)

	var leftSum float32
	var rightSum float32
	for i := 0; i < sampleSize; i++ {
		leftSum += leftValues[i]
		rightSum += rightValues[i]
	}
	leftAvg := leftSum / float32(sampleSize)
	rightAvg := rightSum / float32(sampleSize)

	leftDB := 20 * math.Log10(float64(leftAvg))
	rightDB := 20 * math.Log10(float64(rightAvg))
	if leftDB > 0 {
		leftDB = 0
	}
	if rightDB > 0 {
		rightDB = 0
	}
	if leftDB < -96 {
		leftDB = -96
	}
	if rightDB < -96 {
		rightDB = -96
	}

	style := tcell.StyleDefault.Background(backgroundColour).Foreground(meterColour1)

	runes := []string{"▏", "▎", "▍", "▌", "▋", "▊", "▉", "█"}
	leftLength := float32(width-2) * float32(96+leftDB) / 96
	rightLength := float32(width-2) * float32(96+rightDB) / 96

	for i := 0; i < int(leftLength); i++ {
		drawText(s, xPos, yPos, 1, 1, style, runes[7])
		xPos++
	}
	remainder := leftLength - float32(int(leftLength))
	if remainder >= 0.125 {
		idx := int(remainder*8) - 1
		drawText(s, xPos, yPos, 1, 1, style, runes[idx])
		xPos++
	}

	yPos++
	xPos = x + 1
	for i := 0; i < int(rightLength); i++ {
		drawText(s, xPos, yPos, 1, 1, style, runes[7])
		xPos++
	}
	if remainder >= 0.125 {
		idx := int(remainder*8) - 1
		drawText(s, xPos, yPos, 1, 1, style, runes[idx])
		xPos++
	}
}

func drawPatterns(s tcell.Screen, player *mod.Player) {
	x, y := 33, 1
	width, height := 94, 33
	drawBox(s, x, y, x+width, y+height)
	xPos := x + 1
	yPos := y + 1

	defaultStyle := tcell.StyleDefault.Background(backgroundColour).Foreground(tcell.GetColor("#626A86"))
	highlightStyle := tcell.StyleDefault.Background(patternHighlightBgColor).Foreground(patternHighlightFgColor).Bold(true)
	numRows := 32
	var lineIdx int
	if player.State.CurrentLine < 16 {
		lineIdx = 0
	} else if player.State.CurrentLine > 48 {
		lineIdx = 32
	} else {
		lineIdx = int(player.State.CurrentLine) - 16
	}

	for rowNum := 0; rowNum < numRows && lineIdx < 64; rowNum++ {
		var style tcell.Style
		patternIdx := player.Song.Positions[player.State.SongPatternPosition]
		pattern := player.Song.Patterns[patternIdx]

		if uint32(lineIdx) == player.State.CurrentLine {
			style = highlightStyle
		} else {
			style = defaultStyle
		}

		row := pattern.Rows[lineIdx]

		rowNumber := fmt.Sprintf("%02d.%02d", player.State.SongPatternPosition, lineIdx)
		drawText(s, xPos, yPos, width-2, 1, style, rowNumber)
		xPos += 5

		for idx, note := range row {
			drawText(s, xPos, yPos, 1, 1, style, "│")
			xPos++
			noteStyle := style
			sampleStyle := style
			effectStyle := style

			if !player.State.Channels[idx].Muted {
				noteStyle = style.Foreground(patternNoteFgColour)
				sampleStyle = style.Foreground(patternSampleFgColour)
				effectStyle = style.Foreground(effectColour)
			}

			if note.NoteName != "" {
				drawText(s, xPos, yPos, 4, 1, noteStyle, note.NoteName+" ")
			} else {
				drawText(s, xPos, yPos, 4, 1, style, "... ")
			}
			xPos += 4

			if note.SampleNumber > 0 {
				sampleNumber := fmt.Sprintf("%02d", note.SampleNumber)
				drawText(s, xPos, yPos, 3, 1, sampleStyle, sampleNumber)
			} else {
				drawText(s, xPos, yPos, 3, 1, style, "..")
			}
			xPos += 3

			if note.Effect > 0 {
				effect := fmt.Sprintf("%x%02x", note.Effect, note.EffectArgument)
				drawText(s, xPos, yPos, 3, 1, effectStyle, effect)
			} else {
				drawText(s, xPos, yPos, 3, 1, style, "...")
			}
			xPos += 3
		}
		lineIdx++
		xPos = x + 1
		yPos++
	}
}

func main() {
	sampleRate := uint32(48000)
	player := mod.NewModPlayer(sampleRate)

	loading := true
	f := load()
	loading = false
	err := player.LoadModFile(f)
	if err != nil {
		panic(err)
	}
	f.Close()

	err = speaker.Init(sampleRate, sampleRate/100)
	if err != nil {
		panic(err)
	}

	done := make(chan bool)
	player.Play()
	speaker.Play(player, func() {
		done <- true
	})

	defStyle := tcell.StyleDefault.Background(backgroundColour).Foreground(tcell.ColorReset)

	// Initialize screen
	s, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("%+v", err)
	}
	if err := s.Init(); err != nil {
		log.Fatalf("%+v", err)
	}
	s.SetStyle(defStyle)
	s.Clear()

	// Event loop
	quit := func() {
		s.Fini()
		os.Exit(0)
	}

	go func() {
		for {
			if loading {
				continue
			}
			s.Show()
			start := time.Now()
			drawSamples(s, player)
			drawPatterns(s, player)
			drawMeters(s, player)
			end := time.Since(start)

			xPos, yPos := 2, 0

			drawText(s, xPos, yPos, 0, 132, defStyle.Foreground(sampleFgColour).Bold(true).Underline(true), "M")
			xPos++
			drawText(s, xPos, yPos, 12, 1, defStyle.Foreground(sampleFgColour).Bold(true), "ixing mode:")
			xPos += 12
			drawText(s, xPos, yPos, 8, 1, defStyle.Foreground(effectColour), player.MixingMode.String())
			xPos += 8

			drawText(s, xPos, yPos, 1, 1, defStyle.Foreground(sampleFgColour).Bold(true).Underline(true), "S")
			xPos++
			drawText(s, xPos, yPos, 9, 1, defStyle.Foreground(sampleFgColour).Bold(true), "tandard:")
			xPos += 9
			drawText(s, xPos, yPos, 8, 1, defStyle.Foreground(effectColour), string(player.Standard))
			xPos += 8

			xPos = 64
			drawText(s, xPos, yPos, 8, 1, defStyle.Foreground(sampleFgColour).Bold(true), "Format:")
			xPos += 8
			drawText(s, xPos, yPos, 8, 1, defStyle.Foreground(patternSampleFgColour), player.Song.Format.Tag)
			xPos += 5

			drawText(s, xPos, yPos, 8, 1, defStyle.Foreground(sampleFgColour).Bold(true), "Position:")
			xPos += 10
			drawText(s, xPos, yPos, 8, 1, defStyle.Foreground(patternSampleFgColour), fmt.Sprintf("%d/%d", player.State.SongPatternPosition, player.Song.NumUsedPatterns))
			xPos += 6

			xPos = 106
			drawText(s, xPos, yPos, 13, 1, defStyle.Foreground(sampleFgColour).Bold(true), "Render time:")
			xPos += 13
			drawText(s, xPos, yPos, 14, 1, defStyle.Foreground(effectColour), fmt.Sprintf("%v", end))
			xPos += 14

			time.Sleep(time.Second / 60)
		}
	}()

	for {
		if loading {
			continue
		}
		// Poll event
		ev := s.PollEvent()

		// Process event
		switch ev := ev.(type) {
		case *tcell.EventResize:
			s.Sync()
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
				quit()
			} else {
				rune := ev.Rune()
				switch rune {
				case 'L', 'l':
					loading = true
					s.Suspend()
					f := load()
					err = player.LoadModFile(f)
					if err != nil {
						panic(err)
					}
					s.Resume()
					s.Clear()
					loading = false
				case 'M', 'm':
					player.MixingMode = (player.MixingMode + 1) % 3
				case '1', '2', '3', '4', '5', '6', '7', '8':
					channelNumber, err := strconv.Atoi(string(rune))
					channelNumber--
					if err == nil && channelNumber < len(player.State.Channels) {
						player.State.Channels[channelNumber].Muted = !player.State.Channels[channelNumber].Muted
					}
				case 's', 'S':
					if player.Standard == "NTSC" {
						player.Standard = "PAL"
					} else {
						player.Standard = "NTSC"
					}
				case 'q', 'Q':
					quit()
				}
			}
		}
	}
}
