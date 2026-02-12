package parser

import "unicode"

func ParseUID(identifier string) (string, bool) {
	if len(identifier) >= 4 && identifier[:2] == "((" && identifier[len(identifier)-2:] == "))" {
		return identifier[2 : len(identifier)-2], true
	}
	if looksLikeTitle(identifier) {
		return "", false
	}
	for _, r := range identifier {
		if !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_') {
			return "", false
		}
	}
	if len(identifier) > 0 && len(identifier) <= 40 {
		return identifier, true
	}
	return "", false
}

func looksLikeTitle(s string) bool {
	for _, r := range s {
		if unicode.IsSpace(r) || (r >= 0x4e00 && r <= 0x9fff) {
			return true
		}
	}
	return false
}
