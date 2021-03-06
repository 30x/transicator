/*
Copyright 2016 The Transicator Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func setConfigDefaults() {

	pflag.IntP("port", "p", -1, "HTTP Listen port")
	viper.SetDefault("port", -1)
	pflag.IntP("secureport", "t", -1, "HTTPS Listen port")
	viper.SetDefault("securePort", -1)
	pflag.Int("mgmtport", -1, "Management port (for health checks)")
	viper.SetDefault("mgmtPort", -1)

	pflag.StringP("dbdir", "d", "", "Location of database files")
	viper.SetDefault("dbDir", "")
	pflag.StringP("url", "u", "", "URL to connect to Postgres")
	viper.SetDefault("pgURL", "")
	pflag.StringP("pgslot", "s", "", "Slot name for Postgres logical replication")
	viper.SetDefault("pgSlot", "")
	pflag.StringP("maxage", "m", "", "Purge records older than this age.")
	viper.SetDefault("maxAgeParam", "")
	pflag.String("cert", "", "TLS certificate PEM file")
	viper.SetDefault("cert", "")
	pflag.String("key", "", "TLS key PEM file (must be unencrypted)")
	viper.SetDefault("key", "")
	pflag.StringP("prefix", "P", "", "Optional prefix URL for all API calls")
	pflag.StringP("selectorcolumn", "S", "", "Set the selector column")
	viper.SetDefault("prefix", "")
	viper.SetDefault("selectorColumn", defaultSelectorColumn)

	pflag.StringP("config", "C", "", "specify the config directory (ONLY) for changeserver.properties")
	pflag.BoolP("debug", "D", false, "Turn on debugging")
	viper.SetDefault("debug", false)
}

func getConfig() error {

	viper.BindPFlag("port", pflag.Lookup("port"))
	viper.BindPFlag("securePort", pflag.Lookup("secureport"))
	viper.BindPFlag("mgmtPort", pflag.Lookup("mgmtport"))

	viper.BindPFlag("dbDir", pflag.Lookup("dbdir"))
	viper.BindPFlag("pgURL", pflag.Lookup("url"))
	viper.BindPFlag("pgSlot", pflag.Lookup("pgslot"))
	viper.BindPFlag("maxAgeParam", pflag.Lookup("maxage"))
	viper.BindPFlag("cert", pflag.Lookup("cert"))
	viper.BindPFlag("key", pflag.Lookup("key"))
	viper.BindPFlag("prefix", pflag.Lookup("prefix"))

	viper.BindPFlag("configFile", pflag.Lookup("config"))
	viper.BindPFlag("debug", pflag.Lookup("debug"))
	viper.BindPFlag("selectorColumn", pflag.Lookup("selectorcolumn"))

	viper.SetConfigName(appName)
	// Load config values from file
	if viper.GetString("configFile") != "" {
		viper.AddConfigPath(viper.GetString("configFile"))
		err := viper.ReadInConfig() // Find and read the config file
		if err != nil {             // Handle errors reading the config file
			return err
		}
	}

	// Load any config values from Environment variables who's name is prefixed TCS_ (Transicator Change Server)
	viper.SetEnvPrefix("tcs") // will be uppercased automatically
	viper.AutomaticEnv()

	return nil

}
