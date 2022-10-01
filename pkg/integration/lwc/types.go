package lwc

import "time"

type Request struct {
	BankTransactions []BankTransactions `json:"bank_transactions"`
}

type BankTransactions struct {
	Cal string `json:"cal"`
}

type Response struct {
	SearchResults []SearchResults `json:"search_results"`
}

type AddressLocatedIn struct {
	LocatedInName        string `json:"located_in_name"`
	LocatedInAddress     string `json:"located_in_address"`
	LocatedInWebsite     string `json:"located_in_website"`
	LocatedInPhoneNumber string `json:"located_in_phone_number"`
}

type AddressIdentifiers struct {
	AddressIdentifierName  string `json:"address_identifier_name"`
	AddressIdentifierValue string `json:"address_identifier_value"`
}

type Address struct {
	SingleLineAddress   string             `json:"single_line_address"`
	AddressLine1        string             `json:"address_line_1"`
	AddressLine2        string             `json:"address_line_2"`
	State               string             `json:"state"`
	Postcode            string             `json:"postcode"`
	Suburb              string             `json:"suburb"`
	CountryName         string             `json:"country_name"`
	Longitude           int                `json:"longitude"`
	Latitude            int                `json:"latitude"`
	LatLonPrecision     int                `json:"lat_lon_precision"`
	Mapable             bool               `json:"mapable"`
	StreetViewAvailable bool               `json:"street_view_available"`
	AddressLocatedIn    AddressLocatedIn   `json:"address_located_in"`
	AddressIdentifiers  AddressIdentifiers `json:"address_identifiers"`
}

type Verification struct {
	IsVerified     bool      `json:"is_verified"`
	LastVerifiedOn time.Time `json:"last_verified_on"`
}

type Contact struct {
	TypeOfContact string `json:"type_of_contact"`
	Value         string `json:"value"`
	DisplayValue  string `json:"display_value"`
	Label         string `json:"label"`
}

type ImageGallery struct {
	ThumbnailURL string `json:"thumbnail_url"`
	LargeURL     string `json:"large_url"`
	ImageTitle   string `json:"image_title"`
	ImageHeight  int    `json:"image_height"`
	ImageWidth   int    `json:"image_width"`
}

type Parent struct {
	CategoryName     string `json:"category_name"`
	ID               int    `json:"id"`
	FullCategoryPath string `json:"full_category_path"`
	Parent           string `json:"parent"`
}

type CategoryIcon struct {
	BlackWhiteURL string `json:"Black_white_url"`
	ColouredURL   string `json:"Coloured_url"`
	Height        int    `json:"height"`
	Width         int    `json:"width"`
}

type RecategorisationInfo struct {
	RuleID       int       `json:"rule_id"`
	RuleDateTime time.Time `json:"rule_date_time"`
}

type Category struct {
	CategoryName            string               `json:"category_name"`
	ID                      int                  `json:"id"`
	FullCategoryPath        string               `json:"full_category_path"`
	Parent                  Parent               `json:"parent"`
	IsSensitive             bool                 `json:"is_sensitive"`
	IsLwcCategory           bool                 `json:"is_lwc_category"`
	IsSubstitutedCategory   bool                 `json:"is_substituted_category"`
	LwcCategoryIcon         CategoryIcon         `json:"lwc_category_icon"`
	LwcCategoryIconCircular CategoryIcon         `json:"lwc_category_icon_circular"`
	CategoryEmoji           string               `json:"category_emoji"`
	RecategorisationInfo    RecategorisationInfo `json:"recategorisation_info"`
	CategoryID              string               `json:"category_id"`
}

type Image struct {
	URL    string `json:"url"`
	Height int    `json:"height"`
	Width  int    `json:"width"`
}

type WebsiteScreenShot struct {
	Thumbnail Image `json:"thumbnail"`
	Banner    Image `json:"banner"`
	Main      Image `json:"main"`
}

type AssociatedWith struct {
	AssociatesName    string `json:"associates_name"`
	AssociatesID      string `json:"associates_id"`
	TypeOfAssociation string `json:"type_of_association"`
}

type Time struct {
	Open  string `json:"open"`
	Close string `json:"close"`
}

type Day struct {
	Status string `json:"status"`
	Times  []Time `json:"times"`
}

type OpeningHours struct {
	IsAlwaysOpen bool `json:"is_always_open"`
	Sunday       Day  `json:"sunday"`
	Monday       Day  `json:"monday"`
	Tuesday      Day  `json:"tuesday"`
	Wednesday    Day  `json:"wednesday"`
	Thursday     Day  `json:"thursday"`
	Friday       Day  `json:"friday"`
	Saturday     Day  `json:"saturday"`
}

type LegalRegistration struct {
	LegalNumber      string `json:"legal_number"`
	LegalNumberLabel string `json:"legal_number_label"`
}

type LegalBusinessInfo struct {
	DateEstablished       time.Time           `json:"date_established"`
	EntityType            string              `json:"entity_type"`
	CurrentMerchantStatus string              `json:"current_merchant_status"`
	MerchantType          string              `json:"merchant_type"`
	MerchantPresence      string              `json:"merchant_presence"`
	ChainName             string              `json:"chain_name"`
	LegalRegistrations    []LegalRegistration `json:"legal_registrations"`
	RegisteredForSalesTax bool                `json:"registered_for_sales_tax"`
	ChainLwcGUID          string              `json:"chain_lwc_guid"`
}

type OverallRating struct {
	OverallRatingScore   int `json:"overall_rating_score"`
	TotalNumberOfReviews int `json:"total_number_of_reviews"`
}

type Rating struct {
	Reviewer        string `json:"reviewer"`
	Score           int    `json:"score"`
	NumberOfReviews int    `json:"number_of_reviews"`
}

type TransactionQueryProcess struct {
	TransactionQueryProcessMarkdown string `json:"transaction_query_process_markdown"`
	TransactionQueryProcessLink     string `json:"transaction_query_process_link"`
}

type Items struct {
	Value        string `json:"value"`
	Type         string `json:"type"`
	DisplayValue string `json:"display_value"`
}

type Paragraphs struct {
	Items []Items `json:"items"`
}

type TransactionQueryProcessJSON struct {
	Paragraphs []Paragraphs `json:"paragraphs"`
}

type AdditionalDetails struct {
	BpayBillerCodes  []int `json:"bpay_biller_codes"`
	DirectEntryCodes []int `json:"direct_entry_codes"`
}

type MerchantCategoryMapping struct {
	TypeOfCategory     string `json:"type_of_category"`
	CategoryIdentifier string `json:"category_identifier"`
	CategoryName       string `json:"category_name"`
	MappingError       string `json:"mapping_error"`
}

type MerchantTag struct {
	ID    int    `json:"id"`
	Label string `json:"label"`
}

type MerchantDetails struct {
	ID                          int                         `json:"id"`
	LwcGUID                     string                      `json:"lwc_guid"`
	DataEnrichmentScore         int                         `json:"data_enrichment_score"`
	MerchantPrimaryName         string                      `json:"merchant_primary_name"`
	PrimaryAddress              Address                     `json:"primary_address"`
	Verification                Verification                `json:"verification"`
	PrimaryContacts             []Contact                   `json:"primary_contacts"`
	SuspiciousMerchantScore     int                         `json:"suspicious_merchant_score"`
	SecondaryContacts           []Contact                   `json:"secondary_contacts"`
	SecondaryAddresses          []Address                   `json:"secondary_addresses"`
	ImageGallery                []ImageGallery              `json:"image_gallery"`
	PrimaryCategory             Category                    `json:"primary_category"`
	OtherCategoriesList         []Category                  `json:"other_categories_list"`
	PrimaryMerchantDescription  string                      `json:"primary_merchant_description"`
	MerchantDescriptors         []string                    `json:"merchant_descriptors"`
	IsPermanentlyClosed         bool                        `json:"is_permanently_closed"`
	MerchantLogo                Image                       `json:"merchant_logo"`
	MerchantLogo3X1             Image                       `json:"merchant_logo_3x1"`
	MerchantLogoCircular        Image                       `json:"merchant_logo_circular"`
	MerchantLogoDark            Image                       `json:"merchant_logo_dark"`
	MerchantLogo3X1Dark         Image                       `json:"merchant_logo_3x1_dark"`
	MerchantLogoCircularDark    Image                       `json:"merchant_logo_circular_dark"`
	WebsiteScreenShot           WebsiteScreenShot           `json:"website_screen_shot"`
	AlsoKnownAs                 []string                    `json:"also_known_as"`
	AssociatedWith              []AssociatedWith            `json:"associated_with"`
	LastUpdated                 time.Time                   `json:"last_updated"`
	OpeningHours                OpeningHours                `json:"opening_hours"`
	LegalBusinessInfo           LegalBusinessInfo           `json:"legal_business_info"`
	OverallRating               OverallRating               `json:"overall_rating"`
	Ratings                     []Rating                    `json:"ratings"`
	PaymentOptions              []string                    `json:"payment_options"`
	TransactionQueryProcess     TransactionQueryProcess     `json:"transaction_query_process"`
	TransactionQueryProcessHTML string                      `json:"transaction_query_process_html"`
	TransactionQueryProcessJSON TransactionQueryProcessJSON `json:"transaction_query_process_json"`
	ReceiptDataAvailable        bool                        `json:"receipt_data_available"`
	RecordIsQuarantined         bool                        `json:"record_is_quarantined"`
	AdditionalDetails           AdditionalDetails           `json:"additional_details"`
	MerchantCategoryMappings    []MerchantCategoryMapping   `json:"merchant_category_mappings"`
	MerchantTags                []MerchantTag               `json:"merchant_tags"`
	BsbNumbers                  []string                    `json:"bsb_numbers"`
	IsPaymentaggregator         bool                        `json:"is_paymentaggregator"`
}

type MerchantSearchResult struct {
	SearchResultsGUID      string          `json:"search_results_guid"`
	LwcID                  int             `json:"LWC_ID"`
	LwcGUID                string          `json:"lwc_guid"`
	Rank                   int             `json:"rank"`
	Score                  int             `json:"score"`
	MatchFeedbackRequested bool            `json:"match_feedback_requested"`
	TypeOfMatch            string          `json:"type_of_match"`
	MerchantDetails        MerchantDetails `json:"merchant_details"`
}

type SearchResults struct {
	MerchantSearchResults []MerchantSearchResult `json:"merchant_search_results"`
}
