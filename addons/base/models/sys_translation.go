package models

import (
	"sumeru/core/sdk"
)

type SysTranslation struct {
	sdk.Model `sumeru:"model=sys.translation"`

	Lang   sdk.String `sumeru:"required,index,string=Language"`
	Src    sdk.Text   `sumeru:"required,string=Source"`
	Value  sdk.Text   `sumeru:"string=Translation"`
	Module sdk.String `sumeru:"index,string=Module"`
}
