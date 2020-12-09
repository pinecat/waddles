package util

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/pelletier/go-toml"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	//Cfg holds the current config information in a Config struct
	Cfg Config
)

// Config holds bot config information
type Config struct {
	Wadl struct {
		LogLevel string `toml:"log-level"`
		Prefix   string `toml:"prefix"`
		Token    string `toml:"token"`
		GuildID  string `toml:"guild-id"`
	} `toml:"waddles" comment:"General Bot Configuration"`
	Db struct {
		Host string `toml:"host"`
		Port string `toml:"port"`
		User string `toml:"user"`
		Pass string `toml:"pass"`
		Name string `toml:"database-name"`
		URL  string `toml:"url" commented:"true" comment:"uncomment to use a postgres URI instead of above"`
	} `toml:"database" comment:"Postgresql Database Connection Information"`
	NitroPerk struct {
		BoosterChannel struct {
			ParentID string `toml:"parent-id" comment:"Discord catagory ID for channels to be managed under"`
		} `toml:"booster-channel" comment:"server booster personal channel options"`
	} `toml:"nitro" comment:"perks related to being a server booster"`
	configDir string
}

//ReadConfig parses the config file into a Config struct
func ReadConfig() *Config {
	configDir := os.Getenv("WADL_CONFIG_DIR")

	if configDir == "" {
		pwd, _ := os.Getwd()
		configDir = pwd + "/config/"
		log.Warn().Msgf("WADL_CONFIG_DIR not set, defaulting to working dir (%s)", configDir)
	}

	if !strings.HasSuffix(configDir, "/") {
		configDir = path.Clean(configDir) + "/"
	}

	Cfg = Config{configDir: configDir}

	configFile := Cfg.GetConfigFileLocation("waddles.toml")

	if !FileExists(configFile) {
		Cfg.configDir = ""

		var bytes bytes.Buffer
		err := toml.NewEncoder(&bytes).Order(toml.OrderPreserve).Encode(Cfg)

		if err != nil {
			log.Panic().Err(err).Msg("Unable to save sample config file.")
		}

		ioutil.WriteFile(configFile, bytes.Bytes(), 0644)
		log.Fatal().Msgf("Config file doesn't exist. An example has been saved in its place.")
	}

	// Read config file from the file
	bytes, err := ioutil.ReadFile(configFile)

	if err != nil {
		log.Fatal().Err(err).Msgf("Unable to read config file at: '%s'", configFile)
	}

	// Unmarshal the config file bytes into a Config struct
	err = toml.Unmarshal(bytes, &Cfg)

	if err != nil {
		log.Fatal().Err(err).Msg("Unable to parse config file.")
	}

	log.Debug().Msgf("Read config file: %s", configFile)
	log.Trace().Msgf("Config Struct: %+v", Cfg)

	logLevel, err := zerolog.ParseLevel(Cfg.Wadl.LogLevel)
	if err != nil {
		log.Warn().Msgf("Supplied config file log level (%s) is invalid. Defaulting to info.", Cfg.Wadl.LogLevel)
		logLevel = zerolog.InfoLevel
	}

	log.Info().Msgf("Log Level set to: %s", logLevel.String())

	// Set global log level
	zerolog.SetGlobalLevel(logLevel)

	return &Cfg
}

//GetConfigFileLocation returns the full path of the requested configFile
func (config Config) GetConfigFileLocation(configFile string) string {
	return config.configDir + configFile
}
