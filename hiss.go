package hiss

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/spf13/viper"

	"github.com/key-inside/hiss/aws"
)

type Option func(*Hiss) error

// Default is `slog.Default()`
func WithLogger(l *slog.Logger) Option {
	return func(h *Hiss) error {
		if l == nil {
			return fmt.Errorf("nil logger")
		}
		h.logger = l
		return nil
	}
}

// Default is "." -> "_"
func WithEnvKeyReplace(oldnew ...string) Option {
	return func(h *Hiss) error {
		h.v.SetEnvKeyReplacer(strings.NewReplacer(oldnew...))
		return nil
	}
}

type Hiss struct {
	v *viper.Viper

	logger  *slog.Logger
	srcUsed string
}

func New(v *viper.Viper, options ...Option) (*Hiss, error) {
	h := &Hiss{v: v}

	options = append(
		[]Option{
			WithLogger(slog.Default()),
			WithEnvKeyReplace(".", "_"),
		},
		options...,
	)

	for _, option := range options {
		if err := option(h); err != nil {
			return nil, fmt.Errorf("failed to apply Hiss option: %w", err)
		}
	}

	return h, nil
}

func (h *Hiss) ReadInSources(srcs []string, awsOps ...aws.Option) error {
	// loads configs from file or aws
	for _, src := range srcs {
		h.srcUsed = src
		if arn.IsARN(src) {
			h.logger.Debug("load config from AWS", "arn", src)
			cfgMap, err := aws.GetConfigMap(src, awsOps...)
			if err != nil {
				return fmt.Errorf("failed to get config: %w", err)
			}
			if err := h.v.MergeConfigMap(cfgMap); err != nil {
				return fmt.Errorf("can't merge config: %w", err)
			}
		} else {
			h.logger.Debug("load config file", "filepath", src)
			h.v.SetConfigFile(src)
			if err := h.v.MergeInConfig(); err != nil {
				return fmt.Errorf("can't merge config: %w", err)
			}
		}
	}

	return nil
}

func (h *Hiss) ConfigSrcUsed() string { return h.srcUsed }
