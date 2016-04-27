package devn

//ScanEscapedLines is exactly like bufio.ScanLines, except lines ending in an
//escape character (\) are joined, without the newline or escape, and having
//stripped the first character where it is a comment character and matches the
//existing first line
func ScanEscapedLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	escape := false
	building := make([]byte, 0, len(data))
	for i, b := range data {
		if b == '\n' {
			if !escape {
				return i + 1, building, nil
			}
		}
		escape = false

		// Remove a # IFF:
		// - the first line of the escaped lines sequence is #
		// - it is the first character after a new line
		if data[0] == '#' && b == data[0] && i > 1 && (data[i-1] == '\n' || data[i-1] == '\r') {
			continue
		}

		if b == '\n' || b == '\r' {
			continue
		}

		if b == '\\' {
			escape = true
			continue
		}
		building = append(building, b)
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), building, nil
	}
	// Request more data.
	return 0, nil, nil
}
