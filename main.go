package main

import (
	"8583/j8583"
	"errors"
	"8583/security"
	"bytes"
	"fmt"
	"strconv"
	"encoding/hex"
	"net"
	"io"
	"8583/utils"
)

func main() {
	m := &j8583.Message{Header:"602200000000"}
	macKey := "1CDC70ABD616015E"
	tdk := "4551E676DFEFE6109252683B64B66E1F"

	//0 = {HashMap$HashMapEntry@5782} "F25_POS_COND_CODE" -> "31"
	//1 = {HashMap$HashMapEntry@5783} "F23_CARD_SERIAL" -> "001"
	//2 = {HashMap$HashMapEntry@5784} "F4_AMOUNT" -> "000000000001"
	//3 = {HashMap$HashMapEntry@5785} "F2_CARD_NUM" -> "000000000001"
	//4 = {HashMap$HashMapEntry@5786} "C2" -> "null"
	//5 = {HashMap$HashMapEntry@5787} "F22_POS_INPUT_STYLE" -> "040"
	//6 = {HashMap$HashMapEntry@5788} "F62_TERMINAL_STATUS" -> "284753193293963468"

	buildSCanCodeMessage(m, "6004010000", "000000000001", "000025", "000001", "00003042", "666100041213175",
		"284753193293963468", "", macKey)

	j8583.PrintMessage(m)
	data, err := m.BytesLenHeader(tdk)
	if err != nil {
		fmt.Println("Error: %s", err.Error())
	}

	//j8583.PrintMessage(m)

	//conn, err := net.DialTimeout("tcp", "192.168.1.102:5811", 30 * time.Second)
	conn, err := net.Dial("tcp", "192.168.1.102:5811")
	if err != nil {
		fmt.Println("Error: %s", err.Error())
	}
	defer conn.Close()

	conn.Write(data)

	var buf bytes.Buffer

	_, err = io.Copy(&buf, conn)
	if err != nil {
		fmt.Println("Error: %s", err.Error())
	}


	//
	//
	//
	//
	//
	//
	//
	//result := make([]byte, 0)
	//
	//conn.Read(result)

	mes, err := j8583.DecodeDes(buf.Bytes(), tdk)
	if err != nil {
		fmt.Println("Error: %s", err.Error())
	}

	j8583.PrintMessage(mes)
}

func buildSCanCodeMessage(m *j8583.Message, tpdu, amount, serialNum, batchNum, terminalID, merchantID, scanCodeId, extOrder, mac string) {
	m.Tpdu = tpdu
	m.Mti = "0200"
	m.Fields = make([]j8583.Field, 65)
	m.Fields[3] = j8583.NewFieldFix(j8583.BCD, 6, "000000")
	m.Fields[4] = j8583.NewFieldFix(j8583.BCD, 12, amount)
	m.Fields[11] = j8583.NewFieldFix(j8583.BCD, 6, serialNum)
	m.Fields[22] = j8583.NewFieldFix(j8583.BCD, 3, "040")
	m.Fields[23] = j8583.NewFieldFix(j8583.BCD, 3, "001")
	m.Fields[25] = j8583.NewFieldFix(j8583.BCD, 2, "31")
	m.Fields[41] = j8583.NewFieldFix(j8583.ASCII, 8, terminalID)
	m.Fields[42] = j8583.NewFieldFix(j8583.ASCII, 15, merchantID)
	m.Fields[49] = j8583.NewFieldFix(j8583.ASCII, 3, "156")

	if len(extOrder) > 0 {
		m.Fields[57] = buildField57(extOrder)
	}

	subField60 := make([]j8583.SubField, 5)
	subField60[0] = j8583.NewSubFieldFix(j8583.BCD, 2, "00")
	subField60[1] = j8583.NewSubFieldFix(j8583.BCD, 6, batchNum)
	subField60[2] = j8583.NewSubFieldFix(j8583.BCD, 3, "003")
	subField60[3] = j8583.NewSubFieldFix(j8583.BCD, 1, "0")
	subField60[4] = j8583.NewSubFieldFix(j8583.BCD, 1, "0")
	m.Fields[60] = j8583.NewFields(j8583.LLLVAR, j8583.BCD, subField60)
	m.Fields[62] = j8583.NewFieldVar(j8583.LLLVAR, j8583.BCD, scanCodeId)
	buildField64(m, mac)
}

func buildField57(extOrder string) j8583.Field {
	field := j8583.Field{IsoType:j8583.LLLVAR, Encoder:j8583.BINARY}
	var buf bytes.Buffer
	buf.WriteString(utils.EncodeToString([]byte("UPLDC2")))

	value := utils.EncodeToString([]byte(extOrder))
	length := strconv.Itoa(len(value) / 2)
	switch len(length) {
	case 1:
		length = "00" + length
	case 2:
		length = "0" + length
	}

	buf.WriteString(utils.EncodeToString([]byte(extOrder)))
	buf.WriteString(value)
	value = buf.String()
	field.Length = len(value) / 2
	field.Value = value
	return field;
}

func buildField64(m *j8583.Message, key string) error {
	data, err := m.BytesFields()
	if err != nil {
		return err
	}
	data=[]byte("test")
	macBytes, err := calcMac(data, key)
	if err != nil {
		return err
	}
	mac := utils.EncodeToString(macBytes)
	m.Fields[64] = j8583.NewFieldFix(j8583.BINARY, 8, mac)
	return nil
}

func calcMac(raw []byte, key string) ([]byte, error) {
	mak, err := hex.DecodeString(key)
	if err != nil {
		return nil, err
	}
	return getMac(mak, raw)
}

func getMac(mak, mab[]byte) ([]byte, error) {
	if len(mak) != 8 {
		return nil, errors.New("input MAK must has 8 bytes")
	}
	if len(mab) <= 0 {
		return nil, errors.New("input mab should not be empty")
	}

	if res := len(mab) % 8; res != 0 {
		temp := make([]byte, res)
		mab = append(mab, temp...)
	}

	n := len(mab) / 8
	result := mab[:8]
	for i := 1; i < n; i++ {
		for j := 0; j < 8; j++ {
			result[j] = result[j] ^ mab[i * 8 + j]
		}
	}

	hexDecBytes := []byte(utils.EncodeToString(result))
	var err error
	result, err = security.EncryptWithDESKey(hexDecBytes[:8], mak)
	if err != nil {
		return nil, err
	}

	for i := 0; i < 8; i++ {
		result[i] = byte(result[i] ^ hexDecBytes[8 + i])
	}
	result, err = security.EncryptWithDESKey(result, mak)
	if err != nil {
		return nil, err
	}

	hexDecBytes = []byte(utils.EncodeToString(result))
	return hexDecBytes[:8], nil
}


