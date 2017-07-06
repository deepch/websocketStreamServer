package ts

import (
	"container/list"
	"logger"
	"mediaTypes/aac"
	"mediaTypes/flv"
	"mediaTypes/h264"
	"mediaTypes/mp3"
)

var crc32Table []uint32

const (
	PMT_ID    = 0xfff
	Video_Id  = 0x100
	Audio_Id  = 0x101
	TS_length = 188
	PCR_HZ    = 27000000
)

func init() {
	crc32Table = make([]uint32, 256)
	for i := uint32(0); i < 256; i++ {
		k := uint32(0)
		for j := (i << 24) | 0x800000; j != 0x80000000; j <<= 1 {
			tmp := ((k ^ j) & 0x80000000)
			if tmp != 0 {
				k = (k << 1) ^ 0x04c11db7
			} else {
				k = (k << 1) ^ 0
			}
		}
		crc32Table[i] = k
	}
}

func Crc32Calculate(buffer []uint8) (crc32reg uint32) {
	crc32reg = 0xFFFFFFFF
	for _, v := range buffer {
		crc32reg = (crc32reg << 8) ^ crc32Table[((crc32reg>>24)^uint32(v))&0xFF]
	}

	return crc32reg
}

type TsCreater struct {
	pcrTimes       int16
	tsVcount       int16
	tsAcount       int16
	audioHeader    []byte
	asc            aac.AudioSpecificConfig
	videoHeader    []byte
	sps            []byte
	pps            []byte
	sei            []byte
	videoTypeId    int
	audioTypeId    int
	audioFrameSize int
	audioSampleHz  int
	audioPts       int64
	beginTime      uint32
	nowTime        uint32
	tsCache        *list.List
	keyframeWrited bool
}

func (this *TsCreater) AddTag(tag *flv.FlvTag) {
	if flv.FLV_TAG_ScriptData == tag.TagType {
		return
	}
	if this.tsCache == nil {
		this.keyframeWrited = false
		this.tsCache = list.New()
	}
	this.nowTime = tag.Timestamp
	if true == this.avHeaderAdded(tag) {
		if 0xffffffff == this.beginTime {
			this.beginTime = tag.Timestamp
			this.addPatPmt()
		}
		var addDts, addPCR bool
		if flv.FLV_TAG_Audio == tag.TagType {
			addDts = false
		} else {
			addDts = true
		}
		if this.pcrTimes == 0 {
			addPCR = true
		} else {
			addPCR = false
		}

		if flv.FLV_TAG_Video == tag.TagType {
			this.pcrTimes++
			if this.pcrTimes == 4 {
				this.pcrTimes = 0
				this.addPatPmt()
			}
		}

		var tsCount, padSize int
		var tmp16 uint16

		var dataPayload []byte
		var payloadSize int

		if flv.FLV_TAG_Audio == tag.TagType {
			if this.audioTypeId == 0xf {
				adth := aac.GenerateADTHeader(this.asc, len(tag.Data)-2)
				payloadSize = len(adth) + len(tag.Data) - 2
				dataPayload = make([]byte, payloadSize)
				copy(dataPayload, adth)
				copy(dataPayload[len(adth):], tag.Data[2:])
			} else if 0x03 == this.audioTypeId {
				payloadSize = len(tag.Data) - 1
				dataPayload = make([]byte, payloadSize)
				copy(dataPayload, tag.Data[1:])
			} else if 0x04 == this.audioTypeId {
				payloadSize = len(tag.Data) - 1
				dataPayload = make([]byte, payloadSize)
				copy(dataPayload, tag.Data[1:])
			} else {
				logger.LOGW("not support audio")
			}
		} else if flv.FLV_TAG_Video == tag.TagType {
			dataPayload = this.videoPayload(tag)
			if nil == dataPayload {
				logger.LOGF(dataPayload)
				return
			}
			payloadSize = len(dataPayload)
		}

		tsCount, padSize = this.getTsCount(payloadSize, addPCR, addDts)

		tsBuf := make([]byte, TS_length)
		cur := 0
		if 1 == tsCount {
			for idx := 0; idx < TS_length; idx++ {
				tsBuf[idx] = 0xff
			}
			cur = 0
			tsBuf[cur] = 0x47
			cur++
			if flv.FLV_TAG_Audio == tag.TagType {
				tmp16 = uint16(0x4000 | Audio_Id)
			} else {
				tmp16 = uint16(0x4000 | Video_Id)
			}
			tsBuf[cur] = byte(tmp16 >> 8)
			cur++
			tsBuf[cur] = byte(tmp16 & 0xff)
			cur++
			if addPCR || padSize > 0 {
				if flv.FLV_TAG_Audio == tag.TagType {
					tsBuf[cur] = byte(0x30 | this.tsAcount)
					cur++
				} else {
					tsBuf[cur] = byte(0x30 | this.tsVcount)
					cur++
				}
			} else {
				if flv.FLV_TAG_Audio == tag.TagType {
					tsBuf[cur] = byte(0x10 | this.tsAcount)
					cur++
				} else {
					tsBuf[cur] = byte(0x10 | this.tsVcount)
					cur++
				}
			}
			if flv.FLV_TAG_Audio == tag.TagType {
				this.tsAcount++
				if this.tsAcount == 16 {
					this.tsAcount = 0
				}
			} else {
				this.tsVcount++
				if this.tsVcount == 16 {
					this.tsVcount = 0
				}
			}

			//!四字节头
			//PCR、PAD
			timeMS := uint64(tag.Timestamp - this.beginTime)
			pcr := uint64(((timeMS * (PCR_HZ / 1000)) / 300) % 0x200000000)
			if addPCR {
				adpLength := 7 + padSize
				tsBuf[cur] = byte(adpLength)
				cur++
				tsBuf[cur] = 0x10
				cur++
				tsBuf[cur] = byte((pcr & 0xfe000000) >> 25)
				cur++
				tsBuf[cur] = byte((pcr & 0x1fe0000) >> 17)
				cur++
				tsBuf[cur] = byte((pcr & 0x1fe00) >> 9)
				cur++
				tsBuf[cur] = byte((pcr & 0x1fe) >> 1)
				cur++
				tsBuf[cur] = byte(((pcr & 1) << 7) | 0x7e)
				cur++
				tsBuf[cur] = 0
				cur++
				cur += padSize
			} else if false == addPCR && padSize > 0 {
				adpLength := padSize - 1
				tsBuf[cur] = byte(adpLength)
				cur++
				if padSize > 1 {
					tsBuf[cur] = 0
					cur += padSize - 1
				}
			}
			//!PCR PAD
			//PES
			tsBuf[cur] = 0x00
			cur++
			tsBuf[cur] = 0x00
			cur++
			tsBuf[cur] = 0x01
			cur++
			if flv.FLV_TAG_Audio == tag.TagType {
				tsBuf[cur] = 0xc0
				cur++
				tmp16 = uint16(payloadSize + 8)
				tsBuf[cur] = byte(tmp16 >> 8)
				cur++
				tsBuf[cur] = byte(tmp16 & 0xff)
				cur++
				tsBuf[cur] = 0x80
				cur++
				tsBuf[cur] = 0x80
				cur++
				tsBuf[cur] = 0x05
				cur++

				audioPtsDelta := int64(90000 * int64(this.audioFrameSize) / int64(this.audioSampleHz))
				this.audioPts += audioPtsDelta
				tsBuf[cur] = byte((0x20) | ((this.audioPts & 0x1c0000000) >> 29) | 1)
				cur++
				tmp16 = uint16(((this.audioPts & 0x3fff8000) >> 14) | 1)
				tsBuf[cur] = byte((tmp16 >> 8) & 0xff)
				cur++
				tsBuf[cur] = byte(tmp16 & 0xff)
				cur++
				tmp16 = uint16((this.audioPts&0x7fff)<<1) | 1
				tsBuf[cur] = byte((tmp16 >> 8) & 0xff)
				cur++
				tsBuf[cur] = byte(tmp16 & 0xff)
				cur++

				copy(tsBuf[cur:], dataPayload)
				cur += payloadSize
			} else {
				tsBuf[cur] = 0xe0
				cur++
				tsBuf[cur] = 0x00
				cur++
				tsBuf[cur] = 0x00
				cur++
				tsBuf[cur] = 0x80
				cur++
				tsBuf[cur] = 0xc0
				cur++
				tsBuf[cur] = 0x0a
				cur++

				tsBuf[cur] = byte((3 << 4) | ((pcr & 0x1c0000000) >> 29) | 1)
				cur++
				tmp16 = uint16(((pcr & 0x3fff8000) >> 14) | 1)
				tsBuf[cur] = byte((tmp16 >> 8) & 0xff)
				cur++
				tsBuf[cur] = byte(tmp16 & 0xff)
				cur++
				tmp16 = uint16(((pcr & 0x7fff) << 1) | 1)
				tsBuf[cur] = byte((tmp16 >> 8) & 0xff)
				cur++
				tsBuf[cur] = byte(tmp16 & 0xff)
				cur++
				tsBuf[cur] = byte((1 << 4) | ((pcr & 0x1c0000000) >> 29) | 1)
				cur++
				tmp16 = uint16(((pcr & 0x3fff8000) >> 14) | 1)
				tsBuf[cur] = byte((tmp16 >> 8) & 0xff)
				cur++
				tsBuf[cur] = byte(tmp16 & 0xff)
				cur++
				tmp16 = uint16(((pcr & 0x7fff) << 1) | 1)
				tsBuf[cur] = byte((tmp16 >> 8) & 0xff)
				cur++
				tsBuf[cur] = byte(tmp16 & 0xff)
				cur++
				copy(tsBuf[cur:], dataPayload)
				cur += payloadSize
			}
			//!PES
			this.appendTsPkt(tsBuf)

		} else {
			//不止一个包的情况
			payloadCur := 0
			for i := 0; i < tsCount; i++ {
				for idx := 0; idx < len(tsBuf); idx++ {
					tsBuf[idx] = 0xff
				}
				cur = 0
				//第一帧
				if 0 == i {
					tsBuf[cur] = 0x47
					cur++
					if flv.FLV_TAG_Audio == tag.TagType {
						tmp16 = uint16(0x4000 | Audio_Id)
					} else {
						tmp16 = uint16(0x4000 | Video_Id)
					}
					tsBuf[cur] = byte(tmp16 >> 8)
					cur++
					tsBuf[cur] = byte(tmp16 & 0xff)
					cur++
					if addPCR {
						if flv.FLV_TAG_Audio == tag.TagType {
							tsBuf[cur] = byte(0x30 | this.tsAcount)
							cur++
						} else {
							tsBuf[cur] = byte(0x30 | this.tsVcount)
							cur++
						}
					} else {
						if flv.FLV_TAG_Audio == tag.TagType {
							tsBuf[cur] = byte(0x10 | this.tsAcount)
							cur++
						} else {
							tsBuf[cur] = byte(0x10 | this.tsVcount)
							cur++
						}
					}

					if flv.FLV_TAG_Audio == tag.TagType {
						this.tsAcount++
						if this.tsAcount == 16 {
							this.tsAcount = 0
						}
					} else {
						this.tsVcount++
						if this.tsVcount == 16 {
							this.tsVcount = 0
						}
					}

					//!四字节头
					//PCR
					timeMS := uint64(tag.Timestamp - this.beginTime)
					pcr := uint64(((timeMS * (PCR_HZ / 1000)) / 300) % 0x200000000)
					if addPCR {
						adpLength := 7
						tsBuf[cur] = byte(adpLength)
						cur++
						tsBuf[cur] = 0x10
						cur++
						tsBuf[cur] = byte((pcr & 0xfe000000) >> 25)
						cur++
						tsBuf[cur] = byte((pcr & 0x1fe0000) >> 17)
						cur++
						tsBuf[cur] = byte((pcr & 0x1fe00) >> 9)
						cur++
						tsBuf[cur] = byte((pcr & 0x1fe) >> 1)
						cur++
						tsBuf[cur] = byte(((pcr & 1) << 7) | 0x7e)
						cur++
						tsBuf[cur] = 0
						cur++
					}
					//!PCR
					//PES头
					tsBuf[cur] = 0x00
					cur++
					tsBuf[cur] = 0x00
					cur++
					tsBuf[cur] = 0x01
					cur++
					if flv.FLV_TAG_Audio == tag.TagType {
						tsBuf[cur] = 0xc0
						cur++
						tmp16 = uint16(payloadSize + 8)
						tsBuf[cur] = byte(tmp16 >> 8)
						cur++
						tsBuf[cur] = byte(tmp16 & 0xff)
						cur++
						tsBuf[cur] = 0x80
						cur++
						tsBuf[cur] = 0x80
						cur++
						tsBuf[cur] = 0x05
						cur++

						audioPtsDelta := int64(90000 * int64(this.audioFrameSize) / int64(this.audioSampleHz))
						this.audioPts += audioPtsDelta
						tsBuf[cur] = byte((0x20) | ((this.audioPts & 0x1c0000000) >> 29) | 1)
						cur++
						tmp16 = uint16(((this.audioPts & 0x3fff8000) >> 14) | 1)
						tsBuf[cur] = byte((tmp16 >> 8) & 0xff)
						cur++
						tsBuf[cur] = byte(tmp16 & 0xff)
						cur++
						tmp16 = uint16((this.audioPts&0x7fff)<<1) | 1
						tsBuf[cur] = byte((tmp16 >> 8) & 0xff)
						cur++
						tsBuf[cur] = byte(tmp16 & 0xff)
						cur++
					} else {

						tsBuf[cur] = 0xe0
						cur++
						tsBuf[cur] = 0x00
						cur++
						tsBuf[cur] = 0x00
						cur++
						tsBuf[cur] = 0x80
						cur++
						tsBuf[cur] = 0xc0
						cur++
						tsBuf[cur] = 0x0a
						cur++

						tsBuf[cur] = byte((3 << 4) | ((pcr & 0x1c0000000) >> 29) | 1)
						cur++
						tmp16 = uint16(((pcr & 0x3fff8000) >> 14) | 1)
						tsBuf[cur] = byte((tmp16 >> 8) & 0xff)
						cur++
						tsBuf[cur] = byte(tmp16 & 0xff)
						cur++
						tmp16 = uint16(((pcr & 0x7fff) << 1) | 1)
						tsBuf[cur] = byte((tmp16 >> 8) & 0xff)
						cur++
						tsBuf[cur] = byte(tmp16 & 0xff)
						cur++
						tsBuf[cur] = byte((1 << 4) | ((pcr & 0x1c0000000) >> 29) | 1)
						cur++
						tmp16 = uint16(((pcr & 0x3fff8000) >> 14) | 1)
						tsBuf[cur] = byte((tmp16 >> 8) & 0xff)
						cur++
						tsBuf[cur] = byte(tmp16 & 0xff)
						cur++
						tmp16 = uint16(((pcr & 0x7fff) << 1) | 1)
						tsBuf[cur] = byte((tmp16 >> 8) & 0xff)
						cur++
						tsBuf[cur] = byte(tmp16 & 0xff)
						cur++
					}
					//!PES头
					copy(tsBuf[cur:], dataPayload[payloadCur:TS_length-cur])
					payloadCur += TS_length - cur
					this.appendTsPkt(tsBuf)
				} else {
					//四字节头
					tsBuf[cur] = 0x47
					cur++
					if flv.FLV_TAG_Audio == tag.TagType {
						tmp16 = uint16(Audio_Id)
					} else {
						tmp16 = uint16(Video_Id)
					}
					tsBuf[cur] = byte(tmp16 >> 8)
					cur++
					tsBuf[cur] = byte(tmp16 & 0xff)
					cur++
					//!3字节头
					if i == tsCount-1 && padSize != 0 {
						//最后一帧，且有pad
						if flv.FLV_TAG_Audio == tag.TagType {
							tsBuf[cur] = byte(0x30 | this.tsAcount)
							cur++
						} else {
							tsBuf[cur] = byte(0x30 | this.tsVcount)
							cur++
						}
						tsBuf[cur] = byte(padSize - 1)
						cur++
						if padSize != 1 {
							tsBuf[cur] = 0
							cur++
						}
						copy(tsBuf[4+padSize:], dataPayload[payloadCur:payloadCur+TS_length-4-padSize])
						payloadCur += TS_length - 4 - padSize
					} else {
						//普通添加数据
						if flv.FLV_TAG_Audio == tag.TagType {
							tsBuf[cur] = byte(0x10 | this.tsAcount)
							cur++
						} else {
							tsBuf[cur] = byte(0x10 | this.tsVcount)
							cur++
						}

						tmps := dataPayload[payloadCur : payloadCur+TS_length-cur]
						copy(tsBuf[cur:], tmps)
						payloadCur += TS_length - cur
					}
					if flv.FLV_TAG_Audio == tag.TagType {
						this.tsAcount++
						if this.tsAcount == 16 {
							this.tsAcount = 0
						}
					} else {
						this.tsVcount++
						if this.tsVcount == 16 {
							this.tsVcount = 0
						}
					}
					this.appendTsPkt(tsBuf)
				}
			}
		}
	}
}

func (this *TsCreater) GetDuration() (sec int) {
	return int(this.nowTime - this.beginTime)
}

func (this *TsCreater) FlushTsList() (tsList *list.List) {
	tsList = this.tsCache
	this.tsCache = list.New()
	return tsList
}

func (this *TsCreater) avHeaderAdded(tag *flv.FlvTag) (headerGeted bool) {
	if this.audioHeader != nil && this.videoHeader != nil {
		return true
	}
	this.beginTime = 0xffffffff
	if tag.TagType == flv.FLV_TAG_Audio {
		if this.audioHeader != nil {
			//防止只有音频的情况
			return true
		}
		this.audioHeader = make([]byte, len(tag.Data))
		copy(this.audioHeader, tag.Data)
		this.parseAudioType(this.audioHeader)
		return false
	}
	if tag.TagType == flv.FLV_TAG_Video {
		if this.videoHeader != nil {
			//防止没有音频的情况
			return true
		}
		this.videoHeader = make([]byte, len(tag.Data))
		copy(this.videoHeader, tag.Data)
		this.videoTypeId = 0x1b
		this.parseAVC(this.videoHeader)
		return false
	}
	return false
}

func (this *TsCreater) parseAudioType(data []byte) {
	audioCodec := data[0] >> 4
	switch audioCodec {
	case flv.SoundFormat_AAC:
		this.audioFrameSize = 1024
		this.asc = aac.GenerateAudioSpecificConfig(data[2:])
		this.audioSampleHz = int(this.asc.SamplingFrequency)
		this.audioTypeId = 0x0f
	case flv.SoundFormat_MP3:
		this.audioFrameSize = 1152
		mp3Header, err := mp3.ParseMP3Header(data[1:])
		if err != nil {
			logger.LOGE(err.Error())
			return
		}
		this.audioSampleHz = mp3Header.SampleRate
		if mp3Header.Version == 3 {
			this.audioTypeId = 0x03
		} else {
			this.audioTypeId = 0x04
		}
	default:
		logger.LOGE("ts audio type not supported", audioCodec)
		return
	}

}

func (this *TsCreater) parseAVC(data []byte) {
	if data[0] == 0x17 && data[1] == 0 {
		//avc
		this.sps, this.pps = h264.GetSpsPpsFromAVC(data[5:])
	}
}

func (this *TsCreater) addPatPmt() {
	cur := 0
	var tmp16 uint16
	var tmp32 uint32
	tsBuf := make([]byte, TS_length)
	for idx := 0; idx < TS_length; idx++ {
		tsBuf[idx] = 0xff
	}
	//pat
	tsBuf[cur] = 0x47
	cur++
	tsBuf[cur] = 0x40
	cur++
	tsBuf[cur] = 0x00
	cur++
	tsBuf[cur] = 0x10
	cur++

	tsBuf[cur] = 0x00 //0个补充字节
	cur++

	tsBuf[cur] = 0x00 //table id
	cur++
	tmp16 = (((0xb0) << 8) | 0xd) //section length
	tsBuf[cur] = byte((tmp16 >> 8) & 0xff)
	cur++
	tsBuf[cur] = byte((tmp16 >> 0) & 0xff)
	cur++
	tsBuf[cur] = 0x00 //transport stream id
	cur++
	tsBuf[cur] = 0x01
	cur++
	tsBuf[cur] = 0xc1 //vesion 0,current valid
	cur++
	tsBuf[cur] = 0x00 //section num
	cur++
	tsBuf[cur] = 0x00 //last section num
	cur++
	tsBuf[cur] = 0x00 //program num
	cur++
	tsBuf[cur] = 0x01
	cur++
	tmp16 = (0xe000 | PMT_ID) //PMT id
	tsBuf[cur] = byte((tmp16 >> 8) & 0xff)
	cur++
	tsBuf[cur] = byte((tmp16 >> 0) & 0xff)
	cur++
	tmp32 = Crc32Calculate(tsBuf[5:cur]) //CRC
	tsBuf[cur] = byte((tmp32 >> 24) & 0xff)
	cur++
	tsBuf[cur] = byte((tmp32 >> 16) & 0xff)
	cur++
	tsBuf[cur] = byte((tmp32 >> 8) & 0xff)
	cur++
	tsBuf[cur] = byte((tmp32 >> 0) & 0xff)
	cur++

	this.appendTsPkt(tsBuf)
	//pmt
	tsBuf = make([]byte, TS_length)
	for idx := 0; idx < TS_length; idx++ {
		tsBuf[idx] = 0xff
	}
	cur = 0

	tsBuf[cur] = 0x47
	cur++
	tmp16 = ((0x40 << 8) | PMT_ID)
	tsBuf[cur] = byte((tmp16 >> 8) & 0xff)
	cur++
	tsBuf[cur] = byte((tmp16 >> 0) & 0xff)
	cur++
	tsBuf[cur] = 0x10
	cur++

	tsBuf[cur] = 0x00 //0个补充字节
	cur++

	tsBuf[cur] = 0x02 //table id
	cur++

	tmp16 = ((0xb0 << 8) | 0x17) //section length
	tsBuf[cur] = byte((tmp16 >> 8) & 0xff)
	cur++
	tsBuf[cur] = byte((tmp16 >> 0) & 0xff)
	cur++
	tsBuf[cur] = 0x00 //transport stream id
	cur++
	tsBuf[cur] = 0x01
	cur++
	tsBuf[cur] = 0xc1 //vesion 0,current valid
	cur++
	tsBuf[cur] = 0x00 //section num
	cur++
	tsBuf[cur] = 0x00 //last section num
	cur++
	tmp16 = (0xe000 | Video_Id) //pcr pid
	tsBuf[cur] = byte((tmp16 >> 8) & 0xff)
	cur++
	tsBuf[cur] = byte((tmp16 >> 0) & 0xff)
	cur++
	tmp16 = 0xf000 //program info length = 0
	tsBuf[cur] = byte((tmp16 >> 8) & 0xff)
	cur++
	tsBuf[cur] = byte((tmp16 >> 0) & 0xff)
	cur++
	//video
	tsBuf[cur] = byte(this.videoTypeId)
	cur++
	tmp16 = (0xe000 | Video_Id)
	tsBuf[cur] = byte((tmp16 >> 8) & 0xff)
	cur++
	tsBuf[cur] = byte((tmp16 >> 0) & 0xff)
	cur++
	tsBuf[cur] = 0xf0
	cur++
	tsBuf[cur] = 0x00
	cur++
	//audio
	tsBuf[cur] = byte(this.audioTypeId)
	cur++
	tmp16 = (0xe000 | Audio_Id)
	tsBuf[cur] = byte((tmp16 >> 8) & 0xff)
	cur++
	tsBuf[cur] = byte((tmp16 >> 0) & 0xff)
	cur++
	tsBuf[cur] = 0xf0
	cur++
	tsBuf[cur] = 0x00
	cur++

	tmp32 = Crc32Calculate(tsBuf[5:cur]) //CRC
	tsBuf[cur] = byte((tmp32 >> 24) & 0xff)
	cur++
	tsBuf[cur] = byte((tmp32 >> 16) & 0xff)
	cur++
	tsBuf[cur] = byte((tmp32 >> 8) & 0xff)
	cur++
	tsBuf[cur] = byte((tmp32 >> 0) & 0xff)
	cur++

	this.appendTsPkt(tsBuf)
}

func (this *TsCreater) appendTsPkt(tsBuf []byte) {
	tmp := make([]byte, len(tsBuf))
	copy(tmp, tsBuf)
	this.tsCache.PushBack(tmp)
}

func (this *TsCreater) getTsCount(dataSize int, addPCR, addDts bool) (tsCount, padSize int) {
	firstValidSize := TS_length - 4
	if addPCR {
		firstValidSize -= 8
	}
	if addDts {
		firstValidSize -= 19
	} else {
		firstValidSize -= 14
	}
	validSize := TS_length - 4

	if dataSize <= firstValidSize {
		tsCount = 1
		padSize = firstValidSize - dataSize
		return tsCount, padSize
	} else {
		size := dataSize
		size -= firstValidSize
		tsCount = size/validSize + 1
		padSize = size % validSize
		if padSize != 0 {
			tsCount++
			padSize = validSize - padSize
		}
		return tsCount, padSize
	}

	return
}

func (this *TsCreater) videoPayload(tag *flv.FlvTag) (payload []byte) {
	if tag.Data[0] == 0x17 && tag.Data[1] == 0 {
		this.parseAVC(tag.Data)
		return nil
	}
	nalCur := 5
	getKeyframe := false
	nalList := list.New()
	totalNalSize := 0
	for nalCur < len(tag.Data) {
		nalSize := 0
		nalSizeSlice := tag.Data[nalCur : nalCur+4]
		nalSize = (int(nalSizeSlice[0]) << 24) | (int(nalSizeSlice[1]) << 16) |
			(int(nalSizeSlice[2]) << 8) | (int(nalSizeSlice[3]) << 0)
		nalCur += 4
		nalType := tag.Data[nalCur] & 0x1f

		switch nalType {
		case h264.Nal_type_sei:
			this.sei = make([]byte, nalSize)
			copy(this.sei, tag.Data[nalCur:nalCur+nalSize])
		case h264.Nal_type_sps:
			this.sps = make([]byte, nalSize)
			copy(this.sps, tag.Data[nalCur:nalCur+nalSize])
		case h264.Nal_type_pps:
			this.pps = make([]byte, nalSize)
			copy(this.pps, tag.Data[nalCur:nalCur+nalSize])
		case h264.Nal_type_idr:
			getKeyframe = true
			this.keyframeWrited = true
			totalNalSize += nalSize + 4
			tmp := make([]byte, nalSize)
			copy(tmp, tag.Data[nalCur:nalCur+nalSize])
			nalList.PushBack(tmp)
		default:
			totalNalSize += nalSize + 4
			tmp := make([]byte, nalSize)
			copy(tmp, tag.Data[nalCur:nalCur+nalSize])
			nalList.PushBack(tmp)
		}
		nalCur += nalSize
	}

	if false == getKeyframe && this.keyframeWrited == false {
		logger.LOGE("no keyframe")
		return nil
	}

	if nalList.Len() == 0 {
		logger.LOGE("no frame")
		return nil
	}

	if getKeyframe {
		payloadSize := totalNalSize + 6

		if len(this.sps) > 0 {
			payloadSize += len(this.sps) + 4
		}
		if len(this.pps) > 0 {
			payloadSize += len(this.pps) + 4
		}
		if len(this.sei) > 0 {
			payloadSize += len(this.sei) + 4
		}
		tmp32 := 0
		payload = make([]byte, payloadSize)
		payload[tmp32] = 0x00
		tmp32++
		payload[tmp32] = 0x00
		tmp32++
		payload[tmp32] = 0x00
		tmp32++
		payload[tmp32] = 0x01
		tmp32++
		payload[tmp32] = 0x09
		tmp32++
		payload[tmp32] = 0x10
		tmp32++

		if len(this.sps) > 0 {
			payload[tmp32] = 0x00
			tmp32++
			payload[tmp32] = 0x00
			tmp32++
			payload[tmp32] = 0x00
			tmp32++
			payload[tmp32] = 0x01
			tmp32++

			copy(payload[tmp32:], this.sps)
			tmp32 += len(this.sps)
		}

		if len(this.pps) > 0 {
			payload[tmp32] = 0x00
			tmp32++
			payload[tmp32] = 0x00
			tmp32++
			payload[tmp32] = 0x00
			tmp32++
			payload[tmp32] = 0x01
			tmp32++
			copy(payload[tmp32:], this.pps)
			tmp32 += len(this.pps)
		}

		if len(this.sei) > 0 {
			payload[tmp32] = 0x00
			tmp32++
			payload[tmp32] = 0x00
			tmp32++
			payload[tmp32] = 0x00
			tmp32++
			payload[tmp32] = 0x01
			tmp32++
			copy(payload[tmp32:], this.sei)
			tmp32 += len(this.sei)
		}

		for e := nalList.Front(); e != nil; e = e.Next() {
			buf := e.Value.([]byte)
			payload[tmp32] = 0x00
			tmp32++
			payload[tmp32] = 0x00
			tmp32++
			payload[tmp32] = 0x00
			tmp32++
			payload[tmp32] = 0x01
			tmp32++
			copy(payload[tmp32:], buf)
			tmp32 += len(buf)
		}
	} else {
		payloadSize := totalNalSize + 6
		payload = make([]byte, payloadSize)
		tmp32 := 0
		payload[tmp32] = 0x00
		tmp32++
		payload[tmp32] = 0x00
		tmp32++
		payload[tmp32] = 0x00
		tmp32++
		payload[tmp32] = 0x01
		tmp32++
		payload[tmp32] = 0x09
		tmp32++
		payload[tmp32] = 0x10
		tmp32++

		for e := nalList.Front(); e != nil; e = e.Next() {
			buf := e.Value.([]byte)
			payload[tmp32] = 0x00
			tmp32++
			payload[tmp32] = 0x00
			tmp32++
			payload[tmp32] = 0x00
			tmp32++
			payload[tmp32] = 0x01
			tmp32++

			copy(payload[tmp32:], buf)
			tmp32 += len(buf)
		}
	}

	return payload
}
