package main

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func connect(settings *Settings) func() error {
	var err error
	db, err = sql.Open("mysql", settings.DBString)
	if err != nil {
		// If the DB cannot be opened the program should kill itself. The service log will handle the error data.
		log.Fatal(err)
	}
	return db.Close
}

/*
Autocomplete data is generated by successfull user queries and stored in the DB by the main application.
Anfrage is the literal string query. Treffer is the amount of results generated by the query at that time.
*/
func loadAutoCompleteData() []AutocompleteTreffer {
	rows, err := db.Query("SELECT Id,Anfrage,Treffer FROM suche_autocomplete")
	if err != nil {
		// If the DB cannot be queried the program should kill itself. The service log will handle the error data.
		log.Fatal(err)
	}
	defer rows.Close()
	entries := make([]AutocompleteTreffer, 0)
	for rows.Next() {
		entry := AutocompleteTreffer{}
		err = rows.Scan(&entry.Id, &entry.Anfrage, &entry.Treffer)
		if err != nil {
			// If the DB cannot be queried the program should kill itself. The service log will handle the error data.
			// Consider changing so that bad rows do not kill the program. However this should never fail.
			log.Fatal(err)
		}
		entries = append(entries, entry)
	}
	return entries
}
