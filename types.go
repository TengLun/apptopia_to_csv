package apptopiaTransform

import (
	"encoding/json"
)

type Publisher struct {
	PublisherID   string `json:"publisher_id"`
	AdRev         string `json:"ad_rev"`
	IapsRev       string `json:"iaps_rev"`
	TotalRev      string `json:"total_rev"`
	Dau           string `json:"dau"`
	Mau           string `json:"mau"`
	Dls           string `json:"dls"`
	PublisherName string `json:"publisher_name"`
	HqCountry     string `json:"hq_country"`
	WebsiteURL    string `json:"website_url"`
	AccountID     string `json:"account_id"`
	KochavaName   string `json:"kochava_name"`
	KochavaURL    string `json:"kochava_url"`
	KochavaID     string `json:"kochava_id"`

	// Parsed or Added data
	Domain string // Of format domain.com, domain.io, etc.
	Apps   []App
}

type RatingsBreakdown struct {
	Num1 int `json:"1"`
	Num2 int `json:"2"`
	Num3 int `json:"3"`
	Num4 int `json:"4"`
	Num5 int `json:"5"`
}

type App struct {
	Sdks                 string           `json:"sdks"`
	SessionLen           float64          `json:"session_len"`
	AppstorePublisherURL string           `json:"appstore_publisher_url"`
	OffersInAppPurchases bool             `json:"offers_in_app_purchases"`
	ReleaseDate          string           `json:"release_date"`
	Mau                  float64          `json:"mau"`
	PriceUsUsd           int              `json:"price_us_usd"`
	AvgRating            int              `json:"avg_rating"`
	NameUs               string           `json:"name_us"`
	AppID                json.RawMessage  `json:"app_id,string"`
	RatingsBreakdown     RatingsBreakdown `json:"ratings_breakdown"`
	DescriptionUs        string           `json:"description_us"`
	AppstoreAppURL       string           `json:"appstore_app_url"`
	CategoryName         string           `json:"category_name"`
	Dau                  float64          `json:"dau"`
	RevAds               int              `json:"rev_ads"`
	Dls                  int              `json:"dls"`
	CurrentVersion       string           `json:"current_version"`
	LastVersionUpdateOn  interface{}      `json:"last_version_update_on"`
	RevDls               int              `json:"rev_dls"`
	Paid                 bool             `json:"paid"`
	PublisherName        string           `json:"publisher_name"`
	Sessions             int              `json:"sessions"`
	TotalRatings         int              `json:"total_ratings"`
	RevIaps              int              `json:"rev_iaps"`
	AgeRestrictions      string           `json:"age_restrictions"`
	VndPublisherID       int              `json:"vnd_publisher_id"`
	AccountID            string           `json:"account_id"`
	KochavaName          string           `json:"kochava_name"`
	KochavaURL           string           `json:"kochava_url"`
	KochavaID            string           `json:"kochava_id"`
	SdkIds               []int            `json:"sdk_ids"`

	// Parsed or Added data
	SDKsParsed []string
	SdkCatList []string
}

type SDK struct {
	Category string `json:"category"`
	ID       int    `json:"id"`
	Name     string `json:"name"`
}
