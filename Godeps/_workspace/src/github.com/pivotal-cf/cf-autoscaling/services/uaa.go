package services

import "github.com/evanfarrar/uaa-sso-golang/uaa"

type UAAInterface interface {
	uaa.ExchangeInterface
	uaa.RefreshInterface
	uaa.LoginURLInterface
	uaa.GetClientTokenInterface
	uaa.GetTokenKeyInterface
}
