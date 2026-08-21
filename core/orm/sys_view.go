package orm

import (
	"sumeru/core/modelmeta"
)

type SysView struct {
	modelmeta.ModelMeta `sumeru:"model=sys.view"`

	Name     modelmeta.String `sumeru:"required,unique"`
	ResModel modelmeta.String `sumeru:"required,column=model"`
	Type     modelmeta.String `sumeru:"required"`
	Arch     modelmeta.Text
	Priority modelmeta.Integer
	Active   modelmeta.Boolean
}
