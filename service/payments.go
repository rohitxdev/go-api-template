package service

import (
	"github.com/rohitxdev/go-api-template/config"
	"github.com/stripe/stripe-go/v76"
)

func init() {
	stripe.Key = config.STRIPE_API_KEY
}
