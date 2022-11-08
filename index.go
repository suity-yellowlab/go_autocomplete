package main

import (
	"index/suffixarray"
	"sort"
)

// Example taken from https://blog.gopheracademy.com/advent-2014/string-matching/

// Offset data is needed to link result-ints to slice of structs.
var offsets []int
var index *suffixarray.Index

func buildIndex(entries []AutocompleteTreffer) {
	// the large byte array can be garbage collected after building the index.
	// Program uses multiples of the size of the table to build the array.
	// At current size of about 3mb not a problem for a long time.
	var data []byte
	for i := range entries {
		// TODO Maybe consider manual copy instead of ...  expanding. At the moment time to build index is neglible.
		data = append(data, []byte(entries[i].Anfrage)...)
		offsets = append(offsets, len(data))

	}
	index = suffixarray.New(data)

}

// possible opimization: oversize results as len(idxs) and resize at return.
func searchIndex(query string) []int {
	idxs := index.Lookup([]byte(query), -1)
	var results []int
	for _, idx := range idxs {
		i := sort.Search(len(offsets), func(i int) bool {
			return offsets[i] > idx
		})

		if idx+len(query) <= offsets[i] {
			results = append(results, i)
		}
	}
	return results
}
