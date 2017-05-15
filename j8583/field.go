package j8583

import (
	"errors"
	"fmt"
	"strings"
	"strconv"
	"encoding/hex"
	"8583/utils"
)

const (
	ASCII = iota
	BINARY
	BCD
	rBCD
)

const (
	ERR_INVALID_ENCODER string = "invalid encoder"
	ERR_INVALID_LENGTH_ENCODER string = "invalid length encoder"
	ERR_INVALID_LENGTH_HEAD string = "invalid length head"
	ERR_MISSING_LENGTH string = "missing length"
	ERR_VALUE_TOO_LONG string = "length of value is longer than definition; type=%s, def_len=%d, len=%d"
	ERR_BAD_RAW string = "bad raw data"
	ERR_PARSE_LENGTH_FAILED string = "parse length head failed"
)

type Field struct {
	IsoType int
	Value   interface{}
	Encoder int
	Length  int
}

type SubField struct {
	IsoType int
	Value   string
	Encoder int
	Length  int
}


// NewNumeric create new Numeric field
func NewField(value  string) *Field {
	return &Field{Value:value}
}

//func NewFieldTLv(isoType, tlvs map[string]*SubField) *Field {
//	field := &Field{IsoType:isoType, Encoder:BINARY, }
//	var value string
//	for k, v := range tlvs {
//		value = value + k + v.Length + v.Value
//	}
//	field.Value = value
//	field.Length = len(value) / 2
//	return field
//}

func NewSubField(isoType, encoder int, subFields []*SubField) *Field {
	field := &Field{IsoType:isoType, Encoder:encoder, }
	var length int
	var value string
	for _, subField := range subFields {
		length = length + subField.Length
		value = value + subField.Value
	}
	field.Length = length
	field.Value = value
	return field
}

func NewFields(isoType, encoder int, subFields []SubField) Field {
	field := &Field{IsoType:isoType, Encoder:encoder, }
	length := 0
	var value string
	for _, subField := range subFields {
		length = length + subField.Length
		value = value + subField.Value
	}
	field.Length = length
	field.Value = value
	return *field
}



func NewFieldFix(encoder, length int, value  string) Field {
	return Field{IsoType:FIXED, Value:value, Encoder:encoder, Length:length}
}

func NewFieldVar(isoType, encoder int, value  string) Field {
	field := &Field{IsoType:isoType, Encoder:encoder, Value:value}
	if encoder==BINARY{
		field.Length=len(value)/2
	}else {
		field.Length=len(value)
	}
	return *field
}

func NewSubFieldFix(encoder, length int, value  string) SubField {
	return SubField{IsoType:FIXED, Value:value, Encoder:encoder, Length:length}
}

func (f *Field) Bytes() ([]byte, error) {
	data := make([]byte, 0)
	switch f.IsoType {
	case LLVAR:
		data = append(data, utils.Int2Byte(((f.Length / 10) << 4) | (f.Length % 10))...)
	case LLLVAR:
		data = append(data, utils.Int2Byte(f.Length / 100 + ((f.Length % 100) << 4) | (f.Length % 10))...)
	case LLLLVAR:
		data = append(data, utils.Int2Byte(((f.Length / 1000) << 4) | (f.Length / 100 % 10))...)
		data = append(data, utils.Int2Byte((((f.Length % 100) / 10) << 4) | (f.Length % 10))...)
	}

	if value, ok := f.Value.(string); ok {
		switch f.Encoder {
		case ASCII:
			data = append(data, []byte(value)...)
		case BINARY:
			hexByte, err := hex.DecodeString(value)
			if err != nil {
				return nil, err
			}
			data = append(data, hexByte...)
		case BCD:
			data = append(data, BCD2Byte(value)...)
		case rBCD:
			data = append(data, RBCD2Byte(value)...)
		}
	}

	if subFields, ok := f.Value.([]SubField); ok {
		for _, sub := range subFields {
			d, err := sub.Bytes();
			if err != nil {
				return nil, err
			}
			data = append(data, d...)
		}
	}
	return data, nil
}

func (f *SubField) Bytes() ([]byte, error) {
	data := make([]byte, 0)
	switch f.IsoType {
	case LLVAR:
		data = append(data, utils.Int2Byte(((f.Length / 10) << 4) | (f.Length % 10))...)
	case LLLVAR:
		data = append(data, utils.Int2Byte(f.Length / 100 + ((f.Length % 100) << 4) | (f.Length % 10))...)
	case LLLLVAR:
		data = append(data, utils.Int2Byte(((f.Length / 1000) << 4) | (f.Length / 100 % 10))...)
		data = append(data, utils.Int2Byte((((f.Length % 100) / 10) << 4) | (f.Length % 10))...)
	}

	switch f.Encoder {
	case ASCII:
		data = append(data, []byte(f.Value)...)
	case BINARY:
		hexByte, err := hex.DecodeString(f.Value)
		if err != nil {
			return nil, err
		}
		data = append(data, hexByte...)
	case BCD:
		data = append(data, BCD2Byte(f.Value)...)
	case rBCD:
		data = append(data, RBCD2Byte(f.Value)...)
	}
	return data, nil
}

func (f *SubField)load(raw []byte) (read int, err error) {
	var contentLen int
	switch f.IsoType {
	case LLVAR:
		read = 1
		contentLen, err = strconv.Atoi(string(bcdr2Ascii(raw[:read], 2)))
		if err != nil {
			return 0, errors.New(ERR_PARSE_LENGTH_FAILED + ": " + string(raw[0]))
		}
	case LLLVAR:
		read = 2
		contentLen, err = strconv.Atoi(string(bcdr2Ascii(raw[:read], 3)))
		if err != nil {
			return 0, errors.New(ERR_PARSE_LENGTH_FAILED + ": " + string(raw[:2]))
		}
	case LLLLVAR:
		read = 2
		contentLen, err = strconv.Atoi(string(bcdr2Ascii(raw[:read], 4)))
		if err != nil {
			return 0, errors.New(ERR_PARSE_LENGTH_FAILED + ": " + string(raw[:2]))
		}
	}

	// parse body:
	switch f.Encoder {
	case BINARY:
		if len(raw) < (read + contentLen) {
			return 0, errors.New(ERR_BAD_RAW)
		}
		f.Value = utils.EncodeToString(raw[read : read + contentLen])
		read += contentLen
	case ASCII:
		if len(raw) < (read + contentLen) {
			return 0, errors.New(ERR_BAD_RAW)
		}
		f.Value = string(raw[read : read + contentLen])
		read += contentLen
	case rBCD:
		fallthrough
	case BCD:
		bcdLen := (contentLen + 1) / 2
		if len(raw) < (read + bcdLen) {
			return 0, errors.New(ERR_BAD_RAW)
		}
		f.Value = string(bcdl2Ascii(raw[read:read + bcdLen], contentLen))
		read += bcdLen
	default:
		return 0, errors.New(ERR_INVALID_ENCODER)
	}
	return read, nil
}

func (f *Field)load(raw []byte) (read int, err error) {
	var contentLen int
	switch f.IsoType {
	case LLVAR:
		read = 1
		contentLen, err = strconv.Atoi(string(bcdr2Ascii(raw[:read], 2)))
		if err != nil {
			return 0, errors.New(ERR_PARSE_LENGTH_FAILED + ": " + string(raw[0]))
		}
	case LLLVAR:
		read = 2
		contentLen, err = strconv.Atoi(string(bcdr2Ascii(raw[:read], 3)))
		if err != nil {
			return 0, errors.New(ERR_PARSE_LENGTH_FAILED + ": " + string(raw[:2]))
		}
	case LLLLVAR:
		read = 2
		contentLen, err = strconv.Atoi(string(bcdr2Ascii(raw[:read], 4)))
		if err != nil {
			return 0, errors.New(ERR_PARSE_LENGTH_FAILED + ": " + string(raw[:2]))
		}
	}

	// parse body:
	switch f.Encoder {
	case BINARY:
		if len(raw) < (read + contentLen) {
			return 0, errors.New(ERR_BAD_RAW)
		}
		f.Value = utils.EncodeToString(raw[read : read + contentLen])
		read += contentLen
	case ASCII:
		if len(raw) < (read + contentLen) {
			return 0, errors.New(ERR_BAD_RAW)
		}
		f.Value = string(raw[read : read + contentLen])
		read += contentLen
	case rBCD:
		fallthrough
	case BCD:
		bcdLen := (contentLen + 1) / 2
		if len(raw) < (read + bcdLen) {
			return 0, errors.New(ERR_BAD_RAW)
		}
		f.Value = string(bcdl2Ascii(raw[read:read + bcdLen], contentLen))
		read += bcdLen
	default:
		return 0, errors.New(ERR_INVALID_ENCODER)
	}

	if value, ok := f.Value.([]SubField); ok {
		for _, subField := range value {
			length, error := subField.load(raw[read:])
			if error != nil {
				return read, error
			}
			read += length
		}
	}
	return read, nil
}

// Bytes encode Numeric field to bytes
func endcode(encoder, length int, value string) ([]byte, error) {
	val := []byte(value)
	if (encoder == rBCD) &&
		len(val) == (length + 1) &&
		(string(val[0:1]) == "0") {
		val = val[1:]
	}

	if len(val) > length {
		return nil, errors.New(fmt.Sprintf(ERR_VALUE_TOO_LONG, "Numeric", length, len(val)))
	}
	if len(val) < length {
		val = append([]byte(strings.Repeat("0", length - len(val))), val...)
	}
	switch encoder {
	case BCD:
		return lbcd(val), nil
	case rBCD:
		return rbcd(val), nil
	case ASCII:
		return val, nil
	default:
		return nil, errors.New(ERR_INVALID_ENCODER)
	}
}
