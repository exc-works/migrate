package migrate

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var versionDigitPattern = regexp.MustCompile(`^[1-9][0-9]*$`)

func SplitFilename(filename string) (string, error) {
	if !strings.HasPrefix(filename, "V") {
		return "", fmt.Errorf("invalid migration filename: %s", filename)
	}
	idx := strings.Index(filename, "__")
	if idx <= 1 {
		return "", fmt.Errorf("invalid migration filename: %s", filename)
	}
	return filename[1:idx], nil
}

func CompareVersion(vi, vj string) bool {
	if versionDigitPattern.MatchString(vi) && versionDigitPattern.MatchString(vj) {
		vii, err := strconv.ParseUint(vi, 10, 64)
		if err == nil {
			vji, err := strconv.ParseUint(vj, 10, 64)
			if err == nil {
				return vii < vji
			}
		}
	}
	return vi < vj
}
