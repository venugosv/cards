package gpay

import (
	"context"

	"github.com/anzx/pkg/log"
)

//nolint:gocyclo
func convertCountry(ctx context.Context, in string) string {
	switch in {
	case "AFG": // Afghanistan
		return "AF"
	case "ALA": // Åland Islands
		return "AX"
	case "ALB": // Albania
		return "AL"
	case "DZA": // Algeria
		return "DZ"
	case "ASM": // American Samoa
		return "AS"
	case "AND": // Andorra
		return "AD"
	case "AGO": // Angola
		return "AO"
	case "AIA": // Anguilla
		return "AI"
	case "ATA": // Antarctica
		return "AQ"
	case "ATG": // Antigua and Barbuda
		return "AG"
	case "ARG": // Argentina
		return "AR"
	case "ARM": // Armenia
		return "AM"
	case "ABW": // Aruba
		return "AW"
	case "AUS": // Australia
		return "AU"
	case "AUT": // Austria
		return "AT"
	case "AZE": // Azerbaijan
		return "AZ"
	case "BHS": // Bahamas
		return "BS"
	case "BHR": // Bahrain
		return "BH"
	case "BGD": // Bangladesh
		return "BD"
	case "BRB": // Barbados
		return "BB"
	case "BLR": // Belarus
		return "BY"
	case "BEL": // Belgium
		return "BE"
	case "BLZ": // Belize
		return "BZ"
	case "BEN": // Benin
		return "BJ"
	case "BMU": // Bermuda
		return "BM"
	case "BTN": // Bhutan
		return "BT"
	case "BOL": // Bolivia (Plurinational State of)
		return "BO"
	case "BES": // "Bonaire, Sint Eustatius and Saba"
		return "BQ"
	case "BIH": // Bosnia and Herzegovina
		return "BA"
	case "BWA": // Botswana
		return "BW"
	case "BVT": // Bouvet Island
		return "BV"
	case "BRA": // Brazil
		return "BR"
	case "IOT": // British Indian Ocean Territory
		return "IO"
	case "BRN": // Brunei Darussalam
		return "BN"
	case "BGR": // Bulgaria
		return "BG"
	case "BFA": // Burkina Faso
		return "BF"
	case "BDI": // Burundi
		return "BI"
	case "CPV": // Cabo Verde
		return "CV"
	case "KHM": // Cambodia
		return "KH"
	case "CMR": // Cameroon
		return "CM"
	case "CAN": // Canada
		return "CA"
	case "CYM": // Cayman Islands
		return "KY"
	case "CAF": // Central African Republic
		return "CF"
	case "TCD": // Chad
		return "TD"
	case "CHL": // Chile
		return "CL"
	case "CHN": // China
		return "CN"
	case "CXR": // Christmas Island
		return "CX"
	case "CCK": // Cocos (Keeling) Islands
		return "CC"
	case "COL": // Colombia
		return "CO"
	case "COM": // Comoros
		return "KM"
	case "COG": // Congo
		return "CG"
	case "COD": // "Congo, Democratic Republic of the"
		return "CD"
	case "COK": // Cook Islands
		return "CK"
	case "CRI": // Costa Rica
		return "CR"
	case "CIV": // Côte d'Ivoire
		return "CI"
	case "HRV": // Croatia
		return "HR"
	case "CUB": // Cuba
		return "CU"
	case "CUW": // Curaçao
		return "CW"
	case "CYP": // Cyprus
		return "CY"
	case "CZE": // Czechia
		return "CZ"
	case "DNK": // Denmark
		return "DK"
	case "DJI": // Djibouti
		return "DJ"
	case "DMA": // Dominica
		return "DM"
	case "DOM": // Dominican Republic
		return "DO"
	case "ECU": // Ecuador
		return "EC"
	case "EGY": // Egypt
		return "EG"
	case "SLV": // El Salvador
		return "SV"
	case "GNQ": // Equatorial Guinea
		return "GQ"
	case "ERI": // Eritrea
		return "ER"
	case "EST": // Estonia
		return "EE"
	case "SWZ": // Eswatini
		return "SZ"
	case "ETH": // Ethiopia
		return "ET"
	case "FLK": // Falkland Islands (Malvinas)
		return "FK"
	case "FRO": // Faroe Islands
		return "FO"
	case "FJI": // Fiji
		return "FJ"
	case "FIN": // Finland
		return "FI"
	case "FRA": // France
		return "FR"
	case "GUF": // French Guiana
		return "GF"
	case "PYF": // French Polynesia
		return "PF"
	case "ATF": // French Southern Territories
		return "TF"
	case "GAB": // Gabon
		return "GA"
	case "GMB": // Gambia
		return "GM"
	case "GEO": // Georgia
		return "GE"
	case "DEU": // Germany
		return "DE"
	case "GHA": // Ghana
		return "GH"
	case "GIB": // Gibraltar
		return "GI"
	case "GRC": // Greece
		return "GR"
	case "GRL": // Greenland
		return "GL"
	case "GRD": // Grenada
		return "GD"
	case "GLP": // Guadeloupe
		return "GP"
	case "GUM": // Guam
		return "GU"
	case "GTM": // Guatemala
		return "GT"
	case "GGY": // Guernsey
		return "GG"
	case "GIN": // Guinea
		return "GN"
	case "GNB": // Guinea-Bissau
		return "GW"
	case "GUY": // Guyana
		return "GY"
	case "HTI": // Haiti
		return "HT"
	case "HMD": // Heard Island and McDonald Islands
		return "HM"
	case "VAT": // Holy See
		return "VA"
	case "HND": // Honduras
		return "HN"
	case "HKG": // Hong Kong
		return "HK"
	case "HUN": // Hungary
		return "HU"
	case "ISL": // Iceland
		return "IS"
	case "IND": // India
		return "IN"
	case "IDN": // Indonesia
		return "ID"
	case "IRN": // Iran (Islamic Republic of)
		return "IR"
	case "IRQ": // Iraq
		return "IQ"
	case "IRL": // Ireland
		return "IE"
	case "IMN": // Isle of Man
		return "IM"
	case "ISR": // Israel
		return "IL"
	case "ITA": // Italy
		return "IT"
	case "JAM": // Jamaica
		return "JM"
	case "JPN": // Japan
		return "JP"
	case "JEY": // Jersey
		return "JE"
	case "JOR": // Jordan
		return "JO"
	case "KAZ": // Kazakhstan
		return "KZ"
	case "KEN": // Kenya
		return "KE"
	case "KIR": // Kiribati
		return "KI"
	case "PRK": // Korea (Democratic People's Republic of)
		return "KP"
	case "KOR": // "Korea, Republic of"
		return "KR"
	case "KWT": // Kuwait
		return "KW"
	case "KGZ": // Kyrgyzstan
		return "KG"
	case "LAO": // Lao People's Democratic Republic
		return "LA"
	case "LVA": // Latvia
		return "LV"
	case "LBN": // Lebanon
		return "LB"
	case "LSO": // Lesotho
		return "LS"
	case "LBR": // Liberia
		return "LR"
	case "LBY": // Libya
		return "LY"
	case "LIE": // Liechtenstein
		return "LI"
	case "LTU": // Lithuania
		return "LT"
	case "LUX": // Luxembourg
		return "LU"
	case "MAC": // Macao
		return "MO"
	case "MDG": // Madagascar
		return "MG"
	case "MWI": // Malawi
		return "MW"
	case "MYS": // Malaysia
		return "MY"
	case "MDV": // Maldives
		return "MV"
	case "MLI": // Mali
		return "ML"
	case "MLT": // Malta
		return "MT"
	case "MHL": // Marshall Islands
		return "MH"
	case "MTQ": // Martinique
		return "MQ"
	case "MRT": // Mauritania
		return "MR"
	case "MUS": // Mauritius
		return "MU"
	case "MYT": // Mayotte
		return "YT"
	case "MEX": // Mexico
		return "MX"
	case "FSM": // Micronesia (Federated States of)
		return "FM"
	case "MDA": // "Moldova, Republic of"
		return "MD"
	case "MCO": // Monaco
		return "MC"
	case "MNG": // Mongolia
		return "MN"
	case "MNE": // Montenegro
		return "ME"
	case "MSR": // Montserrat
		return "MS"
	case "MAR": // Morocco
		return "MA"
	case "MOZ": // Mozambique
		return "MZ"
	case "MMR": // Myanmar
		return "MM"
	case "NAM": // Namibia
		return "NA"
	case "NRU": // Nauru
		return "NR"
	case "NPL": // Nepal
		return "NP"
	case "NLD": // Netherlands
		return "NL"
	case "NCL": // New Caledonia
		return "NC"
	case "NZL": // New Zealand
		return "NZ"
	case "NIC": // Nicaragua
		return "NI"
	case "NER": // Niger
		return "NE"
	case "NGA": // Nigeria
		return "NG"
	case "NIU": // Niue
		return "NU"
	case "NFK": // Norfolk Island
		return "NF"
	case "MKD": // North Macedonia
		return "MK"
	case "MNP": // Northern Mariana Islands
		return "MP"
	case "NOR": // Norway
		return "NO"
	case "OMN": // Oman
		return "OM"
	case "PAK": // Pakistan
		return "PK"
	case "PLW": // Palau
		return "PW"
	case "PSE": // "Palestine, State of"
		return "PS"
	case "PAN": // Panama
		return "PA"
	case "PNG": // Papua New Guinea
		return "PG"
	case "PRY": // Paraguay
		return "PY"
	case "PER": // Peru
		return "PE"
	case "PHL": // Philippines
		return "PH"
	case "PCN": // Pitcairn
		return "PN"
	case "POL": // Poland
		return "PL"
	case "PRT": // Portugal
		return "PT"
	case "PRI": // Puerto Rico
		return "PR"
	case "QAT": // Qatar
		return "QA"
	case "REU": // Réunion
		return "RE"
	case "ROU": // Romania
		return "RO"
	case "RUS": // Russian Federation
		return "RU"
	case "RWA": // Rwanda
		return "RW"
	case "BLM": // Saint Barthélemy
		return "BL"
	case "SHN": // "Saint Helena, Ascension and Tristan da Cunha"
		return "SH"
	case "KNA": // Saint Kitts and Nevis
		return "KN"
	case "LCA": // Saint Lucia
		return "LC"
	case "MAF": // Saint Martin (French part)
		return "MF"
	case "SPM": // Saint Pierre and Miquelon
		return "PM"
	case "VCT": // Saint Vincent and the Grenadines
		return "VC"
	case "WSM": // Samoa
		return "WS"
	case "SMR": // San Marino
		return "SM"
	case "STP": // Sao Tome and Principe
		return "ST"
	case "SAU": // Saudi Arabia
		return "SA"
	case "SEN": // Senegal
		return "SN"
	case "SRB": // Serbia
		return "RS"
	case "SYC": // Seychelles
		return "SC"
	case "SLE": // Sierra Leone
		return "SL"
	case "SGP": // Singapore
		return "SG"
	case "SXM": // Sint Maarten (Dutch part)
		return "SX"
	case "SVK": // Slovakia
		return "SK"
	case "SVN": // Slovenia
		return "SI"
	case "SLB": // Solomon Islands
		return "SB"
	case "SOM": // Somalia
		return "SO"
	case "ZAF": // South Africa
		return "ZA"
	case "SGS": // South Georgia and the South Sandwich Islands
		return "GS"
	case "SSD": // South Sudan
		return "SS"
	case "ESP": // Spain
		return "ES"
	case "LKA": // Sri Lanka
		return "LK"
	case "SDN": // Sudan
		return "SD"
	case "SUR": // Suriname
		return "SR"
	case "SJM": // Svalbard and Jan Mayen
		return "SJ"
	case "SWE": // Sweden
		return "SE"
	case "CHE": // Switzerland
		return "CH"
	case "SYR": // Syrian Arab Republic
		return "SY"
	case "TWN": // "Taiwan, Province of China"
		return "TW"
	case "TJK": // Tajikistan
		return "TJ"
	case "TZA": // "Tanzania, United Republic of"
		return "TZ"
	case "THA": // Thailand
		return "TH"
	case "TLS": // Timor-Leste
		return "TL"
	case "TGO": // Togo
		return "TG"
	case "TKL": // Tokelau
		return "TK"
	case "TON": // Tonga
		return "TO"
	case "TTO": // Trinidad and Tobago
		return "TT"
	case "TUN": // Tunisia
		return "TN"
	case "TUR": // Turkey
		return "TR"
	case "TKM": // Turkmenistan
		return "TM"
	case "TCA": // Turks and Caicos Islands
		return "TC"
	case "TUV": // Tuvalu
		return "TV"
	case "UGA": // Uganda
		return "UG"
	case "UKR": // Ukraine
		return "UA"
	case "ARE": // United Arab Emirates
		return "AE"
	case "GBR": // United Kingdom of Great Britain and Northern Ireland
		return "GB"
	case "USA": // United States of America
		return "US"
	case "UMI": // United States Minor Outlying Islands
		return "UM"
	case "URY": // Uruguay
		return "UY"
	case "UZB": // Uzbekistan
		return "UZ"
	case "VUT": // Vanuatu
		return "VU"
	case "VEN": // Venezuela (Bolivarian Republic of)
		return "VE"
	case "VNM": // Viet Nam
		return "VN"
	case "VGB": // Virgin Islands (British)
		return "VG"
	case "VIR": // Virgin Islands (U.S.)
		return "VI"
	case "WLF": // Wallis and Futuna
		return "WF"
	case "ESH": // Western Sahara
		return "EH"
	case "YEM": // Yemen
		return "YE"
	case "ZMB": // Zambia
		return "ZM"
	case "ZWE": // Zimbabwe
		return "ZW"
	default:
		log.Debug(ctx, "unable to convert ISO 3166-1 alpha-3 to alpha-2", log.Str("CountryCode", in))
		return in
	}
}
