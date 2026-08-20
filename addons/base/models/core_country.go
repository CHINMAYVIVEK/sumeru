package models

import (
	"sumeru/core/sdk"
)

type CoreCountry struct {
	sdk.Model `sumeru:"model=core.country"`

	Name      sdk.String  `sumeru:"required,string=Country"`
	Code      sdk.String  `sumeru:"required,unique,index,string=Code"`
	PhoneCode sdk.String  `sumeru:"string=Phone Code"`
	Active    sdk.Boolean `sumeru:"string=Active,default=true"`
}
