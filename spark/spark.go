package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var VERSION = "v0.0.1"

const (
	StatusOK                  = 200
	StatusCreated             = 201
	StatusBadRequest          = 400
	StatusUnauthorized        = 401
	StatusForbidden           = 403
	StatusNotFound            = 404
	StatusInternalServerError = 500
	ErrorNotFound             = "Resource not found"
	ErrorUnauthorized         = "Unauthorized access"
)

func createFolders() {
	folders := []string{
		APP_SETTINGS.templates,
		APP_SETTINGS.layouts,
		APP_SETTINGS.views,
		APP_SETTINGS.static,
	}
	if APP_SETTINGS.uploads != "" {
		folders = append(folders, APP_SETTINGS.uploads)
	}

	for _, f := range folders {
		if f == "" {
			continue
		}
		if _, err := os.Stat(f); os.IsNotExist(err) {
			if err := os.MkdirAll(f, 0755); err != nil && APP_SETTINGS.debug {
				log.Printf("failed to create folder %s: %v", f, err)
			} else if APP_SETTINGS.debug {
				log.Printf("created missing folder %s", f)
			}
		}
	}
}

var DB *gorm.DB

func InitDB() {
	var err error
	switch APP_SETTINGS.db_driver {
	case "sqlite":
		DB, err = gorm.Open(sqlite.Open(APP_SETTINGS.database), &gorm.Config{})
	case "mysql":
		DB, err = gorm.Open(mysql.Open(APP_SETTINGS.database), &gorm.Config{})
	case "postgres":
		DB, err = gorm.Open(postgres.Open(APP_SETTINGS.database), &gorm.Config{})
	default:
		log.Fatal("unsupported db driver:", APP_SETTINGS.db_driver)
	}

	if err != nil {
		log.Fatal("failed to connect database:", err)
	}
}

func AutoMigrate(models ...interface{}) {
	err := DB.AutoMigrate(models...)
	if err != nil {
		log.Fatal("migration failed:", err)
	}
}

var Router *gin.Engine

// Beginner-friendly wrapper around gin.Context
type Request struct {
	*gin.Context
	Route  string            // the matched route (e.g. "/user/:id")
	Params map[string]string // path params
	Query  map[string]string // query params (?a=1&b=2)
}

// Route = handle both GET and POST
func Route(handler func(Request)) {
	// catch-all, let the handler check METHOD + PATH
	Router.Any("/*path", func(c *gin.Context) {
		req := buildRequest(c, c.FullPath())
		handler(req)
	})
}

// buildRequest extracts params and query
func buildRequest(c *gin.Context, route string) Request {
	// extract params
	params := make(map[string]string)
	for _, p := range c.Params {
		params[p.Key] = p.Value
	}

	// extract query params
	query := make(map[string]string)
	for k, v := range c.Request.URL.Query() {
		if len(v) > 0 {
			query[k] = v[0]
		}
	}

	return Request{
		Context: c,
		Route:   route,
		Params:  params,
		Query:   query,
	}
}

func (r Request) GET() bool  { return r.Request.Method == "GET" }
func (r Request) POST() bool { return r.Request.Method == "POST" }

// Get = GET only
func Get(path string, handler func(Request)) {
	Router.GET(path, func(c *gin.Context) {
		handler(buildRequest(c, path))
	})
}

// Post = POST only
func Post(path string, handler func(Request)) {
	Router.POST(path, func(c *gin.Context) {
		handler(buildRequest(c, path))
	})
}

// Run starts the server (default :8000)
func Run() {
	addr := APP_SETTINGS.address
	if addr == "" {
		addr = ":8000" // safe default
	}
	Router.Run(addr)
}

// Middleware type for beginners
type Middleware func(Request)

// Use adds global middleware
func Use(m Middleware) {
	Router.Use(func(c *gin.Context) {
		m(buildRequest(c, c.FullPath()))
		c.Next() // continue to next middleware/handler
	})
}

// UseOn adds middleware only for a specific route
func UseOn(path string, m Middleware) {
	Router.Use(func(c *gin.Context) {
		if c.FullPath() == path {
			m(buildRequest(c, path))
		}
		c.Next()
	})
}

// -------- Static & File Upload Helpers -------- //

// ServeStatic = serve files from APP_SETTINGS.static folder
func ServeStatic() {
	if APP_SETTINGS.static != "" && APP_SETTINGS.static_url != "" {
		Router.Static(APP_SETTINGS.static_url, APP_SETTINGS.static)
	}
}

// SaveUploadedFile = save an uploaded file to target directory
func (r Request) SaveUploadedFile(field string, dst ...string) error {
	file, err := r.FormFile(field)
	if err != nil {
		return err
	}

	target := ""
	if len(dst) > 0 && dst[0] != "" {
		target = dst[0] // use custom path if provided
	} else if APP_SETTINGS.uploads != "" {
		target = filepath.Join(APP_SETTINGS.uploads, file.Filename) // default upload folder
	} else {
		return fmt.Errorf("no upload destination provided and APP_SETTINGS.uploads not set")
	}

	return r.Context.SaveUploadedFile(file, target)
}

// HashPassword hashes a plaintext password
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

// CheckPassword verifies a plaintext password against a hash
func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

var JWT_SECRET = []byte("supersecret") // default, can be overridden

// GenerateToken creates a JWT for a user
func GenerateToken(userID string, expiry time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(expiry).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(JWT_SECRET)
}

// VerifyToken parses and validates a JWT
func VerifyToken(tokenStr string) (*jwt.Token, jwt.MapClaims, error) {
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return JWT_SECRET, nil
	})
	return token, claims, err
}

// JWTMiddleware checks for Authorization header
func JWTMiddleware() Middleware {
	return func(r Request) {
		auth := r.GetHeader("Authorization")
		if auth == "" {
			r.AbortWithStatusJSON(401, gin.H{"error": "missing token"})
			return
		}
		token, claims, err := VerifyToken(auth)
		if err != nil || !token.Valid {
			r.AbortWithStatusJSON(401, gin.H{"error": "invalid token"})
			return
		}
		// Attach claims to context for later use
		r.Set("claims", claims)
		r.Next()
	}
}

func LoggerMiddleware() Middleware {
	return func(r Request) {
		log.Printf("[%s] %s", r.Request.Method, r.Request.URL.Path)
		r.Next()
	}
}

func CORSMiddleware() Middleware {
	return func(r Request) {
		r.Header("Access-Control-Allow-Origin", "*")
		r.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		r.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		if r.Request.Method == "OPTIONS" {
			r.AbortWithStatus(204)
			return
		}
		r.Next()
	}
}

// Automatically init the module
func init() {
	createFolders()
	Router = gin.Default()
	InitDB()
	ServeStatic()

	if APP_SETTINGS.need_template {
		Router.HTMLRender = func() multitemplate.Renderer {
			r := multitemplate.NewRenderer()

			// load layouts
			layouts, err := filepath.Glob(filepath.Join(APP_SETTINGS.layouts, "*.html"))
			if err != nil && APP_SETTINGS.debug {
				log.Fatal("failed to read layouts:", err)
			}

			// load views
			views, err := filepath.Glob(filepath.Join(APP_SETTINGS.views, "*.html"))
			if err != nil && APP_SETTINGS.debug {
				log.Fatal("failed to read views:", err)
			}

			// add templates
			for _, view := range views {
				var files []string
				if APP_SETTINGS.use_layouts && len(layouts) > 0 {
					files = append(layouts, view) // layouts + view
				} else {
					files = []string{view} // only view
				}

				name := filepath.Base(view)
				r.AddFromFiles(name, files...)
			}

			// if no views at all
			if len(views) == 0 && APP_SETTINGS.debug {
				log.Println("no views found in", APP_SETTINGS.views)
			}

			return r
		}()
	}
}
