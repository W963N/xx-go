package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
)

const (
	FLAG_TITLE   = "XX-go"
	FLAG_WHOAMI  = "W963N"
	FLAG_NEWLINE = "\n"
)

func init() {
	flag.CommandLine.Init(FLAG_TITLE, flag.ContinueOnError)
	flag.CommandLine.Usage = func() {
		o := flag.CommandLine.Output()
		fmt.Fprintf(o, FLAG_NEWLINE+"%s by "+FLAG_WHOAMI+FLAG_NEWLINE,
			flag.CommandLine.Name())
		fmt.Fprintf(o, FLAG_NEWLINE)
		fmt.Fprintf(o, "@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@"+FLAG_NEWLINE)
		fmt.Fprintf(o, "@@@@@@@@@@@@@  .@@@@@@@@@.  @@@@@@@@@@@@"+FLAG_NEWLINE)
		fmt.Fprintf(o, "@@@@@@@@@ (@@@@@@@@@@@@@@@@@@@) @@@@@@@@"+FLAG_NEWLINE)
		fmt.Fprintf(o, "@@@@@@ @@ @@@@             @@@@ @@ @@@@@"+FLAG_NEWLINE)
		fmt.Fprintf(o, "@@@@ @@@@@ @ @@@         @@@ @ @@@@ @@@@"+FLAG_NEWLINE)
		fmt.Fprintf(o, "@@@ @@@@@@ @@@@@@       @@@@@@ @@@@@ @@@"+FLAG_NEWLINE)
		fmt.Fprintf(o, "@@ @@@@@@,@@@@@@@       @@@@@@@,@@@@@ @@"+FLAG_NEWLINE)
		fmt.Fprintf(o, "@@ @@@@@@ @@@@@@@@     @@@@@@@@ @@@@@@ @"+FLAG_NEWLINE)
		fmt.Fprintf(o, "@       @@@      @     @      @@@      @"+FLAG_NEWLINE)
		fmt.Fprintf(o, "@@ @@@ @@@@@@   @@@   @@@   @@@@@@ @@ @@"+FLAG_NEWLINE)
		fmt.Fprintf(o, "@@@ @@ @@@@@@@@@@@@   @@@@@@@@@@@@ @@ @@"+FLAG_NEWLINE)
		fmt.Fprintf(o, "@@@@  @@@@@@@@@@@@@@ @@@@@@@@@@@@@ @ @@@"+FLAG_NEWLINE)
		fmt.Fprintf(o, "@@@@@  @@@@@@@@@@@@@ @@@@@@@@@@@@   @@@@"+FLAG_NEWLINE)
		fmt.Fprintf(o, "@@@@@@@   @@@@@@@  @@@  @@@@@@@   @@@@@@"+FLAG_NEWLINE)
		fmt.Fprintf(o, "@@@@@@@@@@ @@@@@@@@@@@@@@@@@@@ @@@@@@@@@"+FLAG_NEWLINE)
		fmt.Fprintf(o, "@@@@@@@@@@@@@@@     @     @@@@@@@@@@@@@@"+FLAG_NEWLINE)
		fmt.Fprintf(o, FLAG_NEWLINE+FLAG_NEWLINE)
		fmt.Fprintf(o, "Options: "+FLAG_NEWLINE)
		fmt.Fprintf(o, FLAG_NEWLINE)
		flag.PrintDefaults()
		fmt.Fprintf(o, FLAG_NEWLINE+FLAG_NEWLINE)
		fmt.Fprintf(o, "                                Hobbyright 2022 walnut 🐿🐿🐿 .")
		fmt.Fprintf(o, FLAG_NEWLINE+FLAG_NEWLINE)
	}
	flag.StringVar(&env_flag, "e", "./env.toml", "path of env.toml.")
	flag.StringVar(&verbose_flag, "v", "error", "Select types(info, warn, error).")
	flag.StringVar(&inFile_flag, "i", "", "File to open.")
	flag.StringVar(&outFile_flag, "o", "outFile", "Output Filename.")
	flag.BoolVar(&rawOut_flag, "r", false, "Dump buffer to stdout instead of writing file.")
	flag.BoolVar(&dumpHex_flag, "x", false, "Dump hex instead of writing file.")
}

var (
	inFile_flag  string
	dumpHex_flag bool
	outFile_flag string
	rawOut_flag  bool
	env_flag     string
	verbose_flag string
)

func changeLogLevel(level string) error {
	switch level {
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	default:
		return errors.New("Don't match level.(info, warn, error)")
	}
	return nil
}

func main() {
	if err := flag.CommandLine.Parse(os.Args[1:]); err != nil {
		if err != flag.ErrHelp {
			fmt.Fprintf(os.Stderr, "error: %s\n", err)
		}
		os.Exit(0)
	}
	if err := changeLogLevel(verbose_flag); err != nil {
		log.Error(err)
		os.Exit(2)
	}

	file, err := os.Open(env_flag)
	if err != nil {
		log.Error("Don't open env file")
		os.Exit(1)
	}
	defer file.Close()
	var config cmdConf
	if err := loadConf(file, &config); err != nil {
		log.Error("Failed to decode toml file")
		os.Exit(1)
	}

	inFile := inFile_flag
	dumpHex := dumpHex_flag
	outFile := outFile_flag
	rawOut := rawOut_flag
	log.Info("inFile: ", inFile)
	log.Info("dumHex: ", dumpHex)
	log.Info("outFile: ", outFile)
	log.Info("rawOut: ", rawOut)

	fp, err := os.Open(inFile)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	defer fp.Close()
	scanner := bufio.NewScanner(fp)
	xxFileLines := []string{}
	for scanner.Scan() {
		xxFileLines = append(xxFileLines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Error(err)
		os.Exit(1)
	}

	out := parseXX(xxFileLines)

	if dumpHex {
		dHex(out)
	} else if rawOut {
		os.Stdout.Write(out)
	} else {
		outFile = config.Output + "/" + outFile
		writeBin(out, outFile)
	}
}
