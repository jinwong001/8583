package main

import (
	"fmt"
	"github.com/ideazxy/iso8583"
	"encoding/hex"
	"strings"
)

type Data struct {
	No   *iso8583.Numeric      `field:"3" length:"6" encode:"bcd"`         // bcd value encoding
	Oper *iso8583.Numeric      `field:"26" length:"2" encode:"ascii"`      // ascii value encoding
	Ret  *iso8583.Alphanumeric `field:"39" length:"2"`
	Sn   *iso8583.Llvar        `field:"45" length:"23" encode:"bcd,ascii"` // bcd length encoding, ascii value encoding
	Info *iso8583.Lllvar       `field:"46" length:"42" encode:"bcd,ascii"`
	Mac  *iso8583.Binary       `field:"64" length:"8"`
}

type CILRequest struct {
	/**
          * 交易金额
          */
	amount             string
	/**
	 * 时间戳
	 */
	timestamp          string
	/**
	 * 商户号
	 */
	merCode            string
	/**
	 * 终端号
	 */
	termCode           string

	/**
	 * 激活码（新激活方式）
	 */
	activeCode         string

	/**
	 * 机器型号（目前 N900 、A8、其他）
	 */
	model              string
	/**
	 * App或SDK的版本号
	 */
	appVersion         string

	/**
	 * 服务端会验证,服务端那边有预先配置
	 */
	snCode             string
	/**
	 * 激活的时候需要上传,用于后续推送
	 */
	deviceToken        string
	/**
	 * 检查更新需要参数
	 */
	apkName            string

	/**
	 * 管理平台查询接口请求参数
	 * 页数
	 */
	page               int
	/**
	 * 管理平台查询接口请求参数
	 * 大小
	 */
	size               int

	/**
	 * 外部订单号 （用于服务平台账单）
	 */
	outOrderNum        string

	/**
	 * 结算请求参数
	 * 结算批次
	 */
	termBatchId        string

	/**
	 * 统计接口请求参数
	 * 统计类型
	 */
	txnType            int

	//交易信息

	/**
	 * 参考号 37域
	 */
	referenceNumber    string
	/**
	 * 原交易时间
	 */
	originalTradeDate  string

	/**
	 * 扫码号
	 */
	scanCodeId         string

	/**
	 * 授权码
	 */
	revAuthCode        string
	/**
	 * 批次号
	 */
	batchNum           string

	/**
	 * 追踪号 11域
	 */
	traceNum           string

	/**
	 * 卡号
	 */
	cardNum            string


	/**
	 * 扣账金额
	 */
	billingAmt         string

	/**
	 * 交易货币代码
	 */
	transCurr          string
	/**
	 * 持卡人扣帐货币代码
	 */
	billingCurr        string

	/**
	 * 交易时间
	 */
	transDatetime      string

	//从POS读取信用卡返回的相关信息

	/**
	 * 二磁道信息  EmvTransInfo
	 */
	secondTrack        []byte

	/**
	 * 55域 用于IC卡的插卡和挥卡
	 */
	field55            string

	/**
	 * 跟密码相关
	 */
	pinEmv             []byte

	/**
    * 从EmvTransInfo 获取的卡有效日期
    */
	cardExpirationDate string

	/**
	 * EmvTransInfo 获取 卡的序列号
	 */
	cardSequenceNumber string

	/**
	 * 电子签名所关联的交易类型
	 */
	original           string

	/**
	 * 签名图片的存储地址
	 */
	jbgPath            string

	/**
	 * 44域 附加响应数据
	 */
	additionalResData  string

	/**
	 * 13域 收卡方所在地日期
	 */
	localTransDate     string

	/**
	 * 12域 收卡方所在地时间
	 */
	localTransTime     string
	/**
	 * 63域
	 */
	reservedPrivate    string

	/**
	 * 交易货币代码
	 */
	transCurrCode      string

	/**
	 * 持卡人扣帐货币代码
	 */
	CardHolderCurrCode string

	serialNum          string

	/**
	 * 每一次查询的时间间隔
	 */
	period             int64

	/**
	 * 查询次数
	 */
	limitTime          int

	/**
	 * 外部订单号
	 * 长度小于64位
	 */
	orderId            string

	/**
         * pos输入服务方式码
         */
	posInputStyle      string
}

type TestISO struct {
	F2  *iso8583.Llnumeric    `field:"2" length:"19" encode:"bcd,rbcd"`
	F3  *iso8583.Numeric      `field:"3" length:"6" encode:"bcd"`
	F4  *iso8583.Numeric      `field:"4" length:"12" encode:"ascii"`
	F7  *iso8583.Numeric      `field:"7" length:"10" encode:"bcd"`
	F11 *iso8583.Numeric      `field:"11" length:"6" encode:"bcd"`
	F12 *iso8583.Numeric      `field:"12" length:"6" encode:"bcd"`
	F13 *iso8583.Numeric      `field:"13" length:"4" encode:"lbcd"`
	F14 *iso8583.Numeric      `field:"14" length:"4" encode:"lbcd"`
	F19 *iso8583.Numeric      `field:"19" length:"3" encode:"rbcd"`
	F22 *iso8583.Numeric      `field:"22" length:"3" encode:"rbcd"`
	F25 *iso8583.Numeric      `field:"25" length:"2" encode:"bcd"`
	F26 *iso8583.Numeric      `field:"26" length:"2" encode:"bcd"`
	F28 *iso8583.Alphanumeric `field:"28" length:"9"`
	F32 *iso8583.Llnumeric    `field:"32" length:"11" encode:"bcd,rbcd"`
	F35 *iso8583.Llnumeric    `field:"35" length:"37" encode:"rbcd,ascii"`
	F37 *iso8583.Alphanumeric `field:"37" length:"12"`
	F39 *iso8583.Alphanumeric `field:"39" length:"2"`
	F41 *iso8583.Alphanumeric `field:"41" length:"8"`
	F42 *iso8583.Alphanumeric `field:"42" length:"15"`
	F43 *iso8583.Alphanumeric `field:"43" length:"40"`
	F45 *iso8583.Llnumeric    `field:"45" length:"75" encode:"ascii,bcd"`
	F49 *iso8583.Numeric      `field:"49" length:"3" encode:"rbcd"`
	F52 *iso8583.Binary       `field:"52" length:"8"`
	F53 *iso8583.Numeric      `field:"53" length:"16" encode:"bcd"`
	F54 *iso8583.Llvar        `field:"54" length:"255" encode:"ascii,ascii"`
	F55 *iso8583.Llvar        `field:"55" length:"255" encode:"bcd,ascii"`
	F56 *iso8583.Lllvar       `field:"56" length:"255" encode:"bcd,ascii"`
	F57 *iso8583.Lllvar       `field:"57" length:"255" encode:"rbcd,ascii"`
	F58 *iso8583.Lllvar       `field:"58" length:"255" encode:"ascii,ascii"`
	F59 *iso8583.Llvar        `field:"59" length:"255" encode:"rbcd,ascii"`
	F60 *iso8583.Lllnumeric   `field:"60" length:"999" encode:"bcd,ascii"`
	F61 *iso8583.Lllnumeric   `field:"60" length:"999" encode:"bcd,rbcd"`
	F63 *iso8583.Lllnumeric   `field:"63" length:"999" encode:"rbcd,bcd"`
	F64 *iso8583.Binary       `field:"64" length:"32"`
}

func main() {
	requst := CILRequest{
		amount:"1",
		cardNum:"6222042600001000",
		cardExpirationDate:"4912",
		secondTrack:[]byte("3001156C887648310854"),
	}

	data := &TestISO{}
	buildSaleCommonRequest(data, requst)
	//TOOD 64 域处理
	msg := iso8583.NewMessage("0800", data)
	msg.MtiEncode = iso8583.BCD
	b, err := msg.Bytes()
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Printf("% x\n", b)
}

func buildSaleCommonRequest(data *TestISO, requst CILRequest) {
	//账号
	data.F2 = iso8583.NewLlnumeric("6222042600001000")
	//处理代码
	data.F3 = iso8583.NewNumeric("000000")
	//金额
	data.F4 = iso8583.NewNumeric("1")
	//流水号
	data.F11 = iso8583.NewNumeric("299999")
	//过期时间
	data.F14 = iso8583.NewNumeric("4912")
	data.F25 = iso8583.NewNumeric("00")

//	binary.BigEndian.Uint16(requst.pinEmv)

	pinEmv := requst.pinEmv;
	inputStyle := requst.posInputStyle

	if len(pinEmv) > 0 {
		data.F26 = iso8583.NewNumeric("06")
		data.F52 = iso8583.NewBinary(pinEmv)
		data.F53 = iso8583.NewNumeric("2600000000000000")
		if inputStyle == "" {
			data.F22 = iso8583.NewNumeric("021")
		} else {
			data.F22 = iso8583.NewNumeric(inputStyle)
		}

	} else {
		if inputStyle == "" {
			data.F22 = iso8583.NewNumeric("022")
		} else {
			data.F22 = iso8583.NewNumeric(inputStyle)
		}
	}
	//终端号
	data.F41 = iso8583.NewAlphanumeric("N9NL10087090")
	//商户号
	data.F42 = iso8583.NewAlphanumeric("")
	//交易货币代码
	data.F49 = iso8583.NewNumeric("156")

	orderId := requst.orderId;
	if orderId != "" {
		setField57ForUPLD()
	}
	//
	//data.F60 = iso8583.NewLllvar([]byte{iso8583.NewNumeric("22").Bytes(iso8583.BCD, 0, 2),
	//	iso8583.NewNumeric(requst.batchNum).Bytes(iso8583.BCD, 0, 6),
	//	iso8583.NewNumeric("003").Bytes(iso8583.BCD, 0, 3),
	//	iso8583.NewNumeric("0").Bytes(iso8583.BCD, 0, 1),
	//	iso8583.NewNumeric("0").Bytes(iso8583.BCD, 0, 1) })






	secondTrack := requst.secondTrack;
	track := strings.Replace(hex.EncodeToString(secondTrack), "D", "=", -1)
	if len(track) > 37 {
		track = track[:37]
	}
	data.F35 = iso8583.NewLlnumeric(track)




	//byte[] mab = request.getMAB();
	//byte[] macBytes = macCalculator.calcMac(mab);
	//String mac = HexCodec.hexEncode(macBytes, 0, macBytes.length);
	////Bit64信息确认码(MAC),数据包的最后一个域，用于验证信息来源的合法性，以及数据包中数据是否未被篡改。
	//request.setValue(64, new IsoField(8, EncodingType.ENCODING_TYPE_BINARY, mac));
}

func setField57ForUPLD() {

}

func getSys()  {


}

