package main

/*


 */

type Settings struct {
	templates     string
	layouts       string
	views         string
	static        string
	static_url    string // URL prefix (e.g. "/static" or "/assets")
	uploads       string
	debug         bool
	need_template bool
	use_layouts   bool
	address       string
	db_driver     string // "sqlite", "mysql", "postgres"
	database      string // file path or DSN
}

var APP_SETTINGS = Settings{
	templates:     "./templates/",
	layouts:       "./templates/layouts/",
	views:         "./templates/views/",
	static:        "./static/",
	static_url:    "/static",
	uploads:       "/uploads/",
	debug:         true,
	need_template: true,
	use_layouts:   true,
	address:       ":8000",
	db_driver:     "sqlite", // default for beginners
	database:      "app.db", // default SQLite file
}
