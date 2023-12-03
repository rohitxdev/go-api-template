package service

import (
	"github.com/stripe/stripe-go/v76"

	"github.com/rohitxdev/go-api-template/env"
)

func init() {
	stripe.Key = env.STRIPE_API_KEY
}
