package rum

func assert1(guard bool, text string) {
	if !guard {
		panic(text)
	}
}

func joinPath(absolutePath, relativePath string) string {
	if relativePath == "" {
		return absolutePath
	}

	buf := absolutePath + "/" + relativePath
	out := make([]byte, 0)
	for _, c := range buf {
		if len(out) != 0 && out[len(out)-1] == '/' && c == '/' {
			continue
		}
		out = append(out, byte(c))
	}
	return string(out)
}

func filterFlags(content string) string {
	for i, char := range content {
		if char == ' ' || char == ';' {
			return content[:i]
		}
	}
	return content
}
