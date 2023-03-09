package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/ahmetalpbalkan/go-cursor"
	"securesystems.engineering/androidstat/adb"
)

func main() {

	// Prepare the flags, provided by the sytem
	argumentADBPath := flag.String("adb", "adb", "[optional] Path to adb executable (default: adb)")
	argumentPackageName := flag.String("p", "", "Package name to be analyzed")
	argumentRefreshRate := flag.Int("t", 0, "Refresh rate in seconds (default: 0)")
	argumentOutputFilename := flag.String("o", "", "[optional] The name of a csv file the results have to be written to.")
	argumentOutputFilter := flag.String("f", "", "[optional] A regular expression to filter the output")
	argumentOutputFileFilter := flag.String("of", "", "[optional] A regular expression to filter the output file columns")

	// Create the function that returns the running processes
	getNameInfo := func() {

		// Parse all attributes
		flag.Parse()

		// Check the package name
		if len(*argumentPackageName) == 0 {
			fmt.Println("Invalid package name")
			os.Exit(1)
		}

		// Check if a path was specified
		if argumentADBPath == nil {
			temppath := "adb"
			argumentADBPath = &temppath
		}

		// Start reading the memory info
		reader, _ := adb.CreateReader(*argumentADBPath)
		nexttime := time.Now()
		starttime := time.Now()

		// Scan the memory
		scanresult, processid, readerror := reader.Scan(*argumentPackageName)

		// Output the header
		fmt.Println(cursor.Hide())
		fmt.Println("\x1b[1m\x1b[97mMeasurement: ", "\x1b[0m"+starttime.Format("02.01.2006 (15:04:05)"), cursor.ClearLineRight())

		// Check the scan result
		if readerror != nil {

			// Write out package information
			fmt.Println("\x1b[1m\x1b[97mProcess:     ", "\x1b[31mINACTIVE\x1b[0m\x1b[97m (\x1b[96m"+*argumentPackageName+"\x1b[97m)", cursor.ClearLineRight())

		} else {

			// Write out package information
			fmt.Println("\x1b[1m\x1b[97mProcess:     ", "\x1b[32mACTIVE\x1b[0m", "(\x1b[96m"+*argumentPackageName, "=> pid:"+processid+"\x1b[97m)", cursor.ClearLineRight())

			// Get a sorted list of all detected keywords
			keywords := make([]string, 0)
			keywordlengthmax := 0
			for keyword := range scanresult {
				keywords = append(keywords, keyword)
				if len(keyword) > keywordlengthmax {
					keywordlengthmax = len(keyword)
				}
			}

			// Check if there are any keywords within the slice
			if len(keywords) > 0 {

				fmt.Println("\x1b[1m\x1b[97mKeywords:\x1b[0m")

				// Sort the keyword list
				sort.Slice(keywords, func(i, j int) bool {
					return strings.Compare(keywords[i], keywords[j]) >= 0
				})

				// Add all the keywords
				for i := 0; i < len(keywords); i++ {
					fmt.Println("   -", keywords[i])
				}
			}
		}

		// Make sure the time interval is as exact as possible
		difference := time.Since(nexttime)
		millisecondstosleep := difference.Milliseconds()
		if millisecondstosleep < 0 {
			time.Sleep(time.Duration(-millisecondstosleep) * time.Millisecond)
		}
	}

	// Create the function that returns the running processes
	getProcessInfo := func() {

		// Parse all attributes
		flag.Parse()

		// Check if a path was specified
		if argumentADBPath == nil {
			temppath := "adb"
			argumentADBPath = &temppath
		}

		// Start reading the memory info
		reader, _ := adb.CreateReader(*argumentADBPath)
		nexttime := time.Now()

		// Scan the memory
		packages := reader.Packages()

		// Output the header
		fmt.Println("\x1b[1m\x1b[97mMeasurement: ", "\x1b[0m"+time.Now().Format("02.01.2006 (15:04:05)"), cursor.ClearLineRight())
		fmt.Println("\x1b[1m\x1b[97mPackages:\x1b[0m")

		// Check the scan result
		if len(packages) == 0 {

			// Write out package information
			fmt.Println("   -", "\x1b[1m\x1b[91mNo packages found\x1b[0m\x1b[97m")

		} else {

			// Sort the keyword list
			sort.Slice(packages, func(i, j int) bool {
				return strings.Compare(packages[i], packages[j]) >= 0
			})

			// Add all the keywords
			for i := 0; i < len(packages); i++ {
				fmt.Println("   -", packages[i])
			}
		}

		// Make sure the time interval is as exact as possible
		difference := time.Since(nexttime)
		millisecondstosleep := difference.Milliseconds()
		if millisecondstosleep < 0 {
			time.Sleep(time.Duration(-millisecondstosleep) * time.Millisecond)
		}
	}

	// Create the function that returns the memory dump
	getMemoryInfo := func() {

		// Parse all attributes
		flag.Parse()

		// Check the package name
		if len(*argumentPackageName) == 0 {
			fmt.Println("Invalid package name (not provided as argument)")
			os.Exit(1)
		}

		// Check if a path was specified
		if argumentADBPath == nil {
			temppath := "adb"
			argumentADBPath = &temppath
		}

		// Parse the refresh rate
		refreshrate := 0
		if argumentRefreshRate != nil {
			refreshrate = *argumentRefreshRate
			if refreshrate < 0 {
				refreshrate = 0
			}
		}

		// Create the output file if needed
		var outputfile *os.File
		var outputfileerror error
		if argumentOutputFilename != nil && len(*argumentOutputFilename) > 0 {
			if outputfile, outputfileerror = os.OpenFile(*argumentOutputFilename, os.O_CREATE, 0644); outputfileerror != nil {
				fmt.Println("Failed to create output file:", *argumentOutputFilename)
				os.Exit(1)
			}
		}

		// Apply the filter
		var outputFileFilter *regexp.Regexp
		if argumentOutputFileFilter != nil && len(*argumentOutputFileFilter) > 0 {
			outputFileFilter = regexp.MustCompile(*argumentOutputFileFilter)
			if outputFileFilter == nil {
				fmt.Println("Failed to apply filter:", *argumentOutputFileFilter)
				os.Exit(1)
			}
		}

		// Apply the filter
		var filter *regexp.Regexp
		if argumentOutputFilter != nil && len(*argumentOutputFilter) > 0 {
			filter = regexp.MustCompile(*argumentOutputFilter)
			if filter == nil {
				fmt.Println("Failed to apply filter:", *argumentOutputFilter)
				os.Exit(1)
			}
		}

		// Clear the screen
		if refreshrate > 0 {
			fmt.Println(cursor.ClearEntireScreen())
			fmt.Println(cursor.MoveTo(0, 0))
		}

		// Start reading the memory info
		haspackage := false
		packagetimestamp := time.Now()
		reader, _ := adb.CreateReader(*argumentADBPath)
		nexttime := time.Now()
		starttime := time.Now()
		filesize := uint64(0)
		var filetempnames []string
		for measurement := 1; ; measurement++ {

			// Get the timestamp of the next refresh
			currenttime := nexttime
			nexttime = nexttime.Add(time.Duration(refreshrate) * time.Second).Truncate(time.Second)
			if nexttime.Before(time.Now()) {
				nexttime = time.Now().Truncate(time.Second)
			}

			// Scan the memory
			scanresult, processid, readerror := reader.Scan(*argumentPackageName)

			fmt.Println(cursor.Hide())
			if refreshrate > 0 {
				fmt.Print(cursor.MoveTo(0, 0))
				fmt.Println(cursor.ClearLineRight())
			}

			// Output the header
			if refreshrate > 0 {
				fmt.Println("\x1b[1m\x1b[97mStart:              ", "\x1b[0m"+starttime.Format("02.01.2006 (15:04:05)"), cursor.ClearLineRight())
				fmt.Println("\x1b[1m\x1b[97mCurrent measurement:", "\x1b[0m"+currenttime.Format("02.01.2006 (15:04:05)"), "[\x1b[96m"+fmt.Sprint(measurement)+"\x1b[97m]", cursor.ClearLineRight())
				fmt.Println("\x1b[1m\x1b[97mNext measurement:   ", "\x1b[0m"+nexttime.Format("02.01.2006 (15:04:05)"), "[\x1b[96m"+fmt.Sprint(measurement+1)+"; rate=1/"+fmt.Sprint(refreshrate)+"sec\x1b[97m]", cursor.ClearLineRight())
			} else {
				fmt.Println("\x1b[1m\x1b[97mMeasurement:        ", "\x1b[0m"+starttime.Format("02.01.2006 (15:04:05)"), cursor.ClearLineRight())
			}

			// Check the scan result
			if readerror != nil {

				if haspackage {
					packagetimestamp = time.Now()
					haspackage = false
				}

				// Write out package information
				fmt.Println("\x1b[1m\x1b[97mProcess:            ", "\x1b[31mINACTIVE \x1b[0m\x1b[97msince "+fmt.Sprint(math.Round(time.Since(packagetimestamp).Seconds()))+"sec", "(\x1b[96m"+*argumentPackageName+"\x1b[97m)", cursor.ClearLineRight())

			} else {

				// Note the time the package was read the last time
				haspackage = true

				// Write out package information
				fmt.Println("\x1b[1m\x1b[97mProcess:            ", "\x1b[32mACTIVE\x1b[0m", "(\x1b[96m"+*argumentPackageName, "=> pid:"+processid+"\x1b[97m)", cursor.ClearLineRight())

				// Get a sorted list of all detected keywords
				keywords := make([]string, 0)
				keywordlengthmax := 0
				for keyword := range scanresult {
					if filter != nil {
						if filter.MatchString(keyword) {
							keywords = append(keywords, keyword)
							if len(keyword) > keywordlengthmax {
								keywordlengthmax = len(keyword)
							}
						}
					} else {
						keywords = append(keywords, keyword)
						if len(keyword) > keywordlengthmax {
							keywordlengthmax = len(keyword)
						}
					}
				}

				// Check if there are any keywords within the slice
				if len(keywords) > 0 {

					// Sort the keyword list
					sort.Slice(keywords, func(i, j int) bool {
						return strings.Compare(keywords[i], keywords[j]) >= 0
					})

					// Add the header to the csv file
					if filetempnames == nil && outputfile != nil {
						for keyword := range scanresult {
							if outputFileFilter != nil {
								if outputFileFilter.MatchString(keyword) {
									filetempnames = append(filetempnames, keyword)
								}
							} else {
								filetempnames = append(filetempnames, keyword)
							}
						}
						stringbuilder := strings.Builder{}
						stringbuilder.WriteString("Time")
						for _, keyword := range keywords {
							stringbuilder.WriteString(";")
							stringbuilder.WriteString(keyword)
						}
						stringbuilder.WriteString("\n")
						outputfile.WriteString(stringbuilder.String())
					}

					fmt.Println(cursor.ClearEntireLine())
					formatstringheader := "\x1b[1m\x1b[97m%-" + fmt.Sprint(keywordlengthmax) + "s \x1b[0m\x1b[90m|\x1b[1m\x1b[97m %8s \x1b[0m\x1b[90m|\x1b[1m\x1b[90m %11s \x1b[0m\x1b[90m|\x1b[1m\x1b[90m %11s \x1b[0m\x1b[90m|\x1b[1m\x1b[90m %11s \x1b[0m\x1b[90m|\x1b[1m\x1b[90m %11s\x1b[0m\x1b[97m%s\n"
					fmt.Printf(formatstringheader, "Name", "Value", "T1000", "T100", "T10", "T1", cursor.ClearLineRight())
					fmt.Printf("\x1b[0m\x1b[90m%s%s%s\x1b[30m\n", strings.Repeat("-", keywordlengthmax+1), "+----------+-------------+-------------+-------------+-------------\x1b[0m\x1b[97m", cursor.ClearLineRight())

					// Output the results
					formatstring := "\x1b[1m\x1b[97m%-" + fmt.Sprint(keywordlengthmax) + "s \x1b[0m\x1b[90m|\x1b[1m\x1b[97m %8d \x1b[0m\x1b[90m| %11.1f | %11.1f | %11.1f | %11.1f\x1b[0m\x1b[97m%s\n"
					for _, keyword := range keywords {
						value := scanresult[keyword]
						p1000, p100, p10, p1 := reader.Trend(keyword, value)
						fmt.Printf(formatstring, keyword, value, p1000, p100, p10, p1, cursor.ClearLineRight())
					}
				}

				if outputfile != nil {
					stringbuilder := strings.Builder{}
					stringbuilder.WriteString(currenttime.Format("2006-01-02 15:04:05"))
					for _, keyword := range filetempnames {
						stringbuilder.WriteString(";")
						stringbuilder.WriteString(fmt.Sprint(scanresult[keyword]))
					}
					stringbuilder.WriteString("\n")
					filecontent := stringbuilder.String()
					filecontentlength := len(filecontent)
					filesize += uint64(filecontentlength)
					outputfile.WriteString(filecontent)
					fmt.Println(cursor.ClearLineRight())
					if filesize > 10485760 {
						fmt.Println("\x1b[0m\x1b[97mAddded\x1b[96m", filecontentlength, "\x1b[97m\x1b[0msymbols to", *argumentOutputFilename, "(\x1b[96mfilesize: "+fmt.Sprint(filesize/1048576)+"MB\x1b[97m\x1b[0m)", cursor.ClearLineRight())
					} else if filesize > 2048 {
						fmt.Println("\x1b[0m\x1b[97mAddded\x1b[96m", filecontentlength, "\x1b[97m\x1b[0msymbols to", *argumentOutputFilename, "(\x1b[96mfilesize: "+fmt.Sprint(filesize/1024)+"kB\x1b[97m\x1b[0m)", cursor.ClearLineRight())
					} else {
						fmt.Println("\x1b[0m\x1b[97mAddded\x1b[96m", filecontentlength, "\x1b[97m\x1b[0msymbols to", *argumentOutputFilename, "(\x1b[96mfilesize: "+fmt.Sprint(filesize)+"B\x1b[97m\x1b[0m)", cursor.ClearLineRight())
					}
				}
			}

			// Clear the rest of the screen
			fmt.Print(cursor.ClearScreenDown())
			fmt.Print(cursor.Show())

			// Make sure the time interval is as exact as possible
			difference := time.Since(nexttime)
			millisecondstosleep := difference.Milliseconds()
			if millisecondstosleep < 0 {
				time.Sleep(time.Duration(-millisecondstosleep) * time.Millisecond)
			}

			// Exit the loop if no refresh is required
			if refreshrate == 0 {
				break
			}
		}
	}

	// Parse all arguments
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "meminfo":
			os.Args = os.Args[1:]
			getMemoryInfo()
		case "packages":
			os.Args = os.Args[1:]
			getProcessInfo()
		case "names":
			os.Args = os.Args[1:]
			getNameInfo()
		default:
			flag.Parse()
		}
	} else {
		fmt.Println("expected 'run', 'processes', or 'names' subcommands")
	}
}
