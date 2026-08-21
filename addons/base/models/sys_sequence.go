package models

import (
	"sumeru/core/sdk"
)

type SysSequence struct {
	sdk.Model `sumeru:"model=sys.sequence"`

	Name       sdk.String  `sumeru:"required,string=Name"`
	Code       sdk.String  `sumeru:"required,unique,index,string=Code"`
	Prefix     sdk.String  `sumeru:"string=Prefix"`
	Suffix     sdk.String  `sumeru:"string=Suffix"`
	Padding    sdk.Integer `sumeru:"string=Padding,default=5"`
	NumberNext sdk.Integer `sumeru:"string=Next Number,default=1"`
	Active     sdk.Boolean `sumeru:"string=Active,default=true"`
}
