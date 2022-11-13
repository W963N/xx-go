package main

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/utf8string"
)

var (
	asciiComments   = []string{"#", ";", "%", "|", "\x1b", "-", "/"}
	twoCharComments = []string{"--", "//"}
	filterList      = []string{",", "$", "\\x", "0x", "h", ":", " "}
	escapes         = map[string]string{"n": "\n", "\\": "\\", "t": "\t", "r": "\r"}
)

const (
	COMMENT_START       = 9472
	COMMENT_END         = 9633
	MULTI_COMMENT_START = "/*"
	MULTI_COMMENT_END   = "*/"
	ESCAPE_SEQ          = "\\"
	DOUBLE_QUATE        = "\""
	SPACE               = " "
)

func getTokenAttributes(inTok *xxToken) {
	testASCII(inTok)
	testComment(inTok)
	testBinary(inTok)
	testHexData(inTok)
	getHexFromString(inTok)
}

func getCommentList() []string {
	cList := []string{}
	for _, c := range asciiComments {
		cList = append(cList, c)
	}
	for _, c := range twoCharComments {
		cList = append(cList, c)
	}
	for i := COMMENT_START; i <= COMMENT_END; i++ {
		cList = append(cList, string(rune(i)))
	}
	return cList
}

func getCommentListPara() []string {
	cList := []string{}
	stream_ascii := make(chan string)
	stream_charComments := make(chan string)
	stream_comments := make(chan string)

	go func(ascii []string) {
		defer close(stream_ascii)
		for _, c := range ascii {
			stream_ascii <- c
		}
	}(asciiComments)
	go func(char []string) {
		defer close(stream_charComments)
		for _, c := range twoCharComments {
			stream_charComments <- c
		}
	}(twoCharComments)
	go func(start, end int) {
		defer close(stream_comments)
		for i := start; i <= end; i++ {
			stream_comments <- string(rune(i))
		}
	}(COMMENT_START, COMMENT_END)
	for set := range stream_ascii {
		cList = append(cList, set)
	}
	for set := range stream_charComments {
		cList = append(cList, set)
	}
	for set := range stream_comments {
		cList = append(cList, set)
	}

	return cList
}

func filterIgnored(inText string) string {
	newText := inText
	for _, f := range filterList {
		newText = strings.ReplaceAll(newText, f, "")
	}
	return newText
}

func include(list []string, target string) bool {
	for _, num := range list {
		if num == target {
			return true
		}
	}
	return false
}

func testCharComment(inChar string) bool {
	tCom := inChar
	o := []rune(tCom)[0]
	if o >= COMMENT_START && o <= COMMENT_END {
		return true
	} else if include(asciiComments, tCom) {
		return true
	} else {
		return false
	}
}

func ascii2hex(inString []string) string {
	outString := []string{}
	for _, char := range inString {
		outString = append(outString,
			hex.EncodeToString([]byte(char)))
	}
	return strings.Join(outString, "")
}

func writeBin(out []byte, outfile string) {
	err := ioutil.WriteFile(outfile, out, 0644)
	if err != nil {
		log.Error(err)
		os.Exit(2)
	}
	log.Info("success: ", outfile)
}

func dHex(inBytes []byte) {
	offs := 0
	maxoffs := 16
	for offs < len(inBytes) {
		bHex := ""
		bAsc := ""

		basenum := (len(inBytes) - offs)
		chklen := maxoffs - basenum
		if 0 < chklen && chklen < maxoffs {
			maxoffs = basenum
		} else if maxoffs > len(inBytes) {
			maxoffs = len(inBytes)
			log.Info("maxoffs :", maxoffs, offs+maxoffs)
		}

		bChunk := inBytes[offs : offs+maxoffs]
		hexChunk := hex.EncodeToString(bChunk)
		bHex_list := []string{}
		bAsc_list := []string{}
		for i := 0; i < maxoffs; i++ {
			offc := i * 2
			bHex = hexChunk[offc : offc+2]
			bAscr, err := hex.DecodeString(hexChunk[offc : offc+2])
			if err != nil {
				log.Error("Failed DecodeString")
			}
			bAsc = string(bAscr)
			if strconv.IsPrint([]rune(bAsc)[0]) && []rune(bAsc)[0] < 0x7F {
				bAsc_list = append(bAsc_list, bAsc)
			} else {
				bAsc_list = append(bAsc_list, ".")
			}
			bHex_list = append(bHex_list, bHex+" ")
		}
		for i := 0; i < 16-maxoffs; i++ {
			bHex_list = append(bHex_list, "   ")
		}
		bHex = strings.Join(bHex_list, "")
		bAsc = strings.Join(bAsc_list, "")
		fmt.Printf("%08x: %s %s\n", offs, bHex, bAsc)
		offs = offs + maxoffs
	}
}

func filterMultLineComments(multilineComment bool, joinedLine, line string) (bool, string, string, bool) {
	lineResult := joinedLine
	joinedLine = ""
	mustContinue := false
	for len(line) > 0 {
		if multilineComment {
			if strings.Contains(line, MULTI_COMMENT_END) {
				_, e, _ := strings.Cut(line, MULTI_COMMENT_END)
				lineResult += e
				line = MULTI_COMMENT_END + e
				multilineComment = false
			} else {
				joinedLine += lineResult
				mustContinue = true
				break
			}
		} else {
			if strings.Contains(line, MULTI_COMMENT_START) {
				s, e, _ := strings.Cut(line, MULTI_COMMENT_START)
				lineResult += s
				line = MULTI_COMMENT_START + e
				multilineComment = true
			} else {
				lineResult += line
				break
			}
		}
	}
	return multilineComment, joinedLine, lineResult, mustContinue
}

func tokenizeXX(xxline string, lineNum int) []xxToken {
	xxline = strings.TrimSpace(xxline)
	tokens := []xxToken{}
	buf := ""
	verbatim := false
	isEscape := false
	isString := false

	xxline_split := strings.Split(xxline, "")
	for i := 0; i < len(xxline_split); i++ {
		c := xxline_split[i]
		if c == ESCAPE_SEQ && !isEscape && verbatim {
			isEscape = true
			continue
		}
		if isEscape {
			if val, ok := escapes[c]; ok {
				buf += val
			} else {
				buf += ESCAPE_SEQ
				buf += c
			}
			isEscape = false
			continue
		}
		if c == DOUBLE_QUATE {
			verbatim = !verbatim
			isString = true
			continue
		}
		if c == SPACE && !verbatim {
			if buf != "" {
				isComment := false
				comments := append(asciiComments, twoCharComments...)
				for _, k := range comments {
					if strings.Contains(buf, k) {
						isComment = true
						break
					}
				}
				var newToken xxToken
				newToken = initXxToken(&newToken, buf, lineNum, isComment, isString)
				tokens = append(tokens, newToken)
				isString = false
			}
			buf = ""
			continue
		}
		buf += c
	}
	var newToken xxToken
	newToken = initXxToken(&newToken, buf, lineNum, false, isString)
	tokens = append(tokens, newToken)

	return tokens
}

func parseXX(xxFile []string) []byte {
	xxOut := []byte{}
	lineNum := 0
	joinedLine := ""
	multilineComment := false
	mustContinue := false

	for _, line := range xxFile {
		lineNum = lineNum + 1
		multilineComment, joinedLine, line, mustContinue =
			filterMultLineComments(multilineComment, joinedLine, line)
		if mustContinue {
			continue
		}
		lineTokens := tokenizeXX(line, lineNum)
		//isComment := false
		linesHexData := ""
		for _, t := range lineTokens {
			getTokenAttributes(&t)
			if t.isComment || t.hasComment {
				//isComment = true
				break
			}
			linesHexData += t.hexData
		}
		out, err := hex.DecodeString(linesHexData)
		if err != nil {
			check_code := utf8string.NewString(linesHexData)
			if check_code.IsASCII() {
				linesHexData_list := strings.Split(linesHexData, "")
				linesHexData_byte := ascii2hex(linesHexData_list)
				out, err = hex.DecodeString(linesHexData_byte)
				if err != nil {
					log.Error(err)
					os.Exit(3)
				}
			} else {
				log.Error(err)
				os.Exit(3)
			}
		}
		xxOut = append(xxOut, out...)
	}
	return xxOut
}
