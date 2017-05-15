package j8583

import (
	"fmt"
	"errors"
	"strconv"
	"encoding/hex"
	"bytes"
	"8583/security"
	"8583/utils"
)

type MessageCalculator interface {
	calcMessage(message []byte) []byte
}

type Message struct {
	Tpdu         string
	Header       string
	Mti          string
	Bitmap       string
	Fields       []Field
	SecondBitmap bool
}

func (m *Message)SetField(i int, field Field) {
	if (i < 1 || i > m.fieldLength()) {
		return
	}
	m.Fields[i] = field;
}

func (m *Message)getField(i int) Field {
	if (i < 1 || i > m.fieldLength()) {
		return Field{}
	}
	return m.Fields[i]
}

func (m *Message)getFieldValue(i int) interface{} {
	if (i < 1 || i > m.fieldLength()) {
		return nil
	}
	return m.Fields[i].Value
}

func (m *Message)fieldLength() int {
	size := 64
	if (m.SecondBitmap) {
		size = size * 2;
	}
	return size;
}

func (m *Message) BytesFields() (ret []byte, err error) {
	mtiBytes := lbcd([]byte(m.Mti))
	ret = append(ret, mtiBytes...)
	byteNum := 8
	if m.SecondBitmap {
		byteNum = 16
	}
	bitmap := make([]byte, byteNum)
	data := make([]byte, 0, 512)

	for byteIndex := 0; byteIndex < byteNum; byteIndex++ {
		for bitIndex := 0; bitIndex < 8; bitIndex++ {

			i := byteIndex * 8 + bitIndex + 1

			// if we need second bitmap (additional 8 bytes) - set first bit in first bitmap
			if m.SecondBitmap && i == 1 {
				step := uint(7 - bitIndex)
				bitmap[byteIndex] |= (0x01 << step)
			}

			for i, f := range m.Fields {
				//判断不好
				if i < 2 || f.Value == nil {
					continue
				}

				// mark 1 in bitmap:
				step := uint(7 - bitIndex)
				bitmap[byteIndex] |= (0x01 << step)

				d, err := f.Bytes();
				if err != nil {
					return nil, err
				}

				data = append(data, d...)
			}
		}
	}

	ret = append(ret, bitmap...)
	ret = append(ret, data...)
	return ret, nil
}


// Bytes marshall Message to bytes
func (m *Message) Bytes(tdk string) (ret []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("Critical error:" + fmt.Sprint(r))
			ret = nil
		}
	}()

	ret = make([]byte, 0)

	tpduBytes, err := m.encodeTpdu()
	if err != nil {
		return nil, err
	}
	ret = append(ret, tpduBytes...)

	if len(m.Header) > 0 {
		headerBytes, err := m.encodeHeader()
		if err != nil {
			return nil, err
		}
		ret = append(ret, headerBytes...)
	}

	fieldsByte, err := m.BytesFields()
	if err != nil {
		return nil, err
	}

	if len(tdk) == 0 {
		ret = append(ret, fieldsByte...)
		return ret, nil
	}

	hexByte1, err := hex.DecodeString("E6")
	if err != nil {
		return nil, err
	}
	ret = append(ret, hexByte1...)
	if value42, ok := m.Fields[42].Value.(string); ok {
		ret = append(ret, []byte(value42)...)
	}
	if value41, ok := m.Fields[42].Value.(string); ok {
		ret = append(ret, []byte(value41)...)
	}
	ret = append(ret, []byte(fmt.Sprintf("%04d", len(fieldsByte)))...)
	ret = append(ret, []byte("000000000000")...)

	hexByte2, err := hex.DecodeString(tdk)
	if err != nil {
		return nil, err
	}

	encryBytes, err := security.EncryptWithDESKey(fieldsByte, hexByte2)
	if err != nil {
		return nil, err
	}

	ret = append(ret, encryBytes...)

	return ret, nil
}

func (m *Message) BytesLenHeader(tdk string) (ret []byte, err error) {
	data, err := m.Bytes(tdk)
	if err != nil {
		return nil, err
	}
	length := len(data)
	buf := bytes.NewBuffer(utils.Int2Byte((length & 0xff00) >> 8))
	buf.WriteByte((byte)((length & 0x00ff)))
	buf.Write(data)
	return buf.Bytes(), nil
}

func (m *Message) encodeTpdu() ([]byte, error) {
	if m.Tpdu == "" {
		return nil, errors.New("tpdu is required")
	}
	if len(m.Tpdu) != 10 {
		return nil, errors.New("tpdu is invalid")
	}

	// check MTI, it must contain only digits
	if _, err := strconv.Atoi(m.Tpdu); err != nil {
		return nil, errors.New("tpdu is invalid")
	}

	return bcd([]byte(m.Tpdu)), nil
}

func (m *Message) encodeHeader() ([]byte, error) {
	//TODO 待确定
	if len(m.Header) != 12 {
		return nil, errors.New("header is invalid")
	}

	// check MTI, it must contain only digits
	if _, err := strconv.Atoi(m.Header); err != nil {
		return nil, errors.New("header is invalid")
	}

	return bcd([]byte(m.Header)), nil
}

func Decode(raw []byte) (m *Message, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("Critical error:" + fmt.Sprint(r))
			m = nil
		}
	}()

	tpdu, err := decodeMti(raw[:5], BCD, 10)
	isoHeader, err := decodeMti(raw[5:11], BCD, 12)
	mti, err := decodeMti(raw[11:13], BCD, 4)
	bitmap := utils.EncodeToString(raw[13:21])
	m = &Message{Tpdu:tpdu, Mti:mti, Header:isoHeader, Bitmap:bitmap, SecondBitmap:false}

	fields := parseFields()

	byteNum := 8
	start := 21
	if raw[start] & 0x80 == 0x80 {
		// 1st bit == 1
		m.SecondBitmap = true
		byteNum = 16
	}
	bitByte := raw[start : start + byteNum]
	start += byteNum

	for byteIndex := 0; byteIndex < byteNum; byteIndex++ {
		for bitIndex := 0; bitIndex < 8; bitIndex++ {
			step := uint(7 - bitIndex)
			if (bitByte[byteIndex] & (0x01 << step)) == 0 {
				continue
			}

			i := byteIndex * 8 + bitIndex + 1
			if i == 1 {
				// field 1 is the second bitmap
				continue
			}
			f, ok := fields[i]
			if !ok {
				return nil, fmt.Errorf("field %d not defined", i)
			}

			l, err := f.load(raw[start:])
			if err != nil {
				return nil, fmt.Errorf("field %d: %s", i, err)
			}
			start += l
		}
	}
	return m, err
}

func DecodeDes(raw []byte, tdk string) (m *Message, err error) {
	minSize := 10 / 2 + 10 % 2 + 12 / 2 + 12 % 2 + 1 / 2 + 1 + 39
	if len(raw) <= 0 {
		return nil, errors.New("buf size is not enough")
	}

	if len(tdk) <= 0 {
		return nil, errors.New("messageCalculator should not be null")
	}

	hexByte, err := hex.DecodeString(tdk)
	if err != nil {
		return nil, err
	}

	encryBytes, err := security.EncryptWithDESKey(raw[minSize:], hexByte)
	if err != nil {
		return nil, err
	}

	data := append(raw[:minSize - 1 - 39], encryBytes...)

	return Decode(data)
}

func parseFields() map[int]*Field {
	fieldmap := make(map[int]*Field, 0)
	fieldmap[2] = &Field{IsoType:LLVAR, Encoder:BCD, }
	fieldmap[3] = &Field{IsoType:FIXED, Encoder:BCD, Length:6, }
	fieldmap[4] = &Field{IsoType:FIXED, Encoder:BCD, Length:12, }
	fieldmap[6] = &Field{IsoType:FIXED, Encoder:BCD, Length:12, }
	fieldmap[10] = &Field{IsoType:FIXED, Encoder:BCD, Length:8, }
	fieldmap[11] = &Field{IsoType:FIXED, Encoder:BCD, Length:6, }
	fieldmap[12] = &Field{IsoType:FIXED, Encoder:BCD, Length:6, }
	fieldmap[13] = &Field{IsoType:FIXED, Encoder:BCD, Length:4, }
	fieldmap[14] = &Field{IsoType:FIXED, Encoder:BCD, Length:4, }
	fieldmap[15] = &Field{IsoType:FIXED, Encoder:BCD, Length:4, }
	fieldmap[22] = &Field{IsoType:FIXED, Encoder:BCD, Length:3, }
	fieldmap[23] = &Field{IsoType:FIXED, Encoder:rBCD, Length:3, }
	fieldmap[25] = &Field{IsoType:FIXED, Encoder:BCD, Length:2, }
	fieldmap[26] = &Field{IsoType:FIXED, Encoder:BCD, Length:2, }

	fieldmap[32] = &Field{IsoType:LLVAR, Encoder:BCD, }
	fieldmap[35] = &Field{IsoType:LLVAR, Encoder:BCD, }

	fieldmap[37] = &Field{IsoType:FIXED, Encoder:ASCII, Length:12, }
	fieldmap[38] = &Field{IsoType:FIXED, Encoder:ASCII, Length:6, }
	fieldmap[39] = &Field{IsoType:FIXED, Encoder:ASCII, Length:2, }
	fieldmap[41] = &Field{IsoType:FIXED, Encoder:ASCII, Length:8, }
	fieldmap[42] = &Field{IsoType:FIXED, Encoder:ASCII, Length:15, }

	fieldmap[44] = &Field{IsoType:LLVAR, Encoder:BCD, }
	fieldmap[46] = &Field{IsoType:LLLVAR, Encoder:BCD, }
	fieldmap[48] = &Field{IsoType:LLLVAR, Encoder:BCD, }

	fieldmap[49] = &Field{IsoType:FIXED, Encoder:ASCII, Length:3, }
	fieldmap[51] = &Field{IsoType:FIXED, Encoder:ASCII, Length:3, }
	fieldmap[52] = &Field{IsoType:FIXED, Encoder:BINARY, Length:8, }
	fieldmap[53] = &Field{IsoType:FIXED, Encoder:BCD, Length:16, }

	fieldmap[54] = &Field{IsoType:LLLVAR, Encoder:ASCII, }
	fieldmap[55] = &Field{IsoType:LLLVAR, Encoder:BINARY, }
	fieldmap[57] = &Field{IsoType:LLLVAR, Encoder:ASCII, }

	subField60 := make([]*SubField, 5)
	subField60[0] = &SubField{IsoType:FIXED, Encoder:BCD, Length:2, }
	subField60[1] = &SubField{IsoType:FIXED, Encoder:BCD, Length:6, }
	subField60[2] = &SubField{IsoType:FIXED, Encoder:BCD, Length:3, }
	subField60[3] = &SubField{IsoType:FIXED, Encoder:BCD, Length:1, }
	subField60[4] = &SubField{IsoType:FIXED, Encoder:BCD, Length:1, }
	fieldmap[60] = NewSubField(LLLVAR, BCD, subField60)

	subField61 := make([]*SubField, 3)
	subField61[0] = &SubField{IsoType:FIXED, Encoder:BCD, Length:6, }
	subField61[1] = &SubField{IsoType:FIXED, Encoder:BCD, Length:6, }
	subField61[2] = &SubField{IsoType:FIXED, Encoder:BCD, Length:4, }
	fieldmap[61] = NewSubField(LLLVAR, BCD, subField61)

	fieldmap[62] = &Field{IsoType:LLLVAR, Encoder:BINARY, }

	subField63 := make([]*SubField, 1)
	subField63[0] = &SubField{IsoType:FIXED, Encoder:BCD, Length:3, }
	fieldmap[63] = NewSubField(LLLVAR, BCD, subField63)

	fieldmap[64] = &Field{IsoType:FIXED, Encoder:BINARY, Length:8, }
	return fieldmap
}

func decodeMti(raw []byte, encode int, length int) (string, error) {
	if encode == BCD {
		length = length / 2
	}
	if len(raw) < length {
		return "", errors.New("bad raw data")
	}

	var result string
	switch encode {
	case ASCII:
		result = string(raw[:length])
	case BCD:
		result = string(bcd2Ascii(raw[:length]))
	default:
		return "", errors.New("invalid encode type")
	}
	return result, nil
}







