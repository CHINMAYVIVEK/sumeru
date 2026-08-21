package orm

import (
	"sumeru/core/modelmeta"
)

type SysActionWindow struct {
	modelmeta.ModelMeta `sumeru:"model=sys.action.window"`

	Name      modelmeta.String `sumeru:"required,unique"`
	CoreModel modelmeta.String `sumeru:"required"`
	ViewMode  modelmeta.String
	Domain    modelmeta.String
	Context   modelmeta.String
	Help      modelmeta.String
}
