package j8583

import (
	"fmt"
	"errors"
	"strconv"
)

//const (
//	/** A fixed-length numeric value. It is zero-filled to the left. */
//	NUMERIC = iota
//	/** A fixed-length alphanumeric value. It is filled with spaces to the right. */
//	ALPHA
//	/** A variable length alphanumeric value with a 2-digit header length. */
//	LLVAR
//	/** A variable length alphanumeric value with a 3-digit header length. */
//	LLLVAR
//	/** A date in format YYYYMMddHHmmss */
//	DATE14
//	/** A date in format MMddHHmmss */
//	DATE10
//	/** A date in format MMdd */
//	DATE4
//	/** A date in format yyMM */
//	DATE_EXP
//	/** Time of day in format HHmmss */
//	TIME
//	/** An amount, expressed in cents with a fixed length of 12. */
//	AMOUNT
//	/** Similar to ALPHA but holds byte arrays instead of strings. */
//	BINARY
//	/** Similar to LLVAR but holds byte arrays instead of strings. */
//	LLBIN
//	/** Similar to LLLVAR but holds byte arrays instead of strings. */
//	LLLBIN
//	/** variable length with 4-digit header length. */
//	LLLLVAR
//	/** variable length byte array with 4-digit header length. */
//	LLLLBIN
//	/** Date in format yyMMddHHmmss. */
//	DATE12
//)

type MessageCalculator interface {
	calcMessage(message []byte) []byte
}

type Message struct {
	Tpdu         string
	Header       string
	Mti          string
	Fields       []Field
	SecondBitmap bool
}

//public IsoField getValue(int index) {
//if (index < 1 || index > 128) {
//return null;
//}
//return fields[index];
//}
//
//public void setValue(int index, IsoField isoField) {
//if (index < 1 || index > 128) {
//return;
//}
//fields[index] = isoField;
//}

func (m *Message)SetField(i int, field Field) {
	if (i < 1 || i > m.fieldLength()) {
		return
	}
	m.Fields[i] = field;
}

func (m *Message)getField(i int) Field {
	if (i < 1 || i > m.fieldLength()) {
		return nil
	}
	return m.Fields[i]
}

func (m *Message)fieldLength() int {
	size := 64
	if (m.SecondBitmap) {
		size = size * 2;
	}
	return size;
}

func (m *Message)BytesWithLenHeader(ret []byte, err error) {
	data, error := m.Bytes();
	if error != nil {
		return nil, err
	}
	return bcd([]byte(len(data) + 2)), nil;
}


// Bytes marshall Message to bytes
func (m *Message) Bytes() (ret []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("Critical error:" + fmt.Sprint(r))
			ret = nil
		}
	}()

	ret = make([]byte, 0)

	// generate MTI:
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

	mtiBytes, err := m.encodeMti()
	if err != nil {
		return nil, err
	}
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
				if i < 2 || f == nil || f.Value == nil {
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

func (m *Message) encodeTpdu() ([]byte, error) {
	if m.Tpdu == "" {
		return nil, errors.New("tpdu is required")
	}
	if len(m.Tpdu) != 4 {
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

func (m *Message) encodeMti() ([]byte, error) {
	if m.Mti == "" {
		return nil, errors.New("MTI is required")
	}
	if len(m.Mti) != 4 {
		return nil, errors.New("MTI is invalid")
	}

	// check MTI, it must contain only digits
	if _, err := strconv.Atoi(m.Mti); err != nil {
		return nil, errors.New("MTI is invalid")
	}

	return bcd([]byte(m.Mti)), nil
}

// Load unmarshall Message from bytes
func (m *Message) Load(raw []byte) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("Critical error:" + fmt.Sprint(r))
		}
	}()

	if m.Mti == "" {
		m.Mti, err = decodeMti(raw, m.MtiEncode)
		if err != nil {
			return err
		}
	}
	start := 4
	if m.MtiEncode == BCD {
		start = 2
	}

	fields := parseFields(m.Data)

	byteNum := 8
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
				return fmt.Errorf("field %d not defined", i)
			}
			l, err := f.Field.Load(raw[start:], f.Encode, f.LenEncode, f.Length)
			if err != nil {
				return fmt.Errorf("field %d: %s", i, err)
			}
			start += l
		}
	}
	return nil
}





