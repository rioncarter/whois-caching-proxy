package persist

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"time"
)


type Persist struct{
	dbPath string
	db *sql.DB
}

//
// Setup the database object
//		This will create a new database if one doesn't exist
func (p *Persist) Setup(dbPath string) error{
	// Assign the sqlite db path
	p.dbPath = dbPath

	//
	// Check if DB exists
	dbAlreadyExists := true
	if _, err := os.Stat(p.dbPath); os.IsNotExist(err){
		dbAlreadyExists = false
	}

	// Get a handle to the SQLite database, using mattn/go-sqlite3
	var err error
	p.db, err = sql.Open("sqlite3", p.dbPath)
	if err != nil{
		return err
	}

	//
	// Getting a handle to the sqlite database will create it. We'll need to populate if it is 'empty'
	if !dbAlreadyExists{
		dbCreateErr := p.createDatabase()
		if dbCreateErr != nil{
			log.Println("Error creating/populating domain database")
			log.Fatal(dbCreateErr)
		}
	}

	return nil
}

// Internal function to populate the database schema (if it doens't already exist)
func (p *Persist) createDatabase() error{
	_, createTableErr := p.db.Exec(`CREATE TABLE domains (
	uid	INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT UNIQUE,
	name	TEXT NOT NULL UNIQUE,
	registration_date	TEXT NOT NULL,
	registration_date_normalized	TEXT NOT NULL
);`)

	if createTableErr != nil{
		return createTableErr
	}

	return nil
}

//
// Insert a domain
func (p *Persist) InsertDomain(domain string, regDateRaw string, regDateNormalized string) error{
	_, insErr := p.db.Exec("INSERT INTO domains(name, registration_date, registration_date_normalized) VALUES(?, ?, ?)",
		domain, regDateRaw, regDateNormalized)

	if insErr != nil{
		return insErr
	}

	return nil
}


//
// Return a Domain object for the domain in question
func (p *Persist) DomainDetails(domain string) *Domain{
	// Get the DB row entries
	entries, err := p.db.Query("select * from domains where name = ?", domain)
	if err != nil{
		log.Println("Unable to retrieve database rows for domain: " + domain)
		log.Fatal(err)
	}
	defer entries.Close()

	// Decode the rows (hopefully just ONE row)
	var domainEntries []Domain
	for entries.Next(){
		d := Domain{}

		rowDecodeErr := entries.Scan(&d.uid, &d.Name, &d.RegisteredRaw, &d.Registered)
		if rowDecodeErr != nil{
			log.Println("Unable to decode database row for domain: " + domain)
			log.Fatal(rowDecodeErr)
		}

		// Ensure golang date object is available
		date, dateParseErr := time.Parse("2006-01", d.Registered)
		if dateParseErr != nil{
			// Only throw an error if the registered date isn't an empty string
			//	Some registrars DO NOT list the registration date for domains!
			if d.Registered != "" {
				log.Println("Unable to parse date for domain: " + domain + ". Attempted to parse: " + d.Registered)
				log.Fatal(dateParseErr)
			}
		}
		d.RegisteredDate = date

		// Using a slice in case there happens to be more than one entry (shouldn't happen)
		domainEntries = append(domainEntries, d)
	}

	// Final check for errors during the SQL query
	entriesDecodeErr := entries.Err()
	if entriesDecodeErr != nil{
		log.Println("Error while decoding database rows for domain: " + domain)
		log.Fatal(entriesDecodeErr)
	}

	// No domain found
	if len(domainEntries) < 1{
		return nil
	}

	return &domainEntries[0]
}