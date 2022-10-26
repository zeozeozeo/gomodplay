package mod

import (
	"fmt"
	"sort"
)

func isStandardNotePeriod(period uint32) bool {
	if period == 0 {
		return true
	}
	return sort.SearchInts(FrequencyTable, int(period)) != len(FrequencyTable)
}

func hasStandardNotesOnly(patterns []Pattern, positions []uint8) bool {
	for _, patternIdx := range positions {
		if int(patternIdx) >= len(patterns) {
			continue
		}
		pattern := patterns[patternIdx]

		for _, row := range pattern.Rows {
			for _, note := range row {
				if !isStandardNotePeriod(note.Period) {
					return false
				}
			}
		}
	}
	return true
}

func parseFormat(format []byte) FormatDescription {
	fd := FormatDescription{
		NumChannels: 4,
		NumSamples:  15,
		Tag:         string(format),
	}

	switch string(format) {
	case "M.K.", "FLT4", "M!K!", "4CHN":
		fd.NumSamples = 31
	case "6CHN":
		fd.NumChannels = 6
		fd.NumSamples = 31
	case "12CH":
		fd.NumChannels = 12
		fd.NumSamples = 31
	case "8CHN", "CD81":
		fd.NumChannels = 8
		fd.NumSamples = 31
	case "CD61":
		panic(fmt.Errorf("unhandled mod format %s", string(format)))
	default:
		//fmt.Printf("Unknown format code %s\n", string(format))
		fd.Tag = ""
	}
	return fd
}

var sindex int = 1

func newSample(data []byte) *Sample {
	sampleName := string(data[0:22])
	size := (uint32(data[23]) + (uint32(data[22]) << 8)) << 1
	repeatOffset := (uint32(data[27]) + (uint32(data[26]))<<8) << 1
	repeatLength := (uint32(data[29]) + (uint32(data[28]))<<8) << 1

	if size > 0 {
		if repeatOffset+repeatLength > size {
			repeatOffset -= (repeatOffset - repeatLength) - size
		}
	}

	s := Sample{
		Name:         sampleName,
		size:         size,
		fineTune:     data[24],
		volume:       data[25],
		repeatOffset: repeatOffset,
		repeatLength: repeatLength,
	}
	sindex++
	return &s
}

func newNote(data []uint8, numSamples uint8) *Note {
	sampleNumber := ((data[2] & 0xf0) >> 4) + (data[0] & 0xf0)
	period := uint32(data[0]&0x0f)*256 + uint32(data[1])
	var noteName string
	if period > 0 {
		noteIdx := sort.SearchInts(FrequencyTable, int(period))
		if noteIdx < len(FrequencyTable) {
			baseNote := NoteTable[noteIdx%len(NoteTable)]
			octave := 4 - noteIdx/len(NoteTable)
			noteName = fmt.Sprintf("%s%d", baseNote, octave)
		}
	}

	n := Note{
		SampleNumber:   sampleNumber,
		Period:         period,
		Effect:         uint8(data[2]) & 0x0f,
		EffectArgument: uint8(data[3]),
		NoteName:       noteName,
	}
	return &n
}

func newPattern(data []byte) *Pattern {
	offset := 0
	numRows := 64
	size := 4

	rows := make([]Row, numRows)
	numSamples := uint8(len(data) / numRows / size)

	for rowIndex := range rows {
		row := make([]Note, numSamples)
		for noteIndex := range row {
			n := newNote(data[offset:offset+size], numSamples)
			row[noteIndex] = *n
			offset += size
		}
		rows[rowIndex] = row
	}

	p := Pattern{
		Rows: rows,
	}
	return &p
}
