package api

var Weekdays = []struct {
	Value  Weekday
	TSName string
}{
	{Monday, string(Monday)},
	{Tuesday, string(Tuesday)},
	{Wednesday, string(Wednesday)},
	{Thursday, string(Thursday)},
	{Friday, string(Friday)},
	{Saturday, string(Saturday)},
	{Sunday, string(Sunday)},
}
