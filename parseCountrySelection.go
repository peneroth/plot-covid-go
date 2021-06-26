package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func getRGBAvalue(s string, pos int) (num, pos2 int) {
	if s[pos] == ':' {
		pos++
	} else {
		fmt.Println("Wrong format, expecting :")
		os.Exit(0)
	}
	// Allow space of tab after : character
	for s[pos] == ' ' || s[pos] == '\t' {
		pos++
	}
	pos2 = pos
	for s[pos2] >= '0' && s[pos2] <= '9' {
		pos2++
	}
	num, err := strconv.Atoi(s[pos:pos2])
	if err != nil {
		fmt.Println("Fail to read color uint8")
		panic(err)
	}
	return num, pos2
}

func parseCountrySelection(filename string, jhData setOfJhData) setOfSelectedCountries {
	var selectedCountries setOfSelectedCountries

	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	// Determine size
	scanner := bufio.NewScanner(f)
	selectedCountries.nbrSelectedCountries = 0
	for scanner.Scan() {
		line := scanner.Text()
		if (len(line) > 0) && (strings.Count(line, "#") == 0) {
			selectedCountries.nbrSelectedCountries++
		}
	}
	selectedCountries.nbrSelectedCountries-- // Remove first row (heading)
	f.Close()
	fmt.Println("nbrSelectedCountries = ", selectedCountries.nbrSelectedCountries)

	selectedCountries.country = make([]countryIndex, selectedCountries.nbrSelectedCountries)
	// selCountry := make([]countryIndex, selectedCountries.nbrSelectedCountries)
	f, err = os.Open(filename)
	if err != nil {
		panic(err)
	}
	// Read data
	scanner = bufio.NewScanner(f)
	i := 0
	j := 0
	for scanner.Scan() {
		s := scanner.Text()
		if i > 0 && strings.Count(s, "#") == 0 {
			line := jhLineSplit(s)
			selectedCountries.country[j].country = strings.TrimSpace(line[0])
			selectedCountries.country[j].polulation, err = strconv.ParseFloat(strings.TrimSpace(line[1]), 64)
			if err != nil {
				fmt.Println("Fail to read population")
				panic(err)
			}
			pos := strings.Index(s, "{") + 1
			num := 0
			for s[pos] != '}' {
				switch s[pos] {
				case 'R':
					pos++
					num, pos = getRGBAvalue(s, pos)
					selectedCountries.country[j].lineColor.R = uint8(num)
				case 'G':
					pos++
					num, pos = getRGBAvalue(s, pos)
					selectedCountries.country[j].lineColor.G = uint8(num)
				case 'B':
					pos++
					num, pos = getRGBAvalue(s, pos)
					selectedCountries.country[j].lineColor.B = uint8(num)
				case 'A':
					pos++
					num, pos = getRGBAvalue(s, pos)
					selectedCountries.country[j].lineColor.A = uint8(num)
				case ' ':
					pos++
				case '\t':
					pos++
				case ',':
					pos++
				default:
					fmt.Println("Wrong country selection file format")
					os.Exit(0)
				}
			}
			i++
			j++
		} else {
			i++
		}
	}
	f.Close()

	// Find index in jhData for the selected countries
	for i := 0; i < selectedCountries.nbrSelectedCountries; i++ {
		for j := 0; j < jhData.nbrOfCountries; j++ {
			if strings.Compare(selectedCountries.country[i].country, jhData.country[j].country) == 0 && strings.Compare(jhData.country[j].province, "") == 0 {
				selectedCountries.country[i].jhIndex = j
				fmt.Println("Country =", jhData.country[j].country, "Province =", jhData.country[j].province, "index =", j)
				break
			}
		}
	}

	return selectedCountries
}
