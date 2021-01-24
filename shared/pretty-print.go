package shared

import (
	"encoding/json"
)

// PrettyPrint JSON prints with nice formatting
func PrettyPrint(data interface{}) string {

	empJSON, _ := json.MarshalIndent(data, "", "  ")

	return string(empJSON)
}
