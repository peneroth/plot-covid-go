package main

import (
	"fmt"
	"image/color"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

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

type setOfSelectedCountries struct {
	country              []countryIndex
	nbrSelectedCountries int
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
	// fileUrl := "https://raw.githubusercontent.com/CSSEGISandData/COVID-19/master/csse_covid_19_data/csse_covid_19_time_series/time_series_covid19_confirmed_global.csv"
	// filename := "time_series_covid19_confirmed_global.csv"
	downloaded := false
	err := DownloadFile(filename, fileUrl, &downloaded)
	if err != nil {
		panic(err)
	}
	if downloaded {
		fmt.Println("Downloaded: " + filename)
	}

	// Read and parse John Hopkins data
	jhData := parseJHData(filename)

	// Read and parse country selection file
	filename = "selected_countries.txt"
	selCountries := parseCountrySelection(filename, jhData)

	// Process data
	var jhProcData procData
	jhProcData.country = make([]countryProcData, jhData.nbrOfCountries)

	for i := 0; i < jhData.nbrOfCountries; i++ {
		// Calc new Deaths
		jhProcData.country[i].newDeaths = make([]int64, jhData.nbrOfDates)
		jhProcData.country[i].newDeaths[0] = 0
		for j := 1; j < jhData.nbrOfDates; j++ {
			jhProcData.country[i].newDeaths[j] = jhData.country[i].deaths[j] - jhData.country[i].deaths[j-1]
		}
		// Calc average value
		avgSize := 7 // must be an odd value
		avgBoarder := (avgSize - 1) / 2
		jhProcData.country[i].newDeathsMean = make([]int64, jhData.nbrOfDates)
		for j := avgBoarder; j < jhData.nbrOfDates-avgBoarder; j++ {
			meanValue := int64(0)
			for k := j - avgBoarder; k < j+avgBoarder; k++ {
				meanValue += jhProcData.country[i].newDeaths[k]
			}
			jhProcData.country[i].newDeathsMean[j] = meanValue / int64(avgSize)
		}
		// Partial moving agerage for the last dates
		for j := jhData.nbrOfDates - avgBoarder; j < jhData.nbrOfDates; j++ {
			meanValue := int64(0)
			for k := j - avgBoarder; k < jhData.nbrOfDates; k++ {
				meanValue += jhProcData.country[i].newDeaths[k]
			}
			jhProcData.country[i].newDeathsMean[j] = meanValue / int64(jhData.nbrOfDates-(j-avgBoarder))
		}
	}

	// Create plot 1
	p := plot.New()
	p.Title.Text = "Title"
	p.X.Tick.Marker = myTicks{}
	p.Y.Label.Text = "Deaths per one millon"
	p.Add(plotter.NewGrid())

	// fmt.Printf("%T\n", plotLines[0].pts)
	pts := make(plotter.XYs, jhData.nbrOfDates)
	for i := 0; i < selCountries.nbrSelectedCountries; i++ {
		for j := range pts {
			pts[j].X = float64(jhData.dates[j])
			// pts[j].Y = float64(jhProcData.country[selCountry[i].jhIndex].newDeathsMean[j]) / selCountry[i].polulation
			pts[j].Y = float64(jhProcData.country[selCountries.country[i].jhIndex].newDeathsMean[j]) / selCountries.country[i].polulation
		}
		plot_line, err := plotter.NewLine(pts)
		if err != nil {
			log.Panic(err)
		}
		plot_line.Color = selCountries.country[i].lineColor
		// fmt.Printf("%T\n", plot_line.Color)
		p.Add(plot_line)
		p.Legend.Add(selCountries.country[i].country, plot_line)
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
	pts2 := make(plotter.XYs, jhData.nbrOfDates)
	for i := 0; i < selCountries.nbrSelectedCountries; i++ {
		for j := range pts {
			pts2[j].X = float64(jhData.dates[j])
			pts2[j].Y = float64(jhData.country[selCountries.country[i].jhIndex].deaths[j]) / selCountries.country[i].polulation
		}
		plot_line2, err := plotter.NewLine(pts2)
		if err != nil {
			log.Panic(err)
		}
		plot_line2.Color = selCountries.country[i].lineColor
		// fmt.Printf("%T\n", plot_line.Color)
		p2.Add(plot_line2)
		p2.Legend.Add(selCountries.country[i].country, plot_line2)
		p2.Legend.Top = false
	}

	err = p2.Save(40*vg.Centimeter, 20*vg.Centimeter, "plot_output2.png")
	if err != nil {
		log.Panic(err)
	}

}

// MyTicks, based on example from https://github.com/gonum/plot/issues/296
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
