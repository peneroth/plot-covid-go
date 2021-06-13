package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func jhLineSplit(s string) []string {

	startpos := 0
	endpos := 1
	var words []string

	for endpos < len(s) {
		// fmt.Println(s[endpos-1 : endpos])
		if strings.Compare(s[endpos-1:endpos], "\"") == 0 {
			endpos++
			// fmt.Println(s[endpos-1 : endpos])
			for strings.Compare(s[endpos-1:endpos], "\"") != 0 {
				endpos++
				// fmt.Println(s[endpos-1 : endpos])
			}
		}
		if strings.Compare(s[endpos-1:endpos], ",") == 0 {
			substring := s[startpos : endpos-1]
			words = append(words, substring)
			// fmt.Println(substring)
			startpos = endpos
			endpos = startpos + 1
		} else {
			endpos++
		}
	}
	substring := s[startpos:endpos]
	words = append(words, substring)
	// fmt.Println(substring)
	return words
}

func parseJHData(filename string) setOfJhData {
	var jhData setOfJhData

	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	// Read line by line
	scanner := bufio.NewScanner(f)
	jhData.nbrOfCountries = 0
	jhData.nbrOfDates = 0
	for scanner.Scan() {
		if jhData.nbrOfCountries == 0 {
			line := scanner.Text()
			jhData.nbrOfDates = strings.Count(line, ",") - 3
		}
		jhData.nbrOfCountries++
	}
	jhData.nbrOfCountries-- // Remove first row (heading)
	f.Close()
	fmt.Println("nbrOfCountries = ", jhData.nbrOfCountries)
	fmt.Println("nbrOfDates = ", jhData.nbrOfDates)

	//
	// Read the data
	//
	// Allocate memory
	jhData.dates_s = make([]string, jhData.nbrOfDates)
	jhData.dates = make([]int64, jhData.nbrOfDates)
	jhData.country = make([]countryData, jhData.nbrOfCountries)
	f, err = os.Open(filename)
	if err != nil {
		panic(err)
	}
	// Read first line
	scanner = bufio.NewScanner(f)
	if !scanner.Scan() {
		fmt.Println("Reached end of file")
		os.Exit(0)
	}
	line := strings.Split(scanner.Text(), ",")
	jhData.heading[0] = line[0]
	jhData.heading[1] = line[1]
	jhData.heading[2] = line[2]
	jhData.heading[3] = line[3]
	for i := 0; i < jhData.nbrOfDates; i++ {
		jhData.dates_s[i] = line[i+4]
		s := strings.Split(jhData.dates_s[i], "/")
		month, err := strconv.Atoi(s[0])
		exitError(err)
		day, err := strconv.Atoi(s[1])
		exitError(err)
		year, err := strconv.Atoi(s[2])
		exitError(err)
		jhData.dates[i] = time.Date(year+2000, time.Month(month), day, 0, 0, 0, 0, time.UTC).Unix()
	}
	for i := 0; i < jhData.nbrOfCountries; i++ {
		scanner.Scan()
		s := scanner.Text()
		line = jhLineSplit(s)
		jhData.country[i].province = line[0]
		jhData.country[i].country = line[1]
		jhData.country[i].lat, err = strconv.ParseFloat(line[2], 64)
		if err != nil {
			jhData.country[i].lat = 0
		}
		jhData.country[i].long, err = strconv.ParseFloat(line[3], 64)
		if err != nil {
			jhData.country[i].long = 0
		}
		jhData.country[i].deaths = make([]int64, jhData.nbrOfDates)
		for j := 0; j < jhData.nbrOfDates; j++ {
			jhData.country[i].deaths[j], err = strconv.ParseInt(line[j+4], 10, 64)
			if err != nil {
				fmt.Println("err 3")
				panic(err)
			}
		}
	}
	f.Close()
	return jhData
}
