package lwc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anzx/fabric-cards/pkg/util/testutil"

	"github.com/stretchr/testify/assert"
)

func getResponse() *Response {
	bytes := []byte("{\"search_telemetry\":\"string\",\"transactions_count\":0,\"search_time_ms\":0,\"total_credits_used\":0,\"fields\":\"string\",\"search_results\":[{\"search_guid\":\"3fa85f64-5717-4562-b3fc-2c963f66afa6\",\"cal\":\"string\",\"fixedcal\":\"string\",\"aid\":\"string\",\"tid\":\"string\",\"mid\":\"string\",\"bank_account_transaction_type\":\"string\",\"direct_entry_code\":\"string\",\"bpay_biller_code\":\"string\",\"terminal_country_code\":\"string\",\"customer_correlation_id\":\"string\",\"transaction_correlation_guid\":\"3fa85f64-5717-4562-b3fc-2c963f66afa6\",\"bank_account_type\":\"string\",\"transaction_bank_3alpha\":\"string\",\"number_of_results\":0,\"highest_score\":0,\"utc_expiry_time\":\"2022-02-22T23:00:14.098Z\",\"merchant_search_results\":[{\"search_results_guid\":\"3fa85f64-5717-4562-b3fc-2c963f66afa6\",\"LWC_ID\":0,\"lwc_guid\":\"3fa85f64-5717-4562-b3fc-2c963f66afa6\",\"rank\":0,\"score\":0,\"match_feedback_requested\":true,\"type_of_match\":\"string\",\"merchant_details\":{\"id\":0,\"lwc_guid\":\"3fa85f64-5717-4562-b3fc-2c963f66afa6\",\"data_enrichment_score\":0,\"merchant_primary_name\":\"string\",\"primary_address\":{\"single_line_address\":\"string\",\"address_line_1\":\"string\",\"address_line_2\":\"string\",\"state\":\"string\",\"postcode\":\"string\",\"suburb\":\"string\",\"country_name\":\"string\",\"longitude\":0,\"latitude\":0,\"lat_lon_precision\":0,\"mapable\":true,\"street_view_available\":true,\"address_located_in\":{\"located_in_name\":\"string\",\"located_in_address\":\"string\",\"located_in_website\":\"string\",\"located_in_phone_number\":\"string\"},\"address_identifiers\":{\"address_identifier_name\":\"string\",\"address_identifier_value\":\"string\"}},\"verification\":{\"is_verified\":true,\"last_verified_on\":\"2022-02-22T23:00:14.098Z\"},\"primary_contacts\":[{\"type_of_contact\":\"string\",\"value\":\"string\",\"display_value\":\"string\",\"label\":\"string\"}],\"suspicious_merchant_score\":0,\"secondary_contacts\":[{\"type_of_contact\":\"string\",\"value\":\"string\",\"display_value\":\"string\",\"label\":\"string\"}],\"secondary_addresses\":[{\"single_line_address\":\"string\",\"address_line_1\":\"string\",\"address_line_2\":\"string\",\"state\":\"string\",\"postcode\":\"string\",\"suburb\":\"string\",\"country_name\":\"string\",\"longitude\":0,\"latitude\":0,\"lat_lon_precision\":0,\"mapable\":true,\"street_view_available\":true,\"address_located_in\":{\"located_in_name\":\"string\",\"located_in_address\":\"string\",\"located_in_website\":\"string\",\"located_in_phone_number\":\"string\"},\"address_identifiers\":{\"address_identifier_name\":\"string\",\"address_identifier_value\":\"string\"}}],\"image_gallery\":[{\"thumbnail_url\":\"string\",\"large_url\":\"string\",\"image_title\":\"string\",\"image_height\":0,\"image_width\":0}],\"primary_category\":{\"category_name\":\"string\",\"id\":0,\"full_category_path\":\"string\",\"parent\":{\"category_name\":\"string\",\"id\":0,\"full_category_path\":\"string\",\"parent\":\"string\"},\"is_sensitive\":true,\"is_lwc_category\":true,\"is_substituted_category\":true,\"lwc_category_icon\":{\"Black_white_url\":\"string\",\"Coloured_url\":\"string\",\"height\":0,\"width\":0},\"lwc_category_icon_circular\":{\"Black_white_url\":\"string\",\"Coloured_url\":\"string\",\"height\":0,\"width\":0},\"category_emoji\":\"string\",\"recategorisation_info\":{\"rule_id\":0,\"rule_date_time\":\"2022-02-22T23:00:14.098Z\"},\"category_id\":\"string\"},\"other_categories_list\":[{\"category_name\":\"string\",\"id\":0,\"full_category_path\":\"string\",\"parent\":{\"category_name\":\"string\",\"id\":0,\"full_category_path\":\"string\",\"parent\":\"string\"},\"is_sensitive\":true,\"is_lwc_category\":true,\"is_substituted_category\":true,\"lwc_category_icon\":{\"Black_white_url\":\"string\",\"Coloured_url\":\"string\",\"height\":0,\"width\":0},\"lwc_category_icon_circular\":{\"Black_white_url\":\"string\",\"Coloured_url\":\"string\",\"height\":0,\"width\":0},\"category_emoji\":\"string\",\"recategorisation_info\":{\"rule_id\":0,\"rule_date_time\":\"2022-02-22T23:00:14.098Z\"},\"category_id\":\"string\"}],\"primary_merchant_description\":\"string\",\"merchant_descriptors\":[\"string\"],\"is_permanently_closed\":true,\"merchant_logo\":{\"url\":\"string\",\"height\":0,\"width\":0},\"merchant_logo_3x1\":{\"url\":\"string\",\"height\":0,\"width\":0},\"merchant_logo_circular\":{\"url\":\"string\",\"height\":0,\"width\":0},\"merchant_logo_dark\":{\"url\":\"string\",\"height\":0,\"width\":0},\"merchant_logo_3x1_dark\":{\"url\":\"string\",\"height\":0,\"width\":0},\"merchant_logo_circular_dark\":{\"url\":\"string\",\"height\":0,\"width\":0},\"website_screen_shot\":{\"thumbnail\":{\"url\":\"string\",\"height\":0,\"width\":0},\"banner\":{\"url\":\"string\",\"height\":0,\"width\":0},\"main\":{\"url\":\"string\",\"height\":0,\"width\":0}},\"also_known_as\":[\"string\"],\"associated_with\":[{\"associates_name\":\"string\",\"associates_id\":\"3fa85f64-5717-4562-b3fc-2c963f66afa6\",\"type_of_association\":\"string\"}],\"last_updated\":\"2022-02-22T23:00:14.098Z\",\"opening_hours\":{\"is_always_open\":true,\"sunday\":{\"status\":\"string\",\"times\":[{\"open\":\"string\",\"close\":\"string\"}]},\"monday\":{\"status\":\"string\",\"times\":[{\"open\":\"string\",\"close\":\"string\"}]},\"tuesday\":{\"status\":\"string\",\"times\":[{\"open\":\"string\",\"close\":\"string\"}]},\"wednesday\":{\"status\":\"string\",\"times\":[{\"open\":\"string\",\"close\":\"string\"}]},\"thursday\":{\"status\":\"string\",\"times\":[{\"open\":\"string\",\"close\":\"string\"}]},\"friday\":{\"status\":\"string\",\"times\":[{\"open\":\"string\",\"close\":\"string\"}]},\"saturday\":{\"status\":\"string\",\"times\":[{\"open\":\"string\",\"close\":\"string\"}]}},\"legal_business_info\":{\"date_established\":\"2022-02-22T23:00:14.098Z\",\"entity_type\":\"string\",\"current_merchant_status\":\"string\",\"merchant_type\":\"string\",\"merchant_presence\":\"string\",\"chain_name\":\"string\",\"legal_registrations\":[{\"legal_number\":\"string\",\"legal_number_label\":\"string\"}],\"registered_for_sales_tax\":true,\"chain_lwc_guid\":\"string\"},\"overall_rating\":{\"overall_rating_score\":0,\"total_number_of_reviews\":0},\"ratings\":[{\"reviewer\":\"string\",\"score\":0,\"number_of_reviews\":0}],\"payment_options\":[\"string\"],\"transaction_query_process\":{\"transaction_query_process_markdown\":\"string\",\"transaction_query_process_link\":\"string\"},\"transaction_query_process_html\":\"string\",\"transaction_query_process_json\":{\"paragraphs\":[{\"items\":[{\"value\":\"string\",\"type\":\"text\",\"display_value\":\"string\"}]}]},\"receipt_data_available\":true,\"record_is_quarantined\":true,\"additional_details\":{\"bpay_biller_codes\":[0],\"direct_entry_codes\":[0]},\"merchant_category_mappings\":[{\"type_of_category\":\"string\",\"category_identifier\":\"string\",\"category_name\":\"string\",\"mapping_error\":\"string\"}],\"merchant_tags\":[{\"id\":0,\"label\":\"string\"}],\"bsb_numbers\":[\"string\"],\"is_paymentaggregator\":true}}],\"bank_message_search_result\":{\"lwc_guid\":\"3fa85f64-5717-4562-b3fc-2c963f66afa6\",\"associated_with\":{\"associates_name\":\"string\",\"associates_id\":\"3fa85f64-5717-4562-b3fc-2c963f66afa6\",\"type_of_association\":\"string\"},\"short_bank_message\":\"string\",\"long_bank_message\":\"string\",\"primary_category\":{\"category_name\":\"string\",\"id\":0,\"full_category_path\":\"string\",\"parent\":{\"category_name\":\"string\",\"id\":0,\"full_category_path\":\"string\",\"parent\":\"string\"},\"is_sensitive\":true,\"is_lwc_category\":true,\"is_substituted_category\":true,\"lwc_category_icon\":{\"Black_white_url\":\"string\",\"Coloured_url\":\"string\",\"height\":0,\"width\":0},\"lwc_category_icon_circular\":{\"Black_white_url\":\"string\",\"Coloured_url\":\"string\",\"height\":0,\"width\":0},\"category_emoji\":\"string\",\"recategorisation_info\":{\"rule_id\":0,\"rule_date_time\":\"2022-02-22T23:00:14.101Z\"},\"category_id\":\"string\"},\"BankMessage_LWC_ID\":0,\"search_results_guid\":\"3fa85f64-5717-4562-b3fc-2c963f66afa6\",\"bank_message_tags\":[{\"id\":0,\"label\":\"string\"}]},\"shared_cal_search_result\":{\"shared_cal_comment\":\"string\",\"shared_name\":\"string\",\"shared_category\":{\"category_name\":\"string\",\"id\":0,\"full_category_path\":\"string\",\"parent\":{\"category_name\":\"string\",\"id\":0,\"full_category_path\":\"string\",\"parent\":\"string\"},\"is_sensitive\":true,\"is_lwc_category\":true,\"is_substituted_category\":true,\"lwc_category_icon\":{\"Black_white_url\":\"string\",\"Coloured_url\":\"string\",\"height\":0,\"width\":0},\"lwc_category_icon_circular\":{\"Black_white_url\":\"string\",\"Coloured_url\":\"string\",\"height\":0,\"width\":0},\"category_emoji\":\"string\",\"recategorisation_info\":{\"rule_id\":0,\"rule_date_time\":\"2022-02-22T23:00:14.101Z\"},\"category_id\":\"string\"}},\"atm_search_results\":[{\"search_results_guid\":\"3fa85f64-5717-4562-b3fc-2c963f66afa6\",\"ATM_ID\":0,\"lwc_guid\":\"3fa85f64-5717-4562-b3fc-2c963f66afa6\",\"rank\":0,\"score\":0,\"match_feedback_requested\":true,\"type_of_match\":\"string\",\"atm_details\":{\"id\":0,\"lwc_guid\":\"3fa85f64-5717-4562-b3fc-2c963f66afa6\",\"primary_category\":{\"category_name\":\"string\",\"id\":0,\"full_category_path\":\"string\",\"parent\":{\"category_name\":\"string\",\"id\":0,\"full_category_path\":\"string\",\"parent\":\"string\"},\"is_sensitive\":true,\"is_lwc_category\":true,\"is_substituted_category\":true,\"lwc_category_icon\":{\"Black_white_url\":\"string\",\"Coloured_url\":\"string\",\"height\":0,\"width\":0},\"lwc_category_icon_circular\":{\"Black_white_url\":\"string\",\"Coloured_url\":\"string\",\"height\":0,\"width\":0},\"category_emoji\":\"string\",\"recategorisation_info\":{\"rule_id\":0,\"rule_date_time\":\"2022-02-22T23:00:14.101Z\"},\"category_id\":\"string\"},\"atm_operator\":{\"associates_name\":\"string\",\"associates_id\":\"3fa85f64-5717-4562-b3fc-2c963f66afa6\",\"type_of_association\":\"string\"},\"primary_address\":{\"single_line_address\":\"string\",\"address_line_1\":\"string\",\"address_line_2\":\"string\",\"state\":\"string\",\"postcode\":\"string\",\"suburb\":\"string\",\"country_name\":\"string\",\"longitude\":0,\"latitude\":0,\"lat_lon_precision\":0,\"mapable\":true,\"street_view_available\":true,\"address_located_in\":{\"located_in_name\":\"string\",\"located_in_address\":\"string\",\"located_in_website\":\"string\",\"located_in_phone_number\":\"string\"},\"address_identifiers\":{\"address_identifier_name\":\"string\",\"address_identifier_value\":\"string\"}},\"primary_contacts\":[{\"type_of_contact\":\"string\",\"value\":\"string\",\"display_value\":\"string\",\"label\":\"string\"}],\"is_mobile\":true,\"atm_name\":\"string\",\"verification\":{\"is_verified\":true,\"last_verified_on\":\"2022-02-22T23:00:14.101Z\"},\"last_updated\":\"2022-02-22T23:00:14.101Z\",\"atm_logo\":{\"url\":\"string\",\"height\":0,\"width\":0},\"atm_logo_3x1\":{\"url\":\"string\",\"height\":0,\"width\":0},\"atm_logo_circular\":{\"url\":\"string\",\"height\":0,\"width\":0},\"atm_logo_dark\":{\"url\":\"string\",\"height\":0,\"width\":0},\"atm_logo_3x1_dark\":{\"url\":\"string\",\"height\":0,\"width\":0},\"atm_logo_circular_dark\":{\"url\":\"string\",\"height\":0,\"width\":0},\"record_is_quarantined\":true}}],\"user_message\":\"string\",\"system_message\":\"string\",\"result_code\":0,\"matched_on_id\":0,\"matched_with\":\"Unknown\",\"lwc_attempting_to_index\":true,\"transaction_search_time_ms\":0,\"credits_used\":0,\"is_quarantined\":true,\"is_recategorised\":true,\"transaction_category\":{\"category_name\":\"string\",\"id\":0,\"full_category_path\":\"string\",\"parent\":{\"category_name\":\"string\",\"id\":0,\"full_category_path\":\"string\",\"parent\":\"string\"},\"is_sensitive\":true,\"is_lwc_category\":true,\"is_substituted_category\":true,\"lwc_category_icon\":{\"Black_white_url\":\"string\",\"Coloured_url\":\"string\",\"height\":0,\"width\":0},\"lwc_category_icon_circular\":{\"Black_white_url\":\"string\",\"Coloured_url\":\"string\",\"height\":0,\"width\":0},\"category_emoji\":\"string\",\"recategorisation_info\":{\"rule_id\":0,\"rule_date_time\":\"2022-02-22T23:00:14.101Z\"},\"category_id\":\"string\"},\"transaction_category_mappings\":[{\"type_of_category\":\"string\",\"category_identifier\":\"string\",\"category_name\":\"string\",\"mapping_error\":\"string\"}],\"transaction_tags\":[{\"id\":0,\"label\":\"string\"}]}]}")
	var response Response

	err := json.Unmarshal(bytes, &response)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return &response
}

func retrieveMerchantResponse() []MerchantDetails {
	bytes := []byte("[{\"id\":0,\"lwc_guid\":\"3fa85f64-5717-4562-b3fc-2c963f66afa6\",\"data_enrichment_score\":0,\"merchant_primary_name\":\"string\",\"primary_address\":{\"single_line_address\":\"string\",\"address_line_1\":\"string\",\"address_line_2\":\"string\",\"state\":\"string\",\"postcode\":\"string\",\"suburb\":\"string\",\"country_name\":\"string\",\"longitude\":0,\"latitude\":0,\"lat_lon_precision\":0,\"mapable\":true,\"street_view_available\":true,\"address_located_in\":{\"located_in_name\":\"string\",\"located_in_address\":\"string\",\"located_in_website\":\"string\",\"located_in_phone_number\":\"string\"},\"address_identifiers\":{\"address_identifier_name\":\"string\",\"address_identifier_value\":\"string\"}},\"verification\":{\"is_verified\":true,\"last_verified_on\":\"2022-02-22T23:00:14.098Z\"},\"primary_contacts\":[{\"type_of_contact\":\"string\",\"value\":\"string\",\"display_value\":\"string\",\"label\":\"string\"}],\"suspicious_merchant_score\":0,\"secondary_contacts\":[{\"type_of_contact\":\"string\",\"value\":\"string\",\"display_value\":\"string\",\"label\":\"string\"}],\"secondary_addresses\":[{\"single_line_address\":\"string\",\"address_line_1\":\"string\",\"address_line_2\":\"string\",\"state\":\"string\",\"postcode\":\"string\",\"suburb\":\"string\",\"country_name\":\"string\",\"longitude\":0,\"latitude\":0,\"lat_lon_precision\":0,\"mapable\":true,\"street_view_available\":true,\"address_located_in\":{\"located_in_name\":\"string\",\"located_in_address\":\"string\",\"located_in_website\":\"string\",\"located_in_phone_number\":\"string\"},\"address_identifiers\":{\"address_identifier_name\":\"string\",\"address_identifier_value\":\"string\"}}],\"image_gallery\":[{\"thumbnail_url\":\"string\",\"large_url\":\"string\",\"image_title\":\"string\",\"image_height\":0,\"image_width\":0}],\"primary_category\":{\"category_name\":\"string\",\"id\":0,\"full_category_path\":\"string\",\"parent\":{\"category_name\":\"string\",\"id\":0,\"full_category_path\":\"string\",\"parent\":\"string\"},\"is_sensitive\":true,\"is_lwc_category\":true,\"is_substituted_category\":true,\"lwc_category_icon\":{\"Black_white_url\":\"string\",\"Coloured_url\":\"string\",\"height\":0,\"width\":0},\"lwc_category_icon_circular\":{\"Black_white_url\":\"string\",\"Coloured_url\":\"string\",\"height\":0,\"width\":0},\"category_emoji\":\"string\",\"recategorisation_info\":{\"rule_id\":0,\"rule_date_time\":\"2022-02-22T23:00:14.098Z\"},\"category_id\":\"string\"},\"other_categories_list\":[{\"category_name\":\"string\",\"id\":0,\"full_category_path\":\"string\",\"parent\":{\"category_name\":\"string\",\"id\":0,\"full_category_path\":\"string\",\"parent\":\"string\"},\"is_sensitive\":true,\"is_lwc_category\":true,\"is_substituted_category\":true,\"lwc_category_icon\":{\"Black_white_url\":\"string\",\"Coloured_url\":\"string\",\"height\":0,\"width\":0},\"lwc_category_icon_circular\":{\"Black_white_url\":\"string\",\"Coloured_url\":\"string\",\"height\":0,\"width\":0},\"category_emoji\":\"string\",\"recategorisation_info\":{\"rule_id\":0,\"rule_date_time\":\"2022-02-22T23:00:14.098Z\"},\"category_id\":\"string\"}],\"primary_merchant_description\":\"string\",\"merchant_descriptors\":[\"string\"],\"is_permanently_closed\":true,\"merchant_logo\":{\"url\":\"string\",\"height\":0,\"width\":0},\"merchant_logo_3x1\":{\"url\":\"string\",\"height\":0,\"width\":0},\"merchant_logo_circular\":{\"url\":\"string\",\"height\":0,\"width\":0},\"merchant_logo_dark\":{\"url\":\"string\",\"height\":0,\"width\":0},\"merchant_logo_3x1_dark\":{\"url\":\"string\",\"height\":0,\"width\":0},\"merchant_logo_circular_dark\":{\"url\":\"string\",\"height\":0,\"width\":0},\"website_screen_shot\":{\"thumbnail\":{\"url\":\"string\",\"height\":0,\"width\":0},\"banner\":{\"url\":\"string\",\"height\":0,\"width\":0},\"main\":{\"url\":\"string\",\"height\":0,\"width\":0}},\"also_known_as\":[\"string\"],\"associated_with\":[{\"associates_name\":\"string\",\"associates_id\":\"3fa85f64-5717-4562-b3fc-2c963f66afa6\",\"type_of_association\":\"string\"}],\"last_updated\":\"2022-02-22T23:00:14.098Z\",\"opening_hours\":{\"is_always_open\":true,\"sunday\":{\"status\":\"string\",\"times\":[{\"open\":\"string\",\"close\":\"string\"}]},\"monday\":{\"status\":\"string\",\"times\":[{\"open\":\"string\",\"close\":\"string\"}]},\"tuesday\":{\"status\":\"string\",\"times\":[{\"open\":\"string\",\"close\":\"string\"}]},\"wednesday\":{\"status\":\"string\",\"times\":[{\"open\":\"string\",\"close\":\"string\"}]},\"thursday\":{\"status\":\"string\",\"times\":[{\"open\":\"string\",\"close\":\"string\"}]},\"friday\":{\"status\":\"string\",\"times\":[{\"open\":\"string\",\"close\":\"string\"}]},\"saturday\":{\"status\":\"string\",\"times\":[{\"open\":\"string\",\"close\":\"string\"}]}},\"legal_business_info\":{\"date_established\":\"2022-02-22T23:00:14.098Z\",\"entity_type\":\"string\",\"current_merchant_status\":\"string\",\"merchant_type\":\"string\",\"merchant_presence\":\"string\",\"chain_name\":\"string\",\"legal_registrations\":[{\"legal_number\":\"string\",\"legal_number_label\":\"string\"}],\"registered_for_sales_tax\":true,\"chain_lwc_guid\":\"string\"},\"overall_rating\":{\"overall_rating_score\":0,\"total_number_of_reviews\":0},\"ratings\":[{\"reviewer\":\"string\",\"score\":0,\"number_of_reviews\":0}],\"payment_options\":[\"string\"],\"transaction_query_process\":{\"transaction_query_process_markdown\":\"string\",\"transaction_query_process_link\":\"string\"},\"transaction_query_process_html\":\"string\",\"transaction_query_process_json\":{\"paragraphs\":[{\"items\":[{\"value\":\"string\",\"type\":\"text\",\"display_value\":\"string\"}]}]},\"receipt_data_available\":true,\"record_is_quarantined\":true,\"additional_details\":{\"bpay_biller_codes\":[0],\"direct_entry_codes\":[0]},\"merchant_category_mappings\":[{\"type_of_category\":\"string\",\"category_identifier\":\"string\",\"category_name\":\"string\",\"mapping_error\":\"string\"}],\"merchant_tags\":[{\"id\":0,\"label\":\"string\"}],\"bsb_numbers\":[\"string\"],\"is_paymentaggregator\":true}]")
	var response []MerchantDetails

	err := json.Unmarshal(bytes, &response)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return response
}

func TestClient_RetrieveMerchants(t *testing.T) {
	tests := []struct {
		name           string
		request        Request
		want           []MerchantDetails
		wantErr        string
		requestHandler http.HandlerFunc
	}{
		{
			name: "successfully request merchants ",
			request: Request{
				BankTransactions: []BankTransactions{{Cal: "2321"}, {Cal: "3123123"}},
			},
			want: retrieveMerchantResponse(),
			requestHandler: func(rw http.ResponseWriter, req *http.Request) {
				data, _ := json.Marshal(getResponse())
				_, _ = rw.Write(data)
			},
		},
		{
			name: "error when passing in empty bank transactions list ",
			request: Request{
				BankTransactions: []BankTransactions{},
			},
			wantErr: "CAL list in request is empty",
			requestHandler: func(rw http.ResponseWriter, req *http.Request) {
				data, _ := json.Marshal(getResponse())
				_, _ = rw.Write(data)
			},
		},
		{
			name: "handle empty response ",
			request: Request{
				BankTransactions: []BankTransactions{{Cal: "2321"}, {Cal: "3123123"}},
			},
			wantErr: "unexpected response from downstream",
			requestHandler: func(rw http.ResponseWriter, req *http.Request) {
				data, _ := json.Marshal([]MerchantDetails{})
				_, _ = rw.Write(data)
			},
		},
		{
			name: "fail to unmarshall response body",
			request: Request{
				BankTransactions: []BankTransactions{{Cal: "2321"}, {Cal: "3123123"}},
			},
			requestHandler: func(rw http.ResponseWriter, req *http.Request) {
				_, _ = rw.Write([]byte{32})
			},
			wantErr: "unexpected response from downstream",
		},
		{
			name: "handle 404",
			request: Request{
				BankTransactions: []BankTransactions{{Cal: "2321"}, {Cal: "3123123"}},
			},
			wantErr: "fabric error: status_code=NotFound, error_code=2, message=failed retrieve merchants request, reason=unexpected response from downstream",
			requestHandler: func(rw http.ResponseWriter, req *http.Request) {
				http.Error(rw, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(test.requestHandler)

			defer server.Client()
			c := &client{
				httpClient: server.Client(),
				baseURL:    server.URL,
			}

			got, err := c.RetrieveMerchants(testutil.GetContext(true), test.request)

			if test.wantErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), test.wantErr)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, test.want, got)
		})
	}
}
