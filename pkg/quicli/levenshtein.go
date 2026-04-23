package quicli

// levenshtein returns the edit distance between two strings.
func levenshtein(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	la, lb := len(ra), len(rb)
	dp := make([][]int, la+1)
	for i := range dp {
		dp[i] = make([]int, lb+1)
		dp[i][0] = i
	}
	for j := range dp[0] {
		dp[0][j] = j
	}
	for i := 1; i <= la; i++ {
		for j := 1; j <= lb; j++ {
			if ra[i-1] == rb[j-1] {
				dp[i][j] = dp[i-1][j-1]
			} else {
				dp[i][j] = 1 + min(dp[i-1][j], dp[i][j-1], dp[i-1][j-1])
			}
		}
	}
	return dp[la][lb]
}

// findClosestSubcommand returns the subcommand name closest to input within
// maxDist edits, or "" if nothing is close enough.
func findClosestSubcommand(subcommands Subcommands, input string) string {
	const maxDist = 2
	best, bestDist := "", maxDist+1
	for _, s := range subcommands {
		if d := levenshtein(input, s.Name); d < bestDist {
			best, bestDist = s.Name, d
		}
		for _, a := range s.Aliases {
			if d := levenshtein(input, a); d < bestDist {
				best, bestDist = s.Name, d
			}
		}
	}
	if bestDist > maxDist {
		return ""
	}
	return best
}
