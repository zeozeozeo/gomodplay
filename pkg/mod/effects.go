package mod

import (
	"fmt"
	"sort"
)

func fineTunePeriod(period uint32, fineTune uint32, useFineTuneTable bool) uint32 {
	if useFineTuneTable {
		idx := sort.SearchInts(FrequencyTable, int(period))
		if idx == len(FrequencyTable) {
			panic("Error looking up finetuneTable")
		}
		return fineTuneTable[fineTune][idx]
	}
	return uint32(float32(period) * scaleFineTune[int8(fineTune)])
}

func changeNote(currentPeriod uint32, change int32) uint32 {
	result := int32(currentPeriod) + change
	if result > 856 {
		result = 856
	}
	if result < 113 {
		result = 113
	}
	return uint32(result)
}

func (p *Player) generateEffect(note *Note, channelNum uint, oldValues *ChannelInfo) {

	channel := p.State.Channels[channelNum]

	switch note.Effect {
	case 0:
		// arpeggio
		switch note.EffectArgument {
		case 0:
		default:
			chordOffset1 := uint32(uint8(note.EffectArgument) >> 4)
			chordOffset2 := uint32(uint8(note.EffectArgument) & 0x0f)
			channel.arpeggioOffsets = []uint32{chordOffset1, chordOffset2}
			channel.arpeggioCounter = 0
		}
	case 1:
		// slide up
		channel.noteChange = -int32(note.EffectArgument)
	case 2:
		// slide down
		channel.noteChange = int32(note.EffectArgument)
	case 3:
		// tone portamento
		if note.Period != 0 {
			channel.periodTarget = channel.period
		} else {
			if channel.lastPortaTarget != 0 {
				channel.periodTarget = channel.lastPortaTarget
			} else {
				channel.periodTarget = oldValues.period
			}
		}
		channel.period = oldValues.period
		if note.EffectArgument != 0 {
			channel.noteChange = int32(note.EffectArgument)
		} else {
			channel.noteChange = channel.lastPortaSpeed
		}
		channel.lastPortaSpeed = channel.noteChange
		channel.lastPortaTarget = channel.periodTarget
		if oldValues.SampleNum == channel.SampleNum {
			channel.samplePos = oldValues.samplePos
		}

	case 4:
		//vibrato
		speed := note.EffectArgument >> 4
		amplitude := note.EffectArgument & 0x0f
		if speed == 0 {
			channel.vibratoSpeed = oldValues.vibratoSpeed
		}
		if amplitude == 0 {
			channel.vibratoDepth = oldValues.vibratoDepth
		}
	case 5:
		// TonePortamentoVolumeSlide
		var volumeChange int16
		if note.EffectArgument&0xf0 != 0 {
			volumeChange = int16(note.EffectArgument >> 4)
		} else {
			volumeChange = -int16(note.EffectArgument)
		}
		channel.volumeChange = float32(volumeChange)
		if note.Period != 0 {
			channel.periodTarget = channel.period
		} else {
			channel.periodTarget = channel.lastPortaTarget
		}
		channel.period = oldValues.period
		channel.samplePos = oldValues.samplePos
		channel.lastPortaTarget = channel.periodTarget
		channel.noteChange = channel.lastPortaSpeed
	case 6:
		//vibratovolumeslide
		var volumeChange int16
		if note.EffectArgument&0xf0 != 0 {
			volumeChange = int16(note.EffectArgument >> 4)
		} else {
			volumeChange = -int16(note.EffectArgument)
		}
		channel.volumeChange = float32(volumeChange)
		channel.vibratoPos = oldValues.vibratoPos
		channel.vibratoSpeed = oldValues.vibratoSpeed
		channel.vibratoDepth = oldValues.vibratoDepth
	case 7:
		// tremolo
		speed := note.EffectArgument >> 4
		amplitude := note.EffectArgument & 0x0f
		if speed == 0 && amplitude == 0 {
			channel.tremoloDepth = oldValues.tremoloDepth
			channel.tremoloSpeed = oldValues.tremoloSpeed
		} else {
			channel.tremoloDepth = int32(amplitude)
			channel.tremoloSpeed = uint32(speed)
		}
	case 8:
		// pan
	case 9:
		// setsampleoffset
		if note.Period != 0 && channel.SampleNum > 0 {
			channel.samplePos = float32(uint16(note.EffectArgument) << 8)
			currentSample := p.Song.Samples[channel.SampleNum-1]
			if uint32(channel.samplePos) > currentSample.size {
				channel.samplePos = float32(uint32(channel.samplePos) % currentSample.size)
			}
		}
	case 10:
		//volumeslide
		var volumeChange int16
		if note.EffectArgument&0xf0 != 0 {
			volumeChange = int16(note.EffectArgument >> 4)
		} else {
			volumeChange = -int16(note.EffectArgument)
		}
		channel.volumeChange = float32(volumeChange)

	case 11:
		//position jump
		if uint32(note.EffectArgument) <= p.State.SongPatternPosition {
			p.State.HasLooped = true
		}
		p.State.NextPosition = int32(note.EffectArgument)
	case 12:
		// Set Volume
		channel.volume = float32(note.EffectArgument)
	case 13:
		// Pattern Break
		nextPatternPos := uint8(((0xf0 & uint32(note.EffectArgument) >> 4) * 10) + (uint32(note.EffectArgument) & 0x0f))
		p.State.NextPatternPosition = int32(nextPatternPos)
		if p.State.NextPatternPosition > 63 {
			p.State.NextPatternPosition = 0
		}
	case 14:
		extEffect := note.EffectArgument >> 4
		extArgument := note.EffectArgument & 0x0f

		switch extEffect {
		case 0:
			// SetHardwareFilter ¯\_(ツ)_/¯
		case 1:
			//FinePortaUp
			channel.period = changeNote(channel.period, -int32(extArgument))
		case 2:
			//FinePortaDown
			channel.period = changeNote(channel.period, int32(extArgument))
		case 3:
			// glissadno
		case 4:
			// setvibratowave
		case 5:
			// setfinetune
		case 6:
			// patternloop
			if extArgument == 0 {
				p.State.PatternLoopPosition = &p.State.CurrentLine
			} else {
				if p.State.PatternLoop == 0 {
					p.State.PatternLoop = int32(extArgument)
				} else {
					p.State.PatternLoop--
				}

				if p.State.PatternLoop > 0 && p.State.PatternLoopPosition != nil {
					p.State.SetPatternPosition = true
				} else {
					p.State.PatternLoopPosition = nil
				}
			}
		case 7:
			//TremoloWaveform
		case 8:
			// CoarsePan
		case 9:
			// Retrigger sample
			channel.retriggerDelay = uint32(extArgument)
			channel.retriggerCounter = 0
		case 10:
			// finevolumeslideup
			vol := channel.volume + float32(extArgument)
			if vol > 64 {
				vol = 64
			}
			channel.volume = vol
		case 11:
			//finevolumeslidedown
			vol := channel.volume - float32(extArgument)
			if vol < 0 {
				vol = 0
			}
			channel.volume = vol
		case 12:
			// Cut note
			channel.cutNoteDelay = uint32(extArgument)
		case 13:
			//delayedsample
		case 14:
			//delayedrow
			p.State.DelayLine = uint32(extArgument)
		case 15:
			//invertloop
		default:
			fmt.Printf("Unhandled extended effect %x\n", extEffect)
		}
	case 15:
		// Set Speed
		if note.EffectArgument <= 31 {
			p.State.SongSpeed = uint32(note.EffectArgument)
		} else {
			vBlanksPerSec := float32(note.EffectArgument) * 0.4
			p.State.SamplesPerVBlank = uint32(float32(p.SampleRate) / vBlanksPerSec)
		}
	default:
		fmt.Printf("Unhandled effect %x\n", note.Effect)
	}
}

func (p *Player) updateEffects() {
	for idx := range p.State.Channels {
		channel := p.State.Channels[idx]
		if channel.SampleNum == 0 {
			continue
		}

		if channel.cutNoteDelay > 0 {
			channel.cutNoteDelay--
			if channel.cutNoteDelay == 0 {
				channel.cutNoteDelay = 0
				channel.size = 0
			}
		}

		if channel.retriggerDelay > 0 {
			channel.retriggerCounter++
			if channel.retriggerDelay == channel.retriggerCounter {
				channel.samplePos = 0
				channel.retriggerCounter = 0
			}
		}
		channel.volume += channel.volumeChange
		if channel.tremoloDepth > 0 {
			baseVol := int32(p.Song.Samples[channel.SampleNum-1].volume)
			tremoloSize := (int32(vibratoTable[uint(channel.tremoloPos&63)]) * channel.tremoloDepth) / 64
			vol := baseVol + tremoloSize
			channel.tremoloPos += channel.tremoloSpeed
			channel.volume = float32(vol)
		}
		if channel.volume < 0 {
			channel.volume = 0
		}
		if channel.volume > 64 {
			channel.volume = 64
		}

		if channel.arpeggioOffsets[0] != 0 && channel.arpeggioOffsets[1] != 0 {
			newPeriod := uint32(0)
			idx := sort.SearchInts(FrequencyTable, int(channel.basePeriod))
			if idx == len(FrequencyTable) {
				panic("Error searching frequency table")
			}
			if channel.arpeggioCounter > 0 {
				noteOffset := idx - int(channel.arpeggioOffsets[channel.arpeggioCounter-1])
				if noteOffset < 0 {
					noteOffset = 0
				}
				newPeriod = uint32(FrequencyTable[noteOffset])
			} else {
				newPeriod = channel.basePeriod
			}
			channel.period = fineTunePeriod(newPeriod, channel.fineTune, p.Song.hasStandardNotes)

			channel.arpeggioCounter++
			if channel.arpeggioCounter >= 3 {
				channel.arpeggioCounter = 0
			}
		}

		if channel.vibratoDepth > 0 {
			period := fineTunePeriod(channel.basePeriod, channel.fineTune, p.Song.hasStandardNotes)
			channel.period = uint32(int32(period) + (int32(vibratoTable[uint(channel.vibratoPos&63)]) * channel.vibratoDepth / 32.0))
			channel.vibratoPos += channel.vibratoSpeed
		} else if channel.noteChange != 0 {

			if channel.periodTarget != 0 {
				if channel.periodTarget > channel.period {
					channel.period = changeNote(channel.period, channel.noteChange)
					if channel.period >= channel.periodTarget {
						channel.period = channel.periodTarget
					}
				} else {
					channel.period = changeNote(channel.period, -channel.noteChange)
					if channel.period <= channel.periodTarget {
						channel.period = channel.periodTarget
					}
				}
			} else {
				// or just moving it
				channel.period = changeNote(channel.period, channel.noteChange)
			}
		}

	}
}
