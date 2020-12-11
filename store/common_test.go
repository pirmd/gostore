package store

import (
	"strings"
)

var (
	testData = []map[string]interface{}{
		{"Title": "Le cinquième élément", "Authors": []string{"Luc"}, "PublicationDate": 1997, "Read": true},
		{"Title": "Mon père, ce héros", "Authors": []string{"Luc Skywalker"}, "PublicationDate": 1980, "Read": false},
		{"Title": "Les misérables", "Authors": []string{"Victor Hugo"}, "PublicationDate": 1862, "Read": true},
		{"Title": "Le nettoyage pour les nuls", "Authors": []string{"Victor"}, "PublicationDate": 1990, "Read": false},
		{"Title": "Songe d'une nuit d'été", "Authors": []string{"W. Shakespear"}, "PublicationDate": 1595, "Read": true},
		{"Title": "Guère épée", "Authors": []string{"McLeod"}, "PublicationDate": 1986, "Read": true},
		{"Title": "Garder sa tête sur les épaules", "Authors": []string{"McLeod"}, "PublicationDate": 1987, "Read": false},
		{"Title": "Being great again", "Authors": []string{"Trump"}, "PublicationDate": 2021, "Read": false},
		{"Title": "Mes recettes de foie gras", "Authors": []string{"Donald Duck"}, "PublicationDate": 2004, "Read": false},
		{"Title": "Un, deux, trois... Soleil !", "Authors": []string{"Louis XIV"}, "PublicationDate": 1711, "Read": false},
		{"Title": "Brise Marine", "Authors": []string{"Mallarmé"}, "PublicationDate": 1899, "Read": true},
		{"Title": "Philosophies comparées", "Authors": []string{"Donald Trump", "Donald Duck"}, "PublicationDate": 2046, "Read": false},
		{"Title": "Dialogues de sourds", "Authors": []string{"Charles-Michel de l'Épée", "D. Trump"}, "PublicationDate": 2018, "Read": false},
	}

	buildKey = func(m map[string]interface{}) string {
		return m["Authors"].([]string)[0] + "/" + m["Title"].(string) + ".tst"
	}
)

type mockROFile struct {
	*strings.Reader
}

func (f *mockROFile) Close() error {
	return nil
}

func newMockROFile(s string) ReadCloser {
	return &mockROFile{
		Reader: strings.NewReader(s),
	}
}
