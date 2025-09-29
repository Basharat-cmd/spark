Spark 🚀

Spark is a beginner-friendly web framework for Go.
It’s built on top of Gin
 and GORM
 but provides simple wrappers that feel like Python’s Django — without the heavy magic.

With Spark you can:

Define routes in one line

Work with GET/POST easily

Add middleware without boilerplate

Use JWT authentication in a beginner-friendly way

Manage static files, templates, uploads, and databases with defaults that just work

✨ Features

Easy Routing: Get, Post, or Route — your choice.

Beginner-friendly Request wrapper: access Params, Query, GET(), POST() directly.

Middleware Made Simple: just Use(LoggerMiddleware()).

JWT Helpers: generate and verify tokens without diving into complex configs.

Database Agnostic: SQLite (default), MySQL, Postgres — pick with one setting.

File Uploads: one-liner save with default upload path.

Template Rendering: layouts and views with auto-discovery.

Constants for Beginners: StatusOK, StatusNotFound, ErrorUnauthorized, etc.

📦 Installation

Clone this repo and build your project by editing main.go.

git clone https://github.com/Basharat-cmd/spark
cd spark
go run main.go

you can change the name from "spark" to something else like "myapp"
but then you had to change the the first line of go.mod:
    module spark
to:
    module myapp

after all just 
    run go mod tidy

and it will install all needed dependencies

Unlike typical Go libraries, Spark is designed to be your starter project, not a go get dependency.

⚡ Quick Start
Example main.go
package main

func main() {
    // Example: simple GET
    Get("/hello", func(r Request) {
        r.String(StatusOK, "Hello, World!")
    })

    // Example: route with params & POST
    Route(func(r Request) {
        if r.Route == "/user/:id" {
            if r.GET() {
                r.String(StatusOK, "GET user id=%s", r.Params["id"])
            } else if r.POST() {
                name := r.PostForm("name")
                r.String(StatusOK, "POST user id=%s, name=%s", r.Params["id"], name)
            }
        }
    })

    // Example: use middleware
    Use(LoggerMiddleware())
    Use(CORSMiddleware())

    Run()
}


Run:

go run main.go


Visit:

http://localhost:8000/hello

🔐 JWT Example
// Generate token
token, _ := GenerateToken("user123", time.Hour)

// Protect route with JWT
Use(JWTMiddleware())

Get("/profile", func(r Request) {
    claims := r.MustGet("claims").(map[string]interface{})
    r.JSON(StatusOK, gin.H{"user": claims["user_id"]})
})

🗄 Database Example
// Example model
type User struct {
    ID    uint   `gorm:"primaryKey"`
    Name  string
    Email string
}

// Auto migrate
AutoMigrate(&User{})

// Create
DB.Create(&User{Name: "Alice", Email: "alice@test.com"})

// Query
var users []User
DB.Find(&users)

⚙️ Settings (APP_SETTINGS)

Spark reads from APP_SETTINGS (edit in settings.go):

var APP_SETTINGS = struct {
    debug        bool
    address      string
    db_driver    string
    database     string
    templates    string
    layouts      string
    views        string
    static       string
    static_url   string
    uploads      string
    need_template bool
    use_layouts   bool
}{
    debug:        true,
    address:      ":8000",
    db_driver:    "sqlite",   // sqlite | mysql | postgres
    database:     "app.db",
    templates:    "templates",
    layouts:      "templates/layouts",
    views:        "templates/views",
    static:       "static",
    static_url:   "/static",
    uploads:      "uploads",
    need_template: true,
    use_layouts:   true,
}

🚧 Roadmap

 Sessions support

 Mailer helper

 Built-in Admin panel (like Django)

 CLI tool to scaffold apps

🤝 Contributing

PRs are