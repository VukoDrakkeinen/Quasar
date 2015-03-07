package qregexp

import (
	"regexp"
	"strings"
)

func MustCompile(str string) *QRegexp {
	expr, lookAroundGroup := transformExpr(str)
	return &QRegexp{regexp.MustCompile(expr), lookAroundGroup}
}

type QRegexp struct {
	*regexp.Regexp
	lookAroundGroup int
}

func (this *QRegexp) Find(b []byte) []byte {
	if this.lookAroundGroup == 0 {
		return this.Regexp.Find(b)
	}
	result := this.Regexp.FindSubmatch(b)
	if len(result) > this.lookAroundGroup {
		return result[this.lookAroundGroup]
	} else {
		return []byte{}
	}
}

func (this *QRegexp) FindString(s string) string {
	if this.lookAroundGroup == 0 {
		return this.Regexp.FindString(s)
	}
	result := this.Regexp.FindStringSubmatch(s)
	if len(result) > this.lookAroundGroup {
		return result[this.lookAroundGroup]
	} else {
		return ""
	}
}

func (this *QRegexp) FindAll(b []byte, n int) [][]byte {
	if this.lookAroundGroup == 0 {
		return this.Regexp.FindAll(b, n)
	}
	result := this.Regexp.FindAllSubmatch(b, n)
	ret := make([][]byte, 0, len(result))
	for _, entry := range result {
		if len(entry) > this.lookAroundGroup {
			ret = append(ret, entry[this.lookAroundGroup])
		}
	}
	return ret
}

func (this *QRegexp) FindAllString(s string, n int) []string {
	if this.lookAroundGroup == 0 {
		return this.Regexp.FindAllString(s, n)
	}
	result := this.Regexp.FindAllStringSubmatch(s, n)
	ret := make([]string, 0, len(result))
	for _, entry := range result {
		if len(entry) > this.lookAroundGroup {
			ret = append(ret, entry[this.lookAroundGroup])
		}
	}
	return ret
}

func transformExpr(str string) (exp string, group int) {
	lb := false
	la := false
	index1 := 0
	index2 := 0
	if strings.HasPrefix(str, "(?<=") {
		lb = true
		index1 = findMatchingParenthesisIndex(str)
	}
	if i := strings.Index(str, "(?="); i != -1 {
		la = true
		index2 = i
	}
	switch {
	case lb && la:
		return str[4:index1-1] + "(" + str[index1:index2] + ")" + str[index2+3:len(str)-1], countCapturingGroups(str[3:index1-1]) + 1
	case lb:
		return str[4:index1-1] + "(" + str[index1:] + ")", countCapturingGroups(str[3:index1-1]) + 1
	case la:
		return "(" + str[:index2] + ")" + str[index2+3:len(str)-1], 1
	default:
		return str, 0
	}
}

func findMatchingParenthesisIndex(str string) int {
	counter := 0
	start := true
	ignore := false
	for i, char := range str {
		switch {
		case counter <= 0 && !start:
			return i
		case char == ')' && !ignore:
			ignore = false
			counter--
		case char == '(' && !ignore:
			ignore = false
			counter++
		case char == '\\' && !ignore:
			ignore = true
		default:
			ignore = false
		}
		start = false
	}
	return len(str)
}

func countCapturingGroups(str string) int {
	counter := 0
	ignore := false
	for _, char := range str {
		switch {
		case char == '(' && !ignore:
			counter++
		case char == '\\' && !ignore:
			ignore = true
		default:
			ignore = false
		}
	}
	return counter
}

/*
func (this *QRegexp) FindFirstStringSubmatch(str string) string {
	result := this.FindStringSubmatch(str)
	if len(result) > this.lookAroundGroup+1 {
		return result[this.lookAroundGroup+1]
	} else {
		return ""
	}
}

func (this *QRegexp) FindFirstSubmatch(b []byte) []byte {
	result := this.FindSubmatch(b)
	if len(result) > this.lookAroundGroup+1 {
		return result[this.lookAroundGroup+1]
	} else {
		return []byte{}
	}
}
*/
