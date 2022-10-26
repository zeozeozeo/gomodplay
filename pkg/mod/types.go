package mod

// Player represents a mod player
type Player struct {
	MixingMode                MixingMode
	SampleRate                uint32
	SongLoaded                bool
	SongPlaying               bool
	Standard                  Standard
	Song                      *Song
	State                     *PlayerState
	clockTicksPerSecond       float32
	clockTicksPerDeviceSample float32
}

// Song respresents currently loaded song
type Song struct {
	Name             string
	NumSamples       uint8
	NumChannels      uint8
	Samples          []*Sample
	SongLength       uint8
	Positions        []uint8
	Patterns         []Pattern
	NumUsedPatterns  uint32
	hasStandardNotes bool
	endPosition      uint32
	Format           FormatDescription
}

// Sample stores the raw sample data as well as loop and volume metadata
type Sample struct {
	Name         string
	data         []int8
	fineTune     uint8
	repeatLength uint32
	repeatOffset uint32
	size         uint32
	volume       uint8
}

// Note defines a sample, period, and effect
type Note struct {
	Effect         uint8
	EffectArgument uint8
	NoteName       string
	Period         uint32
	SampleNumber   uint8
}

// Row is just an array of notes, 1 per channel
type Row []Note

// Pattern defines the 63 rows that make up a pattern
type Pattern struct {
	Rows []Row
}

// MixingMode defines how we map channels to speakers
type MixingMode int

const (
	// AmigaMixingMode hard pans 1+3 left and 2+4 right
	AmigaMixingMode MixingMode = iota
	// StereoMixingMode reduces channel separation a little
	StereoMixingMode
	// MonoMixingMode does what you'd expect
	MonoMixingMode
)

func (mode MixingMode) String() string {
	return [...]string{"Amiga", "Stereo", "Mono"}[mode]
}

// PlayerState is the current state of the modplayer
type PlayerState struct {
	Channels                  []*ChannelInfo
	CurrentLine               uint32
	SongPatternPosition       uint32
	clockTicksPerDeviceSample float32
	CurrentVBlank             uint32
	CurrentVBlankSample       uint32
	DelayLine                 uint32
	HasLooped                 bool
	NextPatternPosition       int32
	NextPosition              int32
	PatternLoop               int32
	PatternLoopPosition       *uint32
	SamplesPerVBlank          uint32
	SetPatternPosition        bool
	SongHasEnded              bool
	SongSpeed                 uint32
	leftChannel               float32
	rightChannel              float32
}

// SampleValues returns the current channel values output
func (ps *PlayerState) SampleValues() (float32, float32) {
	return ps.leftChannel, ps.rightChannel
}

// ChannelInfo defines the current state of each output channel
type ChannelInfo struct {
	SampleNum        uint8
	arpeggioCounter  uint32
	arpeggioOffsets  []uint32
	basePeriod       uint32
	cutNoteDelay     uint32
	fineTune         uint32
	lastPortaSpeed   int32
	lastPortaTarget  uint32
	noteChange       int32
	period           uint32
	periodTarget     uint32
	retriggerCounter uint32
	retriggerDelay   uint32
	samplePos        float32
	size             uint32
	tremoloDepth     int32
	tremoloPos       uint32
	tremoloSpeed     uint32
	vibratoDepth     int32
	vibratoPos       uint32
	vibratoSpeed     uint32
	volume           float32
	volumeChange     float32
	Muted            bool
}

// FormatDescription stores the parsed data of a particular mod format/version
type FormatDescription struct {
	Tag         string
	NumChannels uint8
	NumSamples  uint8
}
