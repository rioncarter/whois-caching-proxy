// Copyright 2018 Rion Carter
//
// Dual-licensed under the terms of the Apache 2 License or GPL v2
package main

import (
	json2 "encoding/json"
	"flag"
	"github.com/domainr/whois"
	"log"
	"net/http"
	"regexp"
	"strings"
	"thunderbird-domain-age-checker-golang-server/persist"
)

var verbose *bool
var p persist.Persist
var suffixes []string
func main() {
	bindPort := flag.String("BindPort", ":9091", "Enter the local port that the Whois caching server should bind to. Default value is ':9091'")
	verbose = flag.Bool("VerboseLog", false, "Set to 'true' to see verbose log messages in stdout/stderr")
	flag.Parse()

	//
	// Initialize database
	p = persist.Persist{}
	p.Setup("domainage.sqlite")

	//
	// Get TLD Suffixes
	suffixes = strings.Split(publicsuffixes, "\n")

	apiHandler := http.NewServeMux()
	apiHandler.HandleFunc("/checkDomain/", WhoisApiHandler)


	//
	// Start the API server listening
	err := http.ListenAndServe(*bindPort, apiHandler)
	if err != nil{
		log.Fatal(err)
	}

}

//
// Does the work of Whois checking and caching
func WhoisApiHandler(writer http.ResponseWriter, r *http.Request) {
	//
	// Get the URL to check
	urlPieces := strings.Split(r.URL.Path, "/")
	checkDomain := urlPieces[2]
	if *verbose{
		log.Println(r.URL.Path)
		log.Println(checkDomain)
	}

	//
	// Ensure we handle subdomains correctly (reference the public suffixes list)
	domainPieces := strings.Split(checkDomain, ".")
	domainLen := len(domainPieces)
	if domainLen > 2{
		tld := domainPieces[domainLen -1]		// tld			(uk)
		nTld := domainPieces[domainLen -2]		// domain		(co)
		nnTld := domainPieces[domainLen -3]		// subdomain	(theregister)
		nTldCombined := nTld+"."+tld
		nnTldCombined := nnTld+"."+nTld+"."+tld

		// Is the 2nd level domain in the suffix list?
		inSuffixList := false
		for _, suffix := range suffixes{
			if suffix == nTldCombined{
				inSuffixList =true
				break
			}
		}

		//
		// Handle IANA registered 2nd level domains vs privately held 2nd level domains
		if inSuffixList{
			// Path if there is a 2ndlevel.domain.tld (IANA)
			checkDomain = nnTldCombined
		} else {
			// Path if the 2ndlevel.domain.tld is privately held
			checkDomain = nTldCombined
		}
	}

	//
	// If we already have the domain cached, return it immediately
	//
	cachedDomain := p.DomainDetails(checkDomain)
	if cachedDomain != nil {
		json, jErr := json2.MarshalIndent(cachedDomain, "", "  ")
		if jErr != nil{
			log.Println("Error Marshaling a cached domain for return: " + cachedDomain.Name)
			log.Fatal(cachedDomain)
		}

		//
		// Write out a response to the caller
		if *verbose{
			// Write out the JSON response to the caller
			log.Println(string(json))
		}
		writer.WriteHeader(200)
		writer.Write([]byte(json))
		return
	}


	//
	//
	// If we DO NOT have the domain cached, take the long path
	//
	//


	//
	// Do the whois check on the appropriate domain
	whoisRequest, _ := whois.NewRequest(checkDomain)            // Returns a prepared whois.Request
	whoisResponse, _ := whois.DefaultClient.Fetch(whoisRequest) // Fetches the request, returns a whois.Response


	//
	// Get the registration date line
	createDateRegexp, _ := regexp.Compile("(?i)(Registered on:.+)|(Creation Date:.+)")
	dateLine := createDateRegexp.FindString(string(whoisResponse.Body))
	dateFormat := ""
	date := ""

	//
	// Check if there is a year and a month (2018-12)
	yearMonthCheck, _ := regexp.Compile("[0-9]{4}-[0-9]{2}")
	ymc := yearMonthCheck.FindString(dateLine)
	if ymc != ""{
		dateFormat="YYYY-MM"
		date = ymc
	}

	//
	// Check if there is a month and a year (12-2018)
	if dateFormat == "" {
		monthYearCheck, _ := regexp.Compile("[0-9]{2}-[0-9]{4}")
		myc := monthYearCheck.FindString(dateLine)
		if myc != ""{
			dateFormat = "MM-YYYY"
			date = myc
		}
	}

	//
	// Check if there is an abbreviated month and year (aug-1996)
	if dateFormat == "" {
		abbMonthYearCheck, _ := regexp.Compile("(?i)(jan|feb|mar|apr|may|jun|jul|aug|sep|oct|nov|dec)-[0-9]{4}")
		amyc := abbMonthYearCheck.FindString(dateLine)
		if amyc != ""{
			dateFormat = "mmm-YYYY"
			date = amyc
		}
	}



	//
	// Normalize to year-month format (2018-12)
	dateRaw := date
	switch dateFormat{
	case "mmm-YYYY":
		// Get number dates from month abbreviations and reverse to get YYYY-MM
		datePieces := strings.Split(date, "-")

		// Need to get a number for each month abbreviation
		mm := ""
		abb := strings.ToUpper(datePieces[0])
		switch abb{
		case "JAN":
			mm = "01"
		case "FEB":
			mm = "02"
		case "MAR":
			mm = "03"
		case "APR":
			mm = "04"
		case "MAY":
			mm = "05"
		case "JUN":
			mm = "06"
		case "JUL":
			mm = "07"
		case "AUG":
			mm = "08"
		case "SEP":
			mm = "09"
		case "OCT":
			mm = "10"
		case "NOV":
			mm = "11"
		case "DEC":
			mm = "12"
		}

		date = datePieces[1] + "-" + mm

	case "MM-YYYY":
		// Reverse to get YYYY-MM
		datePieces := strings.Split(date, "-")
		date = datePieces[1] + "-" + datePieces[0]
	default:
		// Nothing to convert
	}


	//
	// Store the domain in the sqlite db
	iErr := p.InsertDomain(checkDomain, dateRaw, date)
	if iErr != nil{
		if !strings.Contains(iErr.Error(), "UNIQUE constraint failed"){
			log.Println("Unable to insert domain into db: " + checkDomain)
			log.Fatal(iErr)
		}
	}


	//
	// Prepare a JSON response for the caller
	d := p.DomainDetails(checkDomain)
	json, jErr := json2.MarshalIndent(d, "", "  ")
	if jErr != nil{
		log.Println("Error Marshaling a domain for return: " + cachedDomain.Name)
		log.Fatal(cachedDomain)
	}

	//
	// Write out a response to the caller
	writer.WriteHeader(200)
	writer.Write([]byte(json))
}