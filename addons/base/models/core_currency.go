package models

import (
	"sumeru/core/sdk"
)

type CoreCurrency struct {
	sdk.Model `sumeru:"model=core.currency"`

	Name   sdk.String  `sumeru:"required,unique,index,string=Currency"`
	Symbol sdk.String  `sumeru:"required,string=Symbol"`
	Active sdk.Boolean `sumeru:"string=Active,default=true"`
}
