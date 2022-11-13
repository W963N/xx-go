package main

import (
	"encoding/hex"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/utf8string"
)

type xxToken struct {
	lineNum     int
	rawData     string
	rawDataLen  int
	normData    string
	normDataLen int
	hexData     string

	isString   bool
	isAscii    bool
	isHex      bool
	isComment  bool
	hasComment bool
}

func initXxToken(inTok *xxToken, inData string, lineNum int, isComment bool, isString bool) xxToken {
	inTok.lineNum = lineNum
	inTok.rawData = inData
	inTok.rawDataLen = len(inData)
	inTok.normData = inData
	inTok.normDataLen = len(inData)
	inTok.hexData = ""

	inTok.isString = isString
	inTok.isAscii = false
	inTok.isComment = isComment
	inTok.hasComment = false
	return *inTok
}

func testASCII(inTok *xxToken) {
	check_code := utf8string.NewString(inTok.normData)
	if check_code.IsASCII() {
		inTok.isAscii = true
	}
}

func testComment(inTok *xxToken) {
	if inTok.normDataLen > 0 {
		firstCharComment := testCharComment(string(inTok.normData[0]))
		if firstCharComment {
			inTok.isComment = true
			return
		} else {
			if inTok.normDataLen > 1 {
				for _, k := range twoCharComments {
					if strings.Contains(inTok.normData[0:2], k) {
						inTok.isComment = true
						return
					}
				}
			}
		}
		if !inTok.isAscii {
			cL := getCommentListPara()
			tempString := inTok.normData
			for _, comment := range cL {
				tempString_split := strings.Split(tempString, "")
				for i := 0; i < len(tempString_split); i++ {
					c := tempString_split[i]
					if strings.Contains(comment, c) {
						tempString = strings.Split(tempString, comment)[0]
						inTok.hasComment = true
						inTok.normData = tempString
						inTok.normDataLen = len(tempString)
					}
				}
			}
		}
	} else {
		return
	}
}

func testBinary(inTok *xxToken) {
	if !inTok.isString {
		if len(inTok.normData) == 10 {
			if inTok.normData[0:2] == "0y" {
				bindata := inTok.normData[2:]
				bindata_split := strings.Split(bindata, "")
				for i := 0; i < len(bindata_split); i++ {
					c := bindata_split[i]
					if (c != "0") && (c != "1") {
						log.Error("non binary")
						return
					}
				}
				bindata16, _ := strconv.ParseInt(bindata, 2, 16)
				bindatastr := strconv.FormatInt(bindata16, 16)
				if len(bindatastr) == 1 {
					bindatastr = "0" + bindatastr
				}
				inTok.normData = bindatastr
			}
		}
	}
}

func testHexData(inTok *xxToken) {
	if !inTok.isComment && !inTok.isString {
		tempData := filterIgnored(inTok.normData)
		encHex := hex.EncodeToString([]byte(tempData))
		testHex, err := hex.DecodeString(encHex)
		if err != nil {
			log.Error("Failed testHexData")
			return
		}
		if len(testHex) != 0 {
			inTok.isHex = true
			inTok.hexData = tempData
			inTok.normData = tempData
			inTok.normDataLen = len(tempData)
		}
	}
}

func getHexFromString(inTok *xxToken) {
	if inTok.isAscii && len(inTok.hexData) == 0 {
		if !inTok.isComment {
			inTok.hexData = ascii2hex(strings.Split(inTok.normData, ""))
		}
	}
}
