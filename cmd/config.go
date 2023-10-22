package main

type config struct {
	API         apiConfig `yaml:"api"`
	DatabaseURL string    `yaml:"databaseURL"`
}

type apiConfig struct {
	Port int `yaml:"port"`
}
