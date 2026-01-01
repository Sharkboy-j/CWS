package internal

func DetermineTimezoneByCoordinates(lat, lon float64) string {
	offset := int(lon / 15.0)

	if offset < -12 {
		offset = -12
	}
	if offset > 14 {
		offset = 14
	}

	timezoneMap := map[int]string{
		-12: "Etc/GMT+12",
		-11: "Pacific/Midway",
		-10: "Pacific/Honolulu",
		-9:  "America/Anchorage",
		-8:  "America/Los_Angeles",
		-7:  "America/Denver",
		-6:  "America/Chicago",
		-5:  "America/New_York",
		-4:  "America/Caracas",
		-3:  "America/Sao_Paulo",
		-2:  "Atlantic/South_Georgia",
		-1:  "Atlantic/Azores",
		0:   "UTC",
		1:   "Europe/Paris",
		2:   "Europe/Berlin",
		3:   "Europe/Minsk",
		4:   "Asia/Dubai",
		5:   "Asia/Karachi",
		6:   "Asia/Dhaka",
		7:   "Asia/Bangkok",
		8:   "Asia/Shanghai",
		9:   "Asia/Tokyo",
		10:  "Australia/Sydney",
		11:  "Pacific/Guadalcanal",
		12:  "Pacific/Auckland",
		13:  "Pacific/Tongatapu",
		14:  "Pacific/Kiritimati",
	}

	if tz, ok := timezoneMap[offset]; ok {
		return tz
	}

	return "UTC"
}
