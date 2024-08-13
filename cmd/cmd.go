package main

import (
	"flag"
	"fmt"

	http "github.com/theyudiriski/billing-service/cmd/server"
)

func main() {
	runnerMap := map[string]func() Runner{
		"api": func() Runner { return http.NewServer() },
	}

	var serverType string
	flag.StringVar(
		&serverType,
		"type",
		"api",
		fmt.Sprintf("provide server to run. one of %v", func() []string {
			var keys []string
			for key := range runnerMap {
				keys = append(keys, key)
			}
			return keys
		}()),
	)
	flag.Parse()

	getRunner, ok := runnerMap[serverType]
	if !ok {
		panic("runner not found")
	}

	if err := RunApp(getRunner()); err != nil {
		panic(fmt.Sprintf("cannot run app %s", serverType))
	}
}
