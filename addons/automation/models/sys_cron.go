package models

import (
	"sumeru/core/sdk"
)

type SysCron struct {
	sdk.Model `sumeru:"model=sys.cron"`

	Name           sdk.String   `sumeru:"required,string=Name"`
	Code           sdk.String   `sumeru:"string=Code"`
	EventName      sdk.String   `sumeru:"string=Event Name"`
	IntervalNumber sdk.Integer  `sumeru:"string=Interval (minutes),default=60"`
	Active         sdk.Boolean  `sumeru:"string=Active,default=true"`
	NextCall       sdk.DateTime `sumeru:"string=Next Call"`
	LastCall       sdk.DateTime `sumeru:"string=Last Call"`
}
