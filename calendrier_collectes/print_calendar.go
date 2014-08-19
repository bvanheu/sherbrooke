// Copyright (c) 2014, Benjamin Vanheuverzwijn <bvanheu@gmail.com>
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//     * Redistributions of source code must retain the above copyright
//       notice, this list of conditions and the following disclaimer.
//     * Redistributions in binary form must reproduce the above copyright
//       notice, this list of conditions and the following disclaimer in the
//       documentation and/or other materials provided with the distribution.
//     * The names of its contributors may be used to endorse or promote
//       products derived from this software without specific prior written
//       permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED. IN NO EVENT SHALL Benjamin Vanheuverzwijn BE LIABLE FOR ANY
// DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
// (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
// LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
// ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF
// THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

//
// This software is used to parse a JSON Sherbrooke file into a printable
// calendar.
//
// USAGE
//  1/ Get the JSON data file from Sherbrooke city website
//      2014 url - http://donnees.ville.sherbrooke.qc.ca/dataset/calendrier-des-collectes/resource/281d2e78-d942-495a-a5db-eeb7c704ac3b
//  2/ Run the print_calendar software with the json file as the first argument
//      (see examples)
//
// EXAMPLE
//  print to stdout
//      $ go run print_calendar.go
//  print to printer
//      $ go run print_calendar.go | lp -o columns=2 -o page-top=72
//

package main

import "encoding/json"
import "fmt"
import "os"
import "path"
import "log"
import "io/ioutil"
import "time"

type Period struct {
	Number      string
	Types       string
	Date_begin  time.Time
	Date_end    time.Time
	Information string
}

type Calendar struct {
	District_name string
	Periods       []Period
}

type CollecteMatieresResiduelle struct {
	Municipality_id string `json:"MUNID"`
	Code_id         string `json:"CODEID"`
	Week_number     string `json:"NO_SEM"`
	Date_begin      string `json:"DT01"`
	Date_end        string `json:"DT02"`
	District        string `json:"ARROND"`
	Type            string `json:"TYPE"`
	Description     string `json:"DESC"`
	Information     string `json:"INFO"`
}

// Json structure
type jsonobject struct {
	CalendrierCollectes struct {
		CollecteMatieresResiduelles []CollecteMatieresResiduelle `json:"COLLECTE_MATIERES_RESIDUELLES"`
	} `json:"CALENDRIER_COLLECTES"`
}

func usage() {
	fmt.Printf("Usage: %s [OPTION] FILE.json\n\n", path.Base(os.Args[0]))
}

func init_log() {
	log.SetPrefix(path.Base(os.Args[0]) + " - ")
	log.SetOutput(os.Stdout)
}

func main() {
	var json_buffer []byte
	var m jsonobject
	var err error

	init_log()

	if len(os.Args) != 2 {
		usage()
		log.Fatal("not enough arguments")
	}

	// Read the JSON file
	json_buffer, err = ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	// JSON -> Objects
	err = json.Unmarshal(json_buffer, &m)
	if err != nil {
		log.Fatal(err)
	}

	// Objects -> calendar
	var previous_cmr CollecteMatieresResiduelle = m.CalendrierCollectes.CollecteMatieresResiduelles[0]
	var montbellevue Calendar
	var period Period

	for _, cmr := range m.CalendrierCollectes.CollecteMatieresResiduelles {
		if cmr.District == "Arrondissement du Mont-Bellevue" {
			if cmr.Week_number != previous_cmr.Week_number {
				// Flush the current period in the slice
				montbellevue.Periods = append(montbellevue.Periods, period)
				// Start a new period
				period.Types = ""
				period.Information = ""
			}

			period.Number = cmr.Week_number
			period.Date_begin, err = time.Parse("2006-01-02", cmr.Date_begin)
			if err != nil {
				log.Fatal(err)
			}
			period.Date_end, err = time.Parse("2006-01-02", cmr.Date_end)
			if err != nil {
				log.Fatal(err)
			}
			period.Types += cmr.Type
			period.Information += cmr.Information

			previous_cmr = cmr
		}
	}
	// Flush the last entry
	montbellevue.Periods = append(montbellevue.Periods, period)

	// Format the parsed data
	var previous_time, current_time string

	for _, p := range montbellevue.Periods {
		current_time = (p.Date_begin.Format("January 2006"))
		if current_time != previous_time {
			fmt.Println("\n\n" + current_time)
			previous_time = current_time
		}

		// Print week number - types - date begin - date end
		fmt.Println(p.Number +
			"\t" + p.Types +
			"\t" + p.Date_begin.Format("01-02") +
			"\t" + p.Date_end.Format("01-02"))

		if p.Information != "" {
			fmt.Println(" `-> " + p.Information)
		}
	}

	fmt.Println("\nLegende\n" +
		"D - Déchets\n" +
		"R - Récupération\n" +
		"C - Compost\n" +
		"S - Sapin\n" +
		"E - Encombrant et bois\n" +
		"B - Carton\n" +
		"F - Feuilles mortes")
}
