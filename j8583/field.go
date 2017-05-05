package j8583

const (
	ASCII = iota
	BINARY
	BCD
	rBCD
)

type CustomField interface {
	decodeField(value string) interface{}
	encodeField(value interface{}) string
}

type Field struct {
	IsoType  int
	Value    interface{}
	Encoder  CustomField
	Length   int
	Encoding int
}

func (f *Field) Bytes() (data []byte, err error) {
	data := make([]byte, 0)

	switch f.IsoType {
	case VAR:
	case LLVAR:
	case LLLVAR:

	}

	switch f.Encoding {
	case ASCII:
		data = []byte(f.Value)
	case BINARY:

	case BCD:
	case rBCD:

	}

	return nil, nil;
}