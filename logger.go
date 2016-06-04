package gemini

import (
	"log"
	"os"
)

var logger *log.Logger

func init() {
	logger = log.New(os.Stderr, "gemini api: ", log.Ldate | log.Ltime | log.Lshortfile)
}
