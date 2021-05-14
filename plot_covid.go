package main

import (
	"bufio"
	"fmt"
	"image/color"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

type countryData struct {
	province string
	country  string
	lat      float64
	long     float64
	data     []int64
}

type setOfData struct {
	heading [4]string
	dates_s []string
	dates   []float64
	country []countryData
}

type countryProcData struct {
	newDeaths     []int64
	newDeathsMean []int64
}

type procData struct {
	country []countryProcData
}

type countryIndex struct {
	country    string
	polulation float64
	jhIndex    int
	lineColor  color.RGBA
}

func exitError(err error) {
	if err != nil {
		panic(err)
	}
}

func isNewerThanHours(t time.Time, hours time.Duration) bool {
	return time.Since(t) < hours*time.Hour
}

// From https://golangcode.com/download-a-file-from-a-url/
// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func DownloadFile(filename string, url string, downloaded *bool) error {

	*downloaded = false
	// Check if the file already exist and is newer than xx hours
	if fileStat, err := os.Stat(filename); os.IsNotExist(err) {
		fmt.Println("file does not exist")
	} else {
		fmt.Println("file does exist")
		if isNewerThanHours(fileStat.ModTime(), 1) {
			fmt.Println("file younger than 1 hours")
			return err
		}
	}

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	*downloaded = true
	return err
}

func main() {

	// Download file from John Hopkins
	fileUrl := "https://raw.githubusercontent.com/CSSEGISandData/COVID-19/master/csse_covid_19_data/csse_covid_19_time_series/time_series_covid19_deaths_global.csv"
	filename := "time_series_covid19_deaths_global.csv"
	// filename := "test.csv"
	downloaded := false
	err := DownloadFile(filename, fileUrl, &downloaded)
	if err != nil {
		panic(err)
	}
	if downloaded {
		fmt.Println("Downloaded: " + filename)
	}

	//
	// Determine data set size (number of dates and number of regions/countries)
	//
	f1, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	// Read line by line
	scanner1 := bufio.NewScanner(f1)
	nbrOfCountries := 0
	nbrOfDates := 0
	for scanner1.Scan() {
		if nbrOfCountries == 0 {
			line := scanner1.Text()
			nbrOfDates = strings.Count(line, ",") - 3
		}
		nbrOfCountries++
	}
	nbrOfCountries-- // Remove first row (heading)
	f1.Close()
	fmt.Println(nbrOfCountries)
	fmt.Println(nbrOfDates)

	//
	// Read the data
	//
	// Allocate memory
	var jhdata setOfData
	jhdata.dates_s = make([]string, nbrOfDates)
	jhdata.dates = make([]float64, nbrOfDates)
	jhdata.country = make([]countryData, nbrOfCountries)
	f2, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	// Read first line
	scanner2 := bufio.NewScanner(f2)
	if !scanner2.Scan() {
		fmt.Println("Reached end of file")
		os.Exit(0)
	}
	line := strings.Split(scanner2.Text(), ",")
	jhdata.heading[0] = line[0]
	jhdata.heading[1] = line[1]
	jhdata.heading[2] = line[2]
	jhdata.heading[3] = line[3]
	for i := 0; i < nbrOfDates; i++ {
		jhdata.dates_s[i] = line[i+4]
		s := strings.Split(jhdata.dates_s[i], "/")
		month, err := strconv.Atoi(s[0])
		exitError(err)
		day, err := strconv.Atoi(s[1])
		exitError(err)
		year, err := strconv.Atoi(s[2])
		exitError(err)
		jhdata.dates[i] = float64(time.Date(year+2000, time.Month(month), day, 0, 0, 0, 0, time.UTC).Unix())
		// fmt.Println(jhdata.dates_s[i], jhdata.dates[i])
	}
	for i := 0; i < nbrOfCountries; i++ {
		scanner2.Scan()
		s := scanner2.Text()
		// Dirty fix!
		s = strings.Replace(s, "\"Korea, South\"", "Korea; South", -1)
		s = strings.Replace(s, "\"Bonaire, Sint Eustatius and Saba\"", "Bonaire; Sint Eustatius and Saba", -1)
		s = strings.Replace(s, "\"Saint Helena, Ascension and Tristan da Cunha\"", "Saint Helena; Ascension and Tristan da Cunha", -1)
		line = strings.Split(s, ",")
		// fmt.Println("i = ", i, " len(line) = ", len(line))
		jhdata.country[i].province = line[0]
		jhdata.country[i].country = line[1]
		jhdata.country[i].lat, err = strconv.ParseFloat(line[2], 64)
		if err != nil {
			jhdata.country[i].lat = 0
		}
		jhdata.country[i].long, err = strconv.ParseFloat(line[3], 64)
		if err != nil {
			jhdata.country[i].long = 0
		}
		jhdata.country[i].data = make([]int64, nbrOfDates)
		for j := 0; j < nbrOfDates; j++ {
			jhdata.country[i].data[j], err = strconv.ParseInt(line[j+4], 10, 64)
			if err != nil {
				fmt.Println("err 3")
				panic(err)
			}
		}
	}
	f2.Close()

	fmt.Println(jhdata.country[71].country)
	fmt.Println(jhdata.country[71].province)
	fmt.Println(jhdata.country[71].data[12])

	// Country selection and polulation
	nbrPlotCountries := 4
	selCountry := make([]countryIndex, nbrPlotCountries)

	selCountry[0].country = "Sweden"
	selCountry[0].polulation = 10
	selCountry[0].lineColor = color.RGBA{G: 255, A: 255}
	selCountry[1].country = "Italy"
	selCountry[1].polulation = 60
	selCountry[1].lineColor = color.RGBA{R: 255, A: 255}
	selCountry[2].country = "France"
	selCountry[2].polulation = 67
	selCountry[2].lineColor = color.RGBA{B: 255, A: 255}
	selCountry[3].country = "Denmark"
	selCountry[3].polulation = 5.8
	selCountry[3].lineColor = color.RGBA{B: 255, R: 255, A: 255}
	for i := 0; i < nbrPlotCountries; i++ {
		for j := 0; j < nbrOfCountries; j++ {
			if strings.Compare(selCountry[i].country, jhdata.country[j].country) == 0 && strings.Compare(jhdata.country[j].province, "") == 0 {
				println("index = ", j, "C = ", jhdata.country[j].country, "P = ", jhdata.country[j].province)
				selCountry[i].jhIndex = j
				break
			}
		}
	}

	// Process data
	var jhProcData procData
	jhProcData.country = make([]countryProcData, nbrOfCountries)

	for i := 0; i < nbrOfCountries; i++ {
		// Calc new Deaths
		jhProcData.country[i].newDeaths = make([]int64, nbrOfDates)
		jhProcData.country[i].newDeaths[0] = 0
		for j := 1; j < nbrOfDates; j++ {
			jhProcData.country[i].newDeaths[j] = jhdata.country[i].data[j] - jhdata.country[i].data[j-1]
		}
		// Calc average value
		avgSize := 7 // must be an odd value
		avgBoarder := (avgSize - 1) / 2
		jhProcData.country[i].newDeathsMean = make([]int64, nbrOfDates)
		for j := avgBoarder; j < nbrOfDates-avgBoarder; j++ {
			meanValue := int64(0)
			for k := j - avgBoarder; k < j+avgBoarder; k++ {
				meanValue += jhProcData.country[i].newDeaths[k]
			}
			jhProcData.country[i].newDeathsMean[j] = meanValue / int64(avgSize)
		}
	}

	// Create plot

	// xticks defines how we convert and display time.Time values.
	// xticks := plot.TimeTicks{Format: "2006-01-02"}

	p := plot.New()
	p.Title.Text = "Title"
	// p.X.Tick.Marker = xticks
	p.Y.Label.Text = "Deaths per one millon"
	p.Add(plotter.NewGrid())

	// fmt.Printf("%T\n", plotLines[0].pts)
	pts := make(plotter.XYs, nbrOfDates)
	for i := 0; i < nbrPlotCountries; i++ {
		for j := range pts {
			pts[j].X = jhdata.dates[j]
			pts[j].Y = float64(jhProcData.country[selCountry[i].jhIndex].newDeathsMean[j]) / selCountry[i].polulation
		}
		plot_line, err := plotter.NewLine(pts)
		if err != nil {
			log.Panic(err)
		}
		plot_line.Color = selCountry[i].lineColor
		// fmt.Printf("%T\n", plot_line.Color)
		p.Add(plot_line)
	}

	err = p.Save(30*vg.Centimeter, 15*vg.Centimeter, "plot_output.png")
	if err != nil {
		log.Panic(err)
	}

}
