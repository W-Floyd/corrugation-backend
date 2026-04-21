/*
Copyright © 2021-2022 William Floyd (william.png2000@gmail.com)

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/peterbourgon/diskv/v3"
	"github.com/spf13/cobra"

	"github.com/spf13/viper"
)

var (
	cfgFile string
	d       *diskv.Diskv
	store   Store
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "corrugation-backend",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: server,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.corrugation-backend.yaml)")

	viper.SetEnvPrefix("CORRUGATION")

	// rootCmd.Flags().StringP("assets", "a", "assets", "Assets location")
	// viper.BindPFlag("assets", rootCmd.Flags().Lookup("assets"))

	rootCmd.Flags().StringP("data", "d", "data", "Data location")
	viper.BindPFlag("data", rootCmd.Flags().Lookup("data"))

	rootCmd.Flags().Int("port", 8083, "Port to run server on")
	viper.BindPFlag("port", rootCmd.Flags().Lookup("port"))

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".corrugation-backend" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".corrugation-backend")
	}

	viper.SetEnvPrefix("CORRUGATION")

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}

}

func AdvancedTransformExample(key string) *diskv.PathKey {
	path := strings.Split(key, "/")
	last := len(path) - 1
	return &diskv.PathKey{
		Path:     path[:last],
		FileName: path[last],
	}
}

// If you provide an AdvancedTransform, you must also provide its
// inverse:

func InverseTransformExample(pathKey *diskv.PathKey) (key string) {
	return strings.Join(pathKey.Path, "/") + "/" + pathKey.FileName
}

func props(v ...any) map[string]any {
	if len(v)%2 != 0 {
		panic("uneven number of key/value pairs")
	}

	m := map[string]any{}
	for i := 0; i < len(v); i += 2 {
		m[fmt.Sprint(v[i])] = v[i+1]
	}

	return m
}

func server(cmd *cobra.Command, args []string) {

	// Initialize a new diskv store, rooted at "data", with a 16MB cache.
	d = diskv.New(diskv.Options{
		BasePath:          viper.GetString("data"),
		AdvancedTransform: AdvancedTransformExample,
		InverseTransform:  InverseTransformExample,
		CacheSizeMax:      16 * 1024 * 1024,
	})

	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	e.Static("/assets", "dist/assets")
	e.File("/favicon.ico", "dist/favicon.ico")

	e.GET("/ws", wsHandler)

	r := e.Group("/api")

	// Restricted group

	r.GET("/store", dumpStore) // If you want the whole JSON file, you can have it...
	r.GET("/store/version", storeVersion)
	r.Any("/reset", func(c echo.Context) error {
		resetStore()
		d.EraseAll()
		return c.NoContent(http.StatusOK)
	})

	r.GET("/qrcode/:id", qrGenerate)

	r.POST("/artifact", uploadArtifact)
	r.GET("/artifact/:id", downloadArtifact)
	r.DELETE("/artifact/:id", deleteArtifact)
	r.GET("/artifact/:id/qrcode", qrGenerate)
	r.GET("/artifact/list", listArtifacts)

	r.GET("/entity", getEntities)
	r.POST("/entity", createEntity)
	r.GET("/entity/:id", getEntity)
	r.PUT("/entity/:id", replaceEntity)
	r.PATCH("/entity/:id", patchEntity)
	r.DELETE("/entity/:id", deleteEntity)
	r.GET("/entity/find/children/:id/full", getContainsFull)
	r.GET("/entity/find/children/:id/full/recursive", getContainsFullRecursive)
	r.GET("/entity/find/locations", findEntitiesWithChildren)
	r.GET("/entity/find/locations/full", findEntitiesWithChildrenFull)
	r.GET("/entity/find/firstid", firstId)
	r.GET("/entity/find/nextid", func(c echo.Context) error {
		return c.JSON(http.StatusOK, store.LastEntityID()+1)
	})
	r.GET("/entity/find/firstfreeid", func(c echo.Context) error {
		free := emptyIDs()
		min := store.LastEntityID() + 1
		for _, v := range free {
			if v < min {
				min = v
			}
		}
		return c.JSON(http.StatusOK, min)
	})
	r.GET("/entity/:id/contains", getContains)
	r.GET("/entity/:id/qrcode", qrGenerate)
	r.GET("/entity/list", listEntities)

	e.GET("/*", func(c echo.Context) error {
		return c.File("dist/index.html")
	})

	if d.Has("store.json") {
		data, err := d.Read("store.json")
		if err != nil {
			e.Logger.Fatal(err)
		}
		json.Unmarshal(data, &store)
	} else {
		resetStore()
	}

	e.Logger.Fatal(e.Start(":" + strconv.Itoa(viper.GetInt("port"))))

}

// chanToSlice reads all data from ch (which must be a chan), returning a
// slice of the data. If ch is a 'T chan' then the return value is of type
// []T inside the returned interface.
// A typical call would be sl := ChanToSlice(ch).([]int)
func chanToSlice(ch interface{}) interface{} {
	chv := reflect.ValueOf(ch)
	slv := reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(ch).Elem()), 0, 0)
	for {
		v, ok := chv.Recv()
		if !ok {
			return slv.Interface()
		}
		slv = reflect.Append(slv, v)
	}
}

func checkForm(requires []string, c echo.Context) error {

	for _, check := range requires {
		if hasForm(check, c) {
			return c.JSON(http.StatusBadRequest, check+" not provided")
		}
	}
	return nil
}

func checkFormFiles(requires []string, c echo.Context) error {

	for _, check := range requires {
		_, err := c.FormFile(check)
		if err != nil {
			return c.JSON(http.StatusBadRequest, check+" not provided")
		}
	}
	return nil
}

func hasForm(formKey string, c echo.Context) bool {
	return c.FormValue(formKey) != ""
}
