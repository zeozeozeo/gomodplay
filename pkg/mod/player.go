package mod

import (
	"errors"
	"fmt"
	"io"
)

// Standard controls NTSC/PAL vBlank timing
type Standard string

const (
	pal  Standard = "PAL"
	ntsc Standard = "NTSC"
)

var clockTicksPerSecond = map[Standard]float32{
	pal:  3546895,
	ntsc: 3579545,
}

func (p *Player) setStandard(standard Standard) {
	p.Standard = standard
	p.clockTicksPerSecond = clockTicksPerSecond[standard]
	p.clockTicksPerDeviceSample = float32(p.clockTicksPerSecond) / float32(p.SampleRate)
}

// NewModPlayer instantiates the mod player
func NewModPlayer(sampleRate uint32) *Player {
	mp := Player{
		SampleRate:  sampleRate,
		SongLoaded:  false,
		SongPlaying: false,
		MixingMode:  StereoMixingMode,
	}
	mp.setStandard(pal)
	return &mp
}

// LoadModFile parses the mod file into the player
func (p *Player) LoadModFile(f io.Reader) error {
	mod, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	songName := mod[0:20]
	format := parseFormat(mod[1080:1084])

	samples := make([]*Sample, format.NumSamples)
	offset := uint32(20)

	sampleSize := uint32(30)
	totalSampleSize := uint32(0)

	for sampleNum := range samples {
		data := mod[offset : offset+sampleSize]
		sample := newSample(data)
		samples[sampleNum] = sample
		totalSampleSize += sample.size
		offset += sampleSize
	}

	numUsedPatterns := uint8(mod[offset])
	endPosition := uint8(mod[offset+1])
	offset += 2
	positions := mod[offset : offset+128]
	offset += 128

	if format.Tag != "" {
		offset += 4
	}

	totalPatternSize := uint32(len(mod)) - totalSampleSize - offset
	singlePatternSize := uint32(format.NumChannels) * 4 * 64
	numPatterns := totalPatternSize / singlePatternSize

	minPatternRequired := uint8(0)
	for _, pattern := range positions[0:numUsedPatterns] {
		if pattern > minPatternRequired {
			minPatternRequired = pattern
		}
	}
	minPatternRequired++

	if uint32(minPatternRequired) > numPatterns {
		fmt.Printf("Overwriting number of patterns from %d to %d", minPatternRequired, numPatterns)
		numPatterns = uint32(minPatternRequired)
	}

	patternSize := uint32(format.NumChannels) * 64 * 4
	patterns := make([]Pattern, numPatterns)
	for patternNum := range patterns {
		end := offset + patternSize
		p := newPattern(mod[offset:end])
		patterns[patternNum] = *p
		offset = end
	}

	offset = uint32(len(mod)) - totalSampleSize

	for idx, sample := range samples {
		samples[idx].data = make([]int8, sample.size)

		for pos := range samples[idx].data {
			samples[idx].data[pos] = int8(mod[offset])
			offset++
		}
	}
	mod = nil

	channels := make([]*ChannelInfo, int(format.NumChannels))
	for idx := range channels {
		channel := ChannelInfo{arpeggioOffsets: []uint32{0, 0}}
		channels[idx] = &channel
	}
	ps := PlayerState{
		Channels:                  channels,
		songSpeed:                 6,
		nextPatternPosition:       -1,
		nextPosition:              -1,
		samplesPerVBlank:          p.SampleRate / 50,
		clockTicksPerDeviceSample: float32(clockTicksPerSecond[p.Standard]) / float32(p.SampleRate),
	}
	s := Song{
		Name:             string(songName),
		NumChannels:      format.NumChannels,
		Patterns:         patterns,
		Positions:        positions,
		Samples:          samples,
		SongLength:       uint8(len(positions)),
		endPosition:      uint32(endPosition),
		hasStandardNotes: hasStandardNotesOnly(patterns, positions),
		NumUsedPatterns:  uint32(numUsedPatterns),
		Format:           format,
	}
	p.State = &ps
	p.Song = &s
	p.SongLoaded = true
	return nil
}

// Err tells Beep there was an error
func (p *Player) Err() error {
	return nil
}

// Stream sends samples
func (p *Player) Stream(samples [][2]float32) (n int, ok bool) {
	if p.State.songHasEnded {
		return 0, false
	}
	for idx := range samples {
		left, right := p.nextSample()
		samples[idx][0] = left
		samples[idx][1] = right
	}
	return len(samples), true
}

// Play begins audio playback
func (p *Player) Play() error {
	if !p.SongLoaded {
		return errors.New("no song loaded")
	}
	p.SongPlaying = true
	return nil
}
