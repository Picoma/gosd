package completions

import (
	"fmt"
	"log"
	"strings"

	"github.com/urfave/cli/v3"

	"gosd/sd"
)

func writeCandidates(command *cli.Command, candidates []sd.Completion) {
	writer := command.Root().Writer
	for _, candidate := range candidates {
		description := strings.NewReplacer("\r", " ", "\n", " ").Replace(candidate.Description)
		var err error
		if description == "" {
			_, err = fmt.Fprintln(writer, candidate.Name)
		} else {
			_, err = fmt.Fprintf(writer, "%s:%s\n", candidate.Name, description)
		}
		if err != nil {
			log.Printf("gosd: write completion: %v", err)
			return
		}
	}
}

func reportError(command *cli.Command, operation string, err error) {
	if _, writeErr := fmt.Fprintf(command.Root().ErrWriter, "gosd: %s: %v\n", operation, err); writeErr != nil {
		log.Printf("gosd: %s: %v (also failed to write diagnostic: %v)", operation, err, writeErr)
	}
}
