package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

const DefaultDelimiter = ","

type Options struct {
	Reader    io.Reader
	Delimiter rune
	NoHeaders bool
	Field     int
	Check     string
	Select    []int
}

var (
	options        *Options = new(Options)
	delim          string
	field          string
	selection      string
	splitSelection []string
)

func init() {
	flag.StringVar(&delim, "d", DefaultDelimiter, "CSV file delimiter character")
	flag.BoolVar(&options.NoHeaders, "no-headers", false, "Indicate that this CSV file has no headers")
	flag.StringVar(&selection, "s", "", "Fields to output (comma separated)")
	flag.Parse()

	// Extract delimeter
	for _, d := range delim {
		options.Delimiter = d
		break
	}

	splitSelection = strings.Split(selection, DefaultDelimiter)

	if flag.NArg() == 2 {
		field = flag.Arg(0)
		options.Check = flag.Arg(1)
		options.Reader = os.Stdin
	} else if flag.NArg() >= 3 {
		file, err := os.Open(flag.Arg(0))
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
		options.Reader = file
		field = flag.Arg(1)
		options.Check = flag.Arg(2)
	} else {
		fmt.Fprintf(os.Stderr, "%s\n", "Missing arguments")
		os.Exit(1)
	}

	options.Select = make([]int, 0, len(splitSelection))

	if options.NoHeaders {
		fieldIndex, err := strconv.Atoi(field)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", "Field must be a number")
			os.Exit(1)
		}
		options.Field = fieldIndex

		for _, selectionName := range splitSelection {
			selectionIndex, err := strconv.Atoi(selectionName)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", "Selection can only contain numbers")
				os.Exit(1)
			}
			options.Select = append(options.Select, selectionIndex)
		}
	}
}

func main() {
	var err error
	reader := csv.NewReader(options.Reader)
	reader.ReuseRecord = true
	reader.Comma = options.Delimiter

	if !options.NoHeaders {
		headers, err := reader.Read()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}

		for n, header := range headers {
			if header == field {
				options.Field = n
			}

			for i, selectionName := range splitSelection {
				if header == selectionName {
					if i < len(options.Select) {
						options.Select = append(options.Select, 0)
						copy(options.Select[i+1:], options.Select[i:])
						options.Select[i] = n
					} else {
						options.Select = append(options.Select, n)
					}
				}
			}
		}
	}

	// Now lets search
	var value string
	var record []string
	var output []string = make([]string, 0, len(options.Select))
	for {
		record, err = reader.Read()
		if err == io.EOF {
			os.Exit(0)
		}

		value = record[options.Field]
		if strings.Contains(value, options.Check) {
			output = output[:0]
			for _, i := range options.Select {
				output = append(output, record[i])
			}
			fmt.Fprintln(os.Stdout, strings.Join(output, "\t"))
		}
	}
}
