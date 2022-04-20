package config

import (
	"html/template"
	"log"
	"path/filepath"
	"github.com/joho/godotenv"
	"os"
	"backend/modules/mail"
)

type AppConfig struct {
	UseCache 		bool
	TemplateCache 	map[string]*template.Template
	InfoLog			*log.Logger
	ErrorLog		*log.Logger
	InProduction	bool
	Addr			string
	Static			string
	Media			string
	Host			string
	Origin			string
	AllowOrigin		string
	CredentialsPath	string
	Email			mail.Mail
}
type APIKey struct {
	GitHubClientId string
	GitHubSecretId string
}

var App AppConfig
var ApiKey APIKey
var infolog *log.Logger
var errorlog *log.Logger

func init() {

	commonSettings()

	appEnv, err := readEnvFile(); if err!= nil {
		// if .env is not in local and production environment
		panic("Not found .env")
	}

	if appEnv == "local" {
		configureLocalSettings()
	} else {
		configureProdSettings()
	}
}

func readEnvFile() (string, error) {

	// local
    err := godotenv.Load(".env.dev"); if err == nil {
		return os.Getenv("APP_ENV"), nil
	}

	// production
	err = godotenv.Load(".env"); if err == nil {
		return os.Getenv("APP_ENV"), nil
	}

	return os.Getenv("APP_ENV"), err
}

func commonSettings() {
	// log
	infolog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorlog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	App.InfoLog = infolog
	App.ErrorLog = errorlog
	App.AllowOrigin = os.Getenv("ALLOW_ORIGIN")
	App.Host = os.Getenv("APP_HOST")
	App.Origin = os.Getenv("APP_ORIGIN")
	App.CredentialsPath = os.Getenv("CREDENTIALS_PATH")

	// file path
	App.Static = "public"
	App.Media = "upload"

	// APIKey
	ApiKey.GitHubClientId = os.Getenv("GITHUB_CLIENT_ID")
	ApiKey.GitHubSecretId = os.Getenv("GITHUB_SECRET_ID")

	// email
	App.Email.From = os.Getenv("EMAIL_FROM")
	App.Email.Username = os.Getenv("EMAIL_USERNAME")
	App.Email.Password = os.Getenv("EMAIL_PASSWORD")
}

func CreateTemplateCache() (map[string]*template.Template, error) {
	myCache := map[string]*template.Template{}

	pages, err := filepath.Glob("./templates/pages/*.html")
	if err != nil {
		return myCache, err
	}
	for _, page := range pages {
		name := filepath.Base(page)
		ts := template.Must(template.New(name).ParseFiles(page))
		matches, err := filepath.Glob("./templates/layouts/app.html")
		if err != nil {
			return myCache, err
		}

		if len(matches) > 0 {
			ts, err = ts.ParseGlob("./templates/layouts/app.html")
			if err != nil {
				return myCache, err
			}
		}

		myCache[name] = ts
	}
	return myCache, nil

}
