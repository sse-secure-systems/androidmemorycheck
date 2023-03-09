package adb

import (
	"fmt"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

// Prepare regular expressions as soon as this module is used for the first time
var (
	regexpMemoryOverviewStart *regexp.Regexp // Seach for the pattern of the memory overview intro
	regexpBracket             *regexp.Regexp // Search for brackets
	regexpHeaderSeparator     *regexp.Regexp // Search for a '---' pattern that separates the header of a table from its content
	regexpKeyValuePairs       *regexp.Regexp // Search for key value pairs with the form "key: value"
)

func init() {

	regexpMemoryOverviewStart = regexp.MustCompile(`^[\t ]*[*]{2}[\t ]+MEMINFO[\t ]in[\t ]pid[\t ][0-9]+[\t ]+[\[][^\]]+[\]][\t ]+[*]{2}[\t ]*$`)
	regexpBracket = regexp.MustCompile(`([(][^)]*[)])`)
	regexpHeaderSeparator = regexp.MustCompile(`^[\t -]+$`)
	regexpKeyValuePairs = regexp.MustCompile(`([a-zA-Z_ ]+)[:][\t ]*([0-9]+)`)
}

// Reader is the adb reader that allowes to extract statistics from adb dumpsys meminfo
type Reader interface {
	Scan(packagename string) (map[string]int, string, error)          // Scans for data from adb dumpsys meminfo
	Packages() []string                                               // Scan for active packages
	Trend(key string, value int) (float64, float64, float64, float64) // Calculates a rough trend for the given key
}

// internalReader implements the reader interface
type internalReader struct {
	adbpath string                              `` // the path to the adb command on the machine
	trends  map[string]*internalTrendCalculator `` // All stored trends
}

// internalTrendCalculator is a helping structure for calculating rough trends
type internalTrendCalculator struct {
	hasvalues bool    `` // True if the trend has values
	p1000     float64 `` // The trend value influenced by 0.1% of the new value
	p100      float64 `` // The trend value influenced by 1% of the new value
	p10       float64 `` // The trend value influenced by 10% of the new value
	p1        float64 `` // The trend value influenced by 0the new value
}

// CreateReader creates a new reader instance for the given adp path
func CreateReader(adbpath string) (Reader, error) {

	// Check the custom adb path - if it does not exist, then just use adb
	adbpath = strings.TrimSpace(adbpath)
	if len(adbpath) == 0 {
		adbpath = "adb"
	}

	// Return a new reader instance
	return &internalReader{
		adbpath: adbpath,
		trends:  map[string]*internalTrendCalculator{},
	}, nil

}

// Scan for active packages
func (reader *internalReader) Packages() []string {
	if reader == nil {
		return nil
	}

	// Get the process id of the package
	readprocessescommand := exec.Command(reader.adbpath, "shell", "ps", "-A", "-o", "NAME")
	processesbytes, readprocessescommanderror := readprocessescommand.Output()
	if readprocessescommanderror != nil {
		return nil
	} else if processesbytes == nil {
		return nil
	}
	processesstring := strings.Replace(strings.TrimSpace(string(processesbytes)), "\r", "", -1)
	if len(processesstring) == 0 {
		return nil
	}
	processesstringelements := strings.Split(processesstring, "\n")
	output := []string{}
	for _, processtringelment := range processesstringelements {
		processtringelment = strings.TrimSpace(processtringelment)
		if !strings.HasPrefix(processtringelment, "[") && processtringelment != "NAME" {
			output = append(output, processtringelment)
		}
	}
	return output

}

// Calculates a rough trend of number behavior for a specific key. It returns four values:
//  1. the difference between the current number and the sliding average which is influenced to 0.1% by the new number
//  2. the difference between the current number and the sliding average which is influenced to 1% by the new number
//  3. the difference between the current number and the sliding average which is influenced to 10% by the new number
func (reader *internalReader) Trend(key string, value int) (float64, float64, float64, float64) {
	if reader == nil {
		return 0, 0, 0, 0
	}
	existingtrend, hasexistingtrend := reader.trends[key]
	if !hasexistingtrend {
		existingtrend := &internalTrendCalculator{}
		reader.trends[key] = existingtrend
	}
	return existingtrend.getTrend(value)
}

// Returns the memory information for the given process
func (reader *internalReader) Scan(packagename string) (map[string]int, string, error) {

	// Prepare the result
	result := map[string]int{}

	// Check if the reader is defined
	if reader == nil {
		return nil, "", fmt.Errorf("invalid reader")
	}

	// Check the package name
	packagename = strings.TrimSpace(packagename)
	if len(packagename) == 0 {
		return nil, "", fmt.Errorf("invalid package name")
	}

	// Get the process id of the package
	readprocessidcommand := exec.Command(reader.adbpath, "shell", "pidof", packagename)
	processidbytes, readprocessidcommanderror := readprocessidcommand.Output()
	if readprocessidcommanderror != nil {
		return nil, "", readprocessidcommanderror
	} else if processidbytes == nil {
		return nil, "", fmt.Errorf("process not found")
	}
	processidstring := strings.TrimSpace(string(processidbytes))
	if len(processidstring) == 0 {
		return nil, "", fmt.Errorf("process not found")
	}

	// Run the adb readmeminfocommand
	readmeminfocommand := exec.Command(reader.adbpath, "shell", "dumpsys", "meminfo", processidstring)
	meminfobytes, readmeminfocommanderror := readmeminfocommand.Output()
	if readmeminfocommanderror != nil {
		return nil, "", readmeminfocommanderror
	} else if meminfobytes == nil {
		return nil, "", fmt.Errorf("process not found")
	}

	// Prpare parsing of the result
	outputstring := string(meminfobytes)
	outputstring = strings.Replace(outputstring, "\r", "", -1)
	outputstringlines := strings.Split(outputstring, "\n")
	outputstringindex := 0

	// Parse the memory usage overview
	if !isMemoryOverviewIntro(outputstringlines, &outputstringindex) {
		return nil, "", fmt.Errorf("invalid response format (no memory overview found)")
	}
	meminfo := parseTable(outputstringlines, &outputstringindex, "MEM ")
	if len(meminfo) == 0 {
		return nil, "", fmt.Errorf("invalid response format (no memory info found)")
	}
	for meminfokey, meminfovalue := range meminfo {
		result[meminfokey] = meminfovalue
	}

	// Parse all the rest
	for ; outputstringindex < len(outputstringlines); outputstringindex++ {
		jumpOverEmptyLines(outputstringlines, &outputstringindex)
		if outputstringindex < len(outputstringlines) {
			currentline := strings.TrimSpace(outputstringlines[outputstringindex])
			currentlinelower := strings.ToLower(currentline)
			if currentlinelower == "app summary" {
				outputstringindex++
				tempmap := parseTable(outputstringlines, &outputstringindex, "APP ")
				if len(tempmap) > 0 {
					for meminfokey, meminfovalue := range tempmap {
						result[meminfokey] = meminfovalue
					}
				}
			} else if currentlinelower == "objects" {
				outputstringindex++
				tempmap := findKeyValueMatches(outputstringlines, &outputstringindex, "OBJECTS ")
				if len(tempmap) > 0 {
					for meminfokey, meminfovalue := range tempmap {
						result[meminfokey] = meminfovalue
					}
				}
			} else if currentlinelower == "sql" {
				outputstringindex++
				tempmap := findKeyValueMatches(outputstringlines, &outputstringindex, "SQL ")
				if len(tempmap) > 0 {
					for meminfokey, meminfovalue := range tempmap {
						result[meminfokey] = meminfovalue
					}
				}
			} else {
				outputstringindex++
			}
		}
	}

	return result, processidstring, nil

}

// getTrend calculates a rough trend of number behavior. It returns four values:
//  1. the difference between the current number and the sliding average which is influenced to 0.1% by the new number
//  2. the difference between the current number and the sliding average which is influenced to 1% by the new number
//  3. the difference between the current number and the sliding average which is influenced to 10% by the new number
//  4. the difference between the last value and the current one
func (trend *internalTrendCalculator) getTrend(value int) (float64, float64, float64, float64) {
	if trend == nil {
		return 0, 0, 0, 0
	}
	if trend.hasvalues {
		p1000old := trend.p1000
		p100old := trend.p100
		p10old := trend.p10
		p1old := trend.p1
		trend.p1000 = trend.p1000*0.999 + float64(value)*0.001
		trend.p100 = trend.p100*0.99 + float64(value)*0.01
		trend.p10 = trend.p10*0.9 + float64(value)*0.1
		trend.p1 = float64(value)
		return trend.p1000 - p1000old, trend.p100 - p100old, trend.p10 - p10old, trend.p1 - p1old
	}
	trend.hasvalues = true
	trend.p1000 = float64(value)
	trend.p100 = float64(value)
	trend.p10 = float64(value)
	trend.p1 = float64(value)
	return 0, 0, 0, 0
}

// Checks if the given line contains the intro for memory information (** MEMINFO in pid <processid> [<packagename>] **)
func isMemoryOverviewIntro(input []string, lineindex *int) bool {
	if lineindex == nil || input == nil {
		return false
	}
	for ; *lineindex < len(input); (*lineindex)++ {
		if regexpMemoryOverviewStart.MatchString(input[*lineindex]) {
			(*lineindex)++
			return true
		}
	}
	return false
}

// Raises the index pointer until the line contains content (or until the input array end was reached)
func jumpOverEmptyLines(input []string, lineindex *int) {
	for lineindex != nil && input != nil && *lineindex < len(input) {
		if len(strings.TrimSpace(input[*lineindex])) > 0 {
			return
		}
		(*lineindex)++
	}
}

// Splits the given input string at the given indizes
func splitAt(input string, indizes []int) []string {
	if indizes == nil {
		return nil
	}
	output := make([]string, 0)
	inputrunes := []rune(input)
	indizes = append(indizes, len(input))
	lastindex := 0
	for index := 0; index < len(indizes); index++ {
		currentindex := indizes[index]
		currentvalue := []rune{}
		for j := lastindex; j < currentindex && j < len(input); j++ {
			currentvalue = append(currentvalue, inputrunes[j])
		}
		currentstring := strings.TrimSpace(string(currentvalue))
		output = append(output, currentstring)
		lastindex = currentindex
	}
	return output
}

// Adds an index every time a subtext ends within the given line
func findIndizesOfEndingText(input string) []int {

	inputrunes := []rune(input)
	inputindex := 0
	findend := false
	output := []int{}
	for ; inputindex < len(input); inputindex++ {
		if !unicode.IsSpace(inputrunes[inputindex]) {
			break
		}
	}
	for ; inputindex < len(input); inputindex++ {
		if findend {
			if unicode.IsSpace(inputrunes[inputindex]) {
				output = append(output, inputindex)
				findend = false
			}
		} else {
			if !unicode.IsSpace(inputrunes[inputindex]) {
				findend = true
			}
		}
	}
	return output
}

// parseTable parses a table within the output. Each table starts with one or more column names followed by multiple '-' as header separator.
// The column width is defined by the end index of the column name. Each content line starts with the name of the line, followed by
// a number of each column. Values in brackets are ignored as well as empty cells
func parseTable(outputstringlines []string, outputstringindex *int, prefix string) map[string]int {

	// Overjump empty lines
	jumpOverEmptyLines(outputstringlines, outputstringindex)

	// Prepare variables
	result := map[string]int{}
	headerstartindex := *outputstringindex
	headerendindex := *outputstringindex
	headerlineindizes := []int{}
	headerlinevalues := []string{}
	namelength := 0

	// Parse the header names
	for ; *outputstringindex < len(outputstringlines); (*outputstringindex)++ {
		currentline := outputstringlines[*outputstringindex]
		if len(strings.TrimSpace(currentline)) == 0 {
			return nil
		} else if regexpHeaderSeparator.MatchString(currentline) {
			headerendindex = *outputstringindex
			(*outputstringindex)++
			break
		} else {
			tempindizes := findIndizesOfEndingText(currentline)
			for _, value := range tempindizes {
				found := false
				for _, existingvalue := range headerlineindizes {
					if existingvalue == value {
						found = true
						break
					}
				}
				if !found {
					headerlineindizes = append(headerlineindizes, value)
				}
			}
		}
	}
	if len(headerlineindizes) == 0 {
		return nil
	}
	sort.Slice(headerlineindizes, func(i, j int) bool {
		return headerlineindizes[i] <= headerlineindizes[j]
	})

	// Read the header names
	for i := headerstartindex; i < headerendindex; i++ {
		currentline := outputstringlines[i]
		currentsplitresult := splitAt(currentline, headerlineindizes)
		for j := 0; j < len(currentsplitresult); j++ {
			if len(headerlinevalues) <= j {
				headerlinevalues = append(headerlinevalues, strings.TrimSpace(currentsplitresult[j]))
			} else {
				headerlinevalues[j] = headerlinevalues[j] + " " + strings.TrimSpace(currentsplitresult[j])
			}
		}
	}

	// Parse the length of the name indizes
	for i := *outputstringindex; i < len(outputstringlines); i++ {
		currentline := outputstringlines[i]
		if len(strings.TrimSpace(currentline)) == 0 {
			break
		}
		tempindizes := findIndizesOfEndingText(currentline)
		if len(tempindizes) > 0 {
			if tempindizes[0] > namelength {
				namelength = tempindizes[0]
			}
		}
	}
	if namelength == 0 || namelength >= headerlineindizes[0] {
		return nil
	}

	// Add the name length to the header indizes and parse the table values
	headerlineindizes = append([]int{namelength}, headerlineindizes...)
	for ; *outputstringindex < len(outputstringlines); (*outputstringindex)++ {
		currentline := outputstringlines[*outputstringindex]
		if len(strings.TrimSpace(currentline)) == 0 {
			break
		}
		splitresult := splitAt(currentline, headerlineindizes)
		if len(splitresult) != len(headerlinevalues)+1 {
			return nil
		}
		for i := 1; i < len(splitresult); i++ {
			value := strings.TrimSpace(regexpBracket.ReplaceAllString(splitresult[i], ""))
			if len(value) > 0 {
				key := prefix + strings.TrimSpace(splitresult[0]) + " " + headerlinevalues[i-1]
				numbervalue, numbervalueerror := strconv.Atoi(value)
				if numbervalueerror == nil {
					result[key] = numbervalue
				}
			}
		}
	}

	// Done
	return result
}

// findKeyValueMatches finds patterns of type key:value within the given lines. The search ends when
// an empty line was reached or the pattern was not matched
func findKeyValueMatches(outputstringlines []string, outputstringindex *int, prefix string) map[string]int {

	// Overjump all empty lines
	jumpOverEmptyLines(outputstringlines, outputstringindex)

	result := map[string]int{}
	for ; *outputstringindex < len(outputstringlines); (*outputstringindex)++ {
		currentline := outputstringlines[*outputstringindex]
		if len(strings.TrimSpace(currentline)) == 0 {
			return result
		}
		matches := regexpKeyValuePairs.FindAllStringSubmatch(currentline, -1)
		if len(matches) == 0 {
			return result
		}
		for _, match := range matches {
			if len(match) != 3 {
				return result
			}
			key := prefix + strings.TrimSpace(match[1])
			value := strings.TrimSpace(regexpBracket.ReplaceAllString(match[2], ""))
			numbervalue, numbervalueerror := strconv.Atoi(value)
			if numbervalueerror == nil {
				result[key] = numbervalue
			}
		}
	}
	return result
}
