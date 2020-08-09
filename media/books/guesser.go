package books

import (
	"regexp"
)

var (
	// reSerieGuesser is a collection of regexp to extract series information
	// from title/subtitles.
	// It should be made of 3 named capturing groups (title, serie, serie number).
	reSerieGuesser = []*regexp.Regexp{
		regexp.MustCompile(`^(?P<title>.+)\s\((?P<serie>.+?)\s(?i:#|Series |n°|)(?P<seriePos>\d+)\)$`),
		regexp.MustCompile(`^(?P<title>.+)\s-\s(?P<serie>.+?)\s(?i:#|Series |n°|)(?P<seriePos>\d+)$`),
		regexp.MustCompile(`^(?P<serie>.+?)\s(?i:#|Series |n°|)(?P<seriePos>\d+)$`),
		regexp.MustCompile(`^Book\s(?P<seriePos>\d+)\sof\s(?P<serie>.+)$`),
	}
)

// GuessSerie uses a simple heuristic to decipher Google books information about
// series hidden in title/subtitle volume information
func GuessSerie(s string) (title string, serieName string, seriePos string) {
	for _, re := range reSerieGuesser {
		if r := submatchMap(re, s); len(r) > 0 {
			return r["title"], r["serie"], r["seriePos"]
		}
	}

	return s, "", ""
}

func submatchMap(re *regexp.Regexp, s string) map[string]string {
	names := re.SubexpNames()
	matches := re.FindStringSubmatch(s)

	r := make(map[string]string)
	for i := range matches {
		if i > 0 {
			r[names[i]] = matches[i]
		}
	}

	return r
}
