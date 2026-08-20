package models

import (
	"sumeru/core/sdk"
)

type SysConfigParameter struct {
	sdk.Model `sumeru:"model=sys.config.parameter"`

	Key   sdk.String `sumeru:"required,unique,index,string=Key"`
	Value sdk.Text   `sumeru:"string=Value"`
}
