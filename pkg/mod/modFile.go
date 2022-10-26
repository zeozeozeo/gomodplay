package mod

func (p *Player) getSongRow() *Row {
	patternIdx := p.Song.Positions[p.State.SongPatternPosition]
	pattern := p.Song.Patterns[patternIdx]
	row := pattern.Rows[p.State.CurrentLine]
	return &row
}

func (p *Player) playNote(note Note, channelNum uint) {
	prevState := *p.State.Channels[channelNum]
	channel := p.State.Channels[channelNum]

	if note.SampleNumber > 0 {
		currentSample := p.Song.Samples[note.SampleNumber-1]
		channel.volume = float32(currentSample.volume)
		channel.size = currentSample.size
		channel.SampleNum = note.SampleNumber
		channel.fineTune = uint32(currentSample.fineTune)
	}

	channel.volumeChange = 0
	channel.noteChange = 0
	channel.retriggerDelay = 0
	channel.vibratoSpeed = 0
	channel.vibratoDepth = 0
	channel.tremoloSpeed = 0
	channel.tremoloDepth = 0
	channel.arpeggioCounter = 0
	channel.arpeggioOffsets[0] = 0
	channel.arpeggioOffsets[1] = 0

	if note.Period != 0 {
		channel.period = fineTunePeriod(note.Period, channel.fineTune, p.Song.hasStandardNotes)
		channel.basePeriod = note.Period
		channel.samplePos = 0
		if channel.SampleNum > 0 {
			currentSample := p.Song.Samples[channel.SampleNum-1]
			channel.size = currentSample.size
		}
	}
	p.generateEffect(&note, channelNum, &prevState)
}

func (p *Player) playLine() {
	if p.State.NextPatternPosition != -1 {
		p.State.SongPatternPosition++
		p.State.CurrentLine = uint32(p.State.NextPatternPosition)
		p.State.NextPatternPosition = -1
	} else if p.State.NextPosition != -1 {
		p.State.SongPatternPosition = uint32(p.State.NextPosition)
		p.State.CurrentLine = 0
		p.State.NextPosition = -1
	}

	if p.State.SongPatternPosition >= p.Song.NumUsedPatterns {
		if p.Song.endPosition < p.Song.NumUsedPatterns {
			p.State.SongPatternPosition = p.Song.endPosition
			p.State.HasLooped = true
		} else {
			p.State.SongHasEnded = true
		}
	}

	row := *p.getSongRow()
	for channelNum := range row {
		note := row[channelNum]
		p.playNote(note, uint(channelNum))
	}

	if p.State.SetPatternPosition && p.State.PatternLoopPosition != nil {
		p.State.SetPatternPosition = false
		p.State.CurrentLine = *p.State.PatternLoopPosition
	} else {
		p.State.CurrentLine++
		if p.State.CurrentLine >= 64 {
			p.State.SongPatternPosition++
			if p.State.SongPatternPosition >= p.Song.NumUsedPatterns {
				p.State.SongHasEnded = true
			}
			p.State.CurrentLine = 0
		}
	}
}

func (p *Player) nextSample() (left float32, right float32) {
	if !(p.SongLoaded && p.SongPlaying) {
		return
	}

	if p.State.CurrentVBlankSample >= p.State.SamplesPerVBlank {
		p.State.CurrentVBlankSample = 0

		p.updateEffects()

		if p.State.CurrentVBlank >= p.State.SongSpeed {
			if p.State.DelayLine > 0 {
				p.State.DelayLine--
			} else {
				p.State.CurrentVBlank = 0
				p.playLine()

			}
		}
		p.State.CurrentVBlank++
	}
	p.State.CurrentVBlankSample++

	for channelNum := range p.State.Channels {
		channel := p.State.Channels[channelNum]
		if channel.size > 2 {
			currentSample := p.Song.Samples[channel.SampleNum-1]
			if channel.size <= 2 {
				continue
			}

			if channel.samplePos >= float32(channel.size) {
				overflow := channel.samplePos - float32(channel.size)
				channel.samplePos = float32(currentSample.repeatOffset) + overflow
				channel.size = currentSample.repeatOffset + currentSample.repeatLength
				if channel.size <= 2 {
					continue
				}
			}

			rawValue := currentSample.data[uint32(channel.samplePos)]
			channelValue := float32(rawValue) / 128 * channel.volume / 64

			if channel.period != 0 {
				channel.samplePos += (p.clockTicksPerDeviceSample / float32(channel.period))
			}

			if channel.Muted {
				continue
			}

			outputChannel := channelNum % 4
			if outputChannel == 0 || outputChannel == 3 {
				left += channelValue
				switch p.MixingMode {
				case StereoMixingMode:
					right += channelValue * 0.33
				case MonoMixingMode:
					right += channelValue
				}

			} else {
				right += channelValue
				switch p.MixingMode {
				case StereoMixingMode:
					left += channelValue * 0.33
				case MonoMixingMode:
					left += channelValue
				}
			}
		}
	}
	p.State.leftChannel = left
	p.State.rightChannel = right
	return
}
