/* Copyright (C) 2023 Nuts community
*
* This program is free software: you can redistribute it and/or modify
* it under the terms of the GNU General Public License as published by
* the Free Software Foundation, either version 3 of the License, or
* (at your option) any later version.
*
* This program is distributed in the hope that it will be useful,
* but WITHOUT ANY WARRANTY; without even the implied warranty of
* MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
* GNU General Public License for more details.
*
* You should have received a copy of the GNU General Public License
* along with this program.  If not, see <https://www.gnu.org/licenses/>.
*
 */

package config

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"os"
	"strings"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/spf13/pflag"
)

const defaultPrefix = "NUTS"
const defaultDelimiter = "."
const configFileFlag = "configfile"
const defaultConfigFile = "server.config.yaml"
const defaultNutsNodeAddress = "http://localhost:1323"

func defaultConfig() Config {
	return Config{
		NutsNodeAddr: defaultNutsNodeAddress,
	}
}

type Config struct {
	// NutsNodeAddr contains the address of the Nuts node. It's also used in the aud field when API security is enabled
	NutsNodeAddr string `koanf:"nutsnodeaddr"`
	// NutsNodeInternalAddr contains the address of the Nuts node for calls to /internal endpoints, in case these are bound to a separate HTTP interface.
	// If empty, NutsNodeAddr is used.
	NutsNodeInternalAddr string `koanf:"nutsnodeinternaladdr"`
	// NutsNodeAPIKeyFile points to the private key used to sign JWTs. If empty Nuts node API security is not enabled
	NutsNodeAPIKeyFile string `koanf:"nutsnodeapikeyfile"`
	// NutsNodeAPIUser contains the API key user that will go into the iss field. It must match the user with the public key from the authorized_keys file in the Nuts node
	NutsNodeAPIUser string `koanf:"nutsnodeapiuser"`
	// NutsNodeAPIAudience dictates the aud field of the created JWT
	NutsNodeAPIAudience string `kaonf:"nutsnodeapiaudience"`
	ApiKey              crypto.Signer
	// WithMockNode enables the mock Nuts node
	WithMockNode bool `koanf:"withmocknode"`
}

func (c Config) Print(writer io.Writer) error {
	if _, err := fmt.Fprintln(writer, "========== CONFIG: =========="); err != nil {
		return err
	}
	var pr Config = c
	data, _ := json.MarshalIndent(pr, "", "  ")
	if _, err := fmt.Fprintln(writer, string(data)); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(writer, "========= END CONFIG ========="); err != nil {
		return err
	}
	return nil
}

func LoadConfig() Config {
	flagset := loadFlagSet(os.Args[1:])

	var k = koanf.New(defaultDelimiter)

	// Prepare koanf for parsing the config file
	configFilePath := resolveConfigFile(flagset)
	// Check if the file exists
	if _, err := os.Stat(configFilePath); err == nil {
		log.Printf("Loading config from file: %s", configFilePath)
		if err := k.Load(file.Provider(configFilePath), yaml.Parser()); err != nil {
			log.Fatalf("error while loading config from file: %v", err)
		}
	} else {
		log.Printf("Using default config because no file was found at: %s", configFilePath)
	}

	config := defaultConfig()

	// Unmarshal values of the config file into the config struct, potentially replacing default values
	if err := k.Unmarshal("", &config); err != nil {
		log.Fatalf("error while unmarshalling config: %v", err)
	}

	// load env flags, can't return error
	_ = k.Load(envProvider(), nil)

	// load cmd flags, without a parser, no error can be returned
	_ = k.Load(posflag.Provider(flagset, defaultDelimiter, k), nil)

	// Load the API key
	if len(config.NutsNodeAPIKeyFile) > 0 {
		bytes, err := os.ReadFile(config.NutsNodeAPIKeyFile)
		if err != nil {
			log.Fatalf("error while reading private key file: %v", err)
		}
		config.ApiKey, err = pemToPrivateKey(bytes)
		if err != nil {
			log.Fatalf("error while decoding private key file: %v", err)
		}
		if len(config.NutsNodeAPIUser) == 0 {
			log.Fatal("nutsnodeapiuser config is required with nutsnodeapikeyfile")
		}
		if len(config.NutsNodeAPIAudience) == 0 {
			log.Fatal("nutsnodeapiaudience config is required with nutsnodeapikeyfile")
		}
	}

	if k.Bool("withmocknode") {
		config.WithMockNode = true
	}

	return config
}

func loadFlagSet(args []string) *pflag.FlagSet {
	f := pflag.NewFlagSet("config", pflag.ContinueOnError)
	// add withmocknode flag
	f.String(configFileFlag, defaultConfigFile, "Nuts monitor config file")
	f.Usage = func() {
		fmt.Println(f.FlagUsages())
		os.Exit(0)
	}
	f.Parse(args)
	return f
}

// resolveConfigFile resolves the path of the config file using the following sources:
// 1. commandline params (using the given flags)
// 2. environment vars,
// 3. default location.
func resolveConfigFile(flagset *pflag.FlagSet) string {

	k := koanf.New(defaultDelimiter)

	// load env flags, can't return error
	_ = k.Load(envProvider(), nil)

	// load cmd flags, without a parser, no error can be returned
	_ = k.Load(posflag.Provider(flagset, defaultDelimiter, k), nil)

	configFile := k.String(configFileFlag)
	return configFile
}

func envProvider() *env.Env {
	return env.Provider(defaultPrefix, defaultDelimiter, func(s string) string {
		return strings.Replace(strings.ToLower(
			strings.TrimPrefix(s, defaultPrefix)), "_", defaultDelimiter, -1)
	})
}

// pemToPrivateKey converts a PEM encoded private key to a Signer interface. It supports EC, RSA and PKIX PEM encoded strings
func pemToPrivateKey(bytes []byte) (signer crypto.Signer, err error) {
	key, _ := ssh.ParseRawPrivateKey(bytes)
	if key == nil {
		err = errors.New("failed to decode PEM file")
		return
	}

	switch k := key.(type) {
	case *rsa.PrivateKey:
		signer = k
	case *ecdsa.PrivateKey:
		signer = k
	default:
		err = fmt.Errorf("unsupported private key type: %T", k)
	}

	return
}
