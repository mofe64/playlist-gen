package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var envLoaded = false

func loadEnv() {
	if envLoaded {
		return
	} else {
		err := godotenv.Load()
		if err != nil {
			log.Fatalln("There was an error loading the env file")
		}
		envLoaded = true
	}
}

func EnvSpotifyClientId() string {
	loadEnv()
	return os.Getenv("spotify_client_id")
}

func EnvSpotifyClientSecret() string {
	loadEnv()
	return os.Getenv("spotify_client_secret")
}

func EnvMongoURI() string {
	loadEnv()
	return os.Getenv("mongo_uri")
}

func EnvDatabaseName() string {
	return os.Getenv("database_name")
}

func EnvHTTPPort() string {
	loadEnv()
	return os.Getenv("http_port")
}

func EnvProfile() string {
	loadEnv()
	return os.Getenv("profile")
}
