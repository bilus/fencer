package countries

import (
	"errors"
)

// TODO: Load from file, same way radiodns does it.
var gccToIsoMapping = map[string]string{"AF0": "AF", "9E0": "AL", "2E0": "DZ", "3E0": "AD", "6D0": "AO", "1A2": "AI", "2A2": "AG", "AA2": "AR", "AE4": "AM", "3A4": "AW", "1F0": "AU", "2F0": "AU", "3F0": "AU", "4F0": "AU", "5F0": "AU", "6F0": "AU", "7F0": "AU", "8F0": "AU", "AE0": "AT", "BE3": "AZ", "FA2": "BS", "EF0": "BH", "3F1": "BD", "5A2": "BB", "FE3": "BY", "6E0": "BE", "6A2": "BZ", "ED0": "BJ", "CA2": "BM", "2F1": "BT", "1A3": "BO", "FE4": "BA", "BD1": "BW", "BA2": "BR", "FA5": "VI", "BF1": "BN", "8E1": "BG", "BD0": "BF", "BF0": "MM", "9D1": "BI", "3F2": "KH", "1D0": "CM", "CA1": "CA", "6D1": "CV", "7A2": "KY", "2D0": "CF", "9D2": "TD", "CA3": "CL", "CF0": "CN", "2A3": "CO", "CD1": "KM", "CD0": "CG", "8A2": "CR", "CD2": "CI", "CE3": "HR", "9A2": "CU", "2E1": "CY", "2E2": "CZ", "9E1": "FO", "3D0": "DJ", "AA3": "DM", "BA3": "DO", "3A2": "EC", "FE0": "EG", "CA4": "SV", "7D0": "GQ", "2E4": "EE", "ED1": "ET", "4A2": "FK", "5F1": "FJ", "6E1": "FI", "FE1": "FR", "8D0": "GA", "8D1": "GM", "CE4": "GE", "DE0": "DE", "1E0": "DE", "3D1": "GH", "AE1": "GI", "1E1": "GR", "FA1": "GL", "DA3": "GD", "1A4": "GT", "9D0": "GN", "AD2": "GW", "FA3": "GY", "DA4": "HT", "2A4": "HN", "FF1": "HK", "BE0": "HU", "AE2": "IS", "5F2": "IN", "CF2": "ID", "8F1": "IR", "BE1": "IQ", "2E3": "IE", "4E0": "IL", "5E0": "IT", "3A3": "JM", "9F2": "JP", "5E1": "JO", "DE3": "KZ", "6D2": "KE", "1F1": "KI", "DF0": "KP", "EF1": "KR", "1F2": "KW", "3E4": "MK", "1F3": "LA", "9E3": "LV", "AE3": "LB", "6D3": "LS", "2D1": "LR", "DE1": "LY", "9E2": "LI", "CE2": "LT", "7E1": "LU", "6F2": "MO", "4D0": "MG", "FD0": "MW", "FF0": "MY", "BF2": "MV", "5D0": "ML", "CE0": "MT", "4D1": "MR", "AD3": "MU", "FA4": "MX", "EF3": "FM", "1E4": "MD", "BE2": "MC", "FF3": "MN", "1E3": "ME", "5A4": "MS", "1E2": "MA", "3D2": "MZ", "1D1": "NA", "7F1": "NR", "EF2": "NP", "8E3": "NL", "9F1": "NZ", "7A3": "NI", "8D2": "NE", "FD1": "NG", "FE2": "NO", "6F1": "OM", "4F1": "PK", "9A3": "PA", "9F3": "PG", "6A3": "PY", "7A4": "PE", "8F2": "PH", "8E4": "PL", "8E0": "PT", "8A3": "PR", "2F2": "QA", "EE1": "RO", "7E0": "RU", "5D3": "RW", "AD1": "SH", "AA4": "KN", "FA6": "PM", "CA5": "VC", "4F2": "WS", "3E1": "SM", "9F0": "SA", "7D1": "SN", "DE2": "RS", "BA4": "SC", "1D2": "SL", "AF2": "SG", "5E2": "SK", "9E4": "SI", "AF1": "SB", "7D2": "SO", "AD0": "ZA", "AD4": "SS", "EE2": "ES", "CF1": "LK", "CD3": "SD", "8A4": "SR", "5D2": "SZ", "EE3": "SE", "4E1": "CH", "DF1": "TW", "5E3": "TJ", "DD1": "TZ", "2F3": "TH", "DD0": "TG", "3F3": "TO", "6A4": "TT", "7E2": "TN", "3E3": "TR", "EE4": "TM", "EA3": "TC", "4D2": "UG", "6E4": "UA", "DF2": "AE", "CE1": "GB", "1A0": "US", "2A0": "US", "3A0": "US", "4A0": "US", "5A0": "US", "6A0": "US", "7A0": "US", "8A0": "US", "9A0": "US", "AA0": "US", "BA0": "US", "DA0": "US", "EA0": "US", "9A4": "UY", "BE4": "UZ", "FF2": "VU", "4E2": "VA", "EA4": "VE", "7F2": "VN", "3D3": "EH", "BF3": "YE", "ED2": "ZM", "2D2": "ZW"}

var MissingCodeError = errors.New("No such country code")

func GccToIso(gcc string) (string, error) {
	iso, found := gccToIsoMapping[gcc]
	if found {
		return iso, nil
	}
	return "", MissingCodeError
}
