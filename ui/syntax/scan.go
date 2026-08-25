package syntax

import "strings"

const jsonPunctuation = "{}[],:"

func scanJSON(text string, out *writer) {
	index := 0
	for index < len(text) {
		character := text[index]
		switch {
		case character == '"':
			end := scanQuoted(text, index)
			kind := roleString
			if followedByColon(text, end) {
				kind = roleKey
			}
			out.token(text[index:end], kind)
			index = end
		case character == '-' || isDigit(character):
			end := scanNumber(text, index)
			out.token(text[index:end], roleNumber)
			index = end
		case strings.IndexByte(jsonPunctuation, character) >= 0:
			out.token(text[index:index+1], rolePunctuation)
			index++
		default:
			if literal := literalAt(text, index); literal != "" {
				out.token(literal, roleLiteral)
				index += len(literal)
				continue
			}
			end := index + 1
			for end < len(text) && !startsJSONToken(text, end) {
				end++
			}
			out.plain(text[index:end])
			index = end
		}
	}
}

func startsJSONToken(text string, index int) bool {
	character := text[index]
	if character == '"' || character == '-' || isDigit(character) {
		return true
	}
	if strings.IndexByte(jsonPunctuation, character) >= 0 {
		return true
	}
	return literalAt(text, index) != ""
}

func literalAt(text string, index int) string {
	for _, literal := range []string{"true", "false", "null"} {
		if strings.HasPrefix(text[index:], literal) {
			return literal
		}
	}
	return ""
}

// scanQuoted returns the index just past the closing quote of the string that
// starts at index.
func scanQuoted(text string, index int) int {
	quote := text[index]
	for cursor := index + 1; cursor < len(text); cursor++ {
		switch text[cursor] {
		case '\\':
			cursor++
		case quote:
			return cursor + 1
		}
	}
	return len(text)
}

func scanNumber(text string, index int) int {
	cursor := index
	if cursor < len(text) && (text[cursor] == '-' || text[cursor] == '+') {
		cursor++
	}
	for cursor < len(text) && (isDigit(text[cursor]) || text[cursor] == '.' ||
		text[cursor] == 'e' || text[cursor] == 'E' || text[cursor] == '-' || text[cursor] == '+') {
		cursor++
	}
	return cursor
}

func followedByColon(text string, index int) bool {
	for cursor := index; cursor < len(text); cursor++ {
		switch text[cursor] {
		case ' ', '\t', '\r', '\n':
			continue
		case ':':
			return true
		default:
			return false
		}
	}
	return false
}

func isDigit(character byte) bool {
	return character >= '0' && character <= '9'
}
