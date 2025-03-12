package test

import (
	"os"
	"reflect"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/key-inside/hiss"
)

func Test_WithViper(t *testing.T) {
	os.Setenv("CHICKEN_SOUND", "cluck")
	os.Setenv("CHICKEN_FOOT", "2")

	v := viper.New()
	v.AutomaticEnv()

	h, _ := hiss.New(v) // no options, no errors

	srcs := []string{
		"./fixtures/config-1.yaml",
		"./fixtures/config-2.yaml",
	}

	if err := h.ReadInSources(srcs); err != nil {
		t.Errorf("Failed to read in %s: %v", h.ConfigSrcUsed(), err)
		return
	}

	for _, animal := range []string{"snake", "cat", "dog", "chicken"} {
		t.Logf("<%s> sound \"%s\", foot %d", animal, v.GetString(animal+".sound"), v.GetInt(animal+".foot"))
	}
}

func Test_WithCobra(t *testing.T) {
	c := &cobra.Command{
		Use:     "hiss",
		Short:   "Hiss",
		Long:    "Hiss of Cobra",
		Version: "v0.0.0",
		Run: func(cmd *cobra.Command, args []string) {
			t.Log("Cobra & Viper hiss!")
		},
	}

	c.PersistentFlags().StringSlice("configs", []string{}, "config source URIs (file or ARN)")

	v := viper.New()
	v.BindPFlag("sources", c.PersistentFlags().Lookup("configs"))
	// v.AutomaticEnv()

	cobra.OnInitialize(func() {
		h, _ := hiss.New(v) // no options, no errors
		if err := h.ReadInSources(v.GetStringSlice("sources")); err != nil {
			t.Errorf("Failed to read in %s: %v", h.ConfigSrcUsed(), err)
			return
		}
	})

	c.SetArgs([]string{"--configs", "./fixtures/config-1.yaml,./fixtures/config-2.yaml"})
	if err := c.Execute(); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	for _, animal := range []string{"snake", "cat", "dog"} {
		t.Logf("<%s> sound \"%s\", foot %d", animal, v.GetString(animal+".sound"), v.GetInt(animal+".foot"))
	}
}

func Test_Unmarshal(t *testing.T) {
	v := viper.New()
	h, _ := hiss.New(v) // no options, no errors

	srcs := []string{
		"./fixtures/config-1.yaml",
		"./fixtures/config-2.yaml",
	}
	if err := h.ReadInSources(srcs); err != nil {
		t.Errorf("Failed to read in %s: %v", h.ConfigSrcUsed(), err)
		return
	}

	type Animal struct {
		Sound string
		Foot  int
	}
	type Config struct {
		Snake Animal
		Cat   Animal
		Dog   Animal
	}
	cfg := Config{}
	if err := v.Unmarshal(&cfg); err != nil {
		t.Errorf("Failed to unmarshal: %v", err)
		return
	}

	fields := reflect.VisibleFields(reflect.TypeOf(cfg))
	for _, f := range fields {
		animal := reflect.ValueOf(cfg).FieldByName(f.Name).Interface().(Animal)
		t.Logf("<%s> sound \"%v\", foot %v", f.Name, animal.Sound, animal.Foot)
	}
}
