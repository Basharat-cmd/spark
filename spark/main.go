package main

func main() {
	Get("/", func(r Request) {
		r.HTML(200, "index.html", nil)
	})

	Post("/submit", func(r Request) {
		name := r.PostForm("name")
		r.String(200, "Hello "+name)
	})

	Run()
}
