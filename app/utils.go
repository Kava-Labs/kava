package app

import (
	"fmt"
	"io/ioutil"
)

// ExportStateToJSON util function to export the app state to JSON
func ExportStateToJSON(app *App, path string) error {
	fmt.Println("exporting app state...")
	appState, _, err := app.ExportAppStateAndValidators(false, nil)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, []byte(appState), 0644)
}
