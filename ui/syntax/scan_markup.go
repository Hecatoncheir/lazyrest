package syntax

import "strings"

func scanXML(text string, out *writer) {
	index := 0
	for index < len(text) {
		offset := strings.IndexByte(text[index:], '<')
		if offset < 0 {
			out.plain(text[index:])
			return
		}
		out.plain(text[index : index+offset])
		index = scanXMLTag(text, index+offset, out)
	}
}

func scanXMLTag(text string, start int, out *writer) int {
	end := xmlTagEnd(text, start)
	tag := text[start:end]
	if strings.HasPrefix(tag, "<!--") {
		out.token(tag, roleComment)
		return end
	}

	// The opening bracket carries any of / ? ! that follow it.
	cursor := 1
	for cursor < len(tag) && strings.IndexByte("/?!", tag[cursor]) >= 0 {
		cursor++
	}
	out.token(tag[:cursor], rolePunctuation)

	nameEnd := cursor
	for nameEnd < len(tag) && isNameByte(tag[nameEnd]) {
		nameEnd++
	}
	out.token(tag[cursor:nameEnd], roleKey)
	cursor = nameEnd

	for cursor < len(tag) {
		character := tag[cursor]
		switch {
		case character == '"' || character == '\'':
			stop := scanQuoted(tag, cursor)
			out.token(tag[cursor:stop], roleString)
			cursor = stop
		case strings.IndexByte("=>/?", character) >= 0:
			out.token(tag[cursor:cursor+1], rolePunctuation)
			cursor++
		case isSpace(character):
			stop := cursor
			for stop < len(tag) && isSpace(tag[stop]) {
				stop++
			}
			out.plain(tag[cursor:stop])
			cursor = stop
		default:
			stop := cursor
			for stop < len(tag) && isNameByte(tag[stop]) {
				stop++
			}
			if stop == cursor {
				stop++
			}
			out.token(tag[cursor:stop], roleVariable)
			cursor = stop
		}
	}
	return end
}

// xmlTagEnd returns the index just past the bracket that closes the tag,
// ignoring brackets inside attribute values.
func xmlTagEnd(text string, start int) int {
	if strings.HasPrefix(text[start:], "<!--") {
		if offset := strings.Index(text[start:], "-->"); offset >= 0 {
			return start + offset + len("-->")
		}
		return len(text)
	}
	for cursor := start + 1; cursor < len(text); cursor++ {
		switch text[cursor] {
		case '"', '\'':
			cursor = scanQuoted(text, cursor) - 1
		case '>':
			return cursor + 1
		}
	}
	return len(text)
}

var graphQLKeywords = map[string]struct{}{
	"query": {}, "mutation": {}, "subscription": {},
	"fragment": {}, "on": {},
}

const graphQLPunctuation = "{}()[]:,=!|&@."

func scanGraphQL(text string, out *writer) {
	index := 0
	for index < len(text) {
		character := text[index]
		switch {
		case character == '#':
			end := strings.IndexByte(text[index:], '\n')
			if end < 0 {
				end = len(text)
			} else {
				end += index
			}
			out.token(text[index:end], roleComment)
			index = end
		case character == '"':
			end := scanQuoted(text, index)
			out.token(text[index:end], roleString)
			index = end
		case character == '$':
			end := index + 1
			for end < len(text) && isNameByte(text[end]) {
				end++
			}
			out.token(text[index:end], roleVariable)
			index = end
		case isDigit(character) || character == '-':
			end := scanNumber(text, index)
			out.token(text[index:end], roleNumber)
			index = end
		case isNameStart(character):
			end := index
			for end < len(text) && isNameByte(text[end]) {
				end++
			}
			word := text[index:end]
			switch {
			case isGraphQLKeyword(word):
				out.token(word, roleKeyword)
			case word == "true" || word == "false" || word == "null":
				out.token(word, roleLiteral)
			default:
				out.plain(word)
			}
			index = end
		case strings.IndexByte(graphQLPunctuation, character) >= 0:
			out.token(text[index:index+1], rolePunctuation)
			index++
		default:
			end := index
			for end < len(text) && !startsGraphQLToken(text[end]) {
				end++
			}
			out.plain(text[index:end])
			index = end
		}
	}
}

func startsGraphQLToken(character byte) bool {
	if character == '#' || character == '"' || character == '$' || character == '-' {
		return true
	}
	if isDigit(character) || isNameStart(character) {
		return true
	}
	return strings.IndexByte(graphQLPunctuation, character) >= 0
}

func isGraphQLKeyword(word string) bool {
	_, found := graphQLKeywords[word]
	return found
}

func isNameStart(character byte) bool {
	return character == '_' ||
		(character >= 'a' && character <= 'z') ||
		(character >= 'A' && character <= 'Z')
}

func isNameByte(character byte) bool {
	return isNameStart(character) || isDigit(character) ||
		character == '-' || character == '.' || character == ':'
}

func isSpace(character byte) bool {
	return character == ' ' || character == '\t' || character == '\r' || character == '\n'
}
