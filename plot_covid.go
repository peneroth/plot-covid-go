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
	deaths   []int64
}

type setOfJhData struct {
	heading [4]string
	dates_s []string
	dates   []int64
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
	fmt.Println("nbrOfCountries = ", nbrOfCountries)
	fmt.Println("nbrOfDates = ", nbrOfDates)

	//
	// Read the data
	//
	// Allocate memory
	var jhData setOfJhData
	jhData.dates_s = make([]string, nbrOfDates)
	jhData.dates = make([]int64, nbrOfDates)
	jhData.country = make([]countryData, nbrOfCountries)
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
	jhData.heading[0] = line[0]
	jhData.heading[1] = line[1]
	jhData.heading[2] = line[2]
	jhData.heading[3] = line[3]
	for i := 0; i < nbrOfDates; i++ {
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
	for i := 0; i < nbrOfCountries; i++ {
		scanner2.Scan()
		s := scanner2.Text()
		// Dirty fix!
		s = strings.Replace(s, "\"Korea, South\"", "Korea; South", -1)
		s = strings.Replace(s, "\"Bonaire, Sint Eustatius and Saba\"", "Bonaire; Sint Eustatius and Saba", -1)
		s = strings.Replace(s, "\"Saint Helena, Ascension and Tristan da Cunha\"", "Saint Helena; Ascension and Tristan da Cunha", -1)
		line = strings.Split(s, ",")
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
		jhData.country[i].deaths = make([]int64, nbrOfDates)
		for j := 0; j < nbrOfDates; j++ {
			jhData.country[i].deaths[j], err = strconv.ParseInt(line[j+4], 10, 64)
			if err != nil {
				fmt.Println("err 3")
				panic(err)
			}
		}
	}
	f2.Close()

	// Country selection and polulation
	nbrPlotCountries := 4
	selCountry := make([]countryIndex, nbrPlotCountries)

	selCountry[0].country = "Sweden"
	selCountry[0].polulation = 10
	selCountry[0].lineColor = color.RGBA{B: 255, A: 255}
	selCountry[1].country = "Italy"
	selCountry[1].polulation = 60
	selCountry[1].lineColor = color.RGBA{R: 255, A: 255}
	selCountry[2].country = "France"
	selCountry[2].polulation = 67
	selCountry[2].lineColor = color.RGBA{G: 255, A: 255}
	selCountry[3].country = "Denmark"
	selCountry[3].polulation = 5.8
	selCountry[3].lineColor = color.RGBA{B: 255, R: 255, A: 255}
	for i := 0; i < nbrPlotCountries; i++ {
		for j := 0; j < nbrOfCountries; j++ {
			if strings.Compare(selCountry[i].country, jhData.country[j].country) == 0 && strings.Compare(jhData.country[j].province, "") == 0 {
				fmt.Println("Country =", jhData.country[j].country, "Province =", jhData.country[j].province, "index =", j)
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
			jhProcData.country[i].newDeaths[j] = jhData.country[i].deaths[j] - jhData.country[i].deaths[j-1]
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
		// Partial moving agerage for the last dates
		for j := nbrOfDates - avgBoarder; j < nbrOfDates; j++ {
			meanValue := int64(0)
			for k := j - avgBoarder; k < nbrOfDates; k++ {
				meanValue += jhProcData.country[i].newDeaths[k]
			}
			jhProcData.country[i].newDeathsMean[j] = meanValue / int64(nbrOfDates-(j-avgBoarder))
		}
	}

	// Create plot 1
	p := plot.New()
	p.Title.Text = "Title"
	p.X.Tick.Marker = myTicks{}
	p.Y.Label.Text = "Deaths per one millon"
	p.Add(plotter.NewGrid())

	// fmt.Printf("%T\n", plotLines[0].pts)
	pts := make(plotter.XYs, nbrOfDates)
	for i := 0; i < nbrPlotCountries; i++ {
		for j := range pts {
			pts[j].X = float64(jhData.dates[j])
			pts[j].Y = float64(jhProcData.country[selCountry[i].jhIndex].newDeathsMean[j]) / selCountry[i].polulation
		}
		plot_line, err := plotter.NewLine(pts)
		if err != nil {
			log.Panic(err)
		}
		plot_line.Color = selCountry[i].lineColor
		// fmt.Printf("%T\n", plot_line.Color)
		p.Add(plot_line)
		p.Legend.Add(selCountry[i].country, plot_line)
		p.Legend.Top = true
	}

	err = p.Save(40*vg.Centimeter, 20*vg.Centimeter, "plot_output.png")
	if err != nil {
		log.Panic(err)
	}

	// Create plot 2
	p2 := plot.New()
	p2.Title.Text = "Title"
	p2.X.Tick.Marker = myTicks{}
	p2.Y.Label.Text = "Deaths per one millon"
	p2.Add(plotter.NewGrid())

	// fmt.Printf("%T\n", plotLines[0].pts)
	pts2 := make(plotter.XYs, nbrOfDates)
	for i := 0; i < nbrPlotCountries; i++ {
		for j := range pts {
			pts2[j].X = float64(jhData.dates[j])
			pts2[j].Y = float64(jhData.country[selCountry[i].jhIndex].deaths[j]) / selCountry[i].polulation
		}
		plot_line2, err := plotter.NewLine(pts2)
		if err != nil {
			log.Panic(err)
		}
		plot_line2.Color = selCountry[i].lineColor
		// fmt.Printf("%T\n", plot_line.Color)
		p2.Add(plot_line2)
		p2.Legend.Add(selCountry[i].country, plot_line2)
		p2.Legend.Top = false
	}

	err = p2.Save(40*vg.Centimeter, 20*vg.Centimeter, "plot_output2.png")
	if err != nil {
		log.Panic(err)
	}

}

// MyTicks, based on example at https://github.com/gonum/plot/issues/296
type myTicks struct{}

// Ticks returns Ticks in the specified range.
func (myTicks) Ticks(min, max float64) []plot.Tick {
	if max <= min {
		panic("illegal range")
	}
	var ticks []plot.Tick
	for i := min; i <= max; i += (86400) {
		myTime := time.Unix(int64(i), 0)
		myYear, myMonth, myDay := myTime.Date()
		if myDay == 1 {
			myDate := IntToString(myYear) + "-" + IntToString(int(myMonth)) + "-" + IntToString(myDay)
			ticks = append(ticks, plot.Tick{Value: i, Label: myDate})
		}
	}
	return ticks
}

func IntToString(i int) string {
	var s string
	if i < 10 {
		s = "0" + strconv.Itoa(i)
	} else {
		s = strconv.Itoa(i)
	}
	return s
}
