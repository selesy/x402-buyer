package payer

import "time"

type Options struct {
	nonceFunc NonceFunc
	nowFunc   NowFunc
}

func NewOptions(opts ...Option) (*Options, error) {
	options := &Options{
		nonceFunc: DefaultNonce,
		nowFunc:   time.Now,
	}

	for _, opt := range opts {
		if err := opt(options); err != nil {
			return nil, err
		}
	}

	return options, nil
}

type Option func(*Options) error

func WithNonceFunc(nonceFunc NonceFunc) Option {
	return func(o *Options) error {
		o.nonceFunc = nonceFunc

		return nil
	}
}

func WithNowFunc(nowFunc NowFunc) Option {
	return func(o *Options) error {
		o.nowFunc = nowFunc

		return nil
	}
}
