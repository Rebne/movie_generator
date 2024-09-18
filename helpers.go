package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type ID struct {
	MovieID string `json:"movieID"`
	Index   int    `json:"index"`
}

type IDdata struct {
	Length int  `json:"length"`
	IDs    []ID `json:"ids"`
}

type MovieData struct {
	Title   string `json:"Title"`
	Year    string `json:"Year"`
	Plot    string `json:"Plot"`
	Runtime string `json:"Runtime"`
	Poster  string `json:"Poster"`
	Genre   string `json:"Genre"`
}

func getMovieIDs(data IDdata) []string {
	movieIDs := make([]string, data.Length)
	for i, id := range data.IDs {
		movieIDs[i] = id.MovieID
	}
	return movieIDs
}

func fetchMovieData(id string) (MovieData, error) {
	url := fmt.Sprintf("http://www.omdbapi.com/?i=%s&apikey=%s", id, API_KEY)
	resp, err := http.Get(url)
	if err != nil {
		return MovieData{}, err
	}
	defer resp.Body.Close()
	var data MovieData
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return MovieData{}, err
	}
	if data.Title == "" {
		return MovieData{}, fmt.Errorf("error, no movie data returned from OMDb API")
	}
	return data, nil
}

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

func formatRuntimeString(duration string) string {
	result := ""
	index := 0
	for index < len(duration) {
		if !isDigit(duration[index]) {
			break
		}
		result += string(duration[index])
		index++
	}

	durationAsInt, err := strconv.Atoi(result)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error converting string %s\n", result)
		return strings.ReplaceAll(duration, " ", "")
	}
	hours := durationAsInt / 60
	minutes := durationAsInt % 60

	return fmt.Sprintf("%dh%dm", hours, minutes)
}

func isValidIMDbID(id string) bool {
	pattern := `^tt\d{7,8}$`
	match, _ := regexp.MatchString(pattern, id)
	return match
}

func addID(dataSet IDdata, movieID string, filename string) error {

	dataSet.IDs = append(dataSet.IDs, ID{MovieID: movieID, Index: dataSet.Length})
	dataSet.Length++
	return writeIdData(filename, dataSet)
}

func removeID(dataSet IDdata, movieID string, filename string) error {
	for i := range dataSet.IDs {
		if dataSet.IDs[i].MovieID == movieID {
			dataSet.IDs = append(dataSet.IDs[:i], dataSet.IDs[i+1:]...)
			dataSet.Length--
			dataSet.reIndexMovieIDs()
			return writeIdData(filename, dataSet)
		}
	}
	return fmt.Errorf("ID not found")
}

func getNewIDs(prevLength int, data IDdata) ([]string, error) {
	var result []string
	for _, item := range data.IDs {
		if item.Index >= prevLength {
			result = append(result, item.MovieID)
		}
	}
	if len(result) == 0 {
		return result, fmt.Errorf("error no new IDS found")
	}
	return result, nil
}

func (data *IDdata) reIndexMovieIDs() {
	for i := 0; i < data.Length; i++ {
		data.IDs[i].Index = i
	}
}

func readIDData(filename string) (IDdata, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return IDdata{}, err
	}

	var data IDdata
	err = json.Unmarshal(file, &data)
	if err != nil {
		return IDdata{}, err
	}
	return data, nil
}

func writeIdData(filename string, data IDdata) error {
	jsonData, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return err
	}
	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		return err
	}
	return nil
}
