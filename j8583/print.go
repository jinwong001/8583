package j8583

import (
	"fmt"
	"strconv"
)

func PrintMessage(m *Message) {
	fmt.Println("[j8583]----------begin---------")
	printField("H000D", m.Tpdu)
	printField("H001D", m.Header)
	printField("F000D", m.Mti)
	printField("F001D", m.Bitmap)

	for i, field := range m.Fields {
		if field.IsoType != FIXED {
			printField(fmt.Sprintf("F%03dL", i), strconv.Itoa(field.Length))
		}
		if value,ok:=field.Value.(string);ok{
			printField(fmt.Sprintf("F%03dD", i), value)
		}
	}
	fmt.Println("[j8583]----------end---------")
	fmt.Println("")
}

func printField(name, value string) {
	fmt.Println("[j8583]" + name + ": " + value)
}
