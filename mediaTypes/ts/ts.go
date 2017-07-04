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
	PCH_HZ    = 27000000
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
	tsCache        list.List
}

func (this *TsCreater) AddTag(tag *flv.FlvTag) {
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
		var tmp32 uint32
		var tmp16 uint16

		var dataPayload []byte
		var payloadSize int

		if flv.FLV_TAG_Audio == tag.TagType {

		} else {

		}
	}
}

func (this *TsCreater) GetDuration() (sec int) {
	return
}

func (this *TsCreater) GetData() (data []byte) {
	return
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
	tmp32 = Crc32Calculate(tsBuf[5:]) //CRC
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

	tmp32 = Crc32Calculate(tsBuf[5:]) //CRC
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
	this.tsCache.PushBack(tsBuf)
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
		padSize = firstValidSize - padSize
		return tsCount, padSize
	} else {
		dataSize -= firstValidSize
		tsCount = dataSize/validSize + 1
		padSize = dataSize % validSize
		if padSize != 0 {
			tsCount++
			padSize = validSize - padSize
		}
		return tsCount, padSize
	}
	return
}
