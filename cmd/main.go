package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/csv"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	apptopia "github.com/tenglun/apptopia_transform"
)

var username *string

func init() {
	username = flag.String("username", "", "flag to define current user profile name")
}

func main() {

	// Get Environmental Variables
	defaultUsername := os.Getenv("APP_S3_USERNAME")
	bucket := os.Getenv("APP_S3_BUCKET")

	if *username == "" {
		*username = defaultUsername
	}

	// Initialize input parameters
	flag.Parse()

	errLog, err := os.OpenFile(fmt.Sprintf("%v_log.txt", time.Now().Unix()), os.O_CREATE, 666)
	if err != nil {

		return
	}

	log.SetOutput(errLog)

	defer errLog.Close()

	log.SetFlags(log.Lshortfile)

	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered in main().32", r)
		}
	}()

	// Sync apptopia bucket to local folder to ensure up-to-date data
	cmd := "aws"
	cmdArgs := []string{"s3", "sync", bucket, fmt.Sprintf("C:/Users/%v/documents/apptopia-dw-kochava", *username)}
	var cmdOutput []byte

	if cmdOutput, err = exec.Command(cmd, cmdArgs...).Output(); err != nil {
		log.Println(err)
		os.Exit(1)
	}

	log.Println(string(cmdOutput))

	// Check if input folder exists
	if _, err := os.Stat(fmt.Sprintf("c:/Users/%v/documents/apptopia-dw-kochava/", *username)); os.IsNotExist(err) {
		log.Println("error; input folder does not exist")
		return
	}

	var files []string

	filepath.Walk(fmt.Sprintf("c:/Users/%v/documents/apptopia-dw-kochava/", *username), func(path string, info os.FileInfo, err error) error {
		files = append(files, path)
		return nil
	})

	for _, f := range files {
		switch {
		case strings.Contains(f, "sdks"):
			err := parseFile(f, "sdks", "")
			if err != nil {
				log.Println(err)
				continue
			}
		default:
			log.Println(fmt.Errorf("running sdk files first: %v", f))
			log.Println()
		}
	}

	for _, f := range files {
		switch {
		case strings.Contains(f, "apps_data"):
			switch {
			case strings.Contains(f, "google"):
				err := parseFile(f, "app", "google")
				if err != nil {
					log.Println(err)
					continue
				}
			case strings.Contains(f, `itunes`):
				err := parseFile(f, "app", "itunes")
				if err != nil {
					log.Println(err)
					continue
				}
			}

		case strings.Contains(f, "publisher_data"):
			err := parseFile(f, "publisher", "")
			if err != nil {
				log.Println(err)
				continue
			}
		default:
			log.Println(fmt.Errorf("%v is not an applicable file, or is an sdk file and has already been processed; skipping", f))
			log.Println()
		}
	}

}

func parseFile(path, inputType string, store string) error {

	logger := log.Logger{}

	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered in parseFile()", r)
		}
	}()

	var pFile *gzip.Reader

	log.Println(path)

	if filepath.Ext(path) == ".gz" || filepath.Ext(path) == ".gzip" {
		gZip, err := os.Open(path)
		if err != nil {
			return err
		}

		defer gZip.Close()

		pFile, err = gzip.NewReader(gZip)
		if err != nil {
			return err
		}
	} else {
		var buf bytes.Buffer
		var err error
		file, err := os.Open(path)
		if err != nil {

			return err
		}
		gzipWriter := gzip.NewWriter(&buf)

		data, err := ioutil.ReadAll(file)
		if err != nil {

			return err
		}

		_, err = gzipWriter.Write(data)
		if err != nil {

			return err
		}

		gzipWriter.Close()

		pFile, err = gzip.NewReader(&buf)
		if err != nil {

			return err
		}

	}

	var lines []string

	var scanner *bufio.Scanner

	scanner = bufio.NewScanner(pFile)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	parser := regexp.MustCompile(`\\([^\.\\]*?)\.`)
	parserFolder := regexp.MustCompile(`\\([^\\]*)\\[^\\]*\.`)

	fileName := parser.FindStringSubmatch(path)
	folderName := parserFolder.FindStringSubmatch(path)

	if len(folderName) < 2 {
		folderName = []string{"", "not_found"}
	}

	output, err := os.OpenFile(fmt.Sprintf("c:/Users/%v/documents/s3_output/%v_%v_%v_output.csv", *username, folderName[1], fileName[1], inputType), os.O_CREATE, 666)
	if err != nil {

		return err
	}

	defer output.Close()

	outputCSV := csv.NewWriter(output)

	switch inputType {
	case "app":

		ps, err := apptopia.ParseAppDataFromArray(lines, *username, store, &logger)
		if err != nil {

			return err
		}

		err = outPutApp(outputCSV, ps)
		if err != nil {

			return err
		}

	case "publisher":
		ps, err := apptopia.ParsePublisherDataFromArray(lines)
		if err != nil {

			return err
		}

		err = outPutPub(outputCSV, ps)
		if err != nil {

			return err
		}

	case "sdks":
		sdks, err := apptopia.ParseSDKDataFromArray(lines)
		if err != nil {

			return err
		}

		err = outPutSDK(outputCSV, sdks)
		if err != nil {

			return err
		}
	default:
	}

	return nil
}

func getSDKCategories() []string {

	c := make([]string, 0)

	file, err := os.Open("data/categories.csv")
	if err != nil {
		fmt.Println(err)
		return c
	}

	defer file.Close()

	csvReader := csv.NewReader(file)

	cArray, err := csvReader.ReadAll()
	if err != nil {
		fmt.Println(err)
		return c
	}

	for i := range cArray {
		if i == 0 {
			continue
		}
		c = append(c, cArray[i][0])
	}

	return c
}

func outPutApp(outputCSV *csv.Writer, ps []apptopia.App) error {

	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered in outPutApp()", r)
		}
	}()

	outPutHeader := []string{"sdks", "session_length", "app_store_publisher_url", "offers_iap", "release_date", "mau", "price_usd", "avg_rating", "name_usa", "app_id", "description_usa", "app_store_app_url", "category_name", "dau", "revenue_ads", "downloads", "current_version", "revenue_downloads", "paid", "publisher_name", "sessions", "total_ratings", "revenue_iaps", "age_restrictions", "vnd_publisher_id", "kochava_id", "kochava_url", "kochava_name", "account_id", "SDK_ids"}

	categories := getSDKCategories()

	outPutHeader = append(outPutHeader, categories...)

	err := outputCSV.Write(outPutHeader)

	if err != nil {

		return err
	}

	for _, app := range ps {

		outputArray := []string{app.Sdks,
			strconv.FormatFloat(app.SessionLen, 'f', 5, 64),
			//app.SessionLen,
			app.AppstorePublisherURL,
			strconv.FormatBool(app.OffersInAppPurchases),
			app.ReleaseDate,
			strconv.FormatFloat(app.Mau, 'f', 5, 64),
			strconv.Itoa(app.PriceUsUsd),
			strconv.Itoa(app.AvgRating),
			app.NameUs,
			string(app.AppID),
			app.DescriptionUs,
			app.AppstoreAppURL,
			app.CategoryName,
			strconv.FormatFloat(app.Dau, 'f', 5, 64),
			strconv.Itoa(app.RevAds),
			strconv.Itoa(app.Dls),
			app.CurrentVersion,
			strconv.Itoa(app.RevDls),
			strconv.FormatBool(app.Paid),
			app.PublisherName,
			strconv.Itoa(app.Sessions),
			strconv.Itoa(app.TotalRatings),
			strconv.Itoa(app.RevIaps),
			app.AgeRestrictions,
			strconv.Itoa(app.VndPublisherID),
			app.KochavaID,
			app.KochavaURL,
			app.KochavaName,
			app.AccountID,
			strings.Trim(strings.Join(strings.Fields(fmt.Sprint(app.SdkIds)), ";"), "[]")}

		outputArray = append(outputArray, app.SdkCatList...)

		err = outputCSV.Write(outputArray)

		if err != nil {

			return err
		}
	}

	outputCSV.Flush()

	return nil
}

func outPutPub(outputCSV *csv.Writer, ps []apptopia.Publisher) error {

	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered in outPutPub()", r)
		}
	}()

	err := outputCSV.Write([]string{"publisher_id",
		"ad_revenue",
		"iaps_revenue",
		"total_revenue",
		"dau",
		"mau",
		"downloads",
		"publisher_name",
		"headquarters_country",
		"website_url",
		"kochava_id",
		"kochava_url",
		"kochava_name",
	})

	if err != nil {

		return err
	}

	for _, publisher := range ps {
		err = outputCSV.Write([]string{
			publisher.PublisherID,
			publisher.AdRev,
			publisher.IapsRev,
			publisher.TotalRev,
			publisher.Dau,
			publisher.Mau,
			publisher.Dls,
			publisher.PublisherName,
			publisher.HqCountry,
			publisher.WebsiteURL,
			publisher.KochavaID,
			publisher.KochavaURL,
			publisher.KochavaName,
		})
		if err != nil {

			return err
		}
	}

	outputCSV.Flush()

	return nil
}

func outPutSDK(outputCSV *csv.Writer, sdks []apptopia.SDK) error {

	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered in outPutSDK()", r)
		}
	}()

	err := outputCSV.Write([]string{
		"category",
		"sdk_id",
		"name",
	})

	if err != nil {

		return err
	}

	for _, sdk := range sdks {
		err := outputCSV.Write([]string{
			sdk.Category,
			fmt.Sprintf("%v", sdk.ID),
			sdk.Name,
		})
		if err != nil {

			return err
		}
	}

	outputCSV.Flush()

	return nil
}
