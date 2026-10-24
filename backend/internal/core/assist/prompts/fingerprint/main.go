// Command fingerprint prints the configuration fingerprint of the prompts and
// settings built into the binary, the same value the health check reports.
package main

import (
	"fmt"

	"geolineai/internal/core/assist/prompts"
)

func main() {
	fmt.Println(prompts.Fingerprint)
}
