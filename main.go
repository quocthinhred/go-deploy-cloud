package main

import (
	"example.com/micro/action"
	"example.com/micro/client/login"
	"gitlab.com/thuocsi.vn-sdk/go-sdk/sdk"
)

var app *sdk.App

func main() {

	app = sdk.NewApp("Autologin project")

	login.InitLoginStgClient()
	login.InitLoginDevClient()

	var work = app.SetupWorker()
	work = work.SetDelay(1)
	work = work.SetRepeatPeriod(24 * 60 * 60)
	work = work.SetTask(action.AutoLoginTask)

	app.Launch()
}
