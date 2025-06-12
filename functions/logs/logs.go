package logs

import (
	"fmt"
	"os"
)

func Logs(message ...any) {
	if os.Getenv("LOG") != "false" {
		fmt.Println(message...)
	}
}
