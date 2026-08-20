package orm

import (
	"sumeru/core/modelmeta"
)

type SysModel struct {
	modelmeta.ModelMeta `sumeru:"model=sys.model"`

	Name        modelmeta.String `sumeru:"required,unique"`
	ResModel    modelmeta.String `sumeru:"required,column=model"`
	Module      modelmeta.String `sumeru:"index,string=Declaring module"`
	Description modelmeta.Text
	Transient   modelmeta.Boolean
}
