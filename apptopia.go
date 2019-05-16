package apptopiaTransform

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

// here is my comment. but maybe I need to add this thing to make it work right
// balls
func ParsePublisherDataFromArray(records []string) ([]Publisher, error) {

	ps := make([]Publisher, 0)

	for _, record := range records {
		var p Publisher

		record = strings.Replace(record, `\x1f`, ``, -1)

		err := json.Unmarshal([]byte(record), &p)
		if err != nil {
			fmt.Println(record)
			return ps, err
		}

		ps = append(ps, p)
	}

	return ps, nil
}

func ParseAppDataFromArray(records []string, username, store string, logger *log.Logger) ([]App, error) {

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in ParseAppDataFromArray()", r)
			logger.Println("Recovered in ParseAppDataFromArray()", r)
		}
	}()

	as := make([]App, 0)

	var sdkLookup *os.File
	var err error

	switch store {
	case "google":
		sdkLookup, err = os.Open(fmt.Sprintf("c:/Users/%v/documents/s3_output/google_play_sdks_sdks_output.csv", username))
		if err != nil {
			return as, err
		}
	case "itunes":
		sdkLookup, err = os.Open(fmt.Sprintf("c:/Users/%v/documents/s3_output/itunes_connect_sdks_sdks_output.csv", username))
		if err != nil {
			return as, err
		}
	default:

	}

	defer sdkLookup.Close()

	sdkCSVReader := csv.NewReader(sdkLookup)

	sdkcsvArray, err := sdkCSVReader.ReadAll()
	if err != nil {
		return as, err
	}

	sdkMap := make(map[string]string)

	for i := range sdkcsvArray {
		if i == 0 {
			continue
		}
		sdkMap[sdkcsvArray[i][2]] = sdkcsvArray[i][0]
	}

	file, err := os.Open("data/categories.csv")
	if err != nil {
		fmt.Println(err)
		return as, err
	}

	defer file.Close()

	csvReader := csv.NewReader(file)

	cArray, err := csvReader.ReadAll()
	if err != nil {
		fmt.Println(err)
		return as, err
	}

	cMap := make(map[string]int)

	for i := range cArray {
		cMap[cArray[i][0]] = i
	}

	for _, record := range records {
		var a App

		record = strings.Replace(record, `"n/a"`, `0`, -1)
		record = strings.Replace(record, `kochava_url":0`, `kochava_url":""`, -1)
		record = strings.Replace(record, `\x1f`, ``, -1)
		//		fmt.Println(record)
		err := json.Unmarshal([]byte(record), &a)
		if err != nil {
			return as, err
		}

		sort := regexp.MustCompile(`[\s;]?([^;]*);?`)

		sdks := sort.FindAllStringSubmatch(a.Sdks, -1)

		sdksArray := make([]string, 0)

		for _, sdk := range sdks {
			sdksArray = append(sdksArray, sdk[1])
		}

		a.SDKsParsed = sdksArray

		as = append(as, a)

	}

	for i, app := range as {
		sdkCatList := make([]string, len(cArray))
		for _, id := range app.SDKsParsed {
			sdkCatList[cMap[sdkMap[id]]] = sdkCatList[cMap[sdkMap[id]]] + fmt.Sprintf("%v;", id)
		}
		as[i].SdkCatList = sdkCatList
	}

	return as, nil
}

func ParseSDKDataFromArray(records []string) ([]SDK, error) {

	sdks := make([]SDK, 0)

	for _, record := range records {
		var s SDK

		record = strings.Replace(record, `\x1f`, ``, -1)

		err := json.Unmarshal([]byte(record), &s)
		if err != nil {
			return sdks, err
		}

		sdks = append(sdks, s)
	}

	return sdks, nil
}
