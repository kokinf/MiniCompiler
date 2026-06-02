package utils

// SuggestSimilar ищет похожие имена из списка
func SuggestSimilar(name string, candidates []string, maxDistance int) string {
	if maxDistance <= 0 {
		maxDistance = 2
	}

	bestMatch := ""
	bestDistance := maxDistance + 1

	for _, candidate := range candidates {
		distance := levenshteinDistance(name, candidate)
		if distance < bestDistance {
			bestDistance = distance
			bestMatch = candidate
		}
	}

	return bestMatch
}

// SuggestFix форматирует подсказку
func SuggestFix(wrong, suggestion string) string {
	if suggestion == "" {
		return ""
	}
	return "возможно, вы имели в виду '" + suggestion + "'?"
}

func levenshteinDistance(s, t string) int {
	if len(s) == 0 {
		return len(t)
	}
	if len(t) == 0 {
		return len(s)
	}

	d := make([][]int, len(s)+1)
	for i := range d {
		d[i] = make([]int, len(t)+1)
		d[i][0] = i
	}
	for j := 0; j <= len(t); j++ {
		d[0][j] = j
	}

	for i := 1; i <= len(s); i++ {
		for j := 1; j <= len(t); j++ {
			cost := 1
			if s[i-1] == t[j-1] {
				cost = 0
			}
			d[i][j] = min3(
				d[i-1][j]+1,
				d[i][j-1]+1,
				d[i-1][j-1]+cost,
			)
		}
	}

	return d[len(s)][len(t)]
}

func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}
