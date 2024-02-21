package internal

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"

	log "github.com/sirupsen/logrus"

	"gopkg.in/ini.v1"
)

func PrintHeader() {
	log.Infof("Maxon Reporter-Go [version %s] by Premysl Karbula", ReporterVersion)
}

func FatalExitOnError(err error) {
	if err != nil {
		ErrorExit("Fatal", err.Error())
	}
}

func ErrorExit(message ...string) {
	var msg string
	argLen := len(message)

	if argLen == 1 {
		msg = fmt.Sprintln("Error:", message[0])
	} else if argLen == 2 {
		msg = fmt.Sprintf("Error [%s]: %s", message[0], message[1])
	} else {
		FatalExitOnError(errors.New("zero or too many arguments to ErrorExit()"))
	}

	log.Fatal(msg)
}

// Returns true if the path is an existing file that can be opened. Returns
// false otherwise.
func IsExistingFile(path string) bool {
	info, err := os.Stat(path)

	// Error while fetching stat or it's adir - not a file for us.
	if err != nil || info.IsDir() {
		return false
	}

	// Error while opening - not a file for us.
	fd, err := os.Open(path)
	if err != nil {
		return false
	}

	fd.Close()
	return true
}

// Regexp replace function supplying the behavior of preg_replace_callback()
// with also the fact that the callback gets a list of all match groups as the
// argument.
func ReplaceAllStringSubmatchFunc(
	re *regexp.Regexp,
	str string,
	repl func([]string) (string, error),
	limit int,
) (string, error) {
	result := ""
	lastIndex := 0

	for index, v := range re.FindAllSubmatchIndex([]byte(str), -1) {
		var err error
		var next string

		groups := []string{}
		for i := 0; i < len(v); i += 2 {
			groups = append(groups, str[v[i]:v[i+1]])
		}

		if limit == -1 || index < limit {
			next, err = repl(groups)
			if err != nil {
				return "", err
			}
		} else {
			next = groups[0]
		}

		result += str[lastIndex:v[0]] + next
		lastIndex = v[1]
	}

	return result + str[lastIndex:], nil
}

// Takes an ordered list of StringMap structs and puts all their key-value
// pairs into a single StringMap, which is then returned.
func MergeResults(results []StringMap) StringMap {
	result := make(StringMap)

	for _, single := range results {
		for k, v := range single {
			result[k] = v
		}
	}

	return result
}

// Takes absolute path and tries to make and return a relative path - relative
// to the path of compiled reporter binary.
func TryMakingRelativePath(absPath string) string {
	rel, err := filepath.Rel(Settings.SelfDir, absPath)
	if err != nil {
		return absPath
	}

	return rel
}

// Reads INI values from bytes and returns them as map [string key: string value].
func readIniValues(bytes []byte) StringMap {
	result := make(StringMap)
	iniData, _ := ini.Load(bytes)
	defaultSection := iniData.Section("")

	for _, k := range defaultSection.Keys() {
		result[k.Name()] = k.Value()
	}

	return result
}

// This function returns true all timeout errors including the value
// context.DeadlineExceeded. That value satisfies the net.Error interface and
// has a Timeout method that always returns true.
// Thanks https://stackoverflow.com/a/56086437/1285669
func isTimeoutError(err error) bool {
	e, ok := err.(net.Error)
	return ok && e.Timeout()
}
