package main

import (
	"fmt"
	"os"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/cmd/kava/cmd"
)

func main() {
	rootCmd := cmd.NewRootCmd()

	//if err := autoCliOpts.EnhanceRootCommand(rootCmd); err != nil {
	//	panic(err)
	//}

	if err := svrcmd.Execute(rootCmd, cmd.EnvPrefix, app.DefaultNodeHome); err != nil {
		fmt.Println("error for main: ", err)
		os.Exit(1)
		//switch e := err.(type) {
		////case server.ErrorCode:
		////	os.Exit(e.Code)
		//
		//default:
		//	os.Exit(1)
		//}
	}
}
