package main

import (
	"fmt"
	"github.com/lanzay/Amass/amass/core"
	"math/rand"
	"time"

	"github.com/lanzay/Amass/amass"
)

func main() {
	// Seed the default pseudo-random number generator
	rand.Seed(time.Now().UTC().UnixNano())

	enum := amass.NewEnumeration()
	go func() {
		for result := range enum.Output {
			//fmt.Printf("%10s %10s %s - %s, %v\n",result.Source, result.Tag, result.Domain, result.Name, result.Addresses)
			fmt.Printf("%-15s %-7s %s - %-30s %v\n", result.Source, result.Tag, result.Domain, result.Name, result.Addresses)
		}
	}()

	//configDebug(enum)
	//configProd(enum)
	configNuBut(enum)
	enum.Config.AddDomain("att.com")
	enum.Config.Dir = "./examples/att.com-test"

	//var err error
	//f , err := os.OpenFile("aa.txt", os.O_APPEND|os.O_CREATE, 066)
	//if err != nil{
	//	log.Println(err)
	//}
	//defer f.Close()
	//enum.Config.DataOptsWriter = f

	enum.Start()
}

func configNuBut(enum *amass.Enumeration) {

	enum.Config.Passive = true             //srv.LowNumberOfNames() Only access the data sources for names and return results?
	enum.Config.Active = false             // Determines if zone transfers will be attempted
	enum.Config.MaxDNSQueries = 100        // The maximum number of concurrent DNS queries
	enum.Config.IncludeUnresolvable = true //TODO excl Alterations // Determines if unresolved DNS names will be output by the enumeration

	//enum.Config.DisabledDataSources = enum.GetAllSourceNames()

	enum.Config.AddNumbers = true
	enum.Config.Alterations = true
	enum.Config.FlipWords = true
	enum.Config.FlipNumbers = true
	enum.Config.AddWords = true
	enum.Config.AddNumbers = true
	enum.Config.MinForWordFlip = 2
	//enum.Config.EditDistance = 0
	enum.Config.AltWordlist = []string{"./wordlists/all.txt"}

	enum.Config.BruteForcing = false
	enum.Config.Recursive = true     // Will recursive brute forcing be performed?
	enum.Config.MinForRecursive = 10 // Will the enumeration including brute forcing techniques

	enum.Config.AddAPIKey("Shodan", &core.APIKey{Key: "PSKINdQe1GyxGgecYz2191H2JoS9qvgD"})

	//enum.ProvidedNames = []string{} // Names already known prior to the enumeration

}

func configDebug(enum *amass.Enumeration) {

	enum.Config.IncludeOutOfScope = true
	enum.Config.Passive = false   //srv.LowNumberOfNames() Only access the data sources for names and return results?
	enum.Config.Active = false    // Determines if zone transfers will be attempted
	enum.Config.MaxDNSQueries = 1 // The maximum number of concurrent DNS queries
	//enum.Config.IncludeUnresolvable = true //TODO excl Alterations // Determines if unresolved DNS names will be output by the enumeration

	//enum.Config.DisabledDataSources = enum.GetAllSourceNames()[1:]

	enum.Config.AddNumbers = false
	enum.Config.Alterations = false
	enum.Config.FlipWords = false
	enum.Config.FlipNumbers = false
	enum.Config.AddWords = false
	enum.Config.AddNumbers = false
	enum.Config.MinForWordFlip = 2
	//enum.Config.EditDistance = 0
	enum.Config.AltWordlist = []string{"./wordlists/all.txt"}

	enum.Config.BruteForcing = false
	enum.Config.Recursive = false    // Will recursive brute forcing be performed?
	enum.Config.MinForRecursive = 10 // Will the enumeration including brute forcing techniques

	//enum.Config.AddAPIKey("Shodan", &core.APIKey{Key: "PSKINdQe1GyxGgecYz2191H2JoS9qvgD"})

	//enum.ProvidedNames = []string{} // Names already known prior to the enumeration

}

func configProd(enum *amass.Enumeration) {

	enum.Config.Passive = false     //srv.LowNumberOfNames() Only access the data sources for names and return results?
	enum.Config.Active = true       // Determines if zone transfers will be attempted
	enum.Config.MaxDNSQueries = 100 // The maximum number of concurrent DNS queries
	//enum.Config.IncludeUnresolvable = true //TODO excl Alterations // Determines if unresolved DNS names will be output by the enumeration

	//enum.Config.DisabledDataSources = enum.GetAllSourceNames()

	enum.Config.AddNumbers = true
	enum.Config.Alterations = true
	enum.Config.FlipWords = true
	enum.Config.FlipNumbers = true
	enum.Config.AddWords = true
	enum.Config.AddNumbers = true
	enum.Config.MinForWordFlip = 2
	//enum.Config.EditDistance = 0
	enum.Config.AltWordlist = []string{"./wordlists/all.txt"}

	enum.Config.BruteForcing = true
	enum.Config.Recursive = true     // Will recursive brute forcing be performed?
	enum.Config.MinForRecursive = 10 // Will the enumeration including brute forcing techniques

	enum.Config.AddAPIKey("Shodan", &core.APIKey{Key: "PSKINdQe1GyxGgecYz2191H2JoS9qvgD"})

	//enum.ProvidedNames = []string{} // Names already known prior to the enumeration

}
