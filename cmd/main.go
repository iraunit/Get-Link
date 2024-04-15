package cmd

func main() {
	app := InitializeApp()
	if err := app.Start(); err != nil {
		panic(err)
	}
}
