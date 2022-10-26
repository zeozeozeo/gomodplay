package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"github.com/zeozeozeo/gomodplay/pkg/mod"

	"github.com/gdamore/tcell/v2"
)

var fileStyle = tcell.StyleDefault.Background(sampleBgColour).Foreground(sampleFgColour)
var fileHighlightStyle = tcell.StyleDefault.Background(sampleHighlightBgColour).Foreground(sampleHighlightFgColour).Bold(true)
var modRegexp = regexp.MustCompile("(?i).mod")

type file struct {
	name       string
	isDir      bool
	size       int64
	moduleName *string
}

func parseDir(path string) ([]file, error) {
	var matchingFiles []file

	if path != "/" {
		parentDir := file{
			name:  "../",
			isDir: true,
		}

		matchingFiles = append(matchingFiles, parentDir)
	}

	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		if f.IsDir() {
			dir := file{
				name:  fmt.Sprintf("%s/", f.Name()),
				size:  f.Size(),
				isDir: true,
			}
			matchingFiles = append(matchingFiles, dir)
			continue
		}
		name := f.Name()
		matched := modRegexp.MatchString(name)
		if matched && name != "go.mod" {
			mod := file{
				name:  name,
				size:  f.Size(),
				isDir: false,
			}
			matchingFiles = append(matchingFiles, mod)
		}
	}

	return matchingFiles, nil
}

func changeDir(s tcell.Screen, dir string) ([]file, error) {
	dir, err := filepath.Abs(fmt.Sprintf("%s/%s", currentState.currentDir, dir))
	if err != nil {
		return nil, err
	}
	currentState.currentDir = dir
	currentState.currentIdx = 0
	return parseDir(currentState.currentDir)
}

type state struct {
	currentDir string
	currentIdx int
	entries    []file
}

var currentState *state

func load() *os.File {
	s, e := tcell.NewScreen()
	defer s.Fini()
	if e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}
	if e := s.Init(); e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}

	defStyle := tcell.StyleDefault.
		Background(tcell.ColorBlack).
		Foreground(tcell.ColorWhite)
	s.SetStyle(defStyle)
	s.Clear()

	// Event loop
	quit := func() {
		s.Fini()
		os.Exit(0)
	}

	currentState = &state{
		currentDir: "./modfiles",
	}

	dirEntries, err := parseDir(currentState.currentDir)
	if err != nil {
		panic(err)
	}
	currentState.entries = dirEntries

	terminate := make(chan bool)

	go func(currentState *state) {
		var m sync.Mutex

		for {
			select {
			case <-terminate:
				return
			default:
				s.Show()
				drawBox(s, 0, 0, 130, 38)
				yPos := 1
				for idx, file := range currentState.entries {
					xPos := 1
					var style tcell.Style
					if idx == currentState.currentIdx {
						style = fileHighlightStyle
					} else {
						style = fileStyle
					}
					drawText(s, xPos, yPos, 32, 1, style, fmt.Sprintf("%-31s", file.name))
					xPos += 32

					if file.isDir {
						drawText(s, xPos, yPos, 9, 1, style, "<dir>")
					} else {
						drawText(s, xPos, yPos, 9, 1, style, fmt.Sprintf("%-8d", file.size))

						xPos += 9

						if file.moduleName != nil {
							drawText(s, xPos, yPos, 20, 1, style, *file.moduleName)
						} else {
							m.Lock()
							path := filepath.Join(currentState.currentDir, file.name)
							f, err := os.Open(path)
							if err == nil {
								defer f.Close()
								player := mod.NewModPlayer(48000)
								player.LoadModFile(f)
								name := player.Song.Name
								currentState.entries[idx].moduleName = &name

							} else {
								msg := ""
								file.moduleName = &msg
							}
							m.Unlock()
						}

					}
					yPos++
				}
				time.Sleep(time.Second / 60)
			}
		}
	}(currentState)

	for {
		switch ev := s.PollEvent().(type) {
		case *tcell.EventResize:
			s.Sync()
		case *tcell.EventKey:
			switch key := ev.Key(); key {
			case tcell.KeyDown:
				if currentState.currentIdx < len(currentState.entries)-1 {
					currentState.currentIdx++
				} else {
					currentState.currentIdx = 0
				}
			case tcell.KeyUp:
				if currentState.currentIdx > 0 {
					currentState.currentIdx--
				} else {
					currentState.currentIdx = len(currentState.entries) - 1
				}
			case tcell.KeyEscape:
				quit()
			case tcell.KeyEnter:
				file := (currentState.entries)[currentState.currentIdx]
				if file.isDir {
					entries, err := changeDir(s, file.name)
					if err != nil {
						panic(err)
					}
					currentState.entries = entries
					s.Clear()
				} else {
					path := filepath.Join(currentState.currentDir, file.name)
					file, err := os.Open(path)
					if err != nil {
						panic(err)
					}
					terminate <- true
					return file
				}
			case tcell.KeyHome:
				currentState.currentIdx = 0
			case tcell.KeyEnd:
				currentState.currentIdx = len(currentState.entries) - 1
			}
		}
	}
}
